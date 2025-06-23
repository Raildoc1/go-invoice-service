package meterutils

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelsdkmeter "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type OpenTelemetryConfig struct {
	ServiceName      string
	CollectorAddress string
}

func SetupMeterProvider(cfg OpenTelemetryConfig) (*otelsdkmeter.MeterProvider, error) {
	ctx := context.Background()
	exp, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithEndpoint(cfg.CollectorAddress),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics provider: %w", err)
	}
	meterProvider := otelsdkmeter.NewMeterProvider(
		otelsdkmeter.WithReader(otelsdkmeter.NewPeriodicReader(exp)),
		otelsdkmeter.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		)),
	)
	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}
