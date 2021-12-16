package mongodb_atlas_instalation

import (
	"context"

	v1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	"github.com/go-logr/logr"
	coreosv1 "github.com/operator-framework/api/pkg/operators/v1"
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

	//MongoDBAtlas CatalogSource
	status, err := r.reconcileCatalogSource(ctx)
	if status != v1.ResultSuccess {
		return status, err
	}

	// MongoDBAtlas subscription
	status, err = r.reconcileSubscription(cr, ctx)
	if status != v1.ResultSuccess {
		return status, err
	}
	// MongoDBAtlas operator group
	status, err = r.reconcileOperatorGroup(ctx)
	if status != v1.ResultSuccess {
		return status, err
	}
	status, err = r.waitForMongoDBAtlasOperator(cr, ctx)
	if status != v1.ResultSuccess {
		return status, err
	}
	return v1.ResultSuccess, nil

}

func (r *Reconciler) Cleanup(ctx context.Context, cr *v1.DBaaSPlatform) (v1.PlatformsInstlnStatus, error) {

	subscription := r.getMongoDBAtlasSubscription(cr)
	err := r.client.Delete(ctx, subscription)
	if err != nil && !errors.IsNotFound(err) {
		return v1.ResultFailed, err
	}

	catalogSource := r.getMongoDBAtlasCatalogSource()
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
		if deployments.Items[d].Name == "mongodb-atlas-operator" {
			err = r.client.Delete(ctx, &deployments.Items[d])
			if err != nil && !errors.IsNotFound(err) {
				return v1.ResultFailed, err
			}
		}
	}

	return v1.ResultSuccess, nil
}
func (r *Reconciler) reconcileSubscription(cr *v1.DBaaSPlatform, ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	subscription := r.getMongoDBAtlasSubscription(cr)
	catalogsource := r.getMongoDBAtlasCatalogSource()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, subscription, func() error {
		if err := ctrl.SetControllerReference(cr, subscription, r.scheme); err != nil {
			return err
		}
		subscription.Spec = &v1alpha1.SubscriptionSpec{
			CatalogSource:          catalogsource.Name,
			CatalogSourceNamespace: catalogsource.Namespace,
			Package:                "mongodb-atlas-kubernetes",
			Channel:                "beta",
			InstallPlanApproval:    v1alpha1.ApprovalAutomatic,
		}

		return nil
	})

	if err != nil {
		return v1.ResultFailed, err
	}
	return v1.ResultSuccess, nil
}
func (r *Reconciler) reconcileOperatorGroup(ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	operatorgroup := r.getMongoDBAtlasOperatorGroup()
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
	catalogsource := r.getMongoDBAtlasCatalogSource()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, catalogsource, func() error {
		catalogsource.Spec = v1alpha1.CatalogSourceSpec{
			SourceType:  v1alpha1.SourceTypeGrpc,
			Image:       reconcilers.MONGODB_ATLAS_CATLOG_IMG,
			DisplayName: "MongoDB Atlas Operator",
		}
		return nil
	})
	if err != nil {
		return v1.ResultFailed, err
	}
	return v1.ResultSuccess, nil
}

func (r *Reconciler) waitForMongoDBAtlasOperator(cr *v1.DBaaSPlatform, ctx context.Context) (v1.PlatformsInstlnStatus, error) {

	deployments := &apiv1.DeploymentList{}
	opts := &client.ListOptions{
		Namespace: cr.Namespace,
	}
	err := r.client.List(ctx, deployments, opts)
	if err != nil {
		return v1.ResultFailed, err
	}

	for _, deployment := range deployments.Items {
		if deployment.Name == "mongodb-atlas-operator" {
			if deployment.Status.ReadyReplicas > 0 {
				return v1.ResultSuccess, nil
			}
		}
	}
	return v1.ResultInProgress, nil
}

func (r *Reconciler) getMongoDBAtlasSubscription(cr *v1.DBaaSPlatform) *v1alpha1.Subscription {
	return &v1alpha1.Subscription{
		ObjectMeta: apimv1.ObjectMeta{
			Name:      "mongodb-atlas-subscription",
			Namespace: cr.Namespace,
		},
	}
}
func (r *Reconciler) getMongoDBAtlasOperatorGroup() *coreosv1.OperatorGroup {
	return &coreosv1.OperatorGroup{
		ObjectMeta: apimv1.ObjectMeta{
			Name:      "global-operators",
			Namespace: reconcilers.INSTALL_NAMESPACE,
		},
	}
}

func (r *Reconciler) getMongoDBAtlasCatalogSource() *v1alpha1.CatalogSource {
	return &v1alpha1.CatalogSource{
		ObjectMeta: apimv1.ObjectMeta{
			Name:      "mongodb-atlas-catalogsource",
			Namespace: reconcilers.CATALOG_NAMESPACE,
		},
	}
}
