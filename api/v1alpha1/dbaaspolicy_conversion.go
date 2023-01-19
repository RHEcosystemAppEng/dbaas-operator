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

// ConvertTo converts this DBaaSPolicy to the Hub version (v1beta1).
func (src *DBaaSPolicy) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.DBaaSPolicy)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	if src.Spec.ConnectionNamespaces != nil {
		dst.Spec.Connections.Namespaces = src.Spec.ConnectionNamespaces
	}
	if src.Spec.ConnectionNsSelector != nil {
		dst.Spec.Connections.NsSelector = src.Spec.ConnectionNsSelector
	}
	if src.Spec.DisableProvisions != nil {
		dst.Spec.DisableProvisions = src.Spec.DisableProvisions
	}

	// Status
	dst.Status = v1beta1.DBaaSPolicyStatus(src.Status)

	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *DBaaSPolicy) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.DBaaSPolicy)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	if src.Spec.Connections.Namespaces != nil {
		dst.Spec.ConnectionNamespaces = src.Spec.Connections.Namespaces
	}
	if src.Spec.Connections.NsSelector != nil {
		dst.Spec.ConnectionNsSelector = src.Spec.Connections.NsSelector
	}
	if src.Spec.DisableProvisions != nil {
		dst.Spec.DisableProvisions = src.Spec.DisableProvisions
	}

	// Status
	dst.Status = DBaaSPolicyStatus(src.Status)

	return nil
}
