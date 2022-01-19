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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testSecretName       = "testsecret"
	testSecretNameUpdate = "testsecretupdate"
	testProviderName     = "mongodb-atlas"
	testInventoryKind    = "MongoDBAtlasInventory"
	testConnectionKind   = "MongoDBAtlasConnection"
	testInstaneKind      = "MongoDBAtlasInstance"
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
			InstanceKind:   testInstaneKind,
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
				CredentialsRef: &NamespacedName{
					Name:      testSecretName,
					Namespace: testNamespace,
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
					inv.Spec.CredentialsRef.Namespace = ""
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
			It("missing required credential fields", func() {
				err := k8sClient.Create(ctx, &testDBaaSInventory)
				Expect(err).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request: spec.credentialsRef: Invalid value: v1alpha1.NamespacedName{Namespace:\"default\", Name:\"testsecret\"}: credentialsRef is invalid: field1 is required in secret testsecret"))
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
					Expect(err).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request: spec.credentialsRef: Invalid value: v1alpha1.NamespacedName{Namespace:\"default\", Name:\"testsecretupdate\"}: credentialsRef is invalid: field1 is required in secret testsecretupdate"))
				})
			})
		})
})

func assertResourceCreation(object client.Object) func() {
	return func() {
		By("creating resource")
		object.SetResourceVersion("")
		Expect(k8sClient.Create(ctx, object)).Should(Succeed())

		By("checking the resource created")
		Eventually(func() bool {
			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(object), object); err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
	}
}

func assertResourceDeletion(object client.Object) func() {
	return func() {
		By("deleting resource")
		Expect(k8sClient.Delete(ctx, object)).Should(Succeed())

		By("checking the resource deleted")
		Eventually(func() bool {
			err := k8sClient.Get(ctx, client.ObjectKeyFromObject(object), object)
			if err != nil && errors.IsNotFound(err) {
				return true
			}
			return false
		}, timeout, interval).Should(BeTrue())
	}
}
