//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConditionalProvisioningParameterData) DeepCopyInto(out *ConditionalProvisioningParameterData) {
	*out = *in
	if in.Dependencies != nil {
		in, out := &in.Dependencies, &out.Dependencies
		*out = make([]FieldDependency, len(*in))
		copy(*out, *in)
	}
	if in.Options != nil {
		in, out := &in.Options, &out.Options
		*out = make([]Option, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConditionalProvisioningParameterData.
func (in *ConditionalProvisioningParameterData) DeepCopy() *ConditionalProvisioningParameterData {
	if in == nil {
		return nil
	}
	out := new(ConditionalProvisioningParameterData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CredentialField) DeepCopyInto(out *CredentialField) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CredentialField.
func (in *CredentialField) DeepCopy() *CredentialField {
	if in == nil {
		return nil
	}
	out := new(CredentialField)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSConnection) DeepCopyInto(out *DBaaSConnection) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSConnection.
func (in *DBaaSConnection) DeepCopy() *DBaaSConnection {
	if in == nil {
		return nil
	}
	out := new(DBaaSConnection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSConnection) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSConnectionList) DeepCopyInto(out *DBaaSConnectionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DBaaSConnection, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSConnectionList.
func (in *DBaaSConnectionList) DeepCopy() *DBaaSConnectionList {
	if in == nil {
		return nil
	}
	out := new(DBaaSConnectionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSConnectionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSConnectionSpec) DeepCopyInto(out *DBaaSConnectionSpec) {
	*out = *in
	out.InventoryRef = in.InventoryRef
	if in.InstanceRef != nil {
		in, out := &in.InstanceRef, &out.InstanceRef
		*out = new(NamespacedName)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSConnectionSpec.
func (in *DBaaSConnectionSpec) DeepCopy() *DBaaSConnectionSpec {
	if in == nil {
		return nil
	}
	out := new(DBaaSConnectionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSConnectionStatus) DeepCopyInto(out *DBaaSConnectionStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.CredentialsRef != nil {
		in, out := &in.CredentialsRef, &out.CredentialsRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
	if in.ConnectionInfoRef != nil {
		in, out := &in.ConnectionInfoRef, &out.ConnectionInfoRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSConnectionStatus.
func (in *DBaaSConnectionStatus) DeepCopy() *DBaaSConnectionStatus {
	if in == nil {
		return nil
	}
	out := new(DBaaSConnectionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSInstance) DeepCopyInto(out *DBaaSInstance) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSInstance.
func (in *DBaaSInstance) DeepCopy() *DBaaSInstance {
	if in == nil {
		return nil
	}
	out := new(DBaaSInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSInstance) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSInstanceList) DeepCopyInto(out *DBaaSInstanceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DBaaSInstance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSInstanceList.
func (in *DBaaSInstanceList) DeepCopy() *DBaaSInstanceList {
	if in == nil {
		return nil
	}
	out := new(DBaaSInstanceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSInstanceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSInstanceSpec) DeepCopyInto(out *DBaaSInstanceSpec) {
	*out = *in
	out.InventoryRef = in.InventoryRef
	if in.ProvisioningParameters != nil {
		in, out := &in.ProvisioningParameters, &out.ProvisioningParameters
		*out = make(map[ProvisioningParameterType]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSInstanceSpec.
func (in *DBaaSInstanceSpec) DeepCopy() *DBaaSInstanceSpec {
	if in == nil {
		return nil
	}
	out := new(DBaaSInstanceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSInstanceStatus) DeepCopyInto(out *DBaaSInstanceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.InstanceInfo != nil {
		in, out := &in.InstanceInfo, &out.InstanceInfo
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSInstanceStatus.
func (in *DBaaSInstanceStatus) DeepCopy() *DBaaSInstanceStatus {
	if in == nil {
		return nil
	}
	out := new(DBaaSInstanceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSInventory) DeepCopyInto(out *DBaaSInventory) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSInventory.
func (in *DBaaSInventory) DeepCopy() *DBaaSInventory {
	if in == nil {
		return nil
	}
	out := new(DBaaSInventory)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSInventory) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSInventoryList) DeepCopyInto(out *DBaaSInventoryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DBaaSInventory, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSInventoryList.
func (in *DBaaSInventoryList) DeepCopy() *DBaaSInventoryList {
	if in == nil {
		return nil
	}
	out := new(DBaaSInventoryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSInventoryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSInventoryPolicy) DeepCopyInto(out *DBaaSInventoryPolicy) {
	*out = *in
	if in.DisableProvisions != nil {
		in, out := &in.DisableProvisions, &out.DisableProvisions
		*out = new(bool)
		**out = **in
	}
	if in.ConnectionNamespaces != nil {
		in, out := &in.ConnectionNamespaces, &out.ConnectionNamespaces
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.ConnectionNsSelector != nil {
		in, out := &in.ConnectionNsSelector, &out.ConnectionNsSelector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSInventoryPolicy.
func (in *DBaaSInventoryPolicy) DeepCopy() *DBaaSInventoryPolicy {
	if in == nil {
		return nil
	}
	out := new(DBaaSInventoryPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSInventorySpec) DeepCopyInto(out *DBaaSInventorySpec) {
	*out = *in
	if in.CredentialsRef != nil {
		in, out := &in.CredentialsRef, &out.CredentialsRef
		*out = new(LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSInventorySpec.
func (in *DBaaSInventorySpec) DeepCopy() *DBaaSInventorySpec {
	if in == nil {
		return nil
	}
	out := new(DBaaSInventorySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSInventoryStatus) DeepCopyInto(out *DBaaSInventoryStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Instances != nil {
		in, out := &in.Instances, &out.Instances
		*out = make([]Instance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSInventoryStatus.
func (in *DBaaSInventoryStatus) DeepCopy() *DBaaSInventoryStatus {
	if in == nil {
		return nil
	}
	out := new(DBaaSInventoryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSOperatorInventorySpec) DeepCopyInto(out *DBaaSOperatorInventorySpec) {
	*out = *in
	out.ProviderRef = in.ProviderRef
	in.DBaaSInventorySpec.DeepCopyInto(&out.DBaaSInventorySpec)
	in.DBaaSInventoryPolicy.DeepCopyInto(&out.DBaaSInventoryPolicy)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSOperatorInventorySpec.
func (in *DBaaSOperatorInventorySpec) DeepCopy() *DBaaSOperatorInventorySpec {
	if in == nil {
		return nil
	}
	out := new(DBaaSOperatorInventorySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSPlatform) DeepCopyInto(out *DBaaSPlatform) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSPlatform.
func (in *DBaaSPlatform) DeepCopy() *DBaaSPlatform {
	if in == nil {
		return nil
	}
	out := new(DBaaSPlatform)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSPlatform) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSPlatformList) DeepCopyInto(out *DBaaSPlatformList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DBaaSPlatform, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSPlatformList.
func (in *DBaaSPlatformList) DeepCopy() *DBaaSPlatformList {
	if in == nil {
		return nil
	}
	out := new(DBaaSPlatformList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSPlatformList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSPlatformSpec) DeepCopyInto(out *DBaaSPlatformSpec) {
	*out = *in
	if in.SyncPeriod != nil {
		in, out := &in.SyncPeriod, &out.SyncPeriod
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSPlatformSpec.
func (in *DBaaSPlatformSpec) DeepCopy() *DBaaSPlatformSpec {
	if in == nil {
		return nil
	}
	out := new(DBaaSPlatformSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSPlatformStatus) DeepCopyInto(out *DBaaSPlatformStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.PlatformsStatus != nil {
		in, out := &in.PlatformsStatus, &out.PlatformsStatus
		*out = make([]PlatformStatus, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSPlatformStatus.
func (in *DBaaSPlatformStatus) DeepCopy() *DBaaSPlatformStatus {
	if in == nil {
		return nil
	}
	out := new(DBaaSPlatformStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSPolicy) DeepCopyInto(out *DBaaSPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSPolicy.
func (in *DBaaSPolicy) DeepCopy() *DBaaSPolicy {
	if in == nil {
		return nil
	}
	out := new(DBaaSPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSPolicyList) DeepCopyInto(out *DBaaSPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DBaaSPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSPolicyList.
func (in *DBaaSPolicyList) DeepCopy() *DBaaSPolicyList {
	if in == nil {
		return nil
	}
	out := new(DBaaSPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSPolicySpec) DeepCopyInto(out *DBaaSPolicySpec) {
	*out = *in
	in.DBaaSInventoryPolicy.DeepCopyInto(&out.DBaaSInventoryPolicy)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSPolicySpec.
func (in *DBaaSPolicySpec) DeepCopy() *DBaaSPolicySpec {
	if in == nil {
		return nil
	}
	out := new(DBaaSPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSPolicyStatus) DeepCopyInto(out *DBaaSPolicyStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSPolicyStatus.
func (in *DBaaSPolicyStatus) DeepCopy() *DBaaSPolicyStatus {
	if in == nil {
		return nil
	}
	out := new(DBaaSPolicyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSProvider) DeepCopyInto(out *DBaaSProvider) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSProvider.
func (in *DBaaSProvider) DeepCopy() *DBaaSProvider {
	if in == nil {
		return nil
	}
	out := new(DBaaSProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSProvider) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSProviderConnection) DeepCopyInto(out *DBaaSProviderConnection) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSProviderConnection.
func (in *DBaaSProviderConnection) DeepCopy() *DBaaSProviderConnection {
	if in == nil {
		return nil
	}
	out := new(DBaaSProviderConnection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSProviderInstance) DeepCopyInto(out *DBaaSProviderInstance) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSProviderInstance.
func (in *DBaaSProviderInstance) DeepCopy() *DBaaSProviderInstance {
	if in == nil {
		return nil
	}
	out := new(DBaaSProviderInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSProviderInventory) DeepCopyInto(out *DBaaSProviderInventory) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSProviderInventory.
func (in *DBaaSProviderInventory) DeepCopy() *DBaaSProviderInventory {
	if in == nil {
		return nil
	}
	out := new(DBaaSProviderInventory)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSProviderList) DeepCopyInto(out *DBaaSProviderList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DBaaSProvider, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSProviderList.
func (in *DBaaSProviderList) DeepCopy() *DBaaSProviderList {
	if in == nil {
		return nil
	}
	out := new(DBaaSProviderList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DBaaSProviderList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSProviderSpec) DeepCopyInto(out *DBaaSProviderSpec) {
	*out = *in
	out.Provider = in.Provider
	if in.CredentialFields != nil {
		in, out := &in.CredentialFields, &out.CredentialFields
		*out = make([]CredentialField, len(*in))
		copy(*out, *in)
	}
	if in.ProvisioningParameters != nil {
		in, out := &in.ProvisioningParameters, &out.ProvisioningParameters
		*out = make(map[ProvisioningParameterType]ProvisioningParameter, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSProviderSpec.
func (in *DBaaSProviderSpec) DeepCopy() *DBaaSProviderSpec {
	if in == nil {
		return nil
	}
	out := new(DBaaSProviderSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DBaaSProviderStatus) DeepCopyInto(out *DBaaSProviderStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DBaaSProviderStatus.
func (in *DBaaSProviderStatus) DeepCopy() *DBaaSProviderStatus {
	if in == nil {
		return nil
	}
	out := new(DBaaSProviderStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseProvider) DeepCopyInto(out *DatabaseProvider) {
	*out = *in
	out.Icon = in.Icon
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseProvider.
func (in *DatabaseProvider) DeepCopy() *DatabaseProvider {
	if in == nil {
		return nil
	}
	out := new(DatabaseProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FieldDependency) DeepCopyInto(out *FieldDependency) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FieldDependency.
func (in *FieldDependency) DeepCopy() *FieldDependency {
	if in == nil {
		return nil
	}
	out := new(FieldDependency)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Instance) DeepCopyInto(out *Instance) {
	*out = *in
	if in.InstanceInfo != nil {
		in, out := &in.InstanceInfo, &out.InstanceInfo
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Instance.
func (in *Instance) DeepCopy() *Instance {
	if in == nil {
		return nil
	}
	out := new(Instance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalObjectReference) DeepCopyInto(out *LocalObjectReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalObjectReference.
func (in *LocalObjectReference) DeepCopy() *LocalObjectReference {
	if in == nil {
		return nil
	}
	out := new(LocalObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespacedName) DeepCopyInto(out *NamespacedName) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespacedName.
func (in *NamespacedName) DeepCopy() *NamespacedName {
	if in == nil {
		return nil
	}
	out := new(NamespacedName)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObservabilityConfig) DeepCopyInto(out *ObservabilityConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObservabilityConfig.
func (in *ObservabilityConfig) DeepCopy() *ObservabilityConfig {
	if in == nil {
		return nil
	}
	out := new(ObservabilityConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Option) DeepCopyInto(out *Option) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Option.
func (in *Option) DeepCopy() *Option {
	if in == nil {
		return nil
	}
	out := new(Option)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PlatformConfig) DeepCopyInto(out *PlatformConfig) {
	*out = *in
	if in.Envs != nil {
		in, out := &in.Envs, &out.Envs
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PlatformConfig.
func (in *PlatformConfig) DeepCopy() *PlatformConfig {
	if in == nil {
		return nil
	}
	out := new(PlatformConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PlatformStatus) DeepCopyInto(out *PlatformStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PlatformStatus.
func (in *PlatformStatus) DeepCopy() *PlatformStatus {
	if in == nil {
		return nil
	}
	out := new(PlatformStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProviderIcon) DeepCopyInto(out *ProviderIcon) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProviderIcon.
func (in *ProviderIcon) DeepCopy() *ProviderIcon {
	if in == nil {
		return nil
	}
	out := new(ProviderIcon)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProvisioningParameter) DeepCopyInto(out *ProvisioningParameter) {
	*out = *in
	if in.ConditionalData != nil {
		in, out := &in.ConditionalData, &out.ConditionalData
		*out = make([]ConditionalProvisioningParameterData, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProvisioningParameter.
func (in *ProvisioningParameter) DeepCopy() *ProvisioningParameter {
	if in == nil {
		return nil
	}
	out := new(ProvisioningParameter)
	in.DeepCopyInto(out)
	return out
}
