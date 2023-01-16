package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

var _ = Describe("DBaaSPlatform controller", func() {
	Describe("trigger reconcile", func() {
		By("creating platform cr with syncPeriod")
		syncPeriod := 180
		cr := &dbaasv1beta1.DBaaSPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-platform",
				Namespace: testNamespace,
				Labels:    map[string]string{"managed-by": "dbaas-operator"},
			},
			Spec: dbaasv1beta1.DBaaSPlatformSpec{
				SyncPeriod: &syncPeriod,
			},
		}
		BeforeEach(assertResourceCreationIfNotExists(cr))
		It("should succeed", func() {
			By("checking the DBaaS resource")
			objectKey := client.ObjectKeyFromObject(cr)
			err := dRec.Get(ctx, objectKey, cr)
			Expect(err).NotTo(HaveOccurred())

			Expect(cr.Spec.SyncPeriod).NotTo(BeNil())
			Expect(FindStatusPlatform(cr.Status.PlatformsStatus, "test")).To(BeNil())
			setStatusPlatform(&cr.Status.PlatformsStatus, dbaasv1beta1.PlatformStatus{
				PlatformName:   "test",
				PlatformStatus: dbaasv1beta1.ResultInProgress,
			})
			setStatusCondition(&cr.Status.Conditions, dbaasv1beta1.DBaaSPlatformReadyType, metav1.ConditionFalse, dbaasv1beta1.InstallationInprogress, "DBaaS platform stack install in progress")
			Expect(FindStatusPlatform(cr.Status.PlatformsStatus, "test")).NotTo(BeNil())
			Expect(cr.Status.Conditions).NotTo(BeEmpty())
			Expect(cr.Status.Conditions[0].Type).To(Equal(dbaasv1beta1.DBaaSPlatformReadyType))
		})
	})
})
