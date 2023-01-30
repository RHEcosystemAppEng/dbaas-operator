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
	"context"
	"fmt"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this DBaaSInstance to the Hub version (v1beta1).
func (src *DBaaSInstance) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.DBaaSInstance)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	if err := src.Spec.ConvertTo(&dst.Spec); err != nil {
		return err
	}

	// Status
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.InstanceID = src.Status.InstanceID
	dst.Status.InstanceInfo = src.Status.InstanceInfo
	dst.Status.Phase = v1beta1.DBaasInstancePhase(src.Status.Phase)

	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *DBaaSInstance) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.DBaaSInstance)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	if err := dst.Spec.ConvertFrom(&src.Spec); err != nil {
		return err
	}

	// Status
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.InstanceID = src.Status.InstanceID
	dst.Status.InstanceInfo = src.Status.InstanceInfo
	dst.Status.Phase = DBaasInstancePhase(src.Status.Phase)

	return nil
}

// ConvertTo convert this DBaaSInstanceSpec to v1beta1.DBaaSInstanceSpec
func (src *DBaaSInstanceSpec) ConvertTo(dst *v1beta1.DBaaSInstanceSpec) error {
	dst.InventoryRef = v1beta1.NamespacedName(src.InventoryRef)
	dst.ProvisioningParameters = map[v1beta1.ProvisioningParameterType]string{}
	dst.ProvisioningParameters[v1beta1.ProvisioningName] = src.Name
	dst.ProvisioningParameters[v1beta1.ProvisioningCloudProvider] = src.CloudProvider
	dst.ProvisioningParameters[v1beta1.ProvisioningRegions] = src.CloudRegion
	if v1beta1.WebhookAPIClient == nil {
		return fmt.Errorf("webhook API client not set")
	}
	inventory := &DBaaSInventory{}
	if err := v1beta1.WebhookAPIClient.Get(context.TODO(), types.NamespacedName{
		Name:      src.InventoryRef.Name,
		Namespace: src.InventoryRef.Namespace,
	}, inventory); err != nil {
		return err
	}

	for key, val := range src.OtherInstanceParams {
		name := ConvertNameTo(inventory.Spec.ProviderRef.Name, key)
		if len(name) == 0 {
			continue
		}
		dst.ProvisioningParameters[name] = val
	}
	dst.ProvisioningParameters[v1beta1.ProvisioningPlan] = v1beta1.ProvisioningPlanFreeTrial
	return nil
}

// ConvertFrom converts the DBaaSInstanceSpec from the v1beta1 to this version.
func (dst *DBaaSInstanceSpec) ConvertFrom(src *v1beta1.DBaaSInstanceSpec) error {
	dst.Name = src.ProvisioningParameters[v1beta1.ProvisioningName]
	dst.CloudProvider = src.ProvisioningParameters[v1beta1.ProvisioningCloudProvider]
	dst.CloudRegion = src.ProvisioningParameters[v1beta1.ProvisioningRegions]
	dst.InventoryRef = NamespacedName(src.InventoryRef)
	dst.OtherInstanceParams = map[string]string{}
	inventory := &v1beta1.DBaaSInventory{}
	if v1beta1.WebhookAPIClient == nil {
		return fmt.Errorf("webhook API client not set")
	}
	if err := v1beta1.WebhookAPIClient.Get(context.TODO(), types.NamespacedName{
		Name:      src.InventoryRef.Name,
		Namespace: src.InventoryRef.Namespace,
	}, inventory); err != nil {
		return err
	}
	for key, val := range src.ProvisioningParameters {
		name := ConvertNameFrom(inventory.Spec.ProviderRef.Name, key)
		if len(name) == 0 {
			continue
		}
		// Special tweaking for mongo atlas
		if name == "ProjectName" {
			name = "projectName"
		}
		dst.OtherInstanceParams[name] = val
	}
	// Special tweaking for crunchy bridge
	if inventory.Spec.ProviderRef.Name == v1beta1.CrunchyBridgeRegistration {
		dst.CloudRegion = "us-east-1"
	}
	return nil
}
