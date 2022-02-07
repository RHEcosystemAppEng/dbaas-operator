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

type PlatformsName string

type PlatformsInstlnStatus string

const (
	CrunchyBridgeInstallation          PlatformsName = "crunchy-bridge"
	MongoDBAtlasInstallation           PlatformsName = "mongodb-atlas"
	DBaaSDynamicPluginInstallation     PlatformsName = "dbaas-dynamic-plugin"
	Csv                                PlatformsName = "Csv"
	ConsoleTelemetryPluginInstallation PlatformsName = "console-telemetry-plugin"
	ServiceBindingInstallation         PlatformsName = "service-binding"
	CockroachDBInstallation            PlatformsName = "cockroachdb-cloud"
	DBaaSQuickStartInstallation        PlatformsName = "dbaas-quick-starts"
)

const (
	ResultSuccess    PlatformsInstlnStatus = "success"
	ResultFailed     PlatformsInstlnStatus = "failed"
	ResultInProgress PlatformsInstlnStatus = "in progress"
)

// DBaaSPlatformSpec defines the desired state of DBaaSPlatform
type DBaaSPlatformSpec struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1440
	// The SyncPeriod set The minimum interval at which the provider operator controllers reconcile, the default value is 180 minutes.
	SyncPeriod *int `json:"syncPeriod,omitempty"`
}

// DBaaSPlatformStatus defines the observed state of DBaaSPlatform
type DBaaSPlatformStatus struct {
	PlatformName   PlatformsName         `json:"platformName"`
	PlatformStatus PlatformsInstlnStatus `json:"platformStatus"`
	LastMessage    string                `json:"lastMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

//+operator-sdk:csv:customresourcedefinitions:displayName="DBaaSPlatform"
// DBaaSPlatform is the Schema for the dbaasplatforms API
type DBaaSPlatform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSPlatformSpec   `json:"spec,omitempty"`
	Status DBaaSPlatformStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBaaSPlatformList contains a list of DBaaSPlatform
type DBaaSPlatformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSPlatform `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSPlatform{}, &DBaaSPlatformList{})
}
