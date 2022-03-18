package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// InstallNamespaceEnvVar is the constant for env variable INSTALL_NAMESPACE
var (
	InstallNamespaceEnvVar = "INSTALL_NAMESPACE"
	typeLabelValue         = "credentials"
	typeLabelKeyMongo      = "atlas.mongodb.com/type"
)

type DBaaSReconciler struct {
	client.Client
	*runtime.Scheme
	InstallNamespace string
}

func (r *DBaaSReconciler) getDBaaSProvider(providerName string, ctx context.Context) (*v1alpha1.DBaaSProvider, error) {
	provider := &v1alpha1.DBaaSProvider{}
	if err := r.Get(ctx, types.NamespacedName{Name: providerName}, provider); err != nil {
		return nil, err
	}
	return provider, nil
}

func (r *DBaaSReconciler) watchDBaaSProviderObject(ctrl controller.Controller, object runtime.Object, providerObjectKind string) error {
	providerObject := unstructured.Unstructured{}
	providerObject.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
		Kind:    providerObjectKind,
	})
	err := ctrl.Watch(
		&source.Kind{
			Type: &providerObject,
		},
		&handler.EnqueueRequestForOwner{
			OwnerType:    object,
			IsController: true,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *DBaaSReconciler) createProviderObject(object client.Object, providerObjectKind string) *unstructured.Unstructured {
	var providerObject unstructured.Unstructured
	providerObject.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
		Kind:    providerObjectKind,
	})
	providerObject.SetNamespace(object.GetNamespace())
	providerObject.SetName(object.GetName())
	return &providerObject
}

func (r *DBaaSReconciler) providerObjectMutateFn(object client.Object, providerObject *unstructured.Unstructured, spec interface{}) controllerutil.MutateFn {
	return func() error {
		providerObject.UnstructuredContent()["spec"] = spec
		providerObject.SetOwnerReferences(nil)
		if err := ctrl.SetControllerReference(object, providerObject, r.Scheme); err != nil {
			return err
		}
		return nil
	}
}

func (r *DBaaSReconciler) parseProviderObject(unstructured *unstructured.Unstructured, object interface{}) error {
	b, err := unstructured.MarshalJSON()
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, object)
	if err != nil {
		return err
	}
	return nil
}

// check if namespace is a valid connection namespace
func (r *DBaaSReconciler) isValidConnectionNS(ctx context.Context, namespace string, inventory *v1alpha1.DBaaSInventory) (bool, error) {
	// valid if in same namespace as inventory
	if namespace == inventory.Namespace {
		return true, nil
	}
	// valid if all namespaces are supported via wildcard
	if contains(inventory.Spec.ConnectionNamespaces, "*") {
		return true, nil
	}
	return contains(inventory.Spec.ConnectionNamespaces, namespace), nil
}

func (r *DBaaSReconciler) reconcileProviderResource(providerName string, DBaaSObject client.Object,
	providerObjectKindFn func(*v1alpha1.DBaaSProvider) string, DBaaSObjectSpecFn func() interface{},
	providerObjectFn func() interface{}, DBaaSObjectSyncStatusFn func(interface{}) metav1.Condition,
	DBaaSObjectConditionsFn func() *[]metav1.Condition, DBaaSObjectReadyType string,
	ctx context.Context, logger logr.Logger) (result ctrl.Result, recErr error) {

	var condition *metav1.Condition
	if cond := apimeta.FindStatusCondition(*DBaaSObjectConditionsFn(), DBaaSObjectReadyType); cond != nil {
		condition = cond.DeepCopy()
	} else {
		condition = &metav1.Condition{
			Type:    DBaaSObjectReadyType,
			Status:  metav1.ConditionFalse,
			Reason:  v1alpha1.ProviderReconcileInprogress,
			Message: v1alpha1.MsgProviderCRReconcileInProgress,
		}
	}

	// This update will make sure the status is always updated in case of any errors or successful result
	defer func(cond *metav1.Condition) {
		apimeta.SetStatusCondition(DBaaSObjectConditionsFn(), *cond)
		if err := r.Client.Status().Update(ctx, DBaaSObject); err != nil {
			if errors.IsConflict(err) {
				logger.V(1).Info("DBaaS object modified, retry syncing status", "DBaaS Object", DBaaSObject)
				// Re-queue and preserve existing recErr
				result = ctrl.Result{Requeue: true}
				return
			}
			logger.Error(err, "Error updating the DBaaS resource status", "DBaaS Object", DBaaSObject)
			if recErr == nil {
				// There is no existing recErr. Set it to the status update error
				recErr = err
			}
		}
	}(condition)

	provider, err := r.getDBaaSProvider(providerName, ctx)
	if err != nil {
		recErr = err
		if errors.IsNotFound(err) {
			logger.Error(err, "Requested DBaaS Provider is not configured in this environment", "DBaaS Provider", providerName)
			*condition = metav1.Condition{Type: DBaaSObjectReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.DBaaSProviderNotFound, Message: err.Error()}
			return
		}
		logger.Error(err, "Error reading configured DBaaS Provider", "DBaaS Provider", providerName)
		return
	}
	logger.Info("Found DBaaS Provider", "DBaaS Provider", providerName)

	providerObject := r.createProviderObject(DBaaSObject, providerObjectKindFn(provider))
	if res, err := controllerutil.CreateOrUpdate(ctx, r.Client, providerObject, r.providerObjectMutateFn(DBaaSObject, providerObject, DBaaSObjectSpecFn())); err != nil {
		if errors.IsConflict(err) {
			logger.V(1).Info("Provider object modified, retry syncing spec", "Provider Object", providerObject)
			result = ctrl.Result{Requeue: true}
			return
		}
		logger.Error(err, "Error reconciling the Provider resource", "Provider Object", providerObject)
		recErr = err
		return
	} else if res != controllerutil.OperationResultNone {
		logger.Info("Provider resource reconciled", "Provider Object", providerObject, "result", res)
	}

	DBaaSProviderObject := providerObjectFn()
	if err := r.parseProviderObject(providerObject, DBaaSProviderObject); err != nil {
		logger.Error(err, "Error parsing the Provider object", "Provider Object", providerObject)
		*condition = metav1.Condition{Type: DBaaSObjectReadyType, Status: metav1.ConditionFalse, Reason: v1alpha1.ProviderParsingError, Message: err.Error()}
		recErr = err
		return
	}

	*condition = DBaaSObjectSyncStatusFn(DBaaSProviderObject)
	return
}

func (r *DBaaSReconciler) checkInventory(inventoryRef v1alpha1.NamespacedName, DBaaSObject client.Object,
	conditionFn func(string, string), ctx context.Context, logger logr.Logger) (inventory *v1alpha1.DBaaSInventory, validNS bool, err error) {
	inventory = &v1alpha1.DBaaSInventory{}
	if err = r.Get(ctx, types.NamespacedName{Namespace: inventoryRef.Namespace, Name: inventoryRef.Name}, inventory); err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "DBaaS Inventory resource not found for DBaaS Object", "DBaaS Object", DBaaSObject, "DBaaS Inventory", inventoryRef)
			conditionFn(v1alpha1.DBaaSInventoryNotFound, err.Error())
			if errCond := r.Client.Status().Update(ctx, DBaaSObject); errCond != nil {
				if errors.IsConflict(errCond) {
					logger.V(1).Info("DBaaS Object modified", "DBaaS Object", DBaaSObject)
				} else {
					logger.Error(errCond, "Error updating the DBaaS Object status", "DBaaS Object", DBaaSObject)
				}
			}
			return
		}
		logger.Error(err, "Error fetching DBaaS Inventory resource reference for DBaaS Object", "DBaaS Object", DBaaSObject, "DBaaS Inventory", inventoryRef)
		return
	}

	validNS, err = r.isValidConnectionNS(ctx, DBaaSObject.GetNamespace(), inventory)
	if err != nil {
		return inventory, validNS, err
	}
	if validNS {
		// The inventory must be in ready status before we can move on
		invCond := apimeta.FindStatusCondition(inventory.Status.Conditions, v1alpha1.DBaaSInventoryReadyType)
		if invCond == nil || invCond.Status == metav1.ConditionFalse {
			err = fmt.Errorf("inventory %v is not ready", inventoryRef)
			logger.Error(err, "Inventory is not ready", "Inventory", inventory.Name, "Namespace", inventory.Namespace)
			conditionFn(v1alpha1.DBaaSInventoryNotReady, v1alpha1.MsgInventoryNotReady)
		} else {
			return
		}
	} else {
		conditionFn(v1alpha1.DBaaSInvalidNamespace, v1alpha1.MsgInvalidNamespace)
	}

	if errCond := r.Client.Status().Update(ctx, DBaaSObject); errCond != nil {
		if errors.IsConflict(errCond) {
			logger.V(1).Info("DBaaS Object modified", "DBaaS Object", DBaaSObject)
		} else {
			logger.Error(errCond, "Error updating the DBaaS Object resource status", "DBaaS Object", DBaaSObject)
		}
	}

	return
}

func (r *DBaaSReconciler) checkCredsRefLabel(ctx context.Context, inventory v1alpha1.DBaaSInventory) error {
	if strings.Contains(inventory.Spec.ProviderRef.Name, "mongodb") &&
		inventory.Spec.CredentialsRef != nil && len(inventory.Spec.CredentialsRef.Name) != 0 {
		namespace := inventory.Spec.CredentialsRef.Namespace
		if len(namespace) == 0 {
			namespace = inventory.Namespace
		}
		secret := corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      inventory.Spec.CredentialsRef.Name,
			Namespace: namespace,
		}, &secret); err != nil {
			return err
		}
		if secret.GetLabels()[typeLabelKeyMongo] != typeLabelValue {
			patchBytes, err := json.Marshal(corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						typeLabelKeyMongo: typeLabelValue,
					},
				},
			})
			if err != nil {
				return err
			}
			if err := r.Patch(ctx, &secret, client.RawPatch(types.StrategicMergePatchType, patchBytes)); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetInstallNamespace returns the operator's install Namespace
func GetInstallNamespace() (string, error) {
	ns, found := os.LookupEnv(InstallNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", InstallNamespaceEnvVar)
	}
	return ns, nil
}

// checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
