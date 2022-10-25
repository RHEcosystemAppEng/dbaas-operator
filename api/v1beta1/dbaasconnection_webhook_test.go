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

package v1beta1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var inventoryName = "test-inventory"
var connectionName = "test-connection"
var databaseServiceID = "test-databaseServiceID"
var databaseServiceName = "test-databaseService"
var databaseServiceType = DatabaseServiceType("test-databaseServiceType")
var testDBaaSConnection = &DBaaSConnection{
	ObjectMeta: metav1.ObjectMeta{
		Name:      connectionName,
		Namespace: testNamespace,
	},
	Spec: DBaaSConnectionSpec{
		InventoryRef: NamespacedName{
			Name:      inventoryName,
			Namespace: testNamespace,
		},
		DatabaseServiceID: databaseServiceID,
	},
}

var _ = Describe("DBaaSConnection Webhook", func() {
	Context("after creating DBaaSConnection", func() {
		BeforeEach(func() {
			By("creating DBaaSConnection")
			testDBaaSConnection.SetResourceVersion("")
			Expect(k8sClient.Create(ctx, testDBaaSConnection)).Should(Succeed())

			By("checking DBaaSConnection created")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnection), &DBaaSConnection{}); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			By("deleting DBaaSConnection")
			Expect(k8sClient.Delete(ctx, testDBaaSConnection)).Should(Succeed())

			By("checking DBaaSConnection deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnection), &DBaaSConnection{})
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		DescribeTable("checking invalid DBaaSConnection updates",
			func(specUpdateFn func(*DBaaSConnectionSpec), expectedErr interface{}) {
				updatedDBaaSConnection := &DBaaSConnection{}
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnection), updatedDBaaSConnection)
				Expect(err).NotTo(HaveOccurred())

				specUpdateFn(&updatedDBaaSConnection.Spec)
				err = k8sClient.Update(ctx, updatedDBaaSConnection)
				Expect(err).Should(MatchError(expectedErr))
			},
			Entry("not allow updating database service ID",
				func(spec *DBaaSConnectionSpec) {
					spec.DatabaseServiceID = "updated-databaseServiceID"
				},
				"admission webhook \"vdbaasconnection.kb.io\" denied the request: "+
					"spec.databaseServiceID: Invalid value: \"updated-databaseServiceID\": databaseServiceID is immutable"),
			Entry("not allow updating inventoryRef",
				func(spec *DBaaSConnectionSpec) {
					spec.InventoryRef.Name = "updated-inventory"
				},
				"admission webhook \"vdbaasconnection.kb.io\" denied the request: "+
					"spec.inventoryRef: Invalid value: v1beta1.NamespacedName{Namespace:\"default\", Name:\"updated-inventory\"}: "+
					"inventoryRef is immutable"),
			Entry("not allow updating databaseServiceRef",
				func(spec *DBaaSConnectionSpec) {
					spec.DatabaseServiceRef = &NamespacedName{
						Name:      "updated-databaseService",
						Namespace: testNamespace,
					}
				},
				"admission webhook \"vdbaasconnection.kb.io\" denied the request: "+
					"spec.databaseServiceRef: Invalid value: v1beta1.NamespacedName{Namespace:\"default\", Name:\"updated-databaseService\"}: "+
					"databaseServiceRef is immutable"),
			Entry("not allow updating databaseServiceType",
				func(spec *DBaaSConnectionSpec) {
					spec.DatabaseServiceType = &databaseServiceType
				},
				"admission webhook \"vdbaasconnection.kb.io\" denied the request: "+
					"spec.databaseServiceType: Invalid value: \"test-databaseServiceType\": databaseServiceType is immutable"),
		)
	})

	Context("after trying to create DBaaSConnection without database service info", func() {
		It("should not allow creating the DBaaSConnection", func() {
			testDBaaSConnectionNoDatabaseService := &DBaaSConnection{
				ObjectMeta: metav1.ObjectMeta{
					Name:      connectionName,
					Namespace: testNamespace,
				},
				Spec: DBaaSConnectionSpec{
					InventoryRef: NamespacedName{
						Name:      inventoryName,
						Namespace: testNamespace,
					},
				},
			}
			err := k8sClient.Create(ctx, testDBaaSConnectionNoDatabaseService)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.databaseServiceID: Invalid value: \"\": either databaseServiceID or databaseServiceRef must be specified"))
		})
	})

	Context("after trying to create DBaaSConnection with both database service ID and database service reference", func() {
		It("should not allow creating the DBaaSConnection", func() {
			testDBaaSConnectionNoDatabaseService := &DBaaSConnection{
				ObjectMeta: metav1.ObjectMeta{
					Name:      connectionName,
					Namespace: testNamespace,
				},
				Spec: DBaaSConnectionSpec{
					InventoryRef: NamespacedName{
						Name:      inventoryName,
						Namespace: testNamespace,
					},
					DatabaseServiceID: databaseServiceID,
					DatabaseServiceRef: &NamespacedName{
						Name:      databaseServiceName,
						Namespace: testNamespace,
					},
				},
			}
			err := k8sClient.Create(ctx, testDBaaSConnectionNoDatabaseService)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.databaseServiceID: Invalid value: \"test-databaseServiceID\": both databaseServiceID and databaseServiceRef are specified"))
		})
	})

	Context("after trying to create DBaaSConnection with both database service reference and database service type", func() {
		It("should not allow creating the DBaaSConnection", func() {
			testDBaaSConnectionNoDatabaseService := &DBaaSConnection{
				ObjectMeta: metav1.ObjectMeta{
					Name:      connectionName,
					Namespace: testNamespace,
				},
				Spec: DBaaSConnectionSpec{
					InventoryRef: NamespacedName{
						Name:      inventoryName,
						Namespace: testNamespace,
					},
					DatabaseServiceRef: &NamespacedName{
						Name:      databaseServiceName,
						Namespace: testNamespace,
					},
					DatabaseServiceType: &databaseServiceType,
				},
			}
			err := k8sClient.Create(ctx, testDBaaSConnectionNoDatabaseService)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.databaseServiceRef: Invalid value: v1beta1.NamespacedName{Namespace:\"default\", Name:\"test-databaseService\"}: when using databaseServiceRef, databaseServiceType must not be specified"))
		})
	})

	Context("after creating DBaaSConnection without database service ID", func() {
		var testDBaaSConnectionNoDatabaseServiceID = &DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: DBaaSConnectionSpec{
				InventoryRef: NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				DatabaseServiceRef: &NamespacedName{
					Name:      databaseServiceName,
					Namespace: testNamespace,
				},
			},
		}

		BeforeEach(func() {
			By("creating DBaaSConnection")
			Expect(k8sClient.Create(ctx, testDBaaSConnectionNoDatabaseServiceID)).Should(Succeed())

			By("checking DBaaSConnection created")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionNoDatabaseServiceID), testDBaaSConnectionNoDatabaseServiceID); err != nil {
					return false
				}
				if len(testDBaaSConnectionNoDatabaseServiceID.Spec.DatabaseServiceID) > 0 {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			By("deleting DBaaSConnection")
			Expect(k8sClient.Delete(ctx, testDBaaSConnectionNoDatabaseServiceID)).Should(Succeed())

			By("checking DBaaSConnection deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionNoDatabaseServiceID), &DBaaSConnection{})
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		It("should not allow setting database service ID", func() {
			By("updating databaseServiceID twice")
			testDBaaSConnectionNoDatabaseServiceID.Spec.DatabaseServiceID = "updated-databaseServiceID"
			err := k8sClient.Update(ctx, testDBaaSConnectionNoDatabaseServiceID)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.databaseServiceID: Invalid value: \"updated-databaseServiceID\": databaseServiceID is immutable"))
		})
	})

	Context("after creating DBaaSConnection without database service reference", func() {
		var testDBaaSConnectionNoDatabaseServiceRef = &DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: DBaaSConnectionSpec{
				InventoryRef: NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				DatabaseServiceID: databaseServiceID,
			},
		}

		BeforeEach(func() {
			By("creating DBaaSConnection")
			Expect(k8sClient.Create(ctx, testDBaaSConnectionNoDatabaseServiceRef)).Should(Succeed())

			By("checking DBaaSConnection created")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionNoDatabaseServiceRef), testDBaaSConnectionNoDatabaseServiceRef); err != nil {
					return false
				}
				if testDBaaSConnectionNoDatabaseServiceRef.Spec.DatabaseServiceRef != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			By("deleting DBaaSConnection")
			Expect(k8sClient.Delete(ctx, testDBaaSConnectionNoDatabaseServiceRef)).Should(Succeed())

			By("checking DBaaSConnection deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionNoDatabaseServiceRef), &DBaaSConnection{})
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		It("should not allow setting database service reference", func() {
			testDBaaSConnectionNoDatabaseServiceRef.Spec.DatabaseServiceRef = &NamespacedName{
				Name:      databaseServiceName,
				Namespace: testNamespace,
			}
			err := k8sClient.Update(ctx, testDBaaSConnectionNoDatabaseServiceRef)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.databaseServiceRef: Invalid value: v1beta1.NamespacedName{Namespace:\"default\", Name:\"test-databaseService\"}: " +
				"databaseServiceRef is immutable"))
		})
	})

	Context("after creating DBaaSConnection with database service type", func() {
		var testDBaaSConnectionWithDatabaseServiceType = &DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: DBaaSConnectionSpec{
				InventoryRef: NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				DatabaseServiceID:   databaseServiceID,
				DatabaseServiceType: &databaseServiceType,
			},
		}

		BeforeEach(func() {
			By("creating DBaaSConnection")
			Expect(k8sClient.Create(ctx, testDBaaSConnectionWithDatabaseServiceType)).Should(Succeed())

			By("checking DBaaSConnection created")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionWithDatabaseServiceType), testDBaaSConnectionWithDatabaseServiceType); err != nil {
					return false
				}
				if testDBaaSConnectionWithDatabaseServiceType.Spec.DatabaseServiceType == nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			By("deleting DBaaSConnection")
			Expect(k8sClient.Delete(ctx, testDBaaSConnectionWithDatabaseServiceType)).Should(Succeed())

			By("checking DBaaSConnection deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionWithDatabaseServiceType), &DBaaSConnection{})
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		It("should not allow unsetting database service type", func() {
			testDBaaSConnectionWithDatabaseServiceType.Spec.DatabaseServiceType = nil
			err := k8sClient.Update(ctx, testDBaaSConnectionWithDatabaseServiceType)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.databaseServiceType: Invalid value: \"null\": databaseServiceType is immutable"))
		})
	})
})
