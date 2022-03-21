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
	oauthzv1 "github.com/openshift/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTenantRbacObjs(t *testing.T) {
	RegisterFailHandler(Fail)
	defer GinkgoRecover()
	// Expect(err).NotTo(HaveOccurred())

	// nil serviceAdminAuthz & empty inventory list
	tenant := v1alpha1.DBaaSTenant{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec:       v1alpha1.DBaaSTenantSpec{},
	}
	clusterRoleName := "dbaas-" + tenant.Name + "-tenant-viewer"
	clusterRoleBindingName := clusterRoleName + "s"
	developerAuthz := &oauthzv1.ResourceAccessReviewResponse{}
	serviceAdminAuthz := &oauthzv1.ResourceAccessReviewResponse{}
	tenantListAuthz := &oauthzv1.ResourceAccessReviewResponse{}
	clusterRole, clusterRolebinding := tenantRbacObjs(tenant, serviceAdminAuthz, developerAuthz, tenantListAuthz)
	Expect(clusterRole).NotTo(BeNil())
	Expect(clusterRole.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding).NotTo(BeNil())
	Expect(clusterRolebinding.Name).To(Equal(clusterRoleBindingName))
	Expect(clusterRolebinding.RoleRef.Name).To(Equal(clusterRoleName))
	Expect(clusterRolebinding.Subjects).To(BeNil())

	// nil serviceAdminAuthz & inventory.spec.authz
	developerAuthz = &oauthzv1.ResourceAccessReviewResponse{UsersSlice: []string{"user1", "user2"}}
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
	developerAuthz = &oauthzv1.ResourceAccessReviewResponse{UsersSlice: []string{"user1"}, GroupsSlice: []string{"group1"}}
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

	// serviceAdminAuthz users w/ duplicates
	serviceAdminAuthz = &oauthzv1.ResourceAccessReviewResponse{
		UsersSlice: []string{"admin1", "admin2", "admin2", "admin3"},
	}
	tenantListAuthz = &oauthzv1.ResourceAccessReviewResponse{
		UsersSlice: []string{"admin3"},
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

	// serviceAdminAuthz groups w/ duplicates
	serviceAdminAuthz = &oauthzv1.ResourceAccessReviewResponse{
		GroupsSlice: []string{"group1", "group1"},
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

	// serviceAdminAuthz users & groups w/ duplicates
	serviceAdminAuthz = &oauthzv1.ResourceAccessReviewResponse{
		UsersSlice:  []string{"user1", "user2", "user2", "system:serviceaccount:openshift-service-ca-operator:service-ca-operator"},
		GroupsSlice: []string{"group1", "group1", "group2"},
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
