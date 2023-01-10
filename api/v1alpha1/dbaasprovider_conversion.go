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
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// notes on writing good spokes https://book.kubebuilder.io/multiversion-tutorial/conversion.html

// ConvertTo converts this DBaaSProvider to the Hub version (v1beta1).
func (src *DBaaSProvider) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.DBaaSProvider)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.AllowsFreeTrial = src.Spec.AllowsFreeTrial
	dst.Spec.ConnectionKind = src.Spec.ConnectionKind
	for i := range src.Spec.CredentialFields {
		dst.Spec.CredentialFields = append(dst.Spec.CredentialFields, v1beta1.CredentialField(src.Spec.CredentialFields[i]))
	}
	dst.Spec.ExternalProvisionDescription = src.Spec.ExternalProvisionDescription
	dst.Spec.ExternalProvisionURL = src.Spec.ExternalProvisionURL
	dst.Spec.InstanceKind = src.Spec.InstanceKind
	for i := range src.Spec.InstanceParameterSpecs {
		dst.Spec.InstanceParameterSpecs = append(dst.Spec.InstanceParameterSpecs, v1beta1.InstanceParameterSpec(src.Spec.InstanceParameterSpecs[i]))
	}
	dst.Spec.InventoryKind = src.Spec.InventoryKind
	dst.Spec.Provider = v1beta1.DatabaseProviderInfo{
		Name:               src.Spec.Provider.Name,
		DisplayName:        src.Spec.Provider.DisplayName,
		DisplayDescription: src.Spec.Provider.DisplayDescription,
		Icon:               v1beta1.ProviderIcon(src.Spec.Provider.Icon),
	}

	// Status
	dst.Status = v1beta1.DBaaSProviderStatus(src.Status)

	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *DBaaSProvider) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.DBaaSProvider)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.AllowsFreeTrial = src.Spec.AllowsFreeTrial
	dst.Spec.ConnectionKind = src.Spec.ConnectionKind
	for i := range src.Spec.CredentialFields {
		dst.Spec.CredentialFields = append(dst.Spec.CredentialFields, CredentialField(src.Spec.CredentialFields[i]))
	}
	dst.Spec.ExternalProvisionDescription = src.Spec.ExternalProvisionDescription
	dst.Spec.ExternalProvisionURL = src.Spec.ExternalProvisionURL
	dst.Spec.InstanceKind = src.Spec.InstanceKind
	for i := range src.Spec.InstanceParameterSpecs {
		dst.Spec.InstanceParameterSpecs = append(dst.Spec.InstanceParameterSpecs, InstanceParameterSpec(src.Spec.InstanceParameterSpecs[i]))
	}
	dst.Spec.InventoryKind = src.Spec.InventoryKind
	dst.Spec.Provider = DatabaseProvider{
		Name:               src.Spec.Provider.Name,
		DisplayName:        src.Spec.Provider.DisplayName,
		DisplayDescription: src.Spec.Provider.DisplayDescription,
		Icon:               ProviderIcon(src.Spec.Provider.Icon),
	}

	// Status
	dst.Status = DBaaSProviderStatus(src.Status)

	return nil
}
