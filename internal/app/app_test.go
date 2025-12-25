package app

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	_, err = db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	return db
}

func TestNew(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Create app with nil event client
	app := New(db, nil)

	if app == nil {
		t.Fatal("Expected app to be created, got nil")
	}

	if app.TaskService == nil {
		t.Error("Expected TaskService to be initialized")
	}

	if app.ProjectService == nil {
		t.Error("Expected ProjectService to be initialized")
	}

	if app.ColumnService == nil {
		t.Error("Expected ColumnService to be initialized")
	}

	if app.LabelService == nil {
		t.Error("Expected LabelService to be initialized")
	}
}

func TestClose(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	app := New(db, nil)

	err := app.Close()
	if err != nil {
		t.Errorf("Expected Close to succeed, got error: %v", err)
	}
}
