package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	serviceName = "message-scheduler-service"
)

var globalLabels = prometheus.Labels{
	"service": serviceName,
}

var (
	KafkaTotalProduceAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_total_produce_messages",
			Help: "Total Kafka produced messages",
		},
		[]string{"topic", "status"},
	)

	KafkaTotalProducedBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_total_produce_bytes",
			Help: "Total Kafka produced bytes",
		},
		[]string{"topic"},
	)
)

func MustInit() {
	prometheus.
		WrapRegistererWith(globalLabels, prometheus.DefaultRegisterer).
		MustRegister(KafkaTotalProduceAttempts, KafkaTotalProducedBytes)
}
