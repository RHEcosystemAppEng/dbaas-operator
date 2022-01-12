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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

var _ = Describe("DBaaSProvider controller", func() {
	Describe("trigger reconcile", func() {
		createdInventoryKind := "createdInventoryKind"
		createdConnectionKind := "createdConnectionKind"

		provider := &v1alpha1.DBaaSProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-create-update-provider",
				Namespace: testNamespace,
			},
			Spec: v1alpha1.DBaaSProviderSpec{
				Provider: v1alpha1.DatabaseProvider{
					Name: "test-create-update-provider",
				},
				InventoryKind:    createdInventoryKind,
				ConnectionKind:   createdConnectionKind,
				CredentialFields: []v1alpha1.CredentialField{},
			},
		}

		iSrc := &unstructured.Unstructured{}
		iSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    createdInventoryKind,
		})
		iOwner := &v1alpha1.DBaaSInventory{}
		cSrc := &unstructured.Unstructured{}
		cSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    createdConnectionKind,
		})
		cOwner := &v1alpha1.DBaaSConnection{}

		BeforeEach(func() { assertNotWatched(iSrc, iOwner, cSrc, cOwner) })
		BeforeEach(assertResourceCreation(provider))
		AfterEach(assertResourceDeletion(provider))
		AfterEach(func() { reset(iSrc, iOwner, cSrc, cOwner) })

		Context("after creating a DBaaSProvider", func() {
			It("should make DBaaSInventory and DBaaSConnection watch the provider inventory and connection", func() {
				assertWatched(iSrc, iOwner, cSrc, cOwner)
			})
		})

		Context("after updating a DBaaSProvider", func() {
			updatedInventoryKind := "updatedInventoryKind"
			updatedConnectionKind := "updatedConnectionKind"

			uiSrc := &unstructured.Unstructured{}
			uiSrc.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   v1alpha1.GroupVersion.Group,
				Version: v1alpha1.GroupVersion.Version,
				Kind:    updatedInventoryKind,
			})
			ucSrc := &unstructured.Unstructured{}
			ucSrc.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   v1alpha1.GroupVersion.Group,
				Version: v1alpha1.GroupVersion.Version,
				Kind:    updatedConnectionKind,
			})

			BeforeEach(func() { assertNotWatched(uiSrc, iOwner, ucSrc, cOwner) })
			AfterEach(func() { reset(uiSrc, iOwner, ucSrc, cOwner) })

			It("should make DBaaSInventory and DBaaSConnection watch the provider inventory and connection", func() {
				updatedProvider := &v1alpha1.DBaaSProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-create-update-provider",
						Namespace: testNamespace,
					},
				}
				err := dRec.Get(ctx, client.ObjectKeyFromObject(updatedProvider), updatedProvider)
				Expect(err).NotTo(HaveOccurred())

				updatedProvider.Spec.InventoryKind = updatedInventoryKind
				updatedProvider.Spec.ConnectionKind = updatedConnectionKind
				Expect(dRec.Update(ctx, updatedProvider)).Should(Succeed())
				Eventually(func() v1alpha1.DBaaSProviderSpec {
					pProvider := &v1alpha1.DBaaSProvider{}
					err := dRec.Get(ctx, client.ObjectKeyFromObject(updatedProvider), pProvider)
					if err != nil {
						return v1alpha1.DBaaSProviderSpec{}
					}
					return pProvider.Spec
				}, timeout, interval).Should(Equal(updatedProvider.Spec))

				assertWatched(uiSrc, iOwner, ucSrc, cOwner)
			})
		})
	})

	Describe("not trigger reconcile", func() {
		Context("after deleting a DBaaSProvider", func() {
			It("should not watch the provider inventory and connection", func() {
				deletedInventoryKind := "deletedInventoryKind"
				deletedConnectionKind := "deletedConnectionKind"

				provider := &v1alpha1.DBaaSProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-delete-provider",
						Namespace: testNamespace,
					},
					Spec: v1alpha1.DBaaSProviderSpec{
						Provider: v1alpha1.DatabaseProvider{
							Name: "test-delete-provider",
						},
						InventoryKind:    deletedInventoryKind,
						ConnectionKind:   deletedConnectionKind,
						CredentialFields: []v1alpha1.CredentialField{},
					},
				}

				iSrc := &unstructured.Unstructured{}
				iSrc.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   v1alpha1.GroupVersion.Group,
					Version: v1alpha1.GroupVersion.Version,
					Kind:    deletedInventoryKind,
				})
				iOwner := &v1alpha1.DBaaSInventory{}
				cSrc := &unstructured.Unstructured{}
				cSrc.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   v1alpha1.GroupVersion.Group,
					Version: v1alpha1.GroupVersion.Version,
					Kind:    deletedConnectionKind,
				})
				cOwner := &v1alpha1.DBaaSConnection{}

				assertNotWatched(iSrc, iOwner, cSrc, cOwner)
				assertResourceCreation(provider)()
				assertWatched(iSrc, iOwner, cSrc, cOwner)

				reset(iSrc, iOwner, cSrc, cOwner)

				assertResourceDeletion(provider)()
				assertNotWatched(iSrc, iOwner, cSrc, cOwner)
			})
		})
	})
})

func assertWatched(iSrc client.Object, iOwner runtime.Object, cSrc client.Object, cOwner runtime.Object) {
	Eventually(func() bool {
		return iCtrl.watched(&watchable{
			source: iSrc,
			owner:  iOwner,
		})
	}, timeout, interval).Should(BeTrue())

	Eventually(func() bool {
		return cCtrl.watched(&watchable{
			source: cSrc,
			owner:  cOwner,
		})
	}, timeout, interval).Should(BeTrue())
}

func assertNotWatched(iSrc client.Object, iOwner runtime.Object, cSrc client.Object, cOwner runtime.Object) {
	Consistently(func() bool {
		return !iCtrl.watched(&watchable{
			source: iSrc,
			owner:  iOwner,
		})
	}, duration, interval).Should(BeTrue())

	Consistently(func() bool {
		return !cCtrl.watched(&watchable{
			source: cSrc,
			owner:  cOwner,
		})
	}, duration, interval).Should(BeTrue())
}

func reset(iSrc client.Object, iOwner runtime.Object, cSrc client.Object, cOwner runtime.Object) {
	iCtrl.delete(&watchable{
		source: iSrc,
		owner:  iOwner,
	})
	cCtrl.delete(&watchable{
		source: cSrc,
		owner:  cOwner,
	})
}
