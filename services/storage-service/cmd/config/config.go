package config

import (
	"errors"
	"flag"
	"fmt"
	"go-invoice-service/common/pkg/flagtypes"
	"os"
	"storage-service/internal/data/postgres"
	"storage-service/internal/grpc"
	"strconv"
	"time"
)

const (
	postgresConnectionStringFlag = "postgres-connection-string"
	postgresConnectionStringEnv  = "POSTGRES_CONNECTION_STRING"
	grpcPortFlag                 = "grpc-port"
	grpcPortEnv                  = "GRPC_PORT"
)

const (
	defaultGRPCPort = 9090
)

var defaultRetryAttempts = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	PostgresConfig postgres.Config
	GRPCConfig     grpc.Config
}

func Load() (*Config, error) {

	postgresConnectionString := ""
	grpcPort := defaultGRPCPort

	// Flags Definition.

	postgresConnectionStringFlagVal := flagtypes.NewString()
	flag.Var(postgresConnectionStringFlagVal, postgresConnectionStringFlag, "Postgres connection string")

	grpcPortFlagVal := flagtypes.NewInt()
	flag.Var(grpcPortFlagVal, grpcPortFlag, "gRPC port")

	flag.Parse()

	// Flags Parse.

	if val, ok := postgresConnectionStringFlagVal.Value(); ok {
		postgresConnectionString = val
	}

	if val, ok := grpcPortFlagVal.Value(); ok {
		grpcPort = val
	}

	// Environment Variables.

	if valStr, ok := os.LookupEnv(postgresConnectionStringEnv); ok {
		postgresConnectionString = valStr
	}

	if valStr, ok := os.LookupEnv(grpcPortEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return &Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, grpcPortEnv)
		}
		grpcPort = val
	}

	// Validation.

	if postgresConnectionString == "" {
		return &Config{}, errors.New("postgres connection string required")
	}

	if grpcPort < 0 || grpcPort > 65535 {
		return &Config{}, errors.New("grpc port must be between 0 and 65535")
	}

	return &Config{
		PostgresConfig: postgres.Config{
			ConnectionString: postgresConnectionString,
		},
		GRPCConfig: grpc.Config{
			Port: uint16(grpcPort),
		},
	}, nil
}
