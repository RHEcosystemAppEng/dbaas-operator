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
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

var _ = Describe("DBaaSProvider controller", func() {
	Context("after creating and updating a DBaaSProvider", func() {
		createdInventoryKind := "DBaaSCreateInventory"
		createdConnectionKind := "DBaaSCreateConnection"
		createdInstanceKind := "DBaaSCreateInstance"

		provider := &v1beta1.DBaaSProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-create-update-provider",
				Namespace: testNamespace,
			},
			Spec: v1beta1.DBaaSProviderSpec{
				Provider: v1beta1.DatabaseProviderInfo{
					Name: "test-create-update-provider",
				},
				InventoryKind:                createdInventoryKind,
				ConnectionKind:               createdConnectionKind,
				InstanceKind:                 createdInstanceKind,
				CredentialFields:             []v1beta1.CredentialField{},
				AllowsFreeTrial:              false,
				ExternalProvisionURL:         "",
				ExternalProvisionDescription: "",
				InstanceParameterSpecs:       []v1beta1.InstanceParameterSpec{},
			},
		}

		iSrc := &unstructured.Unstructured{}
		iSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    createdInventoryKind,
		})
		iOwner := &v1beta1.DBaaSInventory{}
		cSrc := &unstructured.Unstructured{}
		cSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    createdConnectionKind,
		})
		cOwner := &v1beta1.DBaaSConnection{}
		inSrc := &unstructured.Unstructured{}
		inSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    createdInstanceKind,
		})
		inOwner := &v1beta1.DBaaSInstance{}

		BeforeEach(func() { assertNotWatched(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner) })
		BeforeEach(assertResourceCreation(provider))
		AfterEach(assertResourceDeletion(provider))
		AfterEach(func() { reset(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner) })

		updatedInventoryKind := "DBaaSUpdateInventory"
		updatedConnectionKind := "DBaaSUpdateConnection"
		updatedInstanceKind := "DBaaSUpdateInstance"

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
		uinSrc := &unstructured.Unstructured{}
		uinSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    updatedInstanceKind,
		})

		AfterEach(func() { reset(uiSrc, iOwner, ucSrc, cOwner, uinSrc, inOwner) })

		It("should make DBaaSInventory, DBaaSConnection and DBaaSInstance watch the provider inventory, connection and instance", func() {
			assertWatched(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner)
			assertNotWatched(uiSrc, iOwner, ucSrc, cOwner, uinSrc, inOwner)

			updatedProvider := &v1beta1.DBaaSProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-create-update-provider",
					Namespace: testNamespace,
				},
			}
			err := dRec.Get(ctx, client.ObjectKeyFromObject(updatedProvider), updatedProvider)
			Expect(err).NotTo(HaveOccurred())

			updatedProvider.Spec.InventoryKind = updatedInventoryKind
			updatedProvider.Spec.ConnectionKind = updatedConnectionKind
			updatedProvider.Spec.InstanceKind = updatedInstanceKind
			Expect(dRec.Update(ctx, updatedProvider)).Should(Succeed())
			Eventually(func() v1beta1.DBaaSProviderSpec {
				pProvider := &v1beta1.DBaaSProvider{}
				err := dRec.Get(ctx, client.ObjectKeyFromObject(updatedProvider), pProvider)
				if err != nil {
					return v1beta1.DBaaSProviderSpec{}
				}
				return pProvider.Spec
			}, timeout).Should(Equal(updatedProvider.Spec))

			assertWatched(uiSrc, iOwner, ucSrc, cOwner, uinSrc, inOwner)
		})
	})

	Context("after deleting a DBaaSProvider", func() {
		deletedInventoryKind := "DBaaSDeleteInventory"
		deletedConnectionKind := "DBaaSDeleteConnection"
		deletedInstanceKind := "DBaaSDeleteInstance"

		provider := &v1beta1.DBaaSProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-delete-provider",
				Namespace: testNamespace,
			},
			Spec: v1beta1.DBaaSProviderSpec{
				Provider: v1beta1.DatabaseProviderInfo{
					Name: "test-delete-provider",
				},
				InventoryKind:                deletedInventoryKind,
				ConnectionKind:               deletedConnectionKind,
				InstanceKind:                 deletedInstanceKind,
				CredentialFields:             []v1beta1.CredentialField{},
				AllowsFreeTrial:              false,
				ExternalProvisionURL:         "",
				ExternalProvisionDescription: "",
				InstanceParameterSpecs:       []v1beta1.InstanceParameterSpec{},
			},
		}

		iSrc := &unstructured.Unstructured{}
		iSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    deletedInventoryKind,
		})
		iOwner := &v1beta1.DBaaSInventory{}
		cSrc := &unstructured.Unstructured{}
		cSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    deletedConnectionKind,
		})
		cOwner := &v1beta1.DBaaSConnection{}
		inSrc := &unstructured.Unstructured{}
		inSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    deletedInstanceKind,
		})
		inOwner := &v1beta1.DBaaSInstance{}

		BeforeEach(func() { assertNotWatched(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner) })
		BeforeEach(assertResourceCreation(provider))
		AfterEach(func() { reset(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner) })

		It("should not watch the provider inventory, connection and instance", func() {
			assertWatched(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner)

			reset(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner)

			assertResourceDeletion(provider)()
			assertNotWatched(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner)
		})
	})
})

func assertWatched(iSrc client.Object, iOwner runtime.Object,
	cSrc client.Object, cOwner runtime.Object, inSrc client.Object, inOwner runtime.Object) {
	Eventually(func() bool {
		return iCtrl.watched(&watchable{
			source: iSrc,
			owner:  iOwner,
		})
	}, timeout).Should(BeTrue())

	Eventually(func() bool {
		return cCtrl.watched(&watchable{
			source: cSrc,
			owner:  cOwner,
		})
	}, timeout).Should(BeTrue())

	Eventually(func() bool {
		return inCtrl.watched(&watchable{
			source: inSrc,
			owner:  inOwner,
		})
	}, timeout).Should(BeTrue())
}

func assertNotWatched(iSrc client.Object, iOwner runtime.Object,
	cSrc client.Object, cOwner runtime.Object, inSrc client.Object, inOwner runtime.Object) {
	Consistently(func() bool {
		return !iCtrl.watched(&watchable{
			source: iSrc,
			owner:  iOwner,
		})
	}).Should(BeTrue())

	Consistently(func() bool {
		return !cCtrl.watched(&watchable{
			source: cSrc,
			owner:  cOwner,
		})
	}).Should(BeTrue())

	Consistently(func() bool {
		return !inCtrl.watched(&watchable{
			source: inSrc,
			owner:  inOwner,
		})
	}).Should(BeTrue())
}

func reset(iSrc client.Object, iOwner runtime.Object,
	cSrc client.Object, cOwner runtime.Object, inSrc client.Object, inOwner runtime.Object) {
	iCtrl.delete(&watchable{
		source: iSrc,
		owner:  iOwner,
	})
	cCtrl.delete(&watchable{
		source: cSrc,
		owner:  cOwner,
	})
	inCtrl.delete(&watchable{
		source: inSrc,
		owner:  inOwner,
	})
}
