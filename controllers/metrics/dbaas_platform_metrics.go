package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

const (
	// Resource label values
	LabelResourceValuePlatform = "dbaas_platform"

	// Metrics names.
	MetricNameDBaaSPlatformInstallationStatus = "dbaas_platform_installation_status"

	// Metrics labels.
	MetricLabelPlatformName = "cloud_platform_name"

	// Event label values

	// Error Code label values
	LabelErrorCdValueErrorFetchingDBaaSPlatformResources = "error_fetching_dbaas_platform_resource"
	LabelErrorCdValueErrorGettingOpenShiftURL            = "error_getting_openshift_url"
	LabelErrorCdValueErrorDeletingPlatform               = "error_deleting_dbaas_platform"
)

// DBaasPlatformInstallationGauge defines a gauge for DBaaSPlatformInstallationStatus
var DBaasPlatformInstallationGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameDBaaSPlatformInstallationStatus,
	Help: "The status of an installation of components and provider operators. values ( success=1, failed=0, in progress=2 ) ",
}, []string{MetricLabelName, MetricLabelStatus, MetricLabelVersion})

// PlatformStackInstallationMetric is used to log duration and success/failure
func PlatformStackInstallationMetric(platform *dbaasv1beta1.DBaaSPlatform, version string, e Execution) {
	duration := time.Since(e.begin)
	for _, cond := range platform.Status.Conditions {
		if cond.Type == dbaasv1beta1.DBaaSPlatformReadyType {
			lastTransitionTime := cond.LastTransitionTime
			duration = lastTransitionTime.Sub(platform.CreationTimestamp.Time)
			DBaasStackInstallationHistogram.With(prometheus.Labels{MetricLabelVersion: version, MetricLabelCreationTimestamp: platform.CreationTimestamp.String()}).Observe(duration.Seconds())
		} else {
			DBaasStackInstallationHistogram.With(prometheus.Labels{MetricLabelVersion: version, MetricLabelCreationTimestamp: platform.CreationTimestamp.String()}).Observe(duration.Seconds())
		}
	}
}

// SetPlatformStatusMetric exposes dbaas_platform_status Metric for each platform
func SetPlatformStatusMetric(platformName dbaasv1beta1.PlatformName, status dbaasv1beta1.PlatformInstlnStatus, version string) {
	if len(platformName) > 0 {
		switch status {

		case dbaasv1beta1.ResultFailed:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(status), MetricLabelVersion: version}).Set(float64(0))
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultSuccess), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultInProgress), MetricLabelVersion: version})
		case dbaasv1beta1.ResultSuccess:
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultInProgress), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultFailed), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.With(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(status), MetricLabelVersion: version}).Set(float64(1))
		case dbaasv1beta1.ResultInProgress:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(status), MetricLabelVersion: version}).Set(float64(2))
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultSuccess), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultFailed), MetricLabelVersion: version})
		}
	}
}

// setPlatformRequestDurationSeconds set the metrics for platform request duration in seconds
func setPlatformRequestDurationSeconds(platform dbaasv1beta1.DBaaSPlatform, account string, execution Execution, event string) {
	switch event {
	case LabelEventValueCreate:
		duration := time.Now().UTC().Sub(platform.CreationTimestamp.Time.UTC())
		UpdateRequestsDurationHistogram(platform.Name, account, platform.Namespace, LabelResourceValuePlatform, event, duration.Seconds())
	case LabelEventValueDelete:
		deletionTimestamp := execution.begin.UTC()
		if platform.DeletionTimestamp != nil {
			deletionTimestamp = platform.DeletionTimestamp.UTC()
		}

		duration := time.Now().UTC().Sub(deletionTimestamp.UTC())
		UpdateRequestsDurationHistogram(platform.Name, account, platform.Namespace, LabelResourceValuePlatform, event, duration.Seconds())
	}
}

// SetPlatformMetrics set the metrics for a platform
func SetPlatformMetrics(platform dbaasv1beta1.DBaaSPlatform, account string, execution Execution, event string, errCd string) {
	setPlatformRequestDurationSeconds(platform, account, execution, event)
	UpdateErrorsTotal(platform.Name, account, platform.Namespace, LabelResourceValuePlatform, event, errCd)
}
