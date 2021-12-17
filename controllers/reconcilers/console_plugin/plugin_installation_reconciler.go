package console_plugin

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
	consoleServingCertSecretName = "console-serving-cert"
)

type Reconciler struct {
	client          client.Client
	logger          logr.Logger
	pluginName      string
	pluginNamespace string
	pluginImage     string
	displayName     string
	envs            []v1.EnvVar
}

func NewReconciler(client client.Client, logger logr.Logger, pluginName string, pluginNamespace string, pluginImage string, displayName string, envs ...v1.EnvVar) reconcilers.PlatformReconciler {
	return &Reconciler{
		client:          client,
		logger:          logger,
		pluginName:      pluginName,
		pluginNamespace: pluginNamespace,
		pluginImage:     pluginImage,
		displayName:     displayName,
		envs:            envs,
	}
}
func (r *Reconciler) Reconcile(ctx context.Context, cr *v1alpha1.DBaaSPlatform, status2 *v1alpha1.DBaaSPlatformStatus) (v1alpha1.PlatformsInstlnStatus, error) {
	status, err := r.reconcileService(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	status, err = r.reconcileDeployment(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	status, err = r.waitForConsolePlugin(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	// create Console Plugin CR resource that includes Console Plugin service name.
	status, err = r.createConsolePluginCR(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	// enabled console plugins the console operator config
	status, err = r.enableConsolePluginConfig(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	status, err = r.waitForConsoleOperator(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) Cleanup(ctx context.Context, cr *v1alpha1.DBaaSPlatform) (v1alpha1.PlatformsInstlnStatus, error) {
	console := r.getOperatorConsole()
	err := r.client.Get(ctx, client.ObjectKeyFromObject(console), console)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}
	console.Spec.Plugins = r.removePlugin(console.Spec.Plugins)
	err = r.client.Update(ctx, console)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}

	plugin := r.getConsolePlugin()
	err = r.client.Delete(ctx, plugin)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	deployment := r.getDeployment()
	err = r.client.Delete(ctx, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	service := r.getService()
	err = r.client.Delete(ctx, service)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileService(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	service := r.getService()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, service, func() error {
		service.Annotations = map[string]string{
			"service.beta.openshift.io/serving-cert-secret-name": consoleServingCertSecretName,
		}
		service.Labels = map[string]string{
			"app":                         r.pluginName,
			"app.kubernetes.io/component": r.pluginName,
			"app.kubernetes.io/instance":  r.pluginName,
			"app.kubernetes.io/part-of":   r.pluginName,
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
			"app": r.pluginName,
		}
		service.Spec.Type = v1.ServiceTypeClusterIP
		service.Spec.SessionAffinity = v1.ServiceAffinityNone
		return nil
	})

	if err != nil {
		if errors.IsConflict(err) {
			return v1alpha1.ResultInProgress, nil
		}
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileDeployment(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	deployment := r.getDeployment()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, deployment, func() error {
		deployment.Labels = map[string]string{
			"app":                                r.pluginName,
			"app.kubernetes.io/component":        r.pluginName,
			"app.kubernetes.io/instance":         r.pluginName,
			"app.kubernetes.io/part-of":          r.pluginName,
			"app.openshift.io/runtime-namespace": r.pluginNamespace,
		}
		replicas := int32(3)
		defaultMode := int32(420)
		percentageOfPods := intstr.FromString("25%")
		deployment.Spec.Replicas = &replicas
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": r.pluginName,
			},
		}
		deployment.Spec.Template.ObjectMeta = metav1.ObjectMeta{
			Labels: map[string]string{
				"app": r.pluginName,
			},
		}
		deployment.Spec.Template.Spec.Containers = []v1.Container{
			{
				Name:  r.pluginName,
				Image: r.pluginImage,
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
				Env: r.envs,
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
		if errors.IsConflict(err) {
			return v1alpha1.ResultInProgress, nil
		}
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) createConsolePluginCR(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	plugin := r.getConsolePlugin()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, plugin, func() error {
		plugin.Spec.DisplayName = r.displayName
		plugin.Spec.Service = consolev1alpha1.ConsolePluginService{
			Name:      r.pluginName,
			Namespace: r.pluginNamespace,
			Port:      int32(9001),
			BasePath:  "/",
		}
		return nil
	})

	if err != nil {
		if errors.IsConflict(err) {
			return v1alpha1.ResultInProgress, nil
		}
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) enableConsolePluginConfig(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	console := r.getOperatorConsole()
	err := r.client.Get(ctx, client.ObjectKeyFromObject(console), console)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}

	console.Spec.Plugins = r.addPlugin(console.Spec.Plugins)
	err = r.client.Update(ctx, console)
	if err != nil {
		if errors.IsConflict(err) {
			return v1alpha1.ResultInProgress, nil
		}
		return v1alpha1.ResultFailed, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) waitForConsolePlugin(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	deployments := &appv1.DeploymentList{}
	opts := &client.ListOptions{
		Namespace: r.pluginNamespace,
	}
	err := r.client.List(ctx, deployments, opts)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}

	for _, deployment := range deployments.Items {
		if deployment.Name == r.pluginName {
			if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
				return v1alpha1.ResultSuccess, nil
			}
		}
	}
	return v1alpha1.ResultInProgress, nil
}

func (r *Reconciler) waitForConsoleOperator(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	console := r.getOperatorConsole()
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

func (r *Reconciler) getService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.pluginName,
			Namespace: r.pluginNamespace,
		},
	}
}

func (r *Reconciler) getDeployment() *appv1.Deployment {
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.pluginName,
			Namespace: r.pluginNamespace,
		},
	}
}

func (r *Reconciler) getConsolePlugin() *consolev1alpha1.ConsolePlugin {
	return &consolev1alpha1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.pluginName,
		},
	}
}

func (r *Reconciler) getOperatorConsole() *operatorv1.Console {
	return &operatorv1.Console{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
	}
}

func (r *Reconciler) addPlugin(plugins []string) []string {
	for _, p := range plugins {
		if p == r.pluginName {
			return plugins
		}
	}

	return append(plugins, r.pluginName)
}

func (r *Reconciler) removePlugin(plugins []string) []string {
	for i, p := range plugins {
		if p == r.pluginName {
			return append(plugins[:i], plugins[i+1:]...)
		}
	}

	return plugins
}
