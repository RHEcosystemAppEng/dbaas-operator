package metrics

import (
	"time"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// Resource label values
	LabelResourceValuePolicy = "dbaas_policy"

	// Error Code label values
	LabelErrorCdValueErrorFetchingDBaaSPolicyResources = "error_fetching_dbaas_policy_resource"
	LabelErrorCdValueErrorResourceQuotaModified        = "error_resource_quota_modified"
	LabelErrorCdValueErrUpdatingResourceQuota          = "error_updating_resource_quota"
	LabelErrorCdValueErrorDeletingPolicy               = "error_deleting_dbaas_policy"
)

// SetPolicyMetrics set the Metrics for policy
func SetPolicyMetrics(policy dbaasv1beta1.DBaaSPolicy, execution Execution, event string, errCd string) {
	setPolicyRequestDurationSeconds(policy, event, execution)
	UpdateErrorsTotal(policy.Name, policy.Name, policy.Namespace, LabelResourceValuePolicy, event, errCd)
}

// setPolicyRequestDurationSeconds set the Metrics for policy request duration in seconds
func setPolicyRequestDurationSeconds(policy dbaasv1beta1.DBaaSPolicy, event string, execution Execution) {
	log := ctrl.Log.WithName("Policy Request Duration for event: " + event)
	switch event {

	case LabelEventValueCreate:
		duration := time.Now().UTC().Sub(policy.CreationTimestamp.Time.UTC())
		UpdateRequestsDurationHistogram(policy.Name, policy.Name, policy.Namespace, LabelResourceValuePolicy, event, duration.Seconds())
		log.Info("Set the request duration for create event")

	case LabelEventValueDelete:
		deletionTimestamp := execution.begin.UTC()
		if policy.DeletionTimestamp != nil {
			deletionTimestamp = policy.DeletionTimestamp.UTC()
		}

		duration := time.Now().UTC().Sub(deletionTimestamp.UTC())
		UpdateRequestsDurationHistogram(policy.Name, policy.Name, policy.Namespace, LabelResourceValuePolicy, event, duration.Seconds())
		log.Info("Set the request duration for delete event")
	}
}
