package metrics

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	otelsdkmeter "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"time"
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

type Config struct {
	ShutdownTimeout time.Duration
}

type Metrics struct {
	exporter *otelprometheus.Exporter
	cfg      Config
}

func New(cfg Config) (*Metrics, error) {
	return &Metrics{
		cfg: cfg,
	}, nil
}

func (m *Metrics) Start() error {
	registrer := prometheus.
		WrapRegistererWith(globalLabels, prometheus.DefaultRegisterer)

	exporter, err := otelprometheus.New(otelprometheus.WithRegisterer(registrer))
	if err != nil {
		return fmt.Errorf("failed to create open telemetry prometheus exporter: %w", err)
	}

	m.exporter = exporter

	err = registerCollectors(
		registrer,
		HttpRequestsTotal,
		HttpRequestDuration,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *Metrics) Shutdown() error {
	if m.exporter == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), m.cfg.ShutdownTimeout)
	defer cancel()
	return m.exporter.Shutdown(ctx)
}

func registerCollectors(registerer prometheus.Registerer, collectors ...prometheus.Collector) error {
	for _, collector := range collectors {
		err := registerer.Register(collector)
		if err != nil {
			return fmt.Errorf("failed to register collector: %w", err)
		}
	}
	return nil
}

func SetupMeterProvider() (*otelsdkmeter.MeterProvider, error) {
	ctx := context.Background()
	exp, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithEndpoint("otel-collector:4318"),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics provider: %w", err)
	}
	meterProvider := otelsdkmeter.NewMeterProvider(
		otelsdkmeter.WithReader(otelsdkmeter.NewPeriodicReader(exp)),
		otelsdkmeter.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("demo-service"),
		)),
	)
	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}

var (
	testOrderMetricGlobal     metric.Int64ObservableGauge
	testTestOrderMetricGlobal metric.Int64Gauge
	testCounter               metric.Int64Counter
)

func InitCustomMetric() {
	metricProvider := otel.GetMeterProvider()
	meter := metricProvider.Meter("test")

	testOrderMetric, err := meter.Int64ObservableGauge("test_order_metric", metric.WithDescription("test desc"))
	if err != nil {
		panic(err)
	}

	fmt.Println("Observe???")
	_, err = meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		observer.ObserveInt64(testOrderMetric, -45)
		fmt.Println("Observerd!!!!")
		return nil
	}, testOrderMetric)
	if err != nil {
		panic(err)
	}

	testTestOrderMetricGlobal1, err := meter.Int64Gauge("test_non_observable")

	testTestOrderMetricGlobal = testTestOrderMetricGlobal1

	kv := attribute.KeyValue{Key: "topic", Value: attribute.StringValue("test")}

	testTestOrderMetricGlobal.Record(
		context.Background(),
		345,
		metric.WithAttributeSet(attribute.NewSet(kv)),
	)

	fmt.Println("Record!!!")
}

//func setupTracerProvider() *sdktrace.TracerProvider {
//	ctx := context.Background()
//	exp, err := otlptracehttp.New(ctx)
//	if err != nil {
//		log.Fatalf("failed to create trace exporter: %v", err)
//	}
//	tp := sdktrace.NewTracerProvider(
//		sdktrace.WithBatcher(exp),
//		sdktrace.WithResource(resource.NewWithAttributes(
//			semconv.SchemaURL,
//			semconv.ServiceName("demo-service"),
//		)),
//	)
//	return tp
//}
