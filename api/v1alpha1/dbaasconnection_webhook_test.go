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
		)
	})
})
