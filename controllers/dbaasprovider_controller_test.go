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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

var _ = Describe("DBaaSProvider controller", func() {
	BeforeEach(func() {
		iCtrl.reset()
		cCtrl.reset()
	})

	Describe("trigger reconcile", func() {
		provider := &v1alpha1.DBaaSProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-provider",
				Namespace: testNamespace,
			},
			Spec: v1alpha1.DBaaSProviderSpec{
				Provider: v1alpha1.DatabaseProvider{
					Name: "test-provider",
				},
				InventoryKind:    testInventoryKind,
				ConnectionKind:   testConnectionKind,
				CredentialFields: []v1alpha1.CredentialField{},
			},
		}
		BeforeEach(assertResourceCreation(provider))
		AfterEach(assertResourceDeletion(provider))

		Context("after creating a DBaaSProvider", func() {
			It("should make DBaaSInventory and DBaaSConnection watch the provider inventory and connection", func() {
				iSrc := &unstructured.Unstructured{}
				iSrc.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   v1alpha1.GroupVersion.Group,
					Version: v1alpha1.GroupVersion.Version,
					Kind:    testInventoryKind,
				})
				iOwner := &v1alpha1.DBaaSInventory{}
				cSrc := &unstructured.Unstructured{}
				cSrc.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   v1alpha1.GroupVersion.Group,
					Version: v1alpha1.GroupVersion.Version,
					Kind:    testConnectionKind,
				})
				cOwner := &v1alpha1.DBaaSConnection{}

				assertWatch(iSrc, iOwner, cSrc, cOwner)
			})
		})

		Context("after updating a DBaaSProvider", func() {
			It("should make DBaaSInventory and DBaaSConnection watch the provider inventory and connection", func() {
				iCtrl.reset()
				cCtrl.reset()

				updatedProvider := &v1alpha1.DBaaSProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-provider",
						Namespace: testNamespace,
					},
				}
				err := dRec.Get(ctx, client.ObjectKeyFromObject(updatedProvider), updatedProvider)
				Expect(err).NotTo(HaveOccurred())

				updatedProvider.Spec.InventoryKind = "CrunchyBridgeInventory"
				updatedProvider.Spec.ConnectionKind = "CrunchyBridgeConnection"
				Expect(dRec.Update(ctx, updatedProvider)).Should(Succeed())
				pProvider := &v1alpha1.DBaaSProvider{}
				Eventually(func() v1alpha1.DBaaSProviderSpec {
					err := dRec.Get(ctx, client.ObjectKeyFromObject(updatedProvider), pProvider)
					if err != nil {
						return v1alpha1.DBaaSProviderSpec{}
					}
					return pProvider.Spec
				}, timeout, interval).Should(Equal(updatedProvider.Spec))

				iSrc := &unstructured.Unstructured{}
				iSrc.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   v1alpha1.GroupVersion.Group,
					Version: v1alpha1.GroupVersion.Version,
					Kind:    "CrunchyBridgeInventory",
				})
				iOwner := &v1alpha1.DBaaSInventory{}
				cSrc := &unstructured.Unstructured{}
				cSrc.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   v1alpha1.GroupVersion.Group,
					Version: v1alpha1.GroupVersion.Version,
					Kind:    "CrunchyBridgeConnection",
				})
				cOwner := &v1alpha1.DBaaSConnection{}

				assertWatch(iSrc, iOwner, cSrc, cOwner)
			})
		})
	})

	Describe("not trigger reconcile", func() {
		Context("after deleting a DBaaSProvider", func() {
			It("should not watch the provider inventory and connection", func() {
				provider := &v1alpha1.DBaaSProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-provider",
						Namespace: testNamespace,
					},
					Spec: v1alpha1.DBaaSProviderSpec{
						Provider: v1alpha1.DatabaseProvider{
							Name: "test-provider",
						},
						InventoryKind:    testInventoryKind,
						ConnectionKind:   testConnectionKind,
						CredentialFields: []v1alpha1.CredentialField{},
					},
				}
				assertResourceCreation(provider)

				iCtrl.reset()
				cCtrl.reset()
				assertResourceDeletion(provider)

				assertNotWatch()
			})
		})
	})
})

func assertWatch(iSrc client.Object, iOwner runtime.Object, cSrc client.Object, cOwner runtime.Object) {
	select {
	case s := <-iCtrl.source:
		Expect(s).Should(Equal(iSrc))
	case <-time.After(timeout):
		Fail("failed to watch with the expected source")
	}
	select {
	case o := <-iCtrl.owner:
		Expect(o).Should(Equal(iOwner))
	case <-time.After(timeout):
		Fail("failed to watch with the expected owner")
	}
	select {
	case s := <-cCtrl.source:
		Expect(s).Should(Equal(cSrc))
	case <-time.After(timeout):
		Fail("failed to watch with the expected source")
	}
	select {
	case o := <-cCtrl.owner:
		Expect(o).Should(Equal(cOwner))
	case <-time.After(timeout):
		Fail("failed to watch with the expected owner")
	}
}

func assertNotWatch() {
	Expect(len(iCtrl.source)).Should(Equal(0))
	Expect(len(iCtrl.owner)).Should(Equal(0))

	Expect(len(cCtrl.source)).Should(Equal(0))
	Expect(len(cCtrl.owner)).Should(Equal(0))
}
