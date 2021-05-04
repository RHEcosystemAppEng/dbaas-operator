package models

import (
	"fmt"
	dbaasv1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ptr "k8s.io/utils/pointer"
	"strings"
)

func Deployment(dbaasConnection *dbaasv1.DBaaSConnection) *v12.Deployment {
	return &v12.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dbaas-deploy",
			Namespace: dbaasConnection.Namespace,
			Labels: map[string]string{
				"managed-by":      "dbaas-operator",
				"owner":           dbaasConnection.Name,
				"owner.kind":      dbaasConnection.Kind,
				"owner.namespace": dbaasConnection.Namespace,
			},
		},
	}
}

func OwnedDeployment(dbaasConnection *dbaasv1.DBaaSConnection) *v12.Deployment {
	return &v12.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dbaas-deploy",
			Namespace: dbaasConnection.Namespace,
			Labels: map[string]string{
				"managed-by":      "dbaas-operator",
				"owner":           dbaasConnection.Name,
				"owner.kind":      dbaasConnection.Kind,
				"owner.namespace": dbaasConnection.Namespace,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:                dbaasConnection.GetUID(),
					APIVersion:         "dbaas.redhat.com/v1",
					BlockOwnerDeletion: ptr.BoolPtr(false),
					Controller:         ptr.BoolPtr(true),
					Kind:               "DBaaSConnection",
					Name:               dbaasConnection.Name,
				},
			},
		},
	}
}

func MutateDeploymentSpec() v12.DeploymentSpec {
	return v12.DeploymentSpec{
		Replicas: ptr.Int32Ptr(0),
		Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"name": "bind-deploy"}},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"name": "bind-deploy"},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:            "bind-deploy",
						Image:           "quay.io/bmozaffa/busybox",
						ImagePullPolicy: v1.PullIfNotPresent,
						Command:         []string{"sh", "-c", "echo The app is running! && sleep 3600"},
						Ports:           []v1.ContainerPort{{ContainerPort: int32(8080), Protocol: v1.ProtocolTCP}},
					},
				},
			},
		},
	}
}

func ConfigMap(dbaasConnection *dbaasv1.DBaaSConnection) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("dbaas-atlas-connection-%s", dbaasConnection.Spec.Cluster.ID),
			Namespace: dbaasConnection.Namespace,
			Labels: map[string]string{
				"managed-by":      "dbaas-operator",
				"owner":           dbaasConnection.Name,
				"owner.kind":      dbaasConnection.Kind,
				"owner.namespace": dbaasConnection.Namespace,
			},
		},
	}
}

func OwnedConfigMap(dbaasConnection *dbaasv1.DBaaSConnection) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("dbaas-atlas-connection-%s", dbaasConnection.Spec.Cluster.ID),
			Namespace: dbaasConnection.Namespace,
			Labels: map[string]string{
				"managed-by":      "dbaas-operator",
				"owner":           dbaasConnection.Name,
				"owner.kind":      dbaasConnection.Kind,
				"owner.namespace": dbaasConnection.Namespace,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:                dbaasConnection.GetUID(),
					APIVersion:         "dbaas.redhat.com/v1",
					BlockOwnerDeletion: ptr.BoolPtr(false),
					Controller:         ptr.BoolPtr(true),
					Kind:               "DBaaSConnection",
					Name:               dbaasConnection.Name,
				},
			},
		},
	}
}

func MutateConfigMapData(dbaasConnection *dbaasv1.DBaaSConnection) map[string]string {
	return map[string]string{
		"database": dbaasConnection.Spec.Cluster.Name,
		"host":     strings.Replace(dbaasConnection.Spec.Cluster.ConnectionString, "mongodb+srv://", "", -1),
	}
}

func Secret(dbaasConnection *dbaasv1.DBaaSConnection) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("dbaas-atlas-connection-%s", dbaasConnection.Spec.Cluster.ID),
			Namespace: dbaasConnection.Namespace,
			Labels: map[string]string{
				"managed-by":      "dbaas-operator",
				"owner":           dbaasConnection.Name,
				"owner.kind":      dbaasConnection.Kind,
				"owner.namespace": dbaasConnection.Namespace,
			},
		},
	}
}

func OwnedSecret(dbaasConnection *dbaasv1.DBaaSConnection) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Opaque",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("dbaas-atlas-connection-%s", dbaasConnection.Spec.Cluster.ID),
			Namespace: dbaasConnection.Namespace,
			Labels: map[string]string{
				"managed-by":      "dbaas-operator",
				"owner":           dbaasConnection.Name,
				"owner.kind":      dbaasConnection.Kind,
				"owner.namespace": dbaasConnection.Namespace,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:                dbaasConnection.GetUID(),
					APIVersion:         "dbaas.redhat.com/v1",
					BlockOwnerDeletion: ptr.BoolPtr(false),
					Controller:         ptr.BoolPtr(true),
					Kind:               "DBaaSConnection",
					Name:               dbaasConnection.Name,
				},
			},
		},
	}
}

func UpdatedConnectionStatus(dbaasConnection *dbaasv1.DBaaSConnection) dbaasv1.DBaaSConnectionStatus {
	return dbaasv1.DBaaSConnectionStatus{
		DBConfigMap:   fmt.Sprintf("dbaas-atlas-connection-%s", dbaasConnection.Spec.Cluster.ID),
		DBCredentials: fmt.Sprintf("dbaas-atlas-connection-%s", dbaasConnection.Spec.Cluster.ID),
	}
}
