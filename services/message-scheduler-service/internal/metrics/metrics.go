package metrics

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	kafkaTotalProduceMessages metric.Int64Counter
	kafkaTotalProducedBytes   metric.Int64Counter
)

func MustInitCustomMetric() {
	metricProvider := otel.GetMeterProvider()

	// Kafka.
	meter := metricProvider.Meter("kafka")

	kafkaTotalProduceMessages = must(
		meter.Int64Counter(
			"kafka_total_produce_messages",
			metric.WithDescription("Total Kafka produced messages"),
		),
	)

	kafkaTotalProducedBytes = must(
		meter.Int64Counter(
			"kafka_total_produce_bytes",
			metric.WithDescription("Total Kafka produced bytes"),
		),
	)
}

func must[TMetric any](res TMetric, err error) TMetric {
	if err != nil {
		panic(err)
	}
	return res
}

func IncKafkaTotalProduceMessages(ctx context.Context, topic string) {
	attrSet := attribute.NewSet(
		attribute.KeyValue{Key: "topic", Value: attribute.StringValue(topic)},
	)
	kafkaTotalProduceMessages.Add(ctx, 1, metric.WithAttributeSet(attrSet))
}

func IncKafkaTotalProducedBytes(ctx context.Context, topic string, bytesCount int64) {
	attrSet := attribute.NewSet(
		attribute.KeyValue{Key: "topic", Value: attribute.StringValue(topic)},
	)
	kafkaTotalProducedBytes.Add(ctx, bytesCount, metric.WithAttributeSet(attrSet))
}
