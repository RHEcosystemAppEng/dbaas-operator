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
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrometheusTemplate is the template that serves as the base for the prometheus deployed by the operator
var resourceSelector = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"app": "dbaas-prometheus",
	},
}

var PrometheusTemplate = promv1.Prometheus{
	ObjectMeta: metav1.ObjectMeta{
		Name:      prometheusName,
		Namespace: namespace,
	},
	Spec: promv1.PrometheusSpec{
		ServiceAccountName:     "prometheus-k8s",
		ServiceMonitorSelector: &resourceSelector,
		PodMonitorSelector:     &resourceSelector,
		RuleSelector:           &resourceSelector,
		EnableAdminAPI:         false,
	},
}

var ServiceMonitorTemplate = promv1.ServiceMonitor{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "dbaas-service-monitor",
		Namespace: namespace,
		Labels: map[string]string{
			"app.kubernetes.io/component": "dbaas-metric-exporter",
			"app.kubernetes.io/name":      "dbaas-metric-exporter",
			"app":                         "dbaas-prometheus",
		},
	},
	Spec: promv1.ServiceMonitorSpec{
		NamespaceSelector: promv1.NamespaceSelector{
			MatchNames: []string{namespace},
		},
		Selector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/component": "dbaas-metric-exporter",
				"app.kubernetes.io/name":      "dbaas-metric-exporter",
				"app":                         "dbaas-prometheus",
			},
		},
		Endpoints: []promv1.Endpoint{
			{
				Path:     "/metrics",
				Port:     "metrics",
				Interval: "1m",
			},
		},
	},
}
