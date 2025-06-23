package config

import (
	"errors"
	"flag"
	"fmt"
	"go-invoice-service/common/pkg/flagtypes"
	"go-invoice-service/common/pkg/promutils"
	"os"
	"strconv"
	"time"
	"validation-service/internal/controllers"
	"validation-service/internal/services"
)

const (
	kafkaAddressFlag       = "kafka-address"
	kafkaAddressEnv        = "KAFKA_ADDRESS"
	storageAddressFlag     = "storage-address"
	storageAddressEnv      = "STORAGE_ADDRESS"
	kafkaPollTimeoutMsFlag = "kafka-poll-timeout-ms"
	kafkaPollTimeoutMsEnv  = "KAFKA_POLL_TIMEOUT_MS"
	prometheusPortFlag     = "prometheus-port"
	prometheusPortEnv      = "PROMETHEUS_PORT"
)

const (
	defaultKafkaAddress       = "localhost:9092"
	defaultStorageAddress     = "localhost:5000"
	defaultKafkaPollTimeoutMs = 100
	defaultShutdownTimeout    = 5 * time.Second
	defaultPrometheusPort     = 9090
)

var defaultRetryAttempts = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	KafkaConsumerConfig   services.KafkaConsumerConfig
	StorageConfig         services.StorageConfig
	KafkaDispatcherConfig controllers.KafkaDispatcherConfig
	PrometheusConfig      promutils.PrometheusConfig
}

func Load() (*Config, error) {

	kafkaAddress := defaultKafkaAddress
	storageAddress := defaultStorageAddress
	kafkaPollTimeoutMs := defaultKafkaPollTimeoutMs
	prometheusPort := defaultPrometheusPort

	// Flags Definition.

	kafkaAddressFlagVal := flagtypes.NewString()
	flag.Var(kafkaAddressFlagVal, kafkaAddressFlag, "Kafka bootstrap server address")

	storageAddressFlagVal := flagtypes.NewString()
	flag.Var(storageAddressFlagVal, storageAddressFlag, "Storage server address")

	kafkaPollTimeoutMsFlagVal := flagtypes.NewInt()
	flag.Var(kafkaPollTimeoutMsFlagVal, kafkaPollTimeoutMsFlag, "Kafka poll timeout (ms)")

	prometheusPortFlagVal := flagtypes.NewInt()
	flag.Var(prometheusPortFlagVal, prometheusPortFlag, "Prometheus port")

	flag.Parse()

	// Flags Parse.

	if val, ok := kafkaAddressFlagVal.Value(); ok {
		kafkaAddress = val
	}

	if val, ok := storageAddressFlagVal.Value(); ok {
		storageAddress = val
	}

	if val, ok := kafkaPollTimeoutMsFlagVal.Value(); ok {
		kafkaPollTimeoutMs = val
	}

	if val, ok := prometheusPortFlagVal.Value(); ok {
		prometheusPort = val
	}

	// Environment Variables.

	if valStr, ok := os.LookupEnv(kafkaAddressEnv); ok {
		kafkaAddress = valStr
	}

	if valStr, ok := os.LookupEnv(storageAddressEnv); ok {
		storageAddress = valStr
	}

	if valStr, ok := os.LookupEnv(kafkaPollTimeoutMsEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return &Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, kafkaPollTimeoutMsEnv)
		}
		kafkaPollTimeoutMs = val
	}

	if valStr, ok := os.LookupEnv(prometheusPortEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return &Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, prometheusPortEnv)
		}
		prometheusPort = val
	}

	// Validation.

	if kafkaPollTimeoutMs < 1 {
		return &Config{}, errors.New("kafka poll timeout must be greater than one")
	}

	if prometheusPort < 0 || prometheusPort > 65535 {
		return &Config{}, errors.New("prometheus port must be between 0 and 65535")
	}

	return &Config{
		KafkaConsumerConfig: services.KafkaConsumerConfig{
			ServerAddress: kafkaAddress,
		},
		StorageConfig: services.StorageConfig{
			ServerAddress: storageAddress,
		},
		KafkaDispatcherConfig: controllers.KafkaDispatcherConfig{
			PollTimeoutMs: kafkaPollTimeoutMs,
		},
		PrometheusConfig: promutils.PrometheusConfig{
			PortToListen:    uint16(prometheusPort),
			ShutdownTimeout: defaultShutdownTimeout,
		},
	}, nil
}
