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

// DBaaSServiceSpec defines the desired state of DBaaSService
type DBaaSServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Provider is the name of the database provider whom we wish to connect with
	Provider DatabaseProvider `json:"provider"`

	// CredentialsSecretName indicates the name of the secret storing the vendor-specific connection credentials
	CredentialsSecretName string `json:"credentialsSecretName"`

	// CredentialsSecretName indicates the namespace of the secret storing the vendor-specific connection credentials
	CredentialsSecretNamespace string `json:"credentialsSecretNamespace"`
}

// DBaaSServiceStatus defines the observed state of DBaaSService
type DBaaSServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Projects reflects the list of entities returned from querying the DB provider
	Projects []DBaaSProject `json:"projects,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DBaaSService is the Schema for the dbaasservices API
type DBaaSService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSServiceSpec   `json:"spec,omitempty"`
	Status DBaaSServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBaaSServiceList contains a list of DBaaSService
type DBaaSServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSService `json:"items"`
}

type DatabaseProvider struct {
	Name string `json:"name"`
}

type DBaaSProject struct {
	ID       string              `json:"id,omitempty"`
	Name     string              `json:"name,omitempty"`
	Clusters []DBaaSCluster      `json:"clusters,omitempty"`
	Users    []DBaaSDatabaseUser `json:"users,omitempty"`
}

type DBaaSProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSProject `json:"items"`
}

type DBaaSCluster struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	CloudProviderName string `json:"cloudProvider,omitempty"`
	CloudRegion       string `json:"cloudRegion,omitempty"`
	InstanceSizeName  string `json:"instanceSizeName,omitempty"`
}

type DBaaSClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSClusterList `json:"items"`
}

type DBaaSDatabaseUser struct {
	Name string `json:"name,omitempty"`
}

type DBaaSDatabaseUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSDatabaseUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSService{}, &DBaaSServiceList{})
}
