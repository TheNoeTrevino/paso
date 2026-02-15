package fixtures

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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

// DefaultPostgresTestConfig returns default PostgreSQL test configuration from environment or defaults.
// Environment variables (if set): PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DATABASE
// Defaults: localhost:5432, postgres/postgres, paso_test
func DefaultPostgresTestConfig() PostgresTestConfig {
	return PostgresTestConfig{
		Host:     getEnv("PG_HOST", "localhost"),
		Port:     getEnv("PG_PORT", "5432"),
		User:     getEnv("PG_USER", "postgres"),
		Password: getEnv("PG_PASSWORD", "postgres"),
		Database: getEnv("PG_DATABASE", "paso_test"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectionString returns PostgreSQL connection string for the configured database.
func (c PostgresTestConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

// ConnectionStringForDB returns a connection string targeting a specific database name.
func (c PostgresTestConfig) ConnectionStringForDB(dbName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, dbName)
}

// SetupPostgresTestDB creates an isolated PostgreSQL test database with the full production
// schema applied via goose migrations. Each call creates a unique database (paso_test_<random>)
// to allow parallel test execution without conflicts.
//
// Requires PostgreSQL to be running. Set environment variables or use defaults:
//   - PG_HOST (default: localhost)
//   - PG_PORT (default: 5432)
//   - PG_USER (default: postgres)
//   - PG_PASSWORD (default: postgres)
//   - PG_DATABASE (default: paso_test) — used as the admin database for creating test databases
func SetupPostgresTestDB(tb testing.TB) *sql.DB {
	tb.Helper()

	config := DefaultPostgresTestConfig()

	// Connect to the admin database to create the test-specific database
	adminDB, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		tb.Skipf("PostgreSQL not available: %v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := adminDB.PingContext(ctx); err != nil {
		_ = adminDB.Close()
		tb.Skipf("PostgreSQL connection failed: %v. "+
			"Ensure PostgreSQL is running at %s:%s with database=%s",
			err, config.Host, config.Port, config.Database)
		return nil
	}

	// Generate a unique database name for this test
	testDBName := uniqueTestDBName()

	// CREATE DATABASE cannot run inside a transaction
	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", testDBName)); err != nil {
		_ = adminDB.Close()
		tb.Fatalf("Failed to create test database %s: %v", testDBName, err)
	}
	_ = adminDB.Close()

	// Connect to the new test-specific database
	testDB, err := sql.Open("postgres", config.ConnectionStringForDB(testDBName))
	if err != nil {
		dropTestDatabase(config, testDBName)
		tb.Fatalf("Failed to connect to test database %s: %v", testDBName, err)
	}

	// Apply production migrations (schema only, no seed data)
	if err := database.RunMigrationsOnly(testDB, database.PostgreSQL); err != nil {
		_ = testDB.Close()
		dropTestDatabase(config, testDBName)
		tb.Fatalf("Failed to run PostgreSQL migrations on %s: %v", testDBName, err)
	}

	// Register cleanup: close connection, then drop the database
	tb.Cleanup(func() {
		_ = testDB.Close()
		dropTestDatabase(config, testDBName)
	})

	return testDB
}

// uniqueTestDBName generates a unique database name like "paso_test_a1b2c3d4".
func uniqueTestDBName() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp if crypto/rand fails
		return fmt.Sprintf("paso_test_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("paso_test_%s", hex.EncodeToString(b))
}

// dropTestDatabase connects to the admin database and drops the specified test database.
func dropTestDatabase(config PostgresTestConfig, dbName string) {
	adminDB, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return
	}
	defer func() { _ = adminDB.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Terminate any remaining connections to the database before dropping
	terminateQuery := fmt.Sprintf(
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()",
		dbName,
	)
	_, _ = adminDB.ExecContext(ctx, terminateQuery)

	_, _ = adminDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
}
