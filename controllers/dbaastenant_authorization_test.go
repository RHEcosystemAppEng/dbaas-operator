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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTenantRbacObjs(t *testing.T) {
	RegisterFailHandler(Fail)
	defer GinkgoRecover()
	// Expect(err).NotTo(HaveOccurred())

	// nil serviceAdminAuthz & empty inventory list
	tenant := v1alpha1.DBaaSTenant{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: v1alpha1.DBaaSTenantSpec{
			Authz: v1alpha1.DBaasUsersGroups{
				Users: []string{"user1", "user2"},
			},
		},
	}
	clusterRoleName := "dbaas-" + tenant.Name + "-tenant-viewer"
	clusterRoleBindingName := clusterRoleName + "s"
	developerAuthz := getDevAuthzFromInventoryList(v1alpha1.DBaaSInventoryList{}, tenant)
	serviceAdminAuthz := v1alpha1.DBaasUsersGroups{}
	tenantListAuthz := v1alpha1.DBaasUsersGroups{}
	clusterRole, clusterRolebinding := tenantRbacObjs(tenant, serviceAdminAuthz, developerAuthz, tenantListAuthz)
	Expect(clusterRole).NotTo(BeNil())
	Expect(clusterRole.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.Subjects).To(BeEmpty())

	// nil serviceAdminAuthz & inventory.spec.authz
	developerAuthz = getDevAuthzFromInventoryList(v1alpha1.DBaaSInventoryList{Items: []v1alpha1.DBaaSInventory{{ObjectMeta: metav1.ObjectMeta{Name: "test"}}}}, tenant)
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, serviceAdminAuthz, developerAuthz, tenantListAuthz)
	Expect(clusterRole).NotTo(BeNil())
	Expect(clusterRole.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.Subjects).To(HaveLen(2))
	Expect(clusterRolebinding.Subjects[0].Name).To(Equal("user1"))
	Expect(clusterRolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[0].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[1].Name).To(Equal("user2"))
	Expect(clusterRolebinding.Subjects[1].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[1].Namespace).To(BeEmpty())

	// nil serviceAdminAuthz
	inventoryList := createInventoryList()
	developerAuthz = getDevAuthzFromInventoryList(inventoryList, tenant)
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, serviceAdminAuthz, developerAuthz, tenantListAuthz)
	Expect(clusterRole).NotTo(BeNil())
	Expect(clusterRole.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.Subjects).To(HaveLen(2))
	Expect(clusterRolebinding.Subjects[0].Name).To(Equal("user1"))
	Expect(clusterRolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[0].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[1].Name).To(Equal("group1"))
	Expect(clusterRolebinding.Subjects[1].Kind).To(Equal("Group"))
	Expect(clusterRolebinding.Subjects[1].Namespace).To(BeEmpty())

	// serviceAdminAuthz.users w/ duplicates
	serviceAdminAuthz = v1alpha1.DBaasUsersGroups{
		Users: []string{"admin1", "admin2", "admin2", "admin3"},
	}
	tenantListAuthz = v1alpha1.DBaasUsersGroups{
		Users: []string{"admin3"},
	}
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, serviceAdminAuthz, developerAuthz, tenantListAuthz)
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.Subjects).To(HaveLen(4))
	Expect(clusterRolebinding.Subjects[0].Name).To(Equal("admin1"))
	Expect(clusterRolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[0].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[1].Name).To(Equal("admin2"))
	Expect(clusterRolebinding.Subjects[1].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[1].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[2].Name).To(Equal("user1"))
	Expect(clusterRolebinding.Subjects[2].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[2].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[3].Name).To(Equal("group1"))
	Expect(clusterRolebinding.Subjects[3].Kind).To(Equal("Group"))
	Expect(clusterRolebinding.Subjects[3].Namespace).To(BeEmpty())

	// serviceAdminAuthz.groups w/ duplicates
	serviceAdminAuthz = v1alpha1.DBaasUsersGroups{
		Groups: []string{"group1", "group1"},
	}
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, serviceAdminAuthz, developerAuthz, tenantListAuthz)
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.Subjects).To(HaveLen(2))
	Expect(clusterRolebinding.Subjects[0].Name).To(Equal("user1"))
	Expect(clusterRolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[0].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[1].Name).To(Equal("group1"))
	Expect(clusterRolebinding.Subjects[1].Kind).To(Equal("Group"))
	Expect(clusterRolebinding.Subjects[1].Namespace).To(BeEmpty())

	// serviceAdminAuthz.users & groups w/ duplicates
	serviceAdminAuthz = v1alpha1.DBaasUsersGroups{
		Users:  []string{"user1", "user2", "user2", "system:serviceaccount:openshift-service-ca-operator:service-ca-operator"},
		Groups: []string{"group1", "group1", "group2"},
	}
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, serviceAdminAuthz, developerAuthz, tenantListAuthz)
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.RoleRef.Kind).To(Equal("ClusterRole"))
	Expect(clusterRolebinding.Subjects).To(HaveLen(5))
	Expect(clusterRolebinding.Subjects[0].Name).To(Equal("user1"))
	Expect(clusterRolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[0].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[0].APIGroup).NotTo(BeEmpty())
	Expect(clusterRolebinding.Subjects[1].Name).To(Equal("user2"))
	Expect(clusterRolebinding.Subjects[1].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[1].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[2].Name).To(Equal("service-ca-operator"))
	Expect(clusterRolebinding.Subjects[2].Kind).To(Equal("ServiceAccount"))
	Expect(clusterRolebinding.Subjects[2].Namespace).To(Equal("openshift-service-ca-operator"))
	Expect(clusterRolebinding.Subjects[2].APIGroup).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[3].Name).To(Equal("group1"))
	Expect(clusterRolebinding.Subjects[3].Kind).To(Equal("Group"))
	Expect(clusterRolebinding.Subjects[3].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[3].APIGroup).NotTo(BeEmpty())
	Expect(clusterRolebinding.Subjects[4].Name).To(Equal("group2"))
	Expect(clusterRolebinding.Subjects[4].Kind).To(Equal("Group"))
	Expect(clusterRolebinding.Subjects[4].Namespace).To(BeEmpty())

	// test matchSlices
	testSlice1 := []string{
		"user1",
		"user1",
		"user2",
		"user2",
		"user4",
		"user6",
	}
	testSlice2 := []string{
		"user1",
		"user3",
		"user5",
		"user6",
		"user6",
	}
	resultingSlice := []string{
		"user1",
		"user6",
	}
	Expect(matchSlices(testSlice1, testSlice2)).To(Equal(resultingSlice))
	Expect(matchSlices(testSlice1, []string{})).To(BeEmpty())

	Expect(removeFromSlice([]string{"user6"}, resultingSlice)).To(Equal([]string{"user1"}))
}

func TestInventoryRbacObjs(t *testing.T) {
	RegisterFailHandler(Fail)
	defer GinkgoRecover()

	// Expect(err).NotTo(HaveOccurred())
	namespace := "test-ns"
	tenantList := createTestTenantList()

	// nil spec.authz w/ default tenant set to wrong namespace
	inventory := v1alpha1.DBaaSInventory{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
	}
	roleName := "dbaas-" + inventory.Name + "-inventory-viewer"
	roleBindingName := roleName + "s"
	// add checks that rolebindings subjects are empty OR nil
	role, rolebinding := inventoryRbacObjs(inventory, tenantList)
	Expect(inventory.Namespace).To(Equal(namespace))
	Expect(role).NotTo(BeNil())
	Expect(role.Name).To(Equal(roleName))
	Expect(role.Namespace).To(Equal(namespace))
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.Namespace).To(Equal(namespace))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.Subjects).To(BeEmpty())

	// nil spec.authz w/ correct default tenant
	tenantList.Items[0].Spec.InventoryNamespace = namespace
	role, rolebinding = inventoryRbacObjs(inventory, tenantList)
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
	role, rolebinding = inventoryRbacObjs(inventory, tenantList)
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
	role, rolebinding = inventoryRbacObjs(inventory, tenantList)
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.Subjects).To(HaveLen(1))
	Expect(rolebinding.Subjects[0].Name).To(Equal("group1"))
	Expect(rolebinding.Subjects[0].Kind).To(Equal("Group"))
	Expect(rolebinding.Subjects[0].Namespace).To(Equal(inventory.Namespace))

	// spec.authz.users & groups w/ duplicates
	inventory.Spec.Authz = v1alpha1.DBaasUsersGroups{
		Users:  []string{"user1", "user2", "user2", "system:serviceaccount:openshift-service-ca-operator:service-ca-operator"},
		Groups: []string{"group1", "group1", "group2"},
	}
	role, rolebinding = inventoryRbacObjs(inventory, tenantList)
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.RoleRef.Kind).To(Equal("Role"))
	Expect(rolebinding.Subjects).To(HaveLen(5))
	Expect(rolebinding.Subjects[0].Name).To(Equal("user1"))
	Expect(rolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(rolebinding.Subjects[0].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[0].APIGroup).NotTo(BeEmpty())
	Expect(rolebinding.Subjects[1].Name).To(Equal("user2"))
	Expect(rolebinding.Subjects[1].Kind).To(Equal("User"))
	Expect(rolebinding.Subjects[1].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[2].Name).To(Equal("service-ca-operator"))
	Expect(rolebinding.Subjects[2].Kind).To(Equal("ServiceAccount"))
	Expect(rolebinding.Subjects[2].Namespace).To(Equal("openshift-service-ca-operator"))
	Expect(rolebinding.Subjects[2].APIGroup).To(BeEmpty())
	Expect(rolebinding.Subjects[3].Name).To(Equal("group1"))
	Expect(rolebinding.Subjects[3].Kind).To(Equal("Group"))
	Expect(rolebinding.Subjects[3].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[3].APIGroup).NotTo(BeEmpty())
	Expect(rolebinding.Subjects[4].Name).To(Equal("group2"))
	Expect(rolebinding.Subjects[4].Kind).To(Equal("Group"))
	Expect(rolebinding.Subjects[4].Namespace).To(Equal(inventory.Namespace))

	// multiple tenants and different authz configs
	newTenants := []v1alpha1.DBaaSTenant{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "tenant2",
			},
			Spec: v1alpha1.DBaaSTenantSpec{
				InventoryNamespace: namespace,
				Authz: v1alpha1.DBaasUsersGroups{
					Users: []string{"tenantUser"},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "tenant3",
			},
			Spec: v1alpha1.DBaaSTenantSpec{
				InventoryNamespace: "otherNS",
				Authz: v1alpha1.DBaasUsersGroups{
					Groups: []string{"otherGroup"},
				},
			},
		},
	}
	tenantList.Items = append(tenantList.Items, newTenants...)
	inventory.Spec.Authz = v1alpha1.DBaasUsersGroups{}
	role, rolebinding = inventoryRbacObjs(inventory, tenantList)
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

func createTestTenantList() v1alpha1.DBaaSTenantList {
	return v1alpha1.DBaaSTenantList{
		Items: []v1alpha1.DBaaSTenant{
			getDefaultTenant("wrong"),
		},
	}
}
