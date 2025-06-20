package config

import (
	"flag"
	"go-invoice-service/api-service/internal/http"
	"go-invoice-service/api-service/internal/services"
	"go-invoice-service/common/pkg/flagtypes"
	"go-invoice-service/common/pkg/jwtfactory"
	"os"
	"time"
)

const (
	httpAddressFlag    = "http-address"
	httpAddressEnv     = "HTTP_ADDRESS"
	storageAddressFlag = "storage-address"
	storageAddressEnv  = "STORAGE_ADDRESS"
)

const (
	defaultHTTPAddress    = "localhost:8080"
	defaultStorageAddress = "localhost:5000"
)

var defaultRetryAttempts = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	StorageConfig    services.StorageConfig
	JWTConfig        jwtfactory.Config
	HTTPServerConfig http.Config
}

func Load() (*Config, error) {

	httpAddress := defaultHTTPAddress
	storageAddress := defaultStorageAddress

	// Flags Definition.

	httpAddressFlagVal := flagtypes.NewString()
	flag.Var(httpAddressFlagVal, httpAddressFlag, "HTTP server address")

	storageAddressFlagVal := flagtypes.NewString()
	flag.Var(storageAddressFlagVal, storageAddressFlag, "Storage server address")

	flag.Parse()

	// Flags Parse.

	if val, ok := httpAddressFlagVal.Value(); ok {
		httpAddress = val
	}

	if val, ok := storageAddressFlagVal.Value(); ok {
		storageAddress = val
	}

	// Environment Variables.

	if valStr, ok := os.LookupEnv(httpAddressEnv); ok {
		httpAddress = valStr
	}

	if valStr, ok := os.LookupEnv(storageAddressEnv); ok {
		storageAddress = valStr
	}

	return &Config{
		JWTConfig: jwtfactory.Config{
			Algorithm:      "HS256",
			Secret:         "secret",
			ExpirationTime: time.Hour,
		},
		StorageConfig: services.StorageConfig{
			ServerAddress: storageAddress,
		},
		HTTPServerConfig: http.Config{
			ServerAddress: httpAddress,
		},
	}, nil
}
