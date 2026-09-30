package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/app"
	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/config"
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
		cfg.GatewayHTTPAddr,
		cfg.LedgerGRPCAddress(),
		cfg.LedgerGRPCDialTimeout,
	)
	if err != nil {
		logger.Error("initialize Gateway application", "error", err)

		return 1
	}

	logger.Info("gateway service started", "http_addr", cfg.GatewayHTTPAddr)
	runErr := application.Run(signalContext)
	if runErr != nil {
		logger.Error("run Gateway application", "error", runErr)
	}
	logger.Info("gateway service stopping")

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancel()

	if err := application.Shutdown(shutdownContext); err != nil {
		logger.Error("stop Gateway application", "error", err)

		return 1
	}

	logger.Info("gateway service stopped")
	if runErr != nil {
		return 1
	}

	return 0
}
