package collectors

import (
	"context"
	"strings"

	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"

	"github.com/prometheus/client_golang/prometheus"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

var _ prometheus.Collector = &DbaasPlatformStoreCollector{}

// DbaasPlaatformStoreCollector is a custom collector for DbaasPlaatformStoreCollector Custom Resource
type DbaasPlatformStoreCollector struct {
	PlatformStatus    *prometheus.Desc
	AllowedNamespaces []string
	KubeconfigPath    string
	config            *rest.Config
}

// NewDbaasPlaatformStoreCollector constructs a collector
func NewDbaasPlatformStoreCollector(opts *options.Options) *DbaasPlatformStoreCollector {

	return &DbaasPlatformStoreCollector{
		PlatformStatus: prometheus.NewDesc(
			prometheus.BuildFQName("dbaas", "platform", "status"), // Metrics object name with undersore
			`Health Status of Platfor. 0=Success, 1=Failure`,
			[]string{"dbaas_platform_status"}, // Metrics name
			nil,
		),
		AllowedNamespaces: opts.AllowedNamespaces,
		KubeconfigPath:    opts.KubeconfigPath,
		config:            opts.Kubeconfig,
	}
}

// Describe implements prometheus.Collector interface
func (c *DbaasPlatformStoreCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.PlatformStatus,
	}

	for _, d := range ds {
		ch <- d
	}
}

// Collect implements prometheus.Collector interface
func (c *DbaasPlatformStoreCollector) Collect(ch chan<- prometheus.Metric) {

	klog.Infof("c.config Inside collector to set values %s", c.config)
	// creates the clientset
	clientset, _ := kubernetes.NewForConfig(c.config)
	klog.Infof("clientset Inside collector to set values %s", clientset)
	// access the API to list pods
	pods, _ := clientset.CoreV1().Pods("openshift-dbaas-operator").List(context.TODO(), v1.ListOptions{})
	dbaas_status := ""
	for _, pod := range pods.Items {
		klog.Infof("check pod name %s", pod.ObjectMeta.Name)
		if strings.Contains(pod.ObjectMeta.Name, "operator-controller-manager") {
			dbaas_status = string(pod.Status.Phase)
		}
	}

	ch <- prometheus.MustNewConstMetric(c.PlatformStatus,
		prometheus.GaugeValue, 0,
		dbaas_status) //Set the required value for the metrics
}
