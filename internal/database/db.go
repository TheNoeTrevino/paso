// Package database handles the initialization and connection to the SQLite db
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB(ctx context.Context) (*sql.DB, error) {
	// adding paso database to the home dir
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	pasoDir := filepath.Join(home, ".paso")
	if err := os.MkdirAll(pasoDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	dbPath := filepath.Join(pasoDir, "tasks.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign key constraints (required for CASCADE deletions)
	_, err = db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	if err != nil {
		slog.Error("Failed to enable foreign keys", "error", err)
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("error closing db", "error", closeErr)
		}
		return nil, err
	}

	// Enable WAL mode for better concurrency
	_, err = db.ExecContext(ctx, "PRAGMA journal_mode = WAL")
	if err != nil {
		slog.Error("Failed to enable WAL mode", "error", err)
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("error closing db", "error", closeErr)
		}
		return nil, err
	}

	// Set busy timeout to 5 seconds (SQLite will retry for this duration)
	_, err = db.ExecContext(ctx, "PRAGMA busy_timeout = 5000")
	if err != nil {
		slog.Error("Failed to set busy timeout", "error", err)
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("error closing db", "error", closeErr)
		}
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("error closing db", "error", closeErr)
		}
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Configure connection pool to reduce contention
	db.SetMaxOpenConns(1) // SQLite benefits from a single writer connection
	db.SetMaxIdleConns(1)

	if err := runMigrations(ctx, db); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("error closing db", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
