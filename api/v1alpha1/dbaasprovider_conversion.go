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

var ProviderFieldsMap map[string]interface{} = map[string]interface{}{
	v1beta1.CrunchyBridgeRegistration: map[string]v1beta1.ProvisioningParameterType{
		"Name":     v1beta1.ProvisioningName,
		"Provider": v1beta1.ProvisioningCloudProvider,
		"TeamID":   v1beta1.ProvisioningTeamProject,
	},
	v1beta1.MongoDBAtlasRegistration: map[string]v1beta1.ProvisioningParameterType{
		"clusterName":  v1beta1.ProvisioningName,
		"providerName": v1beta1.ProvisioningCloudProvider,
		"ProjectName":  v1beta1.ProvisioningTeamProject,
	},
	v1beta1.RdsRegistration: map[string]v1beta1.ProvisioningParameterType{
		"Engine": v1beta1.ProvisioningDatabaseType,
	},
}

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
	dst.Spec.InventoryKind = src.Spec.InventoryKind
	dst.Spec.Provider = v1beta1.DatabaseProviderInfo{
		Name:               src.Spec.Provider.Name,
		DisplayName:        src.Spec.Provider.DisplayName,
		DisplayDescription: src.Spec.Provider.DisplayDescription,
		Icon:               v1beta1.ProviderIcon(src.Spec.Provider.Icon),
	}
	dst.Spec.ProvisioningParameters = map[v1beta1.ProvisioningParameterType]v1beta1.ProvisioningParameter{}
	for _, v := range src.Spec.InstanceParameterSpecs {
		name := ConvertNameTo(src.Name, v.Name)
		if len(name) == 0 {
			continue // this field is not supported in v1beta1
		}
		if name == v1beta1.ProvisioningCloudProvider {
			var defaultValue string
			if len(v.DefaultValue) > 0 {
				defaultValue = v.DefaultValue
			} else {
				defaultValue = "AWS"
			}
			s := v1beta1.ProvisioningParameter{
				DisplayName: v.DisplayName,
				ConditionalData: []v1beta1.ConditionalProvisioningParameterData{
					{
						Dependencies: []v1beta1.FieldDependency{
							{
								Field: v1beta1.ProvisioningPlan,
								Value: v1beta1.ProvisioningPlanFreeTrial,
							},
						},
						Options: []v1beta1.Option{
							{
								Value:        defaultValue,
								DisplayValue: defaultValue,
							},
						},
						DefaultValue: defaultValue,
					},
				},
			}
			dst.Spec.ProvisioningParameters[name] = s
		} else {
			// Others are input fields
			dst.Spec.ProvisioningParameters[name] = v1beta1.ProvisioningParameter{
				DisplayName: v.DisplayName,
			}
		}
	}

	//v1alpha1 only supports freetrial
	dst.Spec.ProvisioningParameters[v1beta1.ProvisioningPlan] = v1beta1.ProvisioningParameter{
		DisplayName: "Hosting plan",
		ConditionalData: []v1beta1.ConditionalProvisioningParameterData{
			{
				Options: []v1beta1.Option{
					{
						Value:        v1beta1.ProvisioningPlanFreeTrial,
						DisplayValue: "Free trial",
					},
				},
				DefaultValue: v1beta1.ProvisioningPlanFreeTrial,
			},
		},
	}

	// Status
	dst.Status = v1beta1.DBaaSProviderStatus{}

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
	for k, v := range src.Spec.ProvisioningParameters {
		var s InstanceParameterSpec
		switch k {
		case v1beta1.ProvisioningCloudProvider, v1beta1.ProvisioningName, v1beta1.ProvisioningDatabaseType, v1beta1.ProvisioningTeamProject:
			var defaultValue string
			if len(v.ConditionalData) > 0 {
				defaultValue = v.ConditionalData[0].DefaultValue
			}
			name := ConvertNameFrom(src.Name, k)
			s = InstanceParameterSpec{
				Name:         name,
				DisplayName:  v.DisplayName,
				Required:     true,
				Type:         "string",
				DefaultValue: defaultValue,
			}
			dst.Spec.InstanceParameterSpecs = append(dst.Spec.InstanceParameterSpecs, s)
		}
	}

	dst.Spec.InventoryKind = src.Spec.InventoryKind
	dst.Spec.Provider = DatabaseProvider{
		Name:               src.Spec.Provider.Name,
		DisplayName:        src.Spec.Provider.DisplayName,
		DisplayDescription: src.Spec.Provider.DisplayDescription,
		Icon:               ProviderIcon(src.Spec.Provider.Icon),
	}

	// Status
	dst.Status = DBaaSProviderStatus{}

	return nil
}

func ConvertNameTo(providerName, name string) v1beta1.ProvisioningParameterType {
	if m, ok := ProviderFieldsMap[providerName]; ok {
		m1 := m.(map[string]v1beta1.ProvisioningParameterType)
		if nameOut, ok := m1[name]; ok {
			return nameOut
		}
	}
	return ""
}

func ConvertNameFrom(providerName string, name v1beta1.ProvisioningParameterType) string {
	if m, ok := ProviderFieldsMap[providerName]; ok {
		m1 := m.(map[string]v1beta1.ProvisioningParameterType)
		for k, v := range m1 {
			if v == name {
				return k
			}
		}
	}
	return ""
}
