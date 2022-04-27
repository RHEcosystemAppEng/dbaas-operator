/*


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

package prometheus

import (
	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	corev1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrometheusTemplate is the template that serves as the base for the prometheus deployed by the operator
var resourceSelector = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app": "dbaas-prometheus",
	},
}

func newOperatorGroup(monitoringNamespace string) *operatorsv1.OperatorGroup {
	return &operatorsv1.OperatorGroup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: operatorsv1.SchemeGroupVersion.String(),
			Kind:       "OperatorGroup",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      prometheusName,
			Namespace: monitoringNamespace,
			Labels:    commonLabels(),
		},
		Spec: operatorsv1.OperatorGroupSpec{
			TargetNamespaces: []string{
				monitoringNamespace,
			},
		},
	}
}

func newNamespace(monitoringNamespace string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   monitoringNamespace,
			Labels: commonLabels(),
		},
	}
}

func newSubscription(monitoringNamespace string) *corev1alpha1.Subscription {
	return &corev1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1alpha1.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      prometheusName,
			Namespace: monitoringNamespace,
			Labels:    commonLabels(),
		},
		Spec: &corev1alpha1.SubscriptionSpec{
			CatalogSource:          "community-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			Package:                "prometheus",
			Channel:                "beta",
			InstallPlanApproval:    corev1alpha1.ApprovalAutomatic,
			StartingCSV:            prometheusCSV,
			Config: &corev1alpha1.SubscriptionConfig{
				Env: []corev1.EnvVar{
					{
						Name:  "DASHBOARD_NAMESPACES_ALL",
						Value: "true",
					},
				},
			},
		},
	}
}

func newPrometheus(monitoringNamespace string) *promv1.Prometheus {
	return &promv1.Prometheus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: promv1.SchemeGroupVersion.String(),
			Kind:       "Prometheus",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      prometheusInstance,
			Namespace: monitoringNamespace,
		},
		Spec: promv1.PrometheusSpec{
			ServiceAccountName:     serviceAccountName,
			ServiceMonitorSelector: &resourceSelector,
			EnableAdminAPI:         false,
		},
	}
}

func newServiceMonitor(operatorNamespace, monitoringNamespace string) *promv1.ServiceMonitor {
	return &promv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			APIVersion: promv1.SchemeGroupVersion.String(),
			Kind:       "ServiceMonitor",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceMonitor,
			Namespace: monitoringNamespace,
			Labels: map[string]string{
				"app": "dbaas-prometheus",
			},
		},
		Spec: promv1.ServiceMonitorSpec{
			NamespaceSelector: promv1.NamespaceSelector{
				MatchNames: []string{operatorNamespace},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "dbaas-prometheus",
				},
			},
			Endpoints: []promv1.Endpoint{
				{
					Path: "/metrics",
					Port: "metrics",
				},
			},
		},
	}
}

func newServiceAccount(monitoringNamespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: monitoringNamespace,
		},
	}
}

func newRole(operatorNamespace string) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: operatorNamespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
}

func newRoleBinding(operatorNamespace, monitoringNamespace string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: operatorNamespace,
		},
		RoleRef: rbacv1.RoleRef{
			Name:     roleName,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      serviceAccountName,
				Namespace: monitoringNamespace,
				Kind:      "ServiceAccount",
			},
		},
	}
}

func commonLabels() map[string]string {
	return map[string]string{
		managedBy: operatorName,
	}
}
