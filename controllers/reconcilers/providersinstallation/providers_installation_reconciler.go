package providersinstallation

import (
	"context"
	"strconv"

	"github.com/go-logr/logr"
	coreosv1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	apiv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
)

type reconciler struct {
	client client.Client
	logger logr.Logger
	scheme *runtime.Scheme
	config v1beta1.PlatformConfig
}

// NewReconciler returns a provider installation reconciler
func NewReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger, config v1beta1.PlatformConfig) reconcilers.PlatformReconciler {
	return &reconciler{
		client: client,
		scheme: scheme,
		logger: logger,
		config: config,
	}
}

// Reconcile reconcile a DBaaSPlatform by creating the catalog source, a subscription and operator group
func (r *reconciler) Reconcile(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {

	status, err := r.reconcileCatalogSource(ctx)
	if status != v1beta1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileSubscription(ctx, cr)
	if status != v1beta1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileOperatorGroup(ctx)
	if status != v1beta1.ResultSuccess {
		return status, err
	}
	status, err = r.waitForOperator(ctx, cr)
	if status != v1beta1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileCSV(ctx, cr)
	if status != v1beta1.ResultSuccess {
		return status, err
	}

	return v1beta1.ResultSuccess, nil

}

// Cleanup cleanup resources associated with the DBaaSPlatform
func (r *reconciler) Cleanup(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {

	subscription := reconcilers.GetSubscription(cr.Namespace, r.config.Name+"-subscription")
	err := r.client.Delete(ctx, subscription)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}

	catalogSource := reconcilers.GetCatalogSource(reconcilers.CatalogNamespace, r.config.Name+"-catalogsource")
	err = r.client.Delete(ctx, catalogSource)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}
	deployments := &apiv1.DeploymentList{}
	opts := &client.ListOptions{
		Namespace: cr.Namespace,
	}
	err = r.client.List(ctx, deployments, opts)
	if err != nil {
		return v1beta1.ResultFailed, err
	}

	for d := range deployments.Items {
		if deployments.Items[d].Name == r.config.DeploymentName {
			err = r.client.Delete(ctx, &deployments.Items[d])
			if err != nil && !errors.IsNotFound(err) {
				return v1beta1.ResultFailed, err
			}
		}
	}

	csv := reconcilers.GetClusterServiceVersion(cr.Namespace, r.config.CSV)
	err = r.client.Delete(ctx, csv)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}

	return v1beta1.ResultSuccess, nil
}
func (r *reconciler) reconcileSubscription(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {

	subscription := reconcilers.GetSubscription(cr.Namespace, r.config.Name+"-subscription")
	catalogsource := reconcilers.GetCatalogSource(reconcilers.CatalogNamespace, r.config.Name+"-catalogsource")
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, subscription, func() error {
		if err := ctrl.SetControllerReference(cr, subscription, r.scheme); err != nil {
			return err
		}
		subscription.Spec = &v1alpha1.SubscriptionSpec{
			CatalogSource:          catalogsource.Name,
			CatalogSourceNamespace: catalogsource.Namespace,
			Package:                r.config.PackageName,
			Channel:                r.config.Channel,
			InstallPlanApproval:    v1alpha1.ApprovalAutomatic,
		}
		if r.config.CSV != "" {
			subscription.Spec.StartingCSV = r.config.CSV
		}
		if cr.Spec.SyncPeriod != nil {
			subscription.Spec.Config = &v1alpha1.SubscriptionConfig{
				Env: []corev1.EnvVar{
					{
						Name:  "SYNC_PERIOD_MIN",
						Value: strconv.Itoa(*cr.Spec.SyncPeriod),
					},
				},
			}
		}

		return nil
	})

	if err != nil {
		return v1beta1.ResultFailed, err
	}
	return v1beta1.ResultSuccess, nil
}
func (r *reconciler) reconcileOperatorGroup(ctx context.Context) (v1beta1.PlatformInstlnStatus, error) {

	operatorgroup := reconcilers.GetOperatorGroup(reconcilers.InstallNamespace, "global-operators")
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, operatorgroup, func() error {
		operatorgroup.Spec = coreosv1.OperatorGroupSpec{}

		return nil
	})
	if err != nil {
		return v1beta1.ResultFailed, err
	}

	return v1beta1.ResultSuccess, nil
}
func (r *reconciler) reconcileCatalogSource(ctx context.Context) (v1beta1.PlatformInstlnStatus, error) {
	catalogsource := reconcilers.GetCatalogSource(reconcilers.CatalogNamespace, r.config.Name+"-catalogsource")
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, catalogsource, func() error {
		catalogsource.Spec = v1alpha1.CatalogSourceSpec{
			SourceType:  v1alpha1.SourceTypeGrpc,
			Image:       r.config.Image,
			DisplayName: r.config.DisplayName,
			GrpcPodConfig: &v1alpha1.GrpcPodConfig{
				SecurityContextConfig: v1alpha1.Legacy,
			},
		}
		return nil
	})
	if err != nil {
		return v1beta1.ResultFailed, err
	}
	return v1beta1.ResultSuccess, nil
}

func (r *reconciler) waitForOperator(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {

	deployments := &apiv1.DeploymentList{}
	opts := &client.ListOptions{
		Namespace: cr.Namespace,
	}
	err := r.client.List(ctx, deployments, opts)
	if err != nil {
		return v1beta1.ResultFailed, err
	}

	for _, deployment := range deployments.Items {
		if deployment.Name == r.config.DeploymentName {
			if deployment.Status.ReadyReplicas > 0 {
				return v1beta1.ResultSuccess, nil
			}
		}
	}
	return v1beta1.ResultInProgress, nil
}

func (r *reconciler) reconcileCSV(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	csv := reconcilers.GetClusterServiceVersion(cr.Namespace, r.config.CSV)
	if err := r.client.Get(ctx, client.ObjectKeyFromObject(csv), csv); err != nil {
		if errors.IsNotFound(err) {
			return v1beta1.ResultInProgress, nil
		}
		return v1beta1.ResultFailed, err
	}

	if set, err := reconcilers.CheckOwnerReferenceSet(cr, csv, r.scheme); err != nil {
		return v1beta1.ResultFailed, err
	} else if set {
		return v1beta1.ResultSuccess, nil
	}

	if err := ctrl.SetControllerReference(cr, csv, r.scheme); err != nil {
		return v1beta1.ResultFailed, err
	}
	if err := r.client.Update(ctx, csv); err != nil {
		return v1beta1.ResultFailed, err
	}
	return v1beta1.ResultInProgress, nil
}
