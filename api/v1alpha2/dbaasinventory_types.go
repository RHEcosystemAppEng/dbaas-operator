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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DBaaSInventoryPolicy sets inventory policy
type DBaaSInventoryPolicy struct {
	// Disable provisioning against inventory accounts
	DisableProvisions *bool `json:"disableProvisions,omitempty"`

	// Namespaces where DBaaSConnections/DBaaSInstances are allowed to reference a policy's inventories.
	// Each inventory can individually override this. Use "*" to allow all namespaces.
	// If not set in either the policy or inventory object, connections will only be allowed in the inventory's namespace.
	ConnectionNamespaces *[]string `json:"connectionNamespaces,omitempty"`

	// Use a label selector to determine namespaces where DBaaSConnections/DBaaSInstances are allowed to reference a policy's inventories.
	// Each inventory can individually override this. A label selector is a label query over a set of resources. The result of matchLabels and
	// matchExpressions are ANDed. An empty label selector matches all objects. A null
	// label selector matches no objects.
	ConnectionNsSelector *metav1.LabelSelector `json:"connectionNsSelector,omitempty"`
}

// DBaaSInventorySpec defines the Inventory Spec to be used by provider operators
type DBaaSInventorySpec struct {
	// The Secret containing the provider-specific connection credentials to use with its API
	// endpoint. The format of the Secret is specified in the provider’s operator in its
	// DBaaSProvider CR (CredentialFields key). The Secret must exist within the same namespace
	// as the Inventory.
	CredentialsRef *LocalObjectReference `json:"credentialsRef"`
}

// DBaaSOperatorInventorySpec defines the desired state of DBaaSInventory
type DBaaSOperatorInventorySpec struct {
	// A reference to a DBaaSProvider CR
	ProviderRef NamespacedName `json:"providerRef"`

	// The properties that will be copied into the provider’s inventory Spec
	DBaaSInventorySpec `json:",inline"`

	// The policy for this inventory
	DBaaSInventoryPolicy `json:",inline"`
}

// DBaaSInventoryStatus defines the observed state of DBaaSInventory
type DBaaSInventoryStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// A list of database services returned from querying the DB provider
	DatabaseServices []DatabaseService `json:"databaseServices,omitempty"`
}

// DatabaseService defines the information of a database service
type DatabaseService struct {
	// A provider-specific identifier for the database service. It may contain one or
	// more pieces of information used by the provider operator to identify the database service.
	ServiceID string `json:"serviceID"`

	// The name of the database service
	ServiceName string `json:"serviceName,omitempty"`

	// The type of the database service
	// +kubebuilder:validation:Enum=instance;cluster
	// +kubebuilder:default=instance
	ServiceType DatabaseServiceType `json:"serviceType,omitempty"`

	// Any other provider-specific information related to this service
	ServiceInfo map[string]string `json:"serviceInfo,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// DBaaSInventory is the Schema for the dbaasinventories API
// Inventory objects must be created in a valid namespace, determined by the existence of a DBaaSPolicy object
//+operator-sdk:csv:customresourcedefinitions:displayName="Provider Account"
type DBaaSInventory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSOperatorInventorySpec `json:"spec,omitempty"`
	Status DBaaSInventoryStatus       `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBaaSInventoryList contains a list of DBaaSInventory
type DBaaSInventoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSInventory `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSInventory{}, &DBaaSInventoryList{})
}

// DBaaSProviderInventory is the schema for unmarshalling provider inventory object
type DBaaSProviderInventory struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSInventorySpec   `json:"spec,omitempty"`
	Status DBaaSInventoryStatus `json:"status,omitempty"`
}
