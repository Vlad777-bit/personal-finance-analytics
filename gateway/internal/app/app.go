package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	authclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/auth"
	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	tokenjwt "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/token/jwt"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
	authtransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/auth"
	budgettransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/budget"
	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http/middleware"
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
	jwtSecret string,
	jwtIssuer string,
) (*App, error) {
	tokenVerifier, err := tokenjwt.New(jwtSecret, jwtIssuer)
	if err != nil {
		return nil, fmt.Errorf("create JWT verifier: %w", err)
	}

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
	csvHandler := transactiontransport.NewCSVHandler(ledgerClient)
	reportHandler := reporttransport.NewHandler(ledgerClient)

	router := httptransport.NewRouter()
	authenticate := middleware.Authenticate(tokenVerifier)
	router.Handle("POST /api/auth/register", http.HandlerFunc(authHandler.Register))
	router.Handle("POST /api/auth/login", http.HandlerFunc(authHandler.Login))
	router.Handle("POST /api/auth/refresh", http.HandlerFunc(authHandler.Refresh))
	router.Handle("POST /api/auth/logout", http.HandlerFunc(authHandler.Logout))
	router.Handle("POST /api/auth/logout-all", authenticate(http.HandlerFunc(authHandler.LogoutAll)))
	router.Handle("PUT /api/budgets/{category}", authenticate(http.HandlerFunc(budgetHandler.Upsert)))
	router.Handle("GET /api/budgets", authenticate(http.HandlerFunc(budgetHandler.GetBudgets)))
	router.Handle("POST /api/transactions", authenticate(http.HandlerFunc(transactionHandler.Create)))
	router.Handle("GET /api/transactions", authenticate(http.HandlerFunc(transactionHandler.GetTransactions)))
	router.Handle("POST /api/transactions/import", authenticate(http.HandlerFunc(csvHandler.Import)))
	router.Handle("GET /api/transactions/export", authenticate(http.HandlerFunc(csvHandler.Export)))
	router.Handle("GET /api/reports/summary", authenticate(http.HandlerFunc(reportHandler.GetSummary)))

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
