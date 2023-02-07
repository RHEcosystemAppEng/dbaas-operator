package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
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

	Describe("install dummy secret and configmap for rds-controller upgrade", func() {
		It("should find the dummy secret and configmap after installation", func() {
			Eventually(func() bool {
				secret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ack-rds-user-secrets", //#nosec G101
						Namespace: testNamespace,
					},
				}
				err := dRec.Get(ctx, client.ObjectKeyFromObject(secret), secret)
				if err != nil {
					return false
				}
				return true
			}, timeout).Should(BeTrue())

			Eventually(func() bool {
				cm := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ack-rds-user-config",
						Namespace: testNamespace,
					},
				}
				err := dRec.Get(ctx, client.ObjectKeyFromObject(cm), cm)
				if err != nil {
					return false
				}
				return true
			}, timeout).Should(BeTrue())
		})
	})

	Describe("delete deployment for rds-controller v0.1.3 upgrade", func() {
		csv := &v1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ack-rds-controller.v0.1.3",
				Namespace: testNamespace,
			},
			Spec: v1alpha1.ClusterServiceVersionSpec{
				DisplayName: "AWS Controllers for Kubernetes - Amazon RDS",
				InstallStrategy: v1alpha1.NamedInstallStrategy{
					StrategyName: "deployment",
					StrategySpec: v1alpha1.StrategyDetailsDeployment{
						DeploymentSpecs: []v1alpha1.StrategyDeploymentSpec{
							{
								Name: "ack-rds-controller",
								Spec: appsv1.DeploymentSpec{
									Replicas: pointer.Int32(1),
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app.kubernetes.io/name": "ack-rds-controller",
										},
									},
									Template: v1.PodTemplateSpec{
										ObjectMeta: metav1.ObjectMeta{
											Labels: map[string]string{
												"app.kubernetes.io/name": "ack-rds-controller",
											},
										},
										Spec: v1.PodSpec{
											Containers: []v1.Container{
												{
													Name:            "controller",
													Image:           "quay.io/ecosystem-appeng/busybox",
													ImagePullPolicy: v1.PullIfNotPresent,
													Command:         []string{"sh", "-c", "echo The app is running! && sleep 3600"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		BeforeEach(assertResourceCreation(csv))
		AfterEach(assertResourceDeletion(csv))

		rdsDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ack-rds-controller",
				Namespace: testNamespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: pointer.Int32(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"name": "ack-rds-controller",
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"name": "ack-rds-controller",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:            "ack-rds-controller",
								Image:           "quay.io/ecosystem-appeng/busybox",
								ImagePullPolicy: v1.PullIfNotPresent,
								Command:         []string{"sh", "-c", "echo The app is running! && sleep 3600"},
							},
						},
					},
				},
			},
		}
		BeforeEach(assertResourceCreation(rdsDeployment))
		AfterEach(assertResourceDeletionIfNotExists(rdsDeployment))

		It("should delete the deployment for the csv error", func() {
			By("making the csv in failed status")
			Eventually(func() bool {
				if err := dRec.Get(ctx, client.ObjectKeyFromObject(csv), csv); err != nil {
					return false
				}

				csv.Status.Phase = v1alpha1.CSVPhaseFailed
				csv.Status.Reason = v1alpha1.CSVReasonComponentFailed
				csv.Status.Message = "install strategy failed: Deployment.apps \"ack-rds-controller\" is invalid: spec.selector: " +
					"Invalid value: v1.LabelSelector{MatchLabels:map[string]string{\"app.kubernetes.io/name\":\"ack-rds-controller\"}, " +
					"MatchExpressions:[]v1.LabelSelectorRequirement(nil)}: field is immutable"

				if err := dRec.Status().Update(ctx, csv); err != nil {
					return false
				}
				return true
			}, timeout).Should(BeTrue())

			By("checking if the deployment is deleted")
			Eventually(func() bool {
				if err := dRec.Get(ctx, client.ObjectKeyFromObject(rdsDeployment), rdsDeployment); err != nil {
					if errors.IsNotFound(err) {
						return true
					}
				}
				return false
			}, timeout).Should(BeTrue())
		})
	})
})
