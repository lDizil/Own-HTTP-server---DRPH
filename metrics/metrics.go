package metrics

import "github.com/prometheus/client_golang/prometheus"

var HttpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total http request",
	},
	[]string{"method", "path", "status"},
)

var HttpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "request_duration_seconds",
		Help: "Request duration in seconds ",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path"},
)

var HttpActiveConn = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "active_conn",
		Help: "number of active conn at every moment",
	},
)

func init() {
    prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(HttpActiveConn)
}