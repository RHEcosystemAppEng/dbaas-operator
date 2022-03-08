/*
Copyright 2022.

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
	"strings"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	oauthzv1 "github.com/openshift/api/authorization/v1"
	oauthzclientv1 "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DBaaSAuthzReconciler struct {
	*DBaaSReconciler
	*oauthzclientv1.AuthorizationV1Client
}

// ResourceAccessReview for Service Admin Authz
// return users/groups who can create both inventory and secret objects in the tenant namespace
func (r *DBaaSAuthzReconciler) getServiceAdminAuthz(ctx context.Context, namespace string) *oauthzv1.ResourceAccessReviewResponse {
	logger := ctrl.LoggerFrom(ctx)
	// admin access review
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
		return &oauthzv1.ResourceAccessReviewResponse{}
	}

	// secret access review
	rar.Action.Resource = "secrets"
	rar.Action.Group = corev1.SchemeGroupVersion.Group
	rar.Action.Version = corev1.SchemeGroupVersion.Version
	secretResponse, err := r.AuthorizationV1Client.ResourceAccessReviews().Create(ctx, rar, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "Error creating ResourceAccessReview", secretResponse.Namespace, secretResponse.EvaluationError)
		return &oauthzv1.ResourceAccessReviewResponse{}
	}

	return &oauthzv1.ResourceAccessReviewResponse{
		// remove results of tenant access review, as they don't need the addtl the perms
		UsersSlice:  matchSlices(inventoryResponse.UsersSlice, secretResponse.UsersSlice),
		GroupsSlice: matchSlices(inventoryResponse.GroupsSlice, secretResponse.GroupsSlice),
	}
}

// ResourceAccessReview for Developer Authz
// return users/groups who can list or get individual inventories in the tenant namespace
func (r *DBaaSAuthzReconciler) getDeveloperAuthz(ctx context.Context, namespace string, inventoryList v1alpha1.DBaaSInventoryList) *oauthzv1.ResourceAccessReviewResponse {
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
		return &oauthzv1.ResourceAccessReviewResponse{}
	}
	users := inventoryResponse.UsersSlice
	groups := inventoryResponse.GroupsSlice

	// determine if any individual inventory access exists separate from list
	//   ... will add additional api load, a new call for every inventory object in valid tenant namespaces
	//   ... but this keeps inventory-specific access functionality. a dev user could have access to a single inventory
	for _, inv := range inventoryList.Items {
		rar.Action.Verb = "get"
		rar.Action.ResourceName = inv.Name
		invResponse, err := r.AuthorizationV1Client.ResourceAccessReviews().Create(ctx, rar, metav1.CreateOptions{})
		if err != nil {
			logger.Error(err, "Error creating ResourceAccessReview", invResponse.Namespace, invResponse.EvaluationError)
			return &oauthzv1.ResourceAccessReviewResponse{}
		}
		users = uniqueStrSlice(append(users, invResponse.UsersSlice...))
		groups = uniqueStrSlice(append(groups, invResponse.GroupsSlice...))
	}

	return &oauthzv1.ResourceAccessReviewResponse{
		UsersSlice:  users,
		GroupsSlice: groups,
	}
}

// ResourceAccessReview for Service Admin Authz
// return users/groups who can list tenant objects at the cluster level
func (r *DBaaSAuthzReconciler) getTenantListAuthz(ctx context.Context) *oauthzv1.ResourceAccessReviewResponse {
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
		return &oauthzv1.ResourceAccessReviewResponse{}
	}

	return tenantResponse
}

// Reconcile tenant to ensure proper RBAC is created. inventoryList should only contain inventory objects for the corresponding tenant namespace.
func (r *DBaaSAuthzReconciler) reconcileTenantRbacObjs(ctx context.Context, tenant v1alpha1.DBaaSTenant, serviceAdminAuthz, developerAuthz, tenantListAuthz *oauthzv1.ResourceAccessReviewResponse) error {
	clusterRole, clusterRolebinding := tenantRbacObjs(tenant, serviceAdminAuthz, developerAuthz, tenantListAuthz)
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

// gets rbac objects for a tenant's users
func tenantRbacObjs(tenant v1alpha1.DBaaSTenant, serviceAdminAuthz, developerAuthz, tenantListAuthz *oauthzv1.ResourceAccessReviewResponse) (rbacv1.ClusterRole, rbacv1.ClusterRoleBinding) {
	clusterRole := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dbaas-" + tenant.Name + "-tenant-viewer",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{v1alpha1.GroupVersion.Group},
				Resources:     []string{"dbaastenants", "dbaastenants/status"},
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
	users := uniqueStrSlice(append(serviceAdminAuthz.UsersSlice, developerAuthz.UsersSlice...))
	groups := uniqueStrSlice(append(serviceAdminAuthz.GroupsSlice, developerAuthz.GroupsSlice...))
	// remove results of tenant access review, as they don't need the addtl the perms
	users = removeFromSlice(tenantListAuthz.UsersSlice, users)
	groups = removeFromSlice(tenantListAuthz.GroupsSlice, groups)

	clusterRoleBinding.Subjects = getSubjects(users, groups, "")

	return clusterRole, clusterRoleBinding
}

// verify no edit or list permissions are assigned to a role
func hasNoEditOrListVerbs(roleObj client.Object) bool {
	var verbs []string
	var roleRules []rbacv1.PolicyRule
	editVerbs := []string{"create", "patch", "update", "delete", "list"}

	kind := roleObj.GetObjectKind().GroupVersionKind().Kind
	if kind == "Role" {
		roleRules = roleObj.(*rbacv1.Role).Rules
	}
	if kind == "ClusterRole" {
		roleRules = roleObj.(*rbacv1.ClusterRole).Rules
	}
	for _, rules := range roleRules {
		verbs = append(verbs, rules.Verbs...)
	}
	for _, verb := range editVerbs {
		if contains(verbs, verb) {
			return false
		}
	}
	return true
}

// create a slice of rbac subjects for use in role bindings
func getSubjects(users, groups []string, namespace string) []rbacv1.Subject {
	subjects := []rbacv1.Subject{}
	for _, user := range users {
		if strings.HasPrefix(user, "system:serviceaccount:") {
			sa := strings.Split(user, ":")
			if len(sa) > 3 {
				subjects = append(subjects, getSubject(sa[3], sa[2], "ServiceAccount"))
			}
		} else {
			subjects = append(subjects, getSubject(user, namespace, "User"))
		}
	}
	for _, group := range groups {
		subjects = append(subjects, getSubject(group, namespace, "Group"))
	}
	if len(subjects) == 0 {
		return nil
	}

	return subjects
}

// create an rbac subject for use in role bindings
func getSubject(name, namespace, rbacObjectKind string) rbacv1.Subject {
	subject := rbacv1.Subject{
		Kind:      rbacObjectKind,
		Name:      name,
		Namespace: namespace,
	}
	if subject.Kind != "ServiceAccount" {
		subject.APIGroup = rbacv1.SchemeGroupVersion.Group
	}
	return subject
}
