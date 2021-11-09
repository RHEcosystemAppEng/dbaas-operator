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

package controllers

import (
	. "github.com/onsi/ginkgo"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

var _ = Describe("DBaaSInstance controller with errors", func() {
	Context("after creating DBaaSInstance without inventory", func() {
		instanceName := "test-instance-no-inventory"
		inventoryRefName := "test-inventory-no-exist-ref"
		DBaaSInstanceSpec := &v1alpha1.DBaaSInstanceSpec{
			InventoryRef: v1alpha1.NamespacedName{
				Name:      inventoryRefName,
				Namespace: testNamespace,
			},
			Name:          "test-instance",
			CloudProvider: "aws",
			CloudRegion:   "test-region",
			OtherInstanceParams: map[string]string{
				"testParam": "test-param",
			},
		}
		createdDBaaSInstance := &v1alpha1.DBaaSInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: testNamespace,
			},
			Spec: *DBaaSInstanceSpec,
		}

		BeforeEach(assertResourceCreation(createdDBaaSInstance))
		AfterEach(assertResourceDeletion(createdDBaaSInstance))
		It("reconcile with error", assertDBaaSResourceStatusUpdated(createdDBaaSInstance, metav1.ConditionFalse, v1alpha1.DBaaSInventoryNotFound))
	})
	Context("after creating DBaaSInstance with inventory that is not ready", func() {
		instanceName := "test-instance-not-ready"
		inventoryName := "test-instance-inventory-not-ready"
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
		DBaaSInstanceSpec := &v1alpha1.DBaaSInstanceSpec{
			InventoryRef: v1alpha1.NamespacedName{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Name:          "test-instance",
			CloudProvider: "aws",
			CloudRegion:   "test-region",
			OtherInstanceParams: map[string]string{
				"testParam": "test-param",
			},
		}
		createdDBaaSInstance := &v1alpha1.DBaaSInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: testNamespace,
			},
			Spec: *DBaaSInstanceSpec,
		}
		lastTransitionTime := getLastTransitionTimeForTest()
		providerInventoryStatus := &v1alpha1.DBaaSInventoryStatus{
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
					Status:             metav1.ConditionFalse,
					Reason:             "BackendError",
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
				},
			},
		}

		BeforeEach(assertResourceCreationIfNotExists(defaultProvider))
		BeforeEach(assertResourceCreationIfNotExists(&defaultTenant))
		BeforeEach(assertInventoryCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionFalse, testInventoryKind, providerInventoryStatus))
		BeforeEach(assertResourceCreationIfNotExists(createdDBaaSInstance))
		AfterEach(assertResourceDeletion(createdDBaaSInstance))
		AfterEach(assertResourceDeletion(createdDBaaSInventory))
		It("reconcile with error", assertDBaaSResourceStatusUpdated(createdDBaaSInstance, metav1.ConditionFalse, v1alpha1.DBaaSInventoryNotReady))
	})
})

var _ = Describe("DBaaSInstance controller - nominal", func() {
	BeforeEach(assertResourceCreationIfNotExists(defaultProvider))
	BeforeEach(assertResourceCreationIfNotExists(&defaultTenant))

	Describe("reconcile", func() {
		Context("after creating DBaaSInventory", func() {
			inventoryRefName := "test-inventory-ref"
			createdDBaaSInventory := &v1alpha1.DBaaSInventory{
				ObjectMeta: metav1.ObjectMeta{
					Name:      inventoryRefName,
					Namespace: testNamespace,
				},
				Spec: v1alpha1.DBaaSOperatorInventorySpec{
					ProviderRef: v1alpha1.NamespacedName{
						Name: testProviderName,
					},
					DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
						CredentialsRef: &v1alpha1.NamespacedName{
							Name:      "test-credentialsRef",
							Namespace: testNamespace,
						},
					},
				},
			}
			lastTransitionTime := getLastTransitionTimeForTest()
			providerInventoryStatus := &v1alpha1.DBaaSInventoryStatus{
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
			BeforeEach(assertInventoryCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
			AfterEach(assertResourceDeletion(createdDBaaSInventory))
			Context("after creating DBaaSInstance", func() {
				instanceName := "test-instance"
				DBaaSInstanceSpec := &v1alpha1.DBaaSInstanceSpec{
					InventoryRef: v1alpha1.NamespacedName{
						Name:      inventoryRefName,
						Namespace: testNamespace,
					},
					Name:          "test-instance",
					CloudProvider: "aws",
					CloudRegion:   "test-region",
					OtherInstanceParams: map[string]string{
						"testParam": "test-param",
					},
				}
				createdDBaaSInstance := &v1alpha1.DBaaSInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      instanceName,
						Namespace: testNamespace,
					},
					Spec: *DBaaSInstanceSpec,
				}
				BeforeEach(assertResourceCreation(createdDBaaSInstance))
				AfterEach(assertResourceDeletion(createdDBaaSInstance))

				It("should create a provider instance", assertProviderResourceCreated(createdDBaaSInstance, testInstanceKind, DBaaSInstanceSpec))
				Context("when updating provider instance status", func() {
					lastTransitionTime := getLastTransitionTimeForTest()
					status := &v1alpha1.DBaaSInstanceStatus{
						Conditions: []metav1.Condition{
							{
								Type:               "ProvisionReady",
								Status:             metav1.ConditionTrue,
								Reason:             "SyncOK",
								LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							},
						},
						InstanceID: "test-instance",
						InstanceInfo: map[string]string{
							"instanceInfo": "test-instance-info",
						},
						Phase: "ready",
					}
					It("should update DBaaSInstance status", assertDBaaSResourceProviderStatusUpdated(createdDBaaSInstance, metav1.ConditionTrue, testInstanceKind, status))
				})

				Context("when updating DBaaSInstance spec", func() {
					DBaaSInstanceSpec := &v1alpha1.DBaaSInstanceSpec{
						InventoryRef: v1alpha1.NamespacedName{
							Name:      inventoryRefName,
							Namespace: testNamespace,
						},
						Name:          "updated-test-instance",
						CloudProvider: "azure",
						CloudRegion:   "updated-test-region",
						OtherInstanceParams: map[string]string{
							"testParam": "updated-test-param",
						},
					}
					It("should update provider instance spec", assertProviderResourceSpecUpdated(createdDBaaSInstance, testInstanceKind, DBaaSInstanceSpec))
				})
			})
		})
	})
})
