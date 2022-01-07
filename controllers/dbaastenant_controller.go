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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// DBaaSTenantReconciler reconciles a DBaaSTenant object
type DBaaSTenantReconciler struct {
	*DBaaSAuthzReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	var tenant v1alpha1.DBaaSTenant
	if err := r.Get(ctx, req.NamespacedName, &tenant); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaS Tenant for reconcile")
		return ctrl.Result{}, err
	}

	// Reconcile tenant related RBAC
	if err := r.reconcileAuthz(ctx, tenant.Spec.InventoryNamespace); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSTenant{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		WithOptions(
			controller.Options{
				MaxConcurrentReconciles: 5,
			},
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

// get tenant names from list
func getTenantNames(tenantList v1alpha1.DBaaSTenantList) (tenantNames []string) {
	for _, tenant := range tenantList.Items {
		tenantNames = append(tenantNames, tenant.Name)
	}
	return
}
