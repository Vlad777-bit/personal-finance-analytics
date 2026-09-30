package transaction_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	transactiontransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/transaction"
)

type fakeLedgerClient struct {
	createTransaction func(
		context.Context,
		ledgerclient.CreateTransactionInput,
	) (ledgerclient.Transaction, error)
}

func (f *fakeLedgerClient) CreateTransaction(
	ctx context.Context,
	input ledgerclient.CreateTransactionInput,
) (ledgerclient.Transaction, error) {
	return f.createTransaction(ctx, input)
}

func TestHandler_Create(t *testing.T) {
	t.Parallel()

	occurredAt := time.Date(2026, time.September, 30, 12, 0, 0, 0, time.UTC)
	createdAt := occurredAt.Add(time.Second)
	tests := []struct {
		name       string
		body       string
		client     *fakeLedgerClient
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{
				"user_id":"user-1",
				"amount":1500,
				"category":"food",
				"description":"lunch",
				"occurred_at":"2026-09-30T12:00:00Z"
			}`,
			client: &fakeLedgerClient{createTransaction: func(
				_ context.Context,
				input ledgerclient.CreateTransactionInput,
			) (ledgerclient.Transaction, error) {
				require.Equal(t, "user-1", input.UserID)
				require.Equal(t, int64(1500), input.Amount)
				require.True(t, input.OccurredAt.Equal(occurredAt))

				return ledgerclient.Transaction{
					ID: "transaction-1", UserID: "user-1", Amount: 1500,
					Category: "food", Description: "lunch",
					OccurredAt: occurredAt, CreatedAt: createdAt,
				}, nil
			}},
			wantStatus: http.StatusCreated,
			wantBody: `{
				"id":"transaction-1",
				"user_id":"user-1",
				"amount":1500,
				"category":"food",
				"description":"lunch",
				"occurred_at":"2026-09-30T12:00:00Z",
				"created_at":"2026-09-30T12:00:01Z"
			}`,
		},
		{
			name:       "invalid input",
			body:       `{"user_id":"","amount":0,"category":"","occurred_at":""}`,
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody: `{
				"error":{
					"code":"invalid_request",
					"message":"user_id, positive amount, category and occurred_at are required"
				}
			}`,
		},
		{
			name:       "invalid occurred_at",
			body:       `{"user_id":"user-1","amount":100,"category":"food","occurred_at":"today"}`,
			client:     clientMustNotBeCalled(t),
			wantStatus: http.StatusBadRequest,
			wantBody: `{
				"error":{"code":"invalid_request","message":"occurred_at must use RFC3339 format"}
			}`,
		},
		{
			name: "budget exceeded",
			body: `{
				"user_id":"user-1","amount":1500,"category":"food",
				"occurred_at":"2026-09-30T12:00:00Z"
			}`,
			client: &fakeLedgerClient{createTransaction: func(
				context.Context,
				ledgerclient.CreateTransactionInput,
			) (ledgerclient.Transaction, error) {
				return ledgerclient.Transaction{}, ledgerclient.ErrBudgetExceeded
			}},
			wantStatus: http.StatusConflict,
			wantBody:   `{"error":{"code":"budget_exceeded","message":"budget exceeded"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := transactiontransport.NewHandler(test.client)
			request := httptest.NewRequest(
				http.MethodPost,
				"/transactions",
				strings.NewReader(test.body),
			)
			recorder := httptest.NewRecorder()

			handler.Create(recorder, request)

			require.Equal(t, test.wantStatus, recorder.Code)
			require.JSONEq(t, test.wantBody, recorder.Body.String())
		})
	}
}

func clientMustNotBeCalled(t *testing.T) *fakeLedgerClient {
	t.Helper()

	return &fakeLedgerClient{createTransaction: func(
		context.Context,
		ledgerclient.CreateTransactionInput,
	) (ledgerclient.Transaction, error) {
		return ledgerclient.Transaction{}, errors.New("unexpected CreateTransaction call")
	}}
}
