/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var dbaaspolicylog = logf.Log.WithName("dbaaspolicy-resource")

// SetupWebhookWithManager sets up the webhook with the Manager.
func (r *DBaaSPolicy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-dbaas-redhat-com-v1alpha1-dbaaspolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=dbaas.redhat.com,resources=dbaaspolicies,verbs=create;update,versions=v1alpha1,name=vdbaaspolicy.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &DBaaSPolicy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSPolicy) ValidateCreate() error {
	dbaaspolicylog.Info("validate create", "name", r.Name)
	return validatePolicy(r)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSPolicy) ValidateUpdate(_ runtime.Object) error {
	dbaaspolicylog.Info("validate update", "name", r.Name)
	return validatePolicy(r)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSPolicy) ValidateDelete() error {
	dbaaspolicylog.Info("validate delete", "name", r.Name)
	return nil
}

func validatePolicy(policy *DBaaSPolicy) error {
	// Check ns selector
	if policy.Spec.ConnectionNsSelector != nil {
		if _, err := metav1.LabelSelectorAsSelector(policy.Spec.ConnectionNsSelector); err != nil {
			return err
		}
	}
	return nil
}
