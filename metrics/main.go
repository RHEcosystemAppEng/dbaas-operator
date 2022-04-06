package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	v1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/collectors"
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/exporter"
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/handler"
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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

	v1alpha1.AddToScheme(scheme.Scheme)

	klog.Infof("API Server & kubeconfigpath %s", opts.KubeconfigPath)
	kubeconfig, err := clientcmd.BuildConfigFromFlags(opts.Apiserver, opts.KubeconfigPath)

	if err != nil {
		klog.Fatalf("failed to create cluster config: %v", err)
	}

	kubeconfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1.GroupName, Version: v1.Version}
	kubeconfig.APIPath = "/apis"
	kubeconfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	kubeconfig.UserAgent = rest.DefaultKubernetesUserAgent()

	opts.Kubeconfig = kubeconfig

	fmt.Print(opts.AllowedNamespaces)
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
