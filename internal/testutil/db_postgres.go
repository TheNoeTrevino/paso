package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// PostgresTestConfig holds PostgreSQL test configuration
type PostgresTestConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// DefaultPostgresTestConfig returns default PostgreSQL test configuration from environment or defaults
// Environment variables (if set): PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DATABASE
// Defaults: localhost:5432, postgres/postgres, postgres
func DefaultPostgresTestConfig() PostgresTestConfig {
	return PostgresTestConfig{
		Host:     getEnv("PG_HOST", "localhost"),
		Port:     getEnv("PG_PORT", "5432"),
		User:     getEnv("PG_USER", "postgres"),
		Password: getEnv("PG_PASSWORD", "postgres"),
		Database: getEnv("PG_DATABASE", "paso_test"),
	}
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectionString returns PostgreSQL connection string
func (c PostgresTestConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

// SetupPostgresTestDB creates and returns a PostgreSQL test database
// It attempts to connect to PostgreSQL and creates a fresh test database
// IMPORTANT: Requires PostgreSQL to be running. Set environment variables or use defaults:
//   - PG_HOST (default: localhost)
//   - PG_PORT (default: 5432)
//   - PG_USER (default: postgres)
//   - PG_PASSWORD (default: postgres)
//   - PG_DATABASE (default: paso_test)
func SetupPostgresTestDB(tb testing.TB) *sql.DB {
	tb.Helper()

	config := DefaultPostgresTestConfig()

	// Try to connect to PostgreSQL
	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		tb.Skipf("PostgreSQL not available at %s:%s - skipping PostgreSQL tests. "+
			"Set PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DATABASE environment variables or use defaults",
			config.Host, config.Port)
		return nil
	}

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		tb.Skipf("PostgreSQL connection failed: %v. "+
			"Ensure PostgreSQL is running at %s:%s with user=%s, password=%s, database=%s",
			err, config.Host, config.Port, config.User, config.Password, config.Database)
		return nil
	}

	// Create schema
	if err := createPostgresTestSchema(ctx, db); err != nil {
		tb.Fatalf("Failed to create PostgreSQL schema: %v", err)
	}

	// Register cleanup
	tb.Cleanup(func() {
		// Clean up test data but keep database for next test
		cleanupPostgresTestData(db)
		_ = db.Close()
	})

	return db
}

// createPostgresTestSchema creates the complete database schema for PostgreSQL testing
func createPostgresTestSchema(ctx context.Context, db *sql.DB) error {
	schema := `
	-- Drop existing tables if they exist (for fresh test setup)
	DROP TABLE IF EXISTS task_comments CASCADE;
	DROP TABLE IF EXISTS task_labels CASCADE;
	DROP TABLE IF EXISTS task_subtasks CASCADE;
	DROP TABLE IF EXISTS tasks CASCADE;
	DROP TABLE IF EXISTS labels CASCADE;
	DROP TABLE IF EXISTS columns CASCADE;
	DROP TABLE IF EXISTS project_counters CASCADE;
	DROP TABLE IF EXISTS projects CASCADE;
	DROP TABLE IF EXISTS relation_types CASCADE;
	DROP TABLE IF EXISTS priorities CASCADE;
	DROP TABLE IF EXISTS types CASCADE;

	-- Types lookup table
	CREATE TABLE types (
		id SERIAL PRIMARY KEY,
		description TEXT NOT NULL UNIQUE
	);

	INSERT INTO types (id, description) VALUES
		(1, 'task'),
		(2, 'feature'),
		(3, 'bug');

	-- Priorities lookup table
	CREATE TABLE priorities (
		id SERIAL PRIMARY KEY,
		description TEXT NOT NULL UNIQUE,
		color TEXT NOT NULL
	);

	INSERT INTO priorities (id, description, color) VALUES
		(1, 'trivial', '#3B82F6'),
		(2, 'low', '#22C55E'),
		(3, 'medium', '#EAB308'),
		(4, 'high', '#F97316'),
		(5, 'critical', '#EF4444');

	-- Relation types lookup table
	CREATE TABLE relation_types (
		id SERIAL PRIMARY KEY,
		p_to_c_label TEXT NOT NULL,
		c_to_p_label TEXT NOT NULL,
		color TEXT NOT NULL,
		is_blocking BOOLEAN NOT NULL DEFAULT false
	);

	INSERT INTO relation_types (id, p_to_c_label, c_to_p_label, color, is_blocking) VALUES
		(1, 'Parent', 'Child', '#6B7280', false),
		(2, 'Blocked By', 'Blocker', '#EF4444', true),
		(3, 'Related To', 'Related To', '#3B82F6', false);

	-- Projects table
	CREATE TABLE projects (
		id BIGSERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Project counters for ticket numbers
	CREATE TABLE project_counters (
		project_id BIGINT PRIMARY KEY,
		next_ticket_number BIGINT DEFAULT 1,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	-- Columns table
	CREATE TABLE columns (
		id BIGSERIAL PRIMARY KEY,
		project_id BIGINT NOT NULL,
		name TEXT NOT NULL,
		prev_id BIGINT,
		next_id BIGINT,
		holds_ready_tasks BOOLEAN NOT NULL DEFAULT false,
		holds_completed_tasks BOOLEAN NOT NULL DEFAULT false,
		holds_in_progress_tasks BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	-- Tasks table
	CREATE TABLE tasks (
		id BIGSERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		column_id BIGINT NOT NULL,
		position BIGINT NOT NULL DEFAULT 0,
		ticket_number BIGINT,
		type_id BIGINT NOT NULL DEFAULT 1,
		priority_id BIGINT NOT NULL DEFAULT 3,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE,
		FOREIGN KEY (type_id) REFERENCES types(id),
		FOREIGN KEY (priority_id) REFERENCES priorities(id),
		UNIQUE(column_id, position)
	);

	-- Task relationships (parent-child, blocking, etc.)
	CREATE TABLE task_subtasks (
		parent_id BIGINT NOT NULL,
		child_id BIGINT NOT NULL,
		relation_type_id BIGINT NOT NULL DEFAULT 1,
		PRIMARY KEY (parent_id, child_id),
		FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (child_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (relation_type_id) REFERENCES relation_types(id)
	);

	-- Labels table
	CREATE TABLE labels (
		id BIGSERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		color TEXT NOT NULL,
		project_id BIGINT NOT NULL,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		UNIQUE(name, project_id)
	);

	-- Task-labels join table
	CREATE TABLE task_labels (
		task_id BIGINT NOT NULL,
		label_id BIGINT NOT NULL,
		PRIMARY KEY (task_id, label_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
	);

	-- Task comments table
	CREATE TABLE task_comments (
		id BIGSERIAL PRIMARY KEY,
		task_id BIGINT NOT NULL,
		content TEXT NOT NULL CHECK(length(content) <= 1000),
		author TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	-- Indexes for performance
	CREATE INDEX idx_tasks_column ON tasks(column_id, position);
	CREATE INDEX idx_columns_project ON columns(project_id);
	CREATE INDEX idx_labels_project ON labels(project_id);
	CREATE INDEX idx_task_labels_label ON task_labels(label_id);
	CREATE INDEX idx_task_subtasks_parent ON task_subtasks(parent_id);
	CREATE INDEX idx_task_subtasks_child ON task_subtasks(child_id);
	CREATE INDEX idx_task_comments_task ON task_comments(task_id);
	CREATE INDEX idx_tasks_column_id ON tasks(column_id);
	CREATE INDEX idx_task_labels_task_id ON task_labels(task_id);
	CREATE INDEX idx_labels_project_id ON labels(project_id);
	CREATE INDEX idx_columns_project_id ON columns(project_id);
	CREATE INDEX idx_task_subtasks_child_id ON task_subtasks(child_id);
	CREATE INDEX idx_task_comments_task_id ON task_comments(task_id);
	CREATE INDEX idx_tasks_type_id ON tasks(type_id);
	CREATE INDEX idx_tasks_priority_id ON tasks(priority_id);

	-- Unique partial indexes for column constraints
	CREATE UNIQUE INDEX idx_columns_ready_per_project ON columns(project_id) WHERE holds_ready_tasks = true;
	CREATE UNIQUE INDEX idx_columns_completed_per_project ON columns(project_id) WHERE holds_completed_tasks = true;
	CREATE UNIQUE INDEX idx_columns_in_progress_per_project ON columns(project_id) WHERE holds_in_progress_tasks = true;
	`

	_, err := db.ExecContext(ctx, schema)
	return err
}

// cleanupPostgresTestData truncates all test tables to reset state for next test
func cleanupPostgresTestData(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cleanup := `
	TRUNCATE TABLE task_comments CASCADE;
	TRUNCATE TABLE task_labels CASCADE;
	TRUNCATE TABLE task_subtasks CASCADE;
	TRUNCATE TABLE tasks CASCADE;
	TRUNCATE TABLE labels CASCADE;
	TRUNCATE TABLE columns CASCADE;
	TRUNCATE TABLE project_counters CASCADE;
	TRUNCATE TABLE projects CASCADE;
	`

	_, _ = db.ExecContext(ctx, cleanup)
}

// CreatePostgresTestProject creates a test project with default columns in PostgreSQL
func CreatePostgresTestProject(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var projectID int64
	err := db.QueryRowContext(ctx,
		"INSERT INTO projects (name, description) VALUES ($1, $2) RETURNING id",
		name, "Test description").Scan(&projectID)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Initialize project counter
	_, err = db.ExecContext(ctx,
		"INSERT INTO project_counters (project_id, next_ticket_number) VALUES ($1, 1)", projectID)
	if err != nil {
		t.Fatalf("Failed to initialize project counter: %v", err)
	}

	// Create default columns
	CreatePostgresTestColumn(t, db, projectID, "Todo")
	CreatePostgresTestColumn(t, db, projectID, "In Progress")
	CreatePostgresTestColumn(t, db, projectID, "Done")

	return projectID
}

// CreatePostgresTestColumn creates a test column in PostgreSQL
func CreatePostgresTestColumn(t *testing.T, db *sql.DB, projectID int64, name string) int64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var columnID int64
	err := db.QueryRowContext(ctx,
		"INSERT INTO columns (project_id, name) VALUES ($1, $2) RETURNING id",
		projectID, name).Scan(&columnID)
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}
	return columnID
}

// CreatePostgresTestTask creates a test task in PostgreSQL
func CreatePostgresTestTask(t *testing.T, db *sql.DB, columnID int64, title string) int64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the next position for this column
	var maxPosition int64
	err := db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(position), -1) FROM tasks WHERE column_id = $1", columnID).Scan(&maxPosition)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Failed to get max position: %v", err)
	}

	nextPosition := maxPosition + 1
	var taskID int64
	err = db.QueryRowContext(ctx,
		"INSERT INTO tasks (column_id, title, position, type_id, priority_id) VALUES ($1, $2, $3, 1, 3) RETURNING id",
		columnID, title, nextPosition).Scan(&taskID)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	return taskID
}

// CreatePostgresTestLabel creates a test label in PostgreSQL
func CreatePostgresTestLabel(t *testing.T, db *sql.DB, projectID int64, name, color string) int64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var labelID int64
	err := db.QueryRowContext(ctx,
		"INSERT INTO labels (project_id, name, color) VALUES ($1, $2, $3) RETURNING id",
		projectID, name, color).Scan(&labelID)
	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}
	return labelID
}
