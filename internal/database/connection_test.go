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
