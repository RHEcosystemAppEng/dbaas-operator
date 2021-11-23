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
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Create provider object", func() {
	It("should create the expected provider object", func() {
		object := &v1alpha1.DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-connection",
				Namespace: "test-namespace",
			},
		}
		result := dRec.createProviderObject(object, "test-kind")

		expected := &unstructured.Unstructured{}
		expected.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "dbaas.redhat.com",
			Version: "v1alpha1",
			Kind:    "test-kind",
		})
		expected.SetNamespace("test-namespace")
		expected.SetName("test-connection")
		Expect(result).Should(Equal(expected))
	})
})

var _ = Describe("Get DBaaSProvider", func() {
	provider := &v1alpha1.DBaaSProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-provider",
		},
		Spec: v1alpha1.DBaaSProviderSpec{
			Provider: v1alpha1.DatabaseProvider{
				Name: "test-provider",
			},
			InventoryKind:    "testInventoryKind",
			ConnectionKind:   "testConnectionKind",
			CredentialFields: []v1alpha1.CredentialField{},
		},
	}
	BeforeEach(assertResourceCreation(provider))
	AfterEach(assertResourceDeletion(provider))

	It("should get the expected DBaaSProvider", func() {
		provider.TypeMeta = metav1.TypeMeta{
			Kind:       "DBaaSProvider",
			APIVersion: v1alpha1.GroupVersion.Group + "/" + v1alpha1.GroupVersion.Version,
		}

		p, err := dRec.getDBaaSProvider("test-provider", ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).Should(Equal(provider))
	})
})

var _ = Describe("Get install Namespace", func() {
	It("should get expected install namespace", func() {
		ns, err := GetInstallNamespace()
		Expect(err).NotTo(HaveOccurred())
		Expect(ns).Should(Equal(testNamespace))
	})
})

var _ = Describe("Parse provider object", func() {
	connectionSpec := v1alpha1.DBaaSConnectionSpec{
		InventoryRef: v1alpha1.NamespacedName{
			Name:      "test-inventory",
			Namespace: "test-namespace",
		},
		InstanceID: "test-instance-id",
	}
	uConnection := &unstructured.Unstructured{}
	uConnection.SetUnstructuredContent(make(map[string]interface{}, 1))
	uConnection.UnstructuredContent()["spec"] = connectionSpec
	eConnection := &v1alpha1.DBaaSProviderConnection{
		Spec: connectionSpec,
	}

	inventorySpec := v1alpha1.DBaaSInventorySpec{
		CredentialsRef: &v1alpha1.NamespacedName{
			Name:      "test-credential-ref",
			Namespace: "test-namespace",
		},
	}
	uInventory := &unstructured.Unstructured{}
	uInventory.SetUnstructuredContent(make(map[string]interface{}, 1))
	uInventory.UnstructuredContent()["spec"] = inventorySpec
	eInventory := &v1alpha1.DBaaSProviderInventory{
		Spec: inventorySpec,
	}

	DescribeTable("should correctly parse the provider object",
		func(object interface{}, unstructured *unstructured.Unstructured, expected interface{}) {
			err := dRec.parseProviderObject(unstructured, object)
			Expect(err).NotTo(HaveOccurred())
			Expect(object).Should(Equal(expected))
		},
		Entry("parse DBaaSConnection", &v1alpha1.DBaaSProviderConnection{}, uConnection, eConnection),
		Entry("parse DBaaSInventory", &v1alpha1.DBaaSProviderInventory{}, uInventory, eInventory),
	)
})

var _ = Describe("Provider object MutateFn", func() {
	It("should create the expected MutateFn", func() {
		object := &v1alpha1.DBaaSConnection{}
		providerObject := &unstructured.Unstructured{}
		providerObject.SetUnstructuredContent(make(map[string]interface{}, 1))
		connectionSpec := &v1alpha1.DBaaSConnectionSpec{
			InventoryRef: v1alpha1.NamespacedName{
				Name:      "test-inventory",
				Namespace: "test-namespace",
			},
			InstanceID: "test-instance-id",
		}
		fn := dRec.providerObjectMutateFn(object, providerObject, connectionSpec)
		err := fn()
		Expect(err).NotTo(HaveOccurred())

		expected := &unstructured.Unstructured{}
		expected.SetUnstructuredContent(make(map[string]interface{}, 1))
		expected.UnstructuredContent()["spec"] = connectionSpec
		err = ctrl.SetControllerReference(object, expected, dRec.Scheme)
		Expect(err).NotTo(HaveOccurred())

		Expect(providerObject).Should(Equal(expected))
	})
})

var _ = Describe("Reconcile provider Object", func() {
	inventory := &unstructured.Unstructured{}
	inventory.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "dbaas.redhat.com",
		Version: "v1alpha1",
		Kind:    testInventoryKind,
	})
	inventory.SetNamespace(testNamespace)
	inventory.SetName("test-reconcile-provider")
	inventory.UnstructuredContent()["spec"] = map[string]interface{}{
		"credentialsRef": map[string]interface{}{
			"name":      "test-credential-ref",
			"namespace": "test-namespace",
		},
	}

	BeforeEach(assertResourceCreation(inventory))
	AfterEach(assertResourceDeletion(inventory))

	It("should update the provider object as expected", func() {
		spec := map[string]interface{}{
			"credentialsRef": map[string]interface{}{
				"name":      "updated-test-credential-ref",
				"namespace": "updated-test-namespace",
			},
		}
		r, err := dRec.reconcileProviderObject(inventory, func() error {
			inventory.UnstructuredContent()["spec"] = spec
			return nil
		}, ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(r).Should(Equal(controllerutil.OperationResultUpdated))

		updatedInventory := &unstructured.Unstructured{}
		updatedInventory.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "dbaas.redhat.com",
			Version: "v1alpha1",
			Kind:    testInventoryKind,
		})
		updatedInventory.SetNamespace(testNamespace)
		updatedInventory.SetName("test-reconcile-provider")
		Eventually(func() interface{} {
			err = dRec.Get(ctx, client.ObjectKeyFromObject(inventory), updatedInventory)
			Expect(err).NotTo(HaveOccurred())
			return updatedInventory.UnstructuredContent()["spec"]
		}, timeout, interval).Should(Equal(spec))
	})
})

var _ = Describe("Watch DBaaS provider Object", func() {
	It("should invoke controller watch with correctly input", func() {
		source := &unstructured.Unstructured{}
		source.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    "test-kind",
		})
		owner := &v1alpha1.DBaaSInventory{}
		spyController := newSpyController(nil)

		err := dRec.watchDBaaSProviderObject(spyController, owner, "test-kind")
		Expect(err).NotTo(HaveOccurred())
		select {
		case s := <-spyController.source:
			Expect(s).Should(Equal(source))
		case <-time.After(timeout):
			Fail("failed to watch with the expected source")
		}
		select {
		case o := <-spyController.owner:
			Expect(o).Should(Equal(owner))
		case <-time.After(timeout):
			Fail("failed to watch with the expected owner")
		}
	})
})

var _ = Describe("list tenants by inventory namespace", func() {
	Context("after creating DBaaSTenants", func() {
		ns := "test-namespace"
		tenant1 := getDefaultTenant(ns)
		tenant1.Name = "test-tenant-1"
		tenant2 := getDefaultTenant(ns)
		tenant2.Name = "test-tenant-2"
		BeforeEach(assertResourceCreation(&tenant1))
		AfterEach(assertResourceDeletion(&tenant1))
		BeforeEach(assertResourceCreation(&tenant2))
		AfterEach(assertResourceDeletion(&tenant2))

		Context("when listing the tenants with the inventory namespace", func() {
			It("should return all the tenants matching the inventory namespace", func() {
				tenant1.TypeMeta = metav1.TypeMeta{
					Kind:       "DBaaSTenant",
					APIVersion: v1alpha1.GroupVersion.Group + "/" + v1alpha1.GroupVersion.Version,
				}
				tenant2.TypeMeta = metav1.TypeMeta{
					Kind:       "DBaaSTenant",
					APIVersion: v1alpha1.GroupVersion.Group + "/" + v1alpha1.GroupVersion.Version,
				}

				tenantList, err := dRec.tenantListByInventoryNS(ctx, ns)
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantList.Items).Should(HaveLen(2))
				Expect(tenantList.Items).Should(ConsistOf(tenant1, tenant2))
			})
		})

		Context("when listing the tenants with an invalid namespace", func() {
			It("should return no tenants", func() {
				tenantList, err := dRec.tenantListByInventoryNS(ctx, "not-test-namespace")
				Expect(err).NotTo(HaveOccurred())
				Expect(tenantList.Items).Should(HaveLen(0))
			})
		})
	})
})

var _ = Describe("Check hasNoEditOrListVerbs function", func() {
	defer GinkgoRecover()

	// ClusterRoles created by operator should not grant 'edit' or 'list' rights
	inventoryList := createInventoryList()
	inventoryAuthz := getDevAuthzFromInventoryList(inventoryList, defaultTenant)
	serviceAdminAuthz := v1alpha1.DBaasUsersGroups{}
	tenantListAuthz := v1alpha1.DBaasUsersGroups{}
	clusterRole, clusterRolebinding := tenantRbacObjs(defaultTenant, serviceAdminAuthz, inventoryAuthz, tenantListAuthz)
	Expect(hasNoEditOrListVerbs(&clusterRole)).To(BeTrue())
	clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
		Verbs: []string{"watch"},
	})
	Expect(hasNoEditOrListVerbs(&clusterRole)).To(BeTrue())

	// ClusterRoles with edit rights should return 'false'
	clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
		Verbs: []string{"patch"},
	})
	Expect(hasNoEditOrListVerbs(&clusterRole)).To(BeFalse())

	// Bindings should be ignored, return 'true'
	Expect(hasNoEditOrListVerbs(&clusterRolebinding)).To(BeTrue())

	// Roles created by operator should not grant 'edit' or 'list' rights
	tenantList := createTestTenantList()
	role, rolebinding := inventoryRbacObjs(inventoryList.Items[0], tenantList)
	Expect(hasNoEditOrListVerbs(&role)).To(BeTrue())

	// Bindings should be ignored, return 'true'
	Expect(hasNoEditOrListVerbs(&rolebinding)).To(BeTrue())

	// Roles with edit rights should return 'false'
	role.Rules = append(role.Rules, rbacv1.PolicyRule{
		Verbs: []string{"create"},
	})
	Expect(hasNoEditOrListVerbs(&role)).To(BeFalse())

	role.Rules = []rbacv1.PolicyRule{
		{
			Verbs: []string{"delete"},
		},
	}
	Expect(hasNoEditOrListVerbs(&role)).To(BeFalse())

	role.Rules = []rbacv1.PolicyRule{
		{
			Verbs: []string{"update"},
		},
	}
	Expect(hasNoEditOrListVerbs(&role)).To(BeFalse())

	role.Rules = []rbacv1.PolicyRule{
		{
			Verbs: []string{"list"},
		},
	}
	Expect(hasNoEditOrListVerbs(&role)).To(BeFalse())
})

var _ = Describe("Check isOwner function", func() {
	defer GinkgoRecover()
	scheme := runtime.NewScheme()

	// error should be thrown due to missing scheme
	ownedObj := &unstructured.Unstructured{}
	owned, err := isOwner(&defaultTenant, ownedObj, scheme)
	Expect(err).NotTo(BeNil())
	Expect(owned).To(BeFalse())

	// with scheme added, error is nil, but owner check should be false
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	owned, err = isOwner(&defaultTenant, ownedObj, scheme)
	Expect(err).To(BeNil())
	Expect(owned).To(BeFalse())

	// with ownership set, owner check should be true
	Expect(ctrl.SetControllerReference(&defaultTenant, ownedObj, scheme)).To(BeNil())
	owned, err = isOwner(&defaultTenant, ownedObj, scheme)
	Expect(err).To(BeNil())
	Expect(owned).To(BeTrue())

	// setting namespaced object as owner of a cluster-scoped object should error
	//   owner check should return false
	inventory := &v1alpha1.DBaaSInventory{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test-ns"},
	}
	ownedObj = &unstructured.Unstructured{}
	Expect(ctrl.SetControllerReference(inventory, ownedObj, scheme)).NotTo(BeNil())
	owned, err = isOwner(inventory, ownedObj, scheme)
	Expect(err).To(BeNil())
	Expect(owned).To(BeFalse())

	// changing to a namespaced object should allow ownership to be set
	//   owner check should return true
	ownedObj.SetNamespace(inventory.GetNamespace())
	Expect(ctrl.SetControllerReference(inventory, ownedObj, scheme)).To(BeNil())
	owned, err = isOwner(inventory, ownedObj, scheme)
	Expect(err).To(BeNil())
	Expect(owned).To(BeTrue())
})

func getLastTransitionTimeForTest() time.Time {
	lastTransitionTime, err := time.Parse(time.RFC3339, "2021-06-30T22:17:55-04:00")
	Expect(err).NotTo(HaveOccurred())
	return lastTransitionTime.In(time.Local)
}
