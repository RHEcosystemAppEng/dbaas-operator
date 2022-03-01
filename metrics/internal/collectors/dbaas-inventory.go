package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	InventoryCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "dbaas_inventory_count",
			Help: "Number of inventory present",
		},
	)
)

func SetInvetoryCount(isAdded bool) {
	// Code to fetch list of inventory and set the length
	// if isAdded {
	// 	InventoryCount.Inc()
	// } else {
	// 	InventoryCount.Dec()
	// }
}
