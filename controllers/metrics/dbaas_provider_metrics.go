package metrics

import (
	"time"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// Resource label values
	LabelResourceValueProvider = "dbaas_provider"

	// Error Code Values
	LabelErrorCdValueErrorFetchingDBaaSProviderResources = "error_fetching_dbaas_provider_resource"
	LabelErrorCdValueErrorWatchingInventoryCR            = "error_watching_inventory_cr"
	LabelErrorCdValueErrorWatchingConnectionCR           = "error_watching_connection_cr"
	LabelErrorCdValueErrorWatchingInstanceCR             = "error_watching_instance_cr"
	LabelErrorCdValueErrorDeletingProvider               = "error_deleting_dbaas_provider"
)

// setProviderRequestDurationSeconds set the metrics for provider request duration in seconds
func setProviderRequestDurationSeconds(provider dbaasv1beta1.DBaaSProvider, account string, execution Execution, event string) {
	log := ctrl.Log.WithName("DBaaSProvider Request Duration for event: " + event)
	switch event {
	case LabelEventValueCreate:
		duration := time.Now().UTC().Sub(provider.CreationTimestamp.Time.UTC())
		UpdateRequestsDurationHistogram(provider.Name, account, provider.Namespace, LabelResourceValueProvider, event, duration.Seconds())
		log.Info("Set the request duration for create event")
	case LabelEventValueDelete:
		deletionTimestamp := execution.begin.UTC()
		if provider.DeletionTimestamp != nil {
			deletionTimestamp = provider.DeletionTimestamp.UTC()
		}

		duration := time.Now().UTC().Sub(deletionTimestamp.UTC())
		UpdateRequestsDurationHistogram(provider.Name, account, provider.Namespace, LabelResourceValueProvider, event, duration.Seconds())
		log.Info("Set the request duration for delete event")
	}
}

// SetProviderMetrics set the metrics for a provider
func SetProviderMetrics(provider dbaasv1beta1.DBaaSProvider, account string, execution Execution, event string, errCd string) {
	log := ctrl.Log.WithName("Setting DBaaSProvider Metrics")
	log.Info("provider - " + provider.Name + " account - " + account + " namespace - " + provider.Namespace + " event - " + event + " errCd - " + errCd)
	setProviderRequestDurationSeconds(provider, account, execution, event)
	UpdateErrorsTotal(provider.Name, account, provider.Namespace, LabelResourceValueProvider, event, errCd)
}
