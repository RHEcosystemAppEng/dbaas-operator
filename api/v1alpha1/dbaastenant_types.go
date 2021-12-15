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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DBaaSTenantSpec defines Tenant inventory namespace and user authorizations
type DBaaSTenantSpec struct {
	// Namespace to watch for DBaaSInventories
	// +kubebuilder:validation:Required
	InventoryNamespace string `json:"inventoryNamespace"`
	// Specify a Tenant’s default Developers for DBaaSInventory “viewer” access
	Authz DBaasUsersGroups `json:"authz,omitempty"`
}

// DBaaSTenantStatus defines the observed state of DBaaSTenant
type DBaaSTenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="Inventory_NS",type=string,JSONPath=`.spec.inventoryNamespace`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

//+operator-sdk:csv:customresourcedefinitions:displayName="DBaaSTenant"
// DBaaSTenant is the Schema for the dbaastenants API
type DBaaSTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSTenantSpec   `json:"spec,omitempty"`
	Status DBaaSTenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBaaSTenantList contains a list of DBaaSTenant
type DBaaSTenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSTenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSTenant{}, &DBaaSTenantList{})
}
