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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DBaaSPolicyReconciler reconciles a DBaaSPolicy object
type DBaaSPolicyReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=resourcequotas,verbs=get;list;create;update;watch
//+kubebuilder:rbac:groups=core,resources=resourcequotas/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	policyList, err := r.policyListByNS(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "unable to list policies")
		return ctrl.Result{}, err
	}
	activePolicy := getActivePolicy(policyList)

	var policy v1alpha1.DBaaSPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued
			if len(policyList.Items) > 0 && activePolicy == nil {
				// reconcile another policy to ensure one is active
				policy = policyList.Items[0]
			} else {
				// child objects getting GC'd, no requeue
				return ctrl.Result{}, nil
			}
		} else {
			logger.Error(err, "Error fetching DBaaS Policy for reconcile")
			return ctrl.Result{}, err
		}
	}
	cond := &metav1.Condition{
		Type:    v1alpha1.DBaaSPolicyReadyType,
		Status:  metav1.ConditionTrue,
		Reason:  v1alpha1.Ready,
		Message: v1alpha1.MsgPolicyReady,
	}

	// if an active policy exists, and it's not this one... set status to false
	if activePolicy != nil &&
		activePolicy.GetName() != policy.Name {
		cond = &metav1.Condition{
			Type:    v1alpha1.DBaaSPolicyReadyType,
			Status:  metav1.ConditionFalse,
			Reason:  v1alpha1.DBaaSPolicyNotReady,
			Message: v1alpha1.MsgPolicyNotReady + " - " + activePolicy.GetName(),
		}
	}

	// if policy is active, create resourcequota
	if cond.Status == metav1.ConditionTrue {
		resQuota := v1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-" + policy.Name,
				Namespace: policy.Namespace,
			},
		}
		if res, err := controllerutil.CreateOrUpdate(ctx, r.Client, &resQuota, func() error {
			resQuota.Spec = v1.ResourceQuotaSpec{
				Hard: v1.ResourceList{
					v1.ResourceName("count/dbaaspolicies." + v1alpha1.GroupVersion.Group): resource.MustParse("1"),
				},
			}
			resQuota.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("ResourceQuota"))
			return ctrl.SetControllerReference(&policy, &resQuota, r.Scheme)
		}); err != nil {
			if errors.IsConflict(err) {
				logger.V(1).Info("ResourceQuota resource modified, retry syncing status", "ResourceQuota", resQuota)
				return ctrl.Result{Requeue: true}, err
			}
			logger.Error(err, "Error updating the ResourceQuota resource status", "ResourceQuota", resQuota)
			return ctrl.Result{}, err
		} else if res != controllerutil.OperationResultNone {
			logger.Info("ResourceQuota resource reconciled", "ResourceQuota", resQuota, "result", res)
		}
	}

	return r.updateStatusCondition(ctx, policy, cond)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSPolicy{}).
		Owns(&v1.ResourceQuota{}).
		Complete(r)
}

func (r *DBaaSPolicyReconciler) updateStatusCondition(ctx context.Context, policy v1alpha1.DBaaSPolicy, cond *metav1.Condition) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	apimeta.SetStatusCondition(&policy.Status.Conditions, *cond)
	if err := r.Client.Status().Update(ctx, &policy); err != nil {
		if errors.IsConflict(err) {
			logger.V(1).Info("DBaaS Policy resource modified, retry syncing status", "DBaaS Policy", policy)
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Error updating the DBaaS Policy resource status", "DBaaS Policy", policy)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// get active policy, return nil if none exists
func getActivePolicy(policyList v1alpha1.DBaaSPolicyList) *v1alpha1.DBaaSPolicy {
	for i := range policyList.Items {
		if apimeta.IsStatusConditionTrue(policyList.Items[i].Status.Conditions, v1alpha1.DBaaSPolicyReadyType) {
			return &policyList.Items[i]
		}
	}
	return nil
}
