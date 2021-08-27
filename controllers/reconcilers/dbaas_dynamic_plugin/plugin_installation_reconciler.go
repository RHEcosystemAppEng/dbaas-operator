package dbaas_dynamic_plugin

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"

	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	operatorv1 "github.com/openshift/api/operator/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
)

const (
	DBaaSDynamicPluginName       = "dbaas-dynamic-plugin"
	consoleServingCertSecretName = "console-serving-cert"
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
func (r *Reconciler) Reconcile(ctx context.Context, cr *v1alpha1.DBaaSPlatform, status2 *v1alpha1.DBaaSPlatformStatus) (v1alpha1.PlatformsInstlnStatus, error) {
	status, err := r.reconcileNamespace(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	status, err = r.reconcileService(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	status, err = r.reconcileDeployment(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	status, err = r.reconcileConsolePlugin(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	status, err = r.reconcileConsole(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	status, err = r.waitForConsole(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) Cleanup(ctx context.Context, cr *v1alpha1.DBaaSPlatform) (v1alpha1.PlatformsInstlnStatus, error) {
	console := GetOperatorConsole()
	err := r.client.Get(ctx, client.ObjectKeyFromObject(console), console)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}
	console.Spec.Plugins = removeDBaaSDynamicPlugin(console.Spec.Plugins)
	err = r.client.Update(ctx, console)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}

	plugin := GetDBaaSDynamicPluginConsolePlugin()
	err = r.client.Delete(ctx, plugin)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	deployment := GetDBaaSDynamicPluginDeployment()
	err = r.client.Delete(ctx, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	service := GetDBaaSDynamicPluginService()
	err = r.client.Delete(ctx, service)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	namespace := GetDBaaSDynamicPluginNamespace()
	err = r.client.Delete(ctx, namespace)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileNamespace(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	namespace := GetDBaaSDynamicPluginNamespace()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, namespace, func() error {
		return nil
	})

	if err != nil {
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileService(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	service := GetDBaaSDynamicPluginService()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, service, func() error {
		service.Annotations = map[string]string{
			"service.alpha.openshift.io/serving-cert-secret-name": consoleServingCertSecretName,
		}
		service.Labels = map[string]string{
			"app":                         DBaaSDynamicPluginName,
			"app.kubernetes.io/component": DBaaSDynamicPluginName,
			"app.kubernetes.io/instance":  DBaaSDynamicPluginName,
			"app.kubernetes.io/part-of":   DBaaSDynamicPluginName,
		}
		service.Spec.Ports = []v1.ServicePort{
			{
				Name:       "9001-tcp",
				Protocol:   v1.ProtocolTCP,
				Port:       int32(9001),
				TargetPort: intstr.FromInt(9001),
			},
		}
		service.Spec.Selector = map[string]string{
			"app": DBaaSDynamicPluginName,
		}
		service.Spec.Type = v1.ServiceTypeClusterIP
		service.Spec.SessionAffinity = v1.ServiceAffinityNone
		return nil
	})

	if err != nil {
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileDeployment(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	deployment := GetDBaaSDynamicPluginDeployment()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, deployment, func() error {
		deployment.Labels = map[string]string{
			"app":                                DBaaSDynamicPluginName,
			"app.kubernetes.io/component":        DBaaSDynamicPluginName,
			"app.kubernetes.io/instance":         DBaaSDynamicPluginName,
			"app.kubernetes.io/part-of":          DBaaSDynamicPluginName,
			"app.openshift.io/runtime-namespace": DBaaSDynamicPluginName,
		}
		replicas := int32(1)
		defaultMode := int32(420)
		percentageOfPods := intstr.FromString("25%")
		deployment.Spec.Replicas = &replicas
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": DBaaSDynamicPluginName,
			},
		}
		deployment.Spec.Template.ObjectMeta = metav1.ObjectMeta{
			Labels: map[string]string{
				"app": DBaaSDynamicPluginName,
			},
		}
		deployment.Spec.Template.Spec.Containers = []v1.Container{
			{
				Name:  DBaaSDynamicPluginName,
				Image: reconcilers.DBAAS_DYNAMIC_PLUGIN_IMG,
				Ports: []v1.ContainerPort{
					{
						ContainerPort: 9001,
						Protocol:      v1.ProtocolTCP,
					},
				},
				ImagePullPolicy: v1.PullAlways,
				Args: []string{
					"--ssl",
					"--cert=/var/serving-cert/tls.crt",
					"--key=/var/serving-cert/tls.key",
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      consoleServingCertSecretName,
						ReadOnly:  true,
						MountPath: "/var/serving-cert",
					},
				},
			},
		}
		deployment.Spec.Template.Spec.Volumes = []v1.Volume{
			{
				Name: consoleServingCertSecretName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName:  consoleServingCertSecretName,
						DefaultMode: &defaultMode,
					},
				},
			},
		}
		deployment.Spec.Template.Spec.RestartPolicy = v1.RestartPolicyAlways
		deployment.Spec.Template.Spec.DNSPolicy = v1.DNSClusterFirst
		deployment.Spec.Strategy = appv1.DeploymentStrategy{
			Type: appv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appv1.RollingUpdateDeployment{
				MaxUnavailable: &percentageOfPods,
				MaxSurge:       &percentageOfPods,
			},
		}
		return nil
	})

	if err != nil {
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileConsolePlugin(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	plugin := GetDBaaSDynamicPluginConsolePlugin()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, plugin, func() error {
		plugin.Spec.DisplayName = "OpenShift DataBase as a Service Dynamic Plugin"
		plugin.Spec.Service = consolev1alpha1.ConsolePluginService{
			Name:      DBaaSDynamicPluginName,
			Namespace: DBaaSDynamicPluginName,
			Port:      int32(9001),
			BasePath:  "/",
		}
		return nil
	})

	if err != nil {
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileConsole(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	console := GetOperatorConsole()
	err := r.client.Get(ctx, client.ObjectKeyFromObject(console), console)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}

	if console.Spec.Plugins == nil {
		console.Spec.Plugins = []string{DBaaSDynamicPluginName}
	} else {
		console.Spec.Plugins = addDBaaSDynamicPlugin(console.Spec.Plugins)
	}
	err = r.client.Update(ctx, console)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) waitForConsole(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	console := GetOperatorConsole()
	err := r.client.Get(ctx, client.ObjectKeyFromObject(console), console)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}

	if console.Status.Conditions != nil {
		for _, condition := range console.Status.Conditions {
			if condition.Type == "DeploymentAvailable" {
				if condition.Status == operatorv1.ConditionTrue {
					return v1alpha1.ResultSuccess, nil
				}
				break
			}
		}
	}

	return v1alpha1.ResultInProgress, nil
}

func GetDBaaSDynamicPluginNamespace() *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: DBaaSDynamicPluginName,
		},
	}
}

func GetDBaaSDynamicPluginService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DBaaSDynamicPluginName,
			Namespace: DBaaSDynamicPluginName,
		},
	}
}

func GetDBaaSDynamicPluginDeployment() *appv1.Deployment {
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DBaaSDynamicPluginName,
			Namespace: DBaaSDynamicPluginName,
		},
	}
}

func GetDBaaSDynamicPluginConsolePlugin() *consolev1alpha1.ConsolePlugin {
	return &consolev1alpha1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: DBaaSDynamicPluginName,
		},
	}
}

func GetOperatorConsole() *operatorv1.Console {
	return &operatorv1.Console{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
	}
}

func addDBaaSDynamicPlugin(plugins []string) []string {
	for _, p := range plugins {
		if p == DBaaSDynamicPluginName {
			return plugins
		}
	}

	return append(plugins, DBaaSDynamicPluginName)
}

func removeDBaaSDynamicPlugin(plugins []string) []string {
	for i, p := range plugins {
		if p == DBaaSDynamicPluginName {
			return append(plugins[:i], plugins[i+1:]...)
		}
	}

	return plugins
}
