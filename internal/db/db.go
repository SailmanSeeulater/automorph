package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // pure-Go SQLite driver, registers itself as "sqlite"
)

//go:embed migrations/*sql
var migrationsFS embed.FS

// Open opens (or creates) the SQLite database at path with productio-friendly settings.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	params := url.Values{}
	params.Add("_pragma", "foreign_keys(1)")     // enforce REFERENCES constraints (off by default in SQLite)
	params.Add("_pragma", "journal_mode(WAL)")   // readers don't block the writer
	params.Add("_pragma", "synchronous(NORMAL)") // safe with WAL, much faster than FULL
	params.Add("_pragma", "busy_timeout(5000)")  // wait up to 5s for a lock instead of failing
	params.Set("_txlock", "immediate")           // take the write lock at BEGIN, avoids "database is locked" mid-transaction

	db, err := sql.Open("sqlite", path+"?"+params.Encode())
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return db, nil
}

// Migrate applies any migrations in migrations/ that haven't run yet.
func Migrate(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	migrations, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrations)
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}
	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	for _, r := range results {
		logger.Info("migration applied", "version", r.Source.Version, "file", r.Source.Path, "dur", r.Duration)
	}
	return nil
}
