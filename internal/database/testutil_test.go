package database

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

// ============================================================================
// DATABASE SETUP HELPERS
// ============================================================================

// setupTestDB creates an in-memory database and runs migrations
// This is the unified test database setup used by all tests
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Enable foreign key constraints
	_, err = db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err := runMigrations(context.Background(), db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clear seeded data for fresh tests
	_, err = db.ExecContext(context.Background(), "DELETE FROM columns")
	if err != nil {
		t.Fatalf("Failed to clear columns: %v", err)
	}
	_, err = db.ExecContext(context.Background(), "DELETE FROM labels")
	if err != nil {
		t.Fatalf("Failed to clear labels: %v", err)
	}

	return db
}

// setupTestDBFile creates a file-based database for testing persistence across restarts
func setupTestDBFile(t *testing.T) (*sql.DB, string) {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "paso-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	db, err := sql.Open("sqlite", tmpfile.Name())
	if err != nil {
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign key constraints
	_, err = db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err := runMigrations(context.Background(), db); err != nil {
		db.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clear default seeded columns for fresh tests
	_, err = db.ExecContext(context.Background(), "DELETE FROM columns")
	if err != nil {
		db.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to clear columns: %v", err)
	}

	return db, tmpfile.Name()
}

// closeAndReopenDB simulates app restart by closing and reopening the database
func closeAndReopenDB(t *testing.T, db *sql.DB, dbPath string) *sql.DB {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}

	newDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}

	// Enable foreign key constraints
	_, err = newDB.ExecContext(context.Background(), "PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	return newDB
}

// ============================================================================
// TEST ASSERTION HELPERS
// ============================================================================

// verifyLinkedListIntegrity checks that all columns are properly linked
func verifyLinkedListIntegrity(t *testing.T, ctx context.Context, repo *Repository, projectID int) {
	t.Helper()
	columns, err := repo.GetColumnsByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) == 0 {
		return // Empty list is valid
	}

	// Verify first column has nil prev_id
	if columns[0].PrevID != nil {
		t.Error("First column should have nil PrevID")
	}

	// Verify last column has nil next_id
	if columns[len(columns)-1].NextID != nil {
		t.Error("Last column should have nil NextID")
	}

	// Verify all middle columns have both pointers
	for i := 1; i < len(columns)-1; i++ {
		if columns[i].PrevID == nil {
			t.Errorf("Middle column %d should have non-nil PrevID", i)
		}
		if columns[i].NextID == nil {
			t.Errorf("Middle column %d should have non-nil NextID", i)
		}
	}

	// Verify pointers form valid chain
	for i := 0; i < len(columns)-1; i++ {
		if columns[i].NextID == nil || *columns[i].NextID != columns[i+1].ID {
			t.Errorf("Column %d NextID should point to column %d", i, i+1)
		}
		if columns[i+1].PrevID == nil || *columns[i+1].PrevID != columns[i].ID {
			t.Errorf("Column %d PrevID should point to column %d", i+1, i)
		}
	}
}
