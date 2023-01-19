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

// ConvertTo converts this DBaaSConnection to the Hub version (v1beta1).
func (src *DBaaSConnection) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.DBaaSConnection)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.InventoryRef = v1beta1.NamespacedName(src.Spec.InventoryRef)
	dst.Spec.InstanceID = src.Spec.InstanceID
	if src.Spec.InstanceRef != nil {
		dst.Spec.InstanceRef = &v1beta1.NamespacedName{
			Name:      src.Spec.InstanceRef.Name,
			Namespace: src.Spec.InstanceRef.Namespace,
		}
	}

	// Status
	dst.Status = v1beta1.DBaaSConnectionStatus(src.Status)

	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *DBaaSConnection) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.DBaaSConnection)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.InventoryRef = NamespacedName(src.Spec.InventoryRef)
	dst.Spec.InstanceID = src.Spec.InstanceID
	if src.Spec.InstanceRef != nil {
		dst.Spec.InstanceRef = &NamespacedName{
			Name:      src.Spec.InstanceRef.Name,
			Namespace: src.Spec.InstanceRef.Namespace,
		}
	}

	// Status
	dst.Status = DBaaSConnectionStatus(src.Status)

	return nil
}
