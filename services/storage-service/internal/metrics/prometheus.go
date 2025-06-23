package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	serviceName = "validation-service"
)

var globalLabels = prometheus.Labels{
	"service": serviceName,
}

var (
	KafkaTotalConsumedMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_total_consumed_messages",
			Help: "Total Kafka consumed messages",
		},
		[]string{"topic", "consumer-group-id"},
	)

	TotalHandledInvoices = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_handled_invoices",
			Help: "Total handled invoices",
		},
		[]string{"status"},
	)
)

func MustInit() {
	prometheus.
		WrapRegistererWith(globalLabels, prometheus.DefaultRegisterer).
		MustRegister(KafkaTotalConsumedMessages, TotalHandledInvoices)
}
