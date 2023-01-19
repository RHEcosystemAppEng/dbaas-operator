package reconcilers

import (
	"context"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

// PlatformReconciler interface for platform reconcilers
type PlatformReconciler interface {
	Reconcile(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error)
	Cleanup(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error)
}
