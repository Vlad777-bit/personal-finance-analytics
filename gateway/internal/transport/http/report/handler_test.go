package report_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/token"
	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/middleware"
	reporttransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/report"
)

type fakeLedgerClient struct {
	getSummary func(
		context.Context,
		ledgerclient.GetSummaryInput,
	) (ledgerclient.Summary, error)
}

func (f *fakeLedgerClient) GetSummary(
	ctx context.Context,
	input ledgerclient.GetSummaryInput,
) (ledgerclient.Summary, error) {
	return f.getSummary(ctx, input)
}

func TestHandler_GetSummary(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	tests := []struct {
		name       string
		query      string
		client     *fakeLedgerClient
		wantStatus int
		wantBody   string
	}{
		{
			name:  "success",
			query: "?user_id=other-user&from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z",
			client: &fakeLedgerClient{getSummary: func(
				_ context.Context,
				input ledgerclient.GetSummaryInput,
			) (ledgerclient.Summary, error) {
				require.Equal(t, ledgerclient.GetSummaryInput{
					UserID: "user-1", From: from, To: to,
				}, input)

				return ledgerclient.Summary{
					UserID: "user-1", From: from, To: to, TotalSpent: 600,
					Categories: []ledgerclient.CategorySummary{
						{
							Category: "food", Spent: 600, BudgetLimit: 500,
							BudgetConfigured: true, Remaining: -100, BudgetExceeded: true,
						},
					},
				}, nil
			}},
			wantStatus: http.StatusOK,
			wantBody: `{
				"user_id":"user-1","from":"2026-10-01T00:00:00Z",
				"to":"2026-11-01T00:00:00Z","total_spent":600,
				"categories":[{"category":"food","spent":600,"budget_limit":500,
				"budget_configured":true,"remaining":-100,"budget_exceeded":true}]
			}`,
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
			query:      "?from=yesterday&to=2026-11-01T00:00:00Z",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"from must use RFC3339 format"}}`,
		},
		{
			name:       "invalid to",
			query:      "?from=2026-10-01T00:00:00Z&to=tomorrow",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"to must use RFC3339 format"}}`,
		},
		{
			name:       "invalid period",
			query:      "?from=2026-11-01T00:00:00Z&to=2026-10-01T00:00:00Z",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"from must be before to"}}`,
		},
		{
			name:  "ledger unavailable",
			query: "?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z",
			client: &fakeLedgerClient{getSummary: func(
				context.Context,
				ledgerclient.GetSummaryInput,
			) (ledgerclient.Summary, error) {
				return ledgerclient.Summary{}, ledgerclient.ErrUnavailable
			}},
			wantStatus: http.StatusServiceUnavailable,
			wantBody: `{"error":{"code":"service_unavailable",
				"message":"service unavailable"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, "/reports/summary"+tt.query, nil)
			request = request.WithContext(middleware.WithIdentity(
				request.Context(),
				token.Identity{UserID: "user-1", Email: "user@example.com"},
			))
			recorder := httptest.NewRecorder()

			reporttransport.NewHandler(tt.client).GetSummary(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.JSONEq(t, tt.wantBody, recorder.Body.String())
		})
	}
}

func clientMustNotBeCalled(t *testing.T) *fakeLedgerClient {
	t.Helper()

	return &fakeLedgerClient{getSummary: func(
		context.Context,
		ledgerclient.GetSummaryInput,
	) (ledgerclient.Summary, error) {
		return ledgerclient.Summary{}, errors.New("unexpected GetSummary call")
	}}
}
