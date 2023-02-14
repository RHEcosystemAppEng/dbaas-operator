package observability

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/util"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	rhobsv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	msoapi "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	clusterIDLabel                     = "cluster_id"
	crNameForMonitoringStack           = "dbaas-operator-mso"
	crNameForServiceMonitor            = "dbaas-operator-service-monitor"
	rhobsRemoteWriteConfigIDKey        = "prom-remote-write-config-id"
	rhobsRemoteWriteConfigName         = "prom-remote-write-config-secret" //#nosec
	rhobsTokenKey                      = "rhobs-token"                     //#nosec
	authTypeDex                 string = "dex"
	authTypeRedhat              string = "redhat-sso"
	ServiceMonitorPeriod        string = "30s"
)

var metricsToInclude = []string{"dbaas_.*$", "csv_succeeded$", "csv_abnormal$", "ALERTS$", "subscription_sync_total"}
var replicas int32 = 1

type reconciler struct {
	client k8sclient.Client
	logger logr.Logger
	scheme *runtime.Scheme
	config v1beta1.PlatformConfig
}

// NewReconciler returns a plugin observability reconciler
func NewReconciler(client k8sclient.Client, scheme *runtime.Scheme, logger logr.Logger) reconcilers.PlatformReconciler {
	return &reconciler{
		client: client,
		scheme: scheme,
		logger: logger,
	}
}

// Reconcile create the CR for Observability Operator
func (r *reconciler) Reconcile(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {

	// clean the observability operator if installed by RHODA in previous version.
	status, err := r.cleanObsOldSubscription(ctx, cr)
	if status != v1beta1.ResultSuccess {
		return status, err
	}

	subscription := reconcilers.GetSubscription("openshift-observability-operator", "observability-operator")
	err = r.client.Get(ctx, k8sclient.ObjectKeyFromObject(subscription), subscription)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}
	if errors.IsNotFound(err) {
		return v1beta1.ResultSuccess, nil
	}

	// create or update observability CR.
	status, err = r.createUpdateObservabilityMonitoringStackCR(ctx, cr)
	if status != v1beta1.ResultSuccess {
		return status, err
	}

	// create observability ServiceMonitor CR.
	status, err = r.createObservabilityServiceMonitorCR(ctx, cr)
	if status != v1beta1.ResultSuccess {
		return status, err
	}
	return v1beta1.ResultSuccess, nil
}

func (r *reconciler) Cleanup(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {

	monitoringStackCR := getDefaultMonitoringStackCR(cr.Namespace)
	err := r.client.Delete(ctx, monitoringStackCR)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}

	serviceMonitorCR := getDefaultServiceMonitor(cr.Namespace)
	err = r.client.Delete(ctx, serviceMonitorCR)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}

	return v1beta1.ResultSuccess, nil

}

func (r *reconciler) cleanObsOldSubscription(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {

	subscription := reconcilers.GetSubscription(cr.Namespace, r.config.Name+"-subscription")
	err := r.client.Delete(ctx, subscription)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}

	csv := reconcilers.GetClusterServiceVersion(cr.Namespace, "observability-operator.v0.0.13")
	err = r.client.Delete(ctx, csv)
	if err != nil && !errors.IsNotFound(err) {
		return v1beta1.ResultFailed, err
	}

	return v1beta1.ResultSuccess, nil
}

func (r *reconciler) createUpdateObservabilityMonitoringStackCR(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	config := reconcilers.GetObservabilityConfig()

	r.logger.V(1).Info(" Observability Configuration", "AuthType", config.AuthType, "AddonName", config.AddonName)
	monitoringStackCR := getDefaultMonitoringStackCR(cr.Namespace)
	if _, err := controllerutil.CreateOrUpdate(ctx, r.client, monitoringStackCR, func() error {
		monitoringStackCR.Labels = map[string]string{
			"managed-by": "dbaas-operator",
		}
		monitoringStackCR.Spec = msoapi.MonitoringStackSpec{
			LogLevel: "debug",
			ResourceSelector: &metav1.LabelSelector{
				MatchLabels: setExporterLables(),
			},
		}
		if config.RemoteWritesURL != "" && config.AuthType != "" && config.AddonName != "" {
			prometheusConfig, _ := r.setPrometheusConfig(ctx, config, monitoringStackCR.Namespace)
			monitoringStackCR.Spec.PrometheusConfig = prometheusConfig
		}

		return controllerutil.SetControllerReference(cr, monitoringStackCR, r.scheme)

	}); err != nil {
		if errors.IsConflict(err) {
			return v1beta1.ResultInProgress, nil
		}
		return v1beta1.ResultFailed, err
	}

	return v1beta1.ResultSuccess, nil

}

func (r *reconciler) createObservabilityServiceMonitorCR(ctx context.Context, cr *v1beta1.DBaaSPlatform) (v1beta1.PlatformInstlnStatus, error) {
	monitoringServiceCR := getDefaultServiceMonitor(cr.Namespace)
	if _, err := controllerutil.CreateOrUpdate(ctx, r.client, monitoringServiceCR, func() error {
		return controllerutil.SetControllerReference(cr, monitoringServiceCR, r.scheme)
	}); err != nil {
		if errors.IsConflict(err) {
			return v1beta1.ResultInProgress, nil
		}
		return v1beta1.ResultFailed, err
	}
	return v1beta1.ResultSuccess, nil
}

func getDefaultMonitoringStackCR(namespace string) *msoapi.MonitoringStack {
	monitoringStackCR := &msoapi.MonitoringStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crNameForMonitoringStack,
			Namespace: namespace,
		},
	}
	return monitoringStackCR
}

func getDefaultServiceMonitor(namespace string) *rhobsv1.ServiceMonitor {
	return &rhobsv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crNameForServiceMonitor,
			Namespace: namespace,
			Labels:    setExporterLables(),
		},
		Spec: rhobsv1.ServiceMonitorSpec{
			Endpoints: []rhobsv1.Endpoint{
				{
					Interval: rhobsv1.Duration(ServiceMonitorPeriod),
					Path:     "/metrics",
					Port:     "metrics",
					Scheme:   "http",
				}},
			Selector: metav1.LabelSelector{
				MatchLabels: setExporterLables(),
			},
		},
	}
}

func setExporterLables() map[string]string {
	return map[string]string{"app": "dbaas-prometheus"}
}
func (r *reconciler) setPrometheusConfig(ctx context.Context, config v1beta1.ObservabilityConfig, namespace string) (*msoapi.PrometheusConfig, error) {

	prometheusConfig := &msoapi.PrometheusConfig{}
	prometheusConfig.Replicas = &replicas

	clusterID, _, err := util.GetClusterIDVersion(ctx, r.client)
	if err != nil {
		r.logger.Error(err, "error in getting clusterID")
		return prometheusConfig, err
	}
	if clusterID != "" {
		prometheusConfig.ExternalLabels = map[string]string{clusterIDLabel: clusterID}
	}

	remoteWriteSpec, _ := r.configureRemoteWrite(ctx, config, namespace)
	prometheusConfig.RemoteWrite = append(prometheusConfig.RemoteWrite, remoteWriteSpec)
	return prometheusConfig, nil

}

// configureRemoteWrite setting up environment params for RemoteWrite based on different Auth Type
func (r *reconciler) configureRemoteWrite(ctx context.Context, config v1beta1.ObservabilityConfig, namespace string) (rhobsv1.RemoteWriteSpec, error) {

	switch config.AuthType {
	case authTypeDex:
		return r.getDexRemoteWriteSpec(ctx, config, namespace)
	case authTypeRedhat:
		return r.getRHOBSRemoteWriteSpec(ctx, config, namespace)
	default:
		return rhobsv1.RemoteWriteSpec{}, fmt.Errorf("unknown auth type %v", config.AuthType)
	}
}

// getDexRemoteWriteSpec setting up internal dev environment params for remote write
func (r *reconciler) getDexRemoteWriteSpec(ctx context.Context, config v1beta1.ObservabilityConfig, namespace string) (rhobsv1.RemoteWriteSpec, error) {

	remoteWriteSpec := rhobsv1.RemoteWriteSpec{}
	if config.RemoteWritesURL != "" {
		rhobsRemoteWriteConfigSecret, err := r.validateSecret(ctx, config, namespace)
		if err != nil {
			return remoteWriteSpec, err
		}
		rhobsSecretData := rhobsRemoteWriteConfigSecret.Data
		rhobsToken, found := rhobsSecretData[rhobsTokenKey]
		if !found {
			return remoteWriteSpec, fmt.Errorf("rhobs secret does not contain a value for key %v", rhobsTokenKey)
		}
		remoteWriteSpec.URL = config.RemoteWritesURL
		remoteWriteSpec.BearerToken = string(rhobsToken)
		remoteWriteSpec.TLSConfig = tlsConfig()
		remoteWriteSpec.WriteRelabelConfigs = writeRelabelConfigs()
	}
	return remoteWriteSpec, nil
}

// getRHOBSRemoteWriteSpec setting up the params for RHOBS remote write
func (r *reconciler) getRHOBSRemoteWriteSpec(ctx context.Context, config v1beta1.ObservabilityConfig, namespace string) (rhobsv1.RemoteWriteSpec, error) {

	remoteWriteSpec := rhobsv1.RemoteWriteSpec{}

	if config.RemoteWritesURL != "" && config.RHSSOTokenURL != "" && config.RHOBSSecretName != "" {
		rhobsRemoteWriteConfigSecret, err := r.validateSecret(ctx, config, namespace)
		if err != nil {
			return remoteWriteSpec, err
		}
		rhobsSecretData := rhobsRemoteWriteConfigSecret.Data
		if _, found := rhobsSecretData[rhobsRemoteWriteConfigIDKey]; !found {
			return remoteWriteSpec, fmt.Errorf("rhobs secret does not contain a value for key %v", rhobsRemoteWriteConfigIDKey)
		}
		if _, found := rhobsSecretData[rhobsRemoteWriteConfigName]; !found {
			return remoteWriteSpec, fmt.Errorf("rhobs secret does not contain a value for key %v", rhobsRemoteWriteConfigName)
		}
		rhobsAudience, found := rhobsSecretData["rhobs-audience"]
		if !found {
			return remoteWriteSpec, fmt.Errorf("rhobs secret does not contain a value for key rhobs-audience")
		}
		remoteWriteSpec.URL = config.RemoteWritesURL
		remoteWriteSpec.OAuth2 = &rhobsv1.OAuth2{
			ClientID: rhobsv1.SecretOrConfigMap{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: config.RHOBSSecretName,
					},
					Key: rhobsRemoteWriteConfigIDKey,
				},
			},
			ClientSecret: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: config.RHOBSSecretName,
				},
				Key: rhobsRemoteWriteConfigName,
			},
			TokenURL:       config.RHSSOTokenURL,
			Scopes:         nil,
			EndpointParams: map[string]string{"audience": string(rhobsAudience)},
		}
		remoteWriteSpec.TLSConfig = tlsConfig()
		remoteWriteSpec.WriteRelabelConfigs = writeRelabelConfigs()
	}
	return remoteWriteSpec, nil
}

func (r *reconciler) validateSecret(ctx context.Context, config v1beta1.ObservabilityConfig, namespace string) (*corev1.Secret, error) {

	rhobsRemoteWriteConfigSecret := &corev1.Secret{}
	rhobsRemoteWriteConfigSecret.Name = config.RHOBSSecretName
	rhobsRemoteWriteConfigSecret.Namespace = namespace
	if err := r.client.Get(ctx, k8sclient.ObjectKeyFromObject(rhobsRemoteWriteConfigSecret), rhobsRemoteWriteConfigSecret); err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("rhobs remote write secret not found in namespace %v", namespace)
		}
		return nil, err
	}
	return rhobsRemoteWriteConfigSecret, nil
}

func tlsConfig() *rhobsv1.TLSConfig {
	return &rhobsv1.TLSConfig{
		SafeTLSConfig: rhobsv1.SafeTLSConfig{
			InsecureSkipVerify: true,
		}}
}

func writeRelabelConfigs() []rhobsv1.RelabelConfig {
	return []rhobsv1.RelabelConfig{{
		SourceLabels: []rhobsv1.LabelName{"__name__"},
		Regex:        "(" + strings.Join(metricsToInclude, "|") + ")",
		Action:       "keep",
	}}
}
