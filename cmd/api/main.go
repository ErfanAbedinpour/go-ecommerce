package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"app/internal/config"
	"app/internal/di"
	"app/internal/infrastructure/persistence/postgres"
	apphttp "app/internal/interfaces/http"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	container, err := di.New(cfg)
	if err != nil {
		return fmt.Errorf("initialize container: %w", err)
	}
	defer container.Close()

	if os.Getenv("RUN_MIGRATIONS") == "true" {
		if err := postgres.RunMigrations(cfg.Database, container.Log); err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}
	}

	router := apphttp.NewRouter(container)

	server := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		container.Log.Info("server starting",
			slog.String("addr", cfg.Server.Addr()),
			slog.String("env", cfg.App.Environment),
			slog.String("version", cfg.App.Version),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		container.Log.Info("shutdown signal received", slog.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	container.Log.Info("server stopped gracefully")
	return nil
}
