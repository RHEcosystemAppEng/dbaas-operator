package main

import (
	"net"
	"net/http"
	"strconv"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"
)

const (
	host                      = "0.0.0.0"
	customResourceMetricsPort = 8080
	exporterMetricsPort       = 8081
)

func main() {

	exporterRegistry := prometheus.NewRegistry()
	// Add exporter self metrics collectors to the registry.

	// serves exporter self metrics
	exporterMux := http.NewServeMux()
	RegisterExporterMuxHandlers(exporterMux, exporterRegistry)

	customResourceMux := http.NewServeMux()
	RegisterExporterMuxHandlers(exporterMux, exporterRegistry)

	var rg run.Group
	rg.Add(listenAndServe(exporterMux, host, exporterMetricsPort))
	rg.Add(listenAndServe(customResourceMux, host, customResourceMetricsPort))

}

func listenAndServe(mux *http.ServeMux, host string, port int) (func() error, func(error)) {
	var listener net.Listener
	serve := func() error {
		addr := net.JoinHostPort(host, strconv.Itoa(port))
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		return http.Serve(listener, mux)
	}
	cleanup := func(error) {
		err := listener.Close()
		if err != nil {
			klog.Errorf("failed to close listener: %v", err)
		}
	}
	return serve, cleanup
}

const (
	metricsPath = "/metrics"
	healthzPath = "/healthz"
)

// RegisterExporterMuxHandlers registers the handlers needed to serve the
// exporter self metrics
func RegisterExporterMuxHandlers(mux *http.ServeMux, exporterRegistry *prometheus.Registry) {
	metricsHandler := promhttp.HandlerFor(exporterRegistry, promhttp.HandlerOpts{})
	mux.Handle(metricsPath, metricsHandler)
}

// RegisterCustomResourceMuxHandlers registers the handlers needed to serve metrics
// about Custom Resources
func RegisterCustomResourceMuxHandlers(mux *http.ServeMux, customResourceRegistry *prometheus.Registry, exporterRegistry *prometheus.Registry) {
	// Instrument metricsPath handler and register it inside the exporterRegistry
	metricsHandler := InstrumentMetricHandler(exporterRegistry,
		promhttp.HandlerFor(customResourceRegistry, promhttp.HandlerOpts{}),
	)
	mux.Handle(metricsPath, metricsHandler)

	// Add healthzPath handler
	mux.HandleFunc(healthzPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// InstrumentMetricHandler is a middleware that wraps the provided http.Handler
// to observe requests sent to the exporter
func InstrumentMetricHandler(registry *prometheus.Registry, handler http.Handler) http.Handler {
	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dbass_requests_total",
		Help: "Total number of scrapes.",
	}, []string{"code"})

	requestsInFlight := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dbass_requests_in_flight",
		Help: "Current number of scrapes being served.",
	})

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "dbass_request_duration_seconds",
		Help: "Duration of all scrapes.",
	}, []string{"code"})

	registry.MustRegister(
		requestsTotal,
		requestsInFlight,
		requestDuration,
	)

	return promhttp.InstrumentHandlerDuration(
		requestDuration,
		promhttp.InstrumentHandlerInFlight(requestsInFlight,
			promhttp.InstrumentHandlerCounter(requestsTotal, handler),
		),
	)
}
