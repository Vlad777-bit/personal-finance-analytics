package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	authclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/auth"
	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
	authtransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/auth"
	budgettransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/budget"
	reporttransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/report"
	transactiontransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/transaction"
)

type App struct {
	server       *http.Server
	ledgerClient *ledgerclient.Client
	authClient   *authclient.Client
}

func New(
	ctx context.Context,
	httpAddress string,
	ledgerAddress string,
	ledgerDialTimeout time.Duration,
	authAddress string,
	authDialTimeout time.Duration,
) (*App, error) {
	ledgerClient, err := ledgerclient.New(
		ctx,
		ledgerAddress,
		ledgerDialTimeout,
	)
	if err != nil {
		return nil, fmt.Errorf("create Ledger client: %w", err)
	}
	authClient, err := authclient.New(ctx, authAddress, authDialTimeout)
	if err != nil {
		if closeErr := ledgerClient.Close(); closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("create Auth client: %w", err),
				fmt.Errorf("close Ledger client: %w", closeErr),
			)
		}

		return nil, fmt.Errorf("create Auth client: %w", err)
	}

	authHandler := authtransport.NewHandler(authClient)
	budgetHandler := budgettransport.NewHandler(ledgerClient)
	transactionHandler := transactiontransport.NewHandler(ledgerClient)
	reportHandler := reporttransport.NewHandler(ledgerClient)

	router := httptransport.NewRouter()

	router.HandleFunc("POST /auth/register", authHandler.Register)
	router.HandleFunc("POST /auth/login", authHandler.Login)
	router.HandleFunc("PUT /budgets/{category}", budgetHandler.Upsert)
	router.HandleFunc("GET /budgets", budgetHandler.GetBudgets)
	router.HandleFunc("POST /transactions", transactionHandler.Create)
	router.HandleFunc("GET /transactions", transactionHandler.GetTransactions)
	router.HandleFunc("GET /reports/summary", reportHandler.GetSummary)

	return &App{
		server: &http.Server{
			Addr:              httpAddress,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
		ledgerClient: ledgerClient,
		authClient:   authClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	serveError := make(chan error, 1)

	go func() {
		serveError <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-serveError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("serve HTTP: %w", err)
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	var shutdownErrors []error

	if err := a.server.Shutdown(ctx); err != nil {
		shutdownErrors = append(
			shutdownErrors,
			fmt.Errorf("shutdown HTTP server: %w", err),
		)
	}

	if err := a.ledgerClient.Close(); err != nil {
		shutdownErrors = append(
			shutdownErrors,
			fmt.Errorf("close Ledger client: %w", err),
		)
	}

	if err := a.authClient.Close(); err != nil {
		shutdownErrors = append(
			shutdownErrors,
			fmt.Errorf("close Auth client: %w", err),
		)
	}

	return errors.Join(shutdownErrors...)
}
