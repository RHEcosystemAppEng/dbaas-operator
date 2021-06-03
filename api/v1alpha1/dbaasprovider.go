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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DBaaSProvider defines a database provider for DBaaS operator
type DBaaSProvider struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Provider is the name of the database provider
	Provider DatabaseProvider `json:"provider"`

	// The name of the inventory CRD defined by the provider
	InventoryKind string `json:"inventoryKind"`

	// The authentication fields the user needs to receive and provide
	AuthenticationFields []AuthenticationField `json:"authenticationFields"`

	// The name of the connection CRD defined by the provider
	ConnectionKind string `json:"connectionKind"`
}

type DatabaseProvider struct {
	Name string `json:"name"`
}

type AuthenticationField struct {
	Name   string `json:"name"`
	Masked bool   `json:"masked"`
}

// DBaaSProviderList contains a list of DBaaSProvider
type DBaaSProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSProvider `json:"items"`
}

// DBaaSInventorySpec defines the desired state of DBaaSInventory
type DBaaSInventorySpec struct {
	// Important: Run "make" to regenerate code after modifying this file

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

// DBaaSConnectionSpec defines the desired state of DBaaSConnection
type DBaaSConnectionSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// A reference to the relevant DBaaSInventory CR
	InventoryRef *corev1.LocalObjectReference `json:"inventory"`

	// The ID of the instance to connect to, as seen in the Status of
	// the referenced DBaaSInventory
	InstanceID string `json:"instanceID"`
}

// DBaaSConnectionStatus defines the observed state of DBaaSConnection
type DBaaSConnectionStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// The connection string for this instance
	ConnectionString string `json:"connectionString,omitempty"`

	// Secret holding username and password
	CredentialsRef *corev1.LocalObjectReference `json:"credentialsRef"`

	// Any other provider-specific information related to this connection
	ConnectionInfo map[string]string `json:"connectionInfo,omitempty"`
}
