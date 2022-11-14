package consoleplugin

import (
	"context"
	"strconv"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	"github.com/go-logr/logr"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	operatorv1 "github.com/openshift/api/operator/v1"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	serviceCertPrefix = "serve-cert-"
	consolePort       = 9001
)

type reconciler struct {
	client client.Client
	logger logr.Logger
	scheme *runtime.Scheme
	config v1beta1.PlatformConfig
}

// NewReconciler returns a plugin installation reconciler
func NewReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger, config v1beta1.PlatformConfig) reconcilers.PlatformReconciler {
	return &reconciler{
		client: client,
		scheme: scheme,
		logger: logger,
		config: config,
	}
}

// Reconcile deploys the dynamic console plugin
func (r *reconciler) Reconcile(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	status, err := r.reconcileService(ctx, cr)
	if status != v1beta1.ResultSuccess {
		return status, err
	}
	status, err = r.reconcileDeployment(ctx, cr)
	if status != v1beta1.ResultSuccess {
		return status, err
	}

	// create Console Plugin CR resource that includes Console Plugin service name.
	status, err = r.createConsolePluginCR(ctx, cr)
	if status != v1beta1.ResultSuccess {
		return status, err
	}
	// enabled console plugins the console operator config
	status, err = r.enableConsolePluginConfig(ctx)
	if status != v1beta1.ResultSuccess {
		return status, err
	}

	return v1beta1.ResultSuccess, nil
}

// Cleanup cleanup resources related to the console plugin
func (r *reconciler) Cleanup(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	console := r.getOperatorConsole()
	err := r.client.Get(ctx, client.ObjectKeyFromObject(console), console)
	if err != nil {
		return v1beta1.ResultFailed, err
	}
	console.Spec.Plugins = r.removePlugin(console.Spec.Plugins)
	err = r.client.Update(ctx, console)
	if err != nil {
		return v1beta1.ResultFailed, err
	}

	plugin := r.getConsolePlugin()
	err = r.client.Delete(ctx, plugin)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}

	deployment := r.getDeployment(cr)
	err = r.client.Delete(ctx, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}

	service := r.getService(cr)
	err = r.client.Delete(ctx, service)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}

	return v1beta1.ResultSuccess, nil
}

func (r *reconciler) reconcileService(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	service := r.getService(cr)
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, service, func() error {
		if err := ctrl.SetControllerReference(cr, service, r.scheme); err != nil {
			return err
		}
		service.Annotations = map[string]string{
			"service.beta.openshift.io/serving-cert-secret-name": serviceCertPrefix + r.config.Name,
		}
		service.Labels = map[string]string{
			"app":                         r.config.Name,
			"app.kubernetes.io/component": r.config.Name,
			"app.kubernetes.io/instance":  r.config.Name,
			"app.kubernetes.io/part-of":   r.config.Name,
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
			"app": r.config.Name,
		}
		service.Spec.Type = v1.ServiceTypeClusterIP
		service.Spec.SessionAffinity = v1.ServiceAffinityNone
		return nil
	})

	if err != nil {
		if errors.IsConflict(err) {
			return v1beta1.ResultInProgress, nil
		}
		return v1beta1.ResultFailed, err
	}
	return v1beta1.ResultSuccess, nil
}

func (r *reconciler) reconcileDeployment(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	deployment := r.getDeployment(cr)
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, deployment, func() error {
		if err := ctrl.SetControllerReference(cr, deployment, r.scheme); err != nil {
			return err
		}
		deployment.Labels = map[string]string{
			"app":                                r.config.Name,
			"app.kubernetes.io/component":        r.config.Name,
			"app.kubernetes.io/instance":         r.config.Name,
			"app.kubernetes.io/part-of":          r.config.Name,
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
				"app": r.config.Name,
			},
		}
		deployment.Spec.Template.ObjectMeta = metav1.ObjectMeta{
			Labels: map[string]string{
				"app": r.config.Name,
			},
		}
		socketHandler := v1.ProbeHandler{
			TCPSocket: &v1.TCPSocketAction{Port: intstr.FromInt(consolePort)},
		}
		deployment.Spec.Template.Spec.Containers = []v1.Container{
			{
				Name:  r.config.Name,
				Image: r.config.Image,
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
						Name:      serviceCertPrefix + r.config.Name,
						ReadOnly:  true,
						MountPath: "/var/serving-cert",
					},
				},
				Env: r.config.Envs,
				SecurityContext: &v1.SecurityContext{
					AllowPrivilegeEscalation: &ptrFalse,
					Capabilities: &v1.Capabilities{
						Drop: []v1.Capability{"ALL"},
					},
					ReadOnlyRootFilesystem: &ptrTrue,
					RunAsNonRoot:           &ptrTrue,
				},
				LivenessProbe: &v1.Probe{
					ProbeHandler:        socketHandler,
					InitialDelaySeconds: 5,
				},
				ReadinessProbe: &v1.Probe{
					ProbeHandler:        socketHandler,
					InitialDelaySeconds: 30,
					PeriodSeconds:       20,
				},
			},
		}
		deployment.Spec.Template.Spec.Volumes = []v1.Volume{
			{
				Name: serviceCertPrefix + r.config.Name,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName:  serviceCertPrefix + r.config.Name,
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
			return v1beta1.ResultInProgress, nil
		}
		return v1beta1.ResultFailed, err
	}

	err = r.client.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return v1beta1.ResultInProgress, nil
		}
		return v1beta1.ResultFailed, err
	}
	if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
		return v1beta1.ResultSuccess, nil
	}
	return v1beta1.ResultInProgress, nil
}

func (r *reconciler) createConsolePluginCR(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	plugin := r.getConsolePlugin()
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, plugin, func() error {
		plugin.Spec.DisplayName = r.config.DisplayName
		plugin.Spec.Service = consolev1alpha1.ConsolePluginService{
			Name:      r.config.Name,
			Namespace: cr.Namespace,
			Port:      int32(consolePort),
			BasePath:  "/",
		}
		return nil
	})

	if err != nil {
		if errors.IsConflict(err) {
			return v1beta1.ResultInProgress, nil
		}
		return v1beta1.ResultFailed, err
	}
	return v1beta1.ResultSuccess, nil
}

func (r *reconciler) enableConsolePluginConfig(ctx context.Context) (v1beta1.PlatformInstlnStatus, error) {
	console := r.getOperatorConsole()
	err := r.client.Get(ctx, client.ObjectKeyFromObject(console), console)
	if err != nil {
		return v1beta1.ResultFailed, err
	}

	if plugins, add := r.addPlugin(console.Spec.Plugins); add {
		console.Spec.Plugins = plugins
		err := r.client.Update(ctx, console)
		if err != nil {
			if errors.IsConflict(err) {
				return v1beta1.ResultInProgress, nil
			}
			return v1beta1.ResultFailed, err
		}
		return v1beta1.ResultInProgress, nil
	}

	if console.Status.Conditions != nil {
		for _, condition := range console.Status.Conditions {
			if condition.Type == "DeploymentAvailable" {
				if condition.Status == operatorv1.ConditionTrue {
					return v1beta1.ResultSuccess, nil
				}
				break
			}
		}
	}
	return v1beta1.ResultInProgress, nil
}

func (r *reconciler) getService(cr *v1beta1.DBaaSPlatform) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.config.Name,
			Namespace: cr.Namespace,
		},
	}
}

func (r *reconciler) getDeployment(cr *v1beta1.DBaaSPlatform) *appv1.Deployment {
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.config.Name,
			Namespace: cr.Namespace,
		},
	}
}

func (r *reconciler) getConsolePlugin() *consolev1alpha1.ConsolePlugin {
	return &consolev1alpha1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.config.Name,
		},
	}
}

func (r *reconciler) getOperatorConsole() *operatorv1.Console {
	return &operatorv1.Console{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
	}
}

func (r *reconciler) addPlugin(plugins []string) ([]string, bool) {
	for _, p := range plugins {
		if p == r.config.Name {
			return plugins, false
		}
	}

	return append(plugins, r.config.Name), true
}

func (r *reconciler) removePlugin(plugins []string) []string {
	for i, p := range plugins {
		if p == r.config.Name {
			return append(plugins[:i], plugins[i+1:]...)
		}
	}

	return plugins
}
