package controllers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
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
	metricNameOperatorVersion                     = "dbaas_operator_version"

	// Metrics labels.
	metricLabelName           = "name"
	metricLabelStatus         = "status"
	metricLabelVersion        = "version"
	metricLabelProvider       = "provider"
	metricLabelAccountName    = "account"
	metricLabelConnectionName = "name"
	metricLabelNameSpace      = "namespace"
	metricLabelInstanceId     = "instance_id"
	metricLabelReason         = "reason"
	metricLabelInstanceName   = "name"
	creationTimestamp         = "creation_timestamp"

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

var DBaaSInventoryStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInventoryStatusReady,
	Help: "The status of DBaaS Provider Account, values ( ready=1, error / not ready=0 )",
}, []string{metricLabelProvider, metricLabelName, metricLabelNameSpace, metricLabelStatus, metricLabelReason, creationTimestamp})

var DBaaSConnectionStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameConnectionStatusReady,
	Help: "The status of DBaaS connections, values ( ready=1, error / not ready=0 )",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceId, metricLabelConnectionName, metricLabelNameSpace, metricLabelStatus, metricLabelReason, creationTimestamp})

var DBaaSInstanceStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInstanceStatusReady,
	Help: "The status of DBaaS instance, values ( ready=1, error / not ready=0 )",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceName, metricLabelNameSpace, metricLabelStatus, metricLabelReason, creationTimestamp})

var DBaaSInstancePhaseGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInstancePhase,
	Help: "Current status phase of the Instance currently managed by RHODA values ( Pending=-1, Creating=0, Ready=1, Unknown=2, Failed=3, Error=4, Deleting=5 ).",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceName, metricLabelNameSpace, creationTimestamp})

var DBaasStackInstallationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaaSStackInstallationTotalDuration,
	Help: "How long in seconds installation of a DBaaS stack takes.",
	Buckets: prometheus.LinearBuckets(
		installationTimeStart,
		installationTimeWidth,
		installationTimeBuckets),
}, []string{metricLabelVersion})

var DBaasInventoryRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaasInventoryDuration,
	Help: "Duration of upstream calls to provider operator/service endpoints",
	Buckets: prometheus.LinearBuckets(installationTimeStart,
		installationTimeWidth,
		installationTimeBuckets),
}, []string{metricLabelProvider, metricLabelName, metricLabelNameSpace, creationTimestamp})

var DBaasOperatorVersionInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameOperatorVersion,
	Help: "The current version of DBaas Operator installed in the cluster",
}, []string{metricLabelVersion})

var DBaasConnectionRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaasConnectionDuration,
	Help: "Request/Response duration of connection account request of upstream calls to provider operator/service endpoints",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceId, metricLabelConnectionName, metricLabelNameSpace, creationTimestamp})

var DBaasInstanceRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaasInstanceDuration,
	Help: "Request/Response duration of instance of upstream calls to provider operator/service endpoints",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceName, metricLabelNameSpace, creationTimestamp})

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
	DBaasOperatorVersionInfo.With(prometheus.Labels{metricLabelVersion: version}).Set(1)
	DBaasStackInstallationHistogram.With(prometheus.Labels{metricLabelVersion: version}).Observe(duration.Seconds())
}

// SetPlatformStatusMetric exposes dbaas_platform_status metric for each platform
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

// SetInventoryMetrics
func SetInventoryMetrics(inventory dbaasv1alpha1.DBaaSInventory, execution Execution) {
	setInventoryStatusMetrics(inventory)
	setInventoryRequestDurationSeconds(execution, inventory)
}

// setInventoryStatusMetrics
func setInventoryStatusMetrics(inventory dbaasv1alpha1.DBaaSInventory) {
	for _, cond := range inventory.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInventoryReadyType {
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: inventory.CreationTimestamp.String()}).Set(1)
				DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.ProviderReconcileInprogress, creationTimestamp: inventory.CreationTimestamp.String()})
				DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSProviderNotFound, creationTimestamp: inventory.CreationTimestamp.String()})
			} else if cond.Reason == dbaasv1alpha1.ProviderReconcileInprogress && cond.Status == metav1.ConditionFalse {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: inventory.CreationTimestamp.String()}).Set(0)
				DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(metav1.ConditionTrue), metricLabelReason: dbaasv1alpha1.Ready, creationTimestamp: inventory.CreationTimestamp.String()})
				DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSProviderNotFound, creationTimestamp: inventory.CreationTimestamp.String()})
			} else {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: inventory.CreationTimestamp.String()}).Set(0)
			}
		}
	}
}

// setInventoryRequestDurationSeconds
func setInventoryRequestDurationSeconds(execution Execution, inventory dbaasv1alpha1.DBaaSInventory) {
	httpDuration := time.Since(execution.begin)
	lastTransitionTime := metav1.Time{}
	for _, cond := range inventory.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInventoryProviderSyncType {
			if cond.Status == metav1.ConditionTrue {
				lastTransitionTime = cond.LastTransitionTime
				httpDuration = lastTransitionTime.Sub(inventory.CreationTimestamp.Time)
				DBaasInventoryRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name,
					metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, creationTimestamp: inventory.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			} else {
				DBaasInventoryRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name,
					metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, creationTimestamp: inventory.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			}
			break
		}
	}
}

// CleanInventoryMetrics
func CleanInventoryMetrics(inventory *dbaasv1alpha1.DBaaSInventory) {
	for _, cond := range inventory.Status.Conditions {
		switch cond.Type {
		case dbaasv1alpha1.DBaaSInventoryReadyType:
			DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: inventory.CreationTimestamp.String()})
		case dbaasv1alpha1.DBaaSInventoryProviderSyncType:
			DBaasInventoryRequestDurationSeconds.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelName: inventory.Name, metricLabelNameSpace: inventory.Namespace, creationTimestamp: inventory.CreationTimestamp.String()})
		}
	}
}

// SetConnectionMetrics
func SetConnectionMetrics(provider string, account string, connection dbaasv1alpha1.DBaaSConnection, execution Execution) {
	setConnectionStatusMetrics(provider, account, connection)
	setConnectionRequestDurationSeconds(provider, account, connection, execution)
}

// setConnectionStatusMetrics
func setConnectionStatusMetrics(provider string, account string, connection dbaasv1alpha1.DBaaSConnection) {
	for _, cond := range connection.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSConnectionReadyType {
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: connection.CreationTimestamp.String()}).Set(1)
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.ProviderReconcileInprogress, creationTimestamp: connection.CreationTimestamp.String()})
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotFound, creationTimestamp: connection.CreationTimestamp.String()})
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotReady, creationTimestamp: connection.CreationTimestamp.String()})
			} else if cond.Reason == dbaasv1alpha1.ProviderReconcileInprogress && cond.Status == metav1.ConditionFalse {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: connection.CreationTimestamp.String()}).Set(0)
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionTrue), metricLabelReason: dbaasv1alpha1.Ready, creationTimestamp: connection.CreationTimestamp.String()})
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotFound, creationTimestamp: connection.CreationTimestamp.String()})
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotReady, creationTimestamp: connection.CreationTimestamp.String()})

			} else {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: connection.CreationTimestamp.String()}).Set(0)
			}
		}
	}
}

// setConnectionRequestDurationSeconds
func setConnectionRequestDurationSeconds(provider string, account string, connection v1alpha1.DBaaSConnection, execution Execution) {
	httpDuration := time.Since(execution.begin)
	lastTransitionTime := metav1.Time{}
	for _, cond := range connection.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSConnectionProviderSyncType {
			if cond.Status == metav1.ConditionTrue {
				lastTransitionTime = cond.LastTransitionTime
				httpDuration = lastTransitionTime.Sub(connection.CreationTimestamp.Time)
				DBaasConnectionRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(),
					metricLabelNameSpace: connection.Namespace, creationTimestamp: connection.CreationTimestamp.String()}).Observe(httpDuration.Seconds())

			} else {
				DBaasConnectionRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(),
					metricLabelNameSpace: connection.Namespace, creationTimestamp: connection.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			}
			break
		}
	}
}

// CleanConnectionMetrics
func CleanConnectionMetrics(provider string, account string, connection *dbaasv1alpha1.DBaaSConnection) {
	for _, cond := range connection.Status.Conditions {
		switch cond.Type {
		case dbaasv1alpha1.DBaaSConnectionReadyType:
			DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: connection.CreationTimestamp.String()})
		case dbaasv1alpha1.DBaaSConnectionProviderSyncType:
			DBaasConnectionRequestDurationSeconds.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(),
				metricLabelNameSpace: connection.Namespace, creationTimestamp: connection.CreationTimestamp.String()})
		}
	}
}

// SetInstanceMetrics
func SetInstanceMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance, execution Execution) {
	setInstanceStatusMetrics(provider, account, instance)
	setInstancePhaseMetrics(provider, account, instance)
	setInstanceRequestDurationSeconds(provider, account, instance, execution)

}

// setInstanceStatusMetrics
func setInstanceStatusMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance) {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInstanceReadyType {
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: instance.CreationTimestamp.String()}).Set(1)
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.ProviderReconcileInprogress, creationTimestamp: instance.CreationTimestamp.String()})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotFound, creationTimestamp: instance.CreationTimestamp.String()})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSProviderNotFound, creationTimestamp: instance.CreationTimestamp.String()})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotReady, creationTimestamp: instance.CreationTimestamp.String()})

			} else if cond.Reason == dbaasv1alpha1.ProviderReconcileInprogress && cond.Status == metav1.ConditionFalse {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: instance.CreationTimestamp.String()}).Set(0)
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionTrue), metricLabelReason: dbaasv1alpha1.Ready, creationTimestamp: instance.CreationTimestamp.String()})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotFound, creationTimestamp: instance.CreationTimestamp.String()})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSProviderNotFound, creationTimestamp: instance.CreationTimestamp.String()})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotReady, creationTimestamp: instance.CreationTimestamp.String()})
			} else {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: instance.CreationTimestamp.String()}).Set(0)
			}
		}
	}
}

// setInstanceRequestDurationSeconds
func setInstanceRequestDurationSeconds(provider string, account string, instance v1alpha1.DBaaSInstance, execution Execution) {
	httpDuration := time.Since(execution.begin)
	lastTransitionTime := metav1.Time{}
	for _, cond := range instance.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInstanceProviderSyncType {
			if cond.Status == metav1.ConditionTrue {
				lastTransitionTime = cond.LastTransitionTime
				httpDuration = lastTransitionTime.Sub(instance.CreationTimestamp.Time)
				DBaasInstanceRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.GetNamespace(), creationTimestamp: instance.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			} else {
				DBaasInstanceRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.GetNamespace(), creationTimestamp: instance.CreationTimestamp.String()}).Observe(httpDuration.Seconds())
			}
			break
		}
	}
}

//setInstancePhaseMetrics
func setInstancePhaseMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance) {
	var phase float64

	switch instance.Status.Phase {
	case dbaasv1alpha1.PhasePending:
		phase = -1
	case dbaasv1alpha1.PhaseCreating:
		phase = 0
	case dbaasv1alpha1.PhaseReady:
		phase = 1
	case dbaasv1alpha1.PhaseUnknown:
		phase = 2
	case dbaasv1alpha1.PhaseFailed:
		phase = 3
	case dbaasv1alpha1.PhaseError:
		phase = 4
	case dbaasv1alpha1.PhaseDeleting:
		phase = 5
	case dbaasv1alpha1.PhaseDeleted:
		phase = 6
	}

	DBaaSInstancePhaseGauge.With(prometheus.Labels{
		metricLabelProvider:     provider,
		metricLabelAccountName:  account,
		metricLabelInstanceName: instance.Name,
		metricLabelNameSpace:    instance.Namespace,
		creationTimestamp:       instance.CreationTimestamp.String(),
	}).Set(phase)
}

// CleanInstanceMetrics
func CleanInstanceMetrics(provider string, account string, instance *dbaasv1alpha1.DBaaSInstance) {
	for _, cond := range instance.Status.Conditions {
		switch cond.Type {
		case dbaasv1alpha1.DBaaSInstanceReadyType:
			DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason, creationTimestamp: instance.CreationTimestamp.String()})
			DBaaSInstancePhaseGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.Name, metricLabelNameSpace: instance.Namespace, creationTimestamp: instance.CreationTimestamp.String()})
		case dbaasv1alpha1.DBaaSInstanceProviderSyncType:
			DBaasInstanceRequestDurationSeconds.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.GetNamespace(), creationTimestamp: instance.CreationTimestamp.String()})
		}
	}
}
