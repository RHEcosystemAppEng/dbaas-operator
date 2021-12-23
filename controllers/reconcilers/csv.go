package reconcilers

import (
	"context"
	alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func CheckOwnerReferenceSet(cr *alpha1.DBaaSPlatform, csv *v1alpha1.ClusterServiceVersion, scheme *runtime.Scheme) (bool, error) {
	gvk, err := apiutil.GVKForObject(cr, scheme)
	if err != nil {
		return false, err
	}
	ref := metav1.OwnerReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       cr.GetName(),
		UID:        cr.GetUID(),
	}

	existing := metav1.GetControllerOf(csv)
	if existing == nil {
		return false, nil
	}

	refGV, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return false, err
	}
	existingGV, err := schema.ParseGroupVersion(existing.APIVersion)
	if err != nil {
		return false, err
	}
	equal := refGV.Group == existingGV.Group && ref.Kind == existing.Kind && ref.Name == existing.Name
	return equal, nil
}

func GetDBaaSOperatorCSV(namespace string, ctx context.Context, serverClient k8sclient.Client) (*v1alpha1.ClusterServiceVersion, error) {
	csv := &v1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DBAAS_CSV,
			Namespace: namespace,
		},
	}
	if err := serverClient.Get(ctx, client.ObjectKeyFromObject(csv), csv); err != nil {
		return nil, err
	}
	return csv, nil
}
