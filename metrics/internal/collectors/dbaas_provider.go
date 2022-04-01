package collectors

import (
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"

	"github.com/prometheus/client_golang/prometheus"

	ctrls "github.com/RHEcosystemAppEng/dbaas-operator/controllers"
)

var _ prometheus.Collector = &DbaasProviderStoreCollector{}

// DbaasProviderStoreCollector is a custom collector for DbaasProviderStoreCollector Custom Resource
type DbaasProviderStoreCollector struct {
	ProviderCount     *prometheus.Desc
	AllowedNamespaces []string
}

// NewDbaasPlaatformStoreCollector constructs a collector
func NewDbaasProviderStoreCollector(opts *options.Options) *DbaasProviderStoreCollector {
	return &DbaasProviderStoreCollector{
		ProviderCount: prometheus.NewDesc(
			prometheus.BuildFQName("dbaas", "provider", "count"), // Metrics object name with underscore
			`Number of provider count`,
			[]string{"dbaas_provider_count"}, // Metrics name
			nil,
		),
		AllowedNamespaces: opts.AllowedNamespaces,
	}
}

// Describe implements prometheus.Collector interface
func (c *DbaasProviderStoreCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.ProviderCount,
	}

	for _, d := range ds {
		ch <- d
	}
}

var sizeOfProvider int

type DBaaSPlatformReconciler struct {
	*ctrls.DBaaSPlatformReconciler
	dbaasProviderStoreCollector *DbaasProviderStoreCollector
}

// Collect implements prometheus.Collector interface
func (c *DbaasProviderStoreCollector) Collect(ch chan<- prometheus.Metric) {

	// klog.Infof("Inside collector to set values %s:%v", c.ProviderCount, 0)
	ch <- prometheus.MustNewConstMetric(c.ProviderCount,
		prometheus.GaugeValue, 0,
		"0") //Set the required value for the metrics
}
