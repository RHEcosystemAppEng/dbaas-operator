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
	"github.com/go-logr/logr"
	atlas "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ptr "k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	dbaasv1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// DBaaSServiceReconciler reconciles a DBaaSService object
type DBaaSServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=dbaasservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=dbaasservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.redhat.com,resources=dbaasservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasservices/status,verbs=get

// Reconcile processes a DBaasService resource to compare & align cluster vs. desired state
func (r *DBaaSServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dbaasservice", req.NamespacedName)

	// fetch DBaaSService object initiating reconcile
	log.Info("reconcile initiated", "object", req.String())
	var dbaasService dbaasv1.DBaaSService
	err := r.Get(ctx, req.NamespacedName, &dbaasService)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			log.Info("DBaasService resource not found, has been deleted")
			return ctrl.Result{}, nil
		}
		// error fetching resource instance, requeue and try again
		r.Log.Error(err, "error fetching DBaasService for reconcile")
		return ctrl.Result{}, err
	}

	// DBaaSService resource found, try and find matching AtlasService
	found := atlas.AtlasService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "atlas.mongodb.com/v1",
			Kind:       "AtlasService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("atlas-service-%s", dbaasService.UID),
			Namespace: dbaasService.Namespace,
		},
	}
	err = r.Get(ctx, client.ObjectKeyFromObject(&found), &found)
	if err != nil {
		if apierrors.IsNotFound(err) {

			// matching AtlasService has not yet been created, make one with ownerReference to our resource
			atlasInstance := atlas.AtlasService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("atlas-service-%s", dbaasService.UID),
					Namespace: dbaasService.Namespace,
					Labels: map[string]string{
						"owner-resource": dbaasService.Name,
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							UID:                dbaasService.GetUID(),
							APIVersion:         "dbaas.redhat.com/v1",
							BlockOwnerDeletion: ptr.BoolPtr(false),
							Controller:         ptr.BoolPtr(true),
							Kind:               "DBaaSService",
							Name:               dbaasService.Name,
						},
					},
				},
			}
			_, err = controllerutil.CreateOrUpdate(ctx, r.Client, &atlasInstance, func() error {
				// modifier callback, mutate spec to match expectation
				atlasInstance.Spec = atlas.AtlasServiceSpec{
					Name: "dbaas-operator-test",
					ConnectionSecret: &atlas.ResourceRef{
						Name: dbaasService.Spec.CredentialsSecretName,
					},
				}
				return nil
			})
			if err != nil {
				r.Log.Error(err, "error creating new matching AtlasService")
				return ctrl.Result{}, err
			}

			// corresponding AtlasService has been created, reconciliations can continue
			log.Info("new matching AtlasService created")
			return ctrl.Result{}, nil
		}
		// error fetching resource instance, requeue and try again
		r.Log.Error(err, "error fetching AtlasService, requeuing")
		return ctrl.Result{}, err
	}

	// matching AtlasService with ownerReference to our resource found, so it's been touched - update our status
	log.Info("AtlasService altered, syncing DBaaSService status")
	err = r.syncServiceStatuses(&dbaasService, &found)
	if err != nil {
		r.Log.Error(err, "error syncing service status")
		return ctrl.Result{}, err
	}

	if err = r.Status().Update(ctx, &dbaasService); err != nil {
		r.Log.Error(err, "error saving modified DBaaSService status")
		return ctrl.Result{}, err
	}

	// reconcile cycle complete
	return ctrl.Result{}, nil
}

func (r *DBaaSServiceReconciler) syncServiceStatuses(dbaas *dbaasv1.DBaaSService, atlas *atlas.AtlasService) error {
	// this is where future generic translation will be nice - manual delta work is not fun
	var serviceStatus dbaasv1.DBaaSServiceStatus
	var dbaasProjects []dbaasv1.DBaaSProject
	atlasProjects := atlas.Status.AtlasProjectServiceList

	for _, project := range atlasProjects {
		dbaasProject := dbaasv1.DBaaSProject{
			ID:       project.ID,
			Name:     project.Name,
			Clusters: []dbaasv1.DBaaSCluster{},
			Users:    []dbaasv1.DBaaSDatabaseUser{},
		}
		for _, cluster := range project.ClusterList {
			dbaasCluster := dbaasv1.DBaaSCluster{
				ID:                cluster.ID,
				Name:              cluster.Name,
				InstanceSizeName:  cluster.InstanceSizeName,
				CloudProviderName: cluster.ProviderName,
				CloudRegion:       cluster.RegionName,
			}
			dbaasProject.Clusters = append(dbaasProject.Clusters, dbaasCluster)
		}
		for _, user := range project.DBUserList {
			dbaasUser := dbaasv1.DBaaSDatabaseUser{
				Name: user,
			}
			dbaasProject.Users = append(dbaasProject.Users, dbaasUser)
		}
		dbaasProjects = append(dbaasProjects, dbaasProject)
	}
	serviceStatus.Projects = dbaasProjects
	serviceStatus.DeepCopyInto(&dbaas.Status)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbaasv1.DBaaSService{}).
		Owns(&atlas.AtlasService{}).
		Complete(r)
}
