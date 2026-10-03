package ledger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

func TestClient_ImportTransactions(t *testing.T) {
	t.Parallel()

	client := &Client{service: &fakeLedgerServiceClient{
		importTransactions: func(
			_ context.Context,
			request *ledgerv1.ImportTransactionsRequest,
		) (*ledgerv1.ImportTransactionsResponse, error) {
			require.Equal(t, "user-1", request.GetUserId())
			require.Equal(t, "csv", request.GetCsvData())

			return &ledgerv1.ImportTransactionsResponse{ImportedCount: 3, FailedCount: 1}, nil
		},
	}}

	got, err := client.ImportTransactions(context.Background(), ImportTransactionsInput{
		UserID: "user-1", CSVData: "csv",
	})
	require.NoError(t, err)
	require.Equal(t, 3, got.ImportedCount)
	require.Equal(t, 1, got.FailedCount)
}

func TestClient_ExportTransactions(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	client := &Client{service: &fakeLedgerServiceClient{
		exportTransactions: func(
			_ context.Context,
			request *ledgerv1.ExportTransactionsRequest,
		) (*ledgerv1.ExportTransactionsResponse, error) {
			require.Equal(t, "user-1", request.GetUserId())
			require.Equal(t, "food", request.GetCategory())
			require.True(t, request.GetFrom().AsTime().Equal(from))
			require.True(t, request.GetTo().AsTime().Equal(to))

			return &ledgerv1.ExportTransactionsResponse{CsvData: "csv"}, nil
		},
	}}

	got, err := client.ExportTransactions(context.Background(), ExportTransactionsInput{
		UserID: "user-1", Category: "food", From: from, To: to,
	})
	require.NoError(t, err)
	require.Equal(t, "csv", got)
}
