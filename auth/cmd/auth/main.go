package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/app"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/config"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)

		return 1
	}

	signalContext, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	application, err := app.New(
		signalContext,
		cfg.DatabaseURL,
		cfg.BcryptCost,
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAccessTTL,
	)
	if err != nil {
		logger.Error("initialize auth application", "error", err)

		return 1
	}

	logger.Info("auth service started")
	runErr := application.Run(signalContext)
	if runErr != nil {
		logger.Error("run auth application", "error", runErr)
	}
	logger.Info("auth service stopping")

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancel()

	if err := application.Shutdown(shutdownContext); err != nil {
		logger.Error("stop auth application", "error", err)

		return 1
	}

	logger.Info("auth service stopped")
	if runErr != nil {
		return 1
	}

	return 0
}
