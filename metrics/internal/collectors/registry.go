package collectors

import (
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// name of the project/exporter
	namespace = "openshift-dbaas-operator"
)

// RegisterCustomResourceCollectors registers the custom resource collectors
// in the given prometheus.Registry
// This is used to expose metrics about the Custom Resources
func RegisterCustomResourceCollectors(registry *prometheus.Registry, opts *options.Options) {
	dbaasPlaatformStoreCollector := NewDbaasPlaatformStoreCollector(opts)
	dbaasPlaatformStoreCollector.Run(opts.StopCh)
	registry.MustRegister(
		dbaasPlaatformStoreCollector,
	)
}
