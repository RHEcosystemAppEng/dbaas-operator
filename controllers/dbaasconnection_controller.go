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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
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
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update

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

	if result, err := r.reconcileDevTopologyResource(&connection, ctx); err != nil {
		logger.Error(err, "Error reconciling Deployment for Developer Topology view")
		return ctrl.Result{}, err
	} else {
		logger.Info("Deployment for Developer Topology view reconciled", "result", result)
	}

	var inventory v1alpha1.DBaaSInventory
	if err := r.Get(ctx, types.NamespacedName{Namespace: connection.Spec.InventoryRef.Namespace, Name: connection.Spec.InventoryRef.Name}, &inventory); err != nil {
		logger.Error(err, "Error fetching DBaaS Inventory resource reference for DBaaS Connection", "DBaaS Inventory", connection.Spec.InventoryRef)
		return ctrl.Result{}, err
	}

	provider, err := r.getDBaaSProvider(inventory.Spec.Provider, ctx)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Requested DBaaS Provider is not configured in this environment", "DBaaS Provider", inventory.Spec.Provider)
			return ctrl.Result{}, err
		}
		logger.Error(err, "Error reading configured DBaaS Provider", "DBaaS Provider", inventory.Spec.Provider)
		return ctrl.Result{}, err
	}
	logger.Info("Found DBaaS Provider", "DBaaS Provider", provider.Provider)

	providerConnection := r.createProviderObject(&connection, provider.ConnectionKind)
	if result, err := r.reconcileProviderObject(providerConnection, r.providerObjectMutateFn(&connection, providerConnection, connection.Spec.DeepCopy()), ctx); err != nil {
		if errors.IsConflict(err) {
			logger.Info("Provider Connection modified, retry syncing spec")
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Error reconciling Provider Connection resource")
		return ctrl.Result{}, err
	} else {
		logger.Info("Provider Connection resource reconciled", "result", result)
	}

	var status v1alpha1.DBaaSConnectionStatus
	if exist, err := r.parseProviderObjectStatus(providerConnection, &status); err != nil {
		logger.Error(err, "Error parsing the status of the Provider Connection resource")
		return ctrl.Result{}, err
	} else if exist {
		err = r.reconcileDBaaSObjectStatus(&connection, ctx, func() error {
			status.DeepCopyInto(&connection.Status)
			return nil
		})
		if err != nil {
			if errors.IsConflict(err) {
				logger.Info("DBaaS Connection modified, retry syncing status")
				return ctrl.Result{Requeue: true}, nil
			}
			logger.Error(err, "Error updating the DBaaS Connection status")
			return ctrl.Result{}, err
		}
		logger.Info("DBaaS Connection status updated")
	} else {
		logger.Info("Provider Connection resource status not found")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSConnectionReconciler) SetupWithManager(mgr ctrl.Manager, providerList v1alpha1.DBaaSProviderList) error {
	owned := r.parseDBaaSProviderConnections(providerList)
	builder := ctrl.NewControllerManagedBy(mgr)
	builder = builder.For(&v1alpha1.DBaaSConnection{})
	builder.Owns(&appv1.Deployment{})
	for _, o := range owned {
		builder = builder.Owns(o)
	}
	return builder.Complete(r)
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
