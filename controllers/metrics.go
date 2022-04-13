package controllers

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

var (
	PlatformStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dbaas_platform_status",
			Help: "status of an installation of components and provider operators",
		},
		[]string{
			"platform",
			"status",
		},
	)
)

// SetPlatformStatus exposes dbaas_platform_status metric for each platform
func SetPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus) {
	if len(platformName) > 0 {
		switch status {

		case dbaasv1alpha1.ResultFailed:
			PlatformStatus.With(prometheus.Labels{"platform": string(platformName), "status": string(status)}).Set(float64(0))
		case dbaasv1alpha1.ResultSuccess:
			PlatformStatus.Delete(prometheus.Labels{"platform": string(platformName), "status": string(dbaasv1alpha1.ResultInProgress)})
			PlatformStatus.Delete(prometheus.Labels{"platform": string(platformName), "status": string(dbaasv1alpha1.ResultFailed)})
			PlatformStatus.With(prometheus.Labels{"platform": string(platformName), "status": string(status)}).Set(float64(1))
		case dbaasv1alpha1.ResultInProgress:
			PlatformStatus.With(prometheus.Labels{"platform": string(platformName), "status": string(status)}).Set(float64(2))
		}

	}
}

// CleanPlatformStatusMetric delete the dbaas_platform_status metric for each platform
func CleanPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus) {
	if len(platformName) > 0 && status == dbaasv1alpha1.ResultSuccess {
		PlatformStatus.Delete(prometheus.Labels{"platform": string(platformName), "status": string(dbaasv1alpha1.ResultSuccess)})
	}
}

var DBaasRequestHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "dbaas_request_duration_seconds",
	Help: "Duration of upstream calls to provider operator/service endpoints",
}, []string{"provider_name", "instance_type", "instance_name", "result", "message"})

var DBaasInstallationtHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "dbaas_installation_duration_seconds",
	Help: "Duration of DBaaS platform stack installation",
}, []string{"message"})

// NewExecution creates an Execution instance and starts the timer
func NewExecution(providerName string, instanceType string, instanceName string) Execution {
	return Execution{
		begin:  time.Now(),
		labels: prometheus.Labels{"provider_name": providerName, "instance_type": instanceType, "instance_name": instanceName},
	}
}

// NewExecution creates an Execution instance and starts the timer
func PlatformInstallStart() Execution {
	return Execution{
		begin:  time.Now(),
		labels: prometheus.Labels{"message": "DBaaS platform stack installation started"},
	}
}

// Execution tracks state for an API execution for emitting metrics
type Execution struct {
	begin  time.Time
	labels prometheus.Labels
}

// Finish is used to log duration and success/failure
func (e *Execution) Finish(conditions []metav1.Condition) {

	for _, cond := range conditions {
		e.labels["result"] = cond.Reason
		e.labels["message"] = cond.Message
	}
	if len(conditions) == 0 {
		e.labels["result"] = "result not found"
		e.labels["message"] = "message not found"
	}

	duration := time.Since(e.begin)
	DBaasRequestHistogram.With(e.labels).Observe(duration.Seconds())
}

// PlatformInstallationFinish is used to log duration and success/failure
func (e *Execution) PlatformInstallationFinish() {
	e.labels["message"] = "DBaaS platform stack installation completed"
	duration := time.Since(e.begin)
	DBaasInstallationtHistogram.With(e.labels).Observe(duration.Seconds())
}
