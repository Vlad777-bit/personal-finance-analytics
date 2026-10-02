package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/cache"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
)

var _ LedgerService = (*service)(nil)

type LedgerService interface {
	CreateTransaction(
		ctx context.Context,
		input CreateTransactionInput,
	) (domain.Transaction, error)

	CreateBudget(
		ctx context.Context,
		input CreateBudgetInput,
	) (domain.Budget, error)

	GetTransactions(
		ctx context.Context,
		input GetTransactionsInput,
	) ([]domain.Transaction, error)

	GetBudgets(
		ctx context.Context,
		input GetBudgetsInput,
	) ([]domain.Budget, error)

	GetSummary(
		ctx context.Context,
		input GetSummaryInput,
	) (domain.Summary, error)

	ImportTransactions(
		ctx context.Context,
		input ImportTransactionsInput,
	) (int, error)

	ExportTransactions(
		ctx context.Context,
		input ExportTransactionsInput,
	) (string, error)
}

type CreateTransactionInput struct {
	UserID      string
	Amount      int64
	Category    string
	Description string
	OccurredAt  time.Time
}

type CreateBudgetInput struct {
	UserID   string
	Category string
	Limit    int64
}

type GetTransactionsInput struct {
	UserID   string
	Category string
	From     time.Time
	To       time.Time
}

type GetBudgetsInput struct {
	UserID string
}

type GetSummaryInput struct {
	UserID string
	From   time.Time
	To     time.Time
}

type ImportTransactionsInput struct {
	UserID  string
	CSVData string
}

type ExportTransactionsInput struct {
	UserID   string
	Category string
	From     time.Time
	To       time.Time
}

type service struct {
	transactionRepository repository.TransactionRepository
	budgetRepository      repository.BudgetRepository
	reportRepository      repository.ReportRepository
	summaryCache          cache.SummaryCache
	summaryCacheTTL       time.Duration
	logger                *slog.Logger
}

func New(
	transactionRepository repository.TransactionRepository,
	budgetRepository repository.BudgetRepository,
	reportRepository repository.ReportRepository,
	summaryCache cache.SummaryCache,
	summaryCacheTTL time.Duration,
	logger *slog.Logger,
) LedgerService {
	return &service{
		transactionRepository: transactionRepository,
		budgetRepository:      budgetRepository,
		reportRepository:      reportRepository,
		summaryCache:          summaryCache,
		summaryCacheTTL:       summaryCacheTTL,
		logger:                logger,
	}
}
