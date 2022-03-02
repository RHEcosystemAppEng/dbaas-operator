package prometheus

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type Reconciler struct {
	client     client.Client
	prometheus *promv1.Prometheus
	cr         *v1alpha1.DBaaSPlatform
	scheme     *runtime.Scheme
}

func NewReconciler(cr *v1alpha1.DBaaSPlatform, client client.Client, scheme *runtime.Scheme) reconcilers.PlatformReconciler {
	return &Reconciler{
		client:     client,
		prometheus: &promv1.Prometheus{},
		cr:         cr,
		scheme:     scheme,
	}
}
func (r *Reconciler) Reconcile(ctx context.Context, cr *v1alpha1.DBaaSPlatform, status2 *v1alpha1.DBaaSPlatformStatus) (v1alpha1.PlatformsInstlnStatus, error) {

	status, err := r.reconcileDeployment(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) Cleanup(ctx context.Context, cr *v1alpha1.DBaaSPlatform) (v1alpha1.PlatformsInstlnStatus, error) {

	deployment := r.prometheus
	err := r.client.Delete(ctx, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileDeployment(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, r.prometheus, func() error {
		// if err := r.own(r.prometheus); err != nil {
		// 	return err
		// }

		desired := PrometheusTemplate.DeepCopy()
		r.prometheus.Spec = desired.Spec
		return nil
	})

	if err != nil {
		if errors.IsConflict(err) {
			return v1alpha1.ResultInProgress, nil
		}
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

// func (r *Reconciler) own(resource metav1.Object) error {
//	 if err := ctrl.SetControllerReference(r.cr, resource, r.scheme); err != nil {
//		 return err
//	 }
//	 return nil
//  }
