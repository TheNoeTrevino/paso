package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// sanitizeConnectionString removes sensitive credentials from connection strings in error messages.
// It replaces passwords in both URL format (postgres://user:password@host) and
// key-value format (password=value) with masked asterisks.
// This prevents accidental leakage of credentials in logs, error messages, or debugging output.
func sanitizeConnectionString(connStr string) string {
	// Sanitize URL format: postgres://user:***@host
	urlPattern := regexp.MustCompile(`(postgres(?:ql)?://[^:]+:)[^@]+(@)`)
	sanitized := urlPattern.ReplaceAllString(connStr, "$1***$2")

	// Sanitize key-value format: password=***
	kvPattern := regexp.MustCompile(`(password=)[^\s]+`)
	sanitized = kvPattern.ReplaceAllString(sanitized, "$1***")

	return sanitized
}

// ParseConnectionString validates and normalizes a PostgreSQL connection string.
// Supports multiple formats:
//   - postgres://user:password@host:port/database
//   - postgresql://user:password@host:port/database
//   - host=localhost port=5432 user=myuser password=mypass dbname=paso
//
// Returns the normalized connection string or an error if the format is invalid.
func ParseConnectionString(connStr string) (string, error) {
	connStr = strings.TrimSpace(connStr)
	if connStr == "" {
		return "", fmt.Errorf("failed to parse connection string: connection string cannot be empty")
	}

	// Check if it's a URL-based connection string (postgres://...)
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		return parseURLConnectionString(connStr)
	}

	// Check if it's a key-value connection string (host=... port=... etc)
	if strings.Contains(connStr, "=") {
		return parseKeyValueConnectionString(connStr)
	}

	return "", fmt.Errorf("failed to parse connection string: invalid format - must be postgres://user:password@host:port/database or key=value format")
}

// parseURLConnectionString parses a PostgreSQL URL-based connection string
func parseURLConnectionString(connStr string) (string, error) {
	// Replace postgres:// with postgresql:// for compatibility
	if strings.HasPrefix(connStr, "postgres://") {
		connStr = "postgresql://" + connStr[11:]
	}

	u, err := url.Parse(connStr)
	if err != nil {
		// Sanitize the connection string in the error message to prevent credential leakage
		sanitized := sanitizeConnectionString(connStr)
		return "", fmt.Errorf("failed to parse PostgreSQL connection URL %s: %w", sanitized, err)
	}

	// Extract components
	user := u.User.Username()
	password, hasPassword := u.User.Password()
	host := u.Hostname()
	port := u.Port()
	database := strings.TrimPrefix(u.Path, "/")

	// Validate required fields
	if user == "" {
		return "", fmt.Errorf("failed to parse connection string: user not specified")
	}
	if host == "" {
		return "", fmt.Errorf("failed to parse connection string: host not specified")
	}
	if database == "" {
		return "", fmt.Errorf("failed to parse connection string: database name not specified")
	}

	// Set defaults
	if port == "" {
		port = "5432"
	}
	if !hasPassword {
		password = ""
	}

	// Build normalized connection string in key-value format
	normalized := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, database,
	)

	// Preserve query parameters from the original URL (except those we already handle)
	query := u.Query()
	for key, values := range query {
		// Skip sslmode since we handle it explicitly
		if key == "sslmode" {
			continue
		}
		// Add other query parameters
		if len(values) > 0 {
			normalized += " " + key + "=" + values[0]
		}
	}

	return normalized, nil
}

// parseKeyValueConnectionString parses a PostgreSQL key-value connection string
func parseKeyValueConnectionString(connStr string) (string, error) {
	// Parse the connection string to validate format
	pairs := strings.Fields(connStr)
	params := make(map[string]string)

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("failed to parse connection string: invalid format %s", pair)
		}
		params[parts[0]] = parts[1]
	}

	// Validate required fields
	if _, hasHost := params["host"]; !hasHost {
		return "", fmt.Errorf("failed to parse connection string: host not specified")
	}
	if _, hasUser := params["user"]; !hasUser {
		return "", fmt.Errorf("failed to parse connection string: user not specified")
	}
	if _, hasDBName := params["dbname"]; !hasDBName {
		return "", fmt.Errorf("failed to parse connection string: database name not specified")
	}

	// Set defaults
	port := "5432"
	if p, ok := params["port"]; ok {
		port = p
	}

	sslmode := "require"
	if s, ok := params["sslmode"]; ok {
		sslmode = s
	}

	password := ""
	if p, ok := params["password"]; ok {
		password = p
	}

	// Return normalized connection string
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		params["host"], port, params["user"], password, params["dbname"], sslmode,
	), nil
}

// TestConnection attempts to connect to a database using the provided connection string and type.
// This is useful for validating credentials before saving the connection to config.
func TestConnection(connStr string, dbType DatabaseType) error {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Determine driver name based on database type
	var driverName string
	switch dbType {
	case SQLite:
		driverName = "sqlite"
	case PostgreSQL:
		driverName = "postgres"
	default:
		return fmt.Errorf("failed to test connection: unsupported database type %s", dbType)
	}

	// Attempt to open a connection
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		// Sanitize the connection string in error messages to prevent credential leakage
		sanitized := sanitizeConnectionString(connStr)
		return fmt.Errorf("failed to open database connection %s: %w", sanitized, err)
	}
	defer func() {
		_ = db.Close() // Ignore close errors - the important part is whether ping succeeded
	}()

	// Attempt to ping the database
	if err := db.PingContext(ctx); err != nil {
		// Sanitize the connection string in error messages to prevent credential leakage
		sanitized := sanitizeConnectionString(connStr)
		return fmt.Errorf("failed to connect to database %s: %w", sanitized, err)
	}

	return nil
}
