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
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	operatorframework "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var testEnv *envtest.Environment
var ctx context.Context
var dRec *DBaaSReconciler
var iCtrl *spyctrl
var cCtrl *spyctrl
var inCtrl *spyctrl

const (
	testNamespace = "default"
	timeout       = time.Second * 30
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
	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = rbacv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = operatorframework.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			filepath.Join("..", "test", "crd"),
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	//+kubebuilder:scaffold:scheme

	ctx = context.Background()

	err = os.Setenv(InstallNamespaceEnvVar, testNamespace)
	Expect(err).NotTo(HaveOccurred())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		ClientDisableCacheFor: []client.Object{
			&operatorframework.ClusterServiceVersion{},
			&corev1.Secret{},
		},
	},
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sManager).NotTo(BeNil())

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

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
		gexec.KillAndWait(4 * time.Second)

		// Teardown the test environment once controller is fnished.
		// Otherwise from Kubernetes 1.21+, teardon timeouts waiting on
		// kube-apiserver to return
		err := testEnv.Stop()
		Expect(err).ToNot(HaveOccurred())
	}()
}, 60)
