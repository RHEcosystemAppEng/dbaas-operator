package prometheus

import (
	"context"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"

	corev1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/controllers/reconcilers"
	"github.com/go-logr/logr"
	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	prometheusName = "dbaas-prometheus-operator"
	prometheusCSV  = "prometheusoperator.0.47.0"
	managedBy      = "app.kubernetes.io/managed-by"
	operatorName   = "dbaas-operator"

	prometheusInstance = "dbaas-prometheus"
	serviceMonitor     = "dbaas-service-monitor"
	serviceAccountName = "dbaas-prometheus-sa"
	roleName           = "dbaas-prometheus-role"
	roleBindingName    = "dbaas-prometheus-rolebinding"
)

type Reconciler struct {
	client              client.Client
	prometheus          *promv1.Prometheus
	cr                  *v1alpha1.DBaaSPlatform
	scheme              *runtime.Scheme
	log                 logr.Logger
	monitoringNamespace string
	operatorNamespace   string
}

func NewReconciler(cr *v1alpha1.DBaaSPlatform, client client.Client, scheme *runtime.Scheme, log logr.Logger, installNamespace string) reconcilers.PlatformReconciler {
	return &Reconciler{
		client:              client,
		prometheus:          &promv1.Prometheus{},
		cr:                  cr,
		scheme:              scheme,
		log:                 log,
		operatorNamespace:   installNamespace,
		monitoringNamespace: installNamespace + "-monitoring",
	}
}
func (r *Reconciler) Reconcile(ctx context.Context, cr *v1alpha1.DBaaSPlatform, status2 *v1alpha1.DBaaSPlatformStatus) (v1alpha1.PlatformsInstlnStatus, error) {

	status, err := r.reconcileNamespace(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileOperatorGroup(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileSubscription(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileServiceAccount(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileRole(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileRoleBinding(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcilePrometheus(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	status, err = r.reconcileServiceMonitor(ctx)
	if status != v1alpha1.ResultSuccess {
		return status, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) Cleanup(ctx context.Context, cr *v1alpha1.DBaaSPlatform) (v1alpha1.PlatformsInstlnStatus, error) {

	deployment := r.prometheus
	err := r.client.Delete(ctx, deployment)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileNamespace(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {

	key := types.NamespacedName{Name: r.monitoringNamespace}
	var namespace corev1.Namespace
	err := r.client.Get(ctx, key, &namespace)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	if errors.IsNotFound(err) {
		r.log.Info("Creating namespace")
		err = r.client.Create(ctx, newNamespace(r.monitoringNamespace))
		return v1alpha1.ResultFailed, err
	}

	// requeue if namespace is marked for deletion
	// TODO: decide if want to use finalizers to prevent deletion but
	// we also need to solve how to properly cleanup / uninstall operator
	if namespace.Status.Phase != corev1.NamespaceActive {
		r.log.Info("Namespace is present but not active", "phase", namespace.Status.Phase)
		return v1alpha1.ResultInProgress, nil
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileOperatorGroup(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	log := r.log.WithValues("Name", prometheusName)
	log.V(6).Info("Reconciling OperatorGroup")

	key := types.NamespacedName{
		Name:      prometheusName,
		Namespace: r.monitoringNamespace,
	}
	var operatorGroup operatorsv1.OperatorGroup

	err := r.client.Get(ctx, key, &operatorGroup)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	// create
	desired := newOperatorGroup(r.monitoringNamespace)
	if errors.IsNotFound(err) {
		log.Info("Creating OperatorGroup")
		err := r.client.Create(ctx, desired)
		return v1alpha1.ResultFailed, err
	}

	// update
	if !reflect.DeepEqual(operatorGroup.Spec, desired.Spec) {
		log.Info("Updating OperatorGroup")
		operatorGroup.Spec = desired.Spec
		err := r.client.Update(ctx, &operatorGroup)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileSubscription(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	log := r.log.WithValues("Name", prometheusName)
	key := types.NamespacedName{
		Name:      prometheusName,
		Namespace: r.monitoringNamespace,
	}
	var subscription corev1alpha1.Subscription
	err := r.client.Get(ctx, key, &subscription)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	// create
	desired := newSubscription(r.monitoringNamespace)
	if errors.IsNotFound(err) {
		log.Info("Creating Prometheus Operator Subscription")
		err := r.client.Create(ctx, desired)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
		return v1alpha1.ResultInProgress, nil
	}

	if subscription.Spec.StartingCSV == desired.Spec.StartingCSV {
		return v1alpha1.ResultSuccess, nil
	}

	r.log.WithValues("Name", subscription.Name).Info("Deleting Subscription")
	if err := r.client.Delete(ctx, &subscription); err != nil {
		return v1alpha1.ResultFailed, err
	}

	r.log.WithValues("Name", subscription.Status.InstalledCSV).Info("Deleting CSV")
	csv := corev1alpha1.ClusterServiceVersion{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1alpha1.SchemeGroupVersion.String(),
			Kind:       "ClusterServiceVersion",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      subscription.Status.InstalledCSV,
			Namespace: r.monitoringNamespace,
			Labels:    commonLabels(),
		},
	}
	if err := r.client.Delete(ctx, &csv); err != nil {
		return v1alpha1.ResultFailed, err
	}

	r.log.WithValues("Name", subscription.Name).Info("Creating Subscription")
	err = r.client.Create(ctx, &subscription)
	if err != nil {
		return v1alpha1.ResultFailed, err
	}
	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcilePrometheus(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	r.log.Info("Reconciling Prometheus")
	key := types.NamespacedName{
		Name:      prometheusInstance,
		Namespace: r.monitoringNamespace,
	}
	var prometheus promv1.Prometheus
	err := r.client.Get(ctx, key, &prometheus)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	desired := newPrometheus(r.monitoringNamespace)
	if errors.IsNotFound(err) {
		r.log.Info("Creating Prometheus")
		err := r.client.Create(ctx, desired)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
		return v1alpha1.ResultInProgress, nil
	}

	// update
	if !reflect.DeepEqual(prometheus.Spec, desired.Spec) {
		r.log.Info("Updating Prometheus")
		prometheus.Spec = desired.Spec
		err := r.client.Update(ctx, &prometheus)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileServiceMonitor(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	r.log.Info("Reconciling Service Monitor")
	key := types.NamespacedName{
		Name:      serviceMonitor,
		Namespace: r.monitoringNamespace,
	}
	var serviceMonitor promv1.ServiceMonitor
	err := r.client.Get(ctx, key, &serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	desired := newServiceMonitor(r.operatorNamespace, r.monitoringNamespace)
	if errors.IsNotFound(err) {
		r.log.Info("Creating Service Monitor")
		err := r.client.Create(ctx, desired)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
		return v1alpha1.ResultInProgress, nil
	}

	// update
	if !reflect.DeepEqual(serviceMonitor.Spec, desired.Spec) {
		r.log.Info("Updating Service Monitor")
		serviceMonitor.Spec = desired.Spec
		err := r.client.Update(ctx, &serviceMonitor)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileRole(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {
	r.log.Info("Reconciling role")

	key := types.NamespacedName{
		Name:      roleName,
		Namespace: r.operatorNamespace,
	}
	var role rbacv1.Role
	err := r.client.Get(ctx, key, &role)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	desired := newRole(r.operatorNamespace)
	if errors.IsNotFound(err) {
		r.log.Info("Creating Role")
		err := r.client.Create(ctx, desired)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
		return v1alpha1.ResultInProgress, nil
	}

	// update
	if !reflect.DeepEqual(role.Rules, desired.Rules) {
		r.log.Info("Updating Service Monitor")
		role.Rules = desired.Rules
		err := r.client.Update(ctx, &role)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileRoleBinding(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {

	r.log.Info("Reconciling role binding")

	key := types.NamespacedName{
		Name:      roleBindingName,
		Namespace: r.operatorNamespace,
	}
	var roleBinding rbacv1.RoleBinding
	err := r.client.Get(ctx, key, &roleBinding)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	desired := newRoleBinding(r.operatorNamespace, r.monitoringNamespace)
	if errors.IsNotFound(err) {
		r.log.Info("Creating RoleBinding")
		err := r.client.Create(ctx, desired)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
		return v1alpha1.ResultInProgress, nil
	}

	// update
	if !reflect.DeepEqual(roleBinding.RoleRef, desired.RoleRef) || !reflect.DeepEqual(roleBinding.Subjects, desired.Subjects) {
		r.log.Info("Updating Service Monitor")
		roleBinding.RoleRef = desired.RoleRef
		roleBinding.Subjects = desired.Subjects
		err := r.client.Update(ctx, &roleBinding)
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
	}

	return v1alpha1.ResultSuccess, nil
}

func (r *Reconciler) reconcileServiceAccount(ctx context.Context) (v1alpha1.PlatformsInstlnStatus, error) {

	r.log.Info("Reconciling service Account")

	key := types.NamespacedName{
		Name:      serviceAccountName,
		Namespace: r.monitoringNamespace,
	}

	var serviceAccount corev1.ServiceAccount
	err := r.client.Get(ctx, key, &serviceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return v1alpha1.ResultFailed, err
	}

	if errors.IsNotFound(err) {
		r.log.Info("Creating service Account")
		err := r.client.Create(ctx, newServiceAccount(r.monitoringNamespace))
		if err != nil {
			return v1alpha1.ResultFailed, err
		}
		return v1alpha1.ResultInProgress, nil
	}

	return v1alpha1.ResultSuccess, nil
}
