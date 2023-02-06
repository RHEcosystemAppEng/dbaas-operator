/*
Copyright 2022.

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
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

// notes on writing good spokes https://book.kubebuilder.io/multiversion-tutorial/conversion.html

// ConvertTo converts this DBaaSInventory to the Hub version (v1beta1).
func (src *DBaaSInventory) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.DBaaSInventory)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.CredentialsRef = (*v1beta1.LocalObjectReference)(src.Spec.CredentialsRef)
	if src.Spec.ConnectionNamespaces != nil {
		setPolicyObj(dst)
		dst.Spec.Policy.Connections.Namespaces = src.Spec.ConnectionNamespaces
	}
	if src.Spec.ConnectionNsSelector != nil {
		setPolicyObj(dst)
		dst.Spec.Policy.Connections.NsSelector = src.Spec.ConnectionNsSelector
	}
	if src.Spec.DisableProvisions != nil {
		setPolicyObj(dst)
		dst.Spec.Policy.DisableProvisions = src.Spec.DisableProvisions
	}
	dst.Spec.ProviderRef = v1beta1.NamespacedName(src.Spec.ProviderRef)

	// Status
	src.Status.ConvertTo(&dst.Status)

	return nil
}

func setPolicyObj(dst *v1beta1.DBaaSInventory) {
	if dst.Spec.Policy == nil {
		dst.Spec.Policy = &v1beta1.DBaaSInventoryPolicy{}
	}
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *DBaaSInventory) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.DBaaSInventory)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.ConvertFrom(&src.Spec)

	// Status
	dst.Status.ConvertFrom(&src.Status)

	return nil
}

// ConvertFrom converts the DBaaSInventorySpec from the v1beta1 to this version.
func (dst *DBaaSOperatorInventorySpec) ConvertFrom(src *v1beta1.DBaaSOperatorInventorySpec) {
	dst.CredentialsRef = (*LocalObjectReference)(src.CredentialsRef)
	if src.Policy != nil {
		dst.ConnectionNamespaces = src.Policy.Connections.Namespaces
		dst.ConnectionNsSelector = src.Policy.Connections.NsSelector
		dst.DisableProvisions = src.Policy.DisableProvisions
	}
	dst.ProviderRef = NamespacedName(src.ProviderRef)
}

// ConvertTo convert this DBaaSInventoryStatus to v1beta1.DBaaSInventoryStatus
func (src *DBaaSInventoryStatus) ConvertTo(dst *v1beta1.DBaaSInventoryStatus) {
	dst.Conditions = src.Conditions
	if src.Instances != nil {
		var services []v1beta1.DatabaseService
		for _, instance := range src.Instances {
			services = append(services, v1beta1.DatabaseService{
				ServiceID:   instance.InstanceID,
				ServiceName: instance.Name,
				ServiceInfo: instance.InstanceInfo,
			})
		}
		dst.DatabaseServices = services
	}
}

// ConvertFrom converts the DBaaSInventoryStatus from the v1beta1 to this version.
func (dst *DBaaSInventoryStatus) ConvertFrom(src *v1beta1.DBaaSInventoryStatus) {
	dst.Conditions = src.Conditions
	if src.DatabaseServices != nil {
		var instances []Instance
		for _, service := range src.DatabaseServices {
			instances = append(instances, Instance{
				InstanceID:   service.ServiceID,
				Name:         service.ServiceName,
				InstanceInfo: service.ServiceInfo,
			})
		}
		dst.Instances = instances
	}
}
