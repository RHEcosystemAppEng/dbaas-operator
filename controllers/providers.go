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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

func getDBaaSProvider(requestedProvider v1alpha1.DatabaseProvider, namespace string, client ctrlclient.Client, ctx context.Context) (v1alpha1.DBaaSProvider, error) {
	if cmList, err := getConfigMapListByController(namespace, client, ctx); err != nil {
		return v1alpha1.DBaaSProvider{}, err
	} else {
		if providers, err := getDBaaSProviders(cmList); err != nil {
			return v1alpha1.DBaaSProvider{}, err
		} else {
			for _, provider := range providers.Items {
				if reflect.DeepEqual(provider.Provider, requestedProvider) {
					return provider, nil
				}
			}
			return v1alpha1.DBaaSProvider{}, apierrors.NewNotFound(schema.GroupResource{
				Resource: strings.ToLower(requestedProvider.Name),
			}, requestedProvider.Name)
		}
	}
}

func getDBaaSProviders(cmList *v1.ConfigMapList) (*v1alpha1.DBaaSProviderList, error) {
	providers := make([]v1alpha1.DBaaSProvider, len(cmList.Items))
	for i, cm := range cmList.Items {
		var provider v1alpha1.DBaaSProvider
		if providerName, exists := cm.Data[providerDataKey]; exists {
			provider.Provider = v1alpha1.DatabaseProvider{Name: providerName}
		} else {
			return nil, errors.New("provider name is missing of the configured DBaaS provider")
		}
		if inventoryKind, exists := cm.Data[inventoryKindDataKey]; exists {
			provider.InventoryKind = inventoryKind
		} else {
			return nil, fmt.Errorf("inventory kind is missing of the configured DBaaS provider %s", provider.Provider.Name)
		}
		if connectionKind, exists := cm.Data[connectionKindDataKey]; exists {
			provider.ConnectionKind = connectionKind
		} else {
			return nil, fmt.Errorf("connection kind is missing of the configured DBaaS provider %s", provider.Provider.Name)
		}
		providers[i] = provider
	}

	return &v1alpha1.DBaaSProviderList{Items: providers}, nil
}

func getDBaaSProviderInventoryObjects(namespace string) ([]unstructured.Unstructured, error) {
	if cmList, err := getConfigMapList(namespace); err != nil {
		return nil, err
	} else {
		if providers, err := getDBaaSProviders(cmList); err != nil {
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

func getDBaaSProviderConnectionObjects(namespace string) ([]unstructured.Unstructured, error) {
	if cmList, err := getConfigMapList(namespace); err != nil {
		return nil, err
	} else {
		if providers, err := getDBaaSProviders(cmList); err != nil {
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

func getConfigMapListByController(namespace string, client ctrlclient.Client, ctx context.Context) (*v1.ConfigMapList, error) {
	logger := log.FromContext(ctx, "dbaasprovider", namespace)

	var cmList v1.ConfigMapList
	opts := []ctrlclient.ListOption{
		ctrlclient.InNamespace(namespace),
		ctrlclient.MatchingLabels{
			relatedToLabelName: relatedToLabelValue,
			typeLabelName:      typeLabelValue,
		},
	}

	if err := client.List(ctx, &cmList, opts...); err != nil {
		logger.Error(err, "Error reading ConfigMaps for the configured DBaaS Providers")
		return nil, err
	}
	return &cmList, nil
}

func getConfigMapList(namespace string) (*v1.ConfigMapList, error) {
	ctx := context.TODO()
	logger := log.FromContext(ctx, "dbaasprovider", namespace)

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error(err, "Error reading ConfigMaps for the configured DBaaS Providers")
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Error reading ConfigMaps for the configured DBaaS Providers")
		return nil, err
	}
	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			map[string]string{
				relatedToLabelName: relatedToLabelValue,
				typeLabelName:      typeLabelValue,
			}).
			String(),
	}

	if cmList, err := clientset.CoreV1().ConfigMaps(namespace).List(ctx, options); err != nil {
		logger.Error(err, "Error reading ConfigMaps for the configured DBaaS Providers")
		return nil, err
	} else {
		return cmList, nil
	}
}

func createProviderCR(object ctrlclient.Object, providerCRKind string, spec interface{}, client ctrlclient.Client, scheme *runtime.Scheme, ctx context.Context) error {
	logger := log.FromContext(ctx, "dbaasprovider", types.NamespacedName{Namespace: object.GetNamespace(), Name: object.GetName()})

	providerCR := &unstructured.Unstructured{}
	providerCR.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   object.GetObjectKind().GroupVersionKind().Group,
		Version: object.GetObjectKind().GroupVersionKind().Version,
		Kind:    providerCRKind,
	})
	providerCR.SetNamespace(object.GetNamespace())
	providerCR.SetName(object.GetName())
	providerCR.UnstructuredContent()["spec"] = spec
	if err := ctrl.SetControllerReference(object, providerCR, scheme); err != nil {
		logger.Error(err, "Error setting controller reference", "providerCR", providerCR)
		return err
	}
	if err := client.Create(ctx, providerCR); err != nil {
		logger.Error(err, "Error creating a provider CR", "providerCR", providerCR)
		return err
	}
	logger.Info("Provider CR resource created", "providerCR", providerCR)
	return nil
}

func updateProviderCR(object ctrlclient.Object, providerCRKind string, spec interface{}, client ctrlclient.Client, scheme *runtime.Scheme, ctx context.Context) error {
	logger := log.FromContext(ctx, "dbaasprovider", types.NamespacedName{Namespace: object.GetNamespace(), Name: object.GetName()})

	providerCR := &unstructured.Unstructured{}
	providerCR.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   object.GetObjectKind().GroupVersionKind().Group,
		Version: object.GetObjectKind().GroupVersionKind().Version,
		Kind:    providerCRKind,
	})
	providerCR.SetNamespace(object.GetNamespace())
	providerCR.SetName(object.GetName())
	providerCR.UnstructuredContent()["spec"] = spec
	if err := ctrl.SetControllerReference(object, providerCR, scheme); err != nil {
		logger.Error(err, "Error setting controller reference", "providerCR", providerCR)
		return err
	}
	if err := client.Update(ctx, providerCR); err != nil {
		logger.Error(err, "Error updating a provider CR", "providerCR", providerCR)
		return err
	}
	logger.Info("Provider CR resource updated", "providerCR", providerCR)
	return nil
}

func getProviderCR(object ctrlclient.Object, providerCRKind string, client ctrlclient.Client, ctx context.Context) (*unstructured.Unstructured, error) {
	gvk := schema.GroupVersionKind{
		Group:   object.GetObjectKind().GroupVersionKind().Group,
		Version: object.GetObjectKind().GroupVersionKind().Version,
		Kind:    providerCRKind,
	}

	logger := log.FromContext(ctx, "dbaasprovider", types.NamespacedName{Namespace: object.GetNamespace(), Name: object.GetName()}, "GVK", gvk)

	var providerCR = unstructured.Unstructured{}
	providerCR.SetGroupVersionKind(gvk)
	if err := client.Get(ctx, ctrlclient.ObjectKeyFromObject(object), &providerCR); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Provider CR resource not found", "providerCR", object.GetName())
			return nil, nil
		}
		logger.Error(err, "Error finding the provider CR", "providerCR", object.GetName())
		return nil, err
	}
	return &providerCR, nil
}
