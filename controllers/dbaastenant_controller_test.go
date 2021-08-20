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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

func TestTenantRbacObjs(t *testing.T) {
	RegisterFailHandler(Fail)
	defer GinkgoRecover()
	// Expect(err).NotTo(HaveOccurred())

	// nil spec.authz.serviceAdmin & empty inventory list
	tenant := v1alpha1.DBaaSTenant{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: v1alpha1.DBaaSTenantSpec{
			Authz: v1alpha1.DBaasAuthz{
				Developer: v1alpha1.DBaasUsersGroups{
					Users: []string{"user1", "user2"},
				},
			},
		},
	}
	clusterRoleName := "dbaas-" + tenant.Name + "-tenant-viewer"
	clusterRoleBindingName := clusterRoleName + "s"
	inventoryAuthz := getAllAuthzFromInventoryList(v1alpha1.DBaaSInventoryList{}, tenant)
	clusterRole, clusterRolebinding := tenantRbacObjs(tenant, inventoryAuthz)
	Expect(clusterRole).NotTo(BeNil())
	Expect(clusterRole.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.Subjects).To(HaveLen(0))

	// nil spec.authz.serviceAdmin & inventory.spec.authz
	inventoryAuthz = getAllAuthzFromInventoryList(v1alpha1.DBaaSInventoryList{Items: []v1alpha1.DBaaSInventory{{ObjectMeta: metav1.ObjectMeta{Name: "test"}}}}, tenant)
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, inventoryAuthz)
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

	// nil spec.authz.serviceAdmin
	inventoryList := createInventoryList()
	inventoryAuthz = getAllAuthzFromInventoryList(inventoryList, tenant)
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, inventoryAuthz)
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

	// spec.authz.serviceAdmin.users w/ duplicates
	tenant.Spec.Authz.ServiceAdmin = v1alpha1.DBaasUsersGroups{
		Users: []string{"admin1", "admin2", "admin2", "admin3"},
	}
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, inventoryAuthz)
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.Subjects).To(HaveLen(5))
	Expect(clusterRolebinding.Subjects[0].Name).To(Equal("admin1"))
	Expect(clusterRolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[0].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[1].Name).To(Equal("admin2"))
	Expect(clusterRolebinding.Subjects[1].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[1].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[2].Name).To(Equal("admin3"))
	Expect(clusterRolebinding.Subjects[2].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[3].Name).To(Equal("user1"))
	Expect(clusterRolebinding.Subjects[3].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[3].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[4].Name).To(Equal("group1"))
	Expect(clusterRolebinding.Subjects[4].Kind).To(Equal("Group"))
	Expect(clusterRolebinding.Subjects[4].Namespace).To(BeEmpty())

	// spec.authz.serviceAdmin.groups w/ duplicates
	tenant.Spec.Authz.ServiceAdmin = v1alpha1.DBaasUsersGroups{
		Groups: []string{"group1", "group1"},
	}
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, inventoryAuthz)
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

	// spec.authz.serviceAdmin.users & groups w/ duplicates
	tenant.Spec.Authz.ServiceAdmin = v1alpha1.DBaasUsersGroups{
		Users:  []string{"user1", "user2", "user2"},
		Groups: []string{"group1", "group1", "group2"},
	}
	clusterRole, clusterRolebinding = tenantRbacObjs(tenant, inventoryAuthz)
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.RoleRef.Kind).To(Equal("ClusterRole"))
	Expect(clusterRolebinding.Subjects).To(HaveLen(4))
	Expect(clusterRolebinding.Subjects[0].Name).To(Equal("user1"))
	Expect(clusterRolebinding.Subjects[0].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[0].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[1].Name).To(Equal("user2"))
	Expect(clusterRolebinding.Subjects[1].Kind).To(Equal("User"))
	Expect(clusterRolebinding.Subjects[1].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[2].Name).To(Equal("group1"))
	Expect(clusterRolebinding.Subjects[2].Kind).To(Equal("Group"))
	Expect(clusterRolebinding.Subjects[2].Namespace).To(BeEmpty())
	Expect(clusterRolebinding.Subjects[3].Name).To(Equal("group2"))
	Expect(clusterRolebinding.Subjects[3].Kind).To(Equal("Group"))
	Expect(clusterRolebinding.Subjects[3].Namespace).To(BeEmpty())
}

func createInventoryList() v1alpha1.DBaaSInventoryList {
	return v1alpha1.DBaaSInventoryList{
		Items: []v1alpha1.DBaaSInventory{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "inventory1",
				},
				Spec: v1alpha1.DBaaSOperatorInventorySpec{
					Authz: v1alpha1.DBaasUsersGroups{
						Users: []string{"user1"},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "inventory2",
				},
				Spec: v1alpha1.DBaaSOperatorInventorySpec{
					Authz: v1alpha1.DBaasUsersGroups{
						Groups: []string{"group1"},
					},
				},
			},
		},
	}
}
