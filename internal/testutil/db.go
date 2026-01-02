package testutil

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const TestAppKey ContextKey = "testApp"

// CaptureOutput captures stdout during function execution
func CaptureOutput(t *testing.T, fn func()) string {
	t.Helper()

	// Save original stdout
	oldStdout := os.Stdout

	// Create pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
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

// SetupTestDB creates an in-memory database with full schema
func SetupTestDB(t *testing.T) *sql.DB {
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

	// Run migrations inline
	if err := createTestSchema(db); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// createTestSchema creates the complete database schema for testing
func createTestSchema(db *sql.DB) error {
	schema := `
	-- Projects table
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Project counters for ticket numbers
	CREATE TABLE IF NOT EXISTS project_counters (
		project_id INTEGER PRIMARY KEY,
		next_ticket_number INTEGER DEFAULT 1,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	-- Columns table
	CREATE TABLE IF NOT EXISTS columns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		prev_id INTEGER,
		next_id INTEGER,
		holds_ready_tasks BOOLEAN NOT NULL DEFAULT 0,
		holds_completed_tasks BOOLEAN NOT NULL DEFAULT 0,
		holds_in_progress_tasks BOOLEAN NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	-- Types lookup table
	CREATE TABLE IF NOT EXISTS types (
		id INTEGER PRIMARY KEY,
		description TEXT NOT NULL UNIQUE
	);

	INSERT OR IGNORE INTO types (id, description) VALUES
		(1, 'task'),
		(2, 'feature'),
		(3, 'bug');

	-- Priorities lookup table
	CREATE TABLE IF NOT EXISTS priorities (
		id INTEGER PRIMARY KEY,
		description TEXT NOT NULL UNIQUE,
		color TEXT NOT NULL
	);

	INSERT OR IGNORE INTO priorities (id, description, color) VALUES
		(1, 'trivial', '#3B82F6'),
		(2, 'low', '#22C55E'),
		(3, 'medium', '#EAB308'),
		(4, 'high', '#F97316'),
		(5, 'critical', '#EF4444');

	-- Relation types
	CREATE TABLE IF NOT EXISTS relation_types (
		id INTEGER PRIMARY KEY,
		p_to_c_label TEXT NOT NULL,
		c_to_p_label TEXT NOT NULL,
		color TEXT NOT NULL,
		is_blocking BOOLEAN NOT NULL DEFAULT 0
	);

	INSERT OR IGNORE INTO relation_types (id, p_to_c_label, c_to_p_label, color, is_blocking) VALUES
		(1, 'Parent', 'Child', '#6B7280', 0),
		(2, 'Blocked By', 'Blocker', '#EF4444', 1),
		(3, 'Related To', 'Related To', '#3B82F6', 0);

	-- Tasks table
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		column_id INTEGER NOT NULL,
		position INTEGER NOT NULL DEFAULT 0,
		ticket_number INTEGER,
		type_id INTEGER NOT NULL DEFAULT 1,
		priority_id INTEGER NOT NULL DEFAULT 3,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE,
		FOREIGN KEY (type_id) REFERENCES types(id),
		FOREIGN KEY (priority_id) REFERENCES priorities(id),
		UNIQUE(column_id, position)
	);

	-- Task relationships (parent-child, blocking, etc.)
	CREATE TABLE IF NOT EXISTS task_subtasks (
		parent_id INTEGER NOT NULL,
		child_id INTEGER NOT NULL,
		relation_type_id INTEGER NOT NULL DEFAULT 1,
		PRIMARY KEY (parent_id, child_id),
		FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (child_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (relation_type_id) REFERENCES relation_types(id)
	);

	-- Labels table
	CREATE TABLE IF NOT EXISTS labels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		color TEXT NOT NULL,
		project_id INTEGER NOT NULL,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		UNIQUE(name, project_id)
	);

	-- Task-labels join table
	CREATE TABLE IF NOT EXISTS task_labels (
		task_id INTEGER NOT NULL,
		label_id INTEGER NOT NULL,
		PRIMARY KEY (task_id, label_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
	);

	-- Task comments table
	CREATE TABLE IF NOT EXISTS task_comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		content TEXT NOT NULL CHECK(length(content) <= 1000),
		author TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	-- Indexes for performance (from 00001_initial_schema)
	CREATE INDEX IF NOT EXISTS idx_tasks_column ON tasks(column_id, position);
	CREATE INDEX IF NOT EXISTS idx_columns_project ON columns(project_id);
	CREATE INDEX IF NOT EXISTS idx_labels_project ON labels(project_id);
	CREATE INDEX IF NOT EXISTS idx_task_labels_label ON task_labels(label_id);
	CREATE INDEX IF NOT EXISTS idx_task_subtasks_parent ON task_subtasks(parent_id);
	CREATE INDEX IF NOT EXISTS idx_task_subtasks_child ON task_subtasks(child_id);
	CREATE INDEX IF NOT EXISTS idx_task_comments_task ON task_comments(task_id);

	-- Unique partial indexes for column constraints
	CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_ready_per_project ON columns(project_id) WHERE holds_ready_tasks = 1;
	CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_completed_per_project ON columns(project_id) WHERE holds_completed_tasks = 1;
	CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_in_progress_per_project ON columns(project_id) WHERE holds_in_progress_tasks = 1;

	-- Additional performance indexes (from 00002_add_performance_indexes)
	CREATE INDEX IF NOT EXISTS idx_tasks_column_id ON tasks(column_id);
	CREATE INDEX IF NOT EXISTS idx_task_labels_task_id ON task_labels(task_id);
	CREATE INDEX IF NOT EXISTS idx_labels_project_id ON labels(project_id);
	CREATE INDEX IF NOT EXISTS idx_columns_project_id ON columns(project_id);
	CREATE INDEX IF NOT EXISTS idx_task_subtasks_child_id ON task_subtasks(child_id);
	CREATE INDEX IF NOT EXISTS idx_task_comments_task_id ON task_comments(task_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_type_id ON tasks(type_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_priority_id ON tasks(priority_id);

	-- Partial indexes for column type queries
	CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_ready_unique ON columns(project_id) WHERE holds_ready_tasks = 1;
	CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_completed_unique ON columns(project_id) WHERE holds_completed_tasks = 1;
	CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_in_progress_unique ON columns(project_id) WHERE holds_in_progress_tasks = 1;
	`

	_, err := db.ExecContext(context.Background(), schema)
	return err
}

// CreateTestProject creates a test project with default columns (Todo, In Progress, Done)
func CreateTestProject(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO projects (name, description) VALUES (?, ?)", name, "Test description")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Initialize project counter
	projectID, _ := result.LastInsertId()
	_, err = db.ExecContext(context.Background(), "INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)", projectID)
	if err != nil {
		t.Fatalf("Failed to initialize project counter: %v", err)
	}

	// Create default columns
	CreateTestColumn(t, db, int(projectID), "Todo")
	CreateTestColumn(t, db, int(projectID), "In Progress")
	CreateTestColumn(t, db, int(projectID), "Done")

	return int(projectID)
}

// CreateTestColumn creates a test column and returns its ID
func CreateTestColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, name)
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}
	columnID, _ := result.LastInsertId()
	return int(columnID)
}

// CreateTestTask creates a test task and returns its ID
func CreateTestTask(t *testing.T, db *sql.DB, columnID int, title string) int {
	t.Helper()
	// Get the next position for this column
	var maxPosition int
	err := db.QueryRowContext(context.Background(),
		"SELECT COALESCE(MAX(position), -1) FROM tasks WHERE column_id = ?", columnID).Scan(&maxPosition)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Failed to get max position: %v", err)
	}

	nextPosition := maxPosition + 1
	result, err := db.ExecContext(context.Background(),
		"INSERT INTO tasks (column_id, title, position, type_id, priority_id) VALUES (?, ?, ?, 1, 3)",
		columnID, title, nextPosition)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	taskID, _ := result.LastInsertId()
	return int(taskID)
}

// Note: SetupCLITest and ExecuteCLICommand are re-exported from testutil/cli package
// to maintain backward compatibility. They cannot be imported directly in this file
// to avoid import cycles, so they must be accessed via testutil.SetupCLITest() which
// dynamically loads them. For now, they are only available by importing testutil/cli directly.

// CreateTestLabel creates a test label and returns its ID
func CreateTestLabel(t *testing.T, db *sql.DB, projectID int, name, color string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO labels (project_id, name, color) VALUES (?, ?, ?)", projectID, name, color)
	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}
	labelID, _ := result.LastInsertId()
	return int(labelID)
}
