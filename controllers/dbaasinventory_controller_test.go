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

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecurityObjs(t *testing.T) {
	RegisterFailHandler(Fail)
	defer GinkgoRecover()

	// Expect(err).NotTo(HaveOccurred())
	namespace := "test-ns"

	// nil spec.authz
	inventory := v1alpha1.DBaaSInventory{
		ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: namespace},
	}
	roleName := "dbaas-" + inventory.Name + "-developer"
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
	Expect(rolebinding.Subjects).To(HaveLen(1))
	Expect(rolebinding.Subjects[0].Name).To(Equal("system:authenticated"))
	Expect(rolebinding.Subjects[0].Namespace).To(Equal(inventory.Namespace))
	Expect(rolebinding.Subjects[0].Kind).To(Equal("Group"))

	// spec.authz.users
	inventory.Spec.Authz = v1alpha1.DBaasUsersGroups{
		Users: []string{"user1", "user2"},
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

	// spec.authz.groups
	inventory.Spec.Authz = v1alpha1.DBaasUsersGroups{
		Groups: []string{"group1"},
	}
	role, rolebinding = inventoryRbacObjs(inventory)
	Expect(rolebinding).NotTo(BeNil())
	Expect(rolebinding.Name).To(Equal(roleBindingName))
	Expect(rolebinding.RoleRef.Name).To(Equal(roleName))
	Expect(rolebinding.Subjects).To(HaveLen(1))
	Expect(rolebinding.Subjects[0].Name).To(Equal("group1"))
	Expect(rolebinding.Subjects[0].Kind).To(Equal("Group"))
	Expect(rolebinding.Subjects[0].Namespace).To(Equal(inventory.Namespace))

	// spec.authz.users & groups
	inventory.Spec.Authz = v1alpha1.DBaasUsersGroups{
		Users:  []string{"user1", "user2"},
		Groups: []string{"group1", "group2"},
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
}
