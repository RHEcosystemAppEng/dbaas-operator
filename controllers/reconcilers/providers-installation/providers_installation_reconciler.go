package providers_installation

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

	v1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
)

type Reconciler struct {
	client                 client.Client
	logger                 logr.Logger
	scheme                 *runtime.Scheme
	operatorName           string
	operatorCSV            string
	operatorDeploymentName string
	operatorCatalogImage   string
	operatorPackageName    string
	operatorChanel         string
	operatorDisplayName    string
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger, operatorName string, operatorCSV string, operatorDeploymentName string, operatorCatalogImage string, operatorPackageName string, operatorChannel string, operatorDisplayName string) reconcilers.PlatformReconciler {
	return &Reconciler{
		client:                 client,
		scheme:                 scheme,
		logger:                 logger,
		operatorName:           operatorName,
		operatorCSV:            operatorCSV,
		operatorDeploymentName: operatorDeploymentName,
		operatorCatalogImage:   operatorCatalogImage,
		operatorPackageName:    operatorPackageName,
		operatorChanel:         operatorChannel,
		operatorDisplayName:    operatorDisplayName,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, cr *v1.DBaaSPlatform, status2 *v1.DBaaSPlatformStatus) (v1.PlatformsInstlnStatus, error) {

	status, err := r.reconcileCatalogSource(ctx)
	if status != v1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileSubscription(cr, ctx)
	if status != v1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileOperatorGroup(ctx)
	if status != v1.ResultSuccess {
		return status, err
	}
	status, err = r.waitForOperator(cr, ctx)
	if status != v1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileCSV(cr, ctx)
	if status != v1.ResultSuccess {
		return status, err
	}
	return v1.ResultSuccess, nil

}

func (r *Reconciler) Cleanup(ctx context.Context, cr *v1.DBaaSPlatform) (v1.PlatformsInstlnStatus, error) {

	subscription := reconcilers.GetSubscription(cr.Namespace, r.operatorName+"-subscription")
	err := r.client.Delete(ctx, subscription)
	if err != nil && !errors.IsNotFound(err) {
		return v1.ResultFailed, err
	}

	catalogSource := reconcilers.GetCatalogSource(reconcilers.CATALOG_NAMESPACE, r.operatorName+"-catalogsource")
	err = r.client.Delete(ctx, catalogSource)
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
		if deployments.Items[d].Name == r.operatorDeploymentName {
			err = r.client.Delete(ctx, &deployments.Items[d])
			if err != nil && !errors.IsNotFound(err) {
				return v1.ResultFailed, err
			}
		}
	}

	csv := reconcilers.GetClusterServiceVersion(cr.Namespace, r.operatorCSV)
	err = r.client.Delete(ctx, csv)
	if err != nil && !errors.IsNotFound(err) {
		return v1.ResultFailed, err
	}

	return v1.ResultSuccess, nil
}
func (r *Reconciler) reconcileSubscription(cr *v1.DBaaSPlatform, ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	subscription := reconcilers.GetSubscription(cr.Namespace, r.operatorName+"-subscription")
	catalogsource := reconcilers.GetCatalogSource(reconcilers.CATALOG_NAMESPACE, r.operatorName+"-catalogsource")
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, subscription, func() error {
		if err := ctrl.SetControllerReference(cr, subscription, r.scheme); err != nil {
			return err
		}
		subscription.Spec = &v1alpha1.SubscriptionSpec{
			CatalogSource:          catalogsource.Name,
			CatalogSourceNamespace: catalogsource.Namespace,
			Package:                r.operatorPackageName,
			Channel:                r.operatorChanel,
			InstallPlanApproval:    v1alpha1.ApprovalAutomatic,
		}
		if cr.Spec.SyncPeriod != nil {
			subscription.Spec.Config = &v1alpha1.SubscriptionConfig{
				Env: []corev1.EnvVar{
					{
						Name:  "SYNC-PERIOD-MIN",
						Value: strconv.Itoa(*cr.Spec.SyncPeriod),
					},
				},
			}
		}

		return nil
	})

	if err != nil {
		return v1.ResultFailed, err
	}
	return v1.ResultSuccess, nil
}
func (r *Reconciler) reconcileOperatorGroup(ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	operatorgroup := reconcilers.GetOperatorGroup(reconcilers.INSTALL_NAMESPACE, "global-operators")
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, operatorgroup, func() error {
		operatorgroup.Spec = coreosv1.OperatorGroupSpec{}

		return nil
	})
	if err != nil {
		return v1.ResultFailed, err
	}

	return v1.ResultSuccess, nil
}
func (r *Reconciler) reconcileCatalogSource(ctx context.Context) (v1.PlatformsInstlnStatus, error) {
	catalogsource := reconcilers.GetCatalogSource(reconcilers.CATALOG_NAMESPACE, r.operatorName+"-catalogsource")
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, catalogsource, func() error {
		catalogsource.Spec = v1alpha1.CatalogSourceSpec{
			SourceType:  v1alpha1.SourceTypeGrpc,
			Image:       r.operatorCatalogImage,
			DisplayName: r.operatorDisplayName,
		}
		return nil
	})
	if err != nil {
		return v1.ResultFailed, err
	}
	return v1.ResultSuccess, nil
}

func (r *Reconciler) waitForOperator(cr *v1.DBaaSPlatform, ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	deployments := &apiv1.DeploymentList{}
	opts := &client.ListOptions{
		Namespace: cr.Namespace,
	}
	err := r.client.List(ctx, deployments, opts)
	if err != nil {
		return v1.ResultFailed, err
	}

	for _, deployment := range deployments.Items {
		if deployment.Name == r.operatorDeploymentName {
			if deployment.Status.ReadyReplicas > 0 {
				return v1.ResultSuccess, nil
			}
		}
	}
	return v1.ResultInProgress, nil
}

func (r *Reconciler) reconcileCSV(cr *v1.DBaaSPlatform, ctx context.Context) (v1.PlatformsInstlnStatus, error) {
	csv := reconcilers.GetClusterServiceVersion(cr.Namespace, r.operatorCSV)
	if err := r.client.Get(ctx, client.ObjectKeyFromObject(csv), csv); err != nil {
		if errors.IsNotFound(err) {
			return v1.ResultInProgress, nil
		}
		return v1.ResultFailed, err
	}

	if set, err := reconcilers.CheckOwnerReferenceSet(cr, csv, r.scheme); err != nil {
		return v1.ResultFailed, err
	} else if set {
		return v1.ResultSuccess, nil
	}

	if err := ctrl.SetControllerReference(cr, csv, r.scheme); err != nil {
		return v1.ResultFailed, err
	}
	if err := r.client.Update(ctx, csv); err != nil {
		return v1.ResultFailed, err
	}
	return v1.ResultInProgress, nil
}
