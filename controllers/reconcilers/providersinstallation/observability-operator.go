package providersinstallation

import (
	"context"
	_ "embed"
	"fmt"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	opapi "github.com/openshift/api/config/v1"
	msoapi "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//go:embed observability-operator-cr.yaml
var observabilityOperatorCRBytes []byte

const (
	versionKeyName = "version"
	clusterIDLabel = "cluster_id"
	crName         = "dbaas-operator-mso"
)

func (r *reconciler) createObservabilityCR(ctx context.Context, cr *dbaasv1alpha1.DBaaSPlatform) (dbaasv1alpha1.PlatformsInstlnStatus, error) {

	monitoringStackCR := &msoapi.MonitoringStack{}
	namespace := cr.Namespace
	monitoringStackList := &msoapi.MonitoringStackList{}
	listOpts := []k8sclient.ListOption{
		k8sclient.InNamespace(namespace),
	}

	err := r.client.List(ctx, monitoringStackList, listOpts...)
	if err != nil {
		return dbaasv1alpha1.ResultFailed, fmt.Errorf("could not get a list of monitoring stack CR: %w", err)
	}
	if len(monitoringStackList.Items) == 0 {

		err = yaml.Unmarshal(observabilityOperatorCRBytes, monitoringStackCR)
		if err != nil {
			return dbaasv1alpha1.ResultFailed, err
		}

		monitoringStackCR.ObjectMeta = metav1.ObjectMeta{
			Name:      crName,
			Namespace: cr.Namespace,
		}

		clusterID, err := getClusterID(ctx, r.client)
		if err != nil {
			return dbaasv1alpha1.ResultFailed, err
		}
		if clusterID != "" {

			monitoringStackCR.Spec.PrometheusConfig.ExternalLabels = map[string]string{clusterIDLabel: clusterID}
		}

		err = controllerutil.SetControllerReference(cr, monitoringStackCR, r.scheme)
		if err != nil {
			return dbaasv1alpha1.ResultFailed, err
		}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.client, monitoringStackCR, func() error {
			monitoringStackCR.Labels = map[string]string{
				"managed-by": "dbaas-operator",
			}

			return nil
		}); err != nil {
			return dbaasv1alpha1.ResultFailed, err
		}
	} else if len(monitoringStackList.Items) == 1 {
		monitoringStackCR = &monitoringStackList.Items[0]
	} else {
		return dbaasv1alpha1.ResultFailed, fmt.Errorf("too many monitoringStackCR resources found. Expecting 1, found %d MonitoringStack resources in %s namespace", len(monitoringStackList.Items), namespace)
	}
	return dbaasv1alpha1.ResultSuccess, nil

}

// getClusterID Returns the cluster id by querying the ClusterVersion resource
func getClusterID(ctx context.Context, client k8sclient.Client) (string, error) {
	v := &opapi.ClusterVersion{}
	selector := k8sclient.ObjectKey{
		Name: versionKeyName,
	}

	err := client.Get(ctx, selector, v)
	if err != nil {
		return "", err
	}

	return string(v.Spec.ClusterID), nil
}
