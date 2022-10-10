/*
Copyright 2021, Red Hat.

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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testSecretName        = "testsecret"
	testSecretNameUpdate  = "testsecretupdate"
	testProviderName      = "mongodb-atlas"
	testInventoryKind     = "MongoDBAtlasInventory"
	testConnectionKind    = "MongoDBAtlasConnection"
	testInstanceKind      = "MongoDBAtlasInstance"
	testSecretNameRDS     = "testsecretrds"
	testInventoryKindRDS  = "RDSInventory"
	testConnectionKindRDS = "RDSConnection"
	testInstanceKindRDS   = "RDSInstance"
	awsAccessKeyID        = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKey    = "AWS_SECRET_ACCESS_KEY" //#nosec G101
	awsRegion             = "AWS_REGION"
	ackResourceTags       = "ACK_RESOURCE_TAGS"
	ackLogLevel           = "ACK_LOG_LEVEL"
)

var (
	testProvider = DBaaSProvider{
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
			InstanceParameterSpecs:       []InstanceParameterSpec{},
		},
	}
	testProviderRDS = DBaaSProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rdsRegistration,
			Namespace: testNamespace,
		},
		Spec: DBaaSProviderSpec{
			Provider: DatabaseProvider{
				Name: rdsRegistration,
			},
			InventoryKind:  testInventoryKindRDS,
			ConnectionKind: testConnectionKindRDS,
			InstanceKind:   testInstanceKindRDS,
			CredentialFields: []CredentialField{
				{
					Key:         awsAccessKeyID,
					DisplayName: "AWS Access Key ID",
					Type:        "maskedstring",
					Required:    true,
				},
				{
					Key:         awsSecretAccessKey,
					DisplayName: "AWS Secret Access Key",
					Type:        "maskedstring",
					Required:    true,
				},
				{
					Key:         awsRegion,
					DisplayName: "AWS Region",
					Type:        "string",
					Required:    true,
				},
				{
					Key:         ackResourceTags,
					DisplayName: "ACK Resource Tags",
					Type:        "string",
					Required:    false,
				},
				{
					Key:         ackLogLevel,
					DisplayName: "ACK Log Level",
					Type:        "string",
					Required:    false,
				},
			},
			AllowsFreeTrial:              true,
			ExternalProvisionURL:         "",
			ExternalProvisionDescription: "",
			InstanceParameterSpecs:       []InstanceParameterSpec{},
		},
	}
	testSecret = corev1.Secret{
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
	testSecret2 = corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Opaque",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"field2": []byte("test2"),
			"field3": []byte("test3"),
		},
	}
	testSecret2update = corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Opaque",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretNameUpdate,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"field2": []byte("test2"),
			"field3": []byte("test3"),
		},
	}
	testSecret3 = corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Opaque",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"field1": []byte("test1"),
		},
	}
	testSecret3update = corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Opaque",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretNameUpdate,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"field1": []byte("test1"),
		},
	}
	testSecretRDS = corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Opaque",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretNameRDS,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"AWS_ACCESS_KEY_ID":     []byte("myaccesskeyid"),
			"AWS_SECRET_ACCESS_KEY": []byte("myaccesskey"),
			"AWS_REGION":            []byte("myregion"),
		},
	}
	testDBaaSInventory = DBaaSInventory{
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
			DBaaSInventoryPolicy: DBaaSInventoryPolicy{
				ConnectionNsSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{Key: "test", Operator: "DoesNotExist"},
					},
				},
			},
		},
	}
)

var _ = Describe("DBaaSInventory Webhook", func() {
	Context("creation succeeds",
		func() {
			BeforeEach(assertResourceCreation(&testProvider))
			AfterEach(assertResourceDeletion(&testProvider))
			Context("without optional fields", func() {
				BeforeEach(assertResourceCreation(&testSecret3))
				AfterEach(assertResourceDeletion(&testSecret3))
				It("should succeed without optional fields", func() {
					inv := testDBaaSInventory.DeepCopy()
					inv.Name = "inv-no-optional"
					err := k8sClient.Create(ctx, inv)
					Expect(err).Should(BeNil())
					assertResourceDeletion(inv)
				})
			})
			Context("with optional fields", func() {
				BeforeEach(assertResourceCreation(&testSecret))
				AfterEach(assertResourceDeletion(&testSecret))
				It("should succeed with optional fields", func() {
					const suffix = "a"
					secret := testSecret.DeepCopy()
					secret.Name = "testsecret" + suffix
					inv := testDBaaSInventory.DeepCopy()
					inv.Name = "testinventory" + suffix
					inv.Spec.CredentialsRef.Name = secret.Name
					secret.SetResourceVersion("")
					Expect(k8sClient.Create(ctx, secret)).Should(Succeed())
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
					Expect(err).Should(BeNil())
					inv.SetResourceVersion("")
					err = k8sClient.Create(ctx, inv)
					Expect(err).Should(BeNil())
					assertResourceDeletion(inv)
					assertResourceDeletion(secret)
				})
			})
			Context("without credentialRef.namespace", func() {
				It("should succeed without credentialRef.namespace", func() {
					const suffix = "b"
					secret := testSecret.DeepCopy()
					secret.Name = "testsecret" + suffix
					inv := testDBaaSInventory.DeepCopy()
					inv.Name = "testinventory" + suffix
					inv.Spec.CredentialsRef.Name = secret.Name
					secret.SetResourceVersion("")
					Expect(k8sClient.Create(ctx, secret)).Should(Succeed())
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
					Expect(err).Should(BeNil())
					inv.SetResourceVersion("")
					err = k8sClient.Create(ctx, inv)
					Expect(err).Should(BeNil())
					assertResourceDeletion(inv)
					assertResourceDeletion(secret)
				})
			})
		})
	Context("creation fails",
		func() {
			BeforeEach(assertResourceCreation(&testSecret2))
			BeforeEach(assertResourceCreation(&testProvider))
			AfterEach(assertResourceDeletion(&testProvider))
			AfterEach(assertResourceDeletion(&testSecret2))
			It("missing required values field", func() {
				inv := testDBaaSInventory.DeepCopy()
				inv.Spec.ConnectionNsSelector = &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{Key: "test", Operator: "In"},
					},
				}
				err := k8sClient.Create(ctx, inv)
				Expect(err).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request: values: Invalid value: []string(nil): for 'in', 'notin' operators, values set can't be empty"))
			})
			It("missing required credential fields", func() {
				err := k8sClient.Create(ctx, &testDBaaSInventory)
				Expect(err).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request: spec.credentialsRef: Invalid value: v1alpha1.LocalObjectReference{Name:\"testsecret\"}: credentialsRef is invalid: field1 is required in secret testsecret"))
			})
		})
	Context("update",
		func() {
			BeforeEach(assertResourceCreation(&testSecret))
			BeforeEach(assertResourceCreation(&testProvider))
			BeforeEach(assertResourceCreation(&testDBaaSInventory))
			AfterEach(assertResourceDeletion(&testDBaaSInventory))
			AfterEach(assertResourceDeletion(&testProvider))
			AfterEach(assertResourceDeletion(&testSecret))
			Context("nominal", func() {
				BeforeEach(assertResourceCreation(&testSecret3update))
				AfterEach(assertResourceDeletion(&testSecret3update))
				It("Update CR should succeed", func() {
					inv := testDBaaSInventory.DeepCopy()
					Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(inv), inv)).Should(Succeed())
					inv.Spec.ConnectionNsSelector.MatchExpressions = nil
					inv.Spec.CredentialsRef.Name = testSecretNameUpdate
					err := k8sClient.Update(ctx, inv)
					Expect(err).Should(BeNil())
				})
			})
			Context("update fails", func() {
				BeforeEach(assertResourceCreation(&testSecret2update))
				AfterEach(assertResourceDeletion(&testSecret2update))
				It("update fails with missing required credential fields", func() {
					inv := testDBaaSInventory.DeepCopy()
					Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(inv), inv)).Should(Succeed())
					inv.Spec.CredentialsRef.Name = testSecretNameUpdate
					err := k8sClient.Update(ctx, inv)
					Expect(err).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request: spec.credentialsRef: Invalid value: v1alpha1.LocalObjectReference{Name:\"testsecretupdate\"}: credentialsRef is invalid: field1 is required in secret testsecretupdate"))
				})
				It("update fails with missing required values field", func() {
					inv := testDBaaSInventory.DeepCopy()
					Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(inv), inv)).Should(Succeed())
					inv.Spec.ConnectionNsSelector.MatchExpressions = []metav1.LabelSelectorRequirement{{Key: "test", Operator: "In"}}
					err := k8sClient.Update(ctx, inv)
					Expect(err).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request: values: Invalid value: []string(nil): for 'in', 'notin' operators, values set can't be empty"))
				})
			})
			It("update fails with provider name change", func() {
				inv := testDBaaSInventory.DeepCopy()
				Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(inv), inv)).Should(Succeed())
				inv.Spec.ProviderRef.Name = "crunchy-registration"
				err := k8sClient.Update(ctx, inv)
				Expect(err).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request: spec.providerRef.name: Invalid value: \"crunchy-registration\": provider name is immutable for provider accounts"))
			})
		})
	Context("After creating DBaaSInventory for RDS", func() {
		testSecretRDS2 := corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Opaque",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      testSecretNameRDS,
				Namespace: testNamespace2,
			},
			Data: map[string][]byte{
				"AWS_ACCESS_KEY_ID":     []byte("myaccesskeyid"),
				"AWS_SECRET_ACCESS_KEY": []byte("myaccesskey"),
				"AWS_REGION":            []byte("myregion"),
			},
		}
		testDBaaSInventoryRDS := DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-inventory-rds",
				Namespace: testNamespace,
			},
			Spec: DBaaSOperatorInventorySpec{
				ProviderRef: NamespacedName{
					Name: rdsRegistration,
				},
				DBaaSInventorySpec: DBaaSInventorySpec{
					CredentialsRef: &LocalObjectReference{
						Name: testSecretNameRDS,
					},
				},
			},
		}
		BeforeEach(assertResourceCreation(&testSecretRDS2))
		BeforeEach(assertResourceCreation(&testSecretRDS))
		BeforeEach(assertResourceCreation(&testProviderRDS))
		BeforeEach(assertResourceCreation(&testDBaaSInventoryRDS))
		AfterEach(assertResourceDeletion(&testDBaaSInventoryRDS))
		AfterEach(assertResourceDeletion(&testProviderRDS))
		AfterEach(assertResourceDeletion(&testSecretRDS))
		AfterEach(assertResourceDeletion(&testSecretRDS2))
		Context("creating another DBaaSInventory for RDS", func() {
			It("should not allow creating DBaaSInventory for RDS in the same namespace", func() {
				testDBaaSInventoryRDSNotAllowed := DBaaSInventory{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-inventory-rds-not-allowed",
						Namespace: testNamespace,
					},
					Spec: DBaaSOperatorInventorySpec{
						ProviderRef: NamespacedName{
							Name: rdsRegistration,
						},
						DBaaSInventorySpec: DBaaSInventorySpec{
							CredentialsRef: &LocalObjectReference{
								Name: testSecretNameRDS,
							},
						},
					},
				}
				By("creating DBaaSInventory for RDS")
				Expect(k8sClient.Create(ctx, &testDBaaSInventoryRDSNotAllowed)).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request:" +
					" only one provider account for RDS can exist in a cluster, but there is already a provider account test-inventory-rds created"))
			})
			It("should not allow creating DBaaSInventory for RDS in another namespace", func() {
				testDBaaSInventoryRDSNotAllowed := DBaaSInventory{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-inventory-rds-not-allowed",
						Namespace: testNamespace2,
					},
					Spec: DBaaSOperatorInventorySpec{
						ProviderRef: NamespacedName{
							Name: rdsRegistration,
						},
						DBaaSInventorySpec: DBaaSInventorySpec{
							CredentialsRef: &LocalObjectReference{
								Name: testSecretNameRDS,
							},
						},
					},
				}
				err := k8sClient.Create(ctx, &testDBaaSInventoryRDSNotAllowed)
				Expect(err).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request:" +
					" only one provider account for RDS can exist in a cluster, but there is already a provider account test-inventory-rds created"))
			})
		})
	})
})
