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

package main

import (
	"flag"
	"fmt"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"go.uber.org/zap/zapcore"
	"golang.org/x/mod/semver"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/RHsyseng/operator-utils/pkg/utils/openshift"
	oauthzv1 "github.com/openshift/api/authorization/v1"
	consolev1 "github.com/openshift/api/console/v1"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	operatorv1 "github.com/openshift/api/operator/v1"
	oauthzclientv1 "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
	coreosv1 "github.com/operator-framework/api/pkg/operators/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers"
	operatorframework "github.com/operator-framework/api/pkg/operators/v1alpha1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(operatorframework.AddToScheme(scheme))
	utilruntime.Must(coreosv1.AddToScheme(scheme))
	utilruntime.Must(consolev1alpha1.Install(scheme))
	utilruntime.Must(operatorv1.Install(scheme))
	utilruntime.Must(oauthzv1.Install(scheme))
	utilruntime.Must(rbacv1.AddToScheme(scheme))
	utilruntime.Must(consolev1.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var logLevel string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&logLevel, "log-level", "info", "Log level.")

	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		//default to info level
		level = zapcore.InfoLevel
	}
	opts := zap.Options{
		Development: true,
		Level:       level,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	cfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e4addb06.redhat.com",
		ClientDisableCacheFor: []client.Object{
			&operatorframework.ClusterServiceVersion{},
			&corev1.Secret{},
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	DBaaSReconciler := &controllers.DBaaSReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	if DBaaSReconciler.InstallNamespace, err = controllers.GetInstallNamespace(); err != nil {
		setupLog.Error(err, "unable to retrieve install namespace. default Tenant object cannot be installed")
	}
	authzReconciler := &controllers.DBaaSAuthzReconciler{
		DBaaSReconciler:       DBaaSReconciler,
		AuthorizationV1Client: oauthzclientv1.NewForConfigOrDie(cfg),
	}
	if err = (&controllers.DBaaSTenantAuthzReconciler{
		DBaaSAuthzReconciler: authzReconciler,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DBaaSTenantAuthz")
		os.Exit(1)
	}
	connectionCtrl, err := (&controllers.DBaaSConnectionReconciler{
		DBaaSReconciler: DBaaSReconciler,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DBaaSConnection")
		os.Exit(1)
	}
	inventoryCtrl, err := (&controllers.DBaaSInventoryReconciler{
		DBaaSReconciler: DBaaSReconciler,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DBaaSInventory")
		os.Exit(1)
	}
	instanceCtrl, err := (&controllers.DBaaSInstanceReconciler{
		DBaaSReconciler: DBaaSReconciler,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DBaaSInstance")
		os.Exit(1)
	}
	if err = (&controllers.DBaaSDefaultTenantReconciler{
		DBaaSReconciler: DBaaSReconciler,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DBaaSDefaultTenant")
		os.Exit(1)
	}
	if err = (&controllers.DBaaSProviderReconciler{
		DBaaSReconciler: DBaaSReconciler,
		ConnectionCtrl:  connectionCtrl,
		InventoryCtrl:   inventoryCtrl,
		InstanceCtrl:    instanceCtrl,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DBaaSProvider")
		os.Exit(1)
	}
	//We'll just make sure to set `ENABLE_WEBHOOKS=false` when we run locally.

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&v1alpha1.DBaaSConnection{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "DBaaSConnection")
			os.Exit(1)
		}
		if err = (&v1alpha1.DBaaSInventory{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "DBaaSInventory")
			os.Exit(1)
		}
		if err = (&v1alpha1.DBaaSTenant{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "DBaaSTenant")
			os.Exit(1)
		}
	}
	if err = (&controllers.DBaaSTenantReconciler{
		DBaaSAuthzReconciler: authzReconciler,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DBaaSTenant")
		os.Exit(1)
	}

	var ocpVersion string
	info, err := openshift.GetPlatformInfo(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to get platform info")
	}
	if info.IsOpenShift() {
		mappedVersion := openshift.MapKnownVersion(info)
		if mappedVersion.Version != "" {
			ocpVersion = semver.MajorMinor("v" + mappedVersion.Version)
			setupLog.Info(fmt.Sprintf("OpenShift Version: %s", ocpVersion))
		} else {
			setupLog.Info("OpenShift version could not be determined.")
		}
	}
	if err = (&controllers.DBaaSPlatformReconciler{
		DBaaSReconciler: DBaaSReconciler,
		Log:             ctrl.Log.WithName("controllers").WithName("DBaaSPlatform"),
		OcpVersion:      ocpVersion,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DBaaSPlatform")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
