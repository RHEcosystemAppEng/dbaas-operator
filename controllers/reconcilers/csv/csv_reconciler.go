package csv

import (
	"context"
	"strings"

	alpha1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"

	"github.com/go-logr/logr"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Reconciler struct {
	client client.Client
	scheme *runtime.Scheme
	logger logr.Logger
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, logger logr.Logger) reconcilers.PlatformReconciler {
	return &Reconciler{
		client: client,
		scheme: scheme,
		logger: logger,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, cr *alpha1.DBaaSPlatform, status *alpha1.DBaaSPlatformStatus) (alpha1.PlatformsInstlnStatus, error) {
	return alpha1.ResultSuccess, nil
}

func (r *Reconciler) Cleanup(ctx context.Context, cr *alpha1.DBaaSPlatform) (alpha1.PlatformsInstlnStatus, error) {

	list := &v1alpha1.ClusterServiceVersionList{}
	opts := &client.ListOptions{
		Namespace: cr.Namespace,
	}
	err := r.client.List(ctx, list, opts)
	if err != nil && !errors.IsNotFound(err) {
		return alpha1.ResultFailed, err
	}

	for c := range list.Items {
		csv := list.Items[c]
		if csv.Namespace == cr.Namespace && strings.HasPrefix(csv.Name, "crunchy-bridge-operator.") {
			err := r.client.Delete(ctx, &csv)
			if err != nil && !errors.IsNotFound(err) {
				return alpha1.ResultFailed, err
			}
			return alpha1.ResultInProgress, nil
		} else if csv.Namespace == cr.Namespace && strings.HasPrefix(csv.Name, "mongodb-atlas-kubernetes.") {
			err := r.client.Delete(ctx, &csv)
			if err != nil && !errors.IsNotFound(err) {
				return alpha1.ResultFailed, err
			}
			return alpha1.ResultInProgress, nil
		} else if csv.Namespace == cr.Namespace && strings.HasPrefix(csv.Name, "service-binding-operator.") {
			err := r.client.Delete(ctx, &csv)
			if err != nil && !errors.IsNotFound(err) {
				return alpha1.ResultFailed, err
			}
			return alpha1.ResultInProgress, nil
		}
	}

	return alpha1.ResultSuccess, nil
}

func CheckOwnerReferenceSet(cr *alpha1.DBaaSPlatform, csv *v1alpha1.ClusterServiceVersion, scheme *runtime.Scheme) (bool, error) {
	gvk, err := apiutil.GVKForObject(cr, scheme)
	if err != nil {
		return false, err
	}
	ref := metav1.OwnerReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       cr.GetName(),
		UID:        cr.GetUID(),
	}

	existing := metav1.GetControllerOf(csv)
	if existing == nil {
		return false, nil
	}

	refGV, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return false, err
	}
	existingGV, err := schema.ParseGroupVersion(existing.APIVersion)
	if err != nil {
		return false, err
	}
	equal := refGV.Group == existingGV.Group && ref.Kind == ref.Kind && ref.Name == existing.Name
	return equal, nil
}

func GetDBaaSOperatorCSV(namespace string, ctx context.Context, serverClient k8sclient.Client) (*v1alpha1.ClusterServiceVersion, error) {

	list := &v1alpha1.ClusterServiceVersionList{}
	opts := &client.ListOptions{
		Namespace: namespace,
	}
	err := serverClient.List(ctx, list, opts)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	for _, csv := range list.Items {
		// Operator CSV
		if csv.Namespace == namespace && strings.HasPrefix(csv.Name, "dbaas-operator.") {
			return &csv, nil

		}
	}

	return nil, nil
}
