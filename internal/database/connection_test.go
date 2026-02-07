package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestConnection_SQLite(t *testing.T) {
	// Create a temporary SQLite database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create the file
	f, err := os.Create(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close test db file: %v", err)
	}

	// Test connection to SQLite database
	err = TestConnection(dbPath, SQLite)
	if err != nil {
		t.Errorf("SQLite connection test failed: %v", err)
	}
}

func TestConnection_SQLite_NonExistent(t *testing.T) {
	// Test connection to non-existent SQLite database (should create it)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nonexistent.db")

	err := TestConnection(dbPath, SQLite)
	if err != nil {
		t.Errorf("SQLite connection to non-existent file should succeed (creates file): %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("SQLite connection should have created the database file")
	}
}

func TestConnection_PostgreSQL_Invalid(t *testing.T) {
	// Test connection to invalid PostgreSQL server (should fail)
	connStr := "host=invalid.invalid port=5432 user=test password=test dbname=test"

	err := TestConnection(connStr, PostgreSQL)
	if err == nil {
		t.Error("PostgreSQL connection to invalid host should fail")
	}
}

func TestConnection_UnsupportedType(t *testing.T) {
	// Test with unsupported database type
	err := TestConnection("dummy", DatabaseType("unsupported"))
	if err == nil {
		t.Error("Connection with unsupported database type should fail")
	}
	if err != nil && !strings.Contains(err.Error(), "unsupported database type") {
		t.Errorf("Expected 'unsupported database type' error, got: %v", err)
	}
}

func TestConnection_PasswordSanitization(t *testing.T) {
	// Test that error messages sanitize passwords
	// This will fail to connect but should not leak the password
	invalidConnStr := "postgres://user:secret123@nonexistent-host:5432/db"
	err := TestConnection(invalidConnStr, PostgreSQL)

	if err == nil {
		t.Fatal("Expected error for connection to nonexistent host")
	}

	// Error message should not contain the password
	if strings.Contains(err.Error(), "secret123") {
		t.Errorf("Error message contains password 'secret123': %v", err)
	}

	// Error message should contain sanitized version
	if !strings.Contains(err.Error(), "***") {
		t.Errorf("Error message should contain sanitized password (***): %v", err)
	}
}

func TestSanitizeConnectionString(t *testing.T) {
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
			result := sanitizeConnectionString(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeConnectionString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseConnectionString(t *testing.T) {
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
			result, err := ParseConnectionString(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseConnectionString() expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ParseConnectionString() error = %q, want substring %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParseConnectionString() unexpected error: %v", err)
				}
				if result == "" {
					t.Error("ParseConnectionString() returned empty string for valid input")
				}
			}
		})
	}
}

func TestParseURLConnectionString(t *testing.T) {
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
			result, err := parseURLConnectionString(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("parseURLConnectionString() expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("parseURLConnectionString() error = %q, want substring %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("parseURLConnectionString() unexpected error: %v", err)
				}
				// For query params, just check that the main parts are present
				// Order may vary for query params
				if strings.Contains(tt.want, "application_name") {
					if !strings.Contains(result, "host=") || !strings.Contains(result, "dbname=") || !strings.Contains(result, "application_name=paso") {
						t.Errorf("parseURLConnectionString() = %q, want to contain basic params + application_name from %q", result, tt.want)
					}
				} else if result != tt.want {
					t.Errorf("parseURLConnectionString() = %q, want %q", result, tt.want)
				}
			}
		})
	}
}

func TestParseKeyValueConnectionString(t *testing.T) {
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
			result, err := parseKeyValueConnectionString(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("parseKeyValueConnectionString() expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("parseKeyValueConnectionString() error = %q, want substring %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("parseKeyValueConnectionString() unexpected error: %v", err)
				}
				if result != tt.want {
					t.Errorf("parseKeyValueConnectionString() = %q, want %q", result, tt.want)
				}
			}
		})
	}
}
