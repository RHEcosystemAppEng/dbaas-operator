package metrics

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/prometheus/client_golang/prometheus"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// Metric Names
	metricNameInstanceStatusReady = "dbaas_instance_status_ready"
	metricNameInstancePhase       = "dbaas_instance_phase"

	metricLabelInstanceID   = "instance_id"
	metricLabelInstanceName = "instance_name"

	// Resource label values
	LabelResourceValueInstance = "dbaas_instance"

	// Error Code Values
	LabelErrorCdValueErrorFetchingDBaaSInstance     = "error_fetching_dbaas_instance_resources"
	LabelErrorCdValueErrorCheckingInstanceInventory = "error_checking_dbaas_instance_inventory"
	LabelErrorCdValueErrorDeletingInstance          = "error_deleting_dbaas_instance"
)

// DBaaSInstanceStatusGauge defines a gauge for DBaaSInstanceStatus
var DBaaSInstanceStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInstanceStatusReady,
	Help: "The status of DBaaS instance, values ( ready=1, error / not ready=0 )",
}, []string{MetricLabelProvider, MetricLabelAccountName, metricLabelInstanceName, MetricLabelNameSpace, MetricLabelStatus, MetricLabelReason})

// DBaaSInstancePhaseGauge defines a gauge for DBaaSInstancePhase
var DBaaSInstancePhaseGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: metricNameInstancePhase,
	Help: "Current status phase of the Instance currently managed by RHODA values ( Pending=-1, Creating=0, Ready=1, Unknown=2, Failed=3, Error=4, Deleting=5 ).",
}, []string{MetricLabelProvider, MetricLabelAccountName, metricLabelInstanceName, MetricLabelNameSpace})

// setInstanceStatusMetrics set the metrics based on instance status
func setInstanceStatusMetrics(provider string, account string, instance dbaasv1beta1.DBaaSInstance) {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == dbaasv1beta1.DBaaSInstanceReadyType {
			DBaaSInstanceStatusGauge.DeletePartialMatch(prometheus.Labels{metricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.Namespace})
			if cond.Reason == dbaasv1beta1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{MetricLabelProvider: provider, MetricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.Namespace, MetricLabelStatus: string(cond.Status), MetricLabelReason: cond.Reason}).Set(1)
			} else {
				DBaaSInstanceStatusGauge.With(prometheus.Labels{MetricLabelProvider: provider, MetricLabelAccountName: account, metricLabelInstanceName: instance.GetName(), MetricLabelNameSpace: instance.Namespace, MetricLabelStatus: string(cond.Status), MetricLabelReason: cond.Reason}).Set(0)
			}
			break
		}
	}
}

// setInstanceRequestDurationSeconds set the metrics for instance request duration in seconds
func setInstanceRequestDurationSeconds(provider string, account string, instance dbaasv1beta1.DBaaSInstance, execution Execution, event string) {
	log := ctrl.Log.WithName("DBaaSInstance Request Duration for event: " + event)
	switch event {
	case LabelEventValueCreate:
		for _, cond := range instance.Status.Conditions {
			if cond.Type == dbaasv1beta1.DBaaSInstanceProviderSyncType {
				if cond.Status == metav1.ConditionTrue {
					duration := time.Now().UTC().Sub(instance.CreationTimestamp.Time.UTC())
					UpdateRequestsDurationHistogram(instance.Spec.InventoryRef.Name, instance.Name, instance.Namespace, LabelResourceValueInstance, event, duration.Seconds())
					log.Info("Set the request duration for create event")
				}
			}
		}
	case LabelEventValueDelete:
		deletionTimestamp := execution.begin.UTC()
		if instance.DeletionTimestamp != nil {
			deletionTimestamp = instance.DeletionTimestamp.UTC()
		}

		duration := time.Now().UTC().Sub(deletionTimestamp.UTC())
		UpdateRequestsDurationHistogram(instance.Spec.InventoryRef.Name, instance.Name, instance.Namespace, LabelResourceValueInstance, event, duration.Seconds())
		log.Info("Set the request duration for delete event")
	}
}

// setInstancePhaseMetrics set the metrics for instance phase
func setInstancePhaseMetrics(provider string, account string, instance dbaasv1beta1.DBaaSInstance) {
	var phase float64

	switch instance.Status.Phase {
	case dbaasv1beta1.InstancePhasePending:
		phase = -1
	case dbaasv1beta1.InstancePhaseCreating:
		phase = 0
	case dbaasv1beta1.InstancePhaseReady:
		phase = 1
	case dbaasv1beta1.InstancePhaseUnknown:
		phase = 2
	case dbaasv1beta1.InstancePhaseFailed:
		phase = 3
	case dbaasv1beta1.InstancePhaseError:
		phase = 4
	case dbaasv1beta1.InstancePhaseDeleting:
		phase = 5
	case dbaasv1beta1.InstancePhaseDeleted:
		phase = 6
	}

	DBaaSInstancePhaseGauge.With(prometheus.Labels{
		MetricLabelProvider:     provider,
		MetricLabelAccountName:  account,
		metricLabelInstanceName: instance.Name,
		MetricLabelNameSpace:    instance.Namespace,
	}).Set(phase)
}

// SetInstanceMetrics set the metrics for an instance
func SetInstanceMetrics(provider string, account string, instance dbaasv1beta1.DBaaSInstance, execution Execution, event string, errCd string) {
	log := ctrl.Log.WithName("Setting DBaaSInstance Metrics")
	log.Info("provider - " + provider + " account - " + account + " namespace - " + instance.Namespace + " event - " + event + " errCd - " + errCd)
	setInstanceStatusMetrics(provider, account, instance)
	setInstancePhaseMetrics(provider, account, instance)
	setInstanceRequestDurationSeconds(provider, account, instance, execution, event)
	UpdateErrorsTotal(provider, account, instance.Namespace, LabelResourceValueInstance, event, errCd)
}
