package reconcilers

import (
	"context"
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

type PlatformReconciler interface {
	Reconcile(ctx context.Context, cr *v1alpha1.DBaaSPlatform, status *v1alpha1.DBaaSPlatformStatus) (v1alpha1.PlatformsInstlnStatus, error)
	Cleanup(ctx context.Context, cr *v1alpha1.DBaaSPlatform) (v1alpha1.PlatformsInstlnStatus, error)
}
