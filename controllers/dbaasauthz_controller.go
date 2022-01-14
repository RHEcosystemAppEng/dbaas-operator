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
	"reflect"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	oauthzv1 "github.com/openshift/api/authorization/v1"
	oauthzclientv1 "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// DBaaSAuthzReconciler reconciles Rbac
type DBaaSAuthzReconciler struct {
	*DBaaSReconciler
	*oauthzclientv1.AuthorizationV1Client
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles/finalizers;rolebindings/finalizers;clusterroles/finalizers;clusterrolebindings/finalizers,verbs=update
//+kubebuilder:rbac:groups="";authorization.openshift.io,resources=localresourceaccessreviews;localsubjectaccessreviews;resourceaccessreviews;selfsubjectrulesreviews;subjectaccessreviews;subjectrulesreviews,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSAuthzReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
func (r *DBaaSAuthzReconciler) SetupWithManager(mgr ctrl.Manager) error {
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

// Reconcile all tenant and inventory RBAC
func (r *DBaaSAuthzReconciler) reconcileAuthz(ctx context.Context, namespace string) (err error) {
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
			if err := r.reconcileTenantRbacObjs(ctx, tenant, inventoryList, serviceAdminAuthz, developerAuthz, tenantListAuthz); err != nil {
				return err
			}
		}

		//
		// Inventory RBAC
		//
		// Reconcile each inventory in the tenant's namespace to ensure proper RBAC is created
		for _, inventory := range inventoryList.Items {
			if err := r.reconcileInventoryRbacObjs(ctx, inventory, tenantList); err != nil {
				return err
			}
		}
	}

	return nil
}

// Reconcile a specific tenant and all related inventory RBAC
func (r *DBaaSAuthzReconciler) reconcileTenantAuthz(ctx context.Context, tenant v1alpha1.DBaaSTenant) (err error) {
	logger := ctrl.LoggerFrom(ctx)

	// Get list of DBaaSInventories from tenant namespace
	var inventoryList v1alpha1.DBaaSInventoryList
	if err := r.List(ctx, &inventoryList, &client.ListOptions{Namespace: tenant.Spec.InventoryNamespace}); err != nil {
		logger.Error(err, "Error fetching DBaaS Inventory List for reconcile")
		return err
	}

	//
	// Tenant RBAC
	//
	serviceAdminAuthz := r.getServiceAdminAuthz(ctx, tenant.Spec.InventoryNamespace)
	developerAuthz := r.getDeveloperAuthz(ctx, tenant.Spec.InventoryNamespace, inventoryList)
	tenantListAuthz := r.getTenantListAuthz(ctx)
	if err := r.reconcileTenantRbacObjs(ctx, tenant, inventoryList, serviceAdminAuthz, developerAuthz, tenantListAuthz); err != nil {
		return err
	}

	return nil
}

// ResourceAccessReview for Service Admin Authz
// return users/groups who can create both inventory and secret objects in the tenant namespace
func (r *DBaaSAuthzReconciler) getServiceAdminAuthz(ctx context.Context, namespace string) v1alpha1.DBaasUsersGroups {
	logger := ctrl.LoggerFrom(ctx)
	// tenant access review
	rar := &oauthzv1.ResourceAccessReview{
		Action: oauthzv1.Action{
			Resource:  "dbaasinventories",
			Verb:      "create",
			Namespace: namespace,
			Group:     v1alpha1.GroupVersion.Group,
			Version:   v1alpha1.GroupVersion.Version,
		},
	}
	rar.SetGroupVersionKind(oauthzv1.SchemeGroupVersion.WithKind("ResourceAccessReview"))
	inventoryResponse, err := r.AuthorizationV1Client.ResourceAccessReviews().Create(ctx, rar, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "Error creating ResourceAccessReview", inventoryResponse.Namespace, inventoryResponse.EvaluationError)
		return v1alpha1.DBaasUsersGroups{}
	}

	// secret access review
	rar.Action.Resource = "secrets"
	rar.Action.Group = corev1.SchemeGroupVersion.Group
	rar.Action.Version = corev1.SchemeGroupVersion.Version
	secretResponse, err := r.AuthorizationV1Client.ResourceAccessReviews().Create(ctx, rar, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "Error creating ResourceAccessReview", secretResponse.Namespace, secretResponse.EvaluationError)
		return v1alpha1.DBaasUsersGroups{}
	}

	return v1alpha1.DBaasUsersGroups{
		// remove results of tenant access review, as they don't need the addtl the perms
		Users:  matchSlices(inventoryResponse.UsersSlice, secretResponse.UsersSlice),
		Groups: matchSlices(inventoryResponse.GroupsSlice, secretResponse.GroupsSlice),
	}
}

// ResourceAccessReview for Developer Authz
// return users/groups who can list inventories in the tenant namespace
func (r *DBaaSAuthzReconciler) getDeveloperAuthz(ctx context.Context, namespace string, inventoryList v1alpha1.DBaaSInventoryList) v1alpha1.DBaasUsersGroups {
	logger := ctrl.LoggerFrom(ctx)

	// inventory access review
	rar := &oauthzv1.ResourceAccessReview{
		Action: oauthzv1.Action{
			Resource:  "dbaasinventories",
			Verb:      "list",
			Namespace: namespace,
			Group:     v1alpha1.GroupVersion.Group,
			Version:   v1alpha1.GroupVersion.Version,
		},
	}
	rar.SetGroupVersionKind(oauthzv1.SchemeGroupVersion.WithKind("ResourceAccessReview"))
	inventoryResponse, err := r.AuthorizationV1Client.ResourceAccessReviews().Create(ctx, rar, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "Error creating ResourceAccessReview", inventoryResponse.Namespace, inventoryResponse.EvaluationError)
		return v1alpha1.DBaasUsersGroups{}
	}

	return v1alpha1.DBaasUsersGroups{
		Users:  inventoryResponse.UsersSlice,
		Groups: inventoryResponse.GroupsSlice,
	}
}

func consolidateDevAuthz(tenant v1alpha1.DBaaSTenant, inventoryList v1alpha1.DBaaSInventoryList, developerAuthz v1alpha1.DBaasUsersGroups) v1alpha1.DBaasUsersGroups {
	newDevAuthz := getDevAuthzFromInventoryList(inventoryList, tenant)
	users := append(developerAuthz.Users, newDevAuthz.Users...)
	groups := append(developerAuthz.Groups, newDevAuthz.Groups...)

	return v1alpha1.DBaasUsersGroups{
		Users:  uniqueStrSlice(users),
		Groups: uniqueStrSlice(groups),
	}
}

// ResourceAccessReview for Service Admin Authz
// return users/groups who can create both inventory and secret objects in the tenant namespace
func (r *DBaaSAuthzReconciler) getTenantListAuthz(ctx context.Context) v1alpha1.DBaasUsersGroups {
	logger := ctrl.LoggerFrom(ctx)
	// tenant access review
	rar := &oauthzv1.ResourceAccessReview{
		Action: oauthzv1.Action{
			Resource: "dbaastenants",
			Verb:     "list",
			Group:    v1alpha1.GroupVersion.Group,
			Version:  v1alpha1.GroupVersion.Version,
		},
	}
	rar.SetGroupVersionKind(oauthzv1.SchemeGroupVersion.WithKind("ResourceAccessReview"))
	tenantResponse, err := r.AuthorizationV1Client.ResourceAccessReviews().Create(ctx, rar, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "Error creating ResourceAccessReview", tenantResponse.Namespace, tenantResponse.EvaluationError)
		return v1alpha1.DBaasUsersGroups{}
	}

	return v1alpha1.DBaasUsersGroups{
		Users:  tenantResponse.UsersSlice,
		Groups: tenantResponse.GroupsSlice,
	}
}

// Reconcile tenant to ensure proper RBAC is created. inventoryList should only contain inventory objects for the corresponding tenant namespace.
func (r *DBaaSAuthzReconciler) reconcileTenantRbacObjs(ctx context.Context, tenant v1alpha1.DBaaSTenant, inventoryList v1alpha1.DBaaSInventoryList, serviceAdminAuthz, developerAuthz, tenantListAuthz v1alpha1.DBaasUsersGroups) error {
	devAuthz := consolidateDevAuthz(tenant, inventoryList, developerAuthz)
	clusterRole, clusterRolebinding := tenantRbacObjs(tenant, serviceAdminAuthz, devAuthz, tenantListAuthz)
	var clusterRoleObj rbacv1.ClusterRole
	if exists, err := r.createRbacObj(&clusterRole, &clusterRoleObj, &tenant, ctx); err != nil {
		return err
	} else if exists {
		if !reflect.DeepEqual(clusterRole.Rules, clusterRoleObj.Rules) {
			clusterRoleObj.Rules = clusterRole.Rules
			if err := r.updateIfOwned(ctx, &tenant, &clusterRoleObj); err != nil {
				return err
			}
		}
	}
	var clusterRoleBindingObj rbacv1.ClusterRoleBinding
	if exists, err := r.createRbacObj(&clusterRolebinding, &clusterRoleBindingObj, &tenant, ctx); err != nil {
		return err
	} else if exists {
		if !reflect.DeepEqual(clusterRolebinding.RoleRef, clusterRoleBindingObj.RoleRef) ||
			!reflect.DeepEqual(clusterRolebinding.Subjects, clusterRoleBindingObj.Subjects) {
			clusterRoleBindingObj.RoleRef = clusterRolebinding.RoleRef
			clusterRoleBindingObj.Subjects = clusterRolebinding.Subjects
			if err := r.updateIfOwned(ctx, &tenant, &clusterRoleBindingObj); err != nil {
				return err
			}
		}
	}

	return nil
}

// Reconcile inventory to ensure proper RBAC is created. tenantList should only contain tenant objects for the corresponding inventory namespace.
func (r *DBaaSAuthzReconciler) reconcileInventoryRbacObjs(ctx context.Context, inventory v1alpha1.DBaaSInventory, tenantList v1alpha1.DBaaSTenantList) error {
	role, rolebinding := inventoryRbacObjs(inventory, tenantList)
	var roleObj rbacv1.Role
	if exists, err := r.createRbacObj(&role, &roleObj, &inventory, ctx); err != nil {
		return err
	} else if exists {
		if !reflect.DeepEqual(role.Rules, roleObj.Rules) {
			roleObj.Rules = role.Rules
			if err := r.updateIfOwned(ctx, &inventory, &roleObj); err != nil {
				return err
			}
		}
	}
	var roleBindingObj rbacv1.RoleBinding
	if exists, err := r.createRbacObj(&rolebinding, &roleBindingObj, &inventory, ctx); err != nil {
		return err
	} else if exists {
		if !reflect.DeepEqual(rolebinding.RoleRef, roleBindingObj.RoleRef) ||
			!reflect.DeepEqual(rolebinding.Subjects, roleBindingObj.Subjects) {
			roleBindingObj.RoleRef = rolebinding.RoleRef
			roleBindingObj.Subjects = rolebinding.Subjects
			if err := r.updateIfOwned(ctx, &inventory, &roleBindingObj); err != nil {
				return err
			}
		}
	}

	return nil
}

// create RBAC object, return true if already exists
func (r *DBaaSAuthzReconciler) createRbacObj(newObj, getObj, owner client.Object, ctx context.Context) (exists bool, err error) {
	name := newObj.GetName()
	namespace := newObj.GetNamespace()
	kind := newObj.GetObjectKind().GroupVersionKind().Kind
	logger := ctrl.LoggerFrom(ctx)
	if hasNoEditOrListVerbs(newObj) {
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, getObj); err != nil {
			if errors.IsNotFound(err) {
				logger.V(1).Info(kind+" resource not found", name, namespace)
				if err = r.createOwnedObject(newObj, owner, ctx); err != nil {
					logger.Error(err, "Error creating resource", name, namespace)
					return false, err
				}
				logger.Info(kind+" resource created", name, namespace)
			} else {
				logger.Error(err, "Error getting the resource", name, namespace)
				return false, err
			}
		} else {
			return true, nil
		}
	} else {
		logger.V(1).Info(kind+" contains edit or list verbs, will not create", name, namespace)
	}
	return false, nil
}
