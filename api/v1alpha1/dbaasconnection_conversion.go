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

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha2"
)

// ConvertTo converts this DBaaSConnection to the Hub version (v1alpha2).
func (src *DBaaSConnection) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.DBaaSConnection)

	dst.Spec.InventoryRef = v1alpha2.NamespacedName{
		Name:      src.Spec.InventoryRef.Name,
		Namespace: src.Spec.InventoryRef.Namespace,
	}
	dst.Spec.DatabaseServiceID = src.Spec.InstanceID
	if src.Spec.InstanceRef != nil {
		dst.Spec.DatabaseServiceRef = &v1alpha2.NamespacedName{
			Name:      src.Spec.InstanceRef.Name,
			Namespace: src.Spec.InstanceRef.Namespace,
		}
	}
	dst.Spec.DatabaseServiceType = v1alpha2.InstanceDatabaseService

	dst.ObjectMeta = src.ObjectMeta

	dst.Status.Conditions = src.Status.Conditions
	dst.Status.CredentialsRef = src.Status.CredentialsRef
	dst.Status.ConnectionInfoRef = src.Status.ConnectionInfoRef

	return nil
}

// ConvertFrom converts from the Hub version (v1alpha2) to this version.
func (dst *DBaaSConnection) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha2.DBaaSConnection)

	dst.Spec.InventoryRef = NamespacedName{
		Name:      src.Spec.InventoryRef.Name,
		Namespace: src.Spec.InventoryRef.Namespace,
	}
	if src.Spec.DatabaseServiceType == v1alpha2.InstanceDatabaseService {
		dst.Spec.InstanceID = src.Spec.DatabaseServiceID
		if src.Spec.DatabaseServiceRef != nil {
			dst.Spec.InstanceRef = &NamespacedName{
				Name:      src.Spec.DatabaseServiceRef.Name,
				Namespace: src.Spec.DatabaseServiceRef.Namespace,
			}
		}
	}

	dst.ObjectMeta = src.ObjectMeta

	dst.Status.Conditions = src.Status.Conditions
	dst.Status.CredentialsRef = src.Status.CredentialsRef
	dst.Status.ConnectionInfoRef = src.Status.ConnectionInfoRef

	return nil
}
