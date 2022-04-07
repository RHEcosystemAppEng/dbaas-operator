package collectors

import (
	"fmt"
	"strconv"

	"github.com/RHEcosystemAppEng/dbaas-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/dbaas-operator/metrics/internal/options"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/prometheus/client_golang/prometheus"

	ctrls "github.com/RHEcosystemAppEng/dbaas-operator/controllers"
)

var _ prometheus.Collector = &DbaasInventoryStoreCollector{}

// DbaasInventoryStoreCollector is a custom collector for DbaasInventoryStoreCollector Custom Resource
type DbaasInventoryStoreCollector struct {
	InventoryCount    *prometheus.Desc
	AllowedNamespaces []string
	config            *rest.Config
}

// NewDbaasPlaatformStoreCollector constructs a collector
func NewDbaasInventoryStoreCollector(opts *options.Options) *DbaasInventoryStoreCollector {
	return &DbaasInventoryStoreCollector{
		InventoryCount: prometheus.NewDesc(
			prometheus.BuildFQName("dbaas", "provider", "count"), // Metrics object name with underscore
			`Number of provider count`,
			[]string{"dbaas_provider_count"}, // Metrics name
			nil,
		),
		AllowedNamespaces: opts.AllowedNamespaces,
		config:            opts.Kubeconfig,
	}
}

// Describe implements prometheus.Collector interface
func (c *DbaasInventoryStoreCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.InventoryCount,
	}

	for _, d := range ds {
		ch <- d
	}
}

type DBaaSPlatformReconciler struct {
	*ctrls.DBaaSPlatformReconciler
	DbaasInventoryStoreCollector *DbaasInventoryStoreCollector
}

// Collect implements prometheus.Collector interface
func (c *DbaasInventoryStoreCollector) Collect(ch chan<- prometheus.Metric) {

	clientSet, err := v1alpha1.NewConfigForInventory(c.config)
	if err != nil {
		panic(err)
	}

	projects, err := clientSet.DbaaSInventory("openshift-dbaas-operator", "dbaasinventories").List(v1.ListOptions{})
	if err != nil {
		fmt.Print(err)
	}

	for _, project := range projects.Items {
		fmt.Printf("projects inventory found: %+v\n", project.Status.Instances[len(project.Status.Instances)-1])
	}

	if err != nil {
		panic(err)
	}

	ch <- prometheus.MustNewConstMetric(c.InventoryCount,
		prometheus.GaugeValue, 0,
		strconv.Itoa(len(projects.Items))) //Set the required value for the metrics
}
