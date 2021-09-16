package controllers

import (
	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	successMetric = "success"
	failureMetric = "failure"
)

var (
	PlatformStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dbaas_platform_status.",
			Help: "Status of an installation of components and provider operators",
		},
		[]string{
			"platform",
			"status",
		},
	)
)

// SetPlatformStatus exposes dbaas_platform_status metric for each platform
func SetPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus) {
	//PlatformStatus.Reset()
	if len(platformName) > 0 {
		PlatformStatus.With(prometheus.Labels{"platform": string(platformName), "status": string(status)}).Set(float64(1))
	}
}

var DBaasRequestHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "dbaas_request_duration_seconds",
	Help: "Duration of upstream calls to provider operator/service endpoints",
}, []string{"provider_name", "instance_type", "instance_name", "action", "outcome"})

// NewExecution creates an Execution instance and starts the timer
func NewExecution(providerName string, instanceType string, instanceName string, action string) Execution {
	return Execution{
		begin:  time.Now(),
		labels: prometheus.Labels{"provider_name": providerName, "instance_type": instanceType, "instance_name": instanceName, "action": action},
	}
}

// Execution tracks state for an API execution for emitting metrics
type Execution struct {
	begin  time.Time
	labels prometheus.Labels
}

// Finish is used to log duration and success/failure
func (e *Execution) Finish(err error) {
	if err == nil {
		e.labels["outcome"] = successMetric
	} else {
		e.labels["outcome"] = failureMetric
	}
	duration := time.Since(e.begin)
	DBaasRequestHistogram.With(e.labels).Observe(duration.Seconds())
}
