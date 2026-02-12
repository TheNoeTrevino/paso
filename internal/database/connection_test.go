package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func TestConnection_SQLite(t *testing.T) {
	t.Parallel()
	// Create a temporary SQLite database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create the file
	f, err := os.Create(dbPath)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Test connection to SQLite database
	err = TestConnection(dbPath, SQLite)
	assert.NoError(t, err)
}

func TestConnection_SQLite_NonExistent(t *testing.T) {
	t.Parallel()
	// Test connection to non-existent SQLite database (should create it)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nonexistent.db")

	err := TestConnection(dbPath, SQLite)
	assert.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(dbPath)
	assert.False(t, os.IsNotExist(err))
}

func TestConnection_PostgreSQL_Invalid(t *testing.T) {
	t.Parallel()
	// Test connection to invalid PostgreSQL server (should fail)
	connStr := "host=invalid.invalid port=5432 user=test password=test dbname=test"

	err := TestConnection(connStr, PostgreSQL)
	assert.Error(t, err)
}

func TestConnection_UnsupportedType(t *testing.T) {
	t.Parallel()
	// Test with unsupported database type
	err := TestConnection("dummy", DatabaseType("unsupported"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "unsupported database type")
}

func TestConnection_PasswordSanitization(t *testing.T) {
	t.Parallel()
	// Test that error messages sanitize passwords
	// This will fail to connect but should not leak the password
	invalidConnStr := "postgres://user:secret123@nonexistent-host:5432/db"
	err := TestConnection(invalidConnStr, PostgreSQL)

	require.Error(t, err)

	// Error message should not contain the password
	assert.False(t, strings.Contains(err.Error(), "secret123"))

	// Error message should contain sanitized version
	assert.Contains(t, err.Error(), "***")
}

func TestSanitizeConnectionString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with password",
			input:    "postgres://user:secret123@localhost:5432/mydb",
			expected: "postgres://user:***@localhost:5432/mydb",
		},
		{
			name:     "URL without password",
			input:    "postgres://user@localhost:5432/mydb",
			expected: "postgres://user@localhost:5432/mydb",
		},
		{
			name:     "postgresql prefix with password",
			input:    "postgresql://admin:password@db.example.com:5432/production",
			expected: "postgresql://admin:***@db.example.com:5432/production",
		},
		{
			name:     "key-value with password",
			input:    "host=localhost port=5432 user=me password=secret dbname=mydb",
			expected: "host=localhost port=5432 user=me password=*** dbname=mydb",
		},
		{
			name:     "key-value without password",
			input:    "host=localhost port=5432 user=me dbname=mydb",
			expected: "host=localhost port=5432 user=me dbname=mydb",
		},
		{
			name:     "password with special characters",
			input:    "postgres://user:p#ssw0rd!@localhost/db",
			expected: "postgres://user:***@localhost/db",
		},
		{
			name:     "key-value password with special chars",
			input:    "host=localhost user=me password=abc@123=xyz dbname=db",
			expected: "host=localhost user=me password=*** dbname=db",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no credentials",
			input:    "some random text without creds",
			expected: "some random text without creds",
		},
		{
			name:     "multiple spaces between key-value pairs",
			input:    "host=localhost  password=secret  user=me",
			expected: "host=localhost  password=***  user=me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()
			result := sanitizeConnectionString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseConnectionString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty string",
			input:     "",
			wantError: true,
			errorMsg:  "connection string cannot be empty",
		},
		{
			name:      "whitespace only",
			input:     "   \t\n  ",
			wantError: true,
			errorMsg:  "connection string cannot be empty",
		},
		{
			name:      "postgres URL format",
			input:     "postgres://user:pass@localhost:5432/mydb",
			wantError: false,
		},
		{
			name:      "postgresql URL format",
			input:     "postgresql://user:pass@localhost:5432/mydb",
			wantError: false,
		},
		{
			name:      "key-value format",
			input:     "host=localhost port=5432 user=me password=pass dbname=mydb",
			wantError: false,
		},
		{
			name:      "invalid format no equals no prefix",
			input:     "localhost mydb user",
			wantError: true,
			errorMsg:  "invalid format",
		},
		{
			name:      "random text",
			input:     "not a connection string",
			wantError: true,
			errorMsg:  "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()
			result, err := ParseConnectionString(tt.input)
			if tt.wantError {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestParseURLConnectionString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
		errorMsg  string
	}{
		{
			name:  "complete URL with all parts",
			input: "postgres://myuser:mypass@localhost:5432/mydb",
			want:  "host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=require",
		},
		{
			name:  "postgresql prefix",
			input: "postgresql://myuser:mypass@localhost:5432/mydb",
			want:  "host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=require",
		},
		{
			name:  "missing port uses default 5432",
			input: "postgres://myuser:mypass@localhost/mydb",
			want:  "host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=require",
		},
		{
			name:  "missing password empty string",
			input: "postgres://myuser@localhost:5432/mydb",
			want:  "host=localhost port=5432 user=myuser password= dbname=mydb sslmode=require",
		},
		{
			name:      "missing user",
			input:     "postgres://:pass@localhost:5432/mydb",
			wantError: true,
			errorMsg:  "user not specified",
		},
		{
			name:      "missing host",
			input:     "postgres://user:pass@:5432/mydb",
			wantError: true,
			errorMsg:  "host not specified",
		},
		{
			name:      "missing database name",
			input:     "postgres://user:pass@localhost:5432",
			wantError: true,
			errorMsg:  "database name not specified",
		},
		{
			name:      "missing database slash only",
			input:     "postgres://user:pass@localhost:5432/",
			wantError: true,
			errorMsg:  "database name not specified",
		},
		{
			name:  "custom port",
			input: "postgres://user:pass@localhost:9999/mydb",
			want:  "host=localhost port=9999 user=user password=pass dbname=mydb sslmode=require",
		},
		{
			name:  "URL with query params application_name",
			input: "postgres://user:pass@localhost:5432/mydb?application_name=paso",
			want:  "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=require application_name=paso",
		},
		{
			name:  "URL with sslmode query param is overridden",
			input: "postgres://user:pass@localhost:5432/mydb?sslmode=disable",
			want:  "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=require",
		},
		{
			name:  "password with special characters",
			input: "postgres://user:p@ss=w0rd!@localhost/mydb",
			want:  "host=localhost port=5432 user=user password=p@ss=w0rd! dbname=mydb sslmode=require",
		},
		{
			name:  "IPv4 host",
			input: "postgres://user:pass@192.168.1.100:5432/mydb",
			want:  "host=192.168.1.100 port=5432 user=user password=pass dbname=mydb sslmode=require",
		},
		{
			name:  "domain name host",
			input: "postgres://user:pass@db.example.com:5432/mydb",
			want:  "host=db.example.com port=5432 user=user password=pass dbname=mydb sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()
			result, err := parseURLConnectionString(tt.input)
			if tt.wantError {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMsg)
			} else {
				require.NoError(t, err)
				// For query params, just check that the main parts are present
				// Order may vary for query params
				if strings.Contains(tt.want, "application_name") {
					assert.Contains(t, result, "host=")
					assert.Contains(t, result, "dbname=")
					assert.Contains(t, result, "application_name=paso")
				} else {
					assert.Equal(t, tt.want, result)
				}
			}
		})
	}
}

func TestParseKeyValueConnectionString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
		errorMsg  string
	}{
		{
			name:  "complete key-value string",
			input: "host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=require",
			want:  "host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=require",
		},
		{
			name:      "missing host",
			input:     "port=5432 user=myuser password=mypass dbname=mydb",
			wantError: true,
			errorMsg:  "host not specified",
		},
		{
			name:      "missing user",
			input:     "host=localhost port=5432 password=mypass dbname=mydb",
			wantError: true,
			errorMsg:  "user not specified",
		},
		{
			name:      "missing dbname",
			input:     "host=localhost port=5432 user=myuser password=mypass",
			wantError: true,
			errorMsg:  "database name not specified",
		},
		{
			name:  "missing port defaults to 5432",
			input: "host=localhost user=myuser password=mypass dbname=mydb",
			want:  "host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=require",
		},
		{
			name:  "missing sslmode defaults to require",
			input: "host=localhost port=5432 user=myuser password=mypass dbname=mydb",
			want:  "host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=require",
		},
		{
			name:  "missing password defaults to empty",
			input: "host=localhost port=5432 user=myuser dbname=mydb",
			want:  "host=localhost port=5432 user=myuser password= dbname=mydb sslmode=require",
		},
		{
			name:  "custom sslmode preserved",
			input: "host=localhost port=5432 user=myuser password=pass dbname=mydb sslmode=disable",
			want:  "host=localhost port=5432 user=myuser password=pass dbname=mydb sslmode=disable",
		},
		{
			name:  "custom port",
			input: "host=localhost port=9999 user=myuser password=pass dbname=mydb",
			want:  "host=localhost port=9999 user=myuser password=pass dbname=mydb sslmode=require",
		},
		{
			name:      "malformed pair no equals",
			input:     "host=localhost port5432 user=myuser dbname=mydb",
			wantError: true,
			errorMsg:  "invalid format",
		},
		{
			name:  "value containing equals sign",
			input: "host=localhost user=myuser password=abc=def dbname=mydb",
			want:  "host=localhost port=5432 user=myuser password=abc=def dbname=mydb sslmode=require",
		},
		{
			name:  "extra whitespace between pairs",
			input: "host=localhost   port=5432  user=myuser  password=pass  dbname=mydb",
			want:  "host=localhost port=5432 user=myuser password=pass dbname=mydb sslmode=require",
		},
		{
			name:  "minimal required fields only",
			input: "host=localhost user=myuser dbname=mydb",
			want:  "host=localhost port=5432 user=myuser password= dbname=mydb sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()
			result, err := parseKeyValueConnectionString(tt.input)
			if tt.wantError {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}
