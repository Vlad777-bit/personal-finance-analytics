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

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/cache"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
	repositorymocks "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/mocks"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
)

const testSummaryCacheTTL = 5 * time.Minute

type fakeSummaryCache struct {
	get func(
		context.Context,
		string,
		time.Time,
		time.Time,
	) (domain.Summary, cache.Version, bool, error)
	set            func(context.Context, domain.Summary, cache.Version, time.Duration) error
	invalidateUser func(context.Context, string) error
}

func (f *fakeSummaryCache) Get(
	ctx context.Context,
	userID string,
	from time.Time,
	to time.Time,
) (domain.Summary, cache.Version, bool, error) {
	if f.get == nil {
		return domain.Summary{}, 0, false, nil
	}

	return f.get(ctx, userID, from, to)
}

func (f *fakeSummaryCache) Set(
	ctx context.Context,
	summary domain.Summary,
	version cache.Version,
	ttl time.Duration,
) error {
	if f.set == nil {
		return nil
	}

	return f.set(ctx, summary, version, ttl)
}

func (f *fakeSummaryCache) InvalidateUser(ctx context.Context, userID string) error {
	if f.invalidateUser == nil {
		return nil
	}

	return f.invalidateUser(ctx, userID)
}

func newTestLedgerService(
	transactionRepository repository.TransactionRepository,
	budgetRepository repository.BudgetRepository,
	reportRepository repository.ReportRepository,
	summaryCaches ...cache.SummaryCache,
) service.LedgerService {
	summaryCache := cache.SummaryCache(&fakeSummaryCache{})
	if len(summaryCaches) > 0 {
		summaryCache = summaryCaches[0]
	}

	return service.New(
		transactionRepository,
		budgetRepository,
		reportRepository,
		summaryCache,
		testSummaryCacheTTL,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func TestService_GetSummaryCache(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	input := service.GetSummaryInput{UserID: "user-1", From: from, To: to}
	want := domain.Summary{
		UserID: "user-1", From: from, To: to,
		Categories: []domain.CategorySummary{},
	}

	tests := []struct {
		name         string
		summaryCache cache.SummaryCache
		prepare      func(*repositorymocks.ReportRepository)
	}{
		{
			name: "cache hit",
			summaryCache: &fakeSummaryCache{get: func(
				_ context.Context,
				userID string,
				gotFrom time.Time,
				gotTo time.Time,
			) (domain.Summary, cache.Version, bool, error) {
				require.Equal(t, "user-1", userID)
				require.Equal(t, from, gotFrom)
				require.Equal(t, to, gotTo)

				return want, 7, true, nil
			}},
			prepare: func(*repositorymocks.ReportRepository) {},
		},
		{
			name: "cache miss stores repository result",
			summaryCache: &fakeSummaryCache{
				get: func(context.Context, string, time.Time, time.Time) (domain.Summary, cache.Version, bool, error) {
					return domain.Summary{}, 7, false, nil
				},
				set: func(_ context.Context, summary domain.Summary, version cache.Version, ttl time.Duration) error {
					require.Equal(t, want, summary)
					require.Equal(t, cache.Version(7), version)
					require.Equal(t, testSummaryCacheTTL, ttl)

					return nil
				},
			},
			prepare: func(reportRepository *repositorymocks.ReportRepository) {
				reportRepository.EXPECT().
					GetSummaryData(mock.Anything, "user-1", from, to).
					Return([]domain.CategoryReportData{}, nil)
			},
		},
		{
			name: "cache errors do not break report",
			summaryCache: &fakeSummaryCache{
				get: func(context.Context, string, time.Time, time.Time) (domain.Summary, cache.Version, bool, error) {
					return domain.Summary{}, 0, false, errors.New("Redis unavailable")
				},
			},
			prepare: func(reportRepository *repositorymocks.ReportRepository) {
				reportRepository.EXPECT().
					GetSummaryData(mock.Anything, "user-1", from, to).
					Return([]domain.CategoryReportData{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reportRepository := repositorymocks.NewReportRepository(t)
			tt.prepare(reportRepository)
			ledgerService := newTestLedgerService(
				repositorymocks.NewTransactionRepository(t),
				repositorymocks.NewBudgetRepository(t),
				reportRepository,
				tt.summaryCache,
			)

			got, err := ledgerService.GetSummary(t.Context(), input)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}

func TestService_InvalidatesSummaryCache(t *testing.T) {
	t.Parallel()

	t.Run("after creating transaction", func(t *testing.T) {
		t.Parallel()

		input := service.CreateTransactionInput{
			UserID: "user-1", Amount: 100, Category: "food", OccurredAt: time.Now(),
		}
		transactionRepository := repositorymocks.NewTransactionRepository(t)
		budgetRepository := repositorymocks.NewBudgetRepository(t)
		budgetRepository.EXPECT().
			GetByCategory(mock.Anything, "user-1", "food").
			Return(domain.Budget{}, domain.ErrBudgetNotFound)
		transactionRepository.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("domain.Transaction")).
			Return(domain.Transaction{UserID: "user-1"}, nil)
		ledgerService := newTestLedgerService(
			transactionRepository,
			budgetRepository,
			repositorymocks.NewReportRepository(t),
			&fakeSummaryCache{invalidateUser: func(_ context.Context, userID string) error {
				require.Equal(t, "user-1", userID)

				return nil
			}},
		)

		_, err := ledgerService.CreateTransaction(t.Context(), input)
		require.NoError(t, err)
	})

	t.Run("after creating budget", func(t *testing.T) {
		t.Parallel()

		input := service.CreateBudgetInput{UserID: "user-1", Category: "food", Limit: 100}
		budgetRepository := repositorymocks.NewBudgetRepository(t)
		budgetRepository.EXPECT().
			Upsert(mock.Anything, mock.AnythingOfType("domain.Budget")).
			Return(domain.Budget{UserID: "user-1"}, nil)
		ledgerService := newTestLedgerService(
			repositorymocks.NewTransactionRepository(t),
			budgetRepository,
			repositorymocks.NewReportRepository(t),
			&fakeSummaryCache{invalidateUser: func(_ context.Context, userID string) error {
				require.Equal(t, "user-1", userID)

				return errors.New("Redis unavailable")
			}},
		)

		_, err := ledgerService.CreateBudget(t.Context(), input)
		require.NoError(t, err)
	})
}
