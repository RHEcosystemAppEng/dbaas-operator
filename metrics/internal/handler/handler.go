package handler

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricsPath = "/metrics"
	healthzPath = "/healthz"
)

// RegisterCustomResourceMuxHandlers registers the handlers needed to serve metrics
// about Custom Resources
func RegisterCustomResourceMuxHandlers(mux *http.ServeMux, customResourceRegistry *prometheus.Registry) {
	// Instrument metricsPath handler and register it inside the exporterRegistry
	metricsHandler := InstrumentMetricHandler(customResourceRegistry,
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
		Name: "dbaas_exporter_requests_total",
		Help: "Total number of scrapes.",
	}, []string{"code"})

	requestsInFlight := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dbaas_exporter_requests_in_flight",
		Help: "Current number of scrapes being served.",
	})

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "dbaas_exporter_request_duration_seconds",
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
