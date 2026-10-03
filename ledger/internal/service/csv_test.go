package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
	repositorymocks "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/mocks"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
)

var errCSVRepository = errors.New("database error")

func TestService_ImportTransactions(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
	csvData := "amount,category,description,occurred_at\n1500,food,lunch," + date.Format(time.RFC3339) + "\n"

	tests := []struct {
		name    string
		input   service.ImportTransactionsInput
		prepare func(*repositorymocks.TransactionRepository)
		want    int
		failed  int
		wantErr error
	}{
		{
			name:  "imports valid transactions",
			input: service.ImportTransactionsInput{UserID: " user-1 ", CSVData: csvData},
			prepare: func(repo *repositorymocks.TransactionRepository) {
				repo.EXPECT().
					CreateWithinBudget(mock.Anything, mock.MatchedBy(func(transaction domain.Transaction) bool {
						return transaction.UserID == "user-1" &&
							transaction.Amount == 1500 &&
							transaction.Category == "food" &&
							transaction.Description == "lunch" &&
							transaction.OccurredAt.Equal(date)
					})).
					Return(domain.Transaction{}, nil)
			},
			want: 1,
		},
		{
			name:    "requires user id",
			input:   service.ImportTransactionsInput{CSVData: csvData},
			prepare: func(*repositorymocks.TransactionRepository) {},
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name:    "rejects invalid header",
			input:   service.ImportTransactionsInput{UserID: "user-1", CSVData: "amount,category\n1,food\n"},
			prepare: func(*repositorymocks.TransactionRepository) {},
			wantErr: domain.ErrInvalidCSV,
		},
		{
			name:    "rejects invalid amount",
			input:   service.ImportTransactionsInput{UserID: "user-1", CSVData: "amount,category,description,occurred_at\ninvalid,food,lunch," + date.Format(time.RFC3339) + "\n"},
			prepare: func(*repositorymocks.TransactionRepository) {},
			failed:  1,
		},
		{
			name:  "wraps repository error",
			input: service.ImportTransactionsInput{UserID: "user-1", CSVData: csvData},
			prepare: func(repo *repositorymocks.TransactionRepository) {
				repo.EXPECT().CreateWithinBudget(mock.Anything, mock.Anything).Return(domain.Transaction{}, errCSVRepository)
			},
			failed: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := repositorymocks.NewTransactionRepository(t)
			tt.prepare(repo)
			ledgerService := newTestLedgerService(
				repo,
				repositorymocks.NewBudgetRepository(t),
				repositorymocks.NewReportRepository(t),
			)

			got, err := ledgerService.ImportTransactions(t.Context(), tt.input)
			require.Equal(t, tt.want, got.ImportedCount)
			require.Equal(t, tt.failed, got.FailedCount)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}

func TestService_ExportTransactions(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	repo := repositorymocks.NewTransactionRepository(t)
	repo.EXPECT().List(mock.Anything, repository.TransactionFilter{
		UserID: "user-1", Category: "food", From: from, To: to,
	}).Return([]domain.Transaction{
		{
			Amount: 1500, Category: "food", Description: "lunch", OccurredAt: from.Add(12 * time.Hour),
		},
	}, nil)

	ledgerService := newTestLedgerService(
		repo,
		repositorymocks.NewBudgetRepository(t),
		repositorymocks.NewReportRepository(t),
	)

	got, err := ledgerService.ExportTransactions(t.Context(), service.ExportTransactionsInput{
		UserID: "user-1", Category: "food", From: from, To: to,
	})
	require.NoError(t, err)
	require.Equal(t,
		"amount,category,description,occurred_at\n1500,food,lunch,2026-10-01T12:00:00Z\n",
		got,
	)
}

func TestService_ImportTransactions_ReturnsRowErrors(t *testing.T) {
	t.Parallel()

	repo := repositorymocks.NewTransactionRepository(t)
	repo.EXPECT().CreateWithinBudget(mock.Anything, mock.Anything).Return(domain.Transaction{}, nil)
	ledgerService := newTestLedgerService(
		repo,
		repositorymocks.NewBudgetRepository(t),
		repositorymocks.NewReportRepository(t),
	)

	result, err := ledgerService.ImportTransactions(t.Context(), service.ImportTransactionsInput{
		UserID: "user-1",
		CSVData: "amount,category,description,occurred_at\n" +
			"100,food,valid,2026-10-01T00:00:00Z\n" +
			"invalid,food,bad,2026-10-01T00:00:00Z\n" +
			"200,food,valid again,2026-10-01T00:00:00Z\n",
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.ImportedCount)
	require.Equal(t, 1, result.FailedCount)
	require.Equal(t, service.ImportTransactionsError{Row: 3, Message: "parse amount: strconv.ParseInt: parsing \"invalid\": invalid syntax"}, result.Errors[0])
}

func TestService_ImportTransactions_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := repositorymocks.NewTransactionRepository(t)
	ledgerService := service.New(
		repo,
		repositorymocks.NewBudgetRepository(t),
		repositorymocks.NewReportRepository(t),
		&fakeSummaryCache{},
		testSummaryCacheTTL,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	_, err := ledgerService.ImportTransactions(ctx, service.ImportTransactionsInput{
		UserID: "user-1",
		CSVData: "amount,category,description,occurred_at\n" +
			"1,food,lunch,2026-10-01T00:00:00Z\n",
	})
	require.ErrorIs(t, err, context.Canceled)
}
