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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DBaaSProviderRegistration defines a database provider for DBaaS operator
type DBaaSProviderRegistrationSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Provider is the name of the database provider
	Provider DatabaseProvider `json:"provider"`

	// The name of the inventory CRD defined by the provider
	InventoryKind string `json:"inventoryKind"`

	// The name of the inventory CRD defined by the provider
	AuthenticationFields []string `json:"authenticationFields"`

	// The name of the connection CRD defined by the provider
	ConnectionKind string `json:"connectionKind"`
}

type DatabaseProvider struct {
	Name string `json:"name"`
}

//+kubebuilder:object:root=true

// DBaaSProviderRegistrationList contains a list of DBaaSProviderRegistrations
type DBaaSProviderRegistration struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Spec            DBaaSProviderRegistrationSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// DBaaSProviderRegistrationList contains a list of DBaaSProviderRegistrations
type DBaaSProviderRegistrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSProviderRegistration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSProviderRegistration{}, &DBaaSProviderRegistrationList{})
}
