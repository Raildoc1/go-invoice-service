package config

import (
	"errors"
	"flag"
	"fmt"
	"go-invoice-service/api-service/internal/httpserver"
	"go-invoice-service/api-service/internal/services"
	"go-invoice-service/common/pkg/flagtypes"
	"go-invoice-service/common/pkg/jwtfactory"
	"go-invoice-service/common/pkg/meterutils"
	"os"
	"strconv"
	"time"
)

const (
	httpAddressFlag          = "http-address"
	httpAddressEnv           = "HTTP_ADDRESS"
	storageAddressFlag       = "storage-address"
	storageAddressEnv        = "STORAGE_ADDRESS"
	jwtPrivateKeyFlag        = "jwt-private-key"
	jwtPrivateKeyEnv         = "JWT_PRIVATE_KEY"
	prometheusPortFlag       = "prometheus-port"
	prometheusPortEnv        = "PROMETHEUS_PORT"
	otelCollectorAddressFlag = "otel-collector-address"
	otelCollectorAddressEnv  = "OTEL_COLLECTOR_ADDRESS"
)

const (
	defaultHTTPAddress          = "localhost:8080"
	defaultStorageAddress       = "localhost:5000"
	defaultJWTPrivateKey        = "private-key"
	defaultShutdownTimeout      = 5 * time.Second
	defaultPrometheusPort       = 9090
	defaultOtelCollectorAddress = "localhost:4318"
)

var defaultRetryAttempts = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	StorageConfig       services.StorageConfig
	JWTConfig           jwtfactory.Config
	HTTPServerConfig    httpserver.Config
	PrometheusConfig    meterutils.PrometheusConfig
	OpenTelemetryConfig meterutils.OpenTelemetryConfig
	ShutdownTimeout     time.Duration
}

func Load() (*Config, error) {

	httpAddress := defaultHTTPAddress
	storageAddress := defaultStorageAddress
	jwtPrivateKey := defaultJWTPrivateKey
	prometheusPort := defaultPrometheusPort
	otelCollectorAddress := defaultOtelCollectorAddress

	// Flags Definition.

	httpAddressFlagVal := flagtypes.NewString()
	flag.Var(httpAddressFlagVal, httpAddressFlag, "HTTP server address")

	storageAddressFlagVal := flagtypes.NewString()
	flag.Var(storageAddressFlagVal, storageAddressFlag, "Storage server address")

	jwtPrivateKeyFlagVal := flagtypes.NewString()
	flag.Var(jwtPrivateKeyFlagVal, jwtPrivateKeyFlag, "JWT private key")

	prometheusPortFlagVal := flagtypes.NewInt()
	flag.Var(prometheusPortFlagVal, prometheusPortFlag, "Prometheus port")

	otelCollectorAddressFlagVal := flagtypes.NewString()
	flag.Var(otelCollectorAddressFlagVal, otelCollectorAddressFlag, "OpenTelemetry Collector address")

	flag.Parse()

	// Flags Parse.

	if val, ok := httpAddressFlagVal.Value(); ok {
		httpAddress = val
	}

	if val, ok := storageAddressFlagVal.Value(); ok {
		storageAddress = val
	}

	if val, ok := jwtPrivateKeyFlagVal.Value(); ok {
		jwtPrivateKey = val
	}

	if val, ok := prometheusPortFlagVal.Value(); ok {
		prometheusPort = val
	}

	if val, ok := otelCollectorAddressFlagVal.Value(); ok {
		otelCollectorAddress = val
	}

	// Environment Variables.

	if valStr, ok := os.LookupEnv(httpAddressEnv); ok {
		httpAddress = valStr
	}

	if valStr, ok := os.LookupEnv(storageAddressEnv); ok {
		storageAddress = valStr
	}

	if valStr, ok := os.LookupEnv(jwtPrivateKeyEnv); ok {
		jwtPrivateKey = valStr
	}

	if valStr, ok := os.LookupEnv(prometheusPortEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return &Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, prometheusPortEnv)
		}
		prometheusPort = val
	}

	if valStr, ok := os.LookupEnv(otelCollectorAddressEnv); ok {
		otelCollectorAddress = valStr
	}

	// Validation.

	if prometheusPort < 0 || prometheusPort > 65535 {
		return &Config{}, errors.New("prometheus port must be between 0 and 65535")
	}

	return &Config{
		JWTConfig: jwtfactory.Config{
			Algorithm:      "HS256",
			Secret:         jwtPrivateKey,
			ExpirationTime: time.Hour,
		},
		StorageConfig: services.StorageConfig{
			ServerAddress: storageAddress,
		},
		HTTPServerConfig: httpserver.Config{
			ServerAddress:   httpAddress,
			ShutdownTimeout: defaultShutdownTimeout,
		},
		PrometheusConfig: meterutils.PrometheusConfig{
			PortToListen:    uint16(prometheusPort),
			ShutdownTimeout: defaultShutdownTimeout,
		},
		OpenTelemetryConfig: meterutils.OpenTelemetryConfig{
			ServiceName:      "api-service",
			CollectorAddress: otelCollectorAddress,
		},
		ShutdownTimeout: defaultShutdownTimeout,
	}, nil
}
