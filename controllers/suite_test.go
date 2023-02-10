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
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/operator-framework/api/pkg/lib/version"
	operatorframework "github.com/operator-framework/api/pkg/operators/v1alpha1"
	rhobsv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/yaml"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc
var dRec *DBaaSReconciler
var iCtrl *spyctrl
var cCtrl *spyctrl
var inCtrl *spyctrl

const (
	testNamespace = "default"
	timeout       = time.Second * 10
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	scheme := runtime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = v1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = rbacv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = operatorframework.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = rhobsv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			filepath.Join("..", "test", "crd"),
		},
		ErrorIfCRDPathMissing: true,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "config", "webhook")},
		},
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	//+kubebuilder:scaffold:scheme

	ctx, cancel = context.WithCancel(context.TODO())

	err = os.Setenv(InstallNamespaceEnvVar, testNamespace)
	Expect(err).NotTo(HaveOccurred())

	webhookInstallOptions := &testEnv.WebhookInstallOptions
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme,
		Host:    webhookInstallOptions.LocalServingHost,
		Port:    webhookInstallOptions.LocalServingPort,
		CertDir: webhookInstallOptions.LocalServingCertDir,
		ClientDisableCacheFor: []client.Object{
			&operatorframework.ClusterServiceVersion{},
			&corev1.Secret{},
		},
	},
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sManager).NotTo(BeNil())

	err = (&v1beta1.DBaaSConnection{}).SetupWebhookWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&v1beta1.DBaaSInstance{}).SetupWebhookWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&v1beta1.DBaaSInventory{}).SetupWebhookWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&v1beta1.DBaaSPolicy{}).SetupWebhookWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&v1beta1.DBaaSProvider{}).SetupWebhookWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	dRec = &DBaaSReconciler{
		Client:           k8sManager.GetClient(),
		Scheme:           k8sManager.GetScheme(),
		InstallNamespace: testNamespace,
	}

	err = (&DBaaSPolicyReconciler{
		DBaaSReconciler: dRec,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	inventoryCtrl, err := (&DBaaSInventoryReconciler{
		DBaaSReconciler: dRec,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	connectionCtrl, err := (&DBaaSConnectionReconciler{
		DBaaSReconciler: dRec,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	instanceCtrl, err := (&DBaaSInstanceReconciler{
		DBaaSReconciler: dRec,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&DBaaSDefaultPolicyReconciler{
		DBaaSReconciler: dRec,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	iCtrl = newSpyController(inventoryCtrl)
	cCtrl = newSpyController(connectionCtrl)
	inCtrl = newSpyController(instanceCtrl)

	err = (&DBaaSProviderReconciler{
		DBaaSReconciler: dRec,
		InventoryCtrl:   iCtrl,
		ConnectionCtrl:  cCtrl,
		InstanceCtrl:    inCtrl,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	mockRDSController(k8sManager)

	createCSV(k8sManager)
	err = (&DBaaSPlatformReconciler{
		DBaaSReconciler: dRec,
		Log:             ctrl.Log.WithName("controllers").WithName("DBaaSPlatform"),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	checkRDSController(k8sManager)

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()
}, 60)

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
	err = os.Unsetenv(InstallNamespaceEnvVar)
	Expect(err).NotTo(HaveOccurred())
})

func createCSV(k8sManager manager.Manager) {
	yamlFile, err := os.ReadFile("../bundle/manifests/dbaas-operator.clusterserviceversion.yaml")
	Expect(err).ToNot(HaveOccurred())
	jsonConversion, err := yaml.YAMLToJSON(yamlFile)
	Expect(err).ToNot(HaveOccurred())

	verValue := gjson.Get(string(jsonConversion), "spec.version")
	newJson, err := sjson.Delete(string(jsonConversion), "spec.version")
	Expect(err).ToNot(HaveOccurred())
	newJson, err = sjson.Delete(newJson, "spec.webhookdefinitions")
	Expect(err).ToNot(HaveOccurred())

	csv := &operatorframework.ClusterServiceVersion{}
	err = json.Unmarshal([]byte(newJson), csv)
	Expect(err).ToNot(HaveOccurred())
	csv.Namespace = testNamespace

	ver := version.OperatorVersion{}
	err = ver.UnmarshalJSON(strconv.AppendQuote([]byte{}, verValue.String()))
	Expect(err).ToNot(HaveOccurred())
	csv.Spec.Version = ver
	Expect(csv.Spec.Version.String()).To(Equal(verValue.String()))

	serverClient, err := client.New(k8sManager.GetConfig(), client.Options{
		Scheme: k8sManager.GetScheme(),
	})
	Expect(err).ToNot(HaveOccurred())
	err = serverClient.Create(ctx, csv)
	Expect(err).ToNot(HaveOccurred())
}

func mockRDSController(k8sManager manager.Manager) {
	serverClient, err := client.New(
		k8sManager.GetConfig(),
		client.Options{
			Scheme: k8sManager.GetScheme(),
		},
	)
	Expect(err).ToNot(HaveOccurred())

	csv := &operatorframework.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ack-rds-controller.v0.1.0",
			Namespace: testNamespace,
		},
		Spec: operatorframework.ClusterServiceVersionSpec{
			DisplayName: "AWS Controllers for Kubernetes - Amazon RDS",
			CustomResourceDefinitions: operatorframework.CustomResourceDefinitions{
				Owned: []operatorframework.CRDDescription{
					{
						Name:    "dbinstances.rds.services.k8s.aws",
						Kind:    "DBInstance",
						Version: "v1alpha1",
					},
					{
						Name:    "dbclusters.rds.services.k8s.aws",
						Kind:    "DBCluster",
						Version: "v1alpha1",
					},
				},
			},
			InstallStrategy: operatorframework.NamedInstallStrategy{
				StrategyName: "deployment",
				StrategySpec: operatorframework.StrategyDetailsDeployment{
					DeploymentSpecs: []operatorframework.StrategyDeploymentSpec{
						{
							Name: "ack-rds-controller",
							Spec: appsv1.DeploymentSpec{
								Replicas: pointer.Int32(1),
								Selector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app.kubernetes.io/name": "ack-rds-controller",
									},
								},
								Template: corev1.PodTemplateSpec{
									ObjectMeta: metav1.ObjectMeta{
										Labels: map[string]string{
											"app.kubernetes.io/name": "ack-rds-controller",
										},
									},
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name:            "controller",
												Image:           "quay.io/ecosystem-appeng/busybox",
												ImagePullPolicy: corev1.PullIfNotPresent,
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
	err = serverClient.Create(ctx, csv)
	Expect(err).ToNot(HaveOccurred())

	subscription := &operatorframework.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ack-rds-controller-alpha-community-operators-openshift-marketplace",
			Namespace: testNamespace,
		},
		Spec: &operatorframework.SubscriptionSpec{
			CatalogSource:          "community-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			Package:                "ack-rds-controller",
			Channel:                "alpha",
			InstallPlanApproval:    "Automatic",
			StartingCSV:            "ack-rds-controller.v0.0.27",
		},
	}
	err = serverClient.Create(ctx, subscription)
	Expect(err).ToNot(HaveOccurred())
}

func checkRDSController(k8sManager manager.Manager) {
	serverClient, err := client.New(
		k8sManager.GetConfig(),
		client.Options{
			Scheme: k8sManager.GetScheme(),
		},
	)
	Expect(err).ToNot(HaveOccurred())

	csv := &operatorframework.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ack-rds-controller.v0.1.0",
			Namespace: testNamespace,
		},
	}
	err = serverClient.Get(ctx, client.ObjectKeyFromObject(csv), csv)
	Expect(err).To(HaveOccurred())
	Expect(errors.IsNotFound(err)).Should(BeTrue())

	subscription := &operatorframework.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ack-rds-controller-alpha-community-operators-openshift-marketplace",
			Namespace: testNamespace,
		},
	}
	err = serverClient.Get(ctx, client.ObjectKeyFromObject(subscription), subscription)
	Expect(err).To(HaveOccurred())
	Expect(errors.IsNotFound(err)).Should(BeTrue())
}
