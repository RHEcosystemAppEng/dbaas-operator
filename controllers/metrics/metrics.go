package metrics

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/prometheus/client_golang/prometheus"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

const (
	// Metrics names.
	MetricNameDBaaSStackInstallationTotalDuration = "dbaas_stack_installation_total_duration_seconds"
	MetricNameDBaaSPlatformInstallationStatus     = "dbaas_platform_installation_status"
	MetricNameInstanceStatusReady                 = "dbaas_instance_status_ready"
	MetricNameDBaasInstanceDuration               = "dbaas_instance_request_duration_seconds"
	MetricNameInstancePhase                       = "dbaas_instance_phase"
	MetricNameOperatorVersion                     = "dbaas_version_info"
	MetricNameDBaaSRequestsDurationSeconds        = "dbaas_requests_duration_seconds"
	MetricNameDBaaSRequestsErrorCount             = "dbaas_requests_error_count"

	// Metrics labels.
	MetricLabelName              = "name"
	MetricLabelStatus            = "status"
	MetricLabelVersion           = "version"
	MetricLabelProvider          = "provider"
	MetricLabelAccountName       = "account"
	MetricLabelNameSpace         = "namespace"
	MetricLabelInstanceID        = "instance_id"
	MetricLabelReason            = "reason"
	MetricLabelInstanceName      = "instance_name"
	MetricLabelCreationTimestamp = "creation_timestamp"
	MetricLabelConsoleULR        = "openshift_url"
	MetricLabelPlatformName      = "cloud_platform_name"
	MetricLabelResource          = "resource"
	MetricLabelEvent             = "event"
	MetricLabelErrorCd           = "error_cd"

	// Event label values
	LabelEventValueCreate = "create"
	LabelEventValueDelete = "delete"

	// Error Code label values
	LabelErrorCdValueResourceNotFound     = "resource_not_found"
	LabelErrorCdValueUnableToListPolicies = "unable_to_list_policies"

	// installationTimeStart base time == 0
	installationTimeStart = 0
	// installationTimeWidth is the width of a bucket in the histogram, here it is 1m
	installationTimeWidth = 60
	// installationTimeBuckets is the number of buckets, here it 10 minutes worth of 1m buckets
	installationTimeBuckets = 10
)

// DBaasPlatformInstallationGauge defines a gauge for DBaaSPlatformInstallationStatus
var DBaasPlatformInstallationGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameDBaaSPlatformInstallationStatus,
	Help: "The status of an installation of components and provider operators. values ( success=1, failed=0, in progress=2 ) ",
}, []string{MetricLabelName, MetricLabelStatus, MetricLabelVersion})

// DBaaSInventoryStatusGauge defines a gauge for DBaaSInventoryStatus
var DBaaSInventoryStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameInventoryStatusReady,
	Help: "The status of DBaaS Provider Account, values ( ready=1, error / not ready=0 )",
}, []string{MetricLabelProvider, MetricLabelName, MetricLabelNameSpace, MetricLabelStatus, MetricLabelReason})

// DBaaSInstanceStatusGauge defines a gauge for DBaaSInstanceStatus
var DBaaSInstanceStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameInstanceStatusReady,
	Help: "The status of DBaaS instance, values ( ready=1, error / not ready=0 )",
}, []string{MetricLabelProvider, MetricLabelAccountName, MetricLabelInstanceName, MetricLabelNameSpace, MetricLabelStatus, MetricLabelReason, MetricLabelCreationTimestamp})

// DBaaSInstancePhaseGauge defines a gauge for DBaaSInstancePhase
var DBaaSInstancePhaseGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameInstancePhase,
	Help: "Current status phase of the Instance currently managed by RHODA values ( Pending=-1, Creating=0, Ready=1, Unknown=2, Failed=3, Error=4, Deleting=5 ).",
}, []string{MetricLabelProvider, MetricLabelAccountName, MetricLabelInstanceName, MetricLabelNameSpace, MetricLabelCreationTimestamp})

// DBaasStackInstallationHistogram defines a histogram for DBaasStackInstallation
var DBaasStackInstallationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: MetricNameDBaaSStackInstallationTotalDuration,
	Help: "Time in seconds installation of a DBaaS stack takes.",
	Buckets: prometheus.LinearBuckets(
		installationTimeStart,
		installationTimeWidth,
		installationTimeBuckets),
}, []string{MetricLabelVersion, MetricLabelCreationTimestamp})

// DBaasInventoryRequestDurationSeconds defines a histogram for DBaasInventoryRequestDuration in seconds
var DBaasInventoryRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: MetricNameDBaasInventoryDuration,
	Help: "Request/Response duration of provider account of upstream calls to provider operator/service endpoints",
	Buckets: prometheus.LinearBuckets(installationTimeStart,
		installationTimeWidth,
		installationTimeBuckets),
}, []string{MetricLabelProvider, MetricLabelName, MetricLabelNameSpace, MetricLabelCreationTimestamp})

// DBaasOperatorVersionInfo defines a gauge for DBaaS Operator version
var DBaasOperatorVersionInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameOperatorVersion,
	Help: "The current version of DBaaS Operator installed in the cluster",
}, []string{MetricLabelVersion, MetricLabelConsoleULR, MetricLabelPlatformName})

// DBaasInstanceRequestDurationSeconds defines a histogram for DBaasInstanceRequestDuration
var DBaasInstanceRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: MetricNameDBaasInstanceDuration,
	Help: "Request/Response duration of instance of upstream calls to provider operator/service endpoints",
}, []string{MetricLabelProvider, MetricLabelAccountName, MetricLabelInstanceName, MetricLabelNameSpace, MetricLabelCreationTimestamp})

// DBaaSRequestsDurationHistogram DBaaS Requests Duration Histogram for all DBaaS Resources
var DBaaSRequestsDurationHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: MetricNameDBaaSRequestsDurationSeconds,
		Help: "Request durations histogram for given resource(e.g. inventory) and for a given event(e.g. create or delete)",
	},
	[]string{MetricLabelProvider, MetricLabelAccountName, MetricLabelNameSpace, MetricLabelResource, MetricLabelEvent})

// DBaaSRequestsErrorsCounter Total errors encountered counter
var DBaaSRequestsErrorsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: MetricNameDBaaSRequestsErrorCount,
		Help: "Total requests for a given resource(e.g. DBaaS Inventory), for a given event(e.g. create or delete), with a given error code (e.g. resource exists, resource not there)",
	},
	[]string{MetricLabelProvider, MetricLabelAccountName, MetricLabelNameSpace, MetricLabelResource, MetricLabelEvent, MetricLabelErrorCd})

// Execution tracks state for an API execution for emitting Metrics
type Execution struct {
	begin time.Time
}

// PlatformInstallStart creates an Execution instance and starts the timer
func PlatformInstallStart() Execution {
	return Execution{
		begin: time.Now().UTC(),
	}
}

// PlatformStackInstallationMetric is used to log duration and success/failure
func (e *Execution) PlatformStackInstallationMetric(platform *dbaasv1alpha1.DBaaSPlatform, version string) {
	duration := time.Since(e.begin)
	for _, cond := range platform.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSPlatformReadyType {
			lastTransitionTime := cond.LastTransitionTime
			duration = lastTransitionTime.Sub(platform.CreationTimestamp.Time)
			DBaasStackInstallationHistogram.With(prometheus.Labels{MetricLabelVersion: version, MetricLabelCreationTimestamp: platform.CreationTimestamp.String()}).Observe(duration.Seconds())
		} else {
			DBaasStackInstallationHistogram.With(prometheus.Labels{MetricLabelVersion: version, MetricLabelCreationTimestamp: platform.CreationTimestamp.String()}).Observe(duration.Seconds())
		}
	}
}

// SetPlatformStatusMetric exposes dbaas_platform_status Metric for each platform
func SetPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 {
		switch status {

		case dbaasv1alpha1.ResultFailed:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(status), MetricLabelVersion: version}).Set(float64(0))
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1alpha1.ResultSuccess), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1alpha1.ResultInProgress), MetricLabelVersion: version})
		case dbaasv1alpha1.ResultSuccess:
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1alpha1.ResultInProgress), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1alpha1.ResultFailed), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.With(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(status), MetricLabelVersion: version}).Set(float64(1))
		case dbaasv1alpha1.ResultInProgress:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(status), MetricLabelVersion: version}).Set(float64(2))
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1alpha1.ResultSuccess), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1alpha1.ResultFailed), MetricLabelVersion: version})

		}
	}
}

// CleanPlatformStatusMetric delete the dbaas_platform_status Metric for each platform
func CleanPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 && status == dbaasv1alpha1.ResultSuccess {
		DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1alpha1.ResultSuccess), MetricLabelVersion: version})
	}
}

// SetOpenShiftInstallationInfoMetric set the Metrics for openshift info
func SetOpenShiftInstallationInfoMetric(operatorVersion string, consoleURL string, platformType string) {
	DBaasOperatorVersionInfo.With(prometheus.Labels{MetricLabelVersion: operatorVersion, MetricLabelConsoleULR: consoleURL, MetricLabelPlatformName: platformType}).Set(1)
}

// SetInstanceMetrics set the Metrics for an instance
func SetInstanceMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance, execution Execution) {
	setInstanceStatusMetrics(provider, account, instance)
	setInstancePhaseMetrics(provider, account, instance)
	setInstanceRequestDurationSeconds(provider, account, instance, execution)

}

// setInstanceStatusMetrics set the Metrics based on instance status
func setInstanceStatusMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance) {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInstanceReadyType {
			DBaaSInstanceStatusGauge.DeletePartialMatch(prometheus.Labels{MetricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.Namespace})
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{MetricLabelProvider: provider, MetricLabelAccountName: account, MetricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.Namespace, MetricLabelStatus: string(cond.Status), MetricLabelReason: cond.Reason, MetricLabelCreationTimestamp: instance.CreationTimestamp.String()}).Set(1)
			} else {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{MetricLabelProvider: provider, MetricLabelAccountName: account, MetricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.Namespace, MetricLabelStatus: string(cond.Status), MetricLabelReason: cond.Reason, MetricLabelCreationTimestamp: instance.CreationTimestamp.String()}).Set(0)
			}
			break
		}
	}
}

// setInstanceRequestDurationSeconds set the Metrics for instance request duration in seconds
func setInstanceRequestDurationSeconds(provider string, account string, instance dbaasv1alpha1.DBaaSInstance, execution Execution) {
	httpDuration := time.Since(execution.begin)
	for _, cond := range instance.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInstanceProviderSyncType {
			if cond.Status == metav1.ConditionTrue {
				lastTransitionTime := cond.LastTransitionTime
				httpDuration = lastTransitionTime.Sub(instance.CreationTimestamp.Time)
				DBaasInstanceRequestDurationSeconds.With(prometheus.Labels{MetricLabelProvider: provider, MetricLabelAccountName: account, MetricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.GetNamespace(), MetricLabelCreationTimestamp: instance.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			} else {
				DBaasInstanceRequestDurationSeconds.With(prometheus.Labels{MetricLabelProvider: provider, MetricLabelAccountName: account, MetricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.GetNamespace(), MetricLabelCreationTimestamp: instance.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			}
			break
		}
	}
}

// setInstancePhaseMetrics set the Metrics for instance phase
func setInstancePhaseMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance) {
	var phase float64

	switch instance.Status.Phase {
	case dbaasv1alpha1.InstancePhasePending:
		phase = -1
	case dbaasv1alpha1.InstancePhaseCreating:
		phase = 0
	case dbaasv1alpha1.InstancePhaseReady:
		phase = 1
	case dbaasv1alpha1.InstancePhaseUnknown:
		phase = 2
	case dbaasv1alpha1.InstancePhaseFailed:
		phase = 3
	case dbaasv1alpha1.InstancePhaseError:
		phase = 4
	case dbaasv1alpha1.InstancePhaseDeleting:
		phase = 5
	case dbaasv1alpha1.InstancePhaseDeleted:
		phase = 6
	}

	DBaaSInstancePhaseGauge.With(prometheus.Labels{
		MetricLabelProvider:          provider,
		MetricLabelAccountName:       account,
		MetricLabelInstanceName:      instance.Name,
		MetricLabelNameSpace:         instance.Namespace,
		MetricLabelCreationTimestamp: instance.CreationTimestamp.String(),
	}).Set(phase)
}

// CleanInstanceMetrics delete instance Metrics based on the condition type
func CleanInstanceMetrics(instance *dbaasv1alpha1.DBaaSInstance) {
	for _, cond := range instance.Status.Conditions {
		switch cond.Type {
		case dbaasv1alpha1.DBaaSInstanceReadyType:
			DBaaSInstanceStatusGauge.DeletePartialMatch(prometheus.Labels{MetricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.Namespace})
			DBaaSInstancePhaseGauge.DeletePartialMatch(prometheus.Labels{MetricLabelInstanceName: instance.Name, MetricLabelNameSpace: instance.Namespace})
		case dbaasv1alpha1.DBaaSInstanceProviderSyncType:
			DBaasInstanceRequestDurationSeconds.DeletePartialMatch(prometheus.Labels{MetricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.GetNamespace()})
		}
	}
}

// UpdateRequestsDurationHistogram Utility function to update request duration histogram
func UpdateRequestsDurationHistogram(provider string, account string, namespace string, resource string, event string, duration float64) {
	DBaaSRequestsDurationHistogram.WithLabelValues(provider, account, namespace, resource, event).Observe(duration)
}

// UpdateErrorsTotal Utility function to update errors total
func UpdateErrorsTotal(provider string, account string, namespace string, resource string, event string, errCd string) {
	if len(errCd) > 0 {
		DBaaSRequestsErrorsCounter.WithLabelValues(provider, account, namespace, resource, event, errCd).Add(1)
	}
}
