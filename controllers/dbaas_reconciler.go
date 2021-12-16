package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
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
	logger := ctrl.LoggerFrom(ctx, owner.GetObjectKind().GroupVersionKind().Kind, owner.GetName())
	name := obj.GetName()
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	if owns, err := isOwner(owner, obj, r.Scheme); !owns {
		logger.V(1).Info(kind+" ownership not verified, won't be updated", "Name", name)
		return err
	}
	if err := r.updateObject(obj, ctx); err != nil {
		logger.Error(err, "Error updating resource", "Name", name)
		return err
	}
	logger.Info(kind+" resource updated", "Name", name)
	return nil
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

// checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
