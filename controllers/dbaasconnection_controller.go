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
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
)

// DBaaSConnectionReconciler reconciles a DBaaSConnection object
type DBaaSConnectionReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx, "DBaaS Connection", req.NamespacedName)

	var connection v1alpha1.DBaaSConnection
	if err := r.Get(ctx, req.NamespacedName, &connection); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.Info("DBaaS Connection resource not found, has been deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaS Connection for reconcile")
		return ctrl.Result{}, err
	}

	var inventory v1alpha1.DBaaSInventory
	if err := r.Get(ctx, types.NamespacedName{Namespace: connection.Spec.InventoryRef.Namespace, Name: connection.Spec.InventoryRef.Name}, &inventory); err != nil {
		logger.Error(err, "Error fetching DBaaS Inventory resource reference for DBaaS Connection", "DBaaS Inventory", connection.Spec.InventoryRef.Name)
		return ctrl.Result{}, err
	}

	provider, err := r.getDBaaSProvider(inventory.Spec.Provider, req.Namespace, ctx)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Requested DBaaS Provider is not configured in this environment", "DBaaS Provider", provider.Provider)
			return ctrl.Result{}, err
		}
		logger.Error(err, "Error reading configured DBaaS Provider", "DBaaS Provider", inventory.Spec.Provider)
		return ctrl.Result{}, err
	}
	logger.Info("Found DBaaS Provider", "DBaaS Provider", provider.Provider)

	if providerConnection, err := r.getProviderObject(&connection, provider.ConnectionKind, ctx); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Provider Connection resource not found", "DBaaS Provider", connection.GetName())
			if err = r.createProviderObject(&connection, provider.ConnectionKind, connection.Spec.DeepCopy(), ctx); err != nil {
				logger.Error(err, "Error creating Provider Connection resource", "DBaaS Provider", connection.GetName())
				return ctrl.Result{}, err
			}
			logger.Info("Provider Connection resource created", "DBaaS Provider", connection.GetName())
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error finding the Provider Connection resource", "DBaaS Provider", connection.GetName())
		return ctrl.Result{}, err
	} else {
		if providerConnectionStatus, exists := providerConnection.UnstructuredContent()["status"]; exists {
			var status v1alpha1.DBaaSConnectionStatus
			if err := decode(providerConnectionStatus, &status); err != nil {
				logger.Error(err, "Error parsing the status of the Provider Connection resource", "DBaaS Provider", providerConnection.GetName())
				return ctrl.Result{}, err
			} else {
				connection.Status = *status.DeepCopy()
				if err := r.Status().Update(ctx, &connection); err != nil {
					logger.Error(err, "Error updating the DBaaS Connection status")
					return ctrl.Result{}, err
				}
				logger.Info("DBaaS Connection status updated")
			}
		} else {
			logger.Info("Provider Connection resource status not found", "DBaaS Provider", providerConnection.GetName())
		}

		if providerConnectionSpec, exists := providerConnection.UnstructuredContent()["spec"]; exists {
			var spec v1alpha1.DBaaSConnectionSpec
			if err := decode(providerConnectionSpec, &spec); err != nil {
				logger.Error(err, "Error parsing the spec of the Provider Connection resource", "DBaaS Provider", providerConnection.GetName())
				return ctrl.Result{}, err
			} else {
				if !reflect.DeepEqual(spec, connection.Spec) {
					if err = r.updateProviderObject(providerConnection, connection.Spec.DeepCopy(), ctx); err != nil {
						logger.Error(err, "Error updating the Provider Connection spec", "DBaaS Provider", providerConnection.GetName())
						return ctrl.Result{}, err
					}
					logger.Info("Provider Connection spec updated")
				}
			}
		} else {
			err = fmt.Errorf("failed to get the spec of the Provider Connection %s", providerConnection.GetName())
			logger.Error(err, "Error getting the spec of the Provider Connection", "DBaaS Provider", providerConnection.GetName())
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSConnectionReconciler) SetupWithManager(mgr ctrl.Manager, namespace string) error {
	logger := ctrl.Log.WithName("controllers").WithName("DBaaS Connection")

	logger.Info("Read configured DBaaS Providers from ConfigMaps")
	owned, err := r.preStartGetDBaaSProviderConnections(namespace)
	if err != nil {
		logger.Error(err, "Error reading the configured DBaaS Providers from ConfigMaps")
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr)
	builder = builder.For(&v1alpha1.DBaaSConnection{})
	for _, o := range owned {
		builder = builder.Owns(&o)
	}
	return builder.Complete(r)
}
