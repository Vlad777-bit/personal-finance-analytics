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
	defaultShutdownTimeout = 10 * time.Second
	defaultGRPCPort        = "9090"
	defaultRedisHost       = "localhost"
	defaultRedisPort       = "6379"
	defaultSummaryCacheTTL = 5 * time.Minute
)

var dotEnvFiles = []string{".env", "../.env"}

type Config struct {
	DatabaseURL     string
	GRPCPort        string
	ShutdownTimeout time.Duration
	RedisAddress    string
	SummaryCacheTTL time.Duration
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
	if !validPort(grpcPort) {
		return Config{}, fmt.Errorf(
			"LEDGER_GRPC_PORT must be a number between 1 and 65535: %q",
			grpcPort,
		)
	}

	redisHost := envOrDefault("REDIS_HOST", defaultRedisHost)
	redisPort := envOrDefault("REDIS_PORT", defaultRedisPort)
	if !validPort(redisPort) {
		return Config{}, fmt.Errorf(
			"REDIS_PORT must be a number between 1 and 65535: %q",
			redisPort,
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

	summaryCacheTTL := defaultSummaryCacheTTL
	if value := os.Getenv("LEDGER_SUMMARY_CACHE_TTL"); value != "" {
		parsedTTL, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse LEDGER_SUMMARY_CACHE_TTL: %w", err)
		}
		if parsedTTL <= 0 {
			return Config{}, errors.New("LEDGER_SUMMARY_CACHE_TTL must be positive")
		}

		summaryCacheTTL = parsedTTL
	}

	return Config{
		DatabaseURL:     databaseURL,
		GRPCPort:        grpcPort,
		ShutdownTimeout: shutdownTimeout,
		RedisAddress:    net.JoinHostPort(redisHost, redisPort),
		SummaryCacheTTL: summaryCacheTTL,
	}, nil
}

func validPort(value string) bool {
	port, err := strconv.ParseUint(value, 10, 16)

	return err == nil && port > 0
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
