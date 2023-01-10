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
	MetricNameConnectionStatusReady = "dbaas_connection_status_ready"

	// Resource label values
	LabelResourceValueConnection = "dbaas_connection"

	// Error Code Values
	LabelErrorCdValueDevTopologyModified           = "dbaas_connection_dev_topology_modified"
	LabelErrorCdValueErrReconcilingWithDevTopology = "dbaas_connection_err_reconciling_with_dev_topology"
	LabelErrorCdValueInvalidNameSpace              = "dbaas_connection_err_invalid_name_space"
	LabelErrorCdCannotReadInstance                 = "dbaas_connection_cannot_read_instance"
	LabelErrorCdValueErrorDeletingConnection       = "error_deleting_connection"

	// Label Names
	MetricLabelConnectionName = "name"
)

// DBaaSConnectionStatusGauge defines a gauge for DBaaSConnectionStatus
var DBaaSConnectionStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameConnectionStatusReady,
	Help: "The status of DBaaS connections, values ( ready=1, error / not ready=0 )",
}, []string{MetricLabelProvider, MetricLabelAccountName, MetricLabelInstanceID, MetricLabelConnectionName, MetricLabelNameSpace, MetricLabelStatus, MetricLabelReason})

// SetConnectionMetrics set the Metrics for a connection
func SetConnectionMetrics(provider string, account string, connection dbaasv1beta1.DBaaSConnection, execution Execution, event string, errCd string) {
	log := ctrl.Log.WithName("Setting DBaaSConnection Metrics")
	log.Info("provider - " + provider + " account - " + account + " namespace - " + connection.Namespace + " event - " + event + " errCd - " + errCd)
	setConnectionStatusMetrics(provider, account, connection)
	setConnectionRequestDurationSeconds(provider, account, connection, execution, event)
	UpdateErrorsTotal(provider, account, connection.Namespace, LabelResourceValueConnection, event, errCd)
}

// setConnectionStatusMetrics set the Metrics based on connection status
func setConnectionStatusMetrics(provider string, account string, connection dbaasv1beta1.DBaaSConnection) {
	for _, cond := range connection.Status.Conditions {
		if cond.Type == dbaasv1beta1.DBaaSConnectionReadyType {
			DBaaSConnectionStatusGauge.DeletePartialMatch(prometheus.Labels{MetricLabelName: connection.Name, MetricLabelNameSpace: connection.Namespace})
			if cond.Reason == dbaasv1beta1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{MetricLabelProvider: provider, MetricLabelAccountName: account, MetricLabelInstanceID: connection.Spec.InstanceID, MetricLabelConnectionName: connection.GetName(), MetricLabelNameSpace: connection.Namespace, MetricLabelStatus: string(cond.Status), MetricLabelReason: cond.Reason}).Set(1)
			} else {
				DBaaSConnectionStatusGauge.With(prometheus.Labels{MetricLabelProvider: provider, MetricLabelAccountName: account, MetricLabelInstanceID: connection.Spec.InstanceID, MetricLabelConnectionName: connection.GetName(), MetricLabelNameSpace: connection.Namespace, MetricLabelStatus: string(cond.Status), MetricLabelReason: cond.Reason}).Set(0)
			}
			break
		}
	}
}

// setConnectionRequestDurationSeconds set the Metrics for connection request duration in seconds
func setConnectionRequestDurationSeconds(provider string, account string, connection dbaasv1beta1.DBaaSConnection, execution Execution, event string) {
	log := ctrl.Log.WithName("Connection Request Duration for event: " + event)
	switch event {
	case LabelEventValueCreate:
		for _, cond := range connection.Status.Conditions {
			if cond.Type == dbaasv1beta1.DBaaSConnectionProviderSyncType {
				if cond.Status == metav1.ConditionTrue {
					duration := time.Now().UTC().Sub(connection.CreationTimestamp.Time.UTC())
					UpdateRequestsDurationHistogram(connection.Spec.InventoryRef.Name, connection.Name, connection.Namespace, LabelResourceValueConnection, event, duration.Seconds())
					log.Info("Set the request duration for create event")
				}
			}
		}
	case LabelEventValueDelete:
		deletionTimestamp := execution.begin.UTC()
		if connection.DeletionTimestamp != nil {
			deletionTimestamp = connection.DeletionTimestamp.UTC()
		}

		duration := time.Now().UTC().Sub(deletionTimestamp.UTC())
		UpdateRequestsDurationHistogram(connection.Spec.InventoryRef.Name, connection.Name, connection.Namespace, LabelResourceValueConnection, event, duration.Seconds())
		log.Info("Set the request duration for delete event")
	}
}
