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
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DBaaSInventoryReconciler reconciles a DBaaSInventory object
type DBaaSInventoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	if providerInventory, err := getProviderCR(&inventory, provider.InventoryKind, r.Client, ctx); err != nil {
		return ctrl.Result{}, err
	} else if providerInventory == nil {
		if err = createProviderCR(&inventory, provider.InventoryKind, inventory.Spec.DBaaSInventorySpec.DeepCopy(), r.Client, r.Scheme, ctx); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	} else {
		if providerInventoryStatus, exists := providerInventory.UnstructuredContent()["status"]; exists {
			var status v1alpha1.DBaaSInventoryStatus
			if err := decode(providerInventoryStatus, &status); err != nil {
				logger.Error(err, "Error parsing the status of the provider inventory CR", "ProviderInventoryStatus", providerInventoryStatus)
				return ctrl.Result{}, err
			} else {
				inventory.Status = *status.DeepCopy()
				if err := r.Status().Update(ctx, &inventory); err != nil {
					logger.Error(err, "Error updating DBaaSInventory status")
					return ctrl.Result{}, err
				}
				logger.Info("DBaaSInventory status updated")
			}
		} else {
			logger.Info("Provider inventory resource status not found", "providerInventory", providerInventory)
		}

		if providerInventorySpec, exists := providerInventory.UnstructuredContent()["spec"]; exists {
			var spec v1alpha1.DBaaSInventorySpec
			if err := decode(providerInventorySpec, &spec); err != nil {
				logger.Error(err, "Error parsing the spec of the provider inventory CR", "ProviderInventorySpec", providerInventorySpec)
				return ctrl.Result{}, err
			} else {
				if !reflect.DeepEqual(spec, inventory.Spec.DBaaSInventorySpec) {
					if err = updateProviderCR(&inventory, provider.InventoryKind, inventory.Spec.DBaaSInventorySpec.DeepCopy(), r.Client, r.Scheme, ctx); err != nil {
						return ctrl.Result{}, err
					}
					logger.Info("Provider inventory spec updated")
				}
			}
		} else {
			err = fmt.Errorf("failed to get the spec of the provider inventory %v", providerInventory)
			logger.Error(err, "Error getting the spec of the provider inventory CR")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSInventoryReconciler) SetupWithManager(mgr ctrl.Manager, namespace string) error {
	owned, err := getDBaaSProviderInventoryObjects(namespace)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr)
	builder = builder.For(&v1alpha1.DBaaSInventory{})
	for _, o := range owned {
		builder = builder.Owns(&o)
	}
	return builder.Complete(r)
}
