package config

import (
	"errors"
	"flag"
	"fmt"
	"go-invoice-service/common/pkg/flagtypes"
	"message-sheduler-service/internal/controllers"
	"message-sheduler-service/internal/services"
	"os"
	"strconv"
	"time"
)

const (
	kafkaAddressFlag     = "kafka-address"
	kafkaAddressEnv      = "KAFKA_ADDRESS"
	storageAddressFlag   = "storage-address"
	storageAddressEnv    = "STORAGE_ADDRESS"
	workersCountFlag     = "workers-count"
	workersCountEnv      = "WORKERS_COUNT"
	retryIntervalFlag    = "retry-interval"
	retryIntervalEnv     = "RETRY_INTERVAL_MS"
	dispatchIntervalFlag = "dispatch-interval"
	dispatchIntervalEnv  = "DISPATCH_INTERVAL_MS"
)

const (
	defaultKafkaAddress     = "localhost:9092"
	defaultStorageAddress   = "localhost:9090"
	defaultWorkersCount     = 3
	defaultRetryInterval    = 30 * time.Second
	defaultDispatchInterval = 1 * time.Second
)

var defaultRetryAttempts = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	KafkaProducerConfig    services.KafkaProducerConfig
	StorageConfig          services.StorageConfig
	OutboxDispatcherConfig controllers.OutboxDispatcherConfig
}

func Load() (*Config, error) {

	kafkaAddress := defaultKafkaAddress
	storageAddress := defaultStorageAddress
	workersCount := defaultWorkersCount
	retryInterval := defaultRetryInterval
	dispatchInterval := defaultDispatchInterval

	// Flags Definition.

	kafkaAddressFlagVal := flagtypes.NewString()
	flag.Var(kafkaAddressFlagVal, kafkaAddressFlag, "Kafka bootstrap server address")

	storageAddressFlagVal := flagtypes.NewString()
	flag.Var(storageAddressFlagVal, storageAddressFlag, "Storage server address")

	workersCountFlagVal := flagtypes.NewInt()
	flag.Var(workersCountFlagVal, workersCountFlag, "Workers count")

	retryIntervalFlagVal := flagtypes.NewInt()
	flag.Var(retryIntervalFlagVal, retryIntervalFlag, "Retry interval (ms)")

	dispatchIntervalFlagVal := flagtypes.NewInt()
	flag.Var(dispatchIntervalFlagVal, dispatchIntervalFlag, "Dispatch interval (ms)")

	flag.Parse()

	// Flags Parse.

	if val, ok := kafkaAddressFlagVal.Value(); ok {
		kafkaAddress = val
	}

	if val, ok := storageAddressFlagVal.Value(); ok {
		storageAddress = val
	}

	if val, ok := workersCountFlagVal.Value(); ok {
		workersCount = val
	}

	if val, ok := retryIntervalFlagVal.Value(); ok {
		retryInterval = time.Duration(val) * time.Millisecond
	}

	if val, ok := dispatchIntervalFlagVal.Value(); ok {
		dispatchInterval = time.Duration(val) * time.Millisecond
	}

	// Environment Variables.

	if valStr, ok := os.LookupEnv(kafkaAddressEnv); ok {
		kafkaAddress = valStr
	}

	if valStr, ok := os.LookupEnv(storageAddressEnv); ok {
		storageAddress = valStr
	}

	if valStr, ok := os.LookupEnv(workersCountEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return &Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, workersCountEnv)
		}
		workersCount = val
	}

	if valStr, ok := os.LookupEnv(retryIntervalEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return &Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, retryIntervalEnv)
		}
		retryInterval = time.Duration(val) * time.Millisecond
	}

	if valStr, ok := os.LookupEnv(dispatchIntervalEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return &Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, dispatchIntervalEnv)
		}
		dispatchInterval = time.Duration(val) * time.Millisecond
	}

	// Validation.

	if workersCount < 1 {
		return &Config{}, errors.New("workers count must be greater than one")
	}

	if retryInterval < time.Duration(0) {
		return &Config{}, errors.New("retry internal must be greater than zero")
	}

	if dispatchInterval < time.Duration(0) {
		return &Config{}, errors.New("dispatch internal must be greater than zero")
	}

	return &Config{
		KafkaProducerConfig: services.KafkaProducerConfig{
			ServerAddress: kafkaAddress,
		},
		StorageConfig: services.StorageConfig{
			ServerAddress: storageAddress,
		},
		OutboxDispatcherConfig: controllers.OutboxDispatcherConfig{
			DispatchInterval: dispatchInterval,
			RetryIn:          retryInterval,
			NumWorkers:       int32(workersCount),
		},
	}, nil
}
