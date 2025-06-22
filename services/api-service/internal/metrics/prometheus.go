package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	serviceName = "api-service"
)

var globalLabels = prometheus.Labels{
	"service": serviceName,
}

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of HTTP requests received",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func MustInit() {
	prometheus.
		WrapRegistererWith(globalLabels, prometheus.DefaultRegisterer).
		MustRegister(HttpRequestsTotal, HttpRequestDuration)
}
