package metrics

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type MetricsCollector struct {
	kafkaTotalConsumedMessages metric.Int64Counter
	totalHandledInvoices       metric.Int64Counter
}

func MustInitCustomMetric() *MetricsCollector {
	metricProvider := otel.GetMeterProvider()

	m := &MetricsCollector{}

	// Kafka.
	kafkaMeter := metricProvider.Meter("kafka")

	m.kafkaTotalConsumedMessages = must(
		kafkaMeter.Int64Counter(
			"kafka_total_consumed_messages",
			metric.WithDescription("Total Kafka consumed messages"),
		),
	)

	// Invoices.
	invoicesMeter := metricProvider.Meter("invoices")

	m.totalHandledInvoices = must(
		invoicesMeter.Int64Counter(
			"total_handled_invoices",
			metric.WithDescription("Total handled invoices"),
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

func (m *MetricsCollector) IncKafkaTotalConsumedMessages(ctx context.Context, topic, consumerGroupID string) {
	attrSet := attribute.NewSet(
		attribute.KeyValue{Key: "topic", Value: attribute.StringValue(topic)},
		attribute.KeyValue{Key: "consumer-group-id", Value: attribute.StringValue(consumerGroupID)},
	)
	m.kafkaTotalConsumedMessages.Add(ctx, 1, metric.WithAttributeSet(attrSet))
}

func (m *MetricsCollector) IncTotalHandledInvoices(ctx context.Context, status string) {
	attrSet := attribute.NewSet(
		attribute.KeyValue{Key: "status", Value: attribute.StringValue(status)},
	)
	m.totalHandledInvoices.Add(ctx, 1, metric.WithAttributeSet(attrSet))
}
