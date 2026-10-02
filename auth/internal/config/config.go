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
	defaultGRPCPort        = "9091"
	defaultShutdownTimeout = 10 * time.Second
	defaultBcryptCost      = 12
	defaultJWTIssuer       = "personal-finance-analytics/auth"
	defaultJWTAccessTTL    = 15 * time.Minute
	minimumBcryptCost      = 4
	maximumBcryptCost      = 31
	minimumJWTSecretBytes  = 32
)

var dotEnvFiles = []string{".env", "../.env"}

type Config struct {
	DatabaseURL     string
	GRPCPort        string
	ShutdownTimeout time.Duration
	BcryptCost      int
	JWTSecret       string
	JWTIssuer       string
	JWTAccessTTL    time.Duration
}

func Load() (Config, error) {
	if err := loadDotEnv(dotEnvFiles...); err != nil {
		return Config{}, err
	}

	databaseURL := os.Getenv("AUTH_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("AUTH_DATABASE_URL is required")
	}

	grpcPort := envOrDefault("AUTH_GRPC_PORT", defaultGRPCPort)
	if !validPort(grpcPort) {
		return Config{}, fmt.Errorf(
			"AUTH_GRPC_PORT must be a number between 1 and 65535: %q",
			grpcPort,
		)
	}

	shutdownTimeout, err := positiveDuration(
		"AUTH_SHUTDOWN_TIMEOUT",
		defaultShutdownTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	bcryptCost, err := bcryptCost()
	if err != nil {
		return Config{}, err
	}

	jwtSecret := os.Getenv("AUTH_JWT_SECRET")
	if len(jwtSecret) < minimumJWTSecretBytes {
		return Config{}, fmt.Errorf(
			"AUTH_JWT_SECRET must be at least %d bytes",
			minimumJWTSecretBytes,
		)
	}

	jwtAccessTTL, err := positiveDuration(
		"AUTH_JWT_ACCESS_TTL",
		defaultJWTAccessTTL,
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DatabaseURL:     databaseURL,
		GRPCPort:        grpcPort,
		ShutdownTimeout: shutdownTimeout,
		BcryptCost:      bcryptCost,
		JWTSecret:       jwtSecret,
		JWTIssuer:       envOrDefault("AUTH_JWT_ISSUER", defaultJWTIssuer),
		JWTAccessTTL:    jwtAccessTTL,
	}, nil
}

func bcryptCost() (int, error) {
	value := os.Getenv("AUTH_BCRYPT_COST")
	if value == "" {
		return defaultBcryptCost, nil
	}

	cost, err := strconv.Atoi(value)
	if err != nil || cost < minimumBcryptCost || cost > maximumBcryptCost {
		return 0, fmt.Errorf(
			"AUTH_BCRYPT_COST must be a number between %d and %d: %q",
			minimumBcryptCost,
			maximumBcryptCost,
			value,
		)
	}

	return cost, nil
}

func positiveDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}

	return duration, nil
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
