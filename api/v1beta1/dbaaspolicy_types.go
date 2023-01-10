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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The specifications for a _DBaaSPolicy_ object.
// Enables administrative capabilities within a namespace, and sets a default inventory policy.
// Policy defaults can be overridden on a per-inventory basis.
type DBaaSPolicySpec struct {
	DBaaSInventoryPolicy `json:",inline"`
}

// Sets the inventory policy.
type DBaaSInventoryPolicy struct {
	// Disables provisioning on inventory accounts.
	DisableProvisions *bool `json:"disableProvisions,omitempty"`
	// Namespaces where DBaaSConnection and DBaaSInstance objects are only allowed to reference a policy's inventories.
	Connections DBaaSConnectionPolicy `json:"connections,omitempty"`
}

// DBaaSConnectionPolicy sets connection policy
type DBaaSConnectionPolicy struct {
	// Namespaces where DBaaSConnection and DBaaSInstance objects are only allowed to reference a policy's inventories.
	// Using an asterisk surrounded by single quotes ('*'), allows all namespaces.
	// If not set in the policy or by an inventory object, connections are only allowed in the inventory's namespace.
	Namespaces *[]string `json:"namespaces,omitempty"`

	// Use a label selector to determine the namespaces where DBaaSConnection and DBaaSInstance objects are only allowed to reference a policy's inventories.
	// A label selector is a label query over a set of resources.
	// Results use a logical AND from matchExpressions and matchLabels queries.
	// An empty label selector matches all objects.
	// A null label selector matches no objects.
	NsSelector *metav1.LabelSelector `json:"nsSelector,omitempty"`
}

// Defines the observed state of a DBaaSPolicy object.
type DBaaSPolicyStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:storageversion
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Active",type=string,JSONPath=`.status.conditions[0].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Enables administrative capabilities within a namespace, and sets a default inventory policy.
// Policy defaults can be overridden on a per-inventory basis.
//+operator-sdk:csv:customresourcedefinitions:displayName="Provider Account Policy"
type DBaaSPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSPolicySpec   `json:"spec,omitempty"`
	Status DBaaSPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Contains a list of DBaaSPolicies.
type DBaaSPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSPolicy{}, &DBaaSPolicyList{})
}
