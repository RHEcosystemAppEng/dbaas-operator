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
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

func TestInventoryRbacObjs(t *testing.T) {
	RegisterFailHandler(Fail)
	defer GinkgoRecover()

	// Expect(err).NotTo(HaveOccurred())
	namespace := "test-ns"
	TenantList = v1alpha1.DBaaSTenantList{
		Items: []v1alpha1.DBaaSTenant{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: v1alpha1.DBaaSTenantSpec{
					InventoryNamespace: "wrong",
					Authz: v1alpha1.DBaasAuthz{
						Developer: v1alpha1.DBaasUsersGroups{
							Groups: []string{"system:authenticated"},
						},
					},
				},
			},
		},
	}
	getTenantNamesandNS()

	// nil spec.authz w/ default tenant set to wrong namespace
	inventory := v1alpha1.DBaaSInventory{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
	}
	roleName := "dbaas-" + inventory.Name + "-inventory-viewer"
	roleBindingName := roleName + "s"
	role, rolebinding := inventoryRbacObjs(inventory)
	Expect(inventory.Namespace).To(Equal(namespace))
	Expect(role).NotTo(BeNil())
	Expect(role.Name).To(Equal(roleName))
	Expect(role.Namespace).To(Equal(namespace))
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.Namespace).To(Equal(namespace))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.Subjects).To(BeNil())

	// nil spec.authz w/ correct default tenant
	TenantList.Items[0].Spec.InventoryNamespace = namespace
	getTenantNamesandNS()
	role, rolebinding = inventoryRbacObjs(inventory)
	Expect(inventory.Namespace).To(Equal(namespace))
	Expect(role).NotTo(BeNil())
	Expect(role.Name).To(Equal(roleName))
	Expect(role.Namespace).To(Equal(namespace))
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.Namespace).To(Equal(namespace))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.Subjects).To(HaveLen(1))
	Expect(rolebinding.Subjects[0].Name).To(Equal("system:authenticated"))
	Expect(rolebinding.Subjects[0].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[0].Kind).To(Equal("Group"))

	// spec.authz.users w/ duplicates
	inventory.Spec.Authz = v1alpha1.DBaasUsersGroups{
		Users: []string{"user1", "user1", "user2"},
	}
	role, rolebinding = inventoryRbacObjs(inventory)
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.Subjects).To(HaveLen(2))
	Expect(rolebinding.Subjects[0].Name).To(Equal("user1"))
	Expect(rolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(rolebinding.Subjects[0].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[1].Name).To(Equal("user2"))
	Expect(rolebinding.Subjects[1].Kind).To(Equal("User"))
	Expect(rolebinding.Subjects[1].Namespace).To(Equal(inventory.Namespace))

	// spec.authz.groups w/ duplicates
	inventory.Spec.Authz = v1alpha1.DBaasUsersGroups{
		Groups: []string{"group1", "group1"},
	}
	role, rolebinding = inventoryRbacObjs(inventory)
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.Subjects).To(HaveLen(1))
	Expect(rolebinding.Subjects[0].Name).To(Equal("group1"))
	Expect(rolebinding.Subjects[0].Kind).To(Equal("Group"))
	Expect(rolebinding.Subjects[0].Namespace).To(Equal(inventory.Namespace))

	// spec.authz.users & groups w/ duplicates
	inventory.Spec.Authz = v1alpha1.DBaasUsersGroups{
		Users:  []string{"user1", "user2", "user2"},
		Groups: []string{"group1", "group1", "group2"},
	}
	role, rolebinding = inventoryRbacObjs(inventory)
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.RoleRef.Kind).To(Equal("Role"))
	Expect(rolebinding.Subjects).To(HaveLen(4))
	Expect(rolebinding.Subjects[0].Name).To(Equal("user1"))
	Expect(rolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(rolebinding.Subjects[0].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[1].Name).To(Equal("user2"))
	Expect(rolebinding.Subjects[1].Kind).To(Equal("User"))
	Expect(rolebinding.Subjects[1].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[2].Name).To(Equal("group1"))
	Expect(rolebinding.Subjects[2].Kind).To(Equal("Group"))
	Expect(rolebinding.Subjects[2].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[3].Name).To(Equal("group2"))
	Expect(rolebinding.Subjects[3].Kind).To(Equal("Group"))
	Expect(rolebinding.Subjects[3].Namespace).To(Equal(inventory.Namespace))

	// multiple tenants same spec.inventoryNamespace and different authz configs
	TenantList.Items = append(TenantList.Items, v1alpha1.DBaaSTenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant2",
		},
		Spec: v1alpha1.DBaaSTenantSpec{
			InventoryNamespace: namespace,
			Authz: v1alpha1.DBaasAuthz{
				Developer: v1alpha1.DBaasUsersGroups{
					Users: []string{"tenantUser"},
				},
			},
		},
	})
	getTenantNamesandNS()
	inventory.Spec.Authz = v1alpha1.DBaasUsersGroups{}
	role, rolebinding = inventoryRbacObjs(inventory)
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.RoleRef.Kind).To(Equal("Role"))
	Expect(rolebinding.Subjects).To(HaveLen(2))
	Expect(rolebinding.Subjects[0].Name).To(Equal("tenantUser"))
	Expect(rolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(rolebinding.Subjects[0].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[1].Name).To(Equal("system:authenticated"))
	Expect(rolebinding.Subjects[1].Kind).To(Equal("Group"))
	Expect(rolebinding.Subjects[1].Namespace).To(Equal(inventory.Namespace))
}

var _ = Describe("DBaaSInventory controller", func() {
	BeforeEach(assertResourceCreationIfNotExists(defaultProvider))
	BeforeEach(assertResourceCreationIfNotExists(defaultTenant))

	Describe("reconcile", func() {
		Context("after creating DBaaSInventory", func() {
			inventoryName := "test-inventory"
			credentialsRefName := "test-credentials-ref"
			DBaaSInventorySpec := &v1alpha1.DBaaSInventorySpec{
				CredentialsRef: &v1alpha1.NamespacedName{
					Name:      credentialsRefName,
					Namespace: testNamespace,
				},
			}
			createdDBaaSInventory := &v1alpha1.DBaaSInventory{
				ObjectMeta: metav1.ObjectMeta{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				Spec: v1alpha1.DBaaSOperatorInventorySpec{
					ProviderRef: v1alpha1.NamespacedName{
						Name: testProviderName,
					},
					DBaaSInventorySpec: *DBaaSInventorySpec,
				},
			}

			BeforeEach(assertResourceCreation(createdDBaaSInventory))
			AfterEach(assertResourceDeletion(createdDBaaSInventory))

			It("should create a provider inventory", assertProviderResourceCreated(createdDBaaSInventory, testInventoryKind, DBaaSInventorySpec))

			Context("when updating provider inventory status", func() {
				lastTransitionTime, err := time.Parse(time.RFC3339, "2021-06-30T22:17:55-04:00")
				Expect(err).NotTo(HaveOccurred())
				lastTransitionTime = lastTransitionTime.In(time.Local)
				status := &v1alpha1.DBaaSInventoryStatus{
					Instances: []v1alpha1.Instance{
						{
							InstanceID: "testInstanceID",
							Name:       "testInstance",
							InstanceInfo: map[string]string{
								"testInstanceInfo": "testInstanceInfo",
							},
						},
					},
					Conditions: []metav1.Condition{
						{
							Type:               "SpecSynced",
							Status:             metav1.ConditionTrue,
							Reason:             "SyncOK",
							LastTransitionTime: metav1.Time{Time: lastTransitionTime},
						},
					},
				}
				It("should update DBaaSInventory status", assertDBaaSResourceStatusUpdated(createdDBaaSInventory, testInventoryKind, status))
			})

			Context("when updating DBaaSInventory spec", func() {
				DBaaSInventorySpec := &v1alpha1.DBaaSInventorySpec{
					CredentialsRef: &v1alpha1.NamespacedName{
						Name:      "updated-test-credentialsRef",
						Namespace: "updated-test-namespace",
					},
				}
				It("should update provider inventory spec", assertProviderResourceSpecUpdated(createdDBaaSInventory, testInventoryKind, DBaaSInventorySpec))
			})
		})
	})
})
