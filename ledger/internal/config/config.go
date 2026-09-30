package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	defaultShutdownTimeout = 10 * time.Second
)

type Config struct {
	DatabaseURL     string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	databaseURL := os.Getenv("LEDGER_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("LEDGER_DATABASE_URL is required")
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

		shutdownTimeout = parsedTimeout
	}

	return Config{
		DatabaseURL:     databaseURL,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}
