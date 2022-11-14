package providersinstallation

import (
	"context"
	"fmt"
	"strings"

	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/util"

	"k8s.io/apimachinery/pkg/api/errors"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	msoapi "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	clusterIDLabel                     = "cluster_id"
	clusterVersionLabel                = "cluster_version"
	crName                             = "dbaas-operator-mso"
	rhobsRemoteWriteConfigIDKey        = "prom-remote-write-config-id"
	rhobsRemoteWriteConfigName         = "prom-remote-write-config-secret" //#nosec
	rhobsTokenKey                      = "rhobs-token"                     //#nosec
	authTypeDex                 string = "dex"
	authTypeRedhat              string = "redhat-sso"
)

var metricsToInclude = []string{"dbaas_.*$", "csv_succeeded$", "csv_abnormal$", "ALERTS$", "subscription_sync_total"}
var replicas int32 = 1

func (r *reconciler) createObservabilityCR(ctx context.Context, cr *dbaasv1beta1.DBaaSPlatform) (dbaasv1beta1.PlatformInstlnStatus, error) {
	config := reconcilers.GetObservabilityConfig()
	monitoringStackCR := getDefaultMonitoringStackCR(cr.Namespace)

	monitoringStackList := &msoapi.MonitoringStackList{}
	listOpts := []k8sclient.ListOption{
		k8sclient.InNamespace(monitoringStackCR.Namespace),
	}
	err := r.client.List(ctx, monitoringStackList, listOpts...)
	if err != nil {
		return dbaasv1beta1.ResultFailed, fmt.Errorf("could not get a list of monitoring stack CR: %w", err)
	}

	if len(monitoringStackList.Items) == 0 {
		if config.RemoteWritesURL != "" && config.AuthType != "" && config.AddonName != "" {
			prometheusConfig, _ := r.setPrometheusConfig(ctx, config, monitoringStackCR.Namespace)
			monitoringStackCR.Spec.PrometheusConfig = prometheusConfig
		}
		err = controllerutil.SetControllerReference(cr, monitoringStackCR, r.scheme)
		if err != nil {
			return dbaasv1beta1.ResultFailed, err
		}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.client, monitoringStackCR, func() error {
			monitoringStackCR.Labels = map[string]string{
				"managed-by": "dbaas-operator",
			}

			return nil
		}); err != nil {
			return dbaasv1beta1.ResultFailed, err
		}
	} else if len(monitoringStackList.Items) == 1 {
		monitoringStackCR = &monitoringStackList.Items[0]
		if config.RemoteWritesURL != "" && config.AuthType != "" && config.AddonName != "" {
			prometheusConfig, _ := r.setPrometheusConfig(ctx, config, monitoringStackCR.Namespace)
			monitoringStackCR.Spec.PrometheusConfig = prometheusConfig
			if _, err := controllerutil.CreateOrUpdate(ctx, r.client, monitoringStackCR, func() error {
				monitoringStackCR.Labels = map[string]string{
					"managed-by": "dbaas-operator",
				}
				return nil
			}); err != nil {
				return dbaasv1beta1.ResultFailed, err
			}
		}

	} else {
		return dbaasv1beta1.ResultFailed, fmt.Errorf("too many monitoringStackCR resources found. Expecting 1, found %d MonitoringStack resources in %s namespace", len(monitoringStackList.Items), cr.Namespace)
	}
	return dbaasv1beta1.ResultSuccess, nil

}

func getDefaultMonitoringStackCR(namespace string) *msoapi.MonitoringStack {
	monitoringStackCR := &msoapi.MonitoringStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crName,
			Namespace: namespace,
		},
		Spec: msoapi.MonitoringStackSpec{
			LogLevel: "debug",
			ResourceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "dbaas-prometheus",
				},
			},
		},
	}
	return monitoringStackCR
}

func (r *reconciler) setPrometheusConfig(ctx context.Context, config dbaasv1beta1.ObservabilityConfig, namespace string) (*msoapi.PrometheusConfig, error) {

	prometheusConfig := &msoapi.PrometheusConfig{}
	prometheusConfig.Replicas = &replicas

	clusterID, clusterVersion, err := util.GetClusterIDVersion(ctx, r.client)
	if err != nil {
		return prometheusConfig, err
	}
	if clusterID != "" && clusterVersion != "" {
		prometheusConfig.ExternalLabels = map[string]string{clusterIDLabel: clusterID, clusterVersionLabel: clusterVersion}
	}

	remoteWriteSpec, _ := r.configureRemoteWrite(ctx, config, namespace)
	prometheusConfig.RemoteWrite = append(prometheusConfig.RemoteWrite, remoteWriteSpec)
	return prometheusConfig, nil

}

// configureRemoteWrite setting up environment params for RemoteWrite based on different Auth Type
func (r *reconciler) configureRemoteWrite(ctx context.Context, config dbaasv1beta1.ObservabilityConfig, namespace string) (monv1.RemoteWriteSpec, error) {

	switch config.AuthType {
	case authTypeDex:
		return r.getDexRemoteWriteSpec(ctx, config, namespace)
	case authTypeRedhat:
		return r.getRHOBSRemoteWriteSpec(ctx, config, namespace)
	default:
		return monv1.RemoteWriteSpec{}, fmt.Errorf("unknown auth type %v", config.AuthType)
	}
}

// getDexRemoteWriteSpec setting up internal dev environment params for remote write
func (r *reconciler) getDexRemoteWriteSpec(ctx context.Context, config dbaasv1beta1.ObservabilityConfig, namespace string) (monv1.RemoteWriteSpec, error) {

	remoteWriteSpec := monv1.RemoteWriteSpec{}
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
func (r *reconciler) getRHOBSRemoteWriteSpec(ctx context.Context, config dbaasv1beta1.ObservabilityConfig, namespace string) (monv1.RemoteWriteSpec, error) {

	remoteWriteSpec := monv1.RemoteWriteSpec{}

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
		remoteWriteSpec.OAuth2 = &monv1.OAuth2{
			ClientID: monv1.SecretOrConfigMap{
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

func (r *reconciler) validateSecret(ctx context.Context, config dbaasv1beta1.ObservabilityConfig, namespace string) (*corev1.Secret, error) {

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

func tlsConfig() *monv1.TLSConfig {
	return &monv1.TLSConfig{
		SafeTLSConfig: monv1.SafeTLSConfig{
			InsecureSkipVerify: true,
		}}
}

func writeRelabelConfigs() []monv1.RelabelConfig {
	return []monv1.RelabelConfig{{
		SourceLabels: []monv1.LabelName{"__name__"},
		Regex:        "(" + strings.Join(metricsToInclude, "|") + ")",
		Action:       "keep",
	}}
}
