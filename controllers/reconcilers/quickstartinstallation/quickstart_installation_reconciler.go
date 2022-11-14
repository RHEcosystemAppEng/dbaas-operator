package quickstartinstallation

import (
	"context"
	_ "embed"

	consolev1 "github.com/openshift/api/console/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
)

//go:embed accessing-the-database-access-menu-for-configuring-and-monitoring-quick-start.yaml
var adminQuickStart []byte

//go:embed accessing-the-developer-workspace-and-adding-a-database-instance-quick-start.yaml
var devInstanceQuickStart []byte

//go:embed connecting-an-application-to-a-database-instance-using-the-topology-view-quick-start.yaml
var devConnectQuickStart []byte

//go:embed installing-the-red-hat-openshift-database-access-add-on-quick-start.yaml
var installAddonQuickStart []byte

var quickStarts = map[string][]byte{
	"accessing-the-database-access-menu-for-configuring-and-monitoring":        adminQuickStart,
	"accessing-the-developer-workspace-and-adding-a-database-instance":         devInstanceQuickStart,
	"connecting-an-application-to-a-database-instance-using-the-topology-view": devConnectQuickStart,
	"installing-the-red-hat-openshift-database-access-add-on":                  installAddonQuickStart,
}

type reconciler struct {
	client client.Client
	logger logr.Logger
	scheme *runtime.Scheme
}

// NewReconciler returns a quickstartinstallation reconciler
func NewReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger) reconcilers.PlatformReconciler {
	return &reconciler{
		client: client,
		scheme: scheme,
		logger: logger,
	}
}

// Reconcile reconciles a quickstart platform
func (r *reconciler) Reconcile(ctx context.Context, _ *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	for qsName, qsBytes := range quickStarts {
		status, err := r.createQuickStartCR(ctx, qsName, qsBytes)
		if status != v1beta1.ResultSuccess {
			return status, err
		}
	}
	return v1beta1.ResultSuccess, nil
}

func (r *reconciler) createQuickStartCR(ctx context.Context, qsName string, qsBytes []byte) (v1beta1.PlatformInstlnStatus, error) {
	quickStart := r.getQuickStartModel(qsName)
	quickStartFromFile := &consolev1.ConsoleQuickStart{}
	err := yaml.Unmarshal(qsBytes, quickStartFromFile)

	if err == nil {
		_, err = controllerutil.CreateOrUpdate(ctx, r.client, quickStart, func() error {
			quickStart.Annotations = map[string]string{
				"categories": "Database management",
			}
			quickStart.Spec = quickStartFromFile.Spec
			return nil
		})
	}

	if err != nil {
		if errors.IsConflict(err) {
			return v1beta1.ResultInProgress, nil
		}
		return v1beta1.ResultFailed, err
	}
	return v1beta1.ResultSuccess, nil
}

func (r *reconciler) getQuickStartModel(name string) *consolev1.ConsoleQuickStart {
	return &consolev1.ConsoleQuickStart{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func (r *reconciler) Cleanup(ctx context.Context, _ *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	for qsName := range quickStarts {
		quickstart := r.getQuickStartModel(qsName)
		err := r.client.Delete(ctx, quickstart)
		if err != nil && !errors.IsNotFound(err) {
			return v1beta1.ResultFailed, err
		}
	}
	return v1beta1.ResultSuccess, nil
}
