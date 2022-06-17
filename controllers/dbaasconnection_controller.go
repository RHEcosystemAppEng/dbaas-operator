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

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

// DBaaSConnectionReconciler reconciles a DBaaSConnection object
type DBaaSConnectionReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	var connection v1alpha1.DBaaSConnection
	if err := r.Get(ctx, req.NamespacedName, &connection); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.V(1).Info("DBaaS Connection resource not found, has been deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaS Connection for reconcile")
		return ctrl.Result{}, err
	}

	if res, err := r.reconcileDevTopologyResource(&connection, ctx); err != nil {
		if errors.IsConflict(err) {
			logger.V(1).Info("Deployment for Developer Topology view modified, retry reconciling")
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Error reconciling Deployment for Developer Topology view")
		return ctrl.Result{}, err
	} else {
		logger.Info("Deployment for Developer Topology view reconciled", "result", res)
	}

	if inventory, validNS, _, err := r.checkInventory(connection.Spec.InventoryRef, &connection, func(reason string, message string) {
		cond := metav1.Condition{
			Type:    v1alpha1.DBaaSConnectionReadyType,
			Status:  metav1.ConditionFalse,
			Reason:  reason,
			Message: message,
		}
		apimeta.SetStatusCondition(&connection.Status.Conditions, cond)
	}, ctx, logger); err != nil {
		return ctrl.Result{}, err
	} else if !validNS {
		return ctrl.Result{}, nil
	} else {
		return r.reconcileProviderResource(inventory.Spec.ProviderRef.Name,
			&connection,
			func(provider *v1alpha1.DBaaSProvider) string {
				return provider.Spec.ConnectionKind
			},
			func() interface{} {
				return connection.Spec.DeepCopy()
			},
			func() interface{} {
				return &v1alpha1.DBaaSProviderConnection{}
			},
			func(i interface{}) metav1.Condition {
				providerConn := i.(*v1alpha1.DBaaSProviderConnection)
				return mergeConnectionStatus(&connection, providerConn)
			},
			func() *[]metav1.Condition {
				return &connection.Status.Conditions
			},
			v1alpha1.DBaaSConnectionReadyType,
			ctx,
			logger,
		)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSConnectionReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSConnection{}).
		WithOptions(
			controller.Options{MaxConcurrentReconciles: 2},
		).
		Build(r)
}

func (r *DBaaSConnectionReconciler) reconcileDevTopologyResource(connection *v1alpha1.DBaaSConnection, ctx context.Context) (controllerutil.OperationResult, error) {
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connection.Name,
			Namespace: connection.Namespace,
		},
	}
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, r.deploymentMutateFn(connection, deployment))
	return result, err
}

func (r *DBaaSConnectionReconciler) deploymentMutateFn(connection *v1alpha1.DBaaSConnection, deployment *appv1.Deployment) controllerutil.MutateFn {
	return func() error {
		deployment.ObjectMeta.Labels = map[string]string{
			"managed-by":      "dbaas-operator",
			"owner":           connection.Name,
			"owner.kind":      connection.Kind,
			"owner.namespace": connection.Namespace,
		}
		deployment.Spec = appv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(0),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "bind-deploy",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "bind-deploy",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "bind-deploy",
							Image:           "quay.io/ecosystem-appeng/busybox",
							ImagePullPolicy: v1.PullIfNotPresent,
							Command:         []string{"sh", "-c", "echo The app is running! && sleep 3600"},
						},
					},
				},
			},
		}
		deployment.OwnerReferences = nil
		if err := ctrl.SetControllerReference(connection, deployment, r.Scheme); err != nil {
			return err
		}
		return nil
	}
}

// mergeConnectionStatus: merge the status from DBaaSProviderConnection into the current DBaaSConnection status
func mergeConnectionStatus(conn *v1alpha1.DBaaSConnection, providerConn *v1alpha1.DBaaSProviderConnection) metav1.Condition {
	providerConn.Status.DeepCopyInto(&conn.Status)
	// Update connection status condition (type: DBaaSConnectionReadyType) based on the provider status
	specSync := apimeta.FindStatusCondition(providerConn.Status.Conditions, v1alpha1.DBaaSConnectionProviderSyncType)
	if specSync != nil && specSync.Status == metav1.ConditionTrue {
		return metav1.Condition{
			Type:    v1alpha1.DBaaSConnectionReadyType,
			Status:  metav1.ConditionTrue,
			Reason:  v1alpha1.Ready,
			Message: v1alpha1.MsgProviderCRStatusSyncDone,
		}
	}
	return metav1.Condition{
		Type:    v1alpha1.DBaaSConnectionReadyType,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.ProviderReconcileInprogress,
		Message: v1alpha1.MsgProviderCRReconcileInProgress,
	}
}
