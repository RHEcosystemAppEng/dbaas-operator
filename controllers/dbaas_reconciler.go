package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

const (
	relatedToLabelKey   = "related-to"
	relatedToLabelValue = "dbaas-operator"
	typeLabelKey        = "type"
	typeLabelValue      = "dbaas-provider-registration"

	providerDataKey       = "provider"
	inventoryKindDataKey  = "inventory_kind"
	connectionKindDataKey = "connection_kind"
)

var ConfigMapSelector = map[string]string{
	relatedToLabelKey: relatedToLabelValue,
	typeLabelKey:      typeLabelValue,
}

type DBaaSReconciler struct {
	client.Client
	*runtime.Scheme
}

func (p *DBaaSReconciler) getDBaaSProvider(requestedProvider v1alpha1.DatabaseProvider, ctx context.Context) (v1alpha1.DBaaSProvider, error) {
	cmList, err := p.getProviderCMList(ctx)
	if err != nil {
		return v1alpha1.DBaaSProvider{}, err
	}

	providers, err := p.ParseDBaaSProviderList(cmList)
	if err != nil {
		return v1alpha1.DBaaSProvider{}, err
	}

	for _, provider := range providers.Items {
		if reflect.DeepEqual(provider.Provider, requestedProvider) {
			return provider, nil
		}
	}
	return v1alpha1.DBaaSProvider{}, apierrors.NewNotFound(schema.GroupResource{
		Group:    v1alpha1.GroupVersion.Group,
		Resource: strings.ToLower(requestedProvider.Name),
	}, requestedProvider.Name)
}

func (p *DBaaSReconciler) ParseDBaaSProviderList(cmList v1.ConfigMapList) (v1alpha1.DBaaSProviderList, error) {
	providers := make([]v1alpha1.DBaaSProvider, len(cmList.Items))
	for i, cm := range cmList.Items {
		var provider v1alpha1.DBaaSProvider
		if providerName, exists := cm.Data[providerDataKey]; exists {
			provider.Provider = v1alpha1.DatabaseProvider{Name: providerName}
		} else {
			return v1alpha1.DBaaSProviderList{}, errors.New("provider name is missing of the configured DBaaS Provider")
		}
		if inventoryKind, exists := cm.Data[inventoryKindDataKey]; exists {
			provider.InventoryKind = inventoryKind
		} else {
			return v1alpha1.DBaaSProviderList{}, fmt.Errorf("inventory kind is missing of the configured DBaaS Provider %s", provider.Provider.Name)
		}
		if connectionKind, exists := cm.Data[connectionKindDataKey]; exists {
			provider.ConnectionKind = connectionKind
		} else {
			return v1alpha1.DBaaSProviderList{}, fmt.Errorf("connection kind is missing of the configured DBaaS Provider %s", provider.Provider.Name)
		}
		providers[i] = provider
	}

	return v1alpha1.DBaaSProviderList{Items: providers}, nil
}

func (p *DBaaSReconciler) parseDBaaSProviderInventories(providerList v1alpha1.DBaaSProviderList) []*unstructured.Unstructured {
	objects := make([]*unstructured.Unstructured, len(providerList.Items))
	for i, provider := range providerList.Items {
		object := &unstructured.Unstructured{}
		object.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    provider.InventoryKind,
		})
		objects[i] = object
	}
	return objects
}

func (p *DBaaSReconciler) parseDBaaSProviderConnections(providerList v1alpha1.DBaaSProviderList) []*unstructured.Unstructured {
	objects := make([]*unstructured.Unstructured, len(providerList.Items))
	for i, provider := range providerList.Items {
		object := &unstructured.Unstructured{}
		object.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    provider.ConnectionKind,
		})
		objects[i] = object
	}
	return objects
}

func (p *DBaaSReconciler) getProviderCMList(ctx context.Context) (v1.ConfigMapList, error) {
	var cmList v1.ConfigMapList
	opts := []client.ListOption{
		client.MatchingLabels(ConfigMapSelector),
	}

	if err := p.List(ctx, &cmList, opts...); err != nil {
		return v1.ConfigMapList{}, err
	}
	return cmList, nil
}

func (p *DBaaSReconciler) PreStartGetProviderCMList() (v1.ConfigMapList, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return v1.ConfigMapList{}, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return v1.ConfigMapList{}, err
	}
	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(ConfigMapSelector).String(),
	}

	if cmList, err := clientset.CoreV1().ConfigMaps("").List(context.TODO(), options); err != nil {
		return v1.ConfigMapList{}, err
	} else {
		return *cmList, nil
	}
}

func (p *DBaaSReconciler) createProviderObject(object client.Object, providerObjectKind string, spec interface{}, ctx context.Context) error {
	var providerObject unstructured.Unstructured
	providerObject.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
		Kind:    providerObjectKind,
	})
	providerObject.SetNamespace(object.GetNamespace())
	providerObject.SetName(object.GetName())
	providerObject.UnstructuredContent()["spec"] = spec
	if err := ctrl.SetControllerReference(object, &providerObject, p.Scheme); err != nil {
		return err
	}
	if err := p.Create(ctx, &providerObject); err != nil {
		return err
	}
	return nil
}

func (p *DBaaSReconciler) updateProviderObject(providerObject unstructured.Unstructured, spec interface{}, ctx context.Context) error {
	providerObject.UnstructuredContent()["spec"] = spec
	if err := p.Update(ctx, &providerObject); err != nil {
		return err
	}
	return nil
}

func (p *DBaaSReconciler) getProviderObject(object client.Object, providerObjectKind string, ctx context.Context) (unstructured.Unstructured, error) {
	var providerObject unstructured.Unstructured
	providerObject.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
		Kind:    providerObjectKind,
	})
	if err := p.Get(ctx, client.ObjectKeyFromObject(object), &providerObject); err != nil {
		return unstructured.Unstructured{}, err
	}
	return providerObject, nil
}
