package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// InstallNamespaceEnvVar is the constant for env variable INSTALL_NAMESPACE
var InstallNamespaceEnvVar = "INSTALL_NAMESPACE"
var inventoryNamespaceKey = ".spec.inventoryNamespace"
var ignoreCreateEvents = predicate.Funcs{
	CreateFunc: func(e event.CreateEvent) bool {
		return false
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		if e.ObjectOld == nil || e.ObjectNew == nil {
			return false
		}
		return e.ObjectNew.GetResourceVersion() != e.ObjectOld.GetResourceVersion()
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return true
	},
	GenericFunc: func(e event.GenericEvent) bool {
		return true
	},
}
var ignoreAllEvents = predicate.Funcs{
	CreateFunc: func(e event.CreateEvent) bool {
		return false
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		return false
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return false
	},
	GenericFunc: func(e event.GenericEvent) bool {
		return false
	},
}

type DBaaSReconciler struct {
	client.Client
	*runtime.Scheme
	InstallNamespace string
}

func (p *DBaaSReconciler) getDBaaSProvider(providerName string, ctx context.Context) (*v1alpha1.DBaaSProvider, error) {
	provider := &v1alpha1.DBaaSProvider{}
	if err := p.Get(ctx, types.NamespacedName{Name: providerName}, provider); err != nil {
		return nil, err
	}
	return provider, nil
}

func (p *DBaaSReconciler) watchDBaaSProviderObject(ctrl controller.Controller, object runtime.Object, providerObjectKind string) error {
	providerObject := unstructured.Unstructured{}
	providerObject.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
		Kind:    providerObjectKind,
	})
	err := ctrl.Watch(
		&source.Kind{
			Type: &providerObject,
		},
		&handler.EnqueueRequestForOwner{
			OwnerType:    object,
			IsController: true,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *DBaaSReconciler) createProviderObject(object client.Object, providerObjectKind string) *unstructured.Unstructured {
	var providerObject unstructured.Unstructured
	providerObject.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
		Kind:    providerObjectKind,
	})
	providerObject.SetNamespace(object.GetNamespace())
	providerObject.SetName(object.GetName())
	return &providerObject
}

func (p *DBaaSReconciler) reconcileProviderObject(providerObject *unstructured.Unstructured, mutateFn controllerutil.MutateFn, ctx context.Context) (controllerutil.OperationResult, error) {
	return controllerutil.CreateOrUpdate(ctx, p.Client, providerObject, mutateFn)
}

func (p *DBaaSReconciler) providerObjectMutateFn(object client.Object, providerObject *unstructured.Unstructured, spec interface{}) controllerutil.MutateFn {
	return func() error {
		providerObject.UnstructuredContent()["spec"] = spec
		providerObject.SetOwnerReferences(nil)
		if err := ctrl.SetControllerReference(object, providerObject, p.Scheme); err != nil {
			return err
		}
		return nil
	}
}

func (p *DBaaSReconciler) parseProviderObject(unstructured *unstructured.Unstructured, object interface{}) error {
	b, err := unstructured.MarshalJSON()
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, object)
	if err != nil {
		return err
	}
	return nil
}

func (r *DBaaSReconciler) createOwnedObject(k8sObj, owner client.Object, ctx context.Context) error {
	if err := ctrl.SetControllerReference(owner, k8sObj, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, k8sObj); err != nil {
		return err
	}
	return nil
}

func (r *DBaaSReconciler) updateObject(k8sObj client.Object, ctx context.Context) error {
	if err := r.Update(ctx, k8sObj); err != nil {
		return err
	}
	return nil
}

// populate Tenant List based on spec.inventoryNamespace
func (r *DBaaSReconciler) tenantListByInventoryNS(ctx context.Context, inventoryNamespace string) (v1alpha1.DBaaSTenantList, error) {
	var tenantListByNS v1alpha1.DBaaSTenantList
	if err := r.List(ctx, &tenantListByNS, client.MatchingFields{inventoryNamespaceKey: inventoryNamespace}); err != nil {
		return v1alpha1.DBaaSTenantList{}, err
	}
	return tenantListByNS, nil
}

// update object upon ownerReference verification
func (r *DBaaSReconciler) updateIfOwned(ctx context.Context, owner, obj client.Object) error {
	logger := ctrl.LoggerFrom(ctx)
	name := obj.GetName()
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	if owns, err := isOwner(owner, obj, r.Scheme); !owns {
		logger.Info(kind+" ownership not verified, won't be updated", "Name", name)
		return err
	}
	if err := r.updateObject(obj, ctx); err != nil {
		logger.Error(err, "Error updating resource", "Name", name)
		return err
	}
	logger.Info(kind+" resource updated", "Name", name)
	return nil
}

// gets rbac objects for an inventory's users
func inventoryRbacObjs(inventory v1alpha1.DBaaSInventory, tenantList v1alpha1.DBaaSTenantList) (rbacv1.Role, rbacv1.RoleBinding) {
	role := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dbaas-" + inventory.Name + "-inventory-viewer",
			Namespace: inventory.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{v1alpha1.GroupVersion.Group},
				Resources:     []string{"dbaasinventories", "dbaasinventories/status"},
				ResourceNames: []string{inventory.Name},
				Verbs:         []string{"get"},
			},
		},
	}
	role.SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("Role"))

	roleBinding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      role.Name + "s",
			Namespace: role.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     "Role",
			Name:     role.Name,
		},
	}
	roleBinding.SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("RoleBinding"))

	// if inventory.Spec.Authz is nil, use tenant defaults for view access to the Inventory object
	var users, groups []string
	if inventory.Spec.Authz.Users == nil && inventory.Spec.Authz.Groups == nil {
		for _, tenant := range tenantList.Items {
			if tenant.Spec.InventoryNamespace == inventory.Namespace {
				users = append(users, tenant.Spec.Authz.Users...)
				groups = append(groups, tenant.Spec.Authz.Groups...)
			}
		}
	} else {
		users = inventory.Spec.Authz.Users
		groups = inventory.Spec.Authz.Groups
	}
	users = uniqueStrSlice(users)
	groups = uniqueStrSlice(groups)

	roleBinding.Subjects = getSubjects(users, groups, role.Namespace)

	return role, roleBinding
}

// gets rbac objects for a tenant's users
func tenantRbacObjs(tenant v1alpha1.DBaaSTenant, serviceAdminAuthz, developerAuthz, tenantListAuthz v1alpha1.DBaasUsersGroups) (rbacv1.ClusterRole, rbacv1.ClusterRoleBinding) {
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
	users := uniqueStrSlice(append(serviceAdminAuthz.Users, developerAuthz.Users...))
	groups := uniqueStrSlice(append(serviceAdminAuthz.Groups, developerAuthz.Groups...))
	// remove results of tenant access review, as they don't need the addtl the perms
	users = removeFromSlice(tenantListAuthz.Users, users)
	groups = removeFromSlice(tenantListAuthz.Groups, groups)

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

// get cumulative authz from all inventories in namespace
func getDevAuthzFromInventoryList(inventoryList v1alpha1.DBaaSInventoryList, tenant v1alpha1.DBaaSTenant) (developerAuthz v1alpha1.DBaasUsersGroups) {
	var tenantDefaults bool
	for _, inventory := range inventoryList.Items {
		if inventory.Namespace == tenant.Spec.InventoryNamespace {
			// if inventory.spec.authz is nil, apply authz from tenant.spec.authz.developer as a default
			if inventory.Spec.Authz.Users == nil && inventory.Spec.Authz.Groups == nil {
				tenantDefaults = true
			} else {
				developerAuthz.Users = append(developerAuthz.Users, inventory.Spec.Authz.Users...)
				developerAuthz.Groups = append(developerAuthz.Groups, inventory.Spec.Authz.Groups...)
			}
		}
	}
	if tenantDefaults {
		developerAuthz.Users = append(developerAuthz.Users, tenant.Spec.Authz.Users...)
		developerAuthz.Groups = append(developerAuthz.Groups, tenant.Spec.Authz.Groups...)
	}
	return developerAuthz
}

// checks if one object is set as owner/controller of another
func isOwner(owner, ownedObj client.Object, scheme *runtime.Scheme) (owns bool, err error) {
	exampleObj := &unstructured.Unstructured{}
	exampleObj.SetNamespace(owner.GetNamespace())
	if err = ctrl.SetControllerReference(owner, exampleObj, scheme); err == nil {
		for _, ownerRef := range exampleObj.GetOwnerReferences() {
			for _, ref := range ownedObj.GetOwnerReferences() {
				if reflect.DeepEqual(ownerRef, ref) {
					owns = true
				}
			}
		}
	}
	return owns, err
}

// GetInstallNamespace returns the operator's install Namespace
func GetInstallNamespace() (string, error) {
	ns, found := os.LookupEnv(InstallNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", InstallNamespaceEnvVar)
	}
	return ns, nil
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

// returns a unique matching subset of the provided slices
func matchSlices(input1, input2 []string) []string {
	m := []string{}
	for _, val1 := range input1 {
		for _, val2 := range input2 {
			if val1 == val2 {
				m = append(m, val1)
			}
		}
	}

	return uniqueStrSlice(m)
}

// returns a unique subset of the provided slice
func uniqueStrSlice(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

// returns a subset of the provided slices, after removing certain entries
func removeFromSlice(entriesToRemove, fromSlice []string) []string {
	r := []string{}
	for _, val := range fromSlice {
		if !contains(entriesToRemove, val) {
			r = append(r, val)
		}

	}

	return r
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

// checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
