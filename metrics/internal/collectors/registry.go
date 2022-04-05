package collectors

import (
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"
	"github.com/prometheus/client_golang/prometheus"
)

// RegisterCustomResourceCollectors registers the custom resource collectors
// in the given prometheus.Registry
// This is used to expose metrics about the Custom Resources
func RegisterCustomResourceCollectors(registry *prometheus.Registry, opts *options.Options) {
	dbaasPlaatformStoreCollector := NewDbaasPlatformStoreCollector(opts)
	registry.MustRegister(
		dbaasPlaatformStoreCollector,
	)
}
