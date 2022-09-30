/*
Copyright 2021, Red Hat.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	testDBaaSPolicy = DBaaSPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpolicy",
			Namespace: testNamespace,
		},
	}
)

var _ = Describe("DBaaSPolicy Webhook", func() {
	Context("creation fails",
		func() {
			It("missing required values field", func() {
				inv := testDBaaSPolicy.DeepCopy()
				inv.SetResourceVersion("")
				inv.Spec = DBaaSPolicySpec{
					DBaaSInventoryPolicy: DBaaSInventoryPolicy{
						ConnectionNsSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "test", Operator: "In"}},
						},
					},
				}
				err := k8sClient.Create(ctx, inv)
				Expect(err).Should(MatchError("admission webhook \"vdbaaspolicy.kb.io\" denied the request: values: Invalid value: []string(nil): for 'in', 'notin' operators, values set can't be empty"))
			})
		})
	Context("without optional fields", func() {
		It("should succeed without optional fields", func() {
			testDBaaSPolicy.SetResourceVersion("")
			err := k8sClient.Create(ctx, &testDBaaSPolicy)
			Expect(err).Should(BeNil())
		})
	})
	Context("update",
		func() {
			Context("nominal", func() {
				It("Update CR should succeed", func() {
					inv := testDBaaSPolicy.DeepCopy()
					Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(inv), inv)).Should(Succeed())
					inv.Spec = DBaaSPolicySpec{
						DBaaSInventoryPolicy: DBaaSInventoryPolicy{
							ConnectionNsSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "test", Operator: "In", Values: []string{"blah"}}},
							},
						},
					}
					err := k8sClient.Update(ctx, inv)
					Expect(err).Should(BeNil())
				})
			})
			Context("update fails", func() {
				It("update fails with missing required values field", func() {
					inv := testDBaaSPolicy.DeepCopy()
					Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(inv), inv)).Should(Succeed())
					inv.Spec.ConnectionNsSelector.MatchExpressions = []metav1.LabelSelectorRequirement{{Key: "test", Operator: "In"}}
					err := k8sClient.Update(ctx, inv)
					Expect(err).Should(MatchError("admission webhook \"vdbaaspolicy.kb.io\" denied the request: values: Invalid value: []string(nil): for 'in', 'notin' operators, values set can't be empty"))
				})
			})
		})
})
