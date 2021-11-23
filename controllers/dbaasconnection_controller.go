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

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
func (r *DBaaSConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, recErr error) {
	logger := ctrl.LoggerFrom(ctx, "DBaaS Connection", req.NamespacedName)

	var connection v1alpha1.DBaaSConnection
	if err := r.Get(ctx, req.NamespacedName, &connection); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.Info("DBaaS Connection resource not found, has been deleted")
			result, recErr = ctrl.Result{}, nil
			return
		}
		logger.Error(err, "Error fetching DBaaS Connection for reconcile")
		result, recErr = ctrl.Result{}, err
		return
	}

	var dbaasCond metav1.Condition
	// This update will make sure the status is always updated in case of any errors or successful result
	defer func(conn *v1alpha1.DBaaSConnection, cond *metav1.Condition) {
		apimeta.SetStatusCondition(&conn.Status.Conditions, *cond)
		if err := r.Client.Status().Update(ctx, conn); err != nil {
			if errors.IsConflict(err) {
				logger.V(1).Info("Connection modified, retry syncing spec")
				// Re-queue and preserve existing recErr
				result = ctrl.Result{Requeue: true}
				return
			}
			logger.Error(err, "Could not update connection status")
			if recErr == nil {
				// There is no existing recErr. Set it to the status update error
				recErr = err
			}
		}
	}(&connection, &dbaasCond)

	if res, err := r.reconcileDevTopologyResource(&connection, ctx); err != nil {
		if errors.IsConflict(err) {
			logger.Info("Deployment for Developer Topology view modified, retry reconciling")
			result, recErr = ctrl.Result{Requeue: true}, nil
			return
		}
		logger.Error(err, "Error reconciling Deployment for Developer Topology view")
		result, recErr = ctrl.Result{}, err
		return
	} else {
		logger.Info("Deployment for Developer Topology view reconciled", "result", res)
	}

	var inventory v1alpha1.DBaaSInventory
	if err := r.Get(ctx, types.NamespacedName{Namespace: connection.Spec.InventoryRef.Namespace, Name: connection.Spec.InventoryRef.Name}, &inventory); err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "DBaaS Inventory resource not found", "DBaaS Inventory", connection.Spec.InventoryRef)
			dbaasCond = metav1.Condition{Type: v1alpha1.DBaaSConnectionReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.DBaaSInventoryNotFound, Message: err.Error()}
			result, recErr = ctrl.Result{}, err
			return
		}
		logger.Error(err, "Error fetching DBaaS Inventory resource reference for DBaaS Connection", "DBaaS Inventory", connection.Spec.InventoryRef)
		result, recErr = ctrl.Result{}, err
		return
	}

	provider, err := r.getDBaaSProvider(inventory.Spec.ProviderRef.Name, ctx)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Requested DBaaS Provider is not configured in this environment", "DBaaS Provider", inventory.Spec.ProviderRef)
			dbaasCond = metav1.Condition{Type: v1alpha1.DBaaSConnectionReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.DBaaSProviderNotFound, Message: err.Error()}
			result, recErr = ctrl.Result{}, err
			return
		}
		logger.Error(err, "Error reading configured DBaaS Provider", "DBaaS Provider", inventory.Spec.ProviderRef)
		result, recErr = ctrl.Result{}, err
		return
	}
	logger.Info("Found DBaaS Provider", "DBaaS Provider", inventory.Spec.ProviderRef)

	// The inventory must be in ready status before we can move on
	invCond := apimeta.FindStatusCondition(inventory.Status.Conditions, v1alpha1.DBaaSInventoryReadyType)
	if invCond == nil || invCond.Status == metav1.ConditionFalse {
		err := fmt.Errorf(v1alpha1.MsgInventoryNotReady)
		logger.Error(err, "Inventory is not ready", "Inventory", inventory.Name, "Namespace", inventory.Namespace)
		dbaasCond = metav1.Condition{Type: v1alpha1.DBaaSConnectionReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.DBaaSInventoryNotReady, Message: v1alpha1.MsgInventoryNotReady}
		result, recErr = ctrl.Result{}, err
		return
	}

	providerConnection := r.createProviderObject(&connection, provider.Spec.ConnectionKind)
	if res, err := r.reconcileProviderObject(providerConnection, r.providerObjectMutateFn(&connection, providerConnection, connection.Spec.DeepCopy()), ctx); err != nil {
		if errors.IsConflict(err) {
			logger.Info("Provider Connection modified, retry syncing spec")
			result, recErr = ctrl.Result{Requeue: true}, nil
			return
		}
		logger.Error(err, "Error reconciling Provider Connection resource")
		result, recErr = ctrl.Result{}, err
		return
	} else {
		logger.Info("Provider Connection resource reconciled", "result", res)
	}

	var DBaaSProviderConnection v1alpha1.DBaaSProviderConnection
	if err := r.parseProviderObject(providerConnection, &DBaaSProviderConnection); err != nil {
		logger.Error(err, "Error parsing the Provider Connection resource")
		dbaasCond = metav1.Condition{Type: v1alpha1.DBaaSConnectionReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.ProviderParsingError, Message: err.Error()}
		result, recErr = ctrl.Result{}, err
		return
	}
	dbaasCond = *mergeConnectionStatus(&connection, &DBaaSProviderConnection)
	result, recErr = ctrl.Result{}, nil
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSConnectionReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSConnection{}).
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
func mergeConnectionStatus(conn *v1alpha1.DBaaSConnection, providerConn *v1alpha1.DBaaSProviderConnection) *metav1.Condition {

	cond := apimeta.FindStatusCondition(conn.Status.Conditions, v1alpha1.DBaaSConnectionReadyType)
	providerConn.Status.DeepCopyInto(&conn.Status)
	if cond != nil {
		conn.Status.Conditions = append(conn.Status.Conditions, *cond)
	}
	// Update connection status condition (type: DBaaSConnectionReadyType) based on the provider status
	specSync := apimeta.FindStatusCondition(providerConn.Status.Conditions, v1alpha1.DBaaSConnectionProviderSyncType)
	if cond != nil && specSync != nil && specSync.Status == metav1.ConditionTrue {
		return &metav1.Condition{Type: v1alpha1.DBaaSConnectionReadyType, Status: metav1.ConditionTrue, Reason: v1alpha1.Ready, Message: v1alpha1.MsgProviderCRStatusSyncDone}
	}
	return &metav1.Condition{Type: v1alpha1.DBaaSConnectionReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.ProviderReconcileInprogress, Message: v1alpha1.MsgProviderCRReconcileInProgress}
}
