package servicebinding

import (
	"context"

	v1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	"github.com/go-logr/logr"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	apiv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Reconciler struct {
	client    client.Client
	logger    logr.Logger
	scheme    *runtime.Scheme
	namespace string
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger, namespace string) reconcilers.PlatformReconciler {
	return &Reconciler{
		client:    client,
		scheme:    scheme,
		logger:    logger,
		namespace: namespace,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, cr *v1.DBaaSPlatform, status2 *v1.DBaaSPlatformStatus) (v1.PlatformsInstlnStatus, error) {

	status, err := r.reconcileSubscription(ctx)
	if status != v1.ResultSuccess {
		return status, err
	}

	status, err = r.waitFoServiceBindingOperator(ctx)
	if status != v1.ResultSuccess {
		return status, err
	}
	return v1.ResultSuccess, nil

}
func (r *Reconciler) Cleanup(ctx context.Context, cr *v1.DBaaSPlatform) (v1.PlatformsInstlnStatus, error) {

	subscription := r.GetServiceBindingSubscription()
	err := r.client.Delete(ctx, subscription)
	if err != nil && !errors.IsNotFound(err) {
		return v1.ResultFailed, err
	}
	deployments := &apiv1.DeploymentList{}
	opts := &client.ListOptions{
		Namespace: r.namespace,
	}
	err = r.client.List(ctx, deployments, opts)
	if err != nil {
		return v1.ResultFailed, err
	}

	for d := range deployments.Items {
		if deployments.Items[d].Name == "service-binding-operator" {
			err = r.client.Delete(ctx, &deployments.Items[d])
			if err != nil && !errors.IsNotFound(err) {
				return v1.ResultFailed, err
			}
		}
	}

	return v1.ResultSuccess, nil
}

func (r *Reconciler) reconcileSubscription(ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	subscription := r.GetServiceBindingSubscription()
	catalogsource := r.GetServiceBindingCatalogSource()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, subscription, func() error {
		subscription.Spec = &v1alpha1.SubscriptionSpec{
			CatalogSource:          catalogsource.Name,
			CatalogSourceNamespace: catalogsource.Namespace,
			Package:                "rh-service-binding-operator",
			Channel:                "preview",
			InstallPlanApproval:    v1alpha1.ApprovalAutomatic,
		}

		return nil
	})

	if err != nil {
		return v1.ResultFailed, err
	}
	return v1.ResultSuccess, nil
}

func (r *Reconciler) GetServiceBindingSubscription() *v1alpha1.Subscription {
	return &v1alpha1.Subscription{
		ObjectMeta: apimv1.ObjectMeta{
			Name:      "rh-service-binding-operator-subscription",
			Namespace: r.namespace,
		},
	}
}

func (r *Reconciler) GetServiceBindingCatalogSource() *v1alpha1.CatalogSource {
	return &v1alpha1.CatalogSource{
		ObjectMeta: apimv1.ObjectMeta{
			Name:      "redhat-operators",
			Namespace: reconcilers.CATALOG_NAMESPACE,
		},
	}
}

func (r *Reconciler) waitFoServiceBindingOperator(ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	deployments := &apiv1.DeploymentList{}
	opts := &client.ListOptions{
		Namespace: r.namespace,
	}
	err := r.client.List(ctx, deployments, opts)
	if err != nil {
		return v1.ResultFailed, err
	}

	for _, deployment := range deployments.Items {
		if deployment.Name == "service-binding-operator" {
			if deployment.Status.ReadyReplicas > 0 {
				return v1.ResultSuccess, nil
			}
		}
	}
	return v1.ResultInProgress, nil
}
