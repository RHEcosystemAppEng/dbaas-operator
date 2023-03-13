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

package controllers

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/observability"

	metrics "github.com/RHEcosystemAppEng/dbaas-operator/controllers/metrics"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/consoleplugin"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/providersinstallation"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/quickstartinstallation"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/util"
	"github.com/go-logr/logr"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

/*
Platform resources to be created or updated explicitly by the DBaaS operator:
	Crunchy Bridge operator: CatalogSource, Subscription, OperatorGroup
	Service Binding operator: Subscription
	DBaaS Dynamic Plugin: Service, Deployment, ConsolePlugin, Console (updated)

Platform resources to be removed by the DBaaS operator by setting owner reference:
	Crunchy Bridge operator: Subscription, CSV
	Service Binding operator: Subscription, CSV
	DBaaS Dynamic Plugin: Service, Deployment

Platform resources NOT to be removed or updated by the DBaaS operator:
	Crunchy Bridge operator: CatalogSource (in different namespace), OperatorGroup (in different namespace)
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
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=monitoringstacks;servicemonitors,verbs=get;list;create;update;watch;delete
//+kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=infrastructures,verbs=get;list;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=consoles,verbs=get;list;watch
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=secrets;configmaps,verbs=get;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSPlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	execution := metrics.PlatformInstallStart()
	logger := log.FromContext(ctx)
	metricLabelErrCdValue := ""
	event := ""

	cr := &v1beta1.DBaaSPlatform{}
	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			logger.V(1).Info("DBaaSPlatform CR not found, has been deleted")
			metricLabelErrCdValue = metrics.LabelErrorCdValueResourceNotFound
			return ctrl.Result{}, nil
		}
		// error fetching DBaaSPlatform instance, requeue and try again
		logger.Error(err, "Error in Get of DBaaSPlatform CR")
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrorFetchingDBaaSProviderResources
		return ctrl.Result{}, err
	}

	// OCPBUGS-4991 - temporary fix until https://github.com/operator-framework/operator-lifecycle-manager/pull/2912 makes it to a release
	if err = r.fixConversionWebhooks(ctx); err != nil {
		logger.Error(err, "Error related to conversion webhook setup")
	}

	// temporary fix for ack-rds-controller v0.1.3 upgrade issue
	if err = r.fixRDSControllerUpgrade(ctx); err != nil {
		logger.Error(err, "Error related to ack-rds-controller v0.1.3 upgrade")
	}

	if cr.DeletionTimestamp != nil {
		event = metrics.LabelEventValueDelete
	} else {
		event = metrics.LabelEventValueCreate
	}

	defer func() {
		metrics.SetPlatformMetrics(*cr, cr.Name, execution, event, metricLabelErrCdValue)
	}()

	var finished = true

	var platforms map[v1beta1.PlatformName]v1beta1.PlatformConfig

	r.setOpenShiftInstallationInfo(ctx, err, logger, cr)

	if cr.DeletionTimestamp == nil {
		platforms = reconcilers.InstallationPlatforms
	}

	nextStatus := cr.Status.DeepCopy()
	nextPlatformStatus := v1beta1.PlatformStatus{}
	for platform, platformConfig := range platforms {
		nextPlatformStatus.PlatformName = platform
		reconciler := r.getReconcilerForPlatform(platformConfig)
		if reconciler != nil {
			var status v1beta1.PlatformInstlnStatus
			var err error

			if cr.DeletionTimestamp == nil {
				status, err = reconciler.Reconcile(ctx, cr)
				metrics.SetPlatformStatusMetric(platform, status, platformConfig.CSV)

			} else {
				status, err = reconciler.Cleanup(ctx, cr)
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
			if status != v1beta1.ResultSuccess {
				if cr.DeletionTimestamp == nil {
					metrics.PlatformStackInstallationMetric(cr, r.operatorNameVersion, execution)
					logger.Info("DBaaS platform stack install in progress", "working platform", platform)
					setStatusCondition(&nextStatus.Conditions, v1beta1.DBaaSPlatformReadyType, metav1.ConditionFalse, v1beta1.InstallationInprogress, "DBaaS platform stack install in progress")
				} else {
					logger.Info("DBaaS platform stack cleanup in progress", "working platform", platform)
					setStatusCondition(&nextStatus.Conditions, v1beta1.DBaaSPlatformReadyType, metav1.ConditionUnknown, v1beta1.InstallationCleanup, "DBaaS platform stack cleanup in progress")
				}
				finished = false
				break
			}
		}
	}
	if cr.DeletionTimestamp == nil && finished && !r.installComplete {
		r.installComplete = true
		setStatusCondition(&nextStatus.Conditions, v1beta1.DBaaSPlatformReadyType, metav1.ConditionTrue, v1beta1.Ready, "DBaaS platform stack installation complete")
		metrics.PlatformStackInstallationMetric(cr, r.operatorNameVersion, execution)
		logger.Info("DBaaS platform stack installation complete")
	}

	return r.updateStatus(cr, nextStatus)
}

//setOpenShiftInstallationInfo sets the metrics for dbaas_version_info
func (r *DBaaSPlatformReconciler) setOpenShiftInstallationInfo(ctx context.Context, err error, logger logr.Logger, cr *v1beta1.DBaaSPlatform) {
	consoleURL, err := util.GetOpenshiftConsoleURL(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Error in getting of openshift consoleURl")
	}

	platformType, err := util.GetOpenshiftPlatform(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Error in getting of openshift platform Type")
	}

	_, clusterVersion, err := util.GetClusterIDVersion(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Error in getting of openshift cluster Version ")
	}
	metrics.SetOpenShiftInstallationInfoMetric(r.operatorNameVersion, consoleURL, string(platformType), cr.CreationTimestamp.String(), clusterVersion)
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
	if err := r.prepareRDSController(context.Background(), client); err != nil {
		return err
	}
	if _, err := r.createPlatformCR(context.Background(), client); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.DBaaSPlatform{}).
		Complete(r)
}

func (r *DBaaSPlatformReconciler) createPlatformCR(ctx context.Context, serverClient k8sclient.Client) (*v1beta1.DBaaSPlatform, error) {
	namespace := r.InstallNamespace
	dbaaSPlatformList := &v1beta1.DBaaSPlatformList{}
	listOpts := []k8sclient.ListOption{
		k8sclient.InNamespace(namespace),
	}
	err := serverClient.List(ctx, dbaaSPlatformList, listOpts...)
	if err != nil {
		return nil, fmt.Errorf("could not get a list of dbaas platform intallation CR: %w", err)
	}

	var cr *v1beta1.DBaaSPlatform
	if len(dbaaSPlatformList.Items) == 0 {
		syncPeriod := 180
		cr = &v1beta1.DBaaSPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-platform",
				Namespace: strings.TrimSpace(namespace),
				Labels:    map[string]string{"managed-by": "dbaas-operator"},
			},
			Spec: v1beta1.DBaaSPlatformSpec{
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

func (r *DBaaSPlatformReconciler) getReconcilerForPlatform(platformConfig v1beta1.PlatformConfig) reconcilers.PlatformReconciler {
	switch platformConfig.Type {
	case v1beta1.TypeOperator:
		return providersinstallation.NewReconciler(r.Client, r.Scheme, r.Log, platformConfig)
	case v1beta1.TypeConsolePlugin:
		return consoleplugin.NewReconciler(r.Client, r.Scheme, r.Log, platformConfig)
	case v1beta1.TypeQuickStart:
		return quickstartinstallation.NewReconciler(r.Client, r.Scheme, r.Log)
	case v1beta1.TypeObservability:
		return observability.NewReconciler(r.Client, r.Scheme, r.Log)
	}

	return nil
}

func (r *DBaaSPlatformReconciler) updateStatus(cr *v1beta1.DBaaSPlatform, nextStatus *v1beta1.DBaaSPlatformStatus) (ctrl.Result, error) {
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
func setStatusPlatform(PlatformsStatus *[]v1beta1.PlatformStatus, newPlatformStatus v1beta1.PlatformStatus) {
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
func FindStatusPlatform(platformsStatus []v1beta1.PlatformStatus, platformName v1beta1.PlatformName) *v1beta1.PlatformStatus {
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

// Delete implements a handler for the Delete event.
func (r *DBaaSPlatformReconciler) Delete(e event.DeleteEvent) error {
	execution := metrics.PlatformInstallStart()
	metricLabelErrCdValue := ""
	log := ctrl.Log.WithName("DBaaSPlatformReconciler DeleteEvent")
	log.V(1).Info("Delete event started")

	platformObj, ok := e.Object.(*v1beta1.DBaaSPlatform)
	if !ok {
		log.Info("Error getting DBaaSPlatform object during delete")
		metricLabelErrCdValue = metrics.LabelErrorCdValueErrorDeletingPlatform
		return nil
	}
	log.V(1).Info("platformObj", "platformObj", objectKeyFromObject(platformObj))

	log.V(1).Info("Calling metrics for deleting of DBaaSProvider")
	metrics.SetPlatformMetrics(*platformObj, platformObj.Name, execution, metrics.LabelEventValueDelete, metricLabelErrCdValue)

	return nil
}

func (r *DBaaSPlatformReconciler) checkConversionWebhookHealth(ctx context.Context, webhooks []v1alpha1.WebhookDescription) (bool, error) {
	for _, webhook := range webhooks {
		for _, crdName := range webhook.ConversionCRDs {
			crd := &apiextensionsv1.CustomResourceDefinition{}
			if err := r.Get(ctx, types.NamespacedName{Name: crdName}, crd); err != nil {
				return false, err
			}
			if crd.Spec.Conversion == nil || crd.Spec.Conversion.Strategy != "Webhook" ||
				crd.Spec.Conversion.Webhook == nil || crd.Spec.Conversion.Webhook.ClientConfig == nil ||
				crd.Spec.Conversion.Webhook.ClientConfig.CABundle == nil {
				return false, nil
			}
		}
	}
	return true, nil
}

func (r *DBaaSPlatformReconciler) fixConversionWebhooks(ctx context.Context) error {
	owner, err := reconcilers.GetDBaaSOperatorCSV(ctx, r.InstallNamespace, r.operatorNameVersion, r.Client)
	if err != nil {
		return err
	}
	if owner.Status.Phase == v1alpha1.CSVPhaseSucceeded || owner.Status.Phase == v1alpha1.CSVPhaseInstalling {
		ok, err := r.checkConversionWebhookHealth(ctx, owner.Spec.WebhookDefinitions)
		if err != nil {
			return err
		}
		labelSelector := k8sclient.MatchingLabels{"olm.owner": r.operatorNameVersion}
		if !ok {
			vwebhooks := &admissionregistrationv1.ValidatingWebhookConfigurationList{}
			if err = r.List(ctx, vwebhooks, labelSelector); err != nil {
				return err
			}
			for i := range vwebhooks.Items {
				if vwebhooks.Items[i].CreationTimestamp.Before(owner.Status.LastUpdateTime) {
					if err = r.Client.Delete(ctx, &vwebhooks.Items[i]); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// Temporary solution to the rds-controller upgrade issue, will revert in the next release
func (r *DBaaSPlatformReconciler) prepareRDSController(ctx context.Context, cli k8sclient.Client) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ack-rds-user-secrets", //#nosec G101
			Namespace: r.DBaaSReconciler.InstallNamespace,
		},
	}
	if err := cli.Get(ctx, k8sclient.ObjectKeyFromObject(secret), secret); err != nil {
		if apierrors.IsNotFound(err) {
			secret.Data = map[string][]byte{
				"AWS_ACCESS_KEY_ID":     []byte("dummy"),
				"AWS_SECRET_ACCESS_KEY": []byte("dummy"), //#nosec G101
			}
			if err := cli.Create(ctx, secret); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ack-rds-user-config",
			Namespace: r.DBaaSReconciler.InstallNamespace,
		},
	}
	if err := cli.Get(ctx, k8sclient.ObjectKeyFromObject(cm), cm); err != nil {
		if apierrors.IsNotFound(err) {
			cm.Data = map[string]string{
				"AWS_REGION":                     "dummy",
				"AWS_ENDPOINT_URL":               "",
				"ACK_ENABLE_DEVELOPMENT_LOGGING": "false",
				"ACK_WATCH_NAMESPACE":            "",
				"ACK_LOG_LEVEL":                  "info",
				"ACK_RESOURCE_TAGS":              "dbaas",
			}
			if err := cli.Create(ctx, cm); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// Temporary solution to the rds-controller v0.1.3 upgrade issue
func (r *DBaaSPlatformReconciler) fixRDSControllerUpgrade(ctx context.Context) error {
	csv := &v1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ack-rds-controller.v0.1.3",
			Namespace: r.DBaaSReconciler.InstallNamespace,
		},
	}
	if err := r.Client.Get(ctx, k8sclient.ObjectKeyFromObject(csv), csv); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if csv.Status.Phase == v1alpha1.CSVPhaseFailed && csv.Status.Reason == v1alpha1.CSVReasonComponentFailed &&
		csv.Status.Message == "install strategy failed: Deployment.apps \"ack-rds-controller\" is invalid: spec.selector: "+
			"Invalid value: v1.LabelSelector{MatchLabels:map[string]string{\"app.kubernetes.io/name\":\"ack-rds-controller\"}, "+
			"MatchExpressions:[]v1.LabelSelectorRequirement(nil)}: field is immutable" {
		ackDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ack-rds-controller",
				Namespace: r.DBaaSReconciler.InstallNamespace,
			},
		}
		if err := r.Client.Delete(ctx, ackDeployment); err != nil {
			return err
		}
		log.FromContext(ctx).Info("Applied fix to the failed RDS controller v0.1.3 installation")
	}
	return nil
}
