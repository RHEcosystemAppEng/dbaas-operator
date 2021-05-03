package models

import (
	"fmt"
	dbaasv1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1"
	atlas "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	v1 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ptr "k8s.io/utils/pointer"
)

func AtlasService(dbaasService *dbaasv1.DBaaSService) *v1.AtlasService {
	return &v1.AtlasService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "atlas.mongodb.com/v1",
			Kind:       "AtlasService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("atlas-service-%s", dbaasService.UID),
			Namespace: dbaasService.Namespace,
		},
	}
}

func OwnedAtlasService(dbaasService *dbaasv1.DBaaSService) *v1.AtlasService {
	return &v1.AtlasService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("atlas-service-%s", dbaasService.UID),
			Namespace: dbaasService.Namespace,
			Labels: map[string]string{
				"owner-resource": dbaasService.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:                dbaasService.GetUID(),
					APIVersion:         "dbaas.redhat.com/v1",
					BlockOwnerDeletion: ptr.BoolPtr(false),
					Controller:         ptr.BoolPtr(true),
					Kind:               "DBaaSService",
					Name:               dbaasService.Name,
				},
			},
		},
	}
}

func MutateAtlasServiceSpec(dbaasService *dbaasv1.DBaaSService) atlas.AtlasServiceSpec {
	return atlas.AtlasServiceSpec{
		Name: "dbaas-operator-test",
		ConnectionSecret: &atlas.ResourceRef{
			Name: dbaasService.Spec.CredentialsSecretName,
		},
	}
}

func DBaaSConnection(dbaasService *dbaasv1.DBaaSService) *dbaasv1.DBaaSConnection {
	return &dbaasv1.DBaaSConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "atlas-connection",
			Namespace: dbaasService.Namespace,
		},
	}
}
