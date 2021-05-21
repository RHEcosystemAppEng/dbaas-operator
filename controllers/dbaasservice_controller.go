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
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/models"
	"github.com/go-logr/logr"
	atlas "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	project2 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/project"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
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
//+kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasdatabaseusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasprojects,verbs=get;list;watch;create;update;patch;delete

// Reconcile processes a DBaasService resource to compare & align cluster vs. desired state
func (r *DBaaSServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dbaasservice", req.NamespacedName)

	// fetch DBaaSService object initiating reconcile
	log.Info("reconcile initiated #2", "object", req.String())
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
	atlasInstance := models.AtlasService(&dbaasService)
	err = r.Get(ctx, client.ObjectKeyFromObject(atlasInstance), atlasInstance)
	if err != nil {
		if apierrors.IsNotFound(err) {

			// matching AtlasService has not yet been created, make one with ownerReference to our resource
			atlasInstance = models.OwnedAtlasService(&dbaasService)
			_, err = controllerutil.CreateOrUpdate(ctx, r.Client, atlasInstance, func() error {
				atlasInstance.Spec = models.MutateAtlasServiceSpec(&dbaasService)
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

	// refetch object prior to status update in case modifications have occurred
	err = r.Get(ctx, req.NamespacedName, &dbaasService)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted since request queued, child objects getting GC'd, no requeue
			log.Info("DBaasService resource not found, has been deleted")
			return ctrl.Result{}, nil
		}
		// error fetching resource instance, requeue and try again
		r.Log.Error(err, "error fetching DBaasService prior to status update")
		return ctrl.Result{}, err
	}

	// matching AtlasService with ownerReference to our resource found, so it's been touched - update our status
	log.Info("AtlasService altered, syncing DBaaSService status")
	err = r.syncServiceStatuses(&dbaasService, atlasInstance)
	if err != nil {
		r.Log.Error(err, "error syncing service status")
		return ctrl.Result{}, err
	}
	if err = r.Status().Update(ctx, &dbaasService); err != nil {
		if apierrors.IsConflict(err) {
			r.Log.Info("conflict at status update, requeue to sort it out")
			return ctrl.Result{}, nil
		} else {
			r.Log.Error(err, "error saving modified DBaaSService status")
			return ctrl.Result{}, err
		}
	}

	// instances have been selected for import & added to our Spec, so we need to create DBaasConnection for each
	if dbaasService.Spec.Imports != nil {
		var importCluster *dbaasv1.DBaaSCluster
		var dbaasProject dbaasv1.DBaaSProject
		for _, importClusterId := range dbaasService.Spec.Imports {
		findCluster:
			for _, project := range dbaasService.Status.Projects {
				r.Log.Info("Looking at project", "project", project)
				for _, cluster := range project.Clusters {
					r.Log.Info("Looking at cluster", "cluster", cluster, "cluster.ID", cluster.ID, "importClusterId", importClusterId)
					if cluster.ID == importClusterId {
						importCluster = cluster.DeepCopy()
						dbaasProject = project
						r.Log.Info("Found the right one", "importCluster", importCluster, "projectName", dbaasProject.Name)
						break findCluster
					}
				}
			}

			if importCluster == nil {
				r.Log.Info("No cluster found for", "importClusterId", importClusterId)
			} else {
				r.Log.Info("Cluster found for", "importClusterId", importClusterId, "importCluster", importCluster)

				dbUser := &atlas.AtlasDatabaseUser{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("dbaas-user-%s", importCluster.ID),
						Namespace: atlasInstance.Namespace,
					},
				}
				if err := r.Get(ctx, client.ObjectKeyFromObject(dbUser), dbUser); err != nil {
					if !apierrors.IsNotFound(err) {
						r.Log.Error(err, "Failed to even try to read database user")
						return ctrl.Result{}, err
					}

					project := atlas.AtlasProject{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("project-%s", dbaasProject.ID),
							Namespace: atlasInstance.Namespace,
						},
						Spec: atlas.AtlasProjectSpec{
							Name:             dbaasProject.Name,
							ConnectionSecret: atlasInstance.Spec.ConnectionSecret,
							ProjectIPAccessList: []project2.IPAccessList{
								{
									CIDRBlock: "0.0.0.0/0",
								},
							},
						},
					}
					r.Log.Info("Will create an atlas project", "project", project)
					_, err = controllerutil.CreateOrUpdate(ctx, r.Client, &project, func() error {
						return nil
					})
					if err != nil {
						r.Log.Error(err, "error reading or creating project")
						return ctrl.Result{}, err
					}

					passwordSecret := &v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      dbUser.Name,
							Namespace: atlasInstance.Namespace,
						},
						Data: map[string][]byte{},
					}
					passwordSecret.Data = map[string][]byte{"password": []byte(rand.String(10))}
					r.Log.Info("Will create a secret for user", "password", passwordSecret.Data["password"])
					_, err = controllerutil.CreateOrUpdate(ctx, r.Client, passwordSecret, func() error {
						return nil
					})
					if err != nil {
						r.Log.Error(err, "error reading or creating secret")
						return ctrl.Result{}, err
					}

					dbUser = &atlas.AtlasDatabaseUser{
						ObjectMeta: metav1.ObjectMeta{
							Name:      dbUser.Name,
							Namespace: atlasInstance.Namespace,
						},
						Spec: atlas.AtlasDatabaseUserSpec{
							Project: atlas.ResourceRef{
								Name: project.Name,
							},
							DatabaseName: "admin",
							Username:     dbUser.Name,
							Roles: []atlas.RoleSpec{
								{
									RoleName:     "readWriteAnyDatabase",
									DatabaseName: "admin",
								},
							},
							Scopes: nil,
							PasswordSecret: &atlas.ResourceRef{
								Name: passwordSecret.Name,
							},
						},
					}
					r.Log.Info("Will create a database user", "databaseUser", *dbUser)
					if err := r.Create(ctx, dbUser); err != nil {
						r.Log.Error(err, "Failed to create database user")
						return ctrl.Result{}, err
					}
				}

				//Happy path, assuming database user was created, and not checking for demo:
				dbPasswordSecret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      dbUser.Spec.Username,
						Namespace: dbUser.Namespace,
					},
				}
				if err := r.Get(ctx, client.ObjectKeyFromObject(dbPasswordSecret), dbPasswordSecret); err != nil {
					r.Log.Error(err, "Failed to read DB user password secret")
					return ctrl.Result{}, err
				}
				importCluster.DatabaseUser = dbaasv1.DBaaSDatabaseUser{
					Name:     dbUser.Spec.Username,
					Password: dbPasswordSecret.Data["password"],
				}

				dbaasConnection := models.DBaaSConnection(&dbaasService)
				_, err = controllerutil.CreateOrUpdate(ctx, r.Client, dbaasConnection, func() error {
					dbaasConnection.Spec = dbaasv1.DBaaSConnectionSpec{
						Type:     "MongoDB",
						Provider: "dbaas-operator",
						Cluster:  importCluster,
					}
					return nil
				})
				if err != nil {
					r.Log.Error(err, "error creating new DBaaSConnection for import selection")
					return ctrl.Result{}, err
				}
			}
		}
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
				ConnectionString:  cluster.ConnectionString,
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
