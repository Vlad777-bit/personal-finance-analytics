package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/app"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/config"
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

	application, err := app.New(signalContext, cfg.DatabaseURL)
	if err != nil {
		logger.Error("initialize ledger application", "error", err)

		return 1
	}

	logger.Info("ledger service started")
	<-signalContext.Done()
	logger.Info("ledger service stopping")

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancel()

	if err := application.Shutdown(shutdownContext); err != nil {
		logger.Error("stop ledger application", "error", err)

		return 1
	}

	logger.Info("ledger service stopped")

	return 0
}
