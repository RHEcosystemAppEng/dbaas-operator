/*
Copyright 2023 The OpenShift Database Access Authors.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Constants for DBaaS condition types, reasons, messages and type labels.
const (
	// DBaaS condition types:
	DBaaSInventoryReadyType         string = "InventoryReady"
	DBaaSInventoryProviderSyncType  string = "SpecSynced"
	DBaaSConnectionReadyType        string = "ConnectionReady"
	DBaaSConnectionProviderSyncType string = "ReadyForBinding"
	DBaaSInstanceReadyType          string = "InstanceReady"
	DBaaSInstanceProviderSyncType   string = "ProvisionReady"
	DBaaSPolicyReadyType            string = "PolicyReady"
	DBaaSPlatformReadyType          string = "PlatformReady"

	// DBaaS condition reasons:
	Ready                          string = "Ready"
	DBaaSPolicyNotFound            string = "DBaaSPolicyNotFound"
	DBaaSPolicyNotReady            string = "DBaaSPolicyNotReady"
	DBaaSProviderNotFound          string = "DBaaSProviderNotFound"
	DBaaSInventoryNotFound         string = "DBaaSInventoryNotFound"
	DBaaSInventoryNotReady         string = "DBaaSInventoryNotReady"
	DBaaSInventoryNotProvisionable string = "DBaaSInventoryNotProvisionable"
	DBaaSInvalidNamespace          string = "InvalidNamespace"
	DBaaSServiceNotAvailable       string = "DBaaSServiceNotAvailable"
	ProviderReconcileInprogress    string = "ProviderReconcileInprogress"
	ProviderReconcileError         string = "ProviderReconcileError"
	ProviderParsingError           string = "ProviderParsingError"
	InstallationInprogress         string = "InstallationInprogress"
	InstallationCleanup            string = "InstallationCleanup"

	// DBaaS condition messages
	MsgProviderCRStatusSyncDone      string = "Provider Custom Resource status sync completed"
	MsgProviderCRReconcileInProgress string = "DBaaS Provider Custom Resource reconciliation in progress"
	MsgInventoryNotReady             string = "Inventory discovery not done"
	MsgInventoryNotProvisionable     string = "Inventory provisioning not allowed"
	MsgPolicyNotFound                string = "Failed to find an active Policy"
	MsgPolicyReady                   string = "Policy is active"
	MsgInvalidNamespace              string = "Invalid connection namespace for the referenced inventory"
	MsgPolicyNotReady                string = "Another active Policy already exists"

	TypeLabelValue    = "credentials"
	TypeLabelKey      = "db-operator/type"
	TypeLabelKeyMongo = "atlas.mongodb.com/type"

	ProvisioningPlanFreeTrial  string = "FREETRIAL"
	ProvisioningPlanServerless string = "SERVERLESS"
	ProvisioningPlanDedicated  string = "DEDICATED"
)

// ProvisioningParameterType defines teh type for provisioning parameters
type ProvisioningParameterType string

// DBaaS provisioning fields
const (
	ProvisioningName                    ProvisioningParameterType = "name"
	ProvisioningPlan                    ProvisioningParameterType = "plan"
	ProvisioningCloudProvider           ProvisioningParameterType = "cloudProvider"
	ProvisioningRegions                 ProvisioningParameterType = "regions"
	ProvisioningAvailabilityZones       ProvisioningParameterType = "availabilityZones"
	ProvisioningNodes                   ProvisioningParameterType = "nodes"
	ProvisioningMachineType             ProvisioningParameterType = "machineType"
	ProvisioningStorageGib              ProvisioningParameterType = "storageGib"
	ProvisioningSpendLimit              ProvisioningParameterType = "spendLimit"
	ProvisioningTeamProject             ProvisioningParameterType = "teamProject"
	ProvisioningDatabaseType            ProvisioningParameterType = "databaseType"
	ProvisioningDedicatedLocationLabel  ProvisioningParameterType = "dedicatedLocationLabel"
	ProvisioningServerlessLocationLabel ProvisioningParameterType = "serverlessLocationLabel"
	ProvisioningHardwareLabel           ProvisioningParameterType = "hardwareLabel"
	ProvisioningPlanLabel               ProvisioningParameterType = "planLabel"
	ProvisioningSpendLimitLabel         ProvisioningParameterType = "spendLimitLabel"
)

// DBaasInstancePhase defines the phases for instance provisioning.
type DBaasInstancePhase string

// DatabaseServiceType defines the supported database service types.
type DatabaseServiceType string

// Constants for the instance phases.
const (
	InstancePhaseUnknown  DBaasInstancePhase = "Unknown"
	InstancePhasePending  DBaasInstancePhase = "Pending"
	InstancePhaseCreating DBaasInstancePhase = "Creating"
	InstancePhaseUpdating DBaasInstancePhase = "Updating"
	InstancePhaseDeleting DBaasInstancePhase = "Deleting"
	InstancePhaseDeleted  DBaasInstancePhase = "Deleted"
	InstancePhaseReady    DBaasInstancePhase = "Ready"
	InstancePhaseError    DBaasInstancePhase = "Error"
	InstancePhaseFailed   DBaasInstancePhase = "Failed"
)

// DBaaSProviderSpec defines the desired state of a DBaaSProvider object.
type DBaaSProviderSpec struct {
	// Contains information about database provider and platform.
	Provider DatabaseProviderInfo `json:"provider"`

	// +kubebuilder:default=dbaas.redhat.com/v1alpha1
	// The DBaaS API group version supported by the provider.
	GroupVersion string `json:"groupVersion"`

	// The name of the inventory custom resource definition (CRD) as defined by the database provider.
	InventoryKind string `json:"inventoryKind"`

	// The name of the connection's custom resource definition (CRD) as defined by the provider.
	ConnectionKind string `json:"connectionKind"`

	// The name of the instance's custom resource definition (CRD) as defined by the provider for provisioning.
	InstanceKind string `json:"instanceKind"`

	// Indicates what information to collect from the user interface and how to display fields in a form.
	CredentialFields []CredentialField `json:"credentialFields"`

	// Indicates whether the provider offers free trials.
	AllowsFreeTrial bool `json:"allowsFreeTrial"`

	// The URL for provisioning instances by using the database provider's web portal.
	ExternalProvisionURL string `json:"externalProvisionURL"`

	// Instructions on how to provision instances by using the database provider's web portal.
	ExternalProvisionDescription string `json:"externalProvisionDescription"`

	// Parameter specifications used by the user interface (UI) for provisioning a database instance.
	ProvisioningParameters map[ProvisioningParameterType]ProvisioningParameter `json:"provisioningParameters,omitempty"`
}

// DBaaSProviderStatus defines the observed state of DBaaSProvider object.
type DBaaSProviderStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// DatabaseProviderInfo defines the information for a DBaaSProvider object.
type DatabaseProviderInfo struct {
	// The name used to specify the service binding origin parameter.
	// For example, 'OpenShift Database Access / Crunchy Bridge'.
	Name string `json:"name"`

	// A user-friendly name for this database provider.
	// For example, 'Crunchy Bridge managed PostgreSQL'.
	DisplayName string `json:"displayName"`

	// Indicates the description text shown for a database provider within the user interface.
	// For example, the catalog tile description.
	DisplayDescription string `json:"displayDescription"`

	// Indicates what icon to display on the catalog tile.
	Icon ProviderIcon `json:"icon"`
}

// ProviderIcon follows the same field and naming formats as a comma-separated values (CSV) file.
type ProviderIcon struct {
	Data      string `json:"base64data"`
	MediaType string `json:"mediatype"`
}

// CredentialField defines the CredentialField object attributes.
type CredentialField struct {
	// The name for this field.
	Key string `json:"key"`

	// A user-friendly name for this field.
	DisplayName string `json:"displayName"`

	// The type of field: string, maskedstring, integer, or boolean.
	Type string `json:"type"`

	// Defines if the field is required or not.
	Required bool `json:"required"`

	// Additional information about the field.
	HelpText string `json:"helpText,omitempty"`
}

// DBaaSInventorySpec defines the inventory specifications for the provider's operators.
type DBaaSInventorySpec struct {
	// The secret containing the provider-specific connection credentials to use with the provider's API endpoint.
	// The format specifies the secret in the provider’s operator for its DBaaSProvider custom resource (CR), such as the CredentialFields key.
	// The secret must exist within the same namespace as the inventory.
	CredentialsRef *LocalObjectReference `json:"credentialsRef"`
}

// LocalObjectReference contains enough information to locate the referenced object inside the same namespace.
type LocalObjectReference struct {
	// Name of the referent.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
}

// DBaaSInventoryStatus defines the inventory status that the provider's operator uses.
type DBaaSInventoryStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// A list of database services returned from querying the database provider.
	DatabaseServices []DatabaseService `json:"databaseServices,omitempty"`
}

// DatabaseService defines the information of a database service.
type DatabaseService struct {
	// A provider-specific identifier for the database service.
	// It can contain one or more pieces of information used by the provider's operator to identify the database service.
	ServiceID string `json:"serviceID"`

	// The name of the database service.
	ServiceName string `json:"serviceName,omitempty"`

	// The type of the database service.
	ServiceType *DatabaseServiceType `json:"serviceType,omitempty"`

	// Any other provider-specific information related to this service.
	ServiceInfo map[string]string `json:"serviceInfo,omitempty"`
}

// Instance defines the information of a database instance.
type Instance struct {
	// A provider-specific identifier for this instance in the database service.
	// It can contain one or more pieces of information used by the provider's operator to identify the instance on the database service.
	InstanceID string `json:"instanceID"`

	// The name of this instance in the database service.
	Name string `json:"name,omitempty"`

	// Any other provider-specific information related to this instance.
	InstanceInfo map[string]string `json:"instanceInfo,omitempty"`
}

// NamespacedName defines the namespace and name of a k8s resource.
type NamespacedName struct {
	// The namespace where an object of a known type is stored.
	Namespace string `json:"namespace,omitempty"`

	// The name for object of a known type.
	Name string `json:"name"`
}

// DBaaSConnectionSpec defines the desired state of a DBaaSConnection object.
type DBaaSConnectionSpec struct {
	// A reference to the relevant DBaaSInventory custom resource (CR).
	InventoryRef NamespacedName `json:"inventoryRef"`

	// The ID of the database service to connect to, as seen in the status of the referenced DBaaSInventory.
	DatabaseServiceID string `json:"databaseServiceID,omitempty"`

	// A reference to the database service CR used, if the DatabaseServiceID is not specified.
	DatabaseServiceRef *NamespacedName `json:"databaseServiceRef,omitempty"`

	// The type of the database service to connect to, as seen in the status of the referenced DBaaSInventory.
	DatabaseServiceType *DatabaseServiceType `json:"databaseServiceType,omitempty"`
}

// DBaaSConnectionStatus defines the observed state of a DBaaSConnection object.
type DBaaSConnectionStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// The secret holding account credentials for accessing the database instance.
	CredentialsRef *corev1.LocalObjectReference `json:"credentialsRef,omitempty"`

	// A ConfigMap object holding non-sensitive information for connecting to the database instance.
	ConnectionInfoRef *corev1.LocalObjectReference `json:"connectionInfoRef,omitempty"`
}

// DBaaSProviderConnection defines the schema for a provider's connection status.
type DBaaSProviderConnection struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSConnectionSpec   `json:"spec,omitempty"`
	Status DBaaSConnectionStatus `json:"status,omitempty"`
}

// DBaaSProviderInventory defines the schema for a provider's inventory status.
type DBaaSProviderInventory struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSInventorySpec   `json:"spec,omitempty"`
	Status DBaaSInventoryStatus `json:"status,omitempty"`
}

// DBaaSInstanceSpec defines the desired state of a DBaaSInstance object.
type DBaaSInstanceSpec struct {
	// A reference to the relevant DBaaSInventory custom resource (CR).
	InventoryRef NamespacedName `json:"inventoryRef"`

	// Parameters with values used for provisioning.
	ProvisioningParameters map[ProvisioningParameterType]string `json:"provisioningParameters,omitempty"`
}

// DBaaSInstanceStatus defines the observed state of a DBaaSInstance.
type DBaaSInstanceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// A provider-specific identifier for this instance in the database service.
	// It can contain one or more pieces of information used by the provider's operator to identify the instance on the database service.
	InstanceID string `json:"instanceID"`

	// Any other provider-specific information related to this instance.
	InstanceInfo map[string]string `json:"instanceInfo,omitempty"`

	// +kubebuilder:validation:Enum=Unknown;Pending;Creating;Updating;Deleting;Deleted;Ready;Error;Failed
	// +kubebuilder:default=Unknown
	// Represents the following cluster provisioning phases.
	// Unknown: An unknown cluster provisioning status.
	// Pending: In the queue, waiting for provisioning to start.
	// Creating: Provisioning is in progress.
	// Updating: Updating the cluster is in progress.
	// Deleting: Cluster deletion is in progress.
	// Deleted: Cluster has been deleted.
	// Ready: Cluster provisioning is done.
	// Error: Cluster provisioning error.
	// Failed: Cluster provisioning failed.
	Phase DBaasInstancePhase `json:"phase"`
}

// DBaaSProviderInstance defines the schema for a provider instance object.
type DBaaSProviderInstance struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBaaSInstanceSpec   `json:"spec,omitempty"`
	Status DBaaSInstanceStatus `json:"status,omitempty"`
}

// Option defines the value and display value for an option in a dropdown menu, radio button, or checkbox.
type Option struct {
	// Value of the option.
	Value string `json:"value,omitempty"`

	// Corresponding display value.
	DisplayValue string `json:"displayValue,omitempty"`
}

// FieldDependency defines the name and value of a dependency field.
type FieldDependency struct {
	// +kubebuilder:validation:Enum=name;plan;cloudProvider;regions;availabilityZones;nodes;machineType;storageGib;spendLimit;teamProject;databaseType;dedicatedLocationLabel;serverlessLocationLabel;hardwareLabel;planLabel;spendLimitLabel
	// Name of the dependency field.
	Field ProvisioningParameterType `json:"field,omitempty"`

	// Value of the dependency field.
	Value string `json:"value,omitempty"`
}

// ProvisioningParameter provides information for a ProvisioningParameter object.
type ProvisioningParameter struct {
	// A user-friendly name for this field.
	DisplayName string `json:"displayName"`

	// Additional information about the field.
	HelpText string `json:"helpText,omitempty"`

	// Lists of additional data containing the options or default values for the field.
	ConditionalData []ConditionalProvisioningParameterData `json:"conditionalData,omitempty"`
}

// ConditionalProvisioningParameterData provides a list of available options with default values for a dropdown menu, or a list of default values entered by the user within the user interface (UI) based on the dependencies.
// A provisioning parameter can have many options lists and default values, depending on the dependency parameters.
// If options lists are present, the field displays a dropdown menu in the UI, otherwise it displays an empty field for user input.
// For example, you can have four different options lists for different regions: one for dedicated clusters on Google Cloud Platform (GCP), one for dedicated clusters on Amazon Web Services (AWS), one for serverless on GCP, and one for serverless on AWS.
type ConditionalProvisioningParameterData struct {
	// List of the dependent fields and values.
	Dependencies []FieldDependency `json:"dependencies,omitempty"`

	// Options displayed in the UI.
	Options []Option `json:"options,omitempty"`

	// Set a default value.
	DefaultValue string `json:"defaultValue,omitempty"`
}
