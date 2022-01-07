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
	"time"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// DBaaSDefaultTenantReconciler reconciles a DBaaSInventory object
type DBaaSDefaultTenantReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSDefaultTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// on operator startup, create default tenant if none exists
	return r.createDefaultTenant(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSDefaultTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// watch deployments if installed to the operator's namespace
	return ctrl.NewControllerManagedBy(mgr).
		Named("defaulttenant").
		For(
			&appsv1.Deployment{},
			builder.WithPredicates(r.ignoreOtherDeployments()),
			builder.OnlyMetadata,
		).
		Complete(r)
}

// only reconcile deployments which reside in the operator's install namespace, and only create events
func (r *DBaaSDefaultTenantReconciler) ignoreOtherDeployments() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object.GetNamespace() == r.InstallNamespace
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// create a default Tenant if one doesn't exist
func (r *DBaaSDefaultTenantReconciler) createDefaultTenant(ctx context.Context) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	defaultTenant := getDefaultTenant(r.InstallNamespace)

	// get list of DBaaSTenants for install/default namespace
	tenantList, err := r.tenantListByInventoryNS(ctx, defaultTenant.Spec.InventoryNamespace)
	if err != nil {
		logger.Error(err, "unable to list tenants")
		return ctrl.Result{}, err
	}

	// if no default tenant exists, create one
	if len(tenantList.Items) == 0 && !contains(getTenantNames(tenantList), defaultTenant.Name) {
		if err := r.Get(ctx, types.NamespacedName{Name: defaultTenant.Name}, &v1alpha1.DBaaSTenant{}); err != nil {
			// proceed only if default tenant not found
			if errors.IsNotFound(err) {
				logger.Info("resource not found", "Name", defaultTenant.Name)
				if err := r.Create(ctx, &defaultTenant); err != nil {
					// trigger retry if creation of default tenant fails
					logger.Error(err, "Error creating DBaaS Tenant resource", "Name", defaultTenant.Name)
					return ctrl.Result{RequeueAfter: time.Duration(30) * time.Second}, err
				}
				logger.Info("creating default DBaaS Tenant resource", "Name", defaultTenant.Name)
			} else {
				logger.Error(err, "Error getting the DBaaS Tenant resource", "Name", defaultTenant.Name)
			}
		}
	}

	return ctrl.Result{}, nil
}

func getDefaultTenant(inventoryNamespace string) v1alpha1.DBaaSTenant {
	return v1alpha1.DBaaSTenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: v1alpha1.DBaaSTenantSpec{
			InventoryNamespace: inventoryNamespace,
			Authz: v1alpha1.DBaasUsersGroups{
				Groups: []string{"system:authenticated"},
			},
		},
	}
}
