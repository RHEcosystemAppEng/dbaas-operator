package main

import (
	"net"
	"net/http"
	"strconv"

	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/handler"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
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
	handler.RegisterExporterMuxHandlers(exporterMux, exporterRegistry)

	customResourceMux := http.NewServeMux()
	handler.RegisterExporterMuxHandlers(exporterMux, exporterRegistry)

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
