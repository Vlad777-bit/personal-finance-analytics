package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testJWTSecret = "0123456789abcdef0123456789abcdef"

func TestLoad(t *testing.T) {
	tests := []struct {
		name            string
		databaseURL     string
		grpcPort        string
		shutdownTimeout string
		bcryptCost      string
		jwtSecret       string
		jwtIssuer       string
		jwtAccessTTL    string
		want            Config
		wantError       bool
	}{
		{
			name:        "defaults",
			databaseURL: "postgres://finance:finance@localhost:5432/finance",
			jwtSecret:   testJWTSecret,
			want: Config{
				DatabaseURL:     "postgres://finance:finance@localhost:5432/finance",
				GRPCPort:        "9091",
				ShutdownTimeout: 10 * time.Second,
				BcryptCost:      12,
				JWTSecret:       testJWTSecret,
				JWTIssuer:       "personal-finance-analytics/auth",
				JWTAccessTTL:    15 * time.Minute,
				JWTRefreshTTL:   30 * 24 * time.Hour,
			},
		},
		{
			name:            "custom values",
			databaseURL:     "postgres://finance:finance@postgres:5432/finance",
			grpcPort:        "19091",
			shutdownTimeout: "3s",
			bcryptCost:      "10",
			jwtSecret:       testJWTSecret + "-custom",
			jwtIssuer:       "custom-auth",
			jwtAccessTTL:    "30m",
			want: Config{
				DatabaseURL:     "postgres://finance:finance@postgres:5432/finance",
				GRPCPort:        "19091",
				ShutdownTimeout: 3 * time.Second,
				BcryptCost:      10,
				JWTSecret:       testJWTSecret + "-custom",
				JWTIssuer:       "custom-auth",
				JWTAccessTTL:    30 * time.Minute,
				JWTRefreshTTL:   30 * 24 * time.Hour,
			},
		},
		{name: "missing database URL", jwtSecret: testJWTSecret, wantError: true},
		{name: "missing JWT secret", databaseURL: "postgres://database", wantError: true},
		{name: "short JWT secret", databaseURL: "postgres://database", jwtSecret: "short", wantError: true},
		{name: "invalid gRPC port", databaseURL: "postgres://database", grpcPort: "grpc", jwtSecret: testJWTSecret, wantError: true},
		{name: "zero gRPC port", databaseURL: "postgres://database", grpcPort: "0", jwtSecret: testJWTSecret, wantError: true},
		{name: "out of range gRPC port", databaseURL: "postgres://database", grpcPort: "65536", jwtSecret: testJWTSecret, wantError: true},
		{name: "invalid shutdown timeout", databaseURL: "postgres://database", shutdownTimeout: "soon", jwtSecret: testJWTSecret, wantError: true},
		{name: "zero shutdown timeout", databaseURL: "postgres://database", shutdownTimeout: "0s", jwtSecret: testJWTSecret, wantError: true},
		{name: "invalid bcrypt cost", databaseURL: "postgres://database", bcryptCost: "expensive", jwtSecret: testJWTSecret, wantError: true},
		{name: "low bcrypt cost", databaseURL: "postgres://database", bcryptCost: "3", jwtSecret: testJWTSecret, wantError: true},
		{name: "high bcrypt cost", databaseURL: "postgres://database", bcryptCost: "32", jwtSecret: testJWTSecret, wantError: true},
		{name: "invalid JWT TTL", databaseURL: "postgres://database", jwtSecret: testJWTSecret, jwtAccessTTL: "later", wantError: true},
		{name: "zero JWT TTL", databaseURL: "postgres://database", jwtSecret: testJWTSecret, jwtAccessTTL: "0s", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AUTH_DATABASE_URL", tt.databaseURL)
			t.Setenv("AUTH_GRPC_PORT", tt.grpcPort)
			t.Setenv("AUTH_SHUTDOWN_TIMEOUT", tt.shutdownTimeout)
			t.Setenv("AUTH_BCRYPT_COST", tt.bcryptCost)
			t.Setenv("AUTH_JWT_SECRET", tt.jwtSecret)
			t.Setenv("AUTH_JWT_ISSUER", tt.jwtIssuer)
			t.Setenv("AUTH_JWT_ACCESS_TTL", tt.jwtAccessTTL)

			got, err := Load()
			if tt.wantError {
				require.Error(t, err)
				require.Empty(t, got)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLoadDotEnv(t *testing.T) {
	t.Run("loads existing file", func(t *testing.T) {
		const (
			key   = "AUTH_DOTENV_TEST_VALUE"
			value = "from-dotenv"
		)

		originalValue, existed := os.LookupEnv(key)
		require.NoError(t, os.Unsetenv(key))
		t.Cleanup(func() {
			if existed {
				require.NoError(t, os.Setenv(key, originalValue))

				return
			}

			require.NoError(t, os.Unsetenv(key))
		})

		filename := filepath.Join(t.TempDir(), ".env")
		err := os.WriteFile(filename, []byte(key+"="+value+"\n"), 0o600)
		require.NoError(t, err)

		err = loadDotEnv(filename)
		require.NoError(t, err)
		require.Equal(t, value, os.Getenv(key))
	})

	t.Run("keeps existing environment value", func(t *testing.T) {
		const databaseURL = "postgres://existing"

		t.Setenv("AUTH_DATABASE_URL", databaseURL)
		filename := filepath.Join(t.TempDir(), ".env")
		err := os.WriteFile(
			filename,
			[]byte("AUTH_DATABASE_URL=postgres://dotenv\n"),
			0o600,
		)
		require.NoError(t, err)

		err = loadDotEnv(filename)
		require.NoError(t, err)
		require.Equal(t, databaseURL, os.Getenv("AUTH_DATABASE_URL"))
	})

	t.Run("ignores missing files", func(t *testing.T) {
		err := loadDotEnv(filepath.Join(t.TempDir(), ".env"))
		require.NoError(t, err)
	})

	t.Run("rejects malformed file", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), ".env")
		err := os.WriteFile(
			filename,
			[]byte("AUTH_DATABASE_URL='unterminated\n"),
			0o600,
		)
		require.NoError(t, err)

		err = loadDotEnv(filename)
		require.Error(t, err)
	})
}
