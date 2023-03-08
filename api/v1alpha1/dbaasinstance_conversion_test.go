/*
Copyright 2023 The OpenShift Database Access Authors.

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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

const (
	testProviderName   = "mongodb-atlas"
	testInventoryKind  = "MongoDBAtlasInventory"
	testConnectionKind = "MongoDBAtlasConnection"
	testInstanceKind   = "MongoDBAtlasInstance"
)

var _ = Context("DBaaSInstance Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		var testSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testsecret",
				Namespace: testNamespace,
				Labels: map[string]string{
					"test": "label",
				},
			},
		}

		var mongoProvider = &v1beta1.DBaaSProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name: testProviderName,
			},
			Spec: v1beta1.DBaaSProviderSpec{
				Provider: v1beta1.DatabaseProviderInfo{
					Name: testProviderName,
				},
				InventoryKind:                testInventoryKind,
				ConnectionKind:               testConnectionKind,
				InstanceKind:                 testInstanceKind,
				CredentialFields:             []v1beta1.CredentialField{},
				AllowsFreeTrial:              false,
				ExternalProvisionURL:         "",
				ExternalProvisionDescription: "",
			},
		}

		inventoryName := "test-inventory"
		DBaaSInventorySpec := &v1beta1.DBaaSInventorySpec{
			CredentialsRef: &v1beta1.LocalObjectReference{
				Name: "testsecret",
			},
		}
		createdDBaaSInventory := &v1beta1.DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: v1beta1.DBaaSOperatorInventorySpec{
				ProviderRef: v1beta1.NamespacedName{
					Name: testProviderName,
				},
				DBaaSInventorySpec: *DBaaSInventorySpec,
			},
		}
		BeforeEach(assertResourceCreation(testSecret))
		BeforeEach(assertResourceCreation(mongoProvider))
		BeforeEach(assertResourceCreation(createdDBaaSInventory))
		AfterEach(assertResourceDeletion(createdDBaaSInventory))
		AfterEach(assertResourceDeletion(mongoProvider))
		AfterEach(assertResourceDeletion(testSecret))
		Specify("converts to and from the same object", func() {
			src := DBaaSInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testName,
					Namespace: testNamespace,
				},
				Spec: DBaaSInstanceSpec{
					Name: "test",
					InventoryRef: NamespacedName{
						Name:      inventoryName,
						Namespace: testNamespace,
					},
					CloudProvider:       "test",
					CloudRegion:         "test",
					OtherInstanceParams: map[string]string{},
				},
				Status: DBaaSInstanceStatus{
					Conditions: []metav1.Condition{
						{
							Type: v1beta1.DBaaSConnectionProviderSyncType,
						},
					},
					InstanceID:   "test",
					InstanceInfo: map[string]string{},
					Phase:        InstancePhaseCreating,
				},
			}
			intermediate := v1beta1.DBaaSInstance{}
			dst := DBaaSInstance{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())
			Expect(dst).To(Equal(src))
		})
	})
})

func assertResourceCreation(object client.Object) func() {
	return func() {
		By("creating resource")
		timeout := 10
		object.SetResourceVersion("")
		err := v1beta1.WebhookAPIClient.Create(context.TODO(), object, &client.CreateOptions{})
		Expect(err).Should(Succeed())
		By("checking the resource created")
		Eventually(func() bool {
			if err := v1beta1.WebhookAPIClient.Get(context.TODO(), client.ObjectKeyFromObject(object), object); err != nil {
				return false
			}
			return true
		}, timeout).Should(BeTrue())
	}
}

func assertResourceDeletion(object client.Object) func() {
	return func() {
		By("deleting resource")
		timeout := 10
		Expect(v1beta1.WebhookAPIClient.Delete(ctx, object)).Should(Succeed())

		By("checking the resource deleted")
		Eventually(func() bool {
			err := v1beta1.WebhookAPIClient.Get(ctx, client.ObjectKeyFromObject(object), object)
			if err != nil && errors.IsNotFound(err) {
				return true
			}
			return false
		}, timeout).Should(BeTrue())
	}
}
