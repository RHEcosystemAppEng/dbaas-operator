package quickstart_installation

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

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
)

//go:embed getting-started-with-openshift-database-access-for-administrators-quick-start.yaml
var gettingstartedadminQuickStart []byte

//go:embed getting-started-with-openshift-database-access-for-developers-quick-start.yaml
var gettingstarteddevQuickStart []byte

var QuickStarts = map[string][]byte{
	"getting-started-with-openshift-database-access-for-administrators": gettingstartedadminQuickStart,
	"getting-started-with-openshift-database-access-for-developers":     gettingstarteddevQuickStart,
}

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

func (r *Reconciler) Reconcile(ctx context.Context, cr *v1alpha1.DBaaSPlatform, platformStatus *v1alpha1.DBaaSPlatformStatus) (v1alpha1.PlatformsInstlnStatus, error) {
	for qsName, qsBytes := range QuickStarts {
		status, err := r.createQuickStartCR(qsName, qsBytes, ctx)
		if status != v1alpha1.ResultSuccess {
			return status, err
		}
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) createQuickStartCR(qsName string, qsBytes []byte, ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
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
			return v1alpha1.ResultInProgress, nil
		}
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) getQuickStartModel(name string) *consolev1.ConsoleQuickStart {
	return &consolev1.ConsoleQuickStart{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func (r *Reconciler) Cleanup(ctx context.Context, cr *v1alpha1.DBaaSPlatform) (v1alpha1.PlatformsInstlnStatus, error) {
	for qsName := range QuickStarts {
		quickstart := r.getQuickStartModel(qsName)
		err := r.client.Delete(ctx, quickstart)
		if err != nil && !errors.IsNotFound(err) {
			return v1alpha1.ResultFailed, err
		}
	}
	return v1alpha1.ResultSuccess, nil
}
