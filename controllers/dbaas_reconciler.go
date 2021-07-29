package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

type DBaaSReconciler struct {
	client.Client
	*runtime.Scheme
}

func (p *DBaaSReconciler) getDBaaSProvider(providerName string, ctx context.Context) (v1alpha1.DBaaSProvider, error) {
	var provider v1alpha1.DBaaSProvider
	if err := p.Get(ctx, types.NamespacedName{Name: providerName}, &provider); err != nil {
		return v1alpha1.DBaaSProvider{}, err
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

func (p *DBaaSReconciler) parseProviderObject(object interface{}, unstructured *unstructured.Unstructured) error {
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

func (p *DBaaSReconciler) reconcileDBaaSObjectStatus(object client.Object, ctx context.Context, f controllerutil.MutateFn) error {
	if err := f(); err != nil {
		return err
	}
	return p.Status().Update(ctx, object)
}

// getInstallNamespace returns the Namespace the operator should be watching for single tenant changes
func (p *DBaaSReconciler) getInstallNamespace() (string, error) {
	// installNamespaceEnvVar is the constant for env variable INSTALL_NAMESPACE
	var installNamespaceEnvVar = "INSTALL_NAMESPACE"

	ns, found := os.LookupEnv(installNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", installNamespaceEnvVar)
	}
	return ns, nil
}
