package providers_installation

import (
	"context"
	_ "embed"
	"fmt"
	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	opapi "github.com/openshift/api/config/v1"
	msoapi "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

//go:embed observability-operator-cr.yaml
var ObservabilityOperatorCRBytes []byte

const (
	VersionKeyName = "version"
	ClusterIDLabel = "cluster_id"
	CRName         = "dbaas-operator-mso"
)

func (r *Reconciler) createObservabilityCR(cr *dbaasv1alpha1.DBaaSPlatform, ctx context.Context) (dbaasv1alpha1.PlatformsInstlnStatus, error) {

	monitoringStackCR := &msoapi.MonitoringStack{}
	namespace := cr.Namespace

	ObservatoriumGateway := os.Getenv("OBSERVATORIUM_GATEWAY")
	ObservatoriumTenant := os.Getenv("OBSERVATORIUM_TENANT")
	ObservatoriumToken := os.Getenv("OBSERVATORIUM_TOKEN")

	monitoringStackList := &msoapi.MonitoringStackList{}
	listOpts := []k8sclient.ListOption{
		k8sclient.InNamespace(namespace),
	}

	err := r.client.List(ctx, monitoringStackList, listOpts...)
	if err != nil {
		return dbaasv1alpha1.ResultFailed, fmt.Errorf("could not get a list of monitoring stack CR: %w", err)
	}
	if len(monitoringStackList.Items) == 0 {

		err = yaml.Unmarshal(ObservabilityOperatorCRBytes, monitoringStackCR)
		if err != nil {
			return dbaasv1alpha1.ResultFailed, err
		}

		monitoringStackCR.ObjectMeta = metav1.ObjectMeta{
			Name:      CRName,
			Namespace: cr.Namespace,
		}

		clusterID, err := GetClusterId(ctx, r.client)
		if err != nil {
			return dbaasv1alpha1.ResultFailed, err
		}
		if clusterID != "" {
			monitoringStackCR.Spec.PrometheusConfig.ExternalLabels = map[string]string{ClusterIDLabel: clusterID}
		}
		if monitoringStackCR.Spec.PrometheusConfig.RemoteWrite != nil {
			remoteWrites := monitoringStackCR.Spec.PrometheusConfig.RemoteWrite[0]
			if ObservatoriumGateway != "" && ObservatoriumTenant != "" && ObservatoriumToken != "" {
				remoteWrites.URL = fmt.Sprintf("%s/api/metrics/v1/%s/api/v1/receive", ObservatoriumGateway, ObservatoriumTenant)
				remoteWrites.BearerToken = strings.TrimSpace(ObservatoriumToken)
				monitoringStackCR.Spec.PrometheusConfig.RemoteWrite[0] = remoteWrites
			}
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
		if monitoringStackCR.Spec.PrometheusConfig.RemoteWrite != nil {
			remoteWrites := monitoringStackCR.Spec.PrometheusConfig.RemoteWrite[0]
			if ObservatoriumGateway != "" && ObservatoriumTenant != "" && ObservatoriumToken != "" {
				remoteWrites.URL = fmt.Sprintf("%s/api/metrics/v1/%s/api/v1/receive", ObservatoriumGateway, ObservatoriumTenant)
				remoteWrites.BearerToken = strings.TrimSpace(ObservatoriumToken)
				monitoringStackCR.Spec.PrometheusConfig.RemoteWrite[0] = remoteWrites
				if _, err := controllerutil.CreateOrUpdate(ctx, r.client, monitoringStackCR, func() error {
					monitoringStackCR.Labels = map[string]string{
						"managed-by": "dbaas-operator",
					}
					return nil
				}); err != nil {
					return dbaasv1alpha1.ResultFailed, err
				}
			}
		}
	} else {
		return dbaasv1alpha1.ResultFailed, fmt.Errorf("too many monitoringStackCR resources found. Expecting 1, found %d MonitoringStack resources in %s namespace", len(monitoringStackList.Items), namespace)
	}
	return dbaasv1alpha1.ResultSuccess, nil

}

// Returns the cluster id by querying the ClusterVersion resource
func GetClusterId(ctx context.Context, client k8sclient.Client) (string, error) {
	v := &opapi.ClusterVersion{}
	selector := k8sclient.ObjectKey{
		Name: VersionKeyName,
	}

	err := client.Get(ctx, selector, v)
	if err != nil {
		return "", err
	}

	return string(v.Spec.ClusterID), nil
}
