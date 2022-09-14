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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

var _ = Describe("DBaaSInstance controller with errors", func() {
	BeforeEach(assertResourceCreationIfNotExists(&testSecret))
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
		DBaaSInventorySpec := &v1alpha1.DBaaSInventorySpec{
			CredentialsRef: &v1alpha1.LocalObjectReference{
				Name: testSecret.Name,
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

		BeforeEach(assertResourceCreationIfNotExists(mongoProvider))
		BeforeEach(assertResourceCreationIfNotExists(&defaultPolicy))
		BeforeEach(assertDBaaSResourceStatusUpdated(&defaultPolicy, metav1.ConditionTrue, v1alpha1.Ready))
		BeforeEach(assertResourceCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionFalse, testInventoryKind, providerInventoryStatus))
		BeforeEach(assertResourceCreationIfNotExists(createdDBaaSInstance))
		AfterEach(assertResourceDeletion(createdDBaaSInstance))
		AfterEach(assertResourceDeletion(createdDBaaSInventory))
		It("reconcile with error", assertDBaaSResourceStatusUpdated(createdDBaaSInstance, metav1.ConditionFalse, v1alpha1.DBaaSInventoryNotReady))
	})
	Context("after creating DBaaSInstance in an invalid namespace", func() {
		instanceName := "test-instance-not-ready-2"
		inventoryName := "test-instance-inventory-diff-ns"
		DBaaSInventorySpec := &v1alpha1.DBaaSInventorySpec{
			CredentialsRef: &v1alpha1.LocalObjectReference{
				Name: testSecret.Name,
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
				DBaaSInventoryPolicy: v1alpha1.DBaaSInventoryPolicy{
					ConnectionNamespaces: &[]string{"valid-ns", "random"},
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
		otherNS := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "other",
			},
		}
		createdDBaaSInstance := &v1alpha1.DBaaSInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: otherNS.Name,
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
					Status:             metav1.ConditionTrue,
					Reason:             "SyncOK",
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
				},
			},
		}

		BeforeEach(assertResourceCreationIfNotExists(&otherNS))
		BeforeEach(assertResourceCreationIfNotExists(mongoProvider))
		BeforeEach(assertResourceCreationIfNotExists(&defaultPolicy))
		BeforeEach(assertDBaaSResourceStatusUpdated(&defaultPolicy, metav1.ConditionTrue, v1alpha1.Ready))
		BeforeEach(assertResourceCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
		BeforeEach(assertResourceCreationIfNotExists(createdDBaaSInstance))
		AfterEach(assertResourceDeletion(createdDBaaSInstance))
		AfterEach(assertResourceDeletion(createdDBaaSInventory))
		It("reconcile with error", assertDBaaSResourceStatusUpdated(createdDBaaSInstance, metav1.ConditionFalse, v1alpha1.DBaaSInvalidNamespace))
	})
})

var _ = Describe("DBaaSInstance controller - nominal", func() {
	BeforeEach(assertResourceCreationIfNotExists(&testSecret))
	BeforeEach(assertResourceCreationIfNotExists(mongoProvider))
	BeforeEach(assertResourceCreationIfNotExists(&defaultPolicy))
	BeforeEach(assertDBaaSResourceStatusUpdated(&defaultPolicy, metav1.ConditionTrue, v1alpha1.Ready))

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
						CredentialsRef: &v1alpha1.LocalObjectReference{
							Name: testSecret.Name,
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
			BeforeEach(assertResourceCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
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
								Type:               v1alpha1.DBaaSInstanceProviderSyncType,
								Status:             metav1.ConditionTrue,
								Reason:             "SyncOK",
								LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							},
						},
						InstanceID: "test-instance",
						InstanceInfo: map[string]string{
							"instanceInfo": "test-instance-info",
						},
						Phase: v1alpha1.InstancePhaseReady,
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

var _ = Describe("DBaaSInstance controller - valid dev namespaces", func() {
	BeforeEach(assertResourceCreationIfNotExists(&testSecret))
	BeforeEach(assertResourceCreationIfNotExists(mongoProvider))
	BeforeEach(assertResourceCreationIfNotExists(&defaultPolicy))
	BeforeEach(assertDBaaSResourceStatusUpdated(&defaultPolicy, metav1.ConditionTrue, v1alpha1.Ready))

	Describe("reconcile", func() {
		Context("after creating DBaaSInventory w/ addtl dev namespace set", func() {
			otherNS := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "other2",
				},
			}
			inventoryRefName := "test-inventory-ref-2"
			createdDBaaSInventory := &v1alpha1.DBaaSInventory{
				ObjectMeta: metav1.ObjectMeta{
					Name:      inventoryRefName,
					Namespace: testNamespace,
				},
				Spec: v1alpha1.DBaaSOperatorInventorySpec{
					ProviderRef: v1alpha1.NamespacedName{
						Name: testProviderName,
					},
					DBaaSInventoryPolicy: v1alpha1.DBaaSInventoryPolicy{
						ConnectionNamespaces: &[]string{otherNS.Name},
					},
					DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
						CredentialsRef: &v1alpha1.LocalObjectReference{
							Name: testSecret.Name,
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

			Context("after creating DBaaSInstance in separate, valid dev namespace", func() {
				instanceName := "test-instance-2"
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
						Namespace: otherNS.Name,
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
								Type:               v1alpha1.DBaaSInstanceProviderSyncType,
								Status:             metav1.ConditionTrue,
								Reason:             "SyncOK",
								LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							},
						},
						InstanceID: "test-instance",
						InstanceInfo: map[string]string{
							"instanceInfo": "test-instance-info",
						},
						Phase: v1alpha1.InstancePhaseReady,
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

			BeforeEach(assertResourceCreationIfNotExists(&otherNS))
			BeforeEach(assertResourceCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
			AfterEach(assertResourceDeletion(createdDBaaSInventory))
		})

		Context("after creating DBaaSInventory w/ wildcard dev namespace set", func() {
			otherNS := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "other3",
				},
			}
			inventoryRefName := "test-inventory-ref-3"
			createdDBaaSInventory := &v1alpha1.DBaaSInventory{
				ObjectMeta: metav1.ObjectMeta{
					Name:      inventoryRefName,
					Namespace: testNamespace,
				},
				Spec: v1alpha1.DBaaSOperatorInventorySpec{
					ProviderRef: v1alpha1.NamespacedName{
						Name: testProviderName,
					},
					DBaaSInventoryPolicy: v1alpha1.DBaaSInventoryPolicy{
						ConnectionNamespaces: &[]string{"*"},
					},
					DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
						CredentialsRef: &v1alpha1.LocalObjectReference{
							Name: testSecret.Name,
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

			Context("after creating DBaaSInstance in separate, valid dev namespace", func() {
				instanceName := "test-instance-3"
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
						Namespace: otherNS.Name,
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
								Type:               v1alpha1.DBaaSInstanceProviderSyncType,
								Status:             metav1.ConditionTrue,
								Reason:             "SyncOK",
								LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							},
						},
						InstanceID: "test-instance",
						InstanceInfo: map[string]string{
							"instanceInfo": "test-instance-info",
						},
						Phase: v1alpha1.InstancePhaseReady,
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

			BeforeEach(assertResourceCreationIfNotExists(&otherNS))
			BeforeEach(assertResourceCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
			AfterEach(assertResourceDeletion(createdDBaaSInventory))
		})

		Context("after creating DBaaSInventory w/ provisioning disabled", func() {
			otherNS := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "other3",
				},
			}
			isTrue := true
			inventoryRefName := "test-inventory-ref-4"
			createdDBaaSInventory := &v1alpha1.DBaaSInventory{
				ObjectMeta: metav1.ObjectMeta{
					Name:      inventoryRefName,
					Namespace: testNamespace,
				},
				Spec: v1alpha1.DBaaSOperatorInventorySpec{
					ProviderRef: v1alpha1.NamespacedName{
						Name: testProviderName,
					},
					DBaaSInventoryPolicy: v1alpha1.DBaaSInventoryPolicy{
						ConnectionNamespaces: &[]string{"*"},
						DisableProvisions:    &isTrue,
					},
					DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
						CredentialsRef: &v1alpha1.LocalObjectReference{
							Name: testSecret.Name,
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

			Context("instance should not provision", func() {
				instanceName := "test-instance-4"
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
						Namespace: otherNS.Name,
					},
					Spec: *DBaaSInstanceSpec,
				}
				BeforeEach(assertResourceCreation(createdDBaaSInstance))
				AfterEach(assertResourceDeletion(createdDBaaSInstance))

				It("should update DBaaSInstance status appropriately", assertDBaaSResourceStatusUpdated(createdDBaaSInstance, metav1.ConditionFalse, v1alpha1.DBaaSInventoryNotProvisionable))
			})

			BeforeEach(assertResourceCreationIfNotExists(&otherNS))
			BeforeEach(assertResourceCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
			AfterEach(assertResourceDeletion(createdDBaaSInventory))
		})
	})
})
