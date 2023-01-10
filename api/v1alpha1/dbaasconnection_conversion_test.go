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
	. "github.com/onsi/gomega"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("DBaaSConnection Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := DBaaSConnection{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testName,
					Namespace: testNamespace,
				},
				Spec: DBaaSConnectionSpec{
					InstanceID: "xxx32xx",
					InventoryRef: NamespacedName{
						Name: "test",
					},
					InstanceRef: &NamespacedName{
						Name: "test",
					},
				},
				Status: DBaaSConnectionStatus{
					Conditions: []metav1.Condition{
						{
							Type: v1beta1.DBaaSConnectionProviderSyncType,
						},
					},
					CredentialsRef:    &v1.LocalObjectReference{Name: "test"},
					ConnectionInfoRef: &v1.LocalObjectReference{Name: "test"},
				},
			}
			intermediate := v1beta1.DBaaSConnection{}
			dst := DBaaSConnection{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())
			Expect(dst).To(Equal(src))
		})
	})
})
