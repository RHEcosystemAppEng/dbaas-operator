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

	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

// DBaaSProviderReconciler reconciles a DBaaSProvider object
type DBaaSProviderReconciler struct {
	*DBaaSReconciler
	ConnectionCtrl controller.Controller
	InventoryCtrl  controller.Controller
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	var provider v1alpha1.DBaaSProvider
	if err := r.Get(ctx, req.NamespacedName, &provider); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.V(1).Info("DBaaS Provider resource not found, has been deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaS Provider for reconcile")
		return ctrl.Result{}, err
	}

	if err := r.watchDBaaSProviderObject(r.InventoryCtrl, &v1alpha1.DBaaSInventory{}, provider.Spec.InventoryKind); err != nil {
		logger.Error(err, "Error watching Provider Inventory CR")
		return ctrl.Result{}, err
	}
	logger.Info("Watching Provider Inventory CR")

	if err := r.watchDBaaSProviderObject(r.ConnectionCtrl, &v1alpha1.DBaaSConnection{}, provider.Spec.ConnectionKind); err != nil {
		logger.Error(err, "Error watching Provider Connection CR")
		return ctrl.Result{}, err
	}
	logger.Info("Watching Provider Connection CR")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSProvider{}, builder.WithPredicates(filterEventPredicate)).
		Complete(r)
}

var filterEventPredicate = predicate.Funcs{
	CreateFunc: func(createEvent event.CreateEvent) bool {
		return true
	},
	UpdateFunc: func(updateEvent event.UpdateEvent) bool {
		return updateEvent.ObjectNew.GetGeneration() != updateEvent.ObjectOld.GetGeneration()
	},
	DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
		return false
	},
	GenericFunc: func(genericEvent event.GenericEvent) bool {
		return false
	},
}
