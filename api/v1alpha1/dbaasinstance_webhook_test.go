/*
Copyright 2022, Red Hat.

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("DBaaSInstance Webhook", func() {
	var (
		testinstanceName = "testinstance"
		provider         = DBaaSProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testProviderName,
				Namespace: testNamespace,
			},
			Spec: DBaaSProviderSpec{
				Provider: DatabaseProvider{
					Name: testProviderName,
				},
				InventoryKind:  testInventoryKind,
				ConnectionKind: testConnectionKind,
				InstanceKind:   testInstanceKind,
				CredentialFields: []CredentialField{
					{
						Key:      "field1",
						Type:     "String",
						Required: true,
					},
					{
						Key:      "field2",
						Type:     "String",
						Required: false,
					},
				},
				AllowsFreeTrial:              false,
				ExternalProvisionURL:         "",
				ExternalProvisionDescription: "",
				InstanceParameterSpecs: []InstanceParameterSpec{
					{
						Name:              "clusterName",
						DisplayName:       "clusterName",
						Type:              "string",
						InstanceFieldName: "name",
						Required:          true,
					},
					{
						Name:        "projectName",
						DisplayName: "projectName",
						Type:        "string",
						Required:    true,
					},
				},
			},
		}
		secret = corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Opaque",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      testSecretName,
				Namespace: testNamespace,
			},
			Data: map[string][]byte{
				"field1": []byte("test1"),
				"field2": []byte("test2"),
				"field3": []byte("test3"),
			},
		}
		inventory = DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: DBaaSOperatorInventorySpec{
				ProviderRef: NamespacedName{
					Name: testProviderName,
				},
				DBaaSInventorySpec: DBaaSInventorySpec{
					CredentialsRef: &LocalObjectReference{
						Name: testSecretName,
					},
				},
			},
		}
		instance = DBaaSInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testinstanceName,
				Namespace: testNamespace,
			},
			Spec: DBaaSInstanceSpec{
				InventoryRef: NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				Name: "testcluster",
				OtherInstanceParams: map[string]string{
					"projectName": "myproject",
				},
			},
		}
	)
	Context("nominal",
		func() {
			BeforeEach(assertResourceCreation(&provider))
			BeforeEach(assertResourceCreation(&secret))
			BeforeEach(assertResourceCreation(&inventory))
			AfterEach(assertResourceDeletion(&secret))
			AfterEach(assertResourceDeletion(&provider))
			AfterEach(assertResourceDeletion(&inventory))
			It("should succeed in creating the intstance", func() {
				inst := instance.DeepCopy()
				err := k8sClient.Create(ctx, inst)
				Expect(err).Should(BeNil())
				assertResourceDeletion(inst)
			})
		})
	Context("validation failures", func() {
		It("should fail without inventory", func() {
			inst := instance.DeepCopy()
			inst.Spec.InventoryRef.Name = "test-inventory-not-exist"
			err := k8sClient.Create(ctx, inst)
			Expect(err).Should(MatchError("admission webhook \"vdbaasinstance.kb.io\" denied the request: DBaaSInventory.dbaas.redhat.com \"test-inventory-not-exist\" not found"))
		})
		Context("with inventory",
			func() {
				BeforeEach(assertResourceCreation(&provider))
				BeforeEach(assertResourceCreation(&secret))
				BeforeEach(assertResourceCreation(&inventory))
				AfterEach(assertResourceDeletion(&secret))
				AfterEach(assertResourceDeletion(&provider))
				AfterEach(assertResourceDeletion(&inventory))
				It("should fail without required common field", func() {
					inst := instance.DeepCopy()
					inst.Spec.Name = ""
					err := k8sClient.Create(ctx, inst)
					Expect(err).Should(MatchError("admission webhook \"vdbaasinstance.kb.io\" denied the request: spec.name: Required value"))
				})
				It("should fail without required non-common field", func() {
					inst := instance.DeepCopy()
					inst.Spec.OtherInstanceParams = map[string]string{}
					err := k8sClient.Create(ctx, inst)
					Expect(err).Should(MatchError("admission webhook \"vdbaasinstance.kb.io\" denied the request: spec.otherInstanceParams.projectName: Required value"))
				})
			})

	})

})
