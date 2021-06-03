package controllers

import (
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *DBaaSInventoryReconciler) getDBaaSProviders(client client.Client, scheme *runtime.Scheme) v1alpha1.DBaaSProviderList {
	return v1alpha1.DBaaSProviderList{
		Items: []v1alpha1.DBaaSProvider{
			{
				Provider: v1alpha1.DatabaseProvider{
					Name: "MongoDB Atlas",
				},
				InventoryKind: "AtlasAccount",
				AuthenticationFields: []v1alpha1.AuthenticationField{
					{
						Name: "Organization ID",
					},
					{
						Name: "Organization Public Key",
					},
					{
						Name: "Organization Private Key",
					},
				},
				ConnectionKind: "AtlasConnection",
			},
		},
	}
}
