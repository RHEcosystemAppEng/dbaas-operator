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
	dst.Status.Conditions = src.Status.Conditions
	for i := range src.Status.Instances {
		dst.Status.Instances = append(dst.Status.Instances, v1beta1.Instance(src.Status.Instances[i]))
	}

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
	dst.Spec.CredentialsRef = (*LocalObjectReference)(src.Spec.CredentialsRef)
	if src.Spec.Policy != nil {
		dst.Spec.ConnectionNamespaces = src.Spec.Policy.Connections.Namespaces
		dst.Spec.ConnectionNsSelector = src.Spec.Policy.Connections.NsSelector
		dst.Spec.DisableProvisions = src.Spec.Policy.DisableProvisions
	}
	dst.Spec.ProviderRef = NamespacedName(src.Spec.ProviderRef)

	// Status
	dst.Status.Conditions = src.Status.Conditions
	for i := range src.Status.Instances {
		dst.Status.Instances = append(dst.Status.Instances, Instance(src.Status.Instances[i]))
	}

	return nil
}
