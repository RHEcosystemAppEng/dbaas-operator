# API Reference

## Packages
- [dbaas.redhat.com/v1beta1](#dbaasredhatcomv1beta1)


## dbaas.redhat.com/v1beta1

Package v1beta1 contains API Schema definitions for the dbaas v1beta1 API group

### Resource Types
- [DBaaSConnection](#dbaasconnection)
- [DBaaSInstance](#dbaasinstance)
- [DBaaSInventory](#dbaasinventory)
- [DBaaSPlatform](#dbaasplatform)
- [DBaaSPolicy](#dbaaspolicy)
- [DBaaSProvider](#dbaasprovider)



#### ConditionalProvisioningParameterData



ConditionalProvisioningParameterData provides a list of available options with default values for a dropdown menu, or a list of default values entered by the user within the user interface (UI) based on the dependencies. A provisioning parameter can have many options lists and default values, depending on the dependency parameters. If options lists are present, the field displays a dropdown menu in the UI, otherwise it displays an empty field for user input. For example, you can have four different options lists for different regions: one for dedicated clusters on Google Cloud Platform (GCP), one for dedicated clusters on Amazon Web Services (AWS), one for serverless on GCP, and one for serverless on AWS.

_Appears in:_
- [ProvisioningParameter](#provisioningparameter)

| Field | Description |
| --- | --- |
| `dependencies` _[FieldDependency](#fielddependency) array_ | List of the dependent fields and values. |
| `options` _[Option](#option) array_ | Options displayed in the UI. |
| `defaultValue` _string_ | Set a default value. |


#### CredentialField



CredentialField defines the CredentialField object attributes.

_Appears in:_
- [DBaaSProviderSpec](#dbaasproviderspec)

| Field | Description |
| --- | --- |
| `key` _string_ | The name for this field. |
| `displayName` _string_ | A user-friendly name for this field. |
| `type` _string_ | The type of field: string, maskedstring, integer, or boolean. |
| `required` _boolean_ | Defines if the field is required or not. |
| `helpText` _string_ | Additional information about the field. |


#### DBaaSConnection



DBaaSConnection defines the schema for the DBaaSConnection API.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `dbaas.redhat.com/v1beta1`
| `kind` _string_ | `DBaaSConnection`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[DBaaSConnectionSpec](#dbaasconnectionspec)_ |  |


#### DBaaSConnectionPolicy



DBaaSConnectionPolicy sets a connection policy.

_Appears in:_
- [DBaaSInventoryPolicy](#dbaasinventorypolicy)

| Field | Description |
| --- | --- |
| `namespaces` _string_ | Namespaces where DBaaSConnection and DBaaSInstance objects are only allowed to reference a policy's inventories. Using an asterisk surrounded by single quotes ('*'), allows all namespaces. If not set in the policy or by an inventory object, connections are only allowed in the inventory's namespace. |
| `nsSelector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta)_ | Use a label selector to determine the namespaces where DBaaSConnection and DBaaSInstance objects are only allowed to reference a policy's inventories. A label selector is a label query over a set of resources. Results use a logical AND from matchExpressions and matchLabels queries. An empty label selector matches all objects. A null label selector matches no objects. |


#### DBaaSConnectionSpec



DBaaSConnectionSpec defines the desired state of a DBaaSConnection object.

_Appears in:_
- [DBaaSConnection](#dbaasconnection)
- [DBaaSProviderConnection](#dbaasproviderconnection)

| Field | Description |
| --- | --- |
| `inventoryRef` _[NamespacedName](#namespacedname)_ | A reference to the relevant DBaaSInventory custom resource (CR). |
| `databaseServiceID` _string_ | The ID of the database service to connect to, as seen in the status of the referenced DBaaSInventory. |
| `databaseServiceRef` _[NamespacedName](#namespacedname)_ | A reference to the database service CR used, if the DatabaseServiceID is not specified. |
| `databaseServiceType` _[DatabaseServiceType](#databaseservicetype)_ | The type of the database service to connect to, as seen in the status of the referenced DBaaSInventory. |


#### DBaaSInstance



DBaaSInstance defines the schema for the DBaaSInstance API.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `dbaas.redhat.com/v1beta1`
| `kind` _string_ | `DBaaSInstance`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[DBaaSInstanceSpec](#dbaasinstancespec)_ |  |


#### DBaaSInstanceSpec



DBaaSInstanceSpec defines the desired state of a DBaaSInstance object.

_Appears in:_
- [DBaaSInstance](#dbaasinstance)
- [DBaaSProviderInstance](#dbaasproviderinstance)

| Field | Description |
| --- | --- |
| `inventoryRef` _[NamespacedName](#namespacedname)_ | A reference to the relevant DBaaSInventory custom resource (CR). |
| `provisioningParameters` _object (keys:[ProvisioningParameterType](#provisioningparametertype), values:string)_ | Parameters with values used for provisioning. |


#### DBaaSInventory



DBaaSInventory defines the schema for the DBaaSInventory API. Inventory objects must be created in a valid namespace, determined by the existence of a DBaaSPolicy object.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `dbaas.redhat.com/v1beta1`
| `kind` _string_ | `DBaaSInventory`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[DBaaSOperatorInventorySpec](#dbaasoperatorinventoryspec)_ |  |


#### DBaaSInventoryPolicy



DBaaSInventoryPolicy sets the inventory policy.

_Appears in:_
- [DBaaSOperatorInventorySpec](#dbaasoperatorinventoryspec)
- [DBaaSPolicySpec](#dbaaspolicyspec)

| Field | Description |
| --- | --- |
| `disableProvisions` _boolean_ | Disables provisioning on inventory accounts. |
| `connections` _[DBaaSConnectionPolicy](#dbaasconnectionpolicy)_ | Namespaces where DBaaSConnection and DBaaSInstance objects are only allowed to reference a policy's inventories. |


#### DBaaSInventorySpec



DBaaSInventorySpec defines the inventory specifications for the provider's operators.

_Appears in:_
- [DBaaSOperatorInventorySpec](#dbaasoperatorinventoryspec)
- [DBaaSProviderInventory](#dbaasproviderinventory)

| Field | Description |
| --- | --- |
| `credentialsRef` _[LocalObjectReference](#localobjectreference)_ | The secret containing the provider-specific connection credentials to use with the provider's API endpoint. The format specifies the secret in the provider’s operator for its DBaaSProvider custom resource (CR), such as the CredentialFields key. The secret must exist within the same namespace as the inventory. |


#### DBaaSOperatorInventorySpec



DBaaSOperatorInventorySpec defines the desired state of a DBaaSInventory object.

_Appears in:_
- [DBaaSInventory](#dbaasinventory)

| Field | Description |
| --- | --- |
| `providerRef` _[NamespacedName](#namespacedname)_ | A reference to a DBaaSProvider custom resource (CR). |
| `DBaaSInventorySpec` _[DBaaSInventorySpec](#dbaasinventoryspec)_ | The properties that will be copied into the provider’s inventory. |
| `policy` _[DBaaSInventoryPolicy](#dbaasinventorypolicy)_ | The policy for this inventory. |


#### DBaaSPlatform



DBaaSPlatform defines the schema for the DBaaSPlatform API.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `dbaas.redhat.com/v1beta1`
| `kind` _string_ | `DBaaSPlatform`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[DBaaSPlatformSpec](#dbaasplatformspec)_ |  |


#### DBaaSPlatformSpec



DBaaSPlatformSpec defines the desired state of a DBaaSPlatform object.

_Appears in:_
- [DBaaSPlatform](#dbaasplatform)

| Field | Description |
| --- | --- |
| `syncPeriod` _integer_ | Sets the minimum interval, which the provider's operator controllers reconcile. The default value is 180 minutes. |


#### DBaaSPolicy



DBaaSPolicy enables administrative capabilities within a namespace, and sets a default inventory policy. Policy defaults can be overridden on a per-inventory basis.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `dbaas.redhat.com/v1beta1`
| `kind` _string_ | `DBaaSPolicy`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[DBaaSPolicySpec](#dbaaspolicyspec)_ |  |


#### DBaaSPolicySpec



DBaaSPolicySpec the specifications for a DBaaSPolicy object.

_Appears in:_
- [DBaaSPolicy](#dbaaspolicy)

| Field | Description |
| --- | --- |
| `DBaaSInventoryPolicy` _[DBaaSInventoryPolicy](#dbaasinventorypolicy)_ |  |


#### DBaaSProvider



DBaaSProvider defines the schema for the DBaaSProvider API.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `dbaas.redhat.com/v1beta1`
| `kind` _string_ | `DBaaSProvider`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[DBaaSProviderSpec](#dbaasproviderspec)_ |  |








#### DBaaSProviderSpec



DBaaSProviderSpec defines the desired state of a DBaaSProvider object.

_Appears in:_
- [DBaaSProvider](#dbaasprovider)

| Field | Description |
| --- | --- |
| `provider` _[DatabaseProviderInfo](#databaseproviderinfo)_ | Contains information about database provider and platform. |
| `groupVersion` _string_ | The DBaaS API group version supported by the provider. |
| `inventoryKind` _string_ | The name of the inventory custom resource definition (CRD) as defined by the database provider. |
| `connectionKind` _string_ | The name of the connection's custom resource definition (CRD) as defined by the provider. |
| `instanceKind` _string_ | The name of the instance's custom resource definition (CRD) as defined by the provider for provisioning. |
| `credentialFields` _[CredentialField](#credentialfield) array_ | Indicates what information to collect from the user interface and how to display fields in a form. |
| `allowsFreeTrial` _boolean_ | Indicates whether the provider offers free trials. |
| `externalProvisionURL` _string_ | The URL for provisioning instances by using the database provider's web portal. |
| `externalProvisionDescription` _string_ | Instructions on how to provision instances by using the database provider's web portal. |
| `provisioningParameters` _object (keys:[ProvisioningParameterType](#provisioningparametertype), values:[ProvisioningParameter](#provisioningparameter))_ | Parameter specifications used by the user interface (UI) for provisioning a database instance. |


#### DatabaseProviderInfo



DatabaseProviderInfo defines the information for a DBaaSProvider object.

_Appears in:_
- [DBaaSProviderSpec](#dbaasproviderspec)

| Field | Description |
| --- | --- |
| `name` _string_ | The name used to specify the service binding origin parameter. For example, 'OpenShift Database Access / Crunchy Bridge'. |
| `displayName` _string_ | A user-friendly name for this database provider. For example, 'Crunchy Bridge managed PostgreSQL'. |
| `displayDescription` _string_ | Indicates the description text shown for a database provider within the user interface. For example, the catalog tile description. |
| `icon` _[ProviderIcon](#providericon)_ | Indicates what icon to display on the catalog tile. |




#### DatabaseServiceType

_Underlying type:_ `string`

DatabaseServiceType defines the supported database service types.

_Appears in:_
- [DBaaSConnectionSpec](#dbaasconnectionspec)
- [DatabaseService](#databaseservice)



#### FieldDependency



FieldDependency defines the name and value of a dependency field.

_Appears in:_
- [ConditionalProvisioningParameterData](#conditionalprovisioningparameterdata)

| Field | Description |
| --- | --- |
| `field` _[ProvisioningParameterType](#provisioningparametertype)_ | Name of the dependency field. |
| `value` _string_ | Value of the dependency field. |




#### LocalObjectReference



LocalObjectReference contains enough information to locate the referenced object inside the same namespace.

_Appears in:_
- [DBaaSInventorySpec](#dbaasinventoryspec)

| Field | Description |
| --- | --- |
| `name` _string_ | Name of the referent. |


#### NamespacedName



NamespacedName defines the namespace and name of a k8s resource.

_Appears in:_
- [DBaaSConnectionSpec](#dbaasconnectionspec)
- [DBaaSInstanceSpec](#dbaasinstancespec)
- [DBaaSOperatorInventorySpec](#dbaasoperatorinventoryspec)

| Field | Description |
| --- | --- |
| `namespace` _string_ | The namespace where an object of a known type is stored. |
| `name` _string_ | The name for object of a known type. |




#### Option



Option defines the value and display value for an option in a dropdown menu, radio button, or checkbox.

_Appears in:_
- [ConditionalProvisioningParameterData](#conditionalprovisioningparameterdata)

| Field | Description |
| --- | --- |
| `value` _string_ | Value of the option. |
| `displayValue` _string_ | Corresponding display value. |






#### ProviderIcon



ProviderIcon follows the same field and naming formats as a comma-separated values (CSV) file.

_Appears in:_
- [DatabaseProviderInfo](#databaseproviderinfo)

| Field | Description |
| --- | --- |
| `base64data` _string_ |  |
| `mediatype` _string_ |  |


#### ProvisioningParameter



ProvisioningParameter provides information for a ProvisioningParameter object.

_Appears in:_
- [DBaaSProviderSpec](#dbaasproviderspec)

| Field | Description |
| --- | --- |
| `displayName` _string_ | A user-friendly name for this field. |
| `helpText` _string_ | Additional information about the field. |
| `conditionalData` _[ConditionalProvisioningParameterData](#conditionalprovisioningparameterdata) array_ | Lists of additional data containing the options or default values for the field. |


#### ProvisioningParameterType

_Underlying type:_ `string`

ProvisioningParameterType defines teh type for provisioning parameters

_Appears in:_
- [DBaaSInstanceSpec](#dbaasinstancespec)
- [DBaaSProviderSpec](#dbaasproviderspec)
- [FieldDependency](#fielddependency)



