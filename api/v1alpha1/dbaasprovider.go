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

const (
	// DBaaS condition types
	DBaaSInventoryReadyType         string = "InventoryReady"
	DBaaSInventoryProviderSyncType  string = "SpecSynced"
	DBaaSConnectionReadyType        string = "ConnectionReady"
	DBaaSConnectionProviderSyncType string = "ReadyForBinding"
	DBaaSInstanceReadyType          string = "InstanceReady"
	DBaaSInstanceProviderSyncType   string = "ProvisionReady"

	// DBaaS condition reasons
	Ready                       string = "Ready"
	DBaaSTenantNotFound         string = "DBaaSTenantNotFound"
	DBaaSProviderNotFound       string = "DBaaSProviderNotFound"
	DBaaSInventoryNotFound      string = "DBaaSInventoryNotFound"
	DBaaSInventoryNotReady      string = "DBaaSInventoryNotReady"
	DBaaSInvalidNamespace       string = "InvalidNamespace"
	ProviderReconcileInprogress string = "ProviderReconcileInprogress"
	ProviderParsingError        string = "ProviderParsingError"

	// DBaaS condition messages
	MsgProviderCRStatusSyncDone      string = "Provider Custom Resource status sync completed"
	MsgProviderCRReconcileInProgress string = "DBaaS Provider Custom Resource reconciliation in progress"
	MsgInventoryNotReady             string = "Inventory discovery not done"
	MsgTenantNotFound                string = "Failed to find DBaaS tenants"
	MsgInvalidNamespace              string = "Invalid connection namespace for the referenced inventory"
)

// DBaaSProviderSpec defines the desired state of DBaaSProvider
type DBaaSProviderSpec struct {
	// Provider contains information about database provider & platform
	Provider DatabaseProvider `json:"provider"`

	// InventoryKind is the name of the inventory resource (CRD) defined by the provider
	InventoryKind string `json:"inventoryKind"`

	// ConnectionKind is the name of the connection resource (CRD) defined by the provider
	ConnectionKind string `json:"connectionKind"`

	// InstanceKind is the name of the instance resource (CRD) defined by the provider for provisioning
	InstanceKind string `json:"instanceKind"`

	// CredentialFields indicates what information to collect from UX & how to display fields in a form
	CredentialFields []CredentialField `json:"credentialFields"`

	// AllowsFreeTrial indicates whether the provider provides free trials
	AllowsFreeTrial bool `json:"allowsFreeTrial"`

	// ExternalProvisionURL URL for provisioning instances through database provider web portal
	ExternalProvisionURL string `json:"externalProvisionURL"`

	// ExternalProvisionDescription instructions on how to provision instances using provider web portal
	ExternalProvisionDescription string `json:"externalProvisionDescription"`

	// InstanceParameterSpecs  indicates what parameters to collect from UX & how to display fields in a form in order to provision an instance
	InstanceParameterSpecs []InstanceParameterSpec `json:"instanceParameterSpecs"`
}

type DatabaseProvider struct {
	// Indicates the name used to specify Service Binding origin parameter (e.g. 'Red Hat DBaas / MongoDB Atlas')
	Name string `json:"name"`

	// A user-friendly name for this database provider (e.g. 'MongoDB Atlas')
	DisplayName string `json:"displayName"`

	// DisplayDescription indicates the description text shown for a Provider within UX (e.g. developer’s catalog tile)
	DisplayDescription string `json:"displayDescription"`

	// Icon information indicates what logo we display on developer catalog tile
	Icon ProviderIcon `json:"icon"`
}

// ProviderIcon follows same field/naming formats as CSV
type ProviderIcon struct {
	Data      string `json:"base64data"`
	MediaType string `json:"mediatype"`
}

type CredentialField struct {
	// The name for this field
	Key string `json:"key"`

	// A user-friendly name for this field
	DisplayName string `json:"displayName"`

	// The type of field (string, maskedstring, integer, boolean)
	Type string `json:"type"`

	// If this field is required or not
	Required bool `json:"required"`
}

// DBaaSInventorySpec defines the Inventory Spec to be used by provider operators
type DBaaSInventorySpec struct {
	// The Secret containing the provider-specific connection credentials to use with its API
	// endpoint. The format of the Secret is specified in the provider’s operator in its
	// DBaaSProvider CR (CredentialFields key). It is recommended to place the Secret in a
	// namespace with limited accessibility.
	CredentialsRef *NamespacedName `json:"credentialsRef"`
}

// DBaaSInventoryStatus defines the Inventory status to be used by provider operators
type DBaaSInventoryStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// A list of instances returned from querying the DB provider
	Instances []Instance `json:"instances,omitempty"`
}

type Instance struct {
	// A provider-specific identifier for this instance in the database service. It may contain one or
	// more pieces of information used by the provider operator to identify the instance on the
	// database service.
	InstanceID string `json:"instanceID"`

	// The name of this instance in the database service
	Name string `json:"name,omitempty"`

	// Any other provider-specific information related to this instance
	InstanceInfo map[string]string `json:"instanceInfo,omitempty"`
}

type NamespacedName struct {
	// The namespace where object of known type is stored
	Namespace string `json:"namespace,omitempty"`

	// The name for object of known type
	Name string `json:"name"`
}

// DBaaSConnectionSpec defines the desired state of DBaaSConnection
type DBaaSConnectionSpec struct {
	// A reference to the relevant DBaaSInventory CR
	InventoryRef NamespacedName `json:"inventoryRef"`

	// The ID of the instance to connect to, as seen in the Status of
	// the referenced DBaaSInventory
	InstanceID string `json:"instanceID"`
}

// DBaaSConnectionStatus defines the observed state of DBaaSConnection
type DBaaSConnectionStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Secret holding the credentials needed for accessing the DB instance
	CredentialsRef *corev1.LocalObjectReference `json:"credentialsRef,omitempty"`

	// A ConfigMap holding non-sensitive information needed for connecting to the DB instance
	ConnectionInfoRef *corev1.LocalObjectReference `json:"connectionInfoRef,omitempty"`
}

// DBaaSProviderConnection is the schema for unmarshalling provider connection object
type DBaaSProviderConnection struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSConnectionSpec   `json:"spec,omitempty"`
	Status DBaaSConnectionStatus `json:"status,omitempty"`
}

// DBaaSProviderInventory is the schema for unmarshalling provider inventory object
type DBaaSProviderInventory struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSInventorySpec   `json:"spec,omitempty"`
	Status DBaaSInventoryStatus `json:"status,omitempty"`
}

// DBaaSInstanceSpec defines the desired state of DBaaSInstance
type DBaaSInstanceSpec struct {
	// A reference to the relevant DBaaSInventory CR
	InventoryRef NamespacedName `json:"inventoryRef"`

	// The name of this instance in the database service
	Name string `json:"name"`

	// Identifies the desired cloud infrastructure provider
	CloudProvider string `json:"cloudProvider,omitempty"`

	// Identifies the requested deployment region within the cloud provider (e.g. us-east-1)
	CloudRegion string `json:"cloudRegion,omitempty"`

	// Any other provider-specific parameters related to the instance provisioning
	OtherInstanceParams map[string]string `json:"otherInstanceParams,omitempty"`
}

// DBaaSInstanceStatus defines the observed state of DBaaSInstance
type DBaaSInstanceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// The ID of the instance,
	InstanceID string `json:"instanceID"`

	// Any other provider-specific information related to this instance
	InstanceInfo map[string]string `json:"instanceInfo,omitempty"`

	// Represents the cluster provisioning phase
	// Pending - provisioning not yet started
	// Creating - provisioning in progress
	// Updating - cluster updating in progress
	// Deleting - cluster deletion in progress
	// Deleted - cluster has been deleted
	// Ready - cluster provisioning complete
	Phase string `json:"phase"`
}

// DBaaSProviderInstance is the schema for unmarshalling provider instance object
type DBaaSProviderInstance struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSInstanceSpec   `json:"spec,omitempty"`
	Status DBaaSInstanceStatus `json:"status,omitempty"`
}

type InstanceParameterSpec struct {
	// The name for this field
	Name string `json:"name"`

	// A user-friendly name for this parameter
	DisplayName string `json:"displayName"`

	// The type of parameter (string, maskedstring, integer, boolean)
	Type string `json:"type"`

	// If this field is required or not
	Required bool `json:"required"`

	// Default value for this field
	DefaultValue string `json:"defaultValue,omitempty"`
}

var InstanceParameterSpecs = InstanceParameterSpec{}
