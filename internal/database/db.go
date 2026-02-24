// Package database handles the initialization and connection to databases
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// MinimumPostgresVersion is the minimum required PostgreSQL version
const MinimumPostgresVersion = 12

// InitDB initializes and returns a database connection based on the provided configuration.
// It supports both SQLite (local) and PostgreSQL (remote) databases.
// For SQLite, it creates the necessary directories and applies SQLite-specific settings.
// For PostgreSQL, it connects to the remote database and applies appropriate settings.
// In both cases, it runs migrations to ensure the schema is up to date.
// The name parameter is used for logging to identify which database is being initialized.
func InitDB(ctx context.Context, cfg Config, name string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	slog.Info("initializing database", "type", cfg.Type, "name", name)

	switch cfg.Type {
	case SQLite:
		slog.Debug("initializing SQLite database", "path", cfg.SQLitePath, "name", name)
		db, err = initSQLiteDB(ctx, cfg)
	case PostgreSQL:
		slog.Debug("initializing PostgreSQL database", "host", cfg.PostgresHost, "port", cfg.PostgresPort, "user", cfg.PostgresUser, "dbname", cfg.PostgresDB, "name", name)
		db, err = initPostgresDB(ctx, cfg, name)
	default:
		return nil, fmt.Errorf("failed to initialize database: unknown database type %s", cfg.Type)
	}

	if err != nil {
		slog.Error("failed to initialize database connection", "type", cfg.Type, "name", name, "error", err)
		return nil, err
	}

	slog.Info("database connection established successfully", "type", cfg.Type, "name", name)

	slog.Info("starting database migrations", "type", cfg.Type, "name", name)
	if err := runMigrations(ctx, db, cfg.Type); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close db after migration error", "error", closeErr)
		}
		slog.Error("migration failed", "type", cfg.Type, "name", name, "error", err)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("database migrations completed successfully", "type", cfg.Type, "name", name)
	return db, nil
}

func initSQLiteDB(ctx context.Context, cfg Config) (*sql.DB, error) {
	pasoDir := filepath.Dir(cfg.SQLitePath)
	if err := os.MkdirAll(pasoDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Enable foreign key constraints (required for CASCADE deletions)
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		slog.Error("failed to enable foreign keys", "error", err)
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close db", "error", closeErr)
		}
		return nil, err
	}

	// Enable WAL mode for concurrent readers alongside a single writer.
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		slog.Error("failed to enable WAL mode", "error", err)
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close db", "error", closeErr)
		}
		return nil, err
	}

	// NORMAL synchronous is safe with WAL: commits are durable after a power loss,
	// only the last few WAL frames might be lost (not the whole DB).
	if _, err := db.ExecContext(ctx, "PRAGMA synchronous = NORMAL"); err != nil {
		slog.Error("failed to set synchronous mode", "error", err)
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close db", "error", closeErr)
		}
		return nil, err
	}

	// Set busy timeout to 5 seconds (SQLite will retry for this duration)
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		slog.Error("failed to set busy timeout", "error", err)
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close db", "error", closeErr)
		}
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close db", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// WAL mode allows concurrent readers, so we open multiple connections.
	// Writes still serialize via SQLite's internal WAL lock.
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)

	return db, nil
}

func initPostgresDB(ctx context.Context, cfg Config, name string) (*sql.DB, error) {
	connStr, err := cfg.ConnectionString()
	if err != nil {
		slog.Error("failed to get connection string", "error", err, "name", name)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}
	slog.Debug("attempting PostgreSQL connection", "name", name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("failed to open PostgreSQL connection", "error", err, "name", name)
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	slog.Debug("testing PostgreSQL connection", "name", name)
	if err := db.PingContext(ctx); err != nil {
		slog.Error("PostgreSQL connection test failed", "error", err, "name", name)
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close db", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	slog.Debug("checking PostgreSQL version compatibility", "name", name, "minimum_version", MinimumPostgresVersion)
	if err := checkPostgresVersion(ctx, db, name); err != nil {
		slog.Error("PostgreSQL version check failed", "error", err, "name", name)
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close db", "error", closeErr)
		}
		return nil, err
	}

	slog.Debug("PostgreSQL connection successful, configuring connection pool", "name", name)
	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 25
	}
	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 5
	}

	slog.Debug("setting connection pool parameters", "maxOpenConns", maxOpenConns, "maxIdleConns", maxIdleConns, "name", name)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(0)

	return db, nil
}

func checkPostgresVersion(ctx context.Context, db *sql.DB, name string) error {
	var versionString string
	if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&versionString); err != nil {
		return fmt.Errorf("failed to query PostgreSQL version: %w", err)
	}

	slog.Debug("retrieved PostgreSQL version string", "version", versionString, "name", name)

	majorVersion, err := parsePostgresVersion(versionString)
	if err != nil {
		return fmt.Errorf("failed to parse PostgreSQL version: %w", err)
	}

	slog.Info("PostgreSQL server version detected", "major_version", majorVersion, "minimum_version", MinimumPostgresVersion, "name", name)

	if majorVersion < MinimumPostgresVersion {
		return fmt.Errorf("failed to validate PostgreSQL version: version %d does not meet minimum requirement of %d", majorVersion, MinimumPostgresVersion)
	}

	slog.Debug("PostgreSQL version check passed", "major_version", majorVersion, "name", name)
	return nil
}

// parsePostgresVersion extracts the major version number from a PostgreSQL version string
// like "PostgreSQL 14.5 on x86_64-pc-linux-gnu, compiled by..."
func parsePostgresVersion(versionString string) (int, error) {
	parts := strings.Fields(versionString)
	if len(parts) < 2 {
		return 0, fmt.Errorf("failed to parse version string: unexpected format %s", versionString)
	}

	versionParts := strings.Split(parts[1], ".")
	if len(versionParts) == 0 {
		return 0, fmt.Errorf("failed to extract version components: %s", parts[1])
	}

	majorVersion, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return 0, fmt.Errorf("failed to parse major version number: %w", err)
	}

	return majorVersion, nil
}
