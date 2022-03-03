/*
Copyright 2021, Red Hat.

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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var dbaasinventorylog = logf.Log.WithName("dbaasinventory-resource")
var inventoryWebhookApiClient client.Client = nil

func (r *DBaaSInventory) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if inventoryWebhookApiClient == nil {
		inventoryWebhookApiClient = mgr.GetClient()
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-dbaas-redhat-com-v1alpha1-dbaasinventory,mutating=false,failurePolicy=fail,sideEffects=None,groups=dbaas.redhat.com,resources=dbaasinventories,verbs=create;update,versions=v1alpha1,name=vdbaasinventory.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &DBaaSInventory{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSInventory) ValidateCreate() error {
	dbaasinventorylog.Info("validate create", "name", r.Name)
	return validateInventory(r)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSInventory) ValidateUpdate(old runtime.Object) error {
	dbaasinventorylog.Info("validate update", "name", r.Name)
	return validateInventory(r)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSInventory) ValidateDelete() error {
	dbaasinventorylog.Info("validate delete", "name", r.Name)
	return nil
}

func validateInventory(inv *DBaaSInventory) error {
	// Retrieve the secret object
	secret := &corev1.Secret{}
	ns := inv.Spec.DBaaSInventorySpec.CredentialsRef.Namespace
	if len(ns) == 0 {
		ns = inv.Namespace
	}
	if err := inventoryWebhookApiClient.Get(context.TODO(), types.NamespacedName{Name: inv.Spec.DBaaSInventorySpec.CredentialsRef.Name, Namespace: ns}, secret); err != nil {
		return err
	}
	// Retrieve the provider object
	provider := &DBaaSProvider{}
	if err := inventoryWebhookApiClient.Get(context.TODO(), types.NamespacedName{Name: inv.Spec.ProviderRef.Name, Namespace: ""}, provider); err != nil {
		return err
	}
	return validateInventoryMandatoryFields(inv, secret, provider)
}

func validateInventoryMandatoryFields(inv *DBaaSInventory, secret *corev1.Secret, provider *DBaaSProvider) error {
	for _, credField := range provider.Spec.CredentialFields {
		if credField.Required {
			if value, ok := secret.Data[credField.Key]; !ok || len(value) == 0 {
				//Required key is missing
				msg := fmt.Sprintf("credentialsRef is invalid: %s is required in secret %s", credField.Key, secret.Name)
				return field.Invalid(field.NewPath("spec").Child("credentialsRef"), *(inv.Spec.CredentialsRef), msg)
			}
		}
	}
	return nil
}
