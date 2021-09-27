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
	"context"
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

const (
	testProviderName   = "mongodb-atlas"
	testInventoryKind  = "MongoDBAtlasInventory"
	testConnectionKind = "MongoDBAtlasConnection"
)

var defaultProvider = &v1alpha1.DBaaSProvider{
	ObjectMeta: metav1.ObjectMeta{
		Name: testProviderName,
	},
	Spec: v1alpha1.DBaaSProviderSpec{
		Provider: v1alpha1.DatabaseProvider{
			Name: testProviderName,
		},
		InventoryKind:    testInventoryKind,
		ConnectionKind:   testConnectionKind,
		CredentialFields: []v1alpha1.CredentialField{},
	},
}

var defaultTenant = getDefaultTenant(testNamespace)

func assertResourceCreationIfNotExists(object client.Object) func() {
	return func() {
		By("checking the resource exists")
		if err := dRec.Get(ctx, client.ObjectKeyFromObject(object), object); err != nil {
			if errors.IsNotFound(err) {
				assertResourceCreation(object)()
			} else {
				Fail(err.Error())
			}
		}
	}
}

func assertResourceCreation(object client.Object) func() {
	return func() {
		By("creating resource")
		object.SetResourceVersion("")
		Expect(dRec.Create(ctx, object)).Should(Succeed())

		By("checking the resource created")
		Eventually(func() bool {
			if err := dRec.Get(ctx, client.ObjectKeyFromObject(object), object); err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
	}
}

func assertResourceDeletion(object client.Object) func() {
	return func() {
		By("deleting resource")
		Expect(dRec.Delete(ctx, object)).Should(Succeed())

		By("checking the resource deleted")
		Eventually(func() bool {
			err := dRec.Get(ctx, client.ObjectKeyFromObject(object), object)
			if err != nil && errors.IsNotFound(err) {
				return true
			}
			return false
		}, timeout, interval).Should(BeTrue())
	}
}

func assertProviderResourceCreated(object client.Object, providerResourceKind string, DBaaSResourceSpec interface{}) func() {
	return func() {
		By("checking a provider resource created")
		objectKey := client.ObjectKeyFromObject(object)
		providerResource := &unstructured.Unstructured{}
		providerResource.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    providerResourceKind,
		})
		Eventually(func() bool {
			if err := dRec.Get(ctx, objectKey, providerResource); err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())

		By("checking the provider resource spec is correct")
		bytes, err := providerResource.MarshalJSON()
		Expect(err).NotTo(HaveOccurred())
		switch v := object.(type) {
		case *v1alpha1.DBaaSInventory:
			providerInventory := &v1alpha1.DBaaSProviderInventory{}
			err := json.Unmarshal(bytes, providerInventory)
			Expect(err).NotTo(HaveOccurred())
			Expect(&providerInventory.Spec).Should(Equal(DBaaSResourceSpec))
			Expect(len(providerInventory.GetOwnerReferences())).Should(Equal(1))
			Expect(providerInventory.GetOwnerReferences()[0].Name).Should(Equal(object.GetName()))
		case *v1alpha1.DBaaSConnection:
			providerConnection := &v1alpha1.DBaaSProviderConnection{}
			err := json.Unmarshal(bytes, providerConnection)
			Expect(err).NotTo(HaveOccurred())
			Expect(&providerConnection.Spec).Should(Equal(DBaaSResourceSpec))
			Expect(len(providerConnection.GetOwnerReferences())).Should(Equal(1))
			Expect(providerConnection.GetOwnerReferences()[0].Name).Should(Equal(object.GetName()))
		default:
			_ = v.GetName() // to avoid syntax error
			Fail("invalid test object")
		}
	}
}

func assertDBaaSResourceStatusUpdated(object client.Object, providerResourceKind string, providerResourceStatus interface{}) func() {
	return func() {
		By("checking the DBaaS resource status has no conditions")
		objectKey := client.ObjectKeyFromObject(object)
		Consistently(func() (int, error) {
			err := dRec.Get(ctx, objectKey, object)
			if err != nil {
				return -1, err
			}
			switch v := object.(type) {
			case *v1alpha1.DBaaSInventory:
				return len(v.Status.Conditions), nil
			case *v1alpha1.DBaaSConnection:
				return len(v.Status.Conditions), nil
			default:
				Fail("invalid test object")
				return -1, err
			}
		}, duration, interval).Should(Equal(0))

		By("getting the provider resource")
		providerResource := &unstructured.Unstructured{}
		providerResource.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    providerResourceKind,
		})
		Eventually(func() bool {
			err := dRec.Get(ctx, objectKey, providerResource)
			if err != nil {
				if errors.IsNotFound(err) {
					return false
				}
				Expect(err).NotTo(HaveOccurred())
			}

			By("updating the provider resource status")
			providerResource.UnstructuredContent()["status"] = providerResourceStatus

			err = dRec.Status().Update(ctx, providerResource)
			if err != nil {
				if errors.IsConflict(err) {
					return false
				}
				Expect(err).NotTo(HaveOccurred())
			}
			return true
		}, timeout, interval).Should(BeTrue())

		By("checking the DBaaS resource status updated")
		Eventually(func() (int, error) {
			err := dRec.Get(ctx, objectKey, object)
			if err != nil {
				return -1, err
			}
			switch v := object.(type) {
			case *v1alpha1.DBaaSInventory:
				return len(v.Status.Conditions), nil
			case *v1alpha1.DBaaSConnection:
				return len(v.Status.Conditions), nil
			default:
				Fail("invalid test object")
				return -1, err
			}
		}, timeout, interval).Should(Equal(1))
		switch v := object.(type) {
		case *v1alpha1.DBaaSInventory:
			Expect(&v.Status).Should(Equal(providerResourceStatus))
		case *v1alpha1.DBaaSConnection:
			Expect(&v.Status).Should(Equal(providerResourceStatus))
		default:
			Fail("invalid test object")
		}
	}
}

func assertProviderResourceSpecUpdated(object client.Object, providerResourceKind string, DBaaSResourceSpec interface{}) func() {
	return func() {
		By("updating the DBaaS resource spec")
		objectKey := client.ObjectKeyFromObject(object)
		Eventually(func() bool {
			err := dRec.Get(ctx, objectKey, object)
			Expect(err).NotTo(HaveOccurred())

			switch v := object.(type) {
			case *v1alpha1.DBaaSInventory:
				v.Spec.DBaaSInventorySpec = *DBaaSResourceSpec.(*v1alpha1.DBaaSInventorySpec)
			case *v1alpha1.DBaaSConnection:
				v.Spec = *DBaaSResourceSpec.(*v1alpha1.DBaaSConnectionSpec)
			default:
				Fail("invalid test object")
			}

			err = dRec.Update(ctx, object)
			if err != nil {
				if errors.IsConflict(err) {
					return false
				}
				Expect(err).NotTo(HaveOccurred())
			}
			return true
		}, timeout, interval).Should(BeTrue())

		By("checking the provider resource status updated")
		providerResource := &unstructured.Unstructured{}
		providerResource.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    providerResourceKind,
		})
		Eventually(func() bool {
			err := dRec.Get(ctx, objectKey, providerResource)
			if err != nil {
				return false
			}

			bytes, err := providerResource.MarshalJSON()
			Expect(err).NotTo(HaveOccurred())
			switch v := object.(type) {
			case *v1alpha1.DBaaSInventory:
				providerInventory := &v1alpha1.DBaaSProviderInventory{}
				err := json.Unmarshal(bytes, providerInventory)
				Expect(err).NotTo(HaveOccurred())
				return reflect.DeepEqual(&providerInventory.Spec, DBaaSResourceSpec)
			case *v1alpha1.DBaaSConnection:
				providerConnection := &v1alpha1.DBaaSProviderConnection{}
				err := json.Unmarshal(bytes, providerConnection)
				Expect(err).NotTo(HaveOccurred())
				return reflect.DeepEqual(&providerConnection.Spec, DBaaSResourceSpec)
			default:
				_ = v.GetName() // to avoid syntax error
				Fail("invalid test object")
				return false
			}
		}, timeout, interval).Should(BeTrue())
	}
}

type SpyController struct {
	controller.Controller
	source chan client.Object
	owner  chan runtime.Object
}

func (c *SpyController) Watch(src source.Source, evthdler handler.EventHandler, prct ...predicate.Predicate) error {
	c.reset()

	switch s := src.(type) {
	case *source.Kind:
		c.source <- s.Type
	default:
		Fail("unexpected source type")
	}

	switch h := evthdler.(type) {
	case *handler.EnqueueRequestForOwner:
		c.owner <- h.OwnerType
	default:
		Fail("unexpected handler type")
	}

	if c.Controller != nil {
		return c.Controller.Watch(src, evthdler, prct...)
	} else {
		return nil
	}
}

func (c *SpyController) Start(ctx context.Context) error {
	if c.Controller != nil {
		return c.Controller.Start(ctx)
	} else {
		return nil
	}
}

func (c *SpyController) GetLogger() logr.Logger {
	if c.Controller != nil {
		return c.Controller.GetLogger()
	} else {
		return nil
	}
}

func (c *SpyController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if c.Controller != nil {
		return c.Controller.Reconcile(ctx, req)
	} else {
		return reconcile.Result{}, nil
	}
}

func (c *SpyController) reset() {
	if len(c.source) > 0 {
		<-c.source
	}
	if len(c.owner) > 0 {
		<-c.owner
	}
}

func newSpyController(ctrl controller.Controller) *SpyController {
	return &SpyController{
		Controller: ctrl,
		source:     make(chan client.Object, 1),
		owner:      make(chan runtime.Object, 1),
	}
}
