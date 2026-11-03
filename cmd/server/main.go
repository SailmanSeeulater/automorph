package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/SailmanSeeulater/automorph/internal/auth"
	"github.com/SailmanSeeulater/automorph/internal/config"
	"github.com/SailmanSeeulater/automorph/internal/db"
	"github.com/SailmanSeeulater/automorph/internal/httpapi"
	"github.com/SailmanSeeulater/automorph/internal/store"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run opens the database, then either runs a subcommand or starts the server.
func run(args []string) error {
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

	if len(args) > 0 {
		switch args[0] {
		case "create-admin":
			return createAdmin(context.Background(), store.New(sqlDB), args[1:])
		default:
			return fmt.Errorf("unknown command %q (available: create-admin)", args[0])
		}
	}
	return serve(cfg, logger, sqlDB)
}

func serve(cfg config.Config, logger *slog.Logger, sqlDB *sql.DB) error {
	queries := store.New(sqlDB)

	srv := &http.Server{
		Addr: cfg.Addr,
		Handler: httpapi.NewRouter(httpapi.Deps{
			Logger:        logger,
			DB:            sqlDB,
			Queries:       queries,
			Auth:          auth.NewService(queries),
			SecureCookies: !cfg.IsDev(),
		}),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Ctrl+C or SIGTERM (systemd stop) cancels ctx.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "url", "http://"+cfg.Addr, "env", cfg.Env)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		// Startup failed, e.g. the port is already in use.
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
