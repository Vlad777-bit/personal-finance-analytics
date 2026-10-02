package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
	dbpgx "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database/pgx"
	passwordbcrypt "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/password/bcrypt"
	userrepository "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/repository/database/user"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
	tokenjwt "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/token/jwt"
)

type App struct {
	database    database.DB
	authService service.AuthService
}

func New(
	ctx context.Context,
	databaseURL string,
	bcryptCost int,
	jwtSecret string,
	jwtIssuer string,
	jwtAccessTTL time.Duration,
) (*App, error) {
	passwordHasher, err := passwordbcrypt.New(bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("create password hasher: %w", err)
	}

	tokenIssuer, err := tokenjwt.New(jwtSecret, jwtIssuer, jwtAccessTTL)
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

	return &App{
		database:    databaseClient,
		authService: authService,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		a.database.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown auth application: %w", ctx.Err())
	}
}
