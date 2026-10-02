package transaction_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	transactiontransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/transaction"
)

func TestHandler_GetTransactions(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	createdAt := from.Add(time.Hour)

	tests := []struct {
		name       string
		query      string
		client     *fakeLedgerClient
		wantStatus int
		wantBody   string
	}{
		{
			name:  "success",
			query: "?user_id=other-user&category=food&from=2026-09-01T00:00:00Z&to=2026-10-01T00:00:00Z",
			client: &fakeLedgerClient{getTransactions: func(
				_ context.Context,
				input ledgerclient.GetTransactionsInput,
			) ([]ledgerclient.Transaction, error) {
				require.Equal(t, "user-1", input.UserID)
				require.Equal(t, "food", input.Category)
				require.True(t, input.From.Equal(from))
				require.True(t, input.To.Equal(to))

				return []ledgerclient.Transaction{
					{
						ID: "transaction-1", UserID: "user-1", Amount: 1500,
						Category: "food", Description: "lunch",
						OccurredAt: from, CreatedAt: createdAt,
					},
				}, nil
			}},
			wantStatus: http.StatusOK,
			wantBody: `[{"id":"transaction-1","user_id":"user-1","amount":1500,
				"category":"food","description":"lunch","occurred_at":"2026-09-01T00:00:00Z",
				"created_at":"2026-09-01T01:00:00Z"}]`,
		},
		{
			name:  "empty result",
			query: "?from=2026-09-01T00:00:00Z&to=2026-10-01T00:00:00Z",
			client: &fakeLedgerClient{getTransactions: func(
				context.Context,
				ledgerclient.GetTransactionsInput,
			) ([]ledgerclient.Transaction, error) {
				return []ledgerclient.Transaction{}, nil
			}},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "missing required query",
			query:      "",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"from and to are required"}}`,
		},
		{
			name:       "invalid from",
			query:      "?from=yesterday&to=2026-10-01T00:00:00Z",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"from must use RFC3339 format"}}`,
		},
		{
			name:       "invalid to",
			query:      "?from=2026-09-01T00:00:00Z&to=tomorrow",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"to must use RFC3339 format"}}`,
		},
		{
			name:       "invalid period",
			query:      "?from=2026-10-01T00:00:00Z&to=2026-09-01T00:00:00Z",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"from must be before to"}}`,
		},
		{
			name:  "ledger unavailable",
			query: "?from=2026-09-01T00:00:00Z&to=2026-10-01T00:00:00Z",
			client: &fakeLedgerClient{getTransactions: func(
				context.Context,
				ledgerclient.GetTransactionsInput,
			) ([]ledgerclient.Transaction, error) {
				return nil, ledgerclient.ErrUnavailable
			}},
			wantStatus: http.StatusServiceUnavailable,
			wantBody: `{"error":{"code":"service_unavailable",
				"message":"service unavailable"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := authenticatedRequest(
				t,
				http.MethodGet,
				"/transactions"+tt.query,
				nil,
			)
			recorder := httptest.NewRecorder()

			transactiontransport.NewHandler(tt.client).
				GetTransactions(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.JSONEq(t, tt.wantBody, recorder.Body.String())
		})
	}
}
