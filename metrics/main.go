package main

import (
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/collectors"
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/exporter"
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/handler"
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func main() {

	opts := options.NewOptions()
	opts.AddFlags()
	// parses the flags and ExitOnError so errors can be ignored
	opts.Parse()
	if opts.Help {
		// prints usage messages for flags
		opts.Usage()
		os.Exit(0)
	}
	//klog.Infof("using options: %+v", opts)

	opts.StopCh = make(chan struct{})
	defer close(opts.StopCh)

	klog.Infof("API Server & kubeconfigpath %s:%v", opts.KubeconfigPath)
	kubeconfig, err := clientcmd.BuildConfigFromFlags(opts.Apiserver, opts.KubeconfigPath)

	if err != nil {
		klog.Fatalf("failed to create cluster config: %v", err)
	}
	opts.Kubeconfig = kubeconfig

	exporterRegistry := prometheus.NewRegistry()
	// Add exporter self metrics collectors to the registry.
	exporter.RegisterExporterCollectors(exporterRegistry)

	customResourceRegistry := prometheus.NewRegistry()
	// Add custom resource collectors to the registry.
	collectors.RegisterCustomResourceCollectors(customResourceRegistry, opts)

	// serves custom resources metrics
	customResourceMux := http.NewServeMux()
	handler.RegisterCustomResourceMuxHandlers(customResourceMux, customResourceRegistry, exporterRegistry)

	var rg run.Group
	rg.Add(listenAndServe(customResourceMux, opts.Host, opts.Port))

	klog.Infof("Running metrics server on %s:%v", opts.Host, opts.Port)
	err = rg.Run()
	if err != nil {
		klog.Fatalf("metrics and telemetry servers terminated: %v", err)
	}
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
