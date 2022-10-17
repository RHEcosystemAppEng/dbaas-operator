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
		// MongoDB provider types
		iSrc := &unstructured.Unstructured{}
		iSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    "MongoDBAtlasInventory",
		})
		iOwner := &v1alpha1.DBaaSInventory{}
		cSrc := &unstructured.Unstructured{}
		cSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    "MongoDBAtlasConnection",
		})
		cOwner := &v1alpha1.DBaaSConnection{}
		inSrc := &unstructured.Unstructured{}
		inSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    "MongoDBAtlasInstance",
		})
		inOwner := &v1alpha1.DBaaSInstance{}

		// Crunchy provider types
		ciSrc := &unstructured.Unstructured{}
		ciSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    "CrunchyBridgeInventory",
		})
		ciOwner := &v1alpha1.DBaaSInventory{}
		ccSrc := &unstructured.Unstructured{}
		ccSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    "CrunchyBridgeConnection",
		})
		ccOwner := &v1alpha1.DBaaSConnection{}
		cinSrc := &unstructured.Unstructured{}
		cinSrc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    "CrunchyBridgeInstance",
		})
		cinOwner := &v1alpha1.DBaaSInstance{}

		provider := &v1alpha1.DBaaSProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-create-update-provider",
				Namespace: testNamespace,
			},
			Spec: v1alpha1.DBaaSProviderSpec{
				Provider: v1alpha1.DatabaseProvider{
					Name: "test-create-update-provider",
				},
				InventoryKind:                iSrc.GetKind(),
				ConnectionKind:               cSrc.GetKind(),
				InstanceKind:                 inSrc.GetKind(),
				CredentialFields:             []v1alpha1.CredentialField{},
				AllowsFreeTrial:              false,
				ExternalProvisionURL:         "",
				ExternalProvisionDescription: "",
				InstanceParameterSpecs:       []v1alpha1.InstanceParameterSpec{},
			},
		}

		BeforeEach(assertResourceCreationIfNotExists(provider))
		Context("after creating a DBaaSProvider", func() {
			It("should make DBaaSInventory, DBaaSConnection and DBaaSInstance watch the provider inventory, connection and instance", func() {
				assertWatched(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner)
			})
		})
		Context("after updating a DBaaSProvider", func() {
			It("updating the provider should make DBaaSInventory, DBaaSConnection and DBaaSInstance watch the new provider inventory, connection and instance", func() {
				updatedProvider := provider.DeepCopy()
				err := dRec.Get(ctx, client.ObjectKeyFromObject(provider), updatedProvider)
				Expect(err).NotTo(HaveOccurred())

				updatedProvider.Spec.InventoryKind = ciSrc.GetKind()
				updatedProvider.Spec.ConnectionKind = ccSrc.GetKind()
				updatedProvider.Spec.InstanceKind = cinSrc.GetKind()
				Expect(dRec.Update(ctx, updatedProvider)).Should(Succeed())
				Eventually(func() v1alpha1.DBaaSProviderSpec {
					pProvider := &v1alpha1.DBaaSProvider{}
					err := dRec.Get(ctx, client.ObjectKeyFromObject(updatedProvider), pProvider)
					if err != nil {
						return v1alpha1.DBaaSProviderSpec{}
					}
					return pProvider.Spec
				}, timeout).Should(Equal(updatedProvider.Spec))

				assertWatched(ciSrc, ciOwner, ccSrc, ccOwner, cinSrc, cinOwner)
			})
		})
		Context("after deleting a DBaaSProvider", func() {
			It("should not watch the provider inventory, connection and instance", func() {
				assertWatched(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner)

				reset(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner)

				assertResourceDeletion(provider)()
				assertNotWatched(iSrc, iOwner, cSrc, cOwner, inSrc, inOwner)
			})
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
