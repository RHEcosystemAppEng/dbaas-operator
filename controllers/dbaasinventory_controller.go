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

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/source"

	metrics "github.com/RHEcosystemAppEng/dbaas-operator/controllers/metrics"
)

// DBaaSInventoryReconciler reconciles a DBaaSInventory object
type DBaaSInventoryReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSInventoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	execution := metrics.PlatformInstallStart()
	logger := ctrl.LoggerFrom(ctx)
	var inventory v1beta1.DBaaSInventory
	metricLabelErrCdValue := ""
	event := ""

	if err := r.Get(ctx, req.NamespacedName, &inventory); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.V(1).Info("DBaaS Inventory resource not found, has been deleted")
			metricLabelErrCdValue = metrics.LabelErrorCdValueResourceNotFound
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaS Inventory for reconcile")
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrorFetchingDBaaSInventoryResources
		return ctrl.Result{}, err
	}

	if inventory.DeletionTimestamp != nil {
		event = metrics.LabelEventValueDelete
	} else {
		event = metrics.LabelEventValueCreate
	}

	policyList, err := r.policyListByNS(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "unable to list policies")
		metricLabelErrCdValue = metrics.LabelErrorCdValueUnableToListPolicies
		return ctrl.Result{}, err
	}
	activePolicy := getActivePolicy(policyList)
	if activePolicy == nil {
		logger.Info("No DBaaSPolicy found for the target namespace", "Namespace", req.Namespace)
		cond := metav1.Condition{
			Type:    v1beta1.DBaaSInventoryReadyType,
			Status:  metav1.ConditionFalse,
			Reason:  v1beta1.DBaaSPolicyNotFound,
			Message: v1beta1.MsgPolicyNotFound,
		}
		apimeta.SetStatusCondition(&inventory.Status.Conditions, cond)
		if err := r.Client.Status().Update(ctx, &inventory); err != nil {
			if errors.IsConflict(err) {
				logger.V(1).Info("DBaaS Inventory resource modified, retry syncing status", "DBaaS Inventory", inventory)
				return ctrl.Result{Requeue: true}, nil
			}
			logger.Error(err, "Error updating the DBaaS Inventory resource status", "DBaaS Inventory", inventory)
			metricLabelErrCdValue = metrics.LabelErrorCdValueErrorUpdatingInventoryStatus
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.checkCredsRefLabel(ctx, inventory); err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	defer func() {
		metrics.SetInventoryMetrics(inventory, execution, event, metricLabelErrCdValue)
	}()

	//
	// Provider Inventory
	//
	return r.reconcileProviderResource(ctx,
		inventory.Spec.ProviderRef.Name,
		&inventory,
		func(provider *v1beta1.DBaaSProvider) string {
			return provider.Spec.InventoryKind
		},
		func() interface{} {
			return inventory.Spec.DeepCopy()
		},
		func() interface{} {
			return &v1beta1.DBaaSProviderInventory{}
		},
		func(i interface{}) metav1.Condition {
			providerInv := i.(*v1beta1.DBaaSProviderInventory)
			return mergeInventoryStatus(&inventory, providerInv)
		},
		func() *[]metav1.Condition {
			return &inventory.Status.Conditions
		},
		v1beta1.DBaaSInventoryReadyType,
		logger,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSInventoryReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.DBaaSInventory{}).
		Watches(&source.Kind{Type: &v1beta1.DBaaSInventory{}}, &EventHandlerWithDelete{Controller: r}).
		WithOptions(
			controller.Options{MaxConcurrentReconciles: 2},
		).
		Build(r)
}

// mergeInventoryStatus: merge the status from DBaaSProviderInventory into the current DBaaSInventory status
func mergeInventoryStatus(inv *v1beta1.DBaaSInventory, providerInv *v1beta1.DBaaSProviderInventory) metav1.Condition {
	providerInv.Status.DeepCopyInto(&inv.Status)
	// Update inventory status condition (type: DBaaSInventoryReadyType) based on the provider status
	specSync := apimeta.FindStatusCondition(providerInv.Status.Conditions, v1beta1.DBaaSInventoryProviderSyncType)
	if specSync != nil && specSync.Status == metav1.ConditionTrue {
		return metav1.Condition{
			Type:    v1beta1.DBaaSInventoryReadyType,
			Status:  metav1.ConditionTrue,
			Reason:  v1beta1.Ready,
			Message: v1beta1.MsgProviderCRStatusSyncDone,
		}
	}
	return metav1.Condition{
		Type:    v1beta1.DBaaSInventoryReadyType,
		Status:  metav1.ConditionFalse,
		Reason:  v1beta1.ProviderReconcileInprogress,
		Message: v1beta1.MsgProviderCRReconcileInProgress,
	}
}

// Delete implements a handler for the Delete event.
func (r *DBaaSInventoryReconciler) Delete(e event.DeleteEvent) error {
	execution := metrics.PlatformInstallStart()
	metricLabelErrCdValue := ""
	log := ctrl.Log.WithName("DBaaSInventoryReconciler DeleteEvent")

	inventory, ok := e.Object.(*v1beta1.DBaaSInventory)

	if !ok {
		log.Info("Error getting inventory object during delete")
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrorDeletingInventory
		return nil
	}

	log.Info("inventoryObj", "inventoryObj", objectKeyFromObject(inventory))

	defer func() {
		log.Info("Calling metrics for deleting of DBaaSInventory")
		metrics.SetInventoryMetrics(*inventory, execution, metrics.LabelEventValueDelete, metricLabelErrCdValue)
	}()

	return nil
}
