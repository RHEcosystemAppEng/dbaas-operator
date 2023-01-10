package reconcilers

import (
	"context"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	coreosv1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// CheckOwnerReferenceSet check if the owner reference has been set for DBaaSPlatform CR
func CheckOwnerReferenceSet(cr *v1beta1.DBaaSPlatform, csv *v1alpha1.ClusterServiceVersion, scheme *runtime.Scheme) (bool, error) {
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

// GetDBaaSOperatorCSV returns the CSV for the DBaaS Operator
func GetDBaaSOperatorCSV(ctx context.Context, namespace string, name string, serverClient k8sclient.Client) (*v1alpha1.ClusterServiceVersion, error) {
	csv := GetClusterServiceVersion(namespace, name)

	if err := serverClient.Get(ctx, k8sclient.ObjectKeyFromObject(csv), csv); err != nil {
		return nil, err
	}
	return csv, nil
}

// GetClusterServiceVersion returns a CSV object
func GetClusterServiceVersion(namespace string, name string) *v1alpha1.ClusterServiceVersion {
	return &v1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

}

// GetSubscription returns a Subscription object
func GetSubscription(namespace string, name string) *v1alpha1.Subscription {

	return &v1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

}

// GetOperatorGroup returns an OperatorGroup object
func GetOperatorGroup(namespace string, name string) *coreosv1.OperatorGroup {

	return &coreosv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

}

// GetCatalogSource returns a CatalogSource object
func GetCatalogSource(namespace string, name string) *v1alpha1.CatalogSource {

	return &v1alpha1.CatalogSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

}
