package testutil

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
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
func CreateTestProject(tb testing.TB, db *sql.DB, d Dialect, name string) int {
	tb.Helper()
	ctx := context.Background()
	query := fmt.Sprintf(
		"INSERT INTO projects (name, description) VALUES (%s, %s)",
		d.Placeholder(1), d.Placeholder(2))
	projectID, err := d.InsertReturningID(ctx, db, query, name, "Test description")
	if err != nil {
		tb.Fatalf("Failed to create test project: %v", err)
	}

	// Initialize project counter
	counterQuery := fmt.Sprintf(
		"INSERT INTO project_counters (project_id, next_ticket_number) VALUES (%s, 1)",
		d.Placeholder(1))
	_, err = db.ExecContext(ctx, counterQuery, projectID)
	if err != nil {
		tb.Fatalf("Failed to initialize project counter: %v", err)
	}

	// Create default columns
	CreateTestColumn(tb, db, d, projectID, "Todo")
	CreateTestColumn(tb, db, d, projectID, "In Progress")
	CreateTestColumn(tb, db, d, projectID, "Done")

	return projectID
}

// CreateTestColumn creates a test column and returns its ID
func CreateTestColumn(tb testing.TB, db *sql.DB, d Dialect, projectID int, name string) int {
	tb.Helper()
	ctx := context.Background()
	query := fmt.Sprintf(
		"INSERT INTO columns (project_id, name) VALUES (%s, %s)",
		d.Placeholder(1), d.Placeholder(2))
	columnID, err := d.InsertReturningID(ctx, db, query, projectID, name)
	if err != nil {
		tb.Fatalf("Failed to create test column: %v", err)
	}
	return columnID
}

// CreateTestTask creates a test task and returns its ID
func CreateTestTask(tb testing.TB, db *sql.DB, d Dialect, columnID int, title string) int {
	tb.Helper()
	ctx := context.Background()

	// Get the next position for this column
	var maxPosition int
	posQuery := fmt.Sprintf(
		"SELECT COALESCE(MAX(position), -1) FROM tasks WHERE column_id = %s",
		d.Placeholder(1))
	err := db.QueryRowContext(ctx, posQuery, columnID).Scan(&maxPosition)
	if err != nil && err != sql.ErrNoRows {
		tb.Fatalf("Failed to get max position: %v", err)
	}

	nextPosition := maxPosition + 1
	query := fmt.Sprintf(
		"INSERT INTO tasks (column_id, title, position, type_id, priority_id) VALUES (%s, %s, %s, 1, 3)",
		d.Placeholder(1), d.Placeholder(2), d.Placeholder(3))
	taskID, err := d.InsertReturningID(ctx, db, query, columnID, title, nextPosition)
	if err != nil {
		tb.Fatalf("Failed to create test task: %v", err)
	}
	return taskID
}

// Note: SetupCLITest and ExecuteCLICommand are re-exported from testutil/cli package
// to maintain backward compatibility. They cannot be imported directly in this file
// to avoid import cycles, so they must be accessed via testutil.SetupCLITest() which
// dynamically loads them. For now, they are only available by importing testutil/cli directly.

// CreateTestLabel creates a test label and returns its ID
func CreateTestLabel(tb testing.TB, db *sql.DB, d Dialect, projectID int, name, color string) int {
	tb.Helper()
	ctx := context.Background()
	query := fmt.Sprintf(
		"INSERT INTO labels (project_id, name, color) VALUES (%s, %s, %s)",
		d.Placeholder(1), d.Placeholder(2), d.Placeholder(3))
	labelID, err := d.InsertReturningID(ctx, db, query, projectID, name, color)
	if err != nil {
		tb.Fatalf("Failed to create test label: %v", err)
	}
	return labelID
}

// CreateTestAssignee creates a test assignee and returns its ID
func CreateTestAssignee(tb testing.TB, db *sql.DB, d Dialect, name string) int {
	tb.Helper()
	ctx := context.Background()
	query := fmt.Sprintf(
		"INSERT INTO assignees (name) VALUES (%s)",
		d.Placeholder(1))
	assigneeID, err := d.InsertReturningID(ctx, db, query, name)
	if err != nil {
		tb.Fatalf("Failed to create test assignee: %v", err)
	}
	return assigneeID
}

// GetTaskColumnID retrieves the column_id for a task by its ID
func GetTaskColumnID(tb testing.TB, db *sql.DB, d Dialect, taskID int) int {
	tb.Helper()
	query := fmt.Sprintf("SELECT column_id FROM tasks WHERE id = %s", d.Placeholder(1))
	var columnID int
	err := db.QueryRowContext(context.Background(), query, taskID).Scan(&columnID)
	if err != nil {
		tb.Fatalf("Failed to query task column_id for task %d: %v", taskID, err)
	}
	return columnID
}

// GetTaskPosition retrieves the position for a task by its ID
func GetTaskPosition(tb testing.TB, db *sql.DB, d Dialect, taskID int) int {
	tb.Helper()
	query := fmt.Sprintf("SELECT position FROM tasks WHERE id = %s", d.Placeholder(1))
	var position int
	err := db.QueryRowContext(context.Background(), query, taskID).Scan(&position)
	if err != nil {
		tb.Fatalf("Failed to query task position for task %d: %v", taskID, err)
	}
	return position
}

// AssertTaskInColumn verifies a task is in the expected column, failing the test if not
func AssertTaskInColumn(tb testing.TB, db *sql.DB, d Dialect, taskID, expectedColumnID int) {
	tb.Helper()
	actualColumnID := GetTaskColumnID(tb, db, d, taskID)
	if actualColumnID != expectedColumnID {
		tb.Errorf("Expected task %d in column %d, got column %d", taskID, expectedColumnID, actualColumnID)
	}
}

// AssertTaskLabelCount verifies the number of labels associated with a task
func AssertTaskLabelCount(tb testing.TB, db *sql.DB, d Dialect, taskID, expectedCount int) {
	tb.Helper()
	query := fmt.Sprintf("SELECT COUNT(*) FROM task_labels WHERE task_id = %s", d.Placeholder(1))
	var count int
	err := db.QueryRowContext(context.Background(), query, taskID).Scan(&count)
	if err != nil {
		tb.Fatalf("Failed to query label count for task %d: %v", taskID, err)
	}
	if count != expectedCount {
		tb.Errorf("Expected task %d to have %d labels, got %d", taskID, expectedCount, count)
	}
}

// AssertRelationExists verifies that a specific relation exists in the database
func AssertRelationExists(tb testing.TB, db *sql.DB, d Dialect, taskID, relatedTaskID int, relationType string) {
	tb.Helper()
	query := fmt.Sprintf(
		"SELECT EXISTS(SELECT 1 FROM task_relations WHERE task_id = %s AND related_task_id = %s AND relation_type = %s)",
		d.Placeholder(1), d.Placeholder(2), d.Placeholder(3))
	var exists bool
	err := db.QueryRowContext(context.Background(), query,
		taskID, relatedTaskID, relationType).Scan(&exists)
	if err != nil {
		tb.Fatalf("Failed to check relation existence: %v", err)
	}
	if !exists {
		tb.Errorf("Expected relation (task=%d, related=%d, type=%s) to exist", taskID, relatedTaskID, relationType)
	}
}

// AssertRelationNotExists verifies that a specific relation does NOT exist in the database
func AssertRelationNotExists(tb testing.TB, db *sql.DB, d Dialect, taskID, relatedTaskID int, relationType string) {
	tb.Helper()
	query := fmt.Sprintf(
		"SELECT EXISTS(SELECT 1 FROM task_relations WHERE task_id = %s AND related_task_id = %s AND relation_type = %s)",
		d.Placeholder(1), d.Placeholder(2), d.Placeholder(3))
	var exists bool
	err := db.QueryRowContext(context.Background(), query,
		taskID, relatedTaskID, relationType).Scan(&exists)
	if err != nil {
		tb.Fatalf("Failed to check relation existence: %v", err)
	}
	if exists {
		tb.Errorf("Expected relation (task=%d, related=%d, type=%s) to NOT exist", taskID, relatedTaskID, relationType)
	}
}

// AssertTaskAssigneeCount verifies the number of assignees associated with a task
func AssertTaskAssigneeCount(tb testing.TB, db *sql.DB, d Dialect, taskID, expectedCount int) {
	tb.Helper()
	query := fmt.Sprintf("SELECT COUNT(*) FROM task_assignees WHERE task_id = %s", d.Placeholder(1))
	var count int
	err := db.QueryRowContext(context.Background(), query, taskID).Scan(&count)
	if err != nil {
		tb.Fatalf("Failed to query assignee count for task %d: %v", taskID, err)
	}
	if count != expectedCount {
		tb.Errorf("Expected task %d to have %d assignees, got %d", taskID, expectedCount, count)
	}
}
