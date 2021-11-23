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

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// DBaaSInventoryReconciler reconciles a DBaaSInventory object
type DBaaSInventoryReconciler struct {
	*DBaaSTenantReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSInventoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, recErr error) {
	logger := ctrl.LoggerFrom(ctx, "DBaaS Inventory", req.NamespacedName)
	var inventory v1alpha1.DBaaSInventory
	if err := r.Get(ctx, req.NamespacedName, &inventory); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.V(1).Info("DBaaS Inventory resource not found, has been deleted")
			result, recErr = ctrl.Result{}, nil
			return
		}
		logger.Error(err, "Error fetching DBaaS Inventory for reconcile")
		result, recErr = ctrl.Result{}, err
		return
	}

	var dbaasCond metav1.Condition
	// This update will make sure the status is always updated in case of any errors or successful result
	defer func(inv *v1alpha1.DBaaSInventory, cond *metav1.Condition) {
		apimeta.SetStatusCondition(&inv.Status.Conditions, *cond)
		if err := r.Client.Status().Update(ctx, inv); err != nil {
			if errors.IsConflict(err) {
				logger.V(1).Info("Inventory modified, retry syncing spec")
				// Re-queue and preserve existing recErr
				result = ctrl.Result{Requeue: true}
				return
			}
			logger.Error(err, "Could not update inventory status")
			if recErr == nil {
				// There is no existing recErr. Set it to the status update error
				recErr = err
			}
		}
	}(&inventory, &dbaasCond)

	tenantList, err := r.tenantListByInventoryNS(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "unable to list tenants")
		result, recErr = ctrl.Result{}, err
		return
	}

	if len(tenantList.Items) == 0 {
		logger.Info("No DBaaS tenant found for the target namespace", "Namespace", req.Namespace)
		dbaasCond = metav1.Condition{Type: v1alpha1.DBaaSInventoryReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.DBaaSTenantNotFound, Message: v1alpha1.MsgTenantNotFound}
		result, recErr = ctrl.Result{}, nil
		return
	}

	//
	// Provider Inventory
	//
	provider, err := r.getDBaaSProvider(inventory.Spec.ProviderRef.Name, ctx)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Requested DBaaS Provider is not configured in this environment", "DBaaS Provider", inventory.Spec.ProviderRef)
			dbaasCond = metav1.Condition{Type: v1alpha1.DBaaSInventoryReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.DBaaSProviderNotFound, Message: err.Error()}
			result, recErr = ctrl.Result{}, err
			return
		}
		logger.Error(err, "Error reading configured DBaaS Provider", "DBaaS Provider", inventory.Spec.ProviderRef)
		result, recErr = ctrl.Result{}, err
		return
	}
	logger.V(1).Info("Found DBaaS Provider", "DBaaS Provider", inventory.Spec.ProviderRef)

	providerInventory := r.createProviderObject(&inventory, provider.Spec.InventoryKind)
	if res, err := r.reconcileProviderObject(providerInventory, r.providerObjectMutateFn(&inventory, providerInventory, inventory.Spec.DeepCopy()), ctx); err != nil {
		if errors.IsConflict(err) {
			logger.V(1).Info("Provider Inventory modified, retry syncing spec")
			result, recErr = ctrl.Result{Requeue: true}, nil
			return
		}
		logger.Error(err, "Error reconciling the Provider Inventory resource")
		result, recErr = ctrl.Result{}, err
		return
	} else {
		logger.V(1).Info("Provider Inventory resource reconciled", "result", res)
	}

	var DBaaSProviderInventory v1alpha1.DBaaSProviderInventory
	if err := r.parseProviderObject(providerInventory, &DBaaSProviderInventory); err != nil {
		logger.Error(err, "Error parsing the Provider Inventory resource")
		dbaasCond = metav1.Condition{Type: v1alpha1.DBaaSInventoryReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.ProviderParsingError, Message: err.Error()}
		result, recErr = ctrl.Result{}, err
		return
	}

	dbaasCond = *mergeInventoryStatus(&inventory, &DBaaSProviderInventory)
	result, recErr = ctrl.Result{}, nil
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSInventoryReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSInventory{}).
		Build(r)
}

// mergeInventoryStatus: merge the status from DBaaSProviderInventory into the current DBaaSInventory status
func mergeInventoryStatus(inv *v1alpha1.DBaaSInventory, providerInv *v1alpha1.DBaaSProviderInventory) *metav1.Condition {
	cond := apimeta.FindStatusCondition(inv.Status.Conditions, v1alpha1.DBaaSInventoryReadyType)
	providerInv.Status.DeepCopyInto(&inv.Status)
	if cond != nil {
		inv.Status.Conditions = append(inv.Status.Conditions, *cond)
	}
	// Update inventory status condition (type: DBaaSInventoryReadyType) based on the provider status
	specSync := apimeta.FindStatusCondition(providerInv.Status.Conditions, v1alpha1.DBaaSInventoryProviderSyncType)
	if cond != nil && specSync != nil && specSync.Status == metav1.ConditionTrue {
		return &metav1.Condition{Type: v1alpha1.DBaaSInventoryReadyType, Status: metav1.ConditionTrue, Reason: v1alpha1.Ready, Message: v1alpha1.MsgProviderCRStatusSyncDone}
	}
	return &metav1.Condition{Type: v1alpha1.DBaaSInventoryReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.ProviderReconcileInprogress, Message: v1alpha1.MsgProviderCRReconcileInProgress}
}
