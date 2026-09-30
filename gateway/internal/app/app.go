package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	ledgerclient "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/client/ledger"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

type App struct {
	server       *http.Server
	ledgerClient *ledgerclient.Client
}

func New(
	ctx context.Context,
	httpAddress string,
	ledgerAddress string,
	ledgerDialTimeout time.Duration,
) (*App, error) {
	ledgerClient, err := ledgerclient.New(
		ctx,
		ledgerAddress,
		ledgerDialTimeout,
	)
	if err != nil {
		return nil, fmt.Errorf("create Ledger client: %w", err)
	}

	return &App{
		server: &http.Server{
			Addr:              httpAddress,
			Handler:           httptransport.NewRouter(),
			ReadHeaderTimeout: 5 * time.Second,
		},
		ledgerClient: ledgerClient,
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

	return errors.Join(shutdownErrors...)
}
