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

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DBaaSPolicy controller", func() {
	BeforeEach(assertResourceCreationIfNotExists(&defaultPolicy))
	BeforeEach(assertDBaaSResourceStatusUpdated(&defaultPolicy, metav1.ConditionTrue, v1beta1.Ready))

	Describe("reconcile", func() {
		Context("w/ status NotReady", func() {
			policy2 := getDefaultPolicy(testNamespace)
			policy2.Name = "test"
			BeforeEach(assertResourceCreationIfNotExists(&policy2))
			BeforeEach(assertDBaaSResourceStatusUpdated(&policy2, metav1.ConditionFalse, v1beta1.DBaaSPolicyNotReady))

			It("should return second policy with existing policy name in status message", func() {
				getPolicy := v1beta1.DBaaSPolicy{}
				err := dRec.Get(ctx, client.ObjectKeyFromObject(&policy2), &getPolicy)
				Expect(err).NotTo(HaveOccurred())
				Expect(getPolicy.Status.Conditions).Should(HaveLen(1))
				Expect(getPolicy.Status.Conditions[0].Message).Should(Equal(v1beta1.MsgPolicyNotReady + " - " + defaultPolicy.GetName()))
			})
		})
	})
})
