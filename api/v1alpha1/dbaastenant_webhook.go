/*
Copyright 2022.

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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var dbaastenantlog = logf.Log.WithName("dbaastenant-resource")

const inventoryNamespaceKey = ".spec.inventoryNamespace"

var tenantWebhookApiClient client.Client

func (r *DBaaSTenant) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if tenantWebhookApiClient == nil {
		tenantWebhookApiClient = mgr.GetClient()
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-dbaas-redhat-com-v1alpha1-dbaastenant,mutating=false,failurePolicy=fail,sideEffects=None,groups=dbaas.redhat.com,resources=dbaastenants,verbs=create;update,versions=v1alpha1,name=vdbaastenant.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &DBaaSTenant{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSTenant) ValidateCreate() error {
	dbaastenantlog.Info("validate create", "name", r.Name)
	return r.validateTenant()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSTenant) ValidateUpdate(old runtime.Object) error {
	dbaastenantlog.Info("validate update", "name", r.Name)
	return r.validateTenant()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSTenant) ValidateDelete() error {
	dbaastenantlog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *DBaaSTenant) validateTenant() error {
	tenantsList := &DBaaSTenantList{}
	if err := tenantWebhookApiClient.List(context.TODO(), tenantsList, client.MatchingFields{inventoryNamespaceKey: r.Spec.InventoryNamespace}); err != nil {
		return err
	}

	if len(tenantsList.Items) > 0 {
		errMsg := fmt.Sprintf("the namespace %s is already managed by tenant %s, it cannot be managed by another tenant", r.Spec.InventoryNamespace, tenantsList.Items[0].Name)
		return field.Invalid(field.NewPath("spec").Child("inventoryNamespace"), r.Spec.InventoryNamespace, errMsg)
	}

	return nil
}
