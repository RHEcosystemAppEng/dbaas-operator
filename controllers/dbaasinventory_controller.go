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

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

// DBaaSInventoryReconciler reconciles a DBaaSInventory object
type DBaaSInventoryReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles/finalizers;rolebindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSInventoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx, "DBaaS Inventory", req.NamespacedName)
	// for now, limit to controller to reconciling inventories in the "install" namespace...
	//     will expand to other namespaces when multi-tenancy logic is added.
	ns, err := r.getInstallNamespace()
	if err == nil && ns == req.Namespace {

		var inventory v1alpha1.DBaaSInventory
		if err := r.Get(ctx, req.NamespacedName, &inventory); err != nil {
			if errors.IsNotFound(err) {
				// CR deleted since request queued, child objects getting GC'd, no requeue
				logger.Info("DBaaS Inventory resource not found, has been deleted")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "Error fetching DBaaS Inventory for reconcile")
			return ctrl.Result{}, err
		}

		//
		// RBAC
		//
		role, rolebinding := inventoryRbacObjs(inventory)
		var roleObj rbacv1.Role
		if err := r.Get(ctx, types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, &roleObj); err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Inventory rbac resource not found", "Inventory RBAC", role.Name)
				if err = r.createRoleObject(role, &inventory, ctx); err != nil {
					logger.Error(err, "Error creating Inventory rbac resource", "Inventory RBAC", role.Name)
					return ctrl.Result{}, err
				}
				logger.Info("Inventory rbac resource created", "Inventory RBAC", role.Name)
			} else {
				logger.Error(err, "Error finding the Inventory rbac resource", "Inventory RBAC", role.Name)
				return ctrl.Result{}, err
			}
		}
		var roleBindingObj rbacv1.RoleBinding
		if err := r.Get(ctx, types.NamespacedName{Name: rolebinding.Name, Namespace: rolebinding.Namespace}, &roleBindingObj); err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Inventory rbac resource not found", "Inventory RBAC", rolebinding.Name)
				if err = r.createRoleBindingObject(rolebinding, &inventory, ctx); err != nil {
					logger.Error(err, "Error creating Inventory rbac resource", "Inventory RBAC", rolebinding.Name)
					return ctrl.Result{}, err
				}
				logger.Info("Inventory rbac resource created", "Inventory RBAC", rolebinding.Name)
			} else {
				logger.Error(err, "Error finding the Inventory rbac resource", "Inventory RBAC", rolebinding.Name)
				return ctrl.Result{}, err
			}
		}

		//
		// Provider Inventory
		//
		provider, err := r.getDBaaSProvider(inventory.Spec.ProviderRef.Name, ctx)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "Requested DBaaS Provider is not configured in this environment", "DBaaS Provider", inventory.Spec.ProviderRef)
				return ctrl.Result{}, err
			}
			logger.Error(err, "Error reading configured DBaaS Provider", "DBaaS Provider", inventory.Spec.ProviderRef)
			return ctrl.Result{}, err
		}
		logger.Info("Found DBaaS Provider", "DBaaS Provider", inventory.Spec.ProviderRef)

		providerInventory := r.createProviderObject(&inventory, provider.Spec.InventoryKind)
		if result, err := r.reconcileProviderObject(providerInventory, r.providerObjectMutateFn(&inventory, providerInventory, inventory.Spec.DeepCopy()), ctx); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Provider Inventory modified, retry syncing spec")
				return ctrl.Result{Requeue: true}, nil
			}
			logger.Error(err, "Error reconciling the Provider Inventory resource")
			return ctrl.Result{}, err
		} else {
			logger.Info("Provider Inventory resource reconciled", "result", result)
		}

		var DBaaSProviderInventory v1alpha1.DBaaSProviderInventory
		if err := r.parseProviderObject(&DBaaSProviderInventory, providerInventory); err != nil {
			logger.Error(err, "Error parsing the Provider Inventory resource")
			return ctrl.Result{}, err
		}
		if err := r.reconcileDBaaSObjectStatus(&inventory, ctx, func() error {
			DBaaSProviderInventory.Status.DeepCopyInto(&inventory.Status)
			return nil
		}); err != nil {
			if errors.IsConflict(err) {
				logger.Info("DBaaS Inventory modified, retry syncing status")
				return ctrl.Result{Requeue: true}, nil
			}
			logger.Error(err, "Error updating the DBaaS Inventory status")
			return ctrl.Result{}, err
		} else {
			logger.Info("DBaaS Inventory status updated")
		}
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSInventoryReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBaaSInventory{}).
		Build(r)
}

// inventoryRbacObjs sets up rbac for an inventory's users
func inventoryRbacObjs(inventory v1alpha1.DBaaSInventory) (rbacv1.Role, rbacv1.RoleBinding) {
	role := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dbaas-" + inventory.Name + "-developer",
			Namespace: inventory.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{v1alpha1.GroupVersion.Group},
				Resources:     []string{"dbaasinventories"},
				ResourceNames: []string{inventory.Name},
				Verbs:         []string{"get", "list", "watch"},
			},
			{
				APIGroups:     []string{v1alpha1.GroupVersion.Group},
				Resources:     []string{"dbaasinventories/status"},
				ResourceNames: []string{inventory.Name},
				Verbs:         []string{"get"},
			},
		},
	}

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
	if inventory.Spec.Authz.Users != nil || inventory.Spec.Authz.Groups != nil {
		for _, user := range inventory.Spec.Authz.Users {
			roleBinding.Subjects = append(roleBinding.Subjects, getSubject(user, role.Namespace, "User"))
		}
		for _, group := range inventory.Spec.Authz.Groups {
			roleBinding.Subjects = append(roleBinding.Subjects, getSubject(group, role.Namespace, "Group"))
		}
	} else {
		roleBinding.Subjects = []rbacv1.Subject{getSubject("system:authenticated", role.Namespace, "Group")}
	}

	return role, roleBinding
}

func getSubject(name, namespace, rbacObjectKind string) rbacv1.Subject {
	return rbacv1.Subject{
		APIGroup:  rbacv1.SchemeGroupVersion.Group,
		Kind:      rbacObjectKind,
		Name:      name,
		Namespace: namespace,
	}
}

func (r *DBaaSInventoryReconciler) createRoleObject(role rbacv1.Role, owner client.Object, ctx context.Context) error {
	var rbacObject unstructured.Unstructured
	rbacObject.SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("Role"))
	rbacObject.SetNamespace(role.Namespace)
	rbacObject.SetName(role.Name)
	rbacObject.UnstructuredContent()["rules"] = role.Rules
	if err := ctrl.SetControllerReference(owner, &rbacObject, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, &rbacObject); err != nil {
		return err
	}
	return nil
}

func (r *DBaaSInventoryReconciler) createRoleBindingObject(roleBinding rbacv1.RoleBinding, owner client.Object, ctx context.Context) error {
	var rbacObject unstructured.Unstructured
	rbacObject.SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("RoleBinding"))
	rbacObject.SetNamespace(roleBinding.Namespace)
	rbacObject.SetName(roleBinding.Name)
	rbacObject.UnstructuredContent()["roleRef"] = roleBinding.RoleRef
	rbacObject.UnstructuredContent()["subjects"] = roleBinding.Subjects
	if err := ctrl.SetControllerReference(owner, &rbacObject, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, &rbacObject); err != nil {
		return err
	}
	return nil
}
