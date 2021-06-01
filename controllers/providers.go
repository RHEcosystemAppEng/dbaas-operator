package controllers

import (
	v1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *DBaaSInventoryReconciler) getDBaaSProviders(client client.Client, scheme *runtime.Scheme) v1.DBaaSProviderRegistrationList {
	return v1.DBaaSProviderRegistrationList{
		Items: []v1.DBaaSProviderRegistration{
			{
				TypeMeta: metav1.TypeMeta{},
				Spec: v1.DBaaSProviderRegistrationSpec{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Provider: v1.DatabaseProvider{
						Name: "MongoDB Atlas",
					},
					InventoryKind:        "AtlasAccount",
					AuthenticationFields: []string{"Organization ID", "Organization Public Key", "Organization Private Key"},
					ConnectionKind:       "AtlasConnection",
				},
			},
		},
	}
}
