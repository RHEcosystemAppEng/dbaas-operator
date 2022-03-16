package collectors

import (
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
)

const (
	// component within the project/exporter
	dbaasSubsystem = "openshift-dbaas-operator"
)

var _ prometheus.Collector = &DbaasPlaatformStoreCollector{}

// DbaasPlaatformStoreCollector is a custom collector for DbaasPlaatformStoreCollector Custom Resource
type DbaasPlaatformStoreCollector struct {
	PlatformStatus    *prometheus.Desc
	Informer          cache.SharedIndexInformer
	AllowedNamespaces []string
}

// NewCephObjectStoreCollector constructs a collector
func NewDbaasPlaatformStoreCollector(opts *options.Options) *DbaasPlaatformStoreCollector {

	sharedIndexInformer := DbaasPlatformStoreInformer(opts)
	if sharedIndexInformer == nil {
		return nil
	}

	return &DbaasPlaatformStoreCollector{
		PlatformStatus: prometheus.NewDesc(
			prometheus.BuildFQName("openshift-dbaas-operator", dbaasSubsystem, "health_status"),
			`Health Status of Platfor. 0=Success, 1=Failure`,
			[]string{"name", "namespace", "rgw_endpoint"},
			nil,
		),
		Informer:          sharedIndexInformer,
		AllowedNamespaces: opts.AllowedNamespaces,
	}
}

// Run starts CephObjectStore informer
func (c *DbaasPlaatformStoreCollector) Run(stopCh <-chan struct{}) {
	go c.Informer.Run(stopCh)
}

// Describe implements prometheus.Collector interface
func (c *DbaasPlaatformStoreCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.PlatformStatus,
	}

	for _, d := range ds {
		ch <- d
	}
}

// Collect implements prometheus.Collector interface
func (c *DbaasPlaatformStoreCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.PlatformStatus,
		prometheus.GaugeValue, 0,
		"mongo-db",
		"openshift-dbaas-operator")
}

// check with arun what to write here
func DbaasPlatformStoreInformer(opts *options.Options) cache.SharedIndexInformer {
	return nil
}
