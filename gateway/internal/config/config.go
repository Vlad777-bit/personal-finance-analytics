package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultHTTPAddress       = ":8080"
	defaultLedgerGRPCHost    = "localhost"
	defaultLedgerGRPCPort    = "9090"
	defaultLedgerDialTimeout = 5 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
)

var dotEnvFiles = []string{".env", "../.env"}

type Config struct {
	GatewayHTTPAddr       string
	LedgerGRPCHost        string
	LedgerGRPCPort        string
	LedgerGRPCDialTimeout time.Duration
	ShutdownTimeout       time.Duration
}

func (c Config) LedgerGRPCAddress() string {
	return net.JoinHostPort(c.LedgerGRPCHost, c.LedgerGRPCPort)
}

func Load() (Config, error) {
	if err := loadDotEnv(dotEnvFiles...); err != nil {
		return Config{}, err
	}

	ledgerGRPCPort := envOrDefault("LEDGER_GRPC_PORT", defaultLedgerGRPCPort)
	parsedPort, err := strconv.ParseUint(ledgerGRPCPort, 10, 16)
	if err != nil || parsedPort == 0 {
		return Config{}, fmt.Errorf(
			"LEDGER_GRPC_PORT must be a number between 1 and 65535: %q",
			ledgerGRPCPort,
		)
	}

	ledgerDialTimeout := defaultLedgerDialTimeout
	if value := os.Getenv("LEDGER_GRPC_DIAL_TIMEOUT"); value != "" {
		parsedTimeout, parseErr := time.ParseDuration(value)
		if parseErr != nil {
			return Config{}, fmt.Errorf(
				"parse LEDGER_GRPC_DIAL_TIMEOUT: %w",
				parseErr,
			)
		}
		if parsedTimeout <= 0 {
			return Config{}, errors.New(
				"LEDGER_GRPC_DIAL_TIMEOUT must be positive",
			)
		}

		ledgerDialTimeout = parsedTimeout
	}

	shutdownTimeout := defaultShutdownTimeout
	if value := os.Getenv("GATEWAY_SHUTDOWN_TIMEOUT"); value != "" {
		parsedTimeout, parseErr := time.ParseDuration(value)
		if parseErr != nil {
			return Config{}, fmt.Errorf(
				"parse GATEWAY_SHUTDOWN_TIMEOUT: %w",
				parseErr,
			)
		}
		if parsedTimeout <= 0 {
			return Config{}, errors.New(
				"GATEWAY_SHUTDOWN_TIMEOUT must be positive",
			)
		}

		shutdownTimeout = parsedTimeout
	}

	return Config{
		GatewayHTTPAddr:       envOrDefault("GATEWAY_HTTP_ADDR", defaultHTTPAddress),
		LedgerGRPCHost:        envOrDefault("LEDGER_GRPC_HOST", defaultLedgerGRPCHost),
		LedgerGRPCPort:        ledgerGRPCPort,
		LedgerGRPCDialTimeout: ledgerDialTimeout,
		ShutdownTimeout:       shutdownTimeout,
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
