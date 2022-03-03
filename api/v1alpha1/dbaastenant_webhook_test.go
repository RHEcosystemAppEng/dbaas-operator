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

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var tenantName = "test-tenant"
var namespaceName = "test-namespace"
var testDBaaSTenant = &DBaaSTenant{
	ObjectMeta: metav1.ObjectMeta{
		Name: tenantName,
	},
	Spec: DBaaSTenantSpec{
		InventoryNamespace: namespaceName,
	},
}

var _ = Describe("DBaaSTenant Webhook", func() {
	Context("after creating DBaaSTenant", func() {
		BeforeEach(func() {
			By("creating DBaaSTenant")
			Expect(k8sClient.Create(ctx, testDBaaSTenant)).Should(Succeed())

			By("checking DBaaSTenant created")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSTenant), &DBaaSTenant{}); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			By("deleting DBaaSTenant")
			Expect(k8sClient.Delete(ctx, testDBaaSTenant)).Should(Succeed())

			By("checking DBaaSTenant deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(testDBaaSTenant), &DBaaSTenant{})
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		Context("after creating DBaaSTenant of the same inventory namespace", func() {
			It("should not allow creating DBaaSTenant", func() {
				testTenant := &DBaaSTenant{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-tenant-1",
					},
					Spec: DBaaSTenantSpec{
						InventoryNamespace: namespaceName,
					},
				}
				By("creating DBaaSTenant")
				Expect(k8sClient.Create(ctx, testTenant)).Should(MatchError("admission webhook \"vdbaastenant.kb.io\" denied the request: " +
					"spec.inventoryNamespace: Invalid value: \"test-namespace\": the namespace test-namespace is already managed by tenant test-tenant, " +
					"it cannot be managed by another tenant"))
			})
		})
	})
})
