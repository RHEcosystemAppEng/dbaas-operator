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

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var inventoryName = "test-inventory"
var connectionName = "test-connection"
var instanceID = "test-instanceID"
var instanceName = "test-instance"
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
		InstanceID: instanceID,
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
			Entry("not allow updating instanceID",
				func(spec *DBaaSConnectionSpec) {
					spec.InstanceID = "updated-instanceID"
				},
				"admission webhook \"vdbaasconnection.kb.io\" denied the request: "+
					"spec.instanceID: Invalid value: \"updated-instanceID\": instanceID is immutable"),
			Entry("not allow updating inventoryRef",
				func(spec *DBaaSConnectionSpec) {
					spec.InventoryRef.Name = "updated-inventory"
				},
				"admission webhook \"vdbaasconnection.kb.io\" denied the request: "+
					"spec.inventoryRef: Invalid value: v1alpha1.NamespacedName{Namespace:\"default\", Name:\"updated-inventory\"}: "+
					"inventoryRef is immutable"),
			Entry("not allow updating instanceRef",
				func(spec *DBaaSConnectionSpec) {
					spec.InstanceRef = &NamespacedName{
						Name:      "updated-instance",
						Namespace: testNamespace,
					}
				},
				"admission webhook \"vdbaasconnection.kb.io\" denied the request: "+
					"spec.instanceRef: Invalid value: v1alpha1.NamespacedName{Namespace:\"default\", Name:\"updated-instance\"}: "+
					"instanceRef is immutable"),
		)
	})

	Context("after trying to create DBaaSConnection without instance info", func() {
		It("should not allow creating the DBaaSConnection", func() {
			testDBaaSConnectionNoInstance := &DBaaSConnection{
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
			err := k8sClient.Create(ctx, testDBaaSConnectionNoInstance)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.instanceID: Invalid value: \"\": either instanceID or instanceRef must be specified"))
		})
	})

	Context("after trying to create DBaaSConnection with both instance ID and instance reference", func() {
		It("should not allow creating the DBaaSConnection", func() {
			testDBaaSConnectionNoInstance := &DBaaSConnection{
				ObjectMeta: metav1.ObjectMeta{
					Name:      connectionName,
					Namespace: testNamespace,
				},
				Spec: DBaaSConnectionSpec{
					InventoryRef: NamespacedName{
						Name:      inventoryName,
						Namespace: testNamespace,
					},
					InstanceID: instanceID,
					InstanceRef: &NamespacedName{
						Name:      instanceName,
						Namespace: testNamespace,
					},
				},
			}
			err := k8sClient.Create(ctx, testDBaaSConnectionNoInstance)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.instanceID: Invalid value: \"test-instanceID\": both instanceID and instanceRef are specified"))
		})
	})

	Context("after creating DBaaSConnection without instance ID", func() {
		var testDBaaSConnectionNoInstanceID = &DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: DBaaSConnectionSpec{
				InventoryRef: NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				InstanceRef: &NamespacedName{
					Name:      instanceName,
					Namespace: testNamespace,
				},
			},
		}

		BeforeEach(func() {
			By("creating DBaaSConnection")
			Expect(k8sClient.Create(ctx, testDBaaSConnectionNoInstanceID)).Should(Succeed())

			By("checking DBaaSConnection created")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionNoInstanceID), testDBaaSConnectionNoInstanceID); err != nil {
					return false
				}
				if len(testDBaaSConnectionNoInstanceID.Spec.InstanceID) > 0 {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			By("deleting DBaaSConnection")
			Expect(k8sClient.Delete(ctx, testDBaaSConnectionNoInstanceID)).Should(Succeed())

			By("checking DBaaSConnection deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionNoInstanceID), &DBaaSConnection{})
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		It("should not allow setting instance ID", func() {
			By("updating instanceID twice")
			testDBaaSConnectionNoInstanceID.Spec.InstanceID = "updated-instanceID"
			err := k8sClient.Update(ctx, testDBaaSConnectionNoInstanceID)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.instanceID: Invalid value: \"updated-instanceID\": instanceID is immutable"))
		})
	})

	Context("after creating DBaaSConnection without instance reference", func() {
		var testDBaaSConnectionNoInstanceRef = &DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: DBaaSConnectionSpec{
				InventoryRef: NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				InstanceID: instanceID,
			},
		}

		BeforeEach(func() {
			By("creating DBaaSConnection")
			Expect(k8sClient.Create(ctx, testDBaaSConnectionNoInstanceRef)).Should(Succeed())

			By("checking DBaaSConnection created")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionNoInstanceRef), testDBaaSConnectionNoInstanceRef); err != nil {
					return false
				}
				if testDBaaSConnectionNoInstanceRef.Spec.InstanceRef != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			By("deleting DBaaSConnection")
			Expect(k8sClient.Delete(ctx, testDBaaSConnectionNoInstanceRef)).Should(Succeed())

			By("checking DBaaSConnection deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSConnectionNoInstanceRef), &DBaaSConnection{})
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		It("should not allow setting instance ID", func() {
			testDBaaSConnectionNoInstanceRef.Spec.InstanceRef = &NamespacedName{
				Name:      instanceName,
				Namespace: testNamespace,
			}
			err := k8sClient.Update(ctx, testDBaaSConnectionNoInstanceRef)
			Expect(err).Should(MatchError("admission webhook \"vdbaasconnection.kb.io\" denied the request: " +
				"spec.instanceRef: Invalid value: v1alpha1.NamespacedName{Namespace:\"default\", Name:\"test-instance\"}: " +
				"instanceRef is immutable"))
		})
	})
})
