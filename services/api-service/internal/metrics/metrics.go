package metrics

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"time"
)

type MetricsCollector struct {
	httpRequestsTotal      metric.Int64Counter
	httpRequestDurationSec metric.Float64Histogram
}

func MustInitCustomMetric() *MetricsCollector {
	metricProvider := otel.GetMeterProvider()

	m := &MetricsCollector{}

	// HTTP.
	meter := metricProvider.Meter("http")

	m.httpRequestsTotal = must(
		meter.Int64Counter(
			"http_requests_total",
			metric.WithDescription("total number of HTTP requests"),
		),
	)

	m.httpRequestDurationSec = must(
		meter.Float64Histogram(
			"http_request_duration_seconds",
			metric.WithDescription("HTTP request durations"),
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

func (m *MetricsCollector) IncHTTPRequestsTotal(ctx context.Context, method, path string, status int) {
	attrSet := attribute.NewSet(
		attribute.KeyValue{Key: "method", Value: attribute.StringValue(method)},
		attribute.KeyValue{Key: "path", Value: attribute.StringValue(path)},
		attribute.KeyValue{Key: "status", Value: attribute.IntValue(status)},
	)
	m.httpRequestsTotal.Add(ctx, 1, metric.WithAttributeSet(attrSet))
}

func (m *MetricsCollector) RecordHTTPRequestDuration(ctx context.Context, method, path string, status int, duration time.Duration) {
	attrSet := attribute.NewSet(
		attribute.KeyValue{Key: "method", Value: attribute.StringValue(method)},
		attribute.KeyValue{Key: "path", Value: attribute.StringValue(path)},
		attribute.KeyValue{Key: "status", Value: attribute.IntValue(status)},
	)
	m.httpRequestDurationSec.Record(ctx, duration.Seconds(), metric.WithAttributeSet(attrSet))
}
