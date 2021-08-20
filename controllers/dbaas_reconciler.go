package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// InstallNamespaceEnvVar is the constant for env variable INSTALL_NAMESPACE
var InstallNamespaceEnvVar = "INSTALL_NAMESPACE"
var inventoryNamespaceKey = ".spec.inventoryNamespace"

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

func (p *DBaaSReconciler) reconcileDBaaSObjectStatus(object client.Object, f controllerutil.MutateFn, ctx context.Context) error {
	if err := f(); err != nil {
		return err
	}
	return p.Status().Update(ctx, object)
}

func (r *DBaaSReconciler) createObject(k8sObj, owner client.Object, ctx context.Context) error {
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

// create RBAC object, return true if already exists
func (r *DBaaSReconciler) createRbacObj(newObj, getObj, owner client.Object, ctx context.Context) (exists bool, err error) {
	name := newObj.GetName()
	namespace := newObj.GetNamespace()
	logger := ctrl.LoggerFrom(ctx, owner.GetObjectKind().GroupVersionKind().Kind+" RBAC", types.NamespacedName{Name: name, Namespace: namespace})
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, getObj); err != nil {
		if errors.IsNotFound(err) {
			logger.V(1).Info("resource not found", name, namespace)
			if err = r.createObject(newObj, owner, ctx); err != nil {
				logger.Error(err, "Error creating resource", name, namespace)
				return false, err
			}
			logger.V(1).Info("resource created", name, namespace)
		} else {
			logger.Error(err, "Error getting the resource", name, namespace)
			return false, err
		}
	} else {
		return true, nil
	}
	return false, nil
}

// populate Tenant List based on spec.inventoryNamespace
func (r *DBaaSReconciler) tenantListByInventoryNS(ctx context.Context, inventoryNamespace string) (v1alpha1.DBaaSTenantList, error) {
	var tenantListByNS v1alpha1.DBaaSTenantList
	if err := r.List(ctx, &tenantListByNS, client.MatchingFields{inventoryNamespaceKey: inventoryNamespace}); err != nil {
		return v1alpha1.DBaaSTenantList{}, err
	}
	return tenantListByNS, nil
}

// Reconcile tenant to ensure proper RBAC is created
func (r *DBaaSReconciler) reconcileTenantRbacObjs(ctx context.Context, tenant v1alpha1.DBaaSTenant, inventoryAuthz v1alpha1.DBaasUsersGroups) error {
	logger := ctrl.LoggerFrom(ctx, "DBaaS Tenant", tenant.Name)

	clusterRole, clusterRolebinding := tenantRbacObjs(tenant, inventoryAuthz)
	var clusterRoleObj rbacv1.ClusterRole
	if exists, err := r.createRbacObj(&clusterRole, &clusterRoleObj, &tenant, ctx); err != nil {
		return err
	} else if exists {
		if !reflect.DeepEqual(clusterRole.Rules, clusterRoleObj.Rules) {
			clusterRoleObj.Rules = clusterRole.Rules
			if err := r.updateObject(&clusterRoleObj, ctx); err != nil {
				logger.Error(err, "Error updating resource", "Name", clusterRoleObj.Name)
				return err
			}
			logger.Info(clusterRoleObj.Kind+" resource updated", "Name", clusterRoleObj.Name)
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
			if err := r.updateObject(&clusterRoleBindingObj, ctx); err != nil {
				logger.Error(err, "Error updating resource", "Name", clusterRoleBindingObj.Name)
				return err
			}
			logger.Info(clusterRoleBindingObj.Kind+" resource updated", "Name", clusterRoleBindingObj.Name)
		}
	}

	return nil
}

// Reconcile inventory to ensure proper RBAC is created
func (r *DBaaSReconciler) reconcileInventoryRbacObjs(ctx context.Context, inventory v1alpha1.DBaaSInventory, tenantList v1alpha1.DBaaSTenantList) error {
	logger := ctrl.LoggerFrom(ctx, "DBaaS Inventory", inventory.Name)

	role, rolebinding := inventoryRbacObjs(inventory, tenantList)
	var roleObj rbacv1.Role
	if exists, err := r.createRbacObj(&role, &roleObj, &inventory, ctx); err != nil {
		return err
	} else if exists {
		if !reflect.DeepEqual(role.Rules, roleObj.Rules) {
			roleObj.Rules = role.Rules
			if err := r.updateObject(&roleObj, ctx); err != nil {
				logger.Error(err, "Error updating resource", roleObj.Name, roleObj.Namespace)
				return err
			}
			logger.V(1).Info(roleObj.Kind+" resource updated", roleObj.Name, roleObj.Namespace)
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
			if err := r.updateObject(&roleBindingObj, ctx); err != nil {
				logger.Error(err, "Error updating resource", roleBindingObj.Name, roleBindingObj.Namespace)
				return err
			}
			logger.V(1).Info(roleBindingObj.Kind+" resource updated", roleBindingObj.Name, roleBindingObj.Namespace)
		}
	}

	return nil
}

// get tenant names and namespaces from list
func getTenantNamesAndNamespaces(tenantList v1alpha1.DBaaSTenantList) (tenantNames, tenantNamespaces []string) {
	return getTenantNames(tenantList), getTenantNamespaces(tenantList)
}

// get tenant names from list
func getTenantNames(tenantList v1alpha1.DBaaSTenantList) (tenantNames []string) {
	for _, tenant := range tenantList.Items {
		tenantNames = append(tenantNames, tenant.Name)
	}
	return tenantNames
}

// get tenant namespaces from list
func getTenantNamespaces(tenantList v1alpha1.DBaaSTenantList) (tenantNamespaces []string) {
	for _, tenant := range tenantList.Items {
		tenantNamespaces = append(tenantNamespaces, tenant.Spec.InventoryNamespace)
	}
	return tenantNamespaces
}

// GetInstallNamespace returns the operator's install Namespace
func GetInstallNamespace() (string, error) {
	ns, found := os.LookupEnv(InstallNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", InstallNamespaceEnvVar)
	}
	return ns, nil
}

// create an rbac subject for use in role bindings
func getSubject(name, namespace, rbacObjectKind string) rbacv1.Subject {
	return rbacv1.Subject{
		APIGroup:  rbacv1.SchemeGroupVersion.Group,
		Kind:      rbacObjectKind,
		Name:      name,
		Namespace: namespace,
	}
}

// returns a unique subset of the provided slice
func uniqueStr(input []string) []string {
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

// checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
