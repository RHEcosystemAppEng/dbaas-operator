/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// DBaaSConnectionReconciler reconciles a DBaaSConnection object
type DBaaSConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "dbaasconnection", req.NamespacedName)

	var connection v1alpha1.DBaaSConnection
	if err := r.Get(ctx, req.NamespacedName, &connection); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.Info("DBaaSConnection resource not found, has been deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaSConnection for reconcile")
		return ctrl.Result{}, err
	}

	if connection.Spec.InventoryRef == nil {
		err := fmt.Errorf("inventory reference is missing for DBaaS connection %s", connection.Name)
		logger.Error(err, "Invalid DBaaSConnection for reconcile")
		return ctrl.Result{}, err
	}

	var inventory v1alpha1.DBaaSInventory
	if err := r.Get(ctx, types.NamespacedName{Namespace: connection.Namespace, Name: connection.Spec.InventoryRef.Name}, &inventory); err != nil {
		logger.Error(err, "Error fetching DBaaSInventory resource reference for DBaaSConnection reconcile", "Inventory", connection.Spec.InventoryRef.Name)
		return ctrl.Result{}, err
	}

	provider, err := getDBaaSProvider(inventory.Spec.Provider, req.Namespace, r.Client, ctx)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Requested DBaaS Provider is not configured in this environment", "Provider", provider.Provider)
			return ctrl.Result{}, err
		}
		logger.Error(err, "Error reading configured DBaaS providers")
		return ctrl.Result{}, err
	}
	logger.Info("Found DBaaS provider", "provider", provider)

	if providerConnection, err := r.getProviderConnection(connection, provider, ctx); err != nil {
		return ctrl.Result{}, err
	} else if providerConnection == nil {
		if err = r.createProviderConnection(connection, provider, ctx); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if !reflect.DeepEqual(providerConnection.UnstructuredContent()["spec"], connection.Spec) {
		if err = r.updateProviderConnection(connection, provider, ctx); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	} else {
		if providerConnectionStatus, exists := providerConnection.UnstructuredContent()["status"]; exists {
			if status, ok := providerConnectionStatus.(v1alpha1.DBaaSConnectionStatus); ok {
				connection.Status.Conditions = status.Conditions
				connection.Status.ConnectionString = status.ConnectionString
				connection.Status.CredentialsRef = status.CredentialsRef
				connection.Status.ConnectionInfo = status.ConnectionInfo
				if err := r.Status().Update(ctx, &connection); err != nil {
					logger.Error(err, "Error updating DBaaSConnection status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
		}

		err = fmt.Errorf("failed to parse the status of the provider connection %v", providerConnection)
		logger.Error(err, "Error parsing the status of the provider connection CR")
		return ctrl.Result{}, err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSConnection{}).
		Complete(r)
}

func (r *DBaaSConnectionReconciler) createProviderConnection(connection v1alpha1.DBaaSConnection, provider v1alpha1.DBaaSProvider, ctx context.Context) error {
	logger := log.FromContext(ctx, "dbaasconnection", types.NamespacedName{Namespace: connection.Namespace, Name: connection.Name})

	providerConnection := &unstructured.Unstructured{}
	providerConnection.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   connection.GroupVersionKind().Group,
		Version: connection.GroupVersionKind().Version,
		Kind:    provider.ConnectionKind,
	})
	providerConnection.SetNamespace(connection.GetNamespace())
	providerConnection.SetName(connection.GetName())
	providerConnection.UnstructuredContent()["spec"] = connection.Spec.DeepCopy()
	if err := r.Create(ctx, providerConnection); err != nil {
		logger.Error(err, "Error creating a provider connection", "providerConnection", providerConnection)
		return err
	}
	logger.Info("Provider connection resource created", "providerConnection", providerConnection)
	return nil
}

func (r *DBaaSConnectionReconciler) updateProviderConnection(connection v1alpha1.DBaaSConnection, provider v1alpha1.DBaaSProvider, ctx context.Context) error {
	logger := log.FromContext(ctx, "dbaasconnection", types.NamespacedName{Namespace: connection.Namespace, Name: connection.Name})

	providerConnection := &unstructured.Unstructured{}
	providerConnection.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   connection.GroupVersionKind().Group,
		Version: connection.GroupVersionKind().Version,
		Kind:    provider.ConnectionKind,
	})
	providerConnection.SetNamespace(connection.GetNamespace())
	providerConnection.SetName(connection.GetName())
	providerConnection.UnstructuredContent()["spec"] = connection.Spec.DeepCopy()
	if err := r.Update(ctx, providerConnection); err != nil {
		logger.Error(err, "Error updating a provider connection", "providerConnection", providerConnection)
		return err
	}
	logger.Info("Provider connection resource updated as ", "providerConnection", providerConnection)
	return nil
}

func (r *DBaaSConnectionReconciler) getProviderConnection(connection v1alpha1.DBaaSConnection, provider v1alpha1.DBaaSProvider, ctx context.Context) (*unstructured.Unstructured, error) {
	gvk := schema.GroupVersionKind{
		Group:   connection.GroupVersionKind().Group,
		Version: connection.GroupVersionKind().Version,
		Kind:    provider.ConnectionKind,
	}

	logger := log.FromContext(ctx, "dbaasconnection", types.NamespacedName{Namespace: connection.Namespace, Name: connection.Name}, "GVK", gvk)

	var providerConnection = unstructured.Unstructured{}
	providerConnection.SetGroupVersionKind(gvk)
	if err := r.Get(ctx, client.ObjectKeyFromObject(&connection), &providerConnection); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Provider connection resource not found", "providerConnection", connection.GetName())
			return nil, nil
		}
		logger.Error(err, "Error finding the provider connection", "providerConnection", connection.GetName())
		return nil, err
	}
	return &providerConnection, nil
}
