package app

import (
	"context"
	"fmt"

	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database/pgx"
	budgetrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/budget"
	transactionrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/transaction"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
)

type App struct {
	database *dbpgx.Client

	service service.LedgerService
}

func New(
	ctx context.Context,
	databaseURL string,
) (*App, error) {
	databaseClient, err := dbpgx.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create database client: %w", err)
	}

	budgetRepository := budgetrepository.New(
		databaseClient,
	)

	transactionRepository := transactionrepository.New(
		databaseClient,
	)

	ledgerService := service.New(
		transactionRepository,
		budgetRepository,
	)

	return &App{
		database: databaseClient,
		service:  ledgerService,
	}, nil
}

func (a *App) Close() {
	a.database.Close()
}

func (a *App) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		a.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown ledger application: %w", ctx.Err())
	}
}
