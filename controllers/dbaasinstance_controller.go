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

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	metrics "github.com/RHEcosystemAppEng/dbaas-operator/controllers/metrics"
)

// DBaaSInstanceReconciler reconciles a DBaaSInstance object
type DBaaSInstanceReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *DBaaSInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	execution := metrics.PlatformInstallStart()
	logger := ctrl.LoggerFrom(ctx)
	var instance v1beta1.DBaaSInstance
	metricLabelErrCdValue := ""
	event := ""

	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.V(1).Info("DBaaS Instance resource not found, has been deleted")
			metricLabelErrCdValue = metrics.LabelErrorCdValueResourceNotFound
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaS Instance for reconcile")
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrorFetchingDBaaSInstance
		return ctrl.Result{}, err
	}

	if instance.DeletionTimestamp != nil {
		event = metrics.LabelEventValueDelete
	} else {
		event = metrics.LabelEventValueCreate
	}

	if inventory, validNS, provision, err := r.checkInventory(ctx, instance.Spec.InventoryRef, &instance, func(reason string, message string) {
		cond := metav1.Condition{
			Type:    v1beta1.DBaaSInstanceReadyType,
			Status:  metav1.ConditionFalse,
			Reason:  reason,
			Message: message,
		}
		apimeta.SetStatusCondition(&instance.Status.Conditions, cond)
		instance.Status.Phase = v1beta1.InstancePhaseError
	}, logger); err != nil {
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrorCheckingInstanceInventory
		return ctrl.Result{}, err
	} else if !validNS {
		return ctrl.Result{}, nil
	} else if !provision {
		return ctrl.Result{}, nil
	} else {
		result, err := r.reconcileProviderResource(ctx,
			inventory.Spec.ProviderRef.Name,
			&instance,
			func(provider *v1beta1.DBaaSProvider) string {
				return provider.Spec.InstanceKind
			},
			func() interface{} {
				return instance.Spec.DeepCopy()
			},
			func() interface{} {
				return &v1beta1.DBaaSProviderInstance{}
			},
			func(i interface{}) metav1.Condition {
				providerInstance := i.(*v1beta1.DBaaSProviderInstance)
				return mergeInstanceStatus(&instance, providerInstance)
			},
			func() *[]metav1.Condition {
				return &instance.Status.Conditions
			},
			v1beta1.DBaaSInstanceReadyType,
			logger,
		)

		defer func() {
			metrics.SetInstanceMetrics(inventory.Spec.ProviderRef.Name, inventory.Name, instance, execution, event, metricLabelErrCdValue)
		}()
		return result, err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSInstanceReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.DBaaSInstance{}).
		Watches(&source.Kind{Type: &v1beta1.DBaaSInstance{}}, &EventHandlerWithDelete{Controller: r}).
		WithOptions(
			controller.Options{MaxConcurrentReconciles: 2},
		).
		Build(r)
}

// mergeInstanceStatus: merge the status from DBaaSProviderInstance into the current DBaaSInstance status
func mergeInstanceStatus(instance *v1beta1.DBaaSInstance, providerInst *v1beta1.DBaaSProviderInstance) metav1.Condition {
	providerInst.Status.DeepCopyInto(&instance.Status)
	if len(instance.Status.Phase) == 0 {
		instance.Status.Phase = v1beta1.InstancePhaseUnknown
	}
	// Update instance status condition (type: DBaaSInstanceReadyType) based on the provider status
	specSync := apimeta.FindStatusCondition(providerInst.Status.Conditions, v1beta1.DBaaSInstanceProviderSyncType)
	if specSync != nil && specSync.Status == metav1.ConditionTrue {
		return metav1.Condition{
			Type:    v1beta1.DBaaSInstanceReadyType,
			Status:  metav1.ConditionTrue,
			Reason:  v1beta1.Ready,
			Message: v1beta1.MsgProviderCRStatusSyncDone,
		}
	}
	return metav1.Condition{
		Type:    v1beta1.DBaaSInstanceReadyType,
		Status:  metav1.ConditionFalse,
		Reason:  v1beta1.ProviderReconcileInprogress,
		Message: v1beta1.MsgProviderCRReconcileInProgress,
	}
}

// Delete implements a handler for the Delete event.
func (r *DBaaSInstanceReconciler) Delete(e event.DeleteEvent) error {
	execution := metrics.PlatformInstallStart()
	metricLabelErrCdValue := ""
	log := ctrl.Log.WithName("DBaaSInstanceReconciler DeleteEvent")
	log.Info("Delete event started")

	instanceObj, ok := e.Object.(*v1beta1.DBaaSInstance)
	if !ok {
		log.Error(nil, "Ignoring malformed Delete()", "Object", e.Object)
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrorDeletingInstance
		return nil
	}
	log.Info("instanceObj", "instanceObj", objectKeyFromObject(instanceObj))

	inventory := &v1beta1.DBaaSInventory{}
	_ = r.Get(context.TODO(), types.NamespacedName{Namespace: instanceObj.Spec.InventoryRef.Namespace, Name: instanceObj.Spec.InventoryRef.Name}, inventory)

	defer func() {
		log.Info("Calling metrics for deleting of DBaaSInstance")
		metrics.SetInstanceMetrics(inventory.Spec.ProviderRef.Name, inventory.Name, *instanceObj, execution, metrics.LabelEventValueDelete, metricLabelErrCdValue)
	}()

	return nil
}
