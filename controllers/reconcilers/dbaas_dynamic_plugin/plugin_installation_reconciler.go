package dbaas_dynamic_plugin

import (
	"context"
	v1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client client.Client
	logger logr.Logger
}

func NewReconciler(client client.Client, logger logr.Logger) reconcilers.PlatformReconciler {
	return &Reconciler{
		client: client,
		logger: logger,
	}
}
func (r *Reconciler) Reconcile(ctx context.Context, cr *v1.DBaaSPlatform, status2 *v1.DBaaSPlatformStatus) (v1.PlatformsInstlnStatus, error) {
	return v1.ResultSuccess, nil
}

func (r *Reconciler) Cleanup(ctx context.Context, cr *v1.DBaaSPlatform) (v1.PlatformsInstlnStatus, error) {
	return v1.ResultSuccess, nil

}
