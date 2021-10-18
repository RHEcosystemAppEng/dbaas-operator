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

	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/console_plugin"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/crunchybridge_installation"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/csv"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/mongodb_atlas_instalation"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers/servicebinding"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dbaasv1alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
)

const (
	DBaaSPlatformFinalizer = "dbaasplaform-cleanup"
	RequeueDelaySuccess    = 10 * time.Second
	RequeueDelayError      = 5 * time.Second
)

// DBaaSPlatformReconciler reconciles a DBaaSPlatform object
type DBaaSPlatformReconciler struct {
	*DBaaSReconciler
	Log                 logr.Logger
	installComplete     bool
	operatorNameVersion string
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=dbaasplatforms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=dbaasplatforms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=dbaasplatforms/finalizers,verbs=update
//+kubebuilder:rbac:groups=operators.coreos.com,resources=catalogsources;subscriptions;operatorgroups;clusterserviceversions,verbs=get;list;create;update;watch
//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;statefulsets,verbs=get;list;create;update;watch
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;create;update;watch
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=get;list;create;update;watch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=consoles,verbs=get;list;update;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DBaaSPlatform object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSPlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	cr := &dbaasv1alpha1.DBaaSPlatform{}
	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			log.Info("DBaaSPlatform CR not found, has been deleted")
			return ctrl.Result{}, nil
		}
		// error fetching DBaaSPlatform instance, requeue and try again
		log.Error(err, "Error in Get of DBaaSPlatform CR")
		return ctrl.Result{}, err
	}
	// Add a cleanup finalizer if not already present
	if cr.DeletionTimestamp == nil && len(cr.Finalizers) == 0 {
		cr.Finalizers = append(cr.Finalizers, DBaaSPlatformFinalizer)
		err = r.Update(ctx, cr)
		return ctrl.Result{}, err
	}
	var finished = true

	var platforms []dbaasv1alpha1.PlatformsName

	if cr.DeletionTimestamp == nil {
		platforms = r.getInstallationPlatforms()

	} else {
		platforms = r.getCleanupPlatforms()
	}

	nextStatus := cr.Status.DeepCopy()

	for _, platform := range platforms {
		nextStatus.PlatformName = platform
		reconciler := r.getReconcilerForPlatform(platform)
		if reconciler != nil {
			var status dbaasv1alpha1.PlatformsInstlnStatus
			var err error

			if cr.DeletionTimestamp == nil {
				status, err = reconciler.Reconcile(ctx, cr, nextStatus)
			} else {
				status, err = reconciler.Cleanup(ctx, cr)
			}

			if err != nil {
				nextStatus.LastMessage = err.Error()
				return ctrl.Result{}, err
			} else {
				// Reset error message when everything went well
				nextStatus.LastMessage = ""
			}

			nextStatus.PlatformStatus = status

			// If a platform is not complete, do not continue with the next
			if status != dbaasv1alpha1.ResultSuccess {
				if cr.DeletionTimestamp == nil {
					log.Info("DBaaS platform stack install in progress", "working platform", platform)
				} else {
					log.Info("DBaaS platform stack cleanup in progress", "working platform", platform)
				}
				finished = false
				break
			}
		}
	}
	if cr.DeletionTimestamp == nil && finished && !r.installComplete {
		r.installComplete = true
		log.Info("DBaaS platform stack installation complete")
	}
	// Ready for deletion?
	// Only remove the finalizer when all platforms were successful
	if cr.DeletionTimestamp != nil && finished {
		log.Info("cleanup platforms complete, removing finalizer")
		cr.Finalizers = []string{}
		err = r.Update(ctx, cr)
		r.installComplete = false
		return ctrl.Result{}, err
	}

	return r.updateStatus(cr, nextStatus)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSPlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Creates a new managed install CR if it is not available
	kubeConfig := mgr.GetConfig()
	client, _ := k8sclient.New(kubeConfig, k8sclient.Options{
		Scheme: mgr.GetScheme(),
	})
	_, err := r.createPlatformCR(context.Background(), client)
	if err != nil {
		return err
	}

	// envVar set for all operators
	if operatorNameEnvVar, found := os.LookupEnv("OPERATOR_CONDITION_NAME"); !found {
		err := fmt.Errorf("OPERATOR_CONDITION_NAME must be set")
		return err
	} else {
		r.operatorNameVersion = operatorNameEnvVar
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
	//owner, _ := csv.GetDBaaSOperatorCSV(namespace, ctx, serverClient)
	var cr *dbaasv1alpha1.DBaaSPlatform
	if len(dbaaSPlatformList.Items) == 0 {

		cr = &dbaasv1alpha1.DBaaSPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-platform",
				Namespace: strings.TrimSpace(namespace),
				Labels:    map[string]string{"managed-by": "dbaas-operator"},
				// committing owner reference to avoid partially cleanup issue, to clean the platform remove the CR manually,
				/*OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         owner.APIVersion,
						Kind:               owner.Kind,
						UID:                owner.UID,
						Name:               owner.Name,
						Controller:         pointer.BoolPtr(true),
						BlockOwnerDeletion: pointer.BoolPtr(false),
					},
				},*/
			},

			Spec: dbaasv1alpha1.DBaaSPlatformSpec{
				Name: "Database as a Service",
			},
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

func (r *DBaaSPlatformReconciler) getInstallationPlatforms() []dbaasv1alpha1.PlatformsName {

	return []dbaasv1alpha1.PlatformsName{
		dbaasv1alpha1.CrunchyBridgeInstallation,
		dbaasv1alpha1.MongoDBAtlasInstallation,
		dbaasv1alpha1.DBassDynamicPluginInstallation,
		dbaasv1alpha1.ConsoleTelemetryPluginInstallation,
		dbaasv1alpha1.ServiceBindingInstallation,
	}

}
func (r *DBaaSPlatformReconciler) getCleanupPlatforms() []dbaasv1alpha1.PlatformsName {
	// since the operator cannot clean up resources automatically
	// disable the cleanup to reduce the permission the operator needs

	// will consider changing the scope of the dbaasplatform CR to be cluster-scoped
	// and use ownerreferences to support auto cleanup in the future
	return nil
	//return []dbaasv1alpha1.PlatformsName{
	//	dbaasv1alpha1.CrunchyBridgeInstallation,
	//	dbaasv1alpha1.MongoDBAtlasInstallation,
	//	dbaasv1alpha1.DBassDynamicPluginInstallation,
	//	dbaasv1alpha1.ConsoleTelemetryPluginInstallation,
	//	dbaasv1alpha1.ServiceBindingInstallation,
	//	dbaasv1alpha1.Csv,
	//}

}

func (r *DBaaSPlatformReconciler) getReconcilerForPlatform(provider dbaasv1alpha1.PlatformsName) reconcilers.PlatformReconciler {
	switch provider {
	case dbaasv1alpha1.CrunchyBridgeInstallation:
		return crunchybridge_installation.NewReconciler(r.Client, r.Scheme, r.Log)
	case dbaasv1alpha1.Csv:
		return csv.NewReconciler(r.Client, r.Log)
	case dbaasv1alpha1.MongoDBAtlasInstallation:
		return mongodb_atlas_instalation.NewReconciler(r.Client, r.Scheme, r.Log)
	case dbaasv1alpha1.DBassDynamicPluginInstallation:
		return console_plugin.NewReconciler(r.Client, r.Log,
			reconcilers.DBAAS_DYNAMIC_PLUGIN_NAME, r.InstallNamespace,
			reconcilers.DBAAS_DYNAMIC_PLUGIN_IMG, reconcilers.DBAAS_DYNAMIC_PLUGIN_DISPLAY_NAME,
			corev1.EnvVar{Name: reconcilers.DBAAS_OPERATOR_VERSION_KEY_ENV, Value: r.operatorNameVersion})
	case dbaasv1alpha1.ConsoleTelemetryPluginInstallation:
		return console_plugin.NewReconciler(r.Client, r.Log,
			reconcilers.CONSOLE_TELEMETRY_PLUGIN_NAME, r.InstallNamespace,
			reconcilers.CONSOLE_TELEMETRY_PLUGIN_IMG, reconcilers.CONSOLE_TELEMETRY_PLUGIN_DISPLAY_NAME,
			corev1.EnvVar{Name: reconcilers.CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY_ENV, Value: reconcilers.CONSOLE_TELEMETRY_PLUGIN_SEGMENT_KEY})
	case dbaasv1alpha1.ServiceBindingInstallation:
		return servicebinding.NewReconciler(r.Client, r.Scheme, r.Log)

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
