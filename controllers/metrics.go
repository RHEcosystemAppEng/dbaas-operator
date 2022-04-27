package controllers

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

const (
	// Metrics names.
	metricNameDBaaSStackInstallationTotalDuration = "dbaas_stack_installation_total_duration_seconds"
	metricNameDBaaSPlatformInstallationStatus     = "dbaas_platform_installation_status"

	// Metrics labels.
	metricLabelName    = "name"
	metricLabelStatus  = "status"
	metricLabelVersion = "version"

	// installationTimeStart base time == 0
	installationTimeStart = 0
	// installationTimeWidth is the width of a bucket in the histogram, here it is 1m
	installationTimeWidth = 60

	// installationTimeBuckets is the number of buckets, here it 10 minutes worth of 1m buckets
	installationTimeBuckets = 10
)

var DBaasPlatformInstallationGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameDBaaSPlatformInstallationStatus,
	Help: "status of an installation of components and provider operators. values ( success=1, failed=0, in progress=2 ) ",
}, []string{metricLabelName, metricLabelStatus, metricLabelVersion})

var DBaasStackInstallationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaaSStackInstallationTotalDuration,
	Help: "How long in seconds installation of a DBaaS stack takes.",
	Buckets: prometheus.LinearBuckets(
		installationTimeStart,
		installationTimeWidth,
		installationTimeBuckets),
}, []string{metricLabelVersion})

// Execution tracks state for an API execution for emitting metrics
type Execution struct {
	begin time.Time
}

// NewExecution creates an Execution instance and starts the timer
func PlatformInstallStart() Execution {
	return Execution{
		begin: time.Now(),
	}
}

// PlatformStackInstallationMetric is used to log duration and success/failure
func (e *Execution) PlatformStackInstallationMetric(version string) {
	duration := time.Since(e.begin)
	DBaasStackInstallationHistogram.With(prometheus.Labels{metricLabelVersion: version}).Observe(duration.Seconds())
}

// SetPlatformStatus exposes dbaas_platform_status metric for each platform
func SetPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 {
		switch status {

		case dbaasv1alpha1.ResultFailed:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(status), metricLabelVersion: version}).Set(float64(0))
		case dbaasv1alpha1.ResultSuccess:
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultInProgress), metricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultFailed), metricLabelVersion: version})
			DBaasPlatformInstallationGauge.With(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(status), metricLabelVersion: version}).Set(float64(1))
		case dbaasv1alpha1.ResultInProgress:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(status), metricLabelVersion: version}).Set(float64(2))
		}

	}
}

// CleanPlatformStatusMetric delete the dbaas_platform_status metric for each platform
func CleanPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 && status == dbaasv1alpha1.ResultSuccess {
		DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultSuccess), metricLabelVersion: version})
	}
}
