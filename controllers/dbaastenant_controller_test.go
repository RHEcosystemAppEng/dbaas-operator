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
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
