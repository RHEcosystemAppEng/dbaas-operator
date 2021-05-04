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
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/models"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dbaasv1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1"
)

// DBaaSConnectionReconciler reconciles a DBaaSConnection object
type DBaaSConnectionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=dbaasconnections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=dbaasconnections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=dbaasconnections/finalizers,verbs=update
//+kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasservices/status,verbs=get
// +kubebuilder:rbac:groups="",resources=secrets;configmaps,verbs=get;list;create;update;delete;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;create;update;delete;watch

func (r *DBaaSConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dbaasconnection", req.NamespacedName)

	// fetch DBaaSConnection object initiating reconcile
	log.Info("reconcile initiated", "object", req.String())
	var dbaasConnection dbaasv1.DBaaSConnection
	err := r.Get(ctx, req.NamespacedName, &dbaasConnection)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			log.Info("DBaaSConnection resource not found, has been deleted")
			return ctrl.Result{}, nil
		} else {
			// error fetching resource instance, requeue and try again
			r.Log.Error(err, "error fetching DBaaSConnection for reconcile")
			return ctrl.Result{}, err
		}
	}

	// DBaaSConnection resource found, try and find matching bindable deployment
	existingDeployment := models.Deployment(&dbaasConnection)
	err = r.Get(ctx, client.ObjectKeyFromObject(existingDeployment), existingDeployment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// matching Deployment has not yet been created, make one with ownerReference
			newDeployment := models.OwnedDeployment(&dbaasConnection)
			_, err = controllerutil.CreateOrUpdate(ctx, r.Client, newDeployment, func() error {
				newDeployment.Spec = models.MutateDeploymentSpec()
				return nil
			})
			if err != nil {
				r.Log.Error(err, "error creating new connection ConfigMap")
				return ctrl.Result{}, err
			}
		} else {
			// error fetching resource instance, requeue and try again
			r.Log.Error(err, "error fetching matching Deployment, requeuing")
			return ctrl.Result{}, err
		}
	}

	// try and find matching ConfigMap
	existingConfigMap := models.ConfigMap(&dbaasConnection)
	err = r.Get(ctx, client.ObjectKeyFromObject(existingConfigMap), existingConfigMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// matching ConfigMap has not yet been created, make one with ownerReference
			connectionConfigMap := models.OwnedConfigMap(&dbaasConnection)
			_, err = controllerutil.CreateOrUpdate(ctx, r.Client, connectionConfigMap, func() error {
				connectionConfigMap.Data = models.MutateConfigMapData(&dbaasConnection)
				return nil
			})
			if err != nil {
				r.Log.Error(err, "error creating new connection ConfigMap")
				return ctrl.Result{}, err
			}
		} else {
			// error fetching resource instance, requeue and try again
			r.Log.Error(err, "error fetching matching ConfigMap, requeuing")
			return ctrl.Result{}, err
		}
	}

	// try and find matching Secret
	existingSecret := models.Secret(&dbaasConnection)
	err = r.Get(ctx, client.ObjectKeyFromObject(existingSecret), existingSecret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// matching Secret has not yet been created, make one with ownerReference
			connectionSecret := models.OwnedSecret(&dbaasConnection)
			_, err = controllerutil.CreateOrUpdate(ctx, r.Client, connectionSecret, func() error {
				connectionSecret.Data = map[string][]byte{
					"username": []byte(dbaasConnection.Spec.Cluster.DatabaseUser.Name),
					"password": dbaasConnection.Spec.Cluster.DatabaseUser.Password,
				}
				return nil
			})
			if err != nil {
				r.Log.Error(err, "error creating new connection Secret")
				return ctrl.Result{}, err
			}
		} else {
			// error fetching resource instance, requeue and try again
			r.Log.Error(err, "error fetching matching Secret, requeuing")
			return ctrl.Result{}, err
		}
	}

	// now that the ConfigMap and Secret exist, add them to status on our dbaasConnection
	log.Info("new connection ConfigMap & Secret created")
	newStatus := models.UpdatedConnectionStatus(&dbaasConnection)

	newStatus.DeepCopyInto(&dbaasConnection.Status)
	if err = r.Status().Update(ctx, &dbaasConnection); err != nil {
		r.Log.Error(err, "error saving modified DBaaSConnection status")
		return ctrl.Result{}, err
	}

	// corresponding ConfigMap & Secret have been made available for binding via status, reconciliations can continue
	log.Info("all DBaaSConnection dependencies created successfully")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbaasv1.DBaaSConnection{}).
		Owns(&v1.Secret{}).
		Owns(&v1.ConfigMap{}).
		Owns(&v12.Deployment{}).
		Complete(r)
}
