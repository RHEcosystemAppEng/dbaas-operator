package servicebinding

import (
	"context"

	v1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	reconcilerscsv "github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/csv"
	"github.com/go-logr/logr"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	apiv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Reconciler struct {
	client client.Client
	logger logr.Logger
	scheme *runtime.Scheme
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger) reconcilers.PlatformReconciler {
	return &Reconciler{
		client: client,
		scheme: scheme,
		logger: logger,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, cr *v1.DBaaSPlatform, status2 *v1.DBaaSPlatformStatus) (v1.PlatformsInstlnStatus, error) {

	status, err := r.reconcileSubscription(cr, ctx)
	if status != v1.ResultSuccess {
		return status, err
	}

	status, err = r.waitFoServiceBindingOperator(cr, ctx)
	if status != v1.ResultSuccess {
		return status, err
	}

	// ServiceBinding csv
	status, err = r.reconcileCSV(cr, ctx)
	if status != v1.ResultSuccess {
		return status, err
	}
	status, err = r.waitForServiceBindingCSV(cr, ctx)
	return v1.ResultSuccess, nil

}
func (r *Reconciler) Cleanup(ctx context.Context, cr *v1.DBaaSPlatform) (v1.PlatformsInstlnStatus, error) {

	subscription := r.GetServiceBindingSubscription(cr)
	err := r.client.Delete(ctx, subscription)
	if err != nil && !errors.IsNotFound(err) {
		return v1.ResultFailed, err
	}
	deployments := &apiv1.DeploymentList{}
	opts := &client.ListOptions{
		Namespace: cr.Namespace,
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

	csv := r.getServiceBindingCSV(cr)
	err = r.client.Delete(ctx, csv)
	if err != nil && !errors.IsNotFound(err) {
		return v1.ResultFailed, err
	}

	return v1.ResultSuccess, nil
}

func (r *Reconciler) reconcileSubscription(cr *v1.DBaaSPlatform, ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	subscription := r.GetServiceBindingSubscription(cr)
	catalogsource := r.GetServiceBindingCatalogSource()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, subscription, func() error {
		if err := ctrl.SetControllerReference(cr, subscription, r.scheme); err != nil {
			return err
		}
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

func (r *Reconciler) GetServiceBindingSubscription(cr *v1.DBaaSPlatform) *v1alpha1.Subscription {
	return &v1alpha1.Subscription{
		ObjectMeta: apimv1.ObjectMeta{
			Name:      "rh-service-binding-operator-subscription",
			Namespace: cr.Namespace,
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

func (r *Reconciler) waitFoServiceBindingOperator(cr *v1.DBaaSPlatform, ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	deployments := &apiv1.DeploymentList{}
	opts := &client.ListOptions{
		Namespace: cr.Namespace,
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

func (r *Reconciler) reconcileCSV(cr *v1.DBaaSPlatform, ctx context.Context) (v1.PlatformsInstlnStatus, error) {
	csv := r.getServiceBindingCSV(cr)
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, csv, func() error {
		if err := ctrl.SetControllerReference(cr, csv, r.scheme); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return v1.ResultFailed, err
	}

	return v1.ResultSuccess, nil
}

func (r *Reconciler) waitForServiceBindingCSV(cr *v1.DBaaSPlatform, ctx context.Context) (v1.PlatformsInstlnStatus, error) {
	csv := r.getServiceBindingCSV(cr)
	err := r.client.Get(ctx, client.ObjectKeyFromObject(csv), csv)
	if err != nil {
		return v1.ResultFailed, err
	}

	if set, err := reconcilerscsv.CheckOwnerReferenceSet(cr, csv, r.scheme); err != nil {
		return v1.ResultFailed, err
	} else if set {
		return v1.ResultSuccess, nil
	}
	return v1.ResultInProgress, nil
}

func (r *Reconciler) getServiceBindingCSV(cr *v1.DBaaSPlatform) *v1alpha1.ClusterServiceVersion {
	return &v1alpha1.ClusterServiceVersion{
		ObjectMeta: apimv1.ObjectMeta{
			Name:      reconcilers.SERVICE_BINDING_CSV,
			Namespace: cr.Namespace,
		},
	}
}
