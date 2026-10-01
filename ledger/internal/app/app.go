package app

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

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
	server   *grpc.Server
	listener net.Listener
}

func New(
	ctx context.Context,
	databaseURL string,
	grpcAddress string,
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

	reportRepository := reportrepository.New(
		databaseClient,
	)

	ledgerService := service.New(
		transactionRepository,
		budgetRepository,
		reportRepository,
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

		return nil, fmt.Errorf("listen for gRPC connections: %w", err)
	}

	return &App{
		database: databaseClient,
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

		return fmt.Errorf("shutdown gRPC server: %w", ctx.Err())
	}

	a.database.Close()

	return nil
}
