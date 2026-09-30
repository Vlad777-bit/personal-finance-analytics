package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultShutdownTimeout = 10 * time.Second
	defaultGRPCPort        = "9090"
)

var dotEnvFiles = []string{".env", "../.env"}

type Config struct {
	DatabaseURL     string
	GRPCPort        string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	if err := loadDotEnv(dotEnvFiles...); err != nil {
		return Config{}, err
	}

	databaseURL := os.Getenv("LEDGER_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("LEDGER_DATABASE_URL is required")
	}

	grpcPort := envOrDefault("LEDGER_GRPC_PORT", defaultGRPCPort)
	parsedGRPCPort, err := strconv.ParseUint(grpcPort, 10, 16)
	if err != nil || parsedGRPCPort == 0 {
		return Config{}, fmt.Errorf(
			"LEDGER_GRPC_PORT must be a number between 1 and 65535: %q",
			grpcPort,
		)
	}

	shutdownTimeout := defaultShutdownTimeout

	if value := os.Getenv("LEDGER_SHUTDOWN_TIMEOUT"); value != "" {
		parsedTimeout, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"parse LEDGER_SHUTDOWN_TIMEOUT: %w",
				err,
			)
		}
		if parsedTimeout <= 0 {
			return Config{}, errors.New(
				"LEDGER_SHUTDOWN_TIMEOUT must be positive",
			)
		}

		shutdownTimeout = parsedTimeout
	}

	return Config{
		DatabaseURL:     databaseURL,
		GRPCPort:        grpcPort,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func loadDotEnv(filenames ...string) error {
	for _, filename := range filenames {
		_, err := os.Stat(filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return fmt.Errorf("stat dotenv file %s: %w", filename, err)
		}

		if err := godotenv.Load(filename); err != nil {
			return fmt.Errorf("load dotenv file %s: %w", filename, err)
		}

		return nil
	}

	return nil
}
