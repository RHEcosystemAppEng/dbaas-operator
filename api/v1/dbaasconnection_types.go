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

// DBaaSConnectionSpec defines the desired state of DBaaSConnection
type DBaaSConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Type will be used by ServiceBindingConverter to indicate the type of connection desired (WIP)
	Type string `json:"type"`

	// Provider will be used by ServiceBindingConverter to indicate the provider type for connection desired (WIP)
	Provider string `json:"provider"`

	// Foo is an example field of DBaaSConnection. Edit dbaasconnection_types.go to remove/update
	Cluster *DBaaSCluster `json:"cluster"`
}

// DBaaSConnectionStatus defines the observed state of DBaaSConnection
type DBaaSConnectionStatus struct {

	// DBConfigMap is the name of the ConfigMap containing the connection info
	DBConfigMap string `json:"dbConfigMap"`

	//+kubebuilder:validation:Required
	// DBCredentials is the name of the Secret containing the database credentials
	DBCredentials string `json:"dbCredentials"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DBaaSConnection is the Schema for the dbaasconnections API
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
