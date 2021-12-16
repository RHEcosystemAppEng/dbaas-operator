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
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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

func assertInventoryCreationWithProviderStatus(object client.Object, inventroyDBaaSStatus metav1.ConditionStatus, providerResourceKind string, DBaaSResourceSpec interface{}) func() {
	return func() {
		assertResourceCreation(object)()
		err := dRec.Get(ctx, client.ObjectKeyFromObject(object), object)
		Expect(err).Should(Succeed())
		assertDBaaSResourceProviderStatusUpdated(object, inventroyDBaaSStatus, providerResourceKind, DBaaSResourceSpec)()
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

func assertDBaaSResourceStatusUpdated(object client.Object, status metav1.ConditionStatus, reason string) func() {
	return func() {
		By("checking the DBaaS resource status")
		objectKey := client.ObjectKeyFromObject(object)

		Eventually(func() (bool, error) {
			err := dRec.Get(ctx, objectKey, object)
			if err != nil {
				return false, err
			}
			switch v := object.(type) {
			case *v1alpha1.DBaaSInventory:
				dbaasConds, _ := splitStatusConditions(v.Status.Conditions, v1alpha1.DBaaSInventoryReadyType)
				return len(dbaasConds) > 0 && dbaasConds[0].Status == status && dbaasConds[0].Reason == reason, nil
			case *v1alpha1.DBaaSConnection:
				dbaasConds, _ := splitStatusConditions(v.Status.Conditions, v1alpha1.DBaaSConnectionReadyType)
				return len(dbaasConds) > 0 && dbaasConds[0].Status == status && dbaasConds[0].Reason == reason, nil
			default:
				Fail("invalid test object")
				return false, err
			}
		}, duration, interval).Should(BeTrue())
	}
}

func assertDBaaSResourceProviderStatusUpdated(object client.Object, inventoryDBaaSStatus metav1.ConditionStatus, providerResourceKind string, providerResourceStatus interface{}) func() {
	return func() {
		By("retrieving current DBaaS resource")
		objectKey := client.ObjectKeyFromObject(object)
		Consistently(func() (int, error) {
			err := dRec.Get(ctx, objectKey, object)
			if err != nil {
				return -1, err
			}
			return 0, nil
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

		By("checking the DBaaS resource provider status updated")
		Eventually(func() (int, error) {
			err := dRec.Get(ctx, objectKey, object)
			if err != nil {
				return -1, err
			}
			switch v := object.(type) {
			case *v1alpha1.DBaaSInventory:
				_, conds := splitStatusConditions(v.Status.Conditions, v1alpha1.DBaaSInventoryReadyType)
				return len(conds), nil
			case *v1alpha1.DBaaSConnection:
				assertInventoryDBaaSStatus(v.Spec.InventoryRef.Name, v.Spec.InventoryRef.Namespace, inventoryDBaaSStatus)()
				_, conds := splitStatusConditions(v.Status.Conditions, v1alpha1.DBaaSConnectionReadyType)
				return len(conds), nil
			default:
				Fail("invalid test object")
				return -1, err
			}
		}, timeout, interval).Should(Equal(1))
		switch v := object.(type) {
		case *v1alpha1.DBaaSInventory:
			assertInventoryStatus(v, v1alpha1.DBaaSInventoryReadyType, inventoryDBaaSStatus, providerResourceStatus)()
		case *v1alpha1.DBaaSConnection:
			assertConnectionStatus(v, v1alpha1.DBaaSConnectionReadyType, providerResourceStatus)()
		default:
			Fail("invalid test object")
		}
	}
}

func splitStatusConditions(conds []metav1.Condition, condType string) (dbaasCond []metav1.Condition, providerCond []metav1.Condition) {
	for _, v := range conds {
		if v.Type != condType { //skip the DBaaS operator specific condition
			providerCond = append(providerCond, v)
		} else {
			dbaasCond = append(dbaasCond, v)
		}
	}
	return
}

func assertInventoryDBaaSStatus(name, namespace string, dbaasStatus metav1.ConditionStatus) func() {
	return func() {
		updatedInv := &v1alpha1.DBaaSInventory{}
		Eventually(func() (int, error) {
			err := dRec.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, updatedInv)
			if err != nil {
				return -1, err
			}
			cond := apimeta.FindStatusCondition(updatedInv.Status.Conditions, v1alpha1.DBaaSInventoryReadyType)
			if cond != nil && cond.Status == dbaasStatus {
				return 0, nil
			}
			return 0, nil
		}, timeout, interval).Should(Equal(0))
	}
}

func assertConnectionDBaaSStatus(name, namespace string, dbaasStatus metav1.ConditionStatus) func() {
	return func() {
		updatedConn := &v1alpha1.DBaaSConnection{}
		Eventually(func() (int, error) {
			err := dRec.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, updatedConn)
			if err != nil {
				return -1, err
			}
			cond := apimeta.FindStatusCondition(updatedConn.Status.Conditions, v1alpha1.DBaaSConnectionReadyType)
			if cond != nil && cond.Status == dbaasStatus {
				return 0, nil
			}
			return 0, nil
		}, timeout, interval).Should(Equal(0))
	}
}

func assertInventoryStatus(inv *v1alpha1.DBaaSInventory, condType string, dbaasStatus metav1.ConditionStatus, providerResourceStatus interface{}) func() {
	return func() {
		status := inv.Status.DeepCopy()
		dbaasConds, providerConds := splitStatusConditions(status.Conditions, condType)
		Expect(len(dbaasConds)).Should(Equal(1))
		Expect(dbaasConds[0].Type).Should(Equal(condType))
		Expect(dbaasConds[0].Status).Should(Equal(dbaasStatus))
		status.Conditions = providerConds
		Expect(status).Should(Equal(providerResourceStatus))
	}
}

func assertConnectionStatus(conn *v1alpha1.DBaaSConnection, condType string, providerResourceStatus interface{}) func() {
	return func() {
		assertConnectionDBaaSStatus(conn.Name, conn.Namespace, metav1.ConditionTrue)()
		status := conn.Status.DeepCopy()
		_, providerConds := splitStatusConditions(status.Conditions, condType)
		status.Conditions = providerConds
		Expect(status).Should(Equal(providerResourceStatus))
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
