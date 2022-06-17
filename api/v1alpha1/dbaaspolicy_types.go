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

// DBaaSPolicySpec enables admin capabilities within a namespace and sets default inventory policy.
// Policy defaults can be overridden on a per-inventory basis.
type DBaaSPolicySpec struct {
	DBaaSInventoryPolicy `json:",inline"`
}

// DBaaSPolicyStatus defines the observed state of DBaaSPolicy
type DBaaSPolicyStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Active",type=string,JSONPath=`.status.conditions[0].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

//+operator-sdk:csv:customresourcedefinitions:displayName="Provider Account Policy"
// DBaaSPolicy enables admin capabilities within a namespace and sets default inventory policy.
// Policy defaults can be overridden on a per-inventory basis.
type DBaaSPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSPolicySpec   `json:"spec,omitempty"`
	Status DBaaSPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBaaSPolicyList contains a list of DBaaSPolicy
type DBaaSPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSPolicy{}, &DBaaSPolicyList{})
}
