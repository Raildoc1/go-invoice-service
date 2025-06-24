package metrics

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type MetricsCollector struct {
	kafkaTotalProduceMessages metric.Int64Counter
	kafkaTotalProducedBytes   metric.Int64Counter
}

func MustInitCustomMetric() *MetricsCollector {
	metricProvider := otel.GetMeterProvider()

	m := &MetricsCollector{}

	// Kafka.
	meter := metricProvider.Meter("kafka")

	m.kafkaTotalProduceMessages = must(
		meter.Int64Counter(
			"kafka_total_produce_messages",
			metric.WithDescription("Total Kafka produced messages"),
		),
	)

	m.kafkaTotalProducedBytes = must(
		meter.Int64Counter(
			"kafka_total_produce_bytes",
			metric.WithDescription("Total Kafka produced bytes"),
		),
	)

	return m
}

func must[TMetric any](res TMetric, err error) TMetric {
	if err != nil {
		panic(err)
	}
	return res
}

func (m *MetricsCollector) IncKafkaTotalProduceMessages(ctx context.Context, topic string) {
	attrSet := attribute.NewSet(
		attribute.KeyValue{Key: "topic", Value: attribute.StringValue(topic)},
	)
	m.kafkaTotalProduceMessages.Add(ctx, 1, metric.WithAttributeSet(attrSet))
}

func (m *MetricsCollector) IncKafkaTotalProducedBytes(ctx context.Context, topic string, bytesCount int64) {
	attrSet := attribute.NewSet(
		attribute.KeyValue{Key: "topic", Value: attribute.StringValue(topic)},
	)
	m.kafkaTotalProducedBytes.Add(ctx, bytesCount, metric.WithAttributeSet(attrSet))
}
