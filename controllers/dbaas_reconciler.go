package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	relatedToLabelName  = "related-to"
	relatedToLabelValue = "dbaas-operator"
	typeLabelName       = "type"
	typeLabelValue      = "dbaas-provider-registration"

	providerDataKey       = "provider"
	inventoryKindDataKey  = "inventory_kind"
	connectionKindDataKey = "connection_kind"

	operatorGroup   = "dbaas.redhat.com"
	operatorVersion = "v1alpha1"
)

type DBaaSReconciler struct {
	client.Client
	*runtime.Scheme
}

func (p *DBaaSReconciler) getDBaaSProvider(requestedProvider v1alpha1.DatabaseProvider, namespace string, ctx context.Context) (v1alpha1.DBaaSProvider, error) {
	if cmList, err := p.getProviderCMList(namespace, ctx); err != nil {
		return v1alpha1.DBaaSProvider{}, err
	} else {
		if providers, err := p.parseDBaaSProviderList(cmList); err != nil {
			return v1alpha1.DBaaSProvider{}, err
		} else {
			for _, provider := range providers.Items {
				if reflect.DeepEqual(provider.Provider, requestedProvider) {
					return provider, nil
				}
			}
			return v1alpha1.DBaaSProvider{}, apierrors.NewNotFound(schema.GroupResource{
				Group:    operatorGroup,
				Resource: strings.ToLower(requestedProvider.Name),
			}, requestedProvider.Name)
		}
	}
}

func (p *DBaaSReconciler) parseDBaaSProviderList(cmList v1.ConfigMapList) (v1alpha1.DBaaSProviderList, error) {
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

func (p *DBaaSReconciler) preStartGetDBaaSProviderInventories(namespace string) ([]unstructured.Unstructured, error) {
	if cmList, err := p.preStartGetProviderCMList(namespace); err != nil {
		return nil, err
	} else {
		if providers, err := p.parseDBaaSProviderList(cmList); err != nil {
			return nil, err
		} else {
			objects := make([]unstructured.Unstructured, len(providers.Items))
			for i, provider := range providers.Items {
				object := unstructured.Unstructured{}
				object.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   operatorGroup,
					Version: operatorVersion,
					Kind:    provider.InventoryKind,
				})
				objects[i] = object
			}
			return objects, nil
		}
	}
}

func (p *DBaaSReconciler) preStartGetDBaaSProviderConnections(namespace string) ([]unstructured.Unstructured, error) {
	if cmList, err := p.preStartGetProviderCMList(namespace); err != nil {
		return nil, err
	} else {
		if providers, err := p.parseDBaaSProviderList(cmList); err != nil {
			return nil, err
		} else {
			objects := make([]unstructured.Unstructured, len(providers.Items))
			for i, provider := range providers.Items {
				object := unstructured.Unstructured{}
				object.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   operatorGroup,
					Version: operatorVersion,
					Kind:    provider.ConnectionKind,
				})
				objects[i] = object
			}
			return objects, nil
		}
	}
}

func (p *DBaaSReconciler) getProviderCMList(namespace string, ctx context.Context) (v1.ConfigMapList, error) {
	var cmList v1.ConfigMapList
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{
			relatedToLabelName: relatedToLabelValue,
			typeLabelName:      typeLabelValue,
		},
	}

	if err := p.List(ctx, &cmList, opts...); err != nil {
		return v1.ConfigMapList{}, err
	}
	return cmList, nil
}

func (p *DBaaSReconciler) preStartGetProviderCMList(namespace string) (v1.ConfigMapList, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return v1.ConfigMapList{}, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return v1.ConfigMapList{}, err
	}
	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			map[string]string{
				relatedToLabelName: relatedToLabelValue,
				typeLabelName:      typeLabelValue,
			}).
			String(),
	}

	if cmList, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), options); err != nil {
		return v1.ConfigMapList{}, err
	} else {
		return *cmList, nil
	}
}

func (p *DBaaSReconciler) createProviderObject(object client.Object, providerObjectKind string, spec interface{}, ctx context.Context) error {
	providerObject := &unstructured.Unstructured{}
	providerObject.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   object.GetObjectKind().GroupVersionKind().Group,
		Version: object.GetObjectKind().GroupVersionKind().Version,
		Kind:    providerObjectKind,
	})
	providerObject.SetNamespace(object.GetNamespace())
	providerObject.SetName(object.GetName())
	providerObject.UnstructuredContent()["spec"] = spec
	if err := ctrl.SetControllerReference(object, providerObject, p.Scheme); err != nil {
		return err
	}
	if err := p.Create(ctx, providerObject); err != nil {
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
	gvk := schema.GroupVersionKind{
		Group:   object.GetObjectKind().GroupVersionKind().Group,
		Version: object.GetObjectKind().GroupVersionKind().Version,
		Kind:    providerObjectKind,
	}

	var providerObject = unstructured.Unstructured{}
	providerObject.SetGroupVersionKind(gvk)
	if err := p.Get(ctx, client.ObjectKeyFromObject(object), &providerObject); err != nil {
		return unstructured.Unstructured{}, err
	}
	return providerObject, nil
}
