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
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type PlatformsName string
type PlatformsInstlnStatus string
type PlatformsType int

const (
	CrunchyBridgeInstallation          PlatformsName = "crunchy-bridge"
	MongoDBAtlasInstallation           PlatformsName = "mongodb-atlas"
	DBaaSDynamicPluginInstallation     PlatformsName = "dbaas-dynamic-plugin"
	ConsoleTelemetryPluginInstallation PlatformsName = "console-telemetry-plugin"
	CockroachDBInstallation            PlatformsName = "cockroachdb-cloud"
	DBaaSQuickStartInstallation        PlatformsName = "dbaas-quick-starts"
)

const (
	TypeQuickStart PlatformsType = iota
	TypeConsolePlugin
	TypeProvider
)

const (
	ResultSuccess    PlatformsInstlnStatus = "success"
	ResultFailed     PlatformsInstlnStatus = "failed"
	ResultInProgress PlatformsInstlnStatus = "in progress"
)

type PlatformConfig struct {
	Name           string
	CSV            string
	DeploymentName string
	Image          string
	PackageName    string
	Channel        string
	DisplayName    string
	Envs           []v1.EnvVar
	Type           PlatformsType
}

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

type DbaaSPlatformInterface interface {
	List(opts metav1.ListOptions) (*DBaaSPlatformList, error)
	Get(name string, options metav1.GetOptions) (*DBaaSPlatform, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type dbaasPlatformClient struct {
	restClient rest.Interface
	ns         string
	ctx        context.Context
}

func (c *dbaasPlatformClient) List(opts metav1.ListOptions) (*DBaaSPlatformList, error) {
	result := DBaaSPlatformList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("dbaasplatforms").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *dbaasPlatformClient) Get(name string, opts metav1.GetOptions) (*DBaaSPlatform, error) {
	result := DBaaSPlatform{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("dbaasplatforms").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *dbaasPlatformClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("dbaasplatforms").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(c.ctx)
}
