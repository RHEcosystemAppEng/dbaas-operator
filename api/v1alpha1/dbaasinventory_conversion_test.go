/*
Copyright 2022.

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

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

var _ = Context("DBaaSInventory Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			pFalse := false
			src := DBaaSInventory{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testName,
					Namespace: testNamespace,
				},
				Spec: DBaaSOperatorInventorySpec{
					DBaaSInventoryPolicy: DBaaSInventoryPolicy{
						DisableProvisions:    &pFalse,
						ConnectionNamespaces: &[]string{"test", "ha"},
						ConnectionNsSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"key": "value",
							},
						},
					},
					ProviderRef: NamespacedName{
						Name:      "trying",
						Namespace: "this",
					},
				},
				Status: DBaaSInventoryStatus{
					Conditions: []metav1.Condition{
						{
							Type: v1beta1.DBaaSInventoryProviderSyncType,
						},
					},
					Instances: []Instance{
						{
							InstanceID: "xx789x",
							Name:       "test",
						},
					},
				},
			}
			intermediate := v1beta1.DBaaSInventory{}
			dst := DBaaSInventory{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())
			Expect(dst).To(Equal(src))
		})
	})
})

var _ = Describe("DBaaSInventoryConversion", func() {
	inventoryName := "testInventory"
	testNamespace := "testNamespace"
	providerName := "testProvider"
	disableProvision := false
	testSecret := "testSecret"
	testConditionType := "conditionType"
	testConditionReason := "conditionReason"
	instanceID := "testInstance"
	instanceName := "testInstanceName"
	clusterID := "testCluster"
	clusterName := "testClusterName"
	instanceType := v1beta1.DatabaseServiceType("instance")
	clusterType := v1beta1.DatabaseServiceType("cluster")

	Context("ConvertTo", func() {
		v1alpha1Inv := DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: DBaaSOperatorInventorySpec{
				ProviderRef: NamespacedName{
					Name:      providerName,
					Namespace: testNamespace,
				},
				DBaaSInventorySpec: DBaaSInventorySpec{
					CredentialsRef: &LocalObjectReference{
						Name: testSecret,
					},
				},
				DBaaSInventoryPolicy: DBaaSInventoryPolicy{
					DisableProvisions: &disableProvision,
					ConnectionNamespaces: &[]string{
						testNamespace,
					},
					ConnectionNsSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "test",
						},
					},
				},
			},
			Status: DBaaSInventoryStatus{
				Conditions: []metav1.Condition{
					{
						Type:   testConditionType,
						Reason: testConditionReason,
						Status: metav1.ConditionTrue,
					},
				},
				Instances: []Instance{
					{
						InstanceID: instanceID,
						Name:       instanceName,
						InstanceInfo: map[string]string{
							"test": "instance",
						},
					},
					{
						InstanceID: clusterID,
						Name:       clusterName,
						InstanceInfo: map[string]string{
							"test": "cluster",
						},
					},
				},
			},
		}

		v1betaInv := v1beta1.DBaaSInventory{}

		expectedInv := v1beta1.DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: v1beta1.DBaaSOperatorInventorySpec{
				ProviderRef: v1beta1.NamespacedName{
					Name:      providerName,
					Namespace: testNamespace,
				},
				DBaaSInventorySpec: v1beta1.DBaaSInventorySpec{
					CredentialsRef: &v1beta1.LocalObjectReference{
						Name: testSecret,
					},
				},
				Policy: &v1beta1.DBaaSInventoryPolicy{
					DisableProvisions: &disableProvision,
					Connections: v1beta1.DBaaSConnectionPolicy{
						Namespaces: &[]string{
							testNamespace,
						},
						NsSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "test",
							},
						},
					},
				},
			},
			Status: v1beta1.DBaaSInventoryStatus{
				Conditions: []metav1.Condition{
					{
						Type:   testConditionType,
						Reason: testConditionReason,
						Status: metav1.ConditionTrue,
					},
				},
				DatabaseServices: []v1beta1.DatabaseService{
					{
						ServiceID:   instanceID,
						ServiceName: instanceName,
						ServiceInfo: map[string]string{
							"test": "instance",
						},
					},
					{
						ServiceID:   clusterID,
						ServiceName: clusterName,
						ServiceInfo: map[string]string{
							"test": "cluster",
						},
					},
				},
			},
		}

		It("should convert v1alpha1 to v1beta1", func() {
			err := v1alpha1Inv.ConvertTo(&v1betaInv)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v1betaInv).Should(Equal(expectedInv))
		})
	})

	Context("ConvertFrom", func() {
		v1alpha1Inv := DBaaSInventory{}

		v1betaInv := v1beta1.DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: v1beta1.DBaaSOperatorInventorySpec{
				ProviderRef: v1beta1.NamespacedName{
					Name:      providerName,
					Namespace: testNamespace,
				},
				DBaaSInventorySpec: v1beta1.DBaaSInventorySpec{
					CredentialsRef: &v1beta1.LocalObjectReference{
						Name: testSecret,
					},
				},
				Policy: &v1beta1.DBaaSInventoryPolicy{
					DisableProvisions: &disableProvision,
					Connections: v1beta1.DBaaSConnectionPolicy{
						Namespaces: &[]string{
							testNamespace,
						},
						NsSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "test",
							},
						},
					},
				},
			},
			Status: v1beta1.DBaaSInventoryStatus{
				Conditions: []metav1.Condition{
					{
						Type:   testConditionType,
						Reason: testConditionReason,
						Status: metav1.ConditionTrue,
					},
				},
				DatabaseServices: []v1beta1.DatabaseService{
					{
						ServiceID:   instanceID,
						ServiceName: instanceName,
						ServiceInfo: map[string]string{
							"test": "instance",
						},
						ServiceType: &instanceType,
					},
					{
						ServiceID:   clusterID,
						ServiceName: clusterName,
						ServiceInfo: map[string]string{
							"test": "cluster",
						},
						ServiceType: &clusterType,
					},
				},
			},
		}

		expectedInv := DBaaSInventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      inventoryName,
				Namespace: testNamespace,
			},
			Spec: DBaaSOperatorInventorySpec{
				ProviderRef: NamespacedName{
					Name:      providerName,
					Namespace: testNamespace,
				},
				DBaaSInventorySpec: DBaaSInventorySpec{
					CredentialsRef: &LocalObjectReference{
						Name: testSecret,
					},
				},
				DBaaSInventoryPolicy: DBaaSInventoryPolicy{
					DisableProvisions: &disableProvision,
					ConnectionNamespaces: &[]string{
						testNamespace,
					},
					ConnectionNsSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "test",
						},
					},
				},
			},
			Status: DBaaSInventoryStatus{
				Conditions: []metav1.Condition{
					{
						Type:   testConditionType,
						Reason: testConditionReason,
						Status: metav1.ConditionTrue,
					},
				},
				Instances: []Instance{
					{
						InstanceID: instanceID,
						Name:       instanceName,
						InstanceInfo: map[string]string{
							"test": "instance",
						},
					},
					{
						InstanceID: clusterID,
						Name:       clusterName,
						InstanceInfo: map[string]string{
							"test": "cluster",
						},
					},
				},
			},
		}

		It("should convert v1alpha1 from v1beta1", func() {
			err := v1alpha1Inv.ConvertFrom(&v1betaInv)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v1alpha1Inv).Should(Equal(expectedInv))
		})
	})
})
