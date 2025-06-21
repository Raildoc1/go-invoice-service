package config

import (
	"flag"
	"go-invoice-service/api-service/internal/httpserver"
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
	jwtPrivateKeyFlag  = "jwt-private-key"
	jwtPrivateKeyEnv   = "JWT_PRIVATE_KEY"
)

const (
	defaultHTTPAddress     = "localhost:8080"
	defaultStorageAddress  = "localhost:9090"
	defaultJWTPrivateKey   = "private-key"
	defaultShutdownTimeout = 5 * time.Second
)

var defaultRetryAttempts = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	StorageConfig    services.StorageConfig
	JWTConfig        jwtfactory.Config
	HTTPServerConfig httpserver.Config
	ShutdownTimeout  time.Duration
}

func Load() (*Config, error) {

	httpAddress := defaultHTTPAddress
	storageAddress := defaultStorageAddress
	jwtPrivateKey := defaultJWTPrivateKey

	// Flags Definition.

	httpAddressFlagVal := flagtypes.NewString()
	flag.Var(httpAddressFlagVal, httpAddressFlag, "HTTP server address")

	storageAddressFlagVal := flagtypes.NewString()
	flag.Var(storageAddressFlagVal, storageAddressFlag, "Storage server address")

	jwtPrivateKeyFlagVal := flagtypes.NewString()
	flag.Var(jwtPrivateKeyFlagVal, jwtPrivateKeyFlag, "JWT private key")

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
		ShutdownTimeout: defaultShutdownTimeout,
	}, nil
}
