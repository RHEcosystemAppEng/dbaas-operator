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
	"strings"
	"time"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// DBaaSDefaultPolicyReconciler reconciles a default DBaaSPolicy object
type DBaaSDefaultPolicyReconciler struct {
	*DBaaSReconciler
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DBaaSDefaultPolicyReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	// Creates a new managed install CR if it is not available
	if err := r.createPlatformCR(context.Background()); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// on operator startup, create default policy if none exists
	return r.createDefaultPolicy(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBaaSDefaultPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// envVar set for all operators
	operatorNameEnvVar, found := os.LookupEnv("OPERATOR_CONDITION_NAME")
	if !found {
		err := fmt.Errorf("OPERATOR_CONDITION_NAME must be set")
		return err
	}
	r.operatorNameVersion = operatorNameEnvVar
	// watch deployments if installed to the operator's namespace
	return ctrl.NewControllerManagedBy(mgr).
		Named("defaultpolicy").
		For(
			&appsv1.Deployment{},
			builder.WithPredicates(r.ignoreOtherDeployments()),
			builder.OnlyMetadata,
		).
		Complete(r)
}

// only reconcile deployments which reside in the operator's install namespace, and only create events
func (r *DBaaSDefaultPolicyReconciler) ignoreOtherDeployments() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object.GetNamespace() == r.InstallNamespace
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// create a default Policy if one doesn't exist
func (r *DBaaSDefaultPolicyReconciler) createDefaultPolicy(ctx context.Context) (ctrl.Result, error) {
	logger := ctrl.LoggerFrom(ctx)

	defaultPolicy := getDefaultPolicy(r.InstallNamespace)

	// get list of DBaaSPolicies for install/default namespace
	policyList, err := r.policyListByNS(ctx, defaultPolicy.Namespace)
	if err != nil {
		logger.Error(err, "unable to list policies")
		return ctrl.Result{}, err
	}

	// if no default policy exists, create one
	if len(policyList.Items) == 0 {
		if err := r.Get(ctx, client.ObjectKeyFromObject(&defaultPolicy), &v1beta1.DBaaSPolicy{}); err != nil {
			// proceed only if default policy not found
			if errors.IsNotFound(err) {
				logger.Info("resource not found", "Name", defaultPolicy.Name)
				if err := r.Create(ctx, &defaultPolicy); err != nil {
					// trigger retry if creation of default policy fails
					logger.Error(err, "Error creating DBaaS Policy resource", "Name", defaultPolicy.Name)
					return ctrl.Result{RequeueAfter: time.Duration(30) * time.Second}, err
				}
				logger.Info("creating default DBaaS Policy resource", "Name", defaultPolicy.Name)
			} else {
				logger.Error(err, "Error getting the DBaaS Policy resource", "Name", defaultPolicy.Name)
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *DBaaSDefaultPolicyReconciler) createPlatformCR(ctx context.Context) error {
	logger := ctrl.LoggerFrom(ctx)
	namespace := r.InstallNamespace
	dbaaSPlatformList := &v1beta1.DBaaSPlatformList{}
	if err := r.List(ctx, dbaaSPlatformList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("could not get a list of dbaas platform intallation CR: %w", err)
	}
	if len(dbaaSPlatformList.Items) == 0 {
		syncPeriod := 180
		cr := &v1beta1.DBaaSPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dbaas-platform",
				Namespace: strings.TrimSpace(namespace),
				Labels:    map[string]string{"managed-by": "dbaas-operator"},
			},
			Spec: v1beta1.DBaaSPlatformSpec{
				SyncPeriod: &syncPeriod,
			},
		}
		if err := r.Get(ctx, client.ObjectKeyFromObject(cr), &v1beta1.DBaaSPlatform{}); err != nil {
			// proceed only if platform object not found
			if errors.IsNotFound(err) {
				owner, err := reconcilers.GetDBaaSOperatorCSV(ctx, namespace, r.operatorNameVersion, r.Client)
				if err != nil {
					return fmt.Errorf("could not create dbaas platform intallation CR: %w", err)
				}
				if err = ctrl.SetControllerReference(owner, cr, r.Scheme); err != nil {
					return fmt.Errorf("could not create dbaas platform intallation CR: %w", err)
				}
				logger.Info("resource not found", "Name", cr.Name)
				if err = r.Create(ctx, cr); err != nil {
					return fmt.Errorf("could not create  CR in %s namespace: %w", namespace, err)
				}
				logger.Info("creating DBaaSPlatform resource", "Name", cr.Name)
			} else {
				return err
			}
		}
	}

	return nil
}

func getDefaultPolicy(inventoryNamespace string) v1beta1.DBaaSPolicy {
	policy := v1beta1.DBaaSPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster",
			Namespace: inventoryNamespace,
		},
		Spec: v1beta1.DBaaSPolicySpec{
			DBaaSInventoryPolicy: v1beta1.DBaaSInventoryPolicy{
				Connections: v1beta1.DBaaSConnectionPolicy{
					Namespaces: &[]string{"*"},
				},
			},
		},
	}
	policy.SetGroupVersionKind(v1beta1.GroupVersion.WithKind("DBaaSPolicy"))
	return policy
}
