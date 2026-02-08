package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/thenoetrevino/paso/internal/database"
)

// PostgresTestConfig holds PostgreSQL test configuration
type PostgresTestConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// DefaultPostgresTestConfig returns default PostgreSQL test configuration from environment or defaults
// Environment variables (if set): PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DATABASE
// Defaults: localhost:5432, postgres/postgres, postgres
func DefaultPostgresTestConfig() PostgresTestConfig {
	return PostgresTestConfig{
		Host:     getEnv("PG_HOST", "localhost"),
		Port:     getEnv("PG_PORT", "5432"),
		User:     getEnv("PG_USER", "postgres"),
		Password: getEnv("PG_PASSWORD", "postgres"),
		Database: getEnv("PG_DATABASE", "paso_test"),
	}
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectionString returns PostgreSQL connection string
func (c PostgresTestConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

// SetupPostgresTestDB creates and returns a PostgreSQL test database with the full production
// schema applied via goose migrations. This ensures the test schema always matches production.
// IMPORTANT: Requires PostgreSQL to be running. Set environment variables or use defaults:
//   - PG_HOST (default: localhost)
//   - PG_PORT (default: 5432)
//   - PG_USER (default: postgres)
//   - PG_PASSWORD (default: postgres)
//   - PG_DATABASE (default: paso_test)
func SetupPostgresTestDB(tb testing.TB) *sql.DB {
	tb.Helper()

	config := DefaultPostgresTestConfig()

	// Try to connect to PostgreSQL
	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		tb.Skipf("PostgreSQL not available at %s:%s - skipping PostgreSQL tests. "+
			"Set PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DATABASE environment variables or use defaults",
			config.Host, config.Port)
		return nil
	}

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		tb.Skipf("PostgreSQL connection failed: %v. "+
			"Ensure PostgreSQL is running at %s:%s with user=%s, password=%s, database=%s",
			err, config.Host, config.Port, config.User, config.Password, config.Database)
		return nil
	}

	// Drop all existing tables to ensure a clean slate for goose migrations.
	// This handles both fresh databases and databases left over from previous test runs
	// (which may have used the old inline schema or an older migration version).
	dropAllPostgresTables(ctx, db)

	// Apply production migrations (schema only, no seed data)
	if err := database.RunMigrationsOnly(db, database.PostgreSQL); err != nil {
		tb.Fatalf("Failed to run PostgreSQL migrations: %v", err)
	}

	// Register cleanup
	tb.Cleanup(func() {
		cleanupPostgresTestData(db)
		_ = db.Close()
	})

	return db
}

// dropAllPostgresTables drops all application tables and the goose version table to ensure
// goose migrations run from scratch on each test setup. Tables are dropped in dependency order.
func dropAllPostgresTables(ctx context.Context, db *sql.DB) {
	drops := `
	DROP TABLE IF EXISTS task_events CASCADE;
	DROP TABLE IF EXISTS task_comments CASCADE;
	DROP TABLE IF EXISTS task_labels CASCADE;
	DROP TABLE IF EXISTS task_subtasks CASCADE;
	DROP TABLE IF EXISTS tasks CASCADE;
	DROP TABLE IF EXISTS assignees CASCADE;
	DROP TABLE IF EXISTS labels CASCADE;
	DROP TABLE IF EXISTS columns CASCADE;
	DROP TABLE IF EXISTS project_counters CASCADE;
	DROP TABLE IF EXISTS projects CASCADE;
	DROP TABLE IF EXISTS relation_types CASCADE;
	DROP TABLE IF EXISTS priorities CASCADE;
	DROP TABLE IF EXISTS types CASCADE;
	DROP TABLE IF EXISTS goose_db_version CASCADE;
	`
	_, _ = db.ExecContext(ctx, drops)
}

// cleanupPostgresTestData truncates all test tables to reset state for next test
func cleanupPostgresTestData(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cleanup := `
	TRUNCATE TABLE task_events CASCADE;
	TRUNCATE TABLE task_comments CASCADE;
	TRUNCATE TABLE task_labels CASCADE;
	TRUNCATE TABLE task_subtasks CASCADE;
	TRUNCATE TABLE tasks CASCADE;
	TRUNCATE TABLE assignees CASCADE;
	TRUNCATE TABLE labels CASCADE;
	TRUNCATE TABLE columns CASCADE;
	TRUNCATE TABLE project_counters CASCADE;
	TRUNCATE TABLE projects CASCADE;
	`

	_, _ = db.ExecContext(ctx, cleanup)
}
