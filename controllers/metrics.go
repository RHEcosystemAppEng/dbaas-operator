package controllers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/prometheus/client_golang/prometheus"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

const (
	// Metrics names.
	metricNameDBaaSStackInstallationTotalDuration = "dbaas_stack_installation_total_duration_seconds"
	metricNameDBaaSPlatformInstallationStatus     = "dbaas_platform_installation_status"
	metricNameInventoryStatusReady                = "dbaas_inventory_status_ready"
	metricNameDBaasInventoryDuration              = "dbaas_inventory_request_duration_seconds"
	metricNameConnectionStatusReady               = "dbaas_connection_status_ready"
	metricNameDBaasConnectionDuration             = "dbaas_connection_request_duration_seconds"
	metricNameInstanceStatusReady                 = "dbaas_instance_status_ready"
	metricNameDBaasInstanceDuration               = "dbaas_instance_request_duration_seconds"
	metricNameInstancePhase                       = "dbaas_instance_phase"
	metricNameOperatorVersion                     = "dbaas_version_info"

	// Metrics labels.
	metricLabelName              = "name"
	metricLabelStatus            = "status"
	metricLabelVersion           = "version"
	metricLabelProvider          = "provider"
	metricLabelAccountName       = "account"
	metricLabelConnectionName    = "name"
	metricLabelNameSpace         = "namespace"
	metricLabelInstanceID        = "instance_id"
	metricLabelReason            = "reason"
	metricLabelInstanceName      = "name"
	metricLabelCreationTimestamp = "creation_timestamp"
	metricLabelConsoleULR        = "openshift_url"
	metricLabelPlatformName      = "cloud_platform_name"

	// installationTimeStart base time == 0
	installationTimeStart = 0
	// installationTimeWidth is the width of a bucket in the histogram, here it is 1m
	installationTimeWidth = 60
	// installationTimeBuckets is the number of buckets, here it 10 minutes worth of 1m buckets
	installationTimeBuckets = 10
)

// DBaasPlatformInstallationGauge defines a gauge for DBaaSPlatformInstallationStatus
var DBaasPlatformInstallationGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameDBaaSPlatformInstallationStatus,
	Help: "The status of an installation of components and provider operators. values ( success=1, failed=0, in progress=2 ) ",
}, []string{metricLabelName, metricLabelStatus, metricLabelVersion})

// DBaaSInventoryStatusGauge defines a gauge for DBaaSInventoryStatus
var DBaaSInventoryStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInventoryStatusReady,
	Help: "The status of DBaaS Provider Account, values ( ready=1, error / not ready=0 )",
}, []string{metricLabelProvider, metricLabelName, metricLabelNameSpace, metricLabelStatus, metricLabelReason, metricLabelCreationTimestamp})

// DBaaSConnectionStatusGauge defines a gauge for DBaaSConnectionStatus
var DBaaSConnectionStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameConnectionStatusReady,
	Help: "The status of DBaaS connections, values ( ready=1, error / not ready=0 )",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceID, metricLabelConnectionName, metricLabelNameSpace, metricLabelStatus, metricLabelReason, metricLabelCreationTimestamp})

// DBaaSInstanceStatusGauge defines a gauge for DBaaSInstanceStatus
var DBaaSInstanceStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInstanceStatusReady,
	Help: "The status of DBaaS instance, values ( ready=1, error / not ready=0 )",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceName, metricLabelNameSpace, metricLabelStatus, metricLabelReason, metricLabelCreationTimestamp})

// DBaaSInstancePhaseGauge defines a gauge for DBaaSInstancePhase
var DBaaSInstancePhaseGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInstancePhase,
	Help: "Current status phase of the Instance currently managed by RHODA values ( Pending=-1, Creating=0, Ready=1, Unknown=2, Failed=3, Error=4, Deleting=5 ).",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceName, metricLabelNameSpace, metricLabelCreationTimestamp})

// DBaasStackInstallationHistogram defines a histogram for DBaasStackInstallation
var DBaasStackInstallationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaaSStackInstallationTotalDuration,
	Help: "Time in seconds installation of a DBaaS stack takes.",
	Buckets: prometheus.LinearBuckets(
		installationTimeStart,
		installationTimeWidth,
		installationTimeBuckets),
}, []string{metricLabelVersion, metricLabelCreationTimestamp})

// DBaasInventoryRequestDurationSeconds defines a histogram for DBaasInventoryRequestDuration in seconds
var DBaasInventoryRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaasInventoryDuration,
	Help: "Request/Response duration of provider account of upstream calls to provider operator/service endpoints",
	Buckets: prometheus.LinearBuckets(installationTimeStart,
		installationTimeWidth,
		installationTimeBuckets),
}, []string{metricLabelProvider, metricLabelName, metricLabelNameSpace, metricLabelCreationTimestamp})

// DBaasOperatorVersionInfo defines a gauge for DBaaS Operator version
var DBaasOperatorVersionInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameOperatorVersion,
	Help: "The current version of DBaaS Operator installed in the cluster",
}, []string{metricLabelVersion, metricLabelConsoleULR, metricLabelPlatformName})

// DBaasConnectionRequestDurationSeconds defines a histogram for DBaasConnectionRequestDuration
var DBaasConnectionRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaasConnectionDuration,
	Help: "Request/Response duration of connection of upstream calls to provider operator/service endpoints",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceID, metricLabelConnectionName, metricLabelNameSpace, metricLabelCreationTimestamp})

// DBaasInstanceRequestDurationSeconds defines a histogram for DBaasInstanceRequestDuration
var DBaasInstanceRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaasInstanceDuration,
	Help: "Request/Response duration of instance of upstream calls to provider operator/service endpoints",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceName, metricLabelNameSpace, metricLabelCreationTimestamp})

// Execution tracks state for an API execution for emitting metrics
type Execution struct {
	begin time.Time
}

// PlatformInstallStart creates an Execution instance and starts the timer
func PlatformInstallStart() Execution {
	return Execution{
		begin: time.Now(),
	}
}

// PlatformStackInstallationMetric is used to log duration and success/failure
func (e *Execution) PlatformStackInstallationMetric(platform *dbaasv1alpha1.DBaaSPlatform, version string) {
	duration := time.Since(e.begin)
	for _, cond := range platform.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSPlatformReadyType {
			lastTransitionTime := cond.LastTransitionTime
			duration = lastTransitionTime.Sub(platform.CreationTimestamp.Time)
			DBaasStackInstallationHistogram.With(prometheus.Labels{metricLabelVersion: version, metricLabelCreationTimestamp: platform.CreationTimestamp.String()}).Observe(duration.Seconds())
		} else {
			DBaasStackInstallationHistogram.With(prometheus.Labels{metricLabelVersion: version, metricLabelCreationTimestamp: platform.CreationTimestamp.String()}).Observe(duration.Seconds())
		}
	}
}

// SetPlatformStatusMetric exposes dbaas_platform_status metric for each platform
func SetPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 {
		switch status {

		case dbaasv1alpha1.ResultFailed:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(status), metricLabelVersion: version}).Set(float64(0))
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultSuccess), metricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultInProgress), metricLabelVersion: version})
		case dbaasv1alpha1.ResultSuccess:
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultInProgress), metricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultFailed), metricLabelVersion: version})
			DBaasPlatformInstallationGauge.With(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(status), metricLabelVersion: version}).Set(float64(1))
		case dbaasv1alpha1.ResultInProgress:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(status), metricLabelVersion: version}).Set(float64(2))
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultSuccess), metricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultFailed), metricLabelVersion: version})

		}
	}
}

// CleanPlatformStatusMetric delete the dbaas_platform_status metric for each platform
func CleanPlatformStatusMetric(platformName dbaasv1alpha1.PlatformsName, status dbaasv1alpha1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 && status == dbaasv1alpha1.ResultSuccess {
		DBaasPlatformInstallationGauge.Delete(prometheus.Labels{metricLabelName: string(platformName), metricLabelStatus: string(dbaasv1alpha1.ResultSuccess), metricLabelVersion: version})
	}
}

// SetOpenShiftInstallationInfoMetric set the metrics for openshift info
func SetOpenShiftInstallationInfoMetric(operatorVersion string, consoleURL string, platformType string) {
	DBaasOperatorVersionInfo.With(prometheus.Labels{metricLabelVersion: operatorVersion, metricLabelConsoleULR: consoleURL, metricLabelPlatformName: platformType}).Set(1)
}

// SetInventoryMetrics set the metrics for inventory
func SetInventoryMetrics(inventory dbaasv1alpha1.DBaaSInventory, execution Execution) {
	setInventoryStatusMetrics(inventory)
	setInventoryRequestDurationSeconds(execution, inventory)
}

// setInventoryStatusMetrics set the metrics for inventory status
func setInventoryStatusMetrics(inventory dbaasv1alpha1.DBaaSInventory) {
	for _, cond := range inventory.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInventoryReadyType {
			DBaaSInventoryStatusGauge.DeletePartialMatch(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace})
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, metricLabelCreationTimestamp: inventory.CreationTimestamp.String()}).Set(1)
			} else {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, metricLabelCreationTimestamp: inventory.CreationTimestamp.String()}).Set(0)
			}
			break
		}
	}
}

// setInventoryRequestDurationSeconds set the metrics for inventory request duration in seconds
func setInventoryRequestDurationSeconds(execution Execution, inventory dbaasv1alpha1.DBaaSInventory) {
	httpDuration := time.Since(execution.begin)
	for _, cond := range inventory.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInventoryProviderSyncType {
			if cond.Status == metav1.ConditionTrue {
				lastTransitionTime := cond.LastTransitionTime
				httpDuration = lastTransitionTime.Sub(inventory.CreationTimestamp.Time)
				DBaasInventoryRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name,
					metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelCreationTimestamp: inventory.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			} else {
				DBaasInventoryRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name,
					metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelCreationTimestamp: inventory.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			}
			break
		}
	}
}

// CleanInventoryMetrics delete inventory metrics based on the condition type
func CleanInventoryMetrics(inventory *dbaasv1alpha1.DBaaSInventory) {
	for _, cond := range inventory.Status.Conditions {
		switch cond.Type {
		case dbaasv1alpha1.DBaaSInventoryReadyType:
			DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, metricLabelCreationTimestamp: inventory.CreationTimestamp.String()})
		case dbaasv1alpha1.DBaaSInventoryProviderSyncType:
			DBaasInventoryRequestDurationSeconds.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelCreationTimestamp: inventory.CreationTimestamp.String()})
		}
	}
}

// SetConnectionMetrics set the metrics for a connection
func SetConnectionMetrics(provider string, account string, connection dbaasv1alpha1.DBaaSConnection, execution Execution) {
	setConnectionStatusMetrics(provider, account, connection)
	setConnectionRequestDurationSeconds(provider, account, connection, execution)
}

// setConnectionStatusMetrics set the metrics based on connection status
func setConnectionStatusMetrics(provider string, account string, connection dbaasv1alpha1.DBaaSConnection) {
	for _, cond := range connection.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSConnectionReadyType {
			DBaaSConnectionStatusGauge.DeletePartialMatch(prometheus.Labels{metricLabelName: connection.Name, metricLabelNameSpace: connection.Namespace})
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceID: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, metricLabelCreationTimestamp: connection.CreationTimestamp.String()}).Set(1)
			} else {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceID: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, metricLabelCreationTimestamp: connection.CreationTimestamp.String()}).Set(0)
			}
			break
		}
	}
}

// setConnectionRequestDurationSeconds set the metrics for connection request duration in seconds
func setConnectionRequestDurationSeconds(provider string, account string, connection dbaasv1alpha1.DBaaSConnection, execution Execution) {
	httpDuration := time.Since(execution.begin)
	for _, cond := range connection.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSConnectionProviderSyncType {
			if cond.Status == metav1.ConditionTrue {
				lastTransitionTime := cond.LastTransitionTime
				httpDuration = lastTransitionTime.Sub(connection.CreationTimestamp.Time)
				DBaasConnectionRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceID: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(),
					metricLabelNameSpace: connection.Namespace, metricLabelCreationTimestamp: connection.CreationTimestamp.String()}).Observe(httpDuration.Seconds())

			} else {
				DBaasConnectionRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceID: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(),
					metricLabelNameSpace: connection.Namespace, metricLabelCreationTimestamp: connection.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			}
			break
		}
	}
}

// CleanConnectionMetrics delete connection metrics based on the condition type
func CleanConnectionMetrics(connection *dbaasv1alpha1.DBaaSConnection) {
	for _, cond := range connection.Status.Conditions {
		switch cond.Type {
		case dbaasv1alpha1.DBaaSConnectionReadyType:
			DBaaSConnectionStatusGauge.DeletePartialMatch(prometheus.Labels{metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace})
		case dbaasv1alpha1.DBaaSConnectionProviderSyncType:
			DBaasConnectionRequestDurationSeconds.DeletePartialMatch(prometheus.Labels{metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace})
		}
	}
}

// SetInstanceMetrics set the metrics for an instance
func SetInstanceMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance, execution Execution) {
	setInstanceStatusMetrics(provider, account, instance)
	setInstancePhaseMetrics(provider, account, instance)
	setInstanceRequestDurationSeconds(provider, account, instance, execution)

}

// setInstanceStatusMetrics set the metrics based on instance status
func setInstanceStatusMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance) {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInstanceReadyType {
			DBaaSInstanceStatusGauge.DeletePartialMatch(prometheus.Labels{metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace})
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, metricLabelCreationTimestamp: instance.CreationTimestamp.String()}).Set(1)
			} else {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, metricLabelCreationTimestamp: instance.CreationTimestamp.String()}).Set(0)
			}
			break
		}
	}
}

// setInstanceRequestDurationSeconds set the metrics for instance request duration in seconds
func setInstanceRequestDurationSeconds(provider string, account string, instance dbaasv1alpha1.DBaaSInstance, execution Execution) {
	httpDuration := time.Since(execution.begin)
	for _, cond := range instance.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInstanceProviderSyncType {
			if cond.Status == metav1.ConditionTrue {
				lastTransitionTime := cond.LastTransitionTime
				httpDuration = lastTransitionTime.Sub(instance.CreationTimestamp.Time)
				DBaasInstanceRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.GetNamespace(), metricLabelCreationTimestamp: instance.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			} else {
				DBaasInstanceRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.GetNamespace(), metricLabelCreationTimestamp: instance.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			}
			break
		}
	}
}

// setInstancePhaseMetrics set the metrics for instance phase
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
		metricLabelProvider:          provider,
		metricLabelAccountName:       account,
		metricLabelInstanceName:      instance.Name,
		metricLabelNameSpace:         instance.Namespace,
		metricLabelCreationTimestamp: instance.CreationTimestamp.String(),
	}).Set(phase)
}

// CleanInstanceMetrics delete instance metrics based on the condition type
func CleanInstanceMetrics(instance *dbaasv1alpha1.DBaaSInstance) {
	for _, cond := range instance.Status.Conditions {
		switch cond.Type {
		case dbaasv1alpha1.DBaaSInstanceReadyType:
			DBaaSInstanceStatusGauge.DeletePartialMatch(prometheus.Labels{metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace})
			DBaaSInstancePhaseGauge.DeletePartialMatch(prometheus.Labels{metricLabelInstanceName: instance.Name, metricLabelNameSpace: instance.Namespace})
		case dbaasv1alpha1.DBaaSInstanceProviderSyncType:
			DBaasInstanceRequestDurationSeconds.DeletePartialMatch(prometheus.Labels{metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.GetNamespace()})
		}
	}
}
