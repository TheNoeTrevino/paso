package database

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
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
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clear seeded data for fresh tests
	_, err = db.Exec("DELETE FROM columns")
	if err != nil {
		t.Fatalf("Failed to clear columns: %v", err)
	}
	_, err = db.Exec("DELETE FROM labels")
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
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clear default seeded columns for fresh tests
	_, err = db.Exec("DELETE FROM columns")
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
	_, err = newDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	return newDB
}

// ============================================================================
// REPOSITORY WRAPPER FUNCTIONS
// These wrappers bridge the old test API to the new Repository pattern
// ============================================================================

// createRepo creates a Repository instance from a raw database connection
func createRepo(db *sql.DB) *Repository {
	return NewRepository(db)
}

// ----------------------------------------------------------------------------
// PROJECT OPERATIONS
// ----------------------------------------------------------------------------

// CreateProject creates a new project
func CreateProject(ctx context.Context, db *sql.DB, name, description string) (*models.Project, error) {
	return createRepo(db).CreateProject(ctx, name, description)
}

// ----------------------------------------------------------------------------
// COLUMN OPERATIONS
// ----------------------------------------------------------------------------

// GetAllColumns retrieves all columns for the default test project (ID 1)
func GetAllColumns(ctx context.Context, db *sql.DB) ([]*models.Column, error) {
	return createRepo(db).GetColumnsByProject(ctx, 1)
}

// CreateColumn creates a new column
func CreateColumn(ctx context.Context, db *sql.DB, name string, projectID int, afterID *int) (*models.Column, error) {
	return createRepo(db).CreateColumn(ctx, name, projectID, afterID)
}

// DeleteColumn deletes a column
func DeleteColumn(ctx context.Context, db *sql.DB, columnID int) error {
	return createRepo(db).DeleteColumn(ctx, columnID)
}

// ----------------------------------------------------------------------------
// TASK OPERATIONS
// ----------------------------------------------------------------------------

// CreateTask creates a new task
func CreateTask(ctx context.Context, db *sql.DB, title, description string, columnID, position int) (*models.Task, error) {
	return createRepo(db).CreateTask(ctx, title, description, columnID, position)
}

// GetTasksByColumn retrieves all tasks for a column with full details
func GetTasksByColumn(ctx context.Context, db *sql.DB, columnID int) ([]*models.Task, error) {
	return createRepo(db).GetTasksByColumn(ctx, columnID)
}

// GetTaskSummariesByColumn retrieves task summaries for a column
func GetTaskSummariesByColumn(ctx context.Context, db *sql.DB, columnID int) ([]*models.TaskSummary, error) {
	return createRepo(db).GetTaskSummariesByColumn(ctx, columnID)
}

// GetTaskDetail retrieves full task details including labels
func GetTaskDetail(ctx context.Context, db *sql.DB, taskID int) (*models.TaskDetail, error) {
	return createRepo(db).GetTaskDetail(ctx, taskID)
}

// GetTaskCountByColumn returns the number of tasks in a column
func GetTaskCountByColumn(ctx context.Context, db *sql.DB, columnID int) (int, error) {
	return createRepo(db).GetTaskCountByColumn(ctx, columnID)
}

// UpdateTask updates a task's title and description
func UpdateTask(ctx context.Context, db *sql.DB, taskID int, title, description string) error {
	return createRepo(db).UpdateTask(ctx, taskID, title, description)
}

// UpdateTaskTitle updates a task's title while preserving its description
func UpdateTaskTitle(ctx context.Context, db *sql.DB, taskID int, title string) error {
	// Get current task to preserve description
	detail, err := createRepo(db).GetTaskDetail(ctx, taskID)
	if err != nil {
		return err
	}
	return createRepo(db).UpdateTask(ctx, taskID, title, detail.Description)
}

// UpdateTaskColumn updates a task's column and position directly
// This method doesn't exist in the Repository, so we use direct SQL
func UpdateTaskColumn(ctx context.Context, db *sql.DB, taskID, columnID, position int) error {
	_, err := db.ExecContext(ctx,
		`UPDATE tasks SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		columnID, position, taskID)
	return err
}

// MoveTaskToNextColumn moves a task to the next column
func MoveTaskToNextColumn(ctx context.Context, db *sql.DB, taskID int) error {
	return createRepo(db).MoveTaskToNextColumn(ctx, taskID)
}

// MoveTaskToPrevColumn moves a task to the previous column
func MoveTaskToPrevColumn(ctx context.Context, db *sql.DB, taskID int) error {
	return createRepo(db).MoveTaskToPrevColumn(ctx, taskID)
}

// DeleteTask deletes a task
func DeleteTask(ctx context.Context, db *sql.DB, taskID int) error {
	return createRepo(db).DeleteTask(ctx, taskID)
}

// ----------------------------------------------------------------------------
// LABEL OPERATIONS
// ----------------------------------------------------------------------------

// CreateLabel creates a new label for a project
func CreateLabel(ctx context.Context, db *sql.DB, projectID int, name, color string) (*models.Label, error) {
	return createRepo(db).CreateLabel(ctx, projectID, name, color)
}

// GetAllLabels retrieves all labels across all projects
func GetAllLabels(ctx context.Context, db *sql.DB) ([]*models.Label, error) {
	// Query all labels directly since there's no Repository method for this
	rows, err := db.QueryContext(ctx, `SELECT id, name, color, project_id FROM labels ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*models.Label
	for rows.Next() {
		label := &models.Label{}
		if err := rows.Scan(&label.ID, &label.Name, &label.Color, &label.ProjectID); err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	return labels, rows.Err()
}

// GetLabelsByProject retrieves all labels for a project
func GetLabelsByProject(ctx context.Context, db *sql.DB, projectID int) ([]*models.Label, error) {
	return createRepo(db).GetLabelsByProject(ctx, projectID)
}

// GetLabelsForTask retrieves all labels associated with a task
func GetLabelsForTask(ctx context.Context, db *sql.DB, taskID int) ([]*models.Label, error) {
	return createRepo(db).GetLabelsForTask(ctx, taskID)
}

// SetTaskLabels sets the labels for a task
func SetTaskLabels(ctx context.Context, db *sql.DB, taskID int, labelIDs []int) error {
	return createRepo(db).SetTaskLabels(ctx, taskID, labelIDs)
}

// ============================================================================
// TEST ASSERTION HELPERS
// ============================================================================

// verifyLinkedListIntegrity checks that all columns are properly linked
func verifyLinkedListIntegrity(t *testing.T, db *sql.DB) {
	t.Helper()
	columns, err := GetAllColumns(context.Background(), db)
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
