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
	"reflect"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	metrics "github.com/RHEcosystemAppEng/dbaas-operator/controllers/metrics"
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
	metricLabelErrCdValue := ""
	event := ""

	var connection v1beta1.DBaaSConnection
	execution := metrics.PlatformInstallStart()
	if err := r.Get(ctx, req.NamespacedName, &connection); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.V(1).Info("DBaaS Connection resource not found, has been deleted")
			metricLabelErrCdValue = metrics.LabelErrorCdValueResourceNotFound
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaS Connection for reconcile")
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrorFetchingDBaaSInventoryResources
		return ctrl.Result{}, err
	}

	if connection.DeletionTimestamp != nil {
		event = metrics.LabelEventValueDelete
	} else {
		event = metrics.LabelEventValueCreate
	}

	res, err := r.reconcileDevTopologyResource(ctx, &connection)
	if err != nil {
		if errors.IsConflict(err) {
			logger.V(1).Info("Deployment for Developer Topology view modified, retry reconciling")
			metricLabelErrCdValue = metrics.LabelErrorCdValueDevTopologyModified
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Error reconciling Deployment for Developer Topology view")
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrReconcilingWithDevTopology
		return ctrl.Result{}, err
	}
	logger.Info("Deployment for Developer Topology view reconciled", "result", res)

	if inventory, validNS, _, err := r.checkInventory(ctx, connection.Spec.InventoryRef, &connection, func(reason string, message string) {
		cond := metav1.Condition{
			Type:    v1beta1.DBaaSConnectionReadyType,
			Status:  metav1.ConditionFalse,
			Reason:  reason,
			Message: message,
		}
		apimeta.SetStatusCondition(&connection.Status.Conditions, cond)
	}, logger); err != nil {
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrCheckingInventory
		return ctrl.Result{}, err
	} else if !validNS {
		metricLabelErrCdValue = metrics.LabelErrorCdValueInvalidNameSpace
		return ctrl.Result{}, nil
	} else {
		spec, err := r.getConnectionSpec(ctx, connection.Spec.DeepCopy())
		if err != nil {
			logger.Error(err, "Cannot read the instance reference")
			cond := metav1.Condition{
				Type:    v1beta1.DBaaSConnectionReadyType,
				Status:  metav1.ConditionFalse,
				Reason:  v1beta1.DBaaSInstanceNotAvailable,
				Message: err.Error(),
			}
			r.updateConnectionStatus(ctx, &connection, &cond)
			metricLabelErrCdValue = metrics.LabelErrorCdCannotReadInstance
			return ctrl.Result{}, err
		}
		result, err := r.reconcileProviderResource(ctx,
			inventory.Spec.ProviderRef.Name,
			&connection,
			func(provider *v1beta1.DBaaSProvider) string {
				return provider.Spec.ConnectionKind
			},
			func() interface{} {
				return spec
			},
			func() interface{} {
				return &v1beta1.DBaaSProviderConnection{}
			},
			func(i interface{}) metav1.Condition {
				providerConn := i.(*v1beta1.DBaaSProviderConnection)
				return mergeConnectionStatus(&connection, providerConn)
			},
			func() *[]metav1.Condition {
				return &connection.Status.Conditions
			},
			v1beta1.DBaaSConnectionReadyType,
			logger,
		)
		defer func() {
			metrics.SetConnectionMetrics(inventory.Spec.ProviderRef.Name, inventory.Name, connection, execution, event, metricLabelErrCdValue)
		}()
		return result, err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSConnectionReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.DBaaSConnection{}).
		Watches(&source.Kind{Type: &v1beta1.DBaaSConnection{}}, &EventHandlerWithDelete{Controller: r}).
		WithOptions(
			controller.Options{MaxConcurrentReconciles: 2},
		).
		Build(r)
}

func (r *DBaaSConnectionReconciler) reconcileDevTopologyResource(ctx context.Context, connection *v1beta1.DBaaSConnection) (controllerutil.OperationResult, error) {
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connection.Name,
			Namespace: connection.Namespace,
		},
	}
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, r.deploymentMutateFn(connection, deployment))
	return result, err
}

func (r *DBaaSConnectionReconciler) deploymentMutateFn(connection *v1beta1.DBaaSConnection, deployment *appv1.Deployment) controllerutil.MutateFn {
	return func() error {
		if deployment.ObjectMeta.Annotations == nil {
			deployment.ObjectMeta.Annotations = make(map[string]string, 4)
		}
		deployment.ObjectMeta.Annotations["managed-by"] = "dbaas-operator"
		deployment.ObjectMeta.Annotations["owner"] = connection.Name
		deployment.ObjectMeta.Annotations["owner.kind"] = connection.Kind
		deployment.ObjectMeta.Annotations["owner.namespace"] = connection.Namespace

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
		return ctrl.SetControllerReference(connection, deployment, r.Scheme)
	}
}

// mergeConnectionStatus: merge the status from DBaaSProviderConnection into the current DBaaSConnection status
func mergeConnectionStatus(conn *v1beta1.DBaaSConnection, providerConn *v1beta1.DBaaSProviderConnection) metav1.Condition {
	providerConn.Status.DeepCopyInto(&conn.Status)
	// Update connection status condition (type: DBaaSConnectionReadyType) based on the provider status
	specSync := apimeta.FindStatusCondition(providerConn.Status.Conditions, v1beta1.DBaaSConnectionProviderSyncType)
	if specSync != nil && specSync.Status == metav1.ConditionTrue {
		return metav1.Condition{
			Type:    v1beta1.DBaaSConnectionReadyType,
			Status:  metav1.ConditionTrue,
			Reason:  v1beta1.Ready,
			Message: v1beta1.MsgProviderCRStatusSyncDone,
		}
	}
	return metav1.Condition{
		Type:    v1beta1.DBaaSConnectionReadyType,
		Status:  metav1.ConditionFalse,
		Reason:  v1beta1.ProviderReconcileInprogress,
		Message: v1beta1.MsgProviderCRReconcileInProgress,
	}
}

// Delete implements a handler for the Delete event.
func (r *DBaaSConnectionReconciler) Delete(e event.DeleteEvent) error {
	execution := metrics.PlatformInstallStart()
	metricLabelErrCdValue := ""
	log := ctrl.Log.WithName("DBaaSConnectionReconciler DeleteEvent")
	log.Info("Delete event started")

	connectionObj, ok := e.Object.(*v1beta1.DBaaSConnection)
	if !ok {
		log.Info("Error getting connection object during delete")
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrorDeletingConnection
		return nil
	}
	log.Info("connectionObj", "connectionObj", objectKeyFromObject(connectionObj))

	inventory := &v1beta1.DBaaSInventory{}
	_ = r.Get(context.TODO(), types.NamespacedName{Namespace: connectionObj.Spec.InventoryRef.Namespace, Name: connectionObj.Spec.InventoryRef.Name}, inventory)

	log.Info("Calling metrics for deleting of DBaaSConnection")
	metrics.SetConnectionMetrics(inventory.Spec.ProviderRef.Name, inventory.Name, *connectionObj, execution, metrics.LabelEventValueDelete, metricLabelErrCdValue)

	return nil

}

func (r *DBaaSConnectionReconciler) getConnectionSpec(ctx context.Context, spec *v1beta1.DBaaSConnectionSpec) (interface{}, error) {
	if len(spec.InstanceID) > 0 {
		return spec, nil
	}

	instanceRef := spec.InstanceRef

	if instanceRef == nil || len(instanceRef.Name) == 0 {
		return nil, fmt.Errorf("instance reference is not properly set")
	}

	instance := &v1beta1.DBaaSInstance{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      instanceRef.Name,
		Namespace: instanceRef.Namespace,
	}, instance); err != nil {
		return nil, fmt.Errorf("cannot read the instance reference")
	}

	if !reflect.DeepEqual(instance.Spec.InventoryRef, spec.InventoryRef) {
		return nil, fmt.Errorf("instance and connection don't use the same inventory reference")
	}

	if len(instance.Status.InstanceID) == 0 {
		return nil, fmt.Errorf("instance ID is not available")
	}

	spec.InstanceID = instance.Status.InstanceID
	spec.InstanceRef = nil
	return spec, nil
}

func (r *DBaaSConnectionReconciler) updateConnectionStatus(ctx context.Context, connection *v1beta1.DBaaSConnection, cond *metav1.Condition) {
	apimeta.SetStatusCondition(&connection.Status.Conditions, *cond)
	logger := ctrl.LoggerFrom(ctx)
	if err := r.Client.Status().Update(ctx, connection); err != nil {
		if errors.IsConflict(err) {
			logger.V(1).Info("DBaaSConnection Object modified", "DBaaSConnection Object", connection)
		} else {
			logger.Error(err, "Error updating the DBaaSConnection Object status", "DBaaSConnection Object", connection)
		}
	}
}
