package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name            string
		databaseURL     string
		grpcPort        string
		shutdownTimeout string
		want            Config
		wantError       bool
	}{
		{
			name:        "default shutdown timeout",
			databaseURL: "postgres://finance:finance@localhost:5432/finance",
			want: Config{
				DatabaseURL:     "postgres://finance:finance@localhost:5432/finance",
				GRPCPort:        "9090",
				ShutdownTimeout: 10 * time.Second,
			},
		},
		{
			name:            "custom shutdown timeout",
			databaseURL:     "postgres://finance:finance@localhost:5432/finance",
			shutdownTimeout: "3s",
			want: Config{
				DatabaseURL:     "postgres://finance:finance@localhost:5432/finance",
				GRPCPort:        "9090",
				ShutdownTimeout: 3 * time.Second,
			},
		},
		{
			name:        "custom gRPC port",
			databaseURL: "postgres://finance:finance@localhost:5432/finance",
			grpcPort:    "19090",
			want: Config{
				DatabaseURL:     "postgres://finance:finance@localhost:5432/finance",
				GRPCPort:        "19090",
				ShutdownTimeout: 10 * time.Second,
			},
		},
		{
			name:        "non-numeric gRPC port",
			databaseURL: "postgres://finance:finance@localhost:5432/finance",
			grpcPort:    "grpc",
			wantError:   true,
		},
		{
			name:        "zero gRPC port",
			databaseURL: "postgres://finance:finance@localhost:5432/finance",
			grpcPort:    "0",
			wantError:   true,
		},
		{
			name:        "out of range gRPC port",
			databaseURL: "postgres://finance:finance@localhost:5432/finance",
			grpcPort:    "65536",
			wantError:   true,
		},
		{
			name:      "missing database URL",
			wantError: true,
		},
		{
			name:            "invalid shutdown timeout",
			databaseURL:     "postgres://finance:finance@localhost:5432/finance",
			shutdownTimeout: "soon",
			wantError:       true,
		},
		{
			name:            "non-positive shutdown timeout",
			databaseURL:     "postgres://finance:finance@localhost:5432/finance",
			shutdownTimeout: "0s",
			wantError:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("LEDGER_DATABASE_URL", test.databaseURL)
			t.Setenv("LEDGER_GRPC_PORT", test.grpcPort)
			t.Setenv("LEDGER_SHUTDOWN_TIMEOUT", test.shutdownTimeout)

			got, err := Load()
			if test.wantError {
				require.Error(t, err)
				require.Empty(t, got)

				return
			}

			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func TestLoadDotEnv(t *testing.T) {
	t.Run("loads existing file", func(t *testing.T) {
		const (
			key   = "LEDGER_DOTENV_TEST_VALUE"
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
		err := os.WriteFile(
			filename,
			[]byte(key+"="+value+"\n"),
			0o600,
		)
		require.NoError(t, err)

		err = loadDotEnv(filename)
		require.NoError(t, err)
		require.Equal(t, value, os.Getenv(key))
	})

	t.Run("keeps existing environment value", func(t *testing.T) {
		const databaseURL = "postgres://existing"

		t.Setenv("LEDGER_DATABASE_URL", databaseURL)

		filename := filepath.Join(t.TempDir(), ".env")
		err := os.WriteFile(
			filename,
			[]byte("LEDGER_DATABASE_URL=postgres://dotenv\n"),
			0o600,
		)
		require.NoError(t, err)

		err = loadDotEnv(filename)
		require.NoError(t, err)
		require.Equal(t, databaseURL, os.Getenv("LEDGER_DATABASE_URL"))
	})

	t.Run("ignores missing files", func(t *testing.T) {
		err := loadDotEnv(filepath.Join(t.TempDir(), ".env"))
		require.NoError(t, err)
	})

	t.Run("rejects malformed file", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), ".env")
		err := os.WriteFile(
			filename,
			[]byte("LEDGER_DATABASE_URL='unterminated\n"),
			0o600,
		)
		require.NoError(t, err)

		err = loadDotEnv(filename)
		require.Error(t, err)
	})
}
