/*
Copyright 2022, Red Hat.

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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var dbaasinstancelog = logf.Log.WithName("DBaaSInstance-resource")
var apiClientInst client.Client = nil

func (r *DBaaSInstance) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if apiClientInst == nil {
		apiClientInst = mgr.GetClient()
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-dbaas-redhat-com-v1alpha1-dbaasinstance,mutating=false,failurePolicy=fail,sideEffects=None,groups=dbaas.redhat.com,resources=dbaasinstances,verbs=create;update,versions=v1alpha1,name=vdbaasinstance.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &DBaaSInstance{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (inst *DBaaSInstance) ValidateCreate() error {
	dbaasinstancelog.Info("validate create", "name", inst.Name)
	return inst.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (inst *DBaaSInstance) ValidateUpdate(old runtime.Object) error {
	dbaasinstancelog.Info("validate update", "name", inst.Name)
	return inst.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DBaaSInstance) ValidateDelete() error {
	dbaasinstancelog.Info("validate delete", "name", r.Name)
	return nil
}

func (inst *DBaaSInstance) validate() error {
	// Retrieve the secret object
	inventory := &DBaaSInventory{}
	ns := inst.Spec.InventoryRef.Namespace
	if len(ns) == 0 {
		ns = inst.Namespace
	}
	if err := apiClientInst.Get(context.TODO(), types.NamespacedName{Name: inst.Spec.InventoryRef.Name, Namespace: ns}, inventory); err != nil {
		return err
	}
	// Retrieve the provider object
	provider := &DBaaSProvider{}
	if err := apiClientInst.Get(context.TODO(), types.NamespacedName{Name: inventory.Spec.ProviderRef.Name, Namespace: ""}, provider); err != nil {
		return err
	}
	return inst.validateFields(provider)
}

func (inst *DBaaSInstance) validateFields(provider *DBaaSProvider) error {
	dbaasinstancelog.Info("instance webhook: validateFields", "name", inst.Name)
	for _, param := range provider.Spec.InstanceParameterSpecs {
		dbaasinstancelog.Info("param", "data", param)
		if param.Required {
			if len(param.InstanceFieldName) > 0 {
				switch param.InstanceFieldName {
				case "name":
					if len(inst.Spec.Name) == 0 {
						return field.Required(field.NewPath("spec").Child(param.InstanceFieldName), "")
					}
					continue
				case "providerName":
					if len(inst.Spec.CloudProvider) == 0 {
						return field.Required(field.NewPath("spec").Child(param.InstanceFieldName), "")
					}
					continue
				case "regionName":
					if len(inst.Spec.CloudRegion) == 0 {
						return field.Required(field.NewPath("spec").Child(param.InstanceFieldName), "")
					}
					continue
				}
			}
			if value, ok := inst.Spec.OtherInstanceParams[param.Name]; !ok || len(value) == 0 {
				//Required key is missing
				return field.Required(field.NewPath("spec").Child("otherInstanceParams").Child(param.Name), "")
			}
		}
	}
	return nil
}
