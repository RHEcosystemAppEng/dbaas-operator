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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

var _ = Describe("DBaaSConnection controller with errors", func() {
	BeforeEach(assertResourceCreationIfNotExists(&testSecret))
	Context("after creating DBaaSConnection without inventory", func() {
		connectionName := "test-connection-no-inventory"
		instanceID := "test-instanceID"
		inventoryRefName := "test-inventory-no-exist-ref"
		DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
			InventoryRef: v1alpha1.NamespacedName{
				Name:      inventoryRefName,
				Namespace: testNamespace,
			},
			InstanceID: instanceID,
		}
		createdDBaaSConnection := &v1alpha1.DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: *DBaaSConnectionSpec,
		}

		BeforeEach(assertResourceCreation(createdDBaaSConnection))
		AfterEach(assertResourceDeletion(createdDBaaSConnection))
		It("reconcile with error", assertDBaaSResourceStatusUpdated(createdDBaaSConnection, metav1.ConditionFalse, v1alpha1.DBaaSInventoryNotFound))
	})
	Context("after creating DBaaSConnection with inventory that is not ready", func() {
		connectionName := "test-connection-not-ready"
		instanceID := "test-instanceID"
		inventoryName := "test-connection-inventory-not-ready"
		DBaaSInventorySpec := &v1alpha1.DBaaSInventorySpec{
			CredentialsRef: &v1alpha1.NamespacedName{
				Name:      testSecret.Name,
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
		DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
			InventoryRef: v1alpha1.NamespacedName{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			InstanceID: instanceID,
		}
		createdDBaaSConnection := &v1alpha1.DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: *DBaaSConnectionSpec,
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
		BeforeEach(assertResourceCreationIfNotExists(createdDBaaSConnection))
		AfterEach(assertResourceDeletion(createdDBaaSConnection))
		AfterEach(assertResourceDeletion(createdDBaaSInventory))
		It("reconcile with error", assertDBaaSResourceStatusUpdated(createdDBaaSConnection, metav1.ConditionFalse, v1alpha1.DBaaSInventoryNotReady))
	})
	Context("after creating DBaaSConnection in an invalid namespace", func() {
		connectionName := "test-connection"
		instanceID := "test-instanceID"
		inventoryName := "test-connection-inventory"
		DBaaSInventorySpec := &v1alpha1.DBaaSInventorySpec{
			CredentialsRef: &v1alpha1.NamespacedName{
				Name:      testSecret.Name,
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
				DBaaSInventorySpec:   *DBaaSInventorySpec,
				ConnectionNamespaces: []string{"valid-ns", "random"},
			},
		}
		DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
			InventoryRef: v1alpha1.NamespacedName{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			InstanceID: instanceID,
		}
		otherNS := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "other",
			},
		}
		createdDBaaSConnection := &v1alpha1.DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: otherNS.Name,
			},
			Spec: *DBaaSConnectionSpec,
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
		BeforeEach(assertResourceCreationIfNotExists(defaultProvider))
		BeforeEach(assertResourceCreationIfNotExists(&defaultTenant))
		BeforeEach(assertInventoryCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
		BeforeEach(assertResourceCreationIfNotExists(createdDBaaSConnection))
		AfterEach(assertResourceDeletion(createdDBaaSConnection))
		AfterEach(assertResourceDeletion(createdDBaaSInventory))
		It("reconcile with error", assertDBaaSResourceStatusUpdated(createdDBaaSConnection, metav1.ConditionFalse, v1alpha1.DBaaSInvalidNamespace))
	})
})

var _ = Describe("DBaaSConnection controller - nominal", func() {
	BeforeEach(assertResourceCreationIfNotExists(&testSecret))
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
							Name:      testSecret.Name,
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

			Context("after creating DBaaSConnection", func() {
				connectionName := "test-connection-1"
				instanceID := "test-instanceID"
				DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
					InventoryRef: v1alpha1.NamespacedName{
						Name:      inventoryRefName,
						Namespace: testNamespace,
					},
					InstanceID: instanceID,
				}
				createdDBaaSConnection := &v1alpha1.DBaaSConnection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      connectionName,
						Namespace: testNamespace,
					},
					Spec: *DBaaSConnectionSpec,
				}
				BeforeEach(assertResourceCreation(createdDBaaSConnection))
				AfterEach(assertResourceDeletion(createdDBaaSConnection))

				It("should create a provider connection", assertProviderResourceCreated(createdDBaaSConnection, testConnectionKind, DBaaSConnectionSpec))
				Context("when updating provider connection status", func() {
					lastTransitionTime := getLastTransitionTimeForTest()
					status := &v1alpha1.DBaaSConnectionStatus{
						Conditions: []metav1.Condition{
							{
								Type:               "ReadyForBinding",
								Status:             metav1.ConditionTrue,
								Reason:             "SyncOK",
								LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							},
						},
						CredentialsRef: &v1.LocalObjectReference{
							Name: testSecret.Name,
						},
						ConnectionInfoRef: &v1.LocalObjectReference{
							Name: "testConnectionInfoRef",
						},
					}
					It("should update DBaaSConnection status", assertDBaaSResourceProviderStatusUpdated(createdDBaaSConnection, metav1.ConditionTrue, testConnectionKind, status))
				})

				Context("when updating DBaaSConnection spec", func() {
					DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
						InventoryRef: v1alpha1.NamespacedName{
							Name:      inventoryRefName,
							Namespace: testNamespace,
						},
						InstanceID: "updated-test-instanceID",
					}
					It("should update provider connection spec", assertProviderResourceSpecUpdated(createdDBaaSConnection, testConnectionKind, DBaaSConnectionSpec))
				})

			})

			BeforeEach(assertInventoryCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
			AfterEach(assertResourceDeletion(createdDBaaSInventory))
		})
	})

})

var _ = Describe("DBaaSConnection controller - valid dev namespaces", func() {
	BeforeEach(assertResourceCreationIfNotExists(&testSecret))
	BeforeEach(assertResourceCreationIfNotExists(defaultProvider))
	BeforeEach(assertResourceCreationIfNotExists(&defaultTenant))

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
					DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
						CredentialsRef: &v1alpha1.NamespacedName{
							Name:      testSecret.Name,
							Namespace: testNamespace,
						},
					},
					ConnectionNamespaces: []string{otherNS.Name},
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

			Context("after creating DBaaSConnections in separate, valid dev namespaces", func() {
				connectionName := "test-connection-2"
				instanceID := "test-instanceID"
				DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
					InventoryRef: v1alpha1.NamespacedName{
						Name:      inventoryRefName,
						Namespace: testNamespace,
					},
					InstanceID: instanceID,
				}
				createdDBaaSConnection := &v1alpha1.DBaaSConnection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      connectionName,
						Namespace: otherNS.Name,
					},
					Spec: *DBaaSConnectionSpec,
				}
				BeforeEach(assertResourceCreation(createdDBaaSConnection))
				AfterEach(assertResourceDeletion(createdDBaaSConnection))

				It("should create a provider connection", assertProviderResourceCreated(createdDBaaSConnection, testConnectionKind, DBaaSConnectionSpec))
				Context("when updating provider connection status", func() {
					lastTransitionTime := getLastTransitionTimeForTest()
					status := &v1alpha1.DBaaSConnectionStatus{
						Conditions: []metav1.Condition{
							{
								Type:               "ReadyForBinding",
								Status:             metav1.ConditionTrue,
								Reason:             "SyncOK",
								LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							},
						},
						CredentialsRef: &v1.LocalObjectReference{
							Name: testSecret.Name,
						},
						ConnectionInfoRef: &v1.LocalObjectReference{
							Name: "testConnectionInfoRef",
						},
					}
					It("should update DBaaSConnection status", assertDBaaSResourceProviderStatusUpdated(createdDBaaSConnection, metav1.ConditionTrue, testConnectionKind, status))
				})

				Context("when updating DBaaSConnection spec", func() {
					DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
						InventoryRef: v1alpha1.NamespacedName{
							Name:      inventoryRefName,
							Namespace: testNamespace,
						},
						InstanceID: "updated-test-instanceID",
					}
					It("should update provider connection spec", assertProviderResourceSpecUpdated(createdDBaaSConnection, testConnectionKind, DBaaSConnectionSpec))
				})

			})

			BeforeEach(assertResourceCreationIfNotExists(&otherNS))
			BeforeEach(assertInventoryCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
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
					DBaaSInventorySpec: v1alpha1.DBaaSInventorySpec{
						CredentialsRef: &v1alpha1.NamespacedName{
							Name:      testSecret.Name,
							Namespace: testNamespace,
						},
					},
					ConnectionNamespaces: []string{"*"},
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

			Context("after creating DBaaSConnections in separate, valid dev namespaces", func() {
				connectionName := "test-connection-3"
				instanceID := "test-instanceID"
				DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
					InventoryRef: v1alpha1.NamespacedName{
						Name:      inventoryRefName,
						Namespace: testNamespace,
					},
					InstanceID: instanceID,
				}
				createdDBaaSConnection := &v1alpha1.DBaaSConnection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      connectionName,
						Namespace: otherNS.Name,
					},
					Spec: *DBaaSConnectionSpec,
				}
				BeforeEach(assertResourceCreation(createdDBaaSConnection))
				AfterEach(assertResourceDeletion(createdDBaaSConnection))

				It("should create a provider connection", assertProviderResourceCreated(createdDBaaSConnection, testConnectionKind, DBaaSConnectionSpec))
				Context("when updating provider connection status", func() {
					lastTransitionTime := getLastTransitionTimeForTest()
					status := &v1alpha1.DBaaSConnectionStatus{
						Conditions: []metav1.Condition{
							{
								Type:               "ReadyForBinding",
								Status:             metav1.ConditionTrue,
								Reason:             "SyncOK",
								LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							},
						},
						CredentialsRef: &v1.LocalObjectReference{
							Name: testSecret.Name,
						},
						ConnectionInfoRef: &v1.LocalObjectReference{
							Name: "testConnectionInfoRef",
						},
					}
					It("should update DBaaSConnection status", assertDBaaSResourceProviderStatusUpdated(createdDBaaSConnection, metav1.ConditionTrue, testConnectionKind, status))
				})

				Context("when updating DBaaSConnection spec", func() {
					DBaaSConnectionSpec := &v1alpha1.DBaaSConnectionSpec{
						InventoryRef: v1alpha1.NamespacedName{
							Name:      inventoryRefName,
							Namespace: testNamespace,
						},
						InstanceID: "updated-test-instanceID",
					}
					It("should update provider connection spec", assertProviderResourceSpecUpdated(createdDBaaSConnection, testConnectionKind, DBaaSConnectionSpec))
				})

			})

			BeforeEach(assertResourceCreationIfNotExists(&otherNS))
			BeforeEach(assertInventoryCreationWithProviderStatus(createdDBaaSInventory, metav1.ConditionTrue, testInventoryKind, providerInventoryStatus))
			AfterEach(assertResourceDeletion(createdDBaaSInventory))
		})
	})

})
