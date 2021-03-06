package controllers

import (
	. "github.com/onsi/ginkgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

var _ = Describe("DBaaSPlatform controller", func() {
	Describe("trigger reconcile", func() {

		By("creating platform cr with syncPeriod")
		syncPeriod := 180
		cr := &dbaasv1alpha1.DBaaSPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-platform",
				Namespace: testNamespace,
				Labels:    map[string]string{"managed-by": "dbaas-operator"},
			},

			Spec: dbaasv1alpha1.DBaaSPlatformSpec{

				SyncPeriod: &syncPeriod,
			},
		}
		BeforeEach(assertResourceCreation(cr))
		AfterEach(assertResourceDeletion(cr))

		By("creating platform cr with empty/nil syncPeriod ")

		cr = &dbaasv1alpha1.DBaaSPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-platform",
				Namespace: testNamespace,
				Labels:    map[string]string{"managed-by": "dbaas-operator"},
			},

			Spec: dbaasv1alpha1.DBaaSPlatformSpec{

				SyncPeriod: nil,
			},
		}
		BeforeEach(assertResourceCreation(cr))
		AfterEach(assertResourceDeletion(cr))

	})
})
