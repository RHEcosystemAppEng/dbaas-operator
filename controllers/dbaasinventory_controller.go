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

// DBaaSInventoryReconciler reconciles a DBaaSInventory object
type DBaaSInventoryReconciler struct {
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
func (r *DBaaSInventoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "dbaasinventory", req.NamespacedName)

	var inventory v1alpha1.DBaaSInventory
	if err := r.Get(ctx, req.NamespacedName, &inventory); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.Info("DBaaSInventory resource not found, has been deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaSInventory for reconcile")
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

	if providerInventory, err := r.getProviderInventory(inventory, provider, ctx); err != nil {
		return ctrl.Result{}, err
	} else if providerInventory == nil {
		if err = r.createProviderInventory(inventory, provider, ctx); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if !reflect.DeepEqual(providerInventory.UnstructuredContent()["spec"], inventory.Spec) {
		if err = r.updateProviderInventory(inventory, provider, ctx); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	} else {
		if providerInventoryStatus, exists := providerInventory.UnstructuredContent()["status"]; exists {
			if status, ok := providerInventoryStatus.(v1alpha1.DBaaSInventoryStatus); ok {
				inventory.Status.Conditions = status.Conditions
				inventory.Status.Type = status.Type
				inventory.Status.Instances = status.Instances
				if err := r.Status().Update(ctx, &inventory); err != nil {
					logger.Error(err, "Error updating DBaaSInventory status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
		}

		err = fmt.Errorf("failed to parse the status of the provider inventory %v", providerInventory)
		logger.Error(err, "Error parsing the status of the provider inventory CR")
		return ctrl.Result{}, err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSInventoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSInventory{}).
		Complete(r)
}

func (r *DBaaSInventoryReconciler) createProviderInventory(inventory v1alpha1.DBaaSInventory, provider v1alpha1.DBaaSProvider, ctx context.Context) error {
	logger := log.FromContext(ctx, "dbaasinventory", types.NamespacedName{Namespace: inventory.Namespace, Name: inventory.Name})

	providerInventory := &unstructured.Unstructured{}
	providerInventory.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   inventory.GroupVersionKind().Group,
		Version: inventory.GroupVersionKind().Version,
		Kind:    provider.InventoryKind,
	})
	providerInventory.SetNamespace(inventory.GetNamespace())
	providerInventory.SetName(inventory.GetName())
	providerInventory.UnstructuredContent()["spec"] = inventory.Spec.DeepCopy()
	if err := r.Create(ctx, providerInventory); err != nil {
		logger.Error(err, "Error creating a provider inventory", "providerInventory", providerInventory)
		return err
	}
	logger.Info("Provider inventory resource created", "providerInventory", providerInventory)
	return nil
}

func (r *DBaaSInventoryReconciler) updateProviderInventory(inventory v1alpha1.DBaaSInventory, provider v1alpha1.DBaaSProvider, ctx context.Context) error {
	logger := log.FromContext(ctx, "dbaasinventory", types.NamespacedName{Namespace: inventory.Namespace, Name: inventory.Name})

	providerInventory := &unstructured.Unstructured{}
	providerInventory.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   inventory.GroupVersionKind().Group,
		Version: inventory.GroupVersionKind().Version,
		Kind:    provider.InventoryKind,
	})
	providerInventory.SetNamespace(inventory.GetNamespace())
	providerInventory.SetName(inventory.GetName())
	providerInventory.UnstructuredContent()["spec"] = inventory.Spec.DeepCopy()
	if err := r.Update(ctx, providerInventory); err != nil {
		logger.Error(err, "Error updating a provider inventory", "providerInventory", providerInventory)
		return err
	}
	logger.Info("Provider inventory resource updated", "providerInventory", providerInventory)
	return nil
}

func (r *DBaaSInventoryReconciler) getProviderInventory(inventory v1alpha1.DBaaSInventory, provider v1alpha1.DBaaSProvider, ctx context.Context) (*unstructured.Unstructured, error) {
	gvk := schema.GroupVersionKind{
		Group:   inventory.GroupVersionKind().Group,
		Version: inventory.GroupVersionKind().Version,
		Kind:    provider.InventoryKind,
	}

	logger := log.FromContext(ctx, "dbaasinventory", types.NamespacedName{Namespace: inventory.Namespace, Name: inventory.Name}, "GVK", gvk)

	var providerInventory = unstructured.Unstructured{}
	providerInventory.SetGroupVersionKind(gvk)
	if err := r.Get(ctx, client.ObjectKeyFromObject(&inventory), &providerInventory); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Provider inventory resource not found", "providerInventory", inventory.GetName())
			return nil, nil
		}
		logger.Error(err, "Error finding the provider inventory", "providerInventory", inventory.GetName())
		return nil, err
	}
	return &providerInventory, nil
}
