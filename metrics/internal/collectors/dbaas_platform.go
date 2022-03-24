package collectors

import (
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"

	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/klog"
)

var _ prometheus.Collector = &DbaasPlaatformStoreCollector{}

// DbaasPlaatformStoreCollector is a custom collector for DbaasPlaatformStoreCollector Custom Resource
type DbaasPlaatformStoreCollector struct {
	PlatformStatus    *prometheus.Desc
	AllowedNamespaces []string
}

// NewDbaasPlaatformStoreCollector constructs a collector
func NewDbaasPlaatformStoreCollector(opts *options.Options) *DbaasPlaatformStoreCollector {

	return &DbaasPlaatformStoreCollector{
		PlatformStatus: prometheus.NewDesc(
			prometheus.BuildFQName("dbaas", "platform", "status"), // Metrics object name with undersore
			`Health Status of Platfor. 0=Success, 1=Failure`,
			[]string{"dbaas_platform_status"}, // Metrics name
			nil,
		),
		AllowedNamespaces: opts.AllowedNamespaces,
	}
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
	klog.Infof("Inside collector to set values %s:%v", c.PlatformStatus, "running")
	ch <- prometheus.MustNewConstMetric(c.PlatformStatus,
		prometheus.GaugeValue, 0,
		"running") //Set the required value for the metrics
}
