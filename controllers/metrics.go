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
	metricNameDBaasInventoryCount                 = "dbaas_inventory_created"
	metricNameDBaasConnectionCount                = "dbaas_connection_count"
	metricNameDBaasTenantCount                    = "dbaas_registerd_tenant_count"
	metricNameDBaasInstanceCount                  = "dbaas_instance_count"

	// Metrics labels.
	metricLabelName     = "name"
	metricLabelStatus   = "status"
	metricLabelVersion  = "version"
	metricLabelProvider = "provider"
	metricLabelMessage  = "message"
	metricAccountName   = "account_name"

	// installationTimeStart base time == 0
	installationTimeStart = 0
	// installationTimeWidth is the width of a bucket in the histogram, here it is 1m
	installationTimeWidth = 60

	// installationTimeBuckets is the number of buckets, here it 10 minutes worth of 1m buckets
	installationTimeBuckets = 10
)

var DBaasPlatformInstallationtGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameDBaaSPlatformInstallationStatus,
	Help: "status of an installation of components and provider operators. values (success=1,failed=0, in progress=2) ",
}, []string{metricLabelName, metricLabelStatus, metricLabelVersion})

var DBaasStackInstallationtHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaaSStackInstallationTotalDuration,
	Help: "How long in seconds installation of a DBaaS stack takes.",
	Buckets: prometheus.LinearBuckets(
		installationTimeStart,
		installationTimeWidth,
		installationTimeBuckets),
}, []string{metricLabelVersion})

// var DBaaSInventoryCount = prometheus.NewGauge(prometheus.GaugeOpts{
// 	Name: metricNameDBaasInventoryCount,
// 	Help: "the total count of dbaas inventory",
// })

var DBaaSInventoryCountGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameDBaasInventoryCount,
	Help: "The number of provider created processed with status",
}, []string{"provider", "account_name", "status"})

var DBaasConnectionCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameDBaasConnectionCount,
	Help: "the total count of dbaas connections",
}, []string{"provider_name", "message"})

var DBaasTenantCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: metricNameDBaasTenantCount,
	Help: "Number of tenant successfully created",
}, []string{"Namespace", "message"})

var DBaasInstanceCount = prometheus.NewCounter(prometheus.CounterOpts{
	Name: metricNameDBaasInstanceCount,
	Help: "The count of dbaas instances",
})

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
	DBaasStackInstallationtHistogram.With(prometheus.Labels{metricLabelVersion: version}).Observe(duration.Seconds())
}

// SetPlatformStatus exposes dbaas_platform_status metric for each platform
func SetPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 {
		switch status {

		case dbaasv1alpha1.ResultFailed:
			DBaasPlatformInstallationtGauge.With(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(status), metricLabelVersion: version}).Set(float64(0))
		case dbaasv1alpha1.ResultSuccess:
			DBaasPlatformInstallationtGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultInProgress), metricLabelVersion: version})
			DBaasPlatformInstallationtGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultFailed), metricLabelVersion: version})
			DBaasPlatformInstallationtGauge.With(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(status), metricLabelVersion: version}).Set(float64(1))
		case dbaasv1alpha1.ResultInProgress:
			DBaasPlatformInstallationtGauge.With(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(status), metricLabelVersion: version}).Set(float64(2))
		}

	}
}

// CleanPlatformStatusMetric delete the dbaas_platform_status metric for each platform
func CleanPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 && status == dbaasv1alpha1.ResultSuccess {
		DBaasPlatformInstallationtGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultSuccess), metricLabelVersion: version})
	}
}

// func IncrementInventoryCount() {
// 	DBaaSInventoryCount.Inc()
// }

// func DecrementInventoryCount() {
// 	DBaaSInventoryCount.Dec()
// }

func SetInventoryCreation(provider string, inventory dbaasv1alpha1.DBaaSInventory) {
	for _, cond := range inventory.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInventoryReadyType {
			DBaaSInventoryCountGauge.WithLabelValues(provider, inventory.GetName(), cond.Reason).Inc()
		}
	}
}

func SetDbaasConnectionMetric(provider string, message string) {
	DBaasConnectionCount.WithLabelValues(provider, message).Inc()
}

func SetDbaasTenantMetric(namespace string, message string) {
	DBaasTenantCount.WithLabelValues(namespace, message).Inc()
}

func SetDbaasInstanceMetric(provider string, message string) {
	DBaasInstanceCount.Inc()
}

// make run ENABLE_WEBHOOKS=false
// http://localhost:8080/metrics

// 	defer func() {
// 		SetInventoryCreation(inventory.Spec.ProviderRef.Name, inventory)
// 	}()
// func SetInventoryCreation(provider string, inventory dbaasv1alpha1.DBaaSInventory) {

// 	for _, cond := range inventory.Status.Conditions {
// 		if cond.Type == dbaasv1alpha1.DBaaSInventoryReadyType {
// 			DBaaSInventoryCountGauge.WithLabelValues(provider, inventory.GetName(), cond.Reason).Inc()
// 		}
// 	}
// }
// var DBaaSInventoryCountGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
// 	Name: "dbaas_inventory_created",
// 	Help: "The number of provider created processed with status",
// }, []string{"provider", "account_name", "status"})
// RHODA Metrics Sync

// make install run INSTALL_NAMESPACE=<your_target_namespace> ENABLE_WEBHOOKS=false
// Continue below by following
