package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database/pgx"
	passwordbcrypt "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/password/bcrypt"
	userrepository "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/repository/database/user"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
	tokenjwt "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/token/jwt"
	grpctransport "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/transport/grpc"
	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

type App struct {
	database database.DB
	server   *grpc.Server
	listener net.Listener
}

func New(
	ctx context.Context,
	databaseURL string,
	bcryptCost int,
	jwtSecret string,
	jwtIssuer string,
	jwtAccessTTL time.Duration,
	jwtRefreshTTL time.Duration,
	grpcAddress string,
) (*App, error) {
	passwordHasher, err := passwordbcrypt.New(bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("create password hasher: %w", err)
	}

	tokenIssuer, err := tokenjwt.New(jwtSecret, jwtIssuer, jwtAccessTTL, jwtRefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("create token issuer: %w", err)
	}

	databaseClient, err := dbpgx.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create database client: %w", err)
	}

	userRepository := userrepository.New(
		databaseClient,
	)

	authService := service.New(
		userRepository,
		passwordHasher,
		tokenIssuer,
	)
	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(
		grpcServer,
		grpctransport.New(authService),
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
