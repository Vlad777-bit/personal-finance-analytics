package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	redisCache "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/cache/redis"
	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database/pgx"
	budgetrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/budget"
	reportrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/report"
	transactionrepository "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository/database/transaction"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/service"
	grpctransport "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/transport/grpc"
	ledgerv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/ledger/v1"
)

type App struct {
	database *dbpgx.Client
	cache    *redisCache.Client
	server   *grpc.Server
	listener net.Listener
}

func New(
	ctx context.Context,
	databaseURL string,
	redisAddress string,
	summaryCacheTTL time.Duration,
	grpcAddress string,
	logger *slog.Logger,
) (*App, error) {
	databaseClient, err := dbpgx.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create database client: %w", err)
	}
	summaryCache, err := redisCache.New(ctx, redisAddress, logger)
	if err != nil {
		databaseClient.Close()

		return nil, fmt.Errorf("create Redis cache: %w", err)
	}

	budgetRepository := budgetrepository.New(
		databaseClient,
	)

	transactionRepository := transactionrepository.New(
		databaseClient,
	)

	reportRepository := reportrepository.New(
		databaseClient,
	)

	ledgerService := service.New(
		transactionRepository,
		budgetRepository,
		reportRepository,
		summaryCache,
		summaryCacheTTL,
		logger,
	)
	grpcServer := grpc.NewServer()
	ledgerv1.RegisterLedgerServiceServer(
		grpcServer,
		grpctransport.New(ledgerService),
	)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		databaseClient.Close()

		return nil, errors.Join(
			fmt.Errorf("listen for gRPC connections: %w", err),
			summaryCache.Close(),
		)
	}

	return &App{
		database: databaseClient,
		cache:    summaryCache,
		server:   grpcServer,
		listener: listener,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	serveError := make(chan error, 1)

	go func() {
		serveError <- a.server.Serve(a.listener)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-serveError:
		return fmt.Errorf("serve gRPC: %w", err)
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		a.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		a.server.Stop()
		<-done
		a.database.Close()
		shutdownErr := fmt.Errorf("shutdown gRPC server: %w", ctx.Err())
		if err := a.cache.Close(); err != nil {
			return errors.Join(
				shutdownErr,
				fmt.Errorf("close summary cache: %w", err),
			)
		}

		return shutdownErr
	}

	a.database.Close()
	if err := a.cache.Close(); err != nil {
		return fmt.Errorf("close summary cache: %w", err)
	}

	return nil
}
