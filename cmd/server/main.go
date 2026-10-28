package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/SailmanSeeulater/automorph/internal/config"
	"github.com/SailmanSeeulater/automorph/internal/db"
	"github.com/SailmanSeeulater/automorph/internal/httpapi"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	// Readable logs locally, JSON in production.
	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, nil)
	if cfg.IsDev() {
		handler = slog.NewTextHandler(os.Stderr, nil)
	}
	logger := slog.New(handler)

	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	dbPath := filepath.Join(cfg.DataDir, "automorph.db")
	sqlDB, err := db.Open(context.Background(), dbPath)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	if err := db.Migrate(context.Background(), sqlDB, logger); err != nil {
		return err
	}
	logger.Info("database ready", "path", dbPath)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.NewRouter(logger, sqlDB),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "url", "http://"+cfg.Addr, "env", cfg.Env)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		// Startup failed, e.g. the port is already is use
		return err
	case <-ctx.Done():
	}

	// Give in-flight requests up to 10s to finish.
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
