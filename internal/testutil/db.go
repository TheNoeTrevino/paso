package testutil

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"os"
	"testing"

	"github.com/thenoetrevino/paso/internal/database"
	_ "modernc.org/sqlite"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const TestAppKey ContextKey = "testApp"

// CaptureOutput captures stdout during function execution
func CaptureOutput(tb testing.TB, fn func()) string {
	tb.Helper()

	// Save original stdout
	oldStdout := os.Stdout

	// Create pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		tb.Fatalf("Failed to create pipe: %v", err)
	}

	// Replace stdout with pipe writer
	os.Stdout = w

	// Channel to collect output
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// Execute function
	fn()

	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Get captured output
	return <-outC
}

// SetupTestDB creates an in-memory database with the full production schema applied via goose migrations.
// This ensures the test schema always matches production, eliminating schema drift.
func SetupTestDB(tb testing.TB) *sql.DB {
	tb.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		tb.Fatalf("Failed to create test database: %v", err)
	}

	// Enable foreign key constraints
	_, err = db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON")
	if err != nil {
		tb.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Apply production migrations (schema only, no seed data)
	if err := database.RunMigrationsOnly(db, database.SQLite); err != nil {
		tb.Fatalf("Failed to run migrations: %v", err)
	}

	tb.Cleanup(func() {
		if err := db.Close(); err != nil {
			tb.Logf("failed to close test db: %v", err)
		}
	})

	return db
}

// CreateTestProject creates a test project with default columns (Todo, In Progress, Done)
func CreateTestProject(tb testing.TB, db *sql.DB, name string) int {
	tb.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO projects (name, description) VALUES (?, ?)", name, "Test description")
	if err != nil {
		tb.Fatalf("Failed to create test project: %v", err)
	}

	// Initialize project counter
	projectID, _ := result.LastInsertId()
	_, err = db.ExecContext(context.Background(), "INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)", projectID)
	if err != nil {
		tb.Fatalf("Failed to initialize project counter: %v", err)
	}

	// Create default columns
	CreateTestColumn(tb, db, int(projectID), "Todo")
	CreateTestColumn(tb, db, int(projectID), "In Progress")
	CreateTestColumn(tb, db, int(projectID), "Done")

	return int(projectID)
}

// CreateTestColumn creates a test column and returns its ID
func CreateTestColumn(tb testing.TB, db *sql.DB, projectID int, name string) int {
	tb.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, name)
	if err != nil {
		tb.Fatalf("Failed to create test column: %v", err)
	}
	columnID, _ := result.LastInsertId()
	return int(columnID)
}

// CreateTestTask creates a test task and returns its ID
func CreateTestTask(tb testing.TB, db *sql.DB, columnID int, title string) int {
	tb.Helper()
	// Get the next position for this column
	var maxPosition int
	err := db.QueryRowContext(context.Background(),
		"SELECT COALESCE(MAX(position), -1) FROM tasks WHERE column_id = ?", columnID).Scan(&maxPosition)
	if err != nil && err != sql.ErrNoRows {
		tb.Fatalf("Failed to get max position: %v", err)
	}

	nextPosition := maxPosition + 1
	result, err := db.ExecContext(context.Background(),
		"INSERT INTO tasks (column_id, title, position, type_id, priority_id) VALUES (?, ?, ?, 1, 3)",
		columnID, title, nextPosition)
	if err != nil {
		tb.Fatalf("Failed to create test task: %v", err)
	}
	taskID, _ := result.LastInsertId()
	return int(taskID)
}

// Note: SetupCLITest and ExecuteCLICommand are re-exported from testutil/cli package
// to maintain backward compatibility. They cannot be imported directly in this file
// to avoid import cycles, so they must be accessed via testutil.SetupCLITest() which
// dynamically loads them. For now, they are only available by importing testutil/cli directly.

// CreateTestLabel creates a test label and returns its ID
func CreateTestLabel(tb testing.TB, db *sql.DB, projectID int, name, color string) int {
	tb.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO labels (project_id, name, color) VALUES (?, ?, ?)", projectID, name, color)
	if err != nil {
		tb.Fatalf("Failed to create test label: %v", err)
	}
	labelID, _ := result.LastInsertId()
	return int(labelID)
}

// CreateTestAssignee creates a test assignee and returns its ID
func CreateTestAssignee(tb testing.TB, db *sql.DB, name string) int {
	tb.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO assignees (name) VALUES (?)", name)
	if err != nil {
		tb.Fatalf("Failed to create test assignee: %v", err)
	}
	assigneeID, _ := result.LastInsertId()
	return int(assigneeID)
}
