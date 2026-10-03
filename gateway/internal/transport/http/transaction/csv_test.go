package transaction_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	transactiontransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/transaction"
)

type fakeCSVClient struct {
	importTransactions func(context.Context, ledgerclient.ImportTransactionsInput) (ledgerclient.ImportTransactionsResult, error)
	exportTransactions func(context.Context, ledgerclient.ExportTransactionsInput) (string, error)
}

func (f *fakeCSVClient) ImportTransactions(
	ctx context.Context,
	input ledgerclient.ImportTransactionsInput,
) (ledgerclient.ImportTransactionsResult, error) {
	return f.importTransactions(ctx, input)
}

func (f *fakeCSVClient) ExportTransactions(
	ctx context.Context,
	input ledgerclient.ExportTransactionsInput,
) (string, error) {
	return f.exportTransactions(ctx, input)
}

func TestHandler_ImportCSV(t *testing.T) {
	t.Parallel()

	client := &fakeCSVClient{
		importTransactions: func(
			_ context.Context,
			input ledgerclient.ImportTransactionsInput,
		) (ledgerclient.ImportTransactionsResult, error) {
			require.Equal(t, "user-1", input.UserID)
			require.Equal(t, "amount,category,description,occurred_at\n", input.CSVData)

			return ledgerclient.ImportTransactionsResult{ImportedCount: 2}, nil
		},
	}
	handler := transactiontransport.NewCSVHandler(client)
	request := authenticatedRequest(t, http.MethodPost, "/transactions/import", strings.NewReader("amount,category,description,occurred_at\n"))
	recorder := httptest.NewRecorder()

	handler.Import(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"imported_count":2,"failed_count":0,"errors":[]}`, recorder.Body.String())
}

func TestHandler_ExportCSV(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	client := &fakeCSVClient{
		exportTransactions: func(
			_ context.Context,
			input ledgerclient.ExportTransactionsInput,
		) (string, error) {
			require.Equal(t, "user-1", input.UserID)
			require.Equal(t, "food", input.Category)
			require.True(t, input.From.Equal(from))
			require.True(t, input.To.Equal(to))

			return "csv\n", nil
		},
	}
	handler := transactiontransport.NewCSVHandler(client)
	request := authenticatedRequest(
		t,
		http.MethodGet,
		"/transactions/export?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z&category=food",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.Export(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "text/csv; charset=utf-8", recorder.Header().Get("Content-Type"))
	require.Equal(t, `attachment; filename="transactions.csv"`, recorder.Header().Get("Content-Disposition"))
	require.Equal(t, "csv\n", recorder.Body.String())
}

func TestHandler_ExportCSV_InvalidPeriod(t *testing.T) {
	t.Parallel()

	client := &fakeCSVClient{exportTransactions: func(
		context.Context,
		ledgerclient.ExportTransactionsInput,
	) (string, error) {
		t.Fatal("unexpected ExportTransactions call")

		return "", nil
	}}
	handler := transactiontransport.NewCSVHandler(client)
	request := authenticatedRequest(t, http.MethodGet, "/transactions/export", nil)
	recorder := httptest.NewRecorder()

	handler.Export(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.JSONEq(t, `{"error":{"code":"invalid_request","message":"from and to must use RFC3339 format and from must be before to"}}`, recorder.Body.String())
}
