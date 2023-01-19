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
	MetricNameInventoryStatusReady = "dbaas_inventory_status_ready"

	// Resource label values
	LabelResourceValueInventory = "dbaas_inventory"

	// Error Code Values
	LabelErrorCdValueErrorFetchingDBaaSInventoryResources = "error_fetching_dbaas_inventory_resources"
	LabelErrorCdValueErrorUpdatingInventoryStatus         = "error_updating_inventory_status"
	LabelErrorCdValueErrorDeletingInventory               = "error_deleting_inventory"
	LabelErrorCdValueErrCheckingInventory                 = "error_checking_inventory"
)

// DBaaSInventoryStatusGauge defines a gauge for DBaaSInventoryStatus
var DBaaSInventoryStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameInventoryStatusReady,
	Help: "The status of DBaaS Provider Account, values ( ready=1, error / not ready=0 )",
}, []string{MetricLabelProvider, MetricLabelName, MetricLabelNameSpace, MetricLabelStatus, MetricLabelReason})

// SetInventoryMetrics set the Metrics for inventory
func SetInventoryMetrics(inventory dbaasv1beta1.DBaaSInventory, execution Execution, event string, errCd string) {
	setInventoryStatusMetrics(inventory)
	setInventoryRequestDurationSeconds(inventory, event, execution)
	UpdateErrorsTotal(inventory.Spec.ProviderRef.Name, inventory.Name, inventory.Namespace, LabelResourceValueInventory, event, errCd)
}

// setInventoryStatusMetrics set the Metrics for inventory status
func setInventoryStatusMetrics(inventory dbaasv1beta1.DBaaSInventory) {
	for _, cond := range inventory.Status.Conditions {
		if cond.Type == dbaasv1beta1.DBaaSInventoryReadyType {
			DBaaSInventoryStatusGauge.DeletePartialMatch(prometheus.Labels{MetricLabelProvider: inventory.Spec.ProviderRef.Name, MetricLabelName: inventory.Name, MetricLabelNameSpace: inventory.Namespace})
			if cond.Reason == dbaasv1beta1.Ready && cond.Status == metav1.ConditionTrue {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{MetricLabelProvider: inventory.Spec.ProviderRef.Name, MetricLabelName: inventory.Name, MetricLabelNameSpace: inventory.Namespace, MetricLabelStatus: string(cond.Status), MetricLabelReason: cond.Reason}).Set(1)
			} else {
				DBaaSInventoryStatusGauge.With(prometheus.Labels{MetricLabelProvider: inventory.Spec.ProviderRef.Name, MetricLabelName: inventory.Name, MetricLabelNameSpace: inventory.Namespace, MetricLabelStatus: string(cond.Status), MetricLabelReason: cond.Reason}).Set(0)
			}
			break
		}
	}
}

// setInventoryRequestDurationSeconds set the Metrics for inventory request duration in seconds
func setInventoryRequestDurationSeconds(inventory dbaasv1beta1.DBaaSInventory, event string, execution Execution) {
	log := ctrl.Log.WithName("Inventory Request Duration for event: " + event)
	switch event {

	case LabelEventValueCreate:
		for _, cond := range inventory.Status.Conditions {
			if cond.Type == dbaasv1beta1.DBaaSInventoryProviderSyncType {
				if cond.Status == metav1.ConditionTrue {
					duration := time.Now().UTC().Sub(inventory.CreationTimestamp.Time.UTC())
					UpdateRequestsDurationHistogram(inventory.Spec.ProviderRef.Name, inventory.Name, inventory.Namespace, LabelResourceValueInventory, event, duration.Seconds())
					log.Info("Set the request duration for create event")
				}
				break
			}
		}

	case LabelEventValueDelete:
		deletionTimestamp := execution.begin.UTC()
		if inventory.DeletionTimestamp != nil {
			deletionTimestamp = inventory.DeletionTimestamp.UTC()
		}

		duration := time.Now().UTC().Sub(deletionTimestamp.UTC())
		UpdateRequestsDurationHistogram(inventory.Spec.ProviderRef.Name, inventory.Name, inventory.Namespace, LabelResourceValueInventory, event, duration.Seconds())
		log.Info("Set the request duration for delete event")
	}
}
