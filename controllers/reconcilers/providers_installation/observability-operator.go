package providers_installation

import (
	"context"
	_ "embed"
	"fmt"
	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	msoapi "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//go:embed observability-operator-cr.yaml
var ObservabilityOperatorCRBytes []byte

func (r *Reconciler) createObservabilityCR(cr *dbaasv1alpha1.DBaaSPlatform, ctx context.Context) (dbaasv1alpha1.PlatformsInstlnStatus, error) {

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

		err = yaml.Unmarshal(ObservabilityOperatorCRBytes, monitoringStackCR)
		if err != nil {
			return dbaasv1alpha1.ResultFailed, err
		}

		monitoringStackCR.ObjectMeta = metav1.ObjectMeta{
			Name:      "dbaas-operator-mso",
			Namespace: cr.Namespace,
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
