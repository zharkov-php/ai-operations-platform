package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"sort"

	"github.com/jackc/pgx/v5"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	if err := run(context.Background(), os.Getenv("DATABASE_URL")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, databaseURL string) error {
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	if _, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return err
	}
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		var applied bool
		if err = conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version=$1)`, entry.Name()).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}
		body, readErr := migrations.ReadFile("migrations/" + entry.Name())
		if readErr != nil {
			return readErr
		}
		tx, beginErr := conn.Begin(ctx)
		if beginErr != nil {
			return beginErr
		}
		if _, execErr := tx.Exec(ctx, string(body)); execErr != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migration %s: %w", entry.Name(), execErr)
		}
		if _, execErr := tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, entry.Name()); execErr != nil {
			_ = tx.Rollback(ctx)
			return execErr
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return commitErr
		}
	}
	return nil
}
