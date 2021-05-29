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

// DBaaSInventorySpec defines the desired state of DBaaSInventory
type DBaaSInventorySpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Provider is the name of the database provider that we wish to connect with
	Provider DatabaseProvider `json:"provider"`

	// The secret storing the vendor-specific connection credentials to
	// use with the API endpoint. The secret may be placed in a separate
	// namespace to control access.
	CredentialsRef *NamespacedName `json:"credentialsRef"`
}

// DBaaSInventoryStatus defines the observed state of DBaaSInventory
type DBaaSInventoryStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// E.g., MongoDB, Postgres
	Type string `json:"type"`

	// A list of instances returned from querying the DB provider
	Instances []Instance `json:"instances,omitempty"`
}

type Instance struct {
	// The ID for this instance in the database service
	InstanceID string `json:"instanceID"`

	// The name of this instance in the database service
	Name string `json:"name,omitempty"`

	// Any other provider-specific information related to this instance
	InstanceInfo map[string]string `json:"extraInfo,omitempty"`
}

type NamespacedName struct {
	// The namespace where object of known type is store
	Namespace string `json:"namespace,omitempty"`

	// The name for object of known type
	Name string `json:"name,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DBaaSInventory is the Schema for the dbaasinventory API
type DBaaSInventory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSInventorySpec   `json:"spec,omitempty"`
	Status DBaaSInventoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBaaSInventoryList contains a list of DBaaSInventories
type DBaaSInventoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSInventory `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSInventory{}, &DBaaSInventoryList{})
}
