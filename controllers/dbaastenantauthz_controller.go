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

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// DBaaSTenantAuthzReconciler reconciles Rbac
type DBaaSTenantAuthzReconciler struct {
	*DBaaSAuthzReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles/finalizers;clusterrolebindings/finalizers,verbs=update
//+kubebuilder:rbac:groups="";authorization.openshift.io,resources=localresourceaccessreviews;localsubjectaccessreviews;resourceaccessreviews;selfsubjectrulesreviews;subjectaccessreviews;subjectrulesreviews,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSTenantAuthzReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	if err := r.reconcileAuthz(ctx, req.Namespace); err != nil {
		if errors.IsConflict(err) {
			logger.Info("Requeued due to update conflict")
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSTenantAuthzReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := ctrl.NewControllerManagedBy(mgr).
		Named("dbaasauthz").
		// set tenant as main object watched by controller, but ignore all events
		// ... we handle tenant events through a map function further down
		For(&v1alpha1.DBaaSTenant{}, builder.WithPredicates(ignoreAllEvents)).
		// all inventory events should trigger full authz reconcile
		Watches(
			&source.Kind{Type: &v1alpha1.DBaaSInventory{}},
			handler.EnqueueRequestsFromMapFunc(nsMapFunc),
		).
		// all role events should trigger full authz reconcile
		Watches(
			&source.Kind{Type: &rbacv1.Role{}},
			handler.EnqueueRequestsFromMapFunc(nsMapFunc),
		).
		// all rolebinding events should trigger full authz reconcile
		Watches(
			&source.Kind{Type: &rbacv1.RoleBinding{}},
			handler.EnqueueRequestsFromMapFunc(nsMapFunc),
			// for most rolebindings, to reduce memory footprint, only cache metadata
			builder.OnlyMetadata,
		).
		// all tenant events should trigger full authz reconcile
		// ... tenant events are transformed to pass inventory namespace in request
		Watches(
			&source.Kind{Type: &v1alpha1.DBaaSTenant{}},
			handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
				namespace := o.(*v1alpha1.DBaaSTenant).Spec.InventoryNamespace
				return getRequest(namespace)
			}),
		).
		WithOptions(
			controller.Options{MaxConcurrentReconciles: 6},
		).
		Complete(r); err != nil {
		return err
	}

	// index tenants by `spec.inventoryNamespace`
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1alpha1.DBaaSTenant{}, inventoryNamespaceKey, func(rawObj client.Object) []string {
		tenant := rawObj.(*v1alpha1.DBaaSTenant)
		inventoryNS := tenant.Spec.InventoryNamespace
		return []string{inventoryNS}
	}); err != nil {
		return err
	}

	return nil
}

// changes name to match an object's namespace. ensures a namespace is only queued once,
// instead of once for each object in a namespace
var nsMapFunc = func(o client.Object) []reconcile.Request {
	namespace := o.GetNamespace()
	return getRequest(namespace)
}

func getRequest(nameNS string) []reconcile.Request {
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      nameNS,
				Namespace: nameNS,
			},
		},
	}
}

// Reconcile all tenant and inventory RBAC
func (r *DBaaSTenantAuthzReconciler) reconcileAuthz(ctx context.Context, namespace string) (err error) {
	logger := ctrl.LoggerFrom(ctx)

	tenantList, err := r.tenantListByInventoryNS(ctx, namespace)
	if err != nil {
		logger.Error(err, "unable to list tenants")
		return err
	}

	// continue only if the request is in a valid tenant namespace
	if len(tenantList.Items) > 0 {

		// Get list of DBaaSInventories from tenant namespace
		var inventoryList v1alpha1.DBaaSInventoryList
		if err := r.List(ctx, &inventoryList, &client.ListOptions{Namespace: namespace}); err != nil {
			logger.Error(err, "Error fetching DBaaS Inventory List for reconcile")
			return err
		}

		//
		// Tenant RBAC
		//
		serviceAdminAuthz := r.getServiceAdminAuthz(ctx, namespace)
		developerAuthz := r.getDeveloperAuthz(ctx, namespace, inventoryList)
		tenantListAuthz := r.getTenantListAuthz(ctx)
		for _, tenant := range tenantList.Items {
			if err := r.reconcileTenantRbacObjs(ctx, tenant, serviceAdminAuthz, developerAuthz, tenantListAuthz); err != nil {
				return err
			}
		}
	}

	return nil
}
