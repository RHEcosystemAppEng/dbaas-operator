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

package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

var _ = Describe("DBaaSInventory controller with errors", func() {
	ns := "testns-no-policy"
	nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	BeforeEach(assertResourceCreationIfNotExists(nsSpec))
	Context("after creating DBaaSInventory without policy in the target namespace", func() {
		inventoryName := "test-inventory-no-policy"
		testSecret2 := testSecret.DeepCopy()
		testSecret2.Namespace = ns
		DBaaSInventorySpec := &v1beta1.DBaaSInventorySpec{
			CredentialsRef: &v1beta1.LocalObjectReference{
				Name: testSecret2.Name,
			},
		}
		testCreatedDBaaSInventory := &v1beta1.DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: ns,
			},
			Spec: v1beta1.DBaaSOperatorInventorySpec{
				ProviderRef: v1beta1.NamespacedName{
					Name: testProviderName,
				},
				DBaaSInventorySpec: *DBaaSInventorySpec,
			},
		}
		BeforeEach(assertResourceCreationIfNotExists(testSecret2))
		BeforeEach(assertResourceCreationIfNotExists(testCreatedDBaaSInventory))
		It("reconcile with error", assertDBaaSResourceStatusUpdated(testCreatedDBaaSInventory, metav1.ConditionFalse, v1beta1.DBaaSPolicyNotFound))
	})

	Context("after creating DBaaSInventory without valid provider", func() {
		inventoryName := "test-inventory-no-provider"
		providerName := "provider-no-exist"
		DBaaSInventorySpec := &v1beta1.DBaaSInventorySpec{
			CredentialsRef: &v1beta1.LocalObjectReference{
				Name: testSecret.Name,
			},
		}
		createdDBaaSInventory := &v1beta1.DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: v1beta1.DBaaSOperatorInventorySpec{
				ProviderRef: v1beta1.NamespacedName{
					Name: providerName,
				},
				DBaaSInventorySpec: *DBaaSInventorySpec,
			},
		}
		BeforeEach(assertResourceCreationIfNotExists(&testSecret))
		BeforeEach(assertResourceCreationIfNotExists(&defaultPolicy))
		BeforeEach(assertDBaaSResourceStatusUpdated(&defaultPolicy, metav1.ConditionTrue, v1beta1.Ready))
		Context("when updating DBaaSInventory spec to non-existant provider", func() {
			It("should not be allow", func() {
				err := dRec.Create(ctx, createdDBaaSInventory)
				Expect(err).Should(MatchError("admission webhook \"vdbaasinventory.kb.io\" denied the request: " +
					"DBaaSProvider.dbaas.redhat.com \"provider-no-exist\" not found"))
			})
		})
	})
})

var _ = Describe("DBaaSInventory controller - nominal", func() {
	BeforeEach(assertResourceCreationIfNotExists(&testSecret))
	BeforeEach(assertResourceCreationIfNotExists(mongoProvider))
	BeforeEach(assertResourceCreationIfNotExists(&defaultPolicy))
	BeforeEach(assertDBaaSResourceStatusUpdated(&defaultPolicy, metav1.ConditionTrue, v1beta1.Ready))

	Describe("reconcile", func() {
		Context("after creating DBaaSInventory", func() {
			inventoryName := "test-inventory"
			DBaaSInventorySpec := &v1beta1.DBaaSInventorySpec{
				CredentialsRef: &v1beta1.LocalObjectReference{
					Name: testSecret.Name,
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

			BeforeEach(assertResourceCreationIfNotExists(&testSecret))
			BeforeEach(assertResourceCreation(createdDBaaSInventory))
			AfterEach(assertResourceDeletion(createdDBaaSInventory))

			It("should create a provider inventory", assertProviderResourceCreated(createdDBaaSInventory, testInventoryKind, DBaaSInventorySpec))

			Context("when updating provider inventory status", func() {
				lastTransitionTime := getLastTransitionTimeForTest()
				status := &v1beta1.DBaaSInventoryStatus{
					Instances: []v1beta1.Instance{
						{
							InstanceID: "testInstanceID",
							Name:       "testInstance",
							InstanceInfo: map[string]string{
								"testInstanceInfo": "testInstanceInfo",
							},
						},
					},
					Conditions: []metav1.Condition{
						{
							Type:               "SpecSynced",
							Status:             metav1.ConditionTrue,
							Reason:             "SyncOK",
							LastTransitionTime: metav1.Time{Time: lastTransitionTime},
						},
					},
				}
				BeforeEach(assertResourceCreationIfNotExists(&testSecret))
				It("should update DBaaSInventory status", assertDBaaSResourceProviderStatusUpdated(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, status))
			})

			Context("when updating DBaaSInventory spec", func() {
				updatedTestSecret := v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "updated-test-credentials",
						Namespace: testNamespace,
					},
				}
				DBaaSInventorySpec := &v1beta1.DBaaSInventorySpec{
					CredentialsRef: &v1beta1.LocalObjectReference{
						Name: updatedTestSecret.Name,
					},
				}
				BeforeEach(assertResourceCreationIfNotExists(&updatedTestSecret))
				It("should update provider inventory spec", assertProviderResourceSpecUpdated(createdDBaaSInventory, testInventoryKind, DBaaSInventorySpec))
				It("should return the secret without error and with proper label", func() {
					getSecret := v1.Secret{}
					err := dRec.Get(ctx, client.ObjectKeyFromObject(&updatedTestSecret), &getSecret)
					Expect(err).NotTo(HaveOccurred())
					labels := getSecret.GetLabels()
					Expect(labels).Should(Not(BeNil()))
					Expect(labels[v1beta1.TypeLabelKeyMongo]).Should(Equal(v1beta1.TypeLabelValue))
				})
			})
		})
	})
})
