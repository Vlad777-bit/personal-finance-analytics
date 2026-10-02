package budget_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	budgettransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/budget"
)

func TestHandler_GetBudgets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		client     *fakeLedgerClient
		wantStatus int
		wantBody   string
	}{
		{
			name:  "success",
			query: "?user_id=other-user",
			client: &fakeLedgerClient{getBudgets: func(
				_ context.Context,
				input ledgerclient.GetBudgetsInput,
			) ([]ledgerclient.Budget, error) {
				require.Equal(t, ledgerclient.GetBudgetsInput{UserID: "user-1"}, input)

				return []ledgerclient.Budget{
					{ID: "budget-1", UserID: "user-1", Category: "food", Limit: 50000},
				}, nil
			}},
			wantStatus: http.StatusOK,
			wantBody: `[{"id":"budget-1","user_id":"user-1","category":"food",
				"limit_amount":50000}]`,
		},
		{
			name:  "empty result",
			query: "",
			client: &fakeLedgerClient{getBudgets: func(
				context.Context,
				ledgerclient.GetBudgetsInput,
			) ([]ledgerclient.Budget, error) {
				return []ledgerclient.Budget{}, nil
			}},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:  "ledger unavailable",
			query: "",
			client: &fakeLedgerClient{getBudgets: func(
				context.Context,
				ledgerclient.GetBudgetsInput,
			) ([]ledgerclient.Budget, error) {
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

			request := authenticatedRequest(t, http.MethodGet, "/budgets"+tt.query, nil)
			recorder := httptest.NewRecorder()

			budgettransport.NewHandler(tt.client).GetBudgets(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.JSONEq(t, tt.wantBody, recorder.Body.String())
		})
	}
}
