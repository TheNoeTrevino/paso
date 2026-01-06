// Package database handles database configuration and initialization
package database

import (
	"fmt"
	"strings"
)

// DatabaseType represents the type of database being used
type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	PostgreSQL DatabaseType = "postgres"
)

// Config holds the configuration needed to initialize a database connection
type Config struct {
	Type DatabaseType

	// For SQLite
	SQLitePath string

	// For PostgreSQL
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresConnStr  string // Original/full connection string with all query parameters

	// Connection pool configuration (applies to PostgreSQL)
	// MaxOpenConns specifies the maximum number of open connections to the database.
	// If set to 0 or negative, it will use the default value (25 for PostgreSQL).
	MaxOpenConns int
	// MaxIdleConns specifies the maximum number of idle connections.
	// If set to 0 or negative, it will use the default value (5 for PostgreSQL).
	MaxIdleConns int
}

// ConnectionString returns the appropriate connection string based on database type
func (c Config) ConnectionString() (string, error) {
	switch c.Type {
	case SQLite:
		return c.SQLitePath, nil
	case PostgreSQL:
		// Use the original connection string if available (preserves all query parameters)
		if c.PostgresConnStr != "" {
			return c.PostgresConnStr, nil
		}
		// Fall back to reconstructing the connection string from components
		// This is used for older connections that don't have the original string stored
		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
			c.PostgresHost,
			c.PostgresPort,
			c.PostgresUser,
			c.PostgresPassword,
			c.PostgresDB,
		)
		return connStr, nil
	default:
		return "", fmt.Errorf("failed to get connection string: unknown database type %s", c.Type)
	}
}

// IsLocalDatabase returns true if using local SQLite
func (c Config) IsLocalDatabase() bool {
	return c.Type == SQLite
}

// IsRemoteDatabase returns true if using remote PostgreSQL
func (c Config) IsRemoteDatabase() bool {
	return c.Type == PostgreSQL
}

// ParsePostgresConnectionString parses a normalized PostgreSQL connection string
// and returns a Config object with the parsed values.
// The normalized connection string should already have query parameters preserved.
func ParsePostgresConnectionString(connStr string) (Config, error) {
	// Parse the key-value format: host=... port=... user=... password=... dbname=...
	pairs := strings.Fields(connStr)
	params := make(map[string]string)

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}

	cfg := Config{
		Type:             PostgreSQL,
		PostgresHost:     params["host"],
		PostgresUser:     params["user"],
		PostgresPassword: params["password"],
		PostgresDB:       params["dbname"],
		PostgresConnStr:  connStr, // Store the full normalized connection string with all parameters
	}

	// Parse port
	if portStr, ok := params["port"]; ok {
		port := 5432 // default
		if p, err := fmt.Sscanf(portStr, "%d", &port); err == nil && p == 1 {
			cfg.PostgresPort = port
		}
	} else {
		cfg.PostgresPort = 5432
	}

	return cfg, nil
}
