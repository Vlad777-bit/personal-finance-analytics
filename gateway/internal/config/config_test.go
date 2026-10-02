package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		httpAddress string
		ledgerHost  string
		ledgerPort  string
		dialTimeout string
		authHost    string
		authPort    string
		authTimeout string
		shutdown    string
		want        Config
		wantError   bool
	}{
		{
			name: "defaults",
			want: Config{
				GatewayHTTPAddr:       ":8080",
				LedgerGRPCHost:        "localhost",
				LedgerGRPCPort:        "9090",
				LedgerGRPCDialTimeout: 5 * time.Second,
				AuthGRPCHost:          "localhost",
				AuthGRPCPort:          "9091",
				AuthGRPCDialTimeout:   5 * time.Second,
				ShutdownTimeout:       10 * time.Second,
			},
		},
		{
			name:        "custom values",
			httpAddress: ":8081",
			ledgerHost:  "ledger",
			ledgerPort:  "19090",
			dialTimeout: "2s",
			authHost:    "auth",
			authPort:    "19091",
			authTimeout: "4s",
			shutdown:    "3s",
			want: Config{
				GatewayHTTPAddr:       ":8081",
				LedgerGRPCHost:        "ledger",
				LedgerGRPCPort:        "19090",
				LedgerGRPCDialTimeout: 2 * time.Second,
				AuthGRPCHost:          "auth",
				AuthGRPCPort:          "19091",
				AuthGRPCDialTimeout:   4 * time.Second,
				ShutdownTimeout:       3 * time.Second,
			},
		},
		{name: "invalid port", ledgerPort: "grpc", wantError: true},
		{name: "zero port", ledgerPort: "0", wantError: true},
		{name: "invalid timeout", dialTimeout: "soon", wantError: true},
		{name: "zero timeout", dialTimeout: "0s", wantError: true},
		{name: "invalid auth port", authPort: "grpc", wantError: true},
		{name: "zero auth port", authPort: "0", wantError: true},
		{name: "invalid auth timeout", authTimeout: "soon", wantError: true},
		{name: "zero auth timeout", authTimeout: "0s", wantError: true},
		{name: "invalid shutdown timeout", shutdown: "later", wantError: true},
		{name: "zero shutdown timeout", shutdown: "0s", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("GATEWAY_HTTP_ADDR", test.httpAddress)
			t.Setenv("LEDGER_GRPC_HOST", test.ledgerHost)
			t.Setenv("LEDGER_GRPC_PORT", test.ledgerPort)
			t.Setenv("LEDGER_GRPC_DIAL_TIMEOUT", test.dialTimeout)
			t.Setenv("AUTH_GRPC_HOST", test.authHost)
			t.Setenv("AUTH_GRPC_PORT", test.authPort)
			t.Setenv("AUTH_GRPC_DIAL_TIMEOUT", test.authTimeout)
			t.Setenv("GATEWAY_SHUTDOWN_TIMEOUT", test.shutdown)

			got, err := Load()
			if test.wantError {
				require.Error(t, err)
				require.Empty(t, got)

				return
			}

			require.NoError(t, err)
			require.Equal(t, test.want, got)
			require.Equal(
				t,
				test.want.LedgerGRPCHost+":"+test.want.LedgerGRPCPort,
				got.LedgerGRPCAddress(),
			)
			require.Equal(
				t,
				test.want.AuthGRPCHost+":"+test.want.AuthGRPCPort,
				got.AuthGRPCAddress(),
			)
		})
	}
}
