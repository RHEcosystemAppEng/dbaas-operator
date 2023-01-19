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

// The name of the platform.
type PlatformsName string

// The status of a platform installation.
type PlatformsInstlnStatus string

// The platform type.
type PlatformsType int

// Supported platform names.
const (
	CrunchyBridgeInstallation      PlatformsName = "crunchy-bridge"
	MongoDBAtlasInstallation       PlatformsName = "mongodb-atlas"
	DBaaSDynamicPluginInstallation PlatformsName = "dbaas-dynamic-plugin"
	CockroachDBInstallation        PlatformsName = "cockroachdb-cloud"
	ObservabilityInstallation      PlatformsName = "observability"
	DBaaSQuickStartInstallation    PlatformsName = "dbaas-quick-starts"
	RDSProviderInstallation        PlatformsName = "rds-provider"
)

// Platform types.
const (
	TypeQuickStart PlatformsType = iota
	TypeConsolePlugin
	TypeOperator
)

// Platform status values.
const (
	ResultSuccess    PlatformsInstlnStatus = "success"
	ResultFailed     PlatformsInstlnStatus = "failed"
	ResultInProgress PlatformsInstlnStatus = "in progress"
)

// Defines parameters for a platform.
type PlatformConfig struct {
	Name           string
	CSV            string
	DeploymentName string
	Image          string
	PackageName    string
	Channel        string
	DisplayName    string
	Envs           []corev1.EnvVar
	Type           PlatformsType
}

// Defines parameters for observatorium.
type ObservabilityConfig struct {
	AuthType        string
	RemoteWritesURL string
	RHSSOTokenURL   string
	AddonName       string
	RHOBSSecretName string
}

// Defines the desired state of a DBaaSPlatform object.
type DBaaSPlatformSpec struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1440
	// The SyncPeriod set The minimum interval at which the provider operator controllers reconcile, the default value is 180 minutes.
	SyncPeriod *int `json:"syncPeriod,omitempty"`
}

// Defines the observed state of a DBaaSPlatform object.
type DBaaSPlatformStatus struct {
	Conditions      []metav1.Condition `json:"conditions,omitempty"`
	PlatformsStatus []PlatformStatus   `json:"platformsStatus"`
}

// Defines the status of a DBaaSPlatform object.
type PlatformStatus struct {
	PlatformName   PlatformsName         `json:"platformName"`
	PlatformStatus PlatformsInstlnStatus `json:"platformStatus"`
	LastMessage    string                `json:"lastMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// The schema for the DBaaSPlatform API.
type DBaaSPlatform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSPlatformSpec   `json:"spec,omitempty"`
	Status DBaaSPlatformStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Contains a list of DBaaSPlatforms.
type DBaaSPlatformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBaaSPlatform `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBaaSPlatform{}, &DBaaSPlatformList{})
}
