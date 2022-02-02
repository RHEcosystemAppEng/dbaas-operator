package console_plugin

import (
	"context"
	"strconv"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
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
	consolePort                  = 9001
)

type Reconciler struct {
	client      client.Client
	logger      logr.Logger
	scheme      *runtime.Scheme
	pluginName  string
	pluginImage string
	displayName string
	envs        []v1.EnvVar
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger, pluginName string, pluginImage string, displayName string, envs ...v1.EnvVar) reconcilers.PlatformReconciler {
	return &Reconciler{
		client:      client,
		scheme:      scheme,
		logger:      logger,
		pluginName:  pluginName,
		pluginImage: pluginImage,
		displayName: displayName,
		envs:        envs,
	}
}
func (r *Reconciler) Reconcile(ctx context.Context, cr *v1alpha1.DBaaSPlatform, status2 *v1alpha1.DBaaSPlatformStatus) (v1alpha1.PlatformsInstlnStatus, error) {
	status, err := r.reconcileService(cr, ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	status, err = r.reconcileDeployment(cr, ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	// create Console Plugin CR resource that includes Console Plugin service name.
	status, err = r.createConsolePluginCR(cr, ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}
	// enabled console plugins the console operator config
	status, err = r.enableConsolePluginConfig(ctx)
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

	deployment := r.getDeployment(cr)
	err = r.client.Delete(ctx, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	service := r.getService(cr)
	err = r.client.Delete(ctx, service)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileService(cr *v1alpha1.DBaaSPlatform, ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	service := r.getService(cr)
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, service, func() error {
		if err := ctrl.SetControllerReference(cr, service, r.scheme); err != nil {
			return err
		}
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
				Name:       strconv.Itoa(consolePort) + "-tcp",
				Protocol:   v1.ProtocolTCP,
				Port:       int32(consolePort),
				TargetPort: intstr.FromInt(consolePort),
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

func (r *Reconciler) reconcileDeployment(cr *v1alpha1.DBaaSPlatform, ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	deployment := r.getDeployment(cr)
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, deployment, func() error {
		if err := ctrl.SetControllerReference(cr, deployment, r.scheme); err != nil {
			return err
		}
		deployment.Labels = map[string]string{
			"app":                                r.pluginName,
			"app.kubernetes.io/component":        r.pluginName,
			"app.kubernetes.io/instance":         r.pluginName,
			"app.kubernetes.io/part-of":          r.pluginName,
			"app.openshift.io/runtime-namespace": cr.Namespace,
		}
		replicas := int32(3)
		defaultMode := int32(420)
		ptrTrue := true
		ptrFalse := false
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
		socketHandler := v1.Handler{
			TCPSocket: &v1.TCPSocketAction{Port: intstr.FromInt(consolePort)},
		}
		deployment.Spec.Template.Spec.Containers = []v1.Container{
			{
				Name:  r.pluginName,
				Image: r.pluginImage,
				Ports: []v1.ContainerPort{
					{
						ContainerPort: consolePort,
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
				SecurityContext: &v1.SecurityContext{
					AllowPrivilegeEscalation: &ptrFalse,
					Capabilities: &v1.Capabilities{
						Drop: []v1.Capability{"ALL"},
					},
				},
				LivenessProbe: &v1.Probe{
					Handler:             socketHandler,
					InitialDelaySeconds: 5,
				},
				ReadinessProbe: &v1.Probe{
					Handler:             socketHandler,
					InitialDelaySeconds: 30,
					PeriodSeconds:       20,
				},
			},
		}
		deployment.Spec.Template.Spec.SecurityContext = &v1.PodSecurityContext{
			RunAsNonRoot: &ptrTrue,
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

	err = r.client.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return v1alpha1.ResultInProgress, nil
		}
		return v1alpha1.ResultFailed, err
	}
	if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
		return v1alpha1.ResultSuccess, nil
	}
	return v1alpha1.ResultInProgress, nil
}

func (r *Reconciler) createConsolePluginCR(cr *v1alpha1.DBaaSPlatform, ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	plugin := r.getConsolePlugin()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, plugin, func() error {
		plugin.Spec.DisplayName = r.displayName
		plugin.Spec.Service = consolev1alpha1.ConsolePluginService{
			Name:      r.pluginName,
			Namespace: cr.Namespace,
			Port:      int32(consolePort),
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

	if plugins, add := r.addPlugin(console.Spec.Plugins); add {
		console.Spec.Plugins = plugins
		err := r.client.Update(ctx, console)
		if err != nil {
			if errors.IsConflict(err) {
				return v1alpha1.ResultInProgress, nil
			}
			return v1alpha1.ResultFailed, err
		}
		return v1alpha1.ResultInProgress, nil
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

func (r *Reconciler) getService(cr *v1alpha1.DBaaSPlatform) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.pluginName,
			Namespace: cr.Namespace,
		},
	}
}

func (r *Reconciler) getDeployment(cr *v1alpha1.DBaaSPlatform) *appv1.Deployment {
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.pluginName,
			Namespace: cr.Namespace,
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

func (r *Reconciler) addPlugin(plugins []string) ([]string, bool) {
	for _, p := range plugins {
		if p == r.pluginName {
			return plugins, false
		}
	}

	return append(plugins, r.pluginName), true
}

func (r *Reconciler) removePlugin(plugins []string) []string {
	for i, p := range plugins {
		if p == r.pluginName {
			return append(plugins[:i], plugins[i+1:]...)
		}
	}

	return plugins
}
