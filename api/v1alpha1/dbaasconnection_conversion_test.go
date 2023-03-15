/*
Copyright 2023 The OpenShift Database Access Authors.

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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

var _ = Context("DBaaSConnection Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := DBaaSConnection{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testName,
					Namespace: testNamespace,
				},
				Spec: DBaaSConnectionSpec{
					InstanceID: "xxx32xx",
					InventoryRef: NamespacedName{
						Name: "test",
					},
					InstanceRef: &NamespacedName{
						Name: "test",
					},
				},
				Status: DBaaSConnectionStatus{
					Conditions: []metav1.Condition{
						{
							Type: v1beta1.DBaaSConnectionProviderSyncType,
						},
					},
					CredentialsRef:    &corev1.LocalObjectReference{Name: "test"},
					ConnectionInfoRef: &corev1.LocalObjectReference{Name: "test"},
				},
			}
			intermediate := v1beta1.DBaaSConnection{}
			dst := DBaaSConnection{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())
			Expect(dst).To(Equal(src))
		})
	})
})

var _ = Describe("DBaaSConnectionConversion", func() {
	inventoryName := "testInventory"
	connectionName := "testConnection"
	testNamespace := "testNamespace"
	databaseServiceID := "testServiceID"
	databaseServiceName := "testServiceName"
	databaseServiceType := v1beta1.DatabaseServiceType("testServiceType")
	testConfigMap := "testConfigMap"
	testSecret := "testSecret"
	testConditionType := "conditionType"
	testConditionReason := "conditionReason"

	Context("ConvertTo", func() {
		v1alpha1Conn := DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: DBaaSConnectionSpec{
				InventoryRef: NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				InstanceID: databaseServiceID,
				InstanceRef: &NamespacedName{
					Name:      databaseServiceName,
					Namespace: testNamespace,
				},
			},
			Status: DBaaSConnectionStatus{
				Conditions: []metav1.Condition{
					{
						Type:   testConditionType,
						Reason: testConditionReason,
						Status: metav1.ConditionTrue,
					},
				},
				CredentialsRef: &corev1.LocalObjectReference{
					Name: testSecret,
				},
				ConnectionInfoRef: &corev1.LocalObjectReference{
					Name: testConfigMap,
				},
			},
		}

		v1betaConn := v1beta1.DBaaSConnection{}

		expectedConn := v1beta1.DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: v1beta1.DBaaSConnectionSpec{
				InventoryRef: v1beta1.NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				DatabaseServiceID: databaseServiceID,
				DatabaseServiceRef: &v1beta1.NamespacedName{
					Name:      databaseServiceName,
					Namespace: testNamespace,
				},
			},
			Status: v1beta1.DBaaSConnectionStatus{
				Conditions: []metav1.Condition{
					{
						Type:   testConditionType,
						Reason: testConditionReason,
						Status: metav1.ConditionTrue,
					},
				},
				CredentialsRef: &corev1.LocalObjectReference{
					Name: testSecret,
				},
				ConnectionInfoRef: &corev1.LocalObjectReference{
					Name: testConfigMap,
				},
			},
		}

		It("should convert v1alpha1 to v1beta1", func() {
			err := v1alpha1Conn.ConvertTo(&v1betaConn)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v1betaConn).Should(Equal(expectedConn))
		})
	})

	Context("ConvertFrom", func() {
		v1alpha1Conn := DBaaSConnection{}

		v1betaConn := v1beta1.DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: v1beta1.DBaaSConnectionSpec{
				InventoryRef: v1beta1.NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				DatabaseServiceID: databaseServiceID,
				DatabaseServiceRef: &v1beta1.NamespacedName{
					Name:      databaseServiceName,
					Namespace: testNamespace,
				},
				DatabaseServiceType: &databaseServiceType,
			},
			Status: v1beta1.DBaaSConnectionStatus{
				Conditions: []metav1.Condition{
					{
						Type:   testConditionType,
						Reason: testConditionReason,
						Status: metav1.ConditionTrue,
					},
				},
				CredentialsRef: &corev1.LocalObjectReference{
					Name: testSecret,
				},
				ConnectionInfoRef: &corev1.LocalObjectReference{
					Name: testConfigMap,
				},
			},
		}

		expectedConn := DBaaSConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectionName,
				Namespace: testNamespace,
			},
			Spec: DBaaSConnectionSpec{
				InventoryRef: NamespacedName{
					Name:      inventoryName,
					Namespace: testNamespace,
				},
				InstanceID: databaseServiceID,
				InstanceRef: &NamespacedName{
					Name:      databaseServiceName,
					Namespace: testNamespace,
				},
			},
			Status: DBaaSConnectionStatus{
				Conditions: []metav1.Condition{
					{
						Type:   testConditionType,
						Reason: testConditionReason,
						Status: metav1.ConditionTrue,
					},
				},
				CredentialsRef: &corev1.LocalObjectReference{
					Name: testSecret,
				},
				ConnectionInfoRef: &corev1.LocalObjectReference{
					Name: testConfigMap,
				},
			},
		}

		It("should convert v1alpha1 from v1beta1", func() {
			err := v1alpha1Conn.ConvertFrom(&v1betaConn)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v1alpha1Conn).Should(Equal(expectedConn))
		})
	})
})
