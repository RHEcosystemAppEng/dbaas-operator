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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DBaaSConnectionSpec defines the desired state of DBaaSConnection
type DBaaSConnectionSpec struct {
	// A reference to the relevant DBaaSInventory CR
	InventoryRef NamespacedName `json:"inventoryRef"`

	// The ID of the database service to connect to, as seen in the Status of
	// the referenced DBaaSInventory
	DatabaseServiceID string `json:"databaseServiceID,omitempty"`

	// A reference to the database service CR that is used if the ID of the
	// service is not specified
	DatabaseServiceRef *NamespacedName `json:"databaseServiceRef,omitempty"`

	// The type of the database service to connect to, as seen in the Status of
	// the referenced DBaaSInventory
	// +kubebuilder:validation:Enum=instance;cluster
	// +kubebuilder:default=instance
	DatabaseServiceType DatabaseServiceType `json:"serviceType,omitempty"`
}

// DBaaSConnectionStatus defines the observed state of DBaaSConnection
type DBaaSConnectionStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Secret holding the credentials needed for accessing the DB instance
	CredentialsRef *corev1.LocalObjectReference `json:"credentialsRef,omitempty"`

	// A ConfigMap holding non-sensitive information needed for connecting to the DB instance
	ConnectionInfoRef *corev1.LocalObjectReference `json:"connectionInfoRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// DBaaSConnection is the Schema for the dbaasconnections API
//+operator-sdk:csv:customresourcedefinitions:displayName="DBaaSConnection"
type DBaaSConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSConnectionSpec   `json:"spec,omitempty"`
	Status DBaaSConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBaaSConnectionList contains a list of DBaaSConnection
type DBaaSConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSConnection{}, &DBaaSConnectionList{})
}

type DBaaSProviderConnection struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSConnectionSpec   `json:"spec,omitempty"`
	Status DBaaSConnectionStatus `json:"status,omitempty"`
}
