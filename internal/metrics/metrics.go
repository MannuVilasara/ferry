package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	RequestCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ferry_request_count_total",
		Help: "Total number of requests received",
	})

	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ferry_active_connections",
		Help: "Current number of active connections",
	})
)

func StartMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe(":9090", mux)
}
