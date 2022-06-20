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
	metricNameConnectionStatusReady               = "dbaas_connection_status_ready"
	metricNameInstanceStatusReady                 = "dbaas_instance_status_ready"
	metricNameInstancePhase                       = "dbaas_instance_phase"
	metricNameDBaasInventoryDuration              = "dbaas_inventory_request_duration_seconds"
	metricNameOperatorVersion                     = "dbaas_operator_version"
	metricNameDBaasConnectionDuration             = "dbaas_connection_request_duration_seconds"
	metricNameDBaasInstanceDuration               = "dbaas_instance_request_duration_seconds"

	// Metrics labels.
	metricLabelName           = "name"
	metricLabelStatus         = "status"
	metricLabelVersion        = "version"
	metricLabelProvider       = "provider"
	metricLabelAccountName    = "account"
	metricLabelConnectionName = "connection"
	metricLabelNameSpace      = "namespace"
	metricLabelInstanceId     = "instance_id"
	metricLabelReason         = "reason"
	metricLabelInstanceName   = "instance_name"

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
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelNameSpace, metricLabelStatus, metricLabelReason})

var DBaaSConnectionStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameConnectionStatusReady,
	Help: "The status of DBaaS connections, values ( ready=1, error / not ready=0 )",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceId, metricLabelConnectionName, metricLabelNameSpace, metricLabelStatus, metricLabelReason})

var DBaaSInstanceStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInstanceStatusReady,
	Help: "The status of DBaaS instance, values ( ready=1, error / not ready=0 )",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceName, metricLabelNameSpace, metricLabelStatus, metricLabelReason})

var DBaaSInstancePhaseGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInstancePhase,
	Help: "Current status phase of the Instance currently managed by RHODA values ( Pending=-1, Creating=0, Ready=1, Unknown=2, Failed=3, Error=4, Deleting=5 ).",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceName, metricLabelNameSpace})

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
}, []string{metricLabelProvider, metricLabelInstanceName, metricLabelNameSpace})

var DBaasOperatorVersionInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameOperatorVersion,
	Help: "The current versoin of DBaas Operator installed in the cluster",
}, []string{metricLabelVersion})

var DBaasConnectionRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaasConnectionDuration,
	Help: "Request/Response duration of connection account request of upstream calls to provider operator/service endpoints",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceId, metricLabelConnectionName, metricLabelNameSpace, metricLabelStatus})

var DBaasInstanceRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: metricNameDBaasInstanceDuration,
	Help: "Request/Response duration of instance of upstream calls to provider operator/service endpoints",
}, []string{metricLabelProvider, metricLabelAccountName, metricLabelInstanceName, metricLabelNameSpace, metricLabelStatus})

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
	DBaasOperatorVersionInfo.With(prometheus.Labels{metricLabelVersion: version}).Set(0)
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

// SetInventoryStatusMetrics
func SetInventoryStatusMetrics(inventory dbaasv1alpha1.DBaaSInventory) {
	for _, cond := range inventory.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInventoryReadyType {
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason}).Set(1)
				DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.ProviderReconcileInprogress})
				DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSProviderNotFound})
			} else if cond.Reason == dbaasv1alpha1.ProviderReconcileInprogress && cond.Status == metav1.ConditionFalse {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason}).Set(0)
				DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(metav1.ConditionTrue), metricLabelReason: dbaasv1alpha1.Ready})
				DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSProviderNotFound})
			} else {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason}).Set(0)
			}
		}
	}
}

func SetInventoryRequestDurationSeconds(execution Execution, inventory dbaasv1alpha1.DBaaSInventory) {
	httpDuration := time.Since(execution.begin)
	for _, cond := range inventory.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInventoryReadyType {
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaasInventoryRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name,
					metricLabelInstanceName: inventory.Name, metricLabelNameSpace: inventory.Namespace}).Observe(httpDuration.Seconds())
			}
		}
	}
}

func SetConnectionRequestDurationSeconds(execution Execution, inventory dbaasv1alpha1.DBaaSInventory, connection v1alpha1.DBaaSConnection) {
	httpDuration := time.Since(execution.begin)
	DBaasConnectionRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: inventory.Namespace, metricLabelStatus: ""}).Observe(httpDuration.Seconds())
}

func SetInstanceRequestDurationSeconds(execution Execution, inventory dbaasv1alpha1.DBaaSInventory, instance v1alpha1.DBaaSInstance) {
	httpDuration := time.Since(execution.begin)
	DBaasInstanceRequestDurationSeconds.With(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.GetNamespace(), metricLabelStatus: ""}).Observe(httpDuration.Seconds())
}

// CleanInventoryStatusMetrics
func CleanInventoryStatusMetrics(inventory *dbaasv1alpha1.DBaaSInventory) {
	for _, cond := range inventory.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInventoryReadyType {
			DBaaSInventoryStatusGauge.Delete(prometheus.Labels{metricLabelProvider: inventory.Spec.ProviderRef.Name, metricLabelAccountName: inventory.Name, metricLabelNameSpace: inventory.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason})
		}
	}
}

// SetConnectionStatusMetrics
func SetConnectionStatusMetrics(provider string, account string, connection dbaasv1alpha1.DBaaSConnection) {
	for _, cond := range connection.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSConnectionReadyType {
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason}).Set(1)
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.ProviderReconcileInprogress})
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotFound})
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotReady})
			} else if cond.Reason == dbaasv1alpha1.ProviderReconcileInprogress && cond.Status == metav1.ConditionFalse {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason}).Set(0)
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionTrue), metricLabelReason: dbaasv1alpha1.Ready})
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotFound})
				DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotReady})

			} else {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason}).Set(0)
			}
		}
	}
}

// CleanConnectionStatusMetrics
func CleanConnectionStatusMetrics(provider string, account string, connection *dbaasv1alpha1.DBaaSConnection) {
	for _, cond := range connection.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSConnectionReadyType {
			DBaaSConnectionStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceId: connection.Spec.InstanceID, metricLabelConnectionName: connection.GetName(), metricLabelNameSpace: connection.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason})
		}
	}
}

// SetInstanceMetrics
func SetInstanceMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance) {
	setInstanceStatusMetrics(provider, account, instance)
	setInstancePhaseMetrics(provider, account, instance)
}

// setInstanceStatusMetrics
func setInstanceStatusMetrics(provider string, account string, instance dbaasv1alpha1.DBaaSInstance) {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInstanceReadyType {
			if cond.Reason == dbaasv1alpha1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason}).Set(1)
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.ProviderReconcileInprogress})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotFound})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSProviderNotFound})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotReady})

			} else if cond.Reason == dbaasv1alpha1.ProviderReconcileInprogress && cond.Status == metav1.ConditionFalse {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason}).Set(0)
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionTrue), metricLabelReason: dbaasv1alpha1.Ready})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotFound})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSProviderNotFound})
				DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(metav1.ConditionFalse), metricLabelReason: dbaasv1alpha1.DBaaSInventoryNotReady})
			} else {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason}).Set(0)
			}
		}
	}
}

// CleanInstanceStatusMetrics
func CleanInstanceStatusMetrics(provider string, account string, instance *dbaasv1alpha1.DBaaSInstance) {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == dbaasv1alpha1.DBaaSInstanceReadyType {
			DBaaSInstanceStatusGauge.Delete(prometheus.Labels{metricLabelProvider: provider, metricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), metricLabelNameSpace: instance.Namespace, metricLabelStatus: string(cond.Status), metricLabelReason: cond.Reason})
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
	}).Set(phase)

}
