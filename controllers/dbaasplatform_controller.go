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
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/util"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/consoleplugin"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/providersinstallation"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/quickstartinstallation"
	"golang.org/x/mod/semver"
	apimeta "k8s.io/apimachinery/pkg/api/meta"

	"github.com/go-logr/logr"

	metrics "github.com/RHEcosystemAppEng/dbaas-operator/controllers/metrics"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

/*
Platform resources to be created or updated explicitly by the DBaaS operator:
	Crunchy Bridge operator: CatalogSource, Subscription, OperatorGroup
	MongoDB Atlas operator: CatalogSource, Subscription, OperatorGroup
	Service Binding operator: Subscription
	DBaaS Dynamic Plugin: Service, Deployment, ConsolePlugin, Console (updated)

Platform resources to be removed by the DBaaS operator by setting owner reference:
	Crunchy Bridge operator: Subscription, CSV
	MongoDB Atlas operator: Subscription, CSV
	Service Binding operator: Subscription, CSV
	DBaaS Dynamic Plugin: Service, Deployment

Platform resources NOT to be removed or updated by the DBaaS operator:
	Crunchy Bridge operator: CatalogSource (in different namespace), OperatorGroup (in different namespace)
	MongoDB Atlas operator: CatalogSource (in different namespace), OperatorGroup (in different namespace)
	DBaaS Dynamic Plugin: ConsolePlugin (cluster-scoped), Console (cluster-scoped)
*/

// Constants for requeue
const (
	RequeueDelaySuccess = 10 * time.Second
	RequeueDelayError   = 5 * time.Second
)

// DBaaSPlatformReconciler reconciles a DBaaSPlatform object
type DBaaSPlatformReconciler struct {
	*DBaaSReconciler
	Log                 logr.Logger
	installComplete     bool
	operatorNameVersion string
	OcpVersion          string
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=*/finalizers,verbs=update
//+kubebuilder:rbac:groups=operators.coreos.com,resources=catalogsources;operatorgroups,verbs=get;list;create;update;watch
//+kubebuilder:rbac:groups=operators.coreos.com,resources=subscriptions,verbs=get;list;create;update;watch;delete
//+kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=get;update;delete
//+kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;statefulsets,verbs=get;list;create;update;watch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;create;update;watch;delete
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins;consolequickstarts,verbs=get;list;create;update;watch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=consoles,verbs=get;list;update;watch
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=monitoringstacks,verbs=get;list;create;update;watch;delete
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=infrastructures,verbs=get;list;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=consoles,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSPlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	cr := &dbaasv1alpha1.DBaaSPlatform{}
	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.V(1).Info("DBaaSPlatform CR not found, has been deleted")
			return ctrl.Result{}, nil
		}
		// error fetching DBaaSPlatform instance, requeue and try again
		logger.Error(err, "Error in Get of DBaaSPlatform CR")
		return ctrl.Result{}, err
	}

	var finished = true
	execution := metrics.PlatformInstallStart()
	var platforms map[dbaasv1alpha1.PlatformsName]dbaasv1alpha1.PlatformConfig

	consoleURL, err := util.GetOpenshiftConsoleURL(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Error in getting of openshift consoleURl")
	}

	platformType, err := util.GetOpenshiftPlatform(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Error in getting of openshift platform Type")
	}
	metrics.SetOpenShiftInstallationInfoMetric(r.operatorNameVersion, consoleURL, string(platformType))
	if cr.DeletionTimestamp == nil {
		platforms = reconcilers.InstallationPlatforms
	}

	nextStatus := cr.Status.DeepCopy()
	nextPlatformStatus := dbaasv1alpha1.PlatformStatus{}
	for platform, platformConfig := range platforms {
		nextPlatformStatus.PlatformName = platform
		reconciler := r.getReconcilerForPlatform(platformConfig)
		if reconciler != nil {
			var status dbaasv1alpha1.PlatformsInstlnStatus
			var err error

			if cr.DeletionTimestamp == nil {
				status, err = reconciler.Reconcile(ctx, cr)
				metrics.SetPlatformStatusMetric(platform, status, platformConfig.CSV)

			} else {
				status, err = reconciler.Cleanup(ctx, cr)
				metrics.CleanPlatformStatusMetric(platform, status, platformConfig.CSV)
			}

			if err != nil {
				nextPlatformStatus.LastMessage = err.Error()
				return ctrl.Result{}, err
			}
			// Reset error message when everything went well
			nextPlatformStatus.LastMessage = ""
			nextPlatformStatus.PlatformStatus = status
			setStatusPlatform(&nextStatus.PlatformsStatus, nextPlatformStatus)

			// If a platform is not complete, do not continue with the next
			if status != dbaasv1alpha1.ResultSuccess {
				if cr.DeletionTimestamp == nil {
					execution.PlatformStackInstallationMetric(cr, r.operatorNameVersion)
					logger.Info("DBaaS platform stack install in progress", "working platform", platform)
					setStatusCondition(&nextStatus.Conditions, dbaasv1alpha1.DBaaSPlatformReadyType, metav1.ConditionFalse, dbaasv1alpha1.InstallationInprogress, "DBaaS platform stack install in progress")
				} else {
					logger.Info("DBaaS platform stack cleanup in progress", "working platform", platform)
					setStatusCondition(&nextStatus.Conditions, dbaasv1alpha1.DBaaSPlatformReadyType, metav1.ConditionUnknown, dbaasv1alpha1.InstallationCleanup, "DBaaS platform stack cleanup in progress")
				}
				finished = false
				break
			}
		}
	}
	if cr.DeletionTimestamp == nil && finished && !r.installComplete {
		r.installComplete = true
		setStatusCondition(&nextStatus.Conditions, dbaasv1alpha1.DBaaSPlatformReadyType, metav1.ConditionTrue, dbaasv1alpha1.Ready, "DBaaS platform stack installation complete")
		execution.PlatformStackInstallationMetric(cr, r.operatorNameVersion)
		logger.Info("DBaaS platform stack installation complete")
	}

	return r.updateStatus(cr, nextStatus)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSPlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// envVar set for all operators
	operatorNameEnvVar, found := os.LookupEnv("OPERATOR_CONDITION_NAME")
	if !found {
		err := fmt.Errorf("OPERATOR_CONDITION_NAME must be set")
		return err
	}
	r.operatorNameVersion = operatorNameEnvVar
	// Creates a new managed install CR if it is not available
	kubeConfig := mgr.GetConfig()
	client, _ := k8sclient.New(kubeConfig, k8sclient.Options{
		Scheme: mgr.GetScheme(),
	})
	_, err := r.createPlatformCR(context.Background(), client)
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbaasv1alpha1.DBaaSPlatform{}).
		Complete(r)
}

func (r *DBaaSPlatformReconciler) createPlatformCR(ctx context.Context, serverClient k8sclient.Client) (*dbaasv1alpha1.DBaaSPlatform, error) {

	namespace := r.InstallNamespace
	dbaaSPlatformList := &dbaasv1alpha1.DBaaSPlatformList{}
	listOpts := []k8sclient.ListOption{
		k8sclient.InNamespace(namespace),
	}
	err := serverClient.List(ctx, dbaaSPlatformList, listOpts...)
	if err != nil {
		return nil, fmt.Errorf("could not get a list of dbaas platform intallation CR: %w", err)
	}

	var cr *dbaasv1alpha1.DBaaSPlatform
	syncPeriod := 180
	if len(dbaaSPlatformList.Items) == 0 {

		cr = &dbaasv1alpha1.DBaaSPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-platform",
				Namespace: strings.TrimSpace(namespace),
				Labels:    map[string]string{"managed-by": "dbaas-operator"},
			},

			Spec: dbaasv1alpha1.DBaaSPlatformSpec{

				SyncPeriod: &syncPeriod,
			},
		}

		owner, err := reconcilers.GetDBaaSOperatorCSV(ctx, namespace, r.operatorNameVersion, serverClient)
		if err != nil {
			return nil, fmt.Errorf("could not create dbaas platform intallation CR: %w", err)
		}
		err = ctrl.SetControllerReference(owner, cr, r.Scheme)
		if err != nil {
			return nil, fmt.Errorf("could not create dbaas platform intallation CR: %w", err)
		}

		err = serverClient.Create(ctx, cr)
		if err != nil {
			return nil, fmt.Errorf("could not create  CR in %s namespace: %w", namespace, err)
		}
	} else if len(dbaaSPlatformList.Items) == 1 {
		cr = &dbaaSPlatformList.Items[0]
	} else {
		return nil, fmt.Errorf("too many DBaaSPlafrom resources found. Expecting 1, found %d DBaaSPlatform resources in %s namespace", len(dbaaSPlatformList.Items), namespace)
	}
	return cr, nil

}

func (r *DBaaSPlatformReconciler) getReconcilerForPlatform(platformConfig dbaasv1alpha1.PlatformConfig) reconcilers.PlatformReconciler {
	switch platformConfig.Type {
	case dbaasv1alpha1.TypeOperator:
		return providersinstallation.NewReconciler(r.Client, r.Scheme, r.Log, platformConfig)
	case dbaasv1alpha1.TypeConsolePlugin:
		if r.OcpVersion != "" && semver.Compare(r.OcpVersion, "v4.9") <= 0 {
			platformConfig.Image += reconcilers.ConsolePlugin49Tag
		}
		return consoleplugin.NewReconciler(r.Client, r.Scheme, r.Log, platformConfig)
	case dbaasv1alpha1.TypeQuickStart:
		return quickstartinstallation.NewReconciler(r.Client, r.Scheme, r.Log)
	}

	return nil
}

func (r *DBaaSPlatformReconciler) updateStatus(cr *dbaasv1alpha1.DBaaSPlatform, nextStatus *dbaasv1alpha1.DBaaSPlatformStatus) (ctrl.Result, error) {
	if !reflect.DeepEqual(&cr.Status, nextStatus) {
		nextStatus.DeepCopyInto(&cr.Status)
		err := r.Client.Status().Update(context.Background(), cr)
		if err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: RequeueDelayError,
			}, err
		}
	}

	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: RequeueDelaySuccess,
	}, nil
}

// setStatusPlatform set the new status for installation of platforms
func setStatusPlatform(PlatformsStatus *[]dbaasv1alpha1.PlatformStatus, newPlatformStatus dbaasv1alpha1.PlatformStatus) {
	if PlatformsStatus == nil {
		return
	}
	existingPlatformStatus := FindStatusPlatform(*PlatformsStatus, newPlatformStatus.PlatformName)
	if existingPlatformStatus == nil {
		*PlatformsStatus = append(*PlatformsStatus, newPlatformStatus)
		return
	}

	if existingPlatformStatus.PlatformStatus != newPlatformStatus.PlatformStatus {
		existingPlatformStatus.PlatformStatus = newPlatformStatus.PlatformStatus
	}

	existingPlatformStatus.PlatformStatus = newPlatformStatus.PlatformStatus
	existingPlatformStatus.LastMessage = newPlatformStatus.LastMessage
}

// FindStatusPlatform finds the platformName in platforms status.
func FindStatusPlatform(platformsStatus []dbaasv1alpha1.PlatformStatus, platformName dbaasv1alpha1.PlatformsName) *dbaasv1alpha1.PlatformStatus {
	for i := range platformsStatus {
		if platformsStatus[i].PlatformName == platformName {
			return &platformsStatus[i]
		}
	}

	return nil
}

// setStatusCondition sets the given condition with the given status,
// reason and message on a resource.
func setStatusCondition(Conditions *[]metav1.Condition, conditionType string, status metav1.ConditionStatus, reason, message string) {
	newCondition := metav1.Condition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	apimeta.SetStatusCondition(Conditions, newCondition)

}
