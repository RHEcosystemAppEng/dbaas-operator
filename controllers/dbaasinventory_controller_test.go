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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

var _ = Describe("DBaaSInventory controller", func() {
	BeforeEach(assertResourceCreationIfNotExists(defaultProvider))
	BeforeEach(assertResourceCreationIfNotExists(&defaultTenant))

	Describe("reconcile", func() {
		Context("after creating DBaaSInventory", func() {
			inventoryName := "test-inventory"
			credentialsRefName := "test-credentials-ref"
			DBaaSInventorySpec := &v1alpha1.DBaaSInventorySpec{
				CredentialsRef: &v1alpha1.NamespacedName{
					Name:      credentialsRefName,
					Namespace: testNamespace,
				},
			}
			createdDBaaSInventory := &v1alpha1.DBaaSInventory{
				ObjectMeta: metav1.ObjectMeta{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				Spec: v1alpha1.DBaaSOperatorInventorySpec{
					ProviderRef: v1alpha1.NamespacedName{
						Name: testProviderName,
					},
					DBaaSInventorySpec: *DBaaSInventorySpec,
				},
			}

			BeforeEach(assertResourceCreation(createdDBaaSInventory))
			AfterEach(assertResourceDeletion(createdDBaaSInventory))

			It("should create a provider inventory", assertProviderResourceCreated(createdDBaaSInventory, testInventoryKind, DBaaSInventorySpec))

			Context("when updating provider inventory status", func() {
				lastTransitionTime, err := time.Parse(time.RFC3339, "2021-06-30T22:17:55-04:00")
				Expect(err).NotTo(HaveOccurred())
				lastTransitionTime = lastTransitionTime.In(time.Local)
				status := &v1alpha1.DBaaSInventoryStatus{
					Instances: []v1alpha1.Instance{
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
				It("should update DBaaSInventory status", assertDBaaSResourceStatusUpdated(createdDBaaSInventory, testInventoryKind, status))
			})

			Context("when updating DBaaSInventory spec", func() {
				DBaaSInventorySpec := &v1alpha1.DBaaSInventorySpec{
					CredentialsRef: &v1alpha1.NamespacedName{
						Name:      "updated-test-credentialsRef",
						Namespace: "updated-test-namespace",
					},
				}
				It("should update provider inventory spec", assertProviderResourceSpecUpdated(createdDBaaSInventory, testInventoryKind, DBaaSInventorySpec))
			})
		})
	})
})
