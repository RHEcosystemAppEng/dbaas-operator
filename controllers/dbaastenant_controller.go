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

	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

// DBaaSTenantReconciler reconciles a DBaaSTenant object
type DBaaSTenantReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles/finalizers;clusterrolebindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx, "DBaaS Tenant", req.NamespacedName)

	// get list of DBaaSTenants
	tenantList := v1alpha1.DBaaSTenantList{}
	if err := r.List(ctx, &tenantList); err != nil {
		logger.Error(err, "Error fetching DBaaS Tenant List for reconcile")
		return ctrl.Result{}, err
	}

	// create default tenant if none exists
	if result, err := r.createDefaultTenant(ctx, tenantList); err != nil {
		return result, err
	}

	var tenant v1alpha1.DBaaSTenant
	if err := r.Get(ctx, req.NamespacedName, &tenant); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error fetching DBaaS Tenant for reconcile")
		return ctrl.Result{}, err
	}

	// Get list of DBaaSInventories from tenant namespace
	var inventoryList v1alpha1.DBaaSInventoryList
	if err := r.List(ctx, &inventoryList, &client.ListOptions{Namespace: tenant.Spec.InventoryNamespace}); err != nil {
		logger.Error(err, "Error fetching DBaaS Inventory List for reconcile")
		return ctrl.Result{}, err
	}

	//
	// Tenant RBAC
	//
	if err := r.reconcileTenantRbacObjs(ctx, tenant, getAllAuthzFromInventoryList(inventoryList, tenant)); err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile each inventory in the tenant's namespace to ensure proper RBAC is created
	for _, inventory := range inventoryList.Items {
		// should we return anything on err for these reconciles?
		if err := r.reconcileInventoryRbacObjs(ctx, inventory, tenantList); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// watch all tenants in the cluster
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSTenant{}).
		Build(r)
	if err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1alpha1.DBaaSTenant{}, inventoryNamespaceKey, func(rawObj client.Object) []string {
		tenant := rawObj.(*v1alpha1.DBaaSTenant)
		inventoryNS := tenant.Spec.InventoryNamespace
		return []string{inventoryNS}
	}); err != nil {
		return err
	}

	// watch deployments if owned by a csv and installed to the operator's namespace
	csvType := &unstructured.Unstructured{}
	csvType.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "operators.coreos.com",
		Kind:    "ClusterServiceVersion",
		Version: "v1alpha1",
	})
	if err = c.Watch(
		&source.Kind{Type: &appsv1.Deployment{}},
		&handler.EnqueueRequestForOwner{OwnerType: csvType},
		r.ignoreOtherDeployments(),
	); err != nil {
		return err
	}
	return nil
}

// only reconcile deployments which reside in the operator's install namespace
func (r *DBaaSTenantReconciler) ignoreOtherDeployments() predicate.Predicate {
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
func (r *DBaaSTenantReconciler) createDefaultTenant(ctx context.Context, tenantList v1alpha1.DBaaSTenantList) (ctrl.Result, error) {
	defaultName := "cluster"
	logger := ctrl.LoggerFrom(ctx, "default DBaaS Tenant", defaultName)

	tenantNames, tenantNamespaces := getTenantNamesAndNamespaces(tenantList)
	if !contains(tenantNames, defaultName) && !contains(tenantNamespaces, r.InstallNamespace) {
		defaultTenant := v1alpha1.DBaaSTenant{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultName,
			},
			Spec: v1alpha1.DBaaSTenantSpec{
				InventoryNamespace: r.InstallNamespace,
				Authz: v1alpha1.DBaasAuthz{
					Developer: v1alpha1.DBaasUsersGroups{
						Groups: []string{"system:authenticated"},
					},
				},
			},
		}

		if err := r.Get(ctx, types.NamespacedName{Name: defaultTenant.Name}, &v1alpha1.DBaaSTenant{}); err != nil {
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

// gets rbac objects for a tenant's users
func tenantRbacObjs(tenant v1alpha1.DBaaSTenant, inventoryAuthz v1alpha1.DBaasUsersGroups) (rbacv1.ClusterRole, rbacv1.ClusterRoleBinding) {
	clusterRole := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dbaas-" + tenant.Name + "-tenant-viewer",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{v1alpha1.GroupVersion.Group},
				Resources:     []string{"dbaastenants"},
				ResourceNames: []string{tenant.Name},
				Verbs:         []string{"get", "list", "watch"},
			},
			{
				APIGroups:     []string{v1alpha1.GroupVersion.Group},
				Resources:     []string{"dbaastenants/status"},
				ResourceNames: []string{tenant.Name},
				Verbs:         []string{"get"},
			},
		},
	}
	clusterRole.SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("ClusterRole"))

	clusterRoleBinding := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRole.Name + "s",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		},
	}
	clusterRoleBinding.SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding"))

	// give view access to all inventory devs and tenant service admins for the Tenant object
	users := uniqueStr(append(tenant.Spec.Authz.ServiceAdmin.Users, inventoryAuthz.Users...))
	groups := uniqueStr(append(tenant.Spec.Authz.ServiceAdmin.Groups, inventoryAuthz.Groups...))

	for _, user := range users {
		clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects, getSubject(user, "", "User"))
	}
	for _, group := range groups {
		clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects, getSubject(group, "", "Group"))
	}

	return clusterRole, clusterRoleBinding
}

// get cumulative authz from all inventories in namespace
func getAllAuthzFromInventoryList(inventoryList v1alpha1.DBaaSInventoryList, tenant v1alpha1.DBaaSTenant) (inventoryAuthz v1alpha1.DBaasUsersGroups) {
	var tenantDefaults bool
	for _, inventory := range inventoryList.Items {
		// if inventory.spec.authz is nil, apply authz from tenant.spec.authz.developer as a default
		if inventory.Spec.Authz.Users == nil && inventory.Spec.Authz.Groups == nil {
			tenantDefaults = true
		} else {
			inventoryAuthz.Users = append(inventoryAuthz.Users, inventory.Spec.Authz.Users...)
			inventoryAuthz.Groups = append(inventoryAuthz.Groups, inventory.Spec.Authz.Groups...)
		}
	}
	if tenantDefaults {
		inventoryAuthz.Users = append(inventoryAuthz.Users, tenant.Spec.Authz.Developer.Users...)
		inventoryAuthz.Groups = append(inventoryAuthz.Groups, tenant.Spec.Authz.Developer.Groups...)
	}
	return inventoryAuthz
}
