package collectors

import (
	"context"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"

	"github.com/prometheus/client_golang/prometheus"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var _ prometheus.Collector = &DbaasPlatformStoreCollector{}

// DbaasPlaatformStoreCollector is a custom collector for DbaasPlaatformStoreCollector Custom Resource
type DbaasPlatformStoreCollector struct {
	PlatformStatus    *prometheus.Desc
	AllowedNamespaces []string
	config            *rest.Config
}

// NewDbaasPlaatformStoreCollector constructs a collector
func NewDbaasPlatformStoreCollector(opts *options.Options) *DbaasPlatformStoreCollector {

	return &DbaasPlatformStoreCollector{
		PlatformStatus: prometheus.NewDesc(
			prometheus.BuildFQName("dbaas", "platform", "status"), // Metrics object name with undersore
			`Health Status of Platfor. 1=Success, 0=Failure`,
			[]string{"dbaas_platform_status"}, // Metrics name
			nil,
		),
		AllowedNamespaces: opts.AllowedNamespaces,
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

	dbaas_status := ""

	clientSet, err := v1alpha1.NewForConfig(c.config)
	if err != nil {
		panic(err)
	}

	projects, err := clientSet.DbaaSPlatform(c.AllowedNamespaces[0], "dbaasplatforms").List(v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, project := range projects.Items {
		dbaas_status = string(project.Status.PlatformStatus)
	}

	ch <- prometheus.MustNewConstMetric(c.PlatformStatus,
		prometheus.GaugeValue, 0,
		dbaas_status) //Set the required value for the metrics
}

func GetResourcesDynamically(dynamic dynamic.Interface, ctx context.Context,
	group string, version string, resource string, namespace string) (
	[]unstructured.Unstructured, error) {

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	list, err := dynamic.Resource(resourceId).Namespace(namespace).
		List(ctx, v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return list.Items, nil
}
