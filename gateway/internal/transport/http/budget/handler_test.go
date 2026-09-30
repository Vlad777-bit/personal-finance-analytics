package budget_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	budgettransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/budget"
)

type fakeLedgerClient struct {
	createBudget func(
		context.Context,
		ledgerclient.CreateBudgetInput,
	) (ledgerclient.Budget, error)
}

func (f *fakeLedgerClient) CreateBudget(
	ctx context.Context,
	input ledgerclient.CreateBudgetInput,
) (ledgerclient.Budget, error) {
	return f.createBudget(ctx, input)
}

func TestHandler_Upsert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		category   string
		client     *fakeLedgerClient
		wantStatus int
		wantBody   string
	}{
		{
			name:     "success",
			body:     `{"user_id":"user-1","limit_amount":50000}`,
			category: "food",
			client: &fakeLedgerClient{createBudget: func(
				_ context.Context,
				input ledgerclient.CreateBudgetInput,
			) (ledgerclient.Budget, error) {
				require.Equal(t, ledgerclient.CreateBudgetInput{
					UserID: "user-1", Category: "food", Limit: 50000,
				}, input)

				return ledgerclient.Budget{
					ID: "budget-1", UserID: "user-1", Category: "food", Limit: 50000,
				}, nil
			}},
			wantStatus: http.StatusOK,
			wantBody: `{
				"id":"budget-1",
				"user_id":"user-1",
				"category":"food",
				"limit_amount":50000
			}`,
		},
		{
			name:       "malformed JSON",
			body:       `{`,
			category:   "food",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"request body must contain valid JSON"}}`,
		},
		{
			name:       "unknown field",
			body:       `{"user_id":"user-1","limit_amount":50000,"currency":"USD"}`,
			category:   "food",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_request","message":"request body must contain valid JSON"}}`,
		},
		{
			name:       "invalid input",
			body:       `{"user_id":"","limit_amount":0}`,
			category:   "food",
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody: `{
				"error":{
					"code":"invalid_request",
					"message":"user_id, category and positive limit_amount are required"
				}
			}`,
		},
		{
			name:     "Ledger unavailable",
			body:     `{"user_id":"user-1","limit_amount":50000}`,
			category: "food",
			client: &fakeLedgerClient{createBudget: func(
				context.Context,
				ledgerclient.CreateBudgetInput,
			) (ledgerclient.Budget, error) {
				return ledgerclient.Budget{}, ledgerclient.ErrUnavailable
			}},
			wantStatus: http.StatusServiceUnavailable,
			wantBody: `{
				"error":{"code":"service_unavailable","message":"service unavailable"}
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := budgettransport.NewHandler(test.client)
			request := httptest.NewRequest(
				http.MethodPut,
				"/budgets/"+test.category,
				strings.NewReader(test.body),
			)
			request.SetPathValue("category", test.category)
			recorder := httptest.NewRecorder()

			handler.Upsert(recorder, request)

			require.Equal(t, test.wantStatus, recorder.Code)
			require.JSONEq(t, test.wantBody, recorder.Body.String())
		})
	}
}

func clientMustNotBeCalled(t *testing.T) *fakeLedgerClient {
	t.Helper()

	return &fakeLedgerClient{createBudget: func(
		context.Context,
		ledgerclient.CreateBudgetInput,
	) (ledgerclient.Budget, error) {
		return ledgerclient.Budget{}, errors.New("unexpected CreateBudget call")
	}}
}
