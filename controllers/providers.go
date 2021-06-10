package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func getDBaaSProvider(requestedProvider v1alpha1.DatabaseProvider, client ctrlclient.Client, ctx context.Context, logger logr.Logger) (v1alpha1.DBaaSProvider, error) {
	if providers, err := getDBaaSProviders(client, ctx, logger); err != nil {
		return v1alpha1.DBaaSProvider{}, err
	} else {
		for _, provider := range providers.Items {
			if reflect.DeepEqual(provider.Provider, requestedProvider) {
				return provider, nil
			}
		}
		notFound := v1alpha1.DBaaSProvider{}
		return notFound, apierrors.NewNotFound(schema.GroupResource{
			Group:    schema.GroupVersionKind{}.Group,
			Resource: requestedProvider.Name,
		}, requestedProvider.Name)
	}
}

func getDBaaSProviders(client ctrlclient.Client, ctx context.Context, logger logr.Logger) (v1alpha1.DBaaSProviderList, error) {
	var cmList v1.ConfigMapList
	opts := []ctrlclient.ListOption{
		ctrlclient.InNamespace("dbaas-operator"),
		ctrlclient.MatchingLabels{"related-to": "dbaas-operator"},
		ctrlclient.MatchingLabels{"type": "dbaas-provider-registration"},
	}

	if err := client.List(ctx, &cmList, opts...); err != nil {
		logger.Error(err, "Error reading ConfigMaps for the configured DBaaS Providers")
		return v1alpha1.DBaaSProviderList{}, err
	}

	providers := make([]v1alpha1.DBaaSProvider, len(cmList.Items))
	for i, cm := range cmList.Items {
		var provider v1alpha1.DBaaSProvider
		if providerName, exists := cm.Data["provider"]; exists {
			provider.Provider = v1alpha1.DatabaseProvider{Name: providerName}
		} else {
			return v1alpha1.DBaaSProviderList{}, errors.New("provider name is missing of the configured DBaaS provider")
		}
		if inventoryKind, exists := cm.Data["inventory_kind"]; exists {
			provider.InventoryKind = inventoryKind
		} else {
			return v1alpha1.DBaaSProviderList{}, fmt.Errorf("inventory kind is missing of the configured DBaaS provider %s", provider.Provider.Name)
		}
		if connectionKind, exists := cm.Data["connection_kind"]; exists {
			provider.ConnectionKind = connectionKind
		} else {
			return v1alpha1.DBaaSProviderList{}, fmt.Errorf("connection kind is missing of the configured DBaaS provider %s", provider.Provider.Name)
		}
		providers[i] = provider
	}

	return v1alpha1.DBaaSProviderList{Items: providers}, nil
}
