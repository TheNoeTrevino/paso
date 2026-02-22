package fixtures

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/database"
	_ "modernc.org/sqlite"
)

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

// DefaultTestLabelColor is the default color used for test labels when the specific color doesn't matter.
const DefaultTestLabelColor = "#FF5733"

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

	// Create default columns with proper flags
	CreateColumnWithFlags(tb, db, d, projectID, "Todo", true, false, false)
	CreateColumnWithFlags(tb, db, d, projectID, "In Progress", false, false, true)
	CreateColumnWithFlags(tb, db, d, projectID, "Done", false, true, false)

	return projectID
}

// CreateTestProjectWithBranch creates a test project with default columns and a git branch association.
func CreateTestProjectWithBranch(tb testing.TB, db *sql.DB, d Dialect, name, branch string) int {
	tb.Helper()
	projectID := CreateTestProject(tb, db, d, name)
	ctx := context.Background()
	query := fmt.Sprintf("UPDATE projects SET git_branch = %s WHERE id = %s",
		d.Placeholder(1), d.Placeholder(2))
	_, err := db.ExecContext(ctx, query, branch, projectID)
	if err != nil {
		tb.Fatalf("Failed to set git branch on project: %v", err)
	}
	return projectID
}

// CreateBareProject creates a test project without any default columns.
// Use this when you need to test behavior with an empty project or custom column layouts.
func CreateBareProject(tb testing.TB, db *sql.DB, d Dialect, name string) int {
	tb.Helper()
	ctx := context.Background()
	query := fmt.Sprintf(
		"INSERT INTO projects (name, description) VALUES (%s, %s)",
		d.Placeholder(1), d.Placeholder(2))
	projectID, err := d.InsertReturningID(ctx, db, query, name, "Test description")
	if err != nil {
		tb.Fatalf("Failed to create bare project: %v", err)
	}

	counterQuery := fmt.Sprintf(
		"INSERT INTO project_counters (project_id, next_ticket_number) VALUES (%s, 1)",
		d.Placeholder(1))
	_, err = db.ExecContext(ctx, counterQuery, projectID)
	if err != nil {
		tb.Fatalf("Failed to initialize project counter: %v", err)
	}

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

// CreateColumnWithFlags creates a column with specific flag settings and returns its ID.
// Use this when you need a column that holds ready, completed, or in-progress tasks.
func CreateColumnWithFlags(tb testing.TB, db *sql.DB, d Dialect, projectID int, name string, holdsReady, holdsCompleted, holdsInProgress bool) int {
	tb.Helper()
	ctx := context.Background()
	query := fmt.Sprintf(
		"INSERT INTO columns (project_id, name, holds_ready_tasks, holds_completed_tasks, holds_in_progress_tasks) VALUES (%s, %s, %s, %s, %s)",
		d.Placeholder(1), d.Placeholder(2), d.Placeholder(3), d.Placeholder(4), d.Placeholder(5))
	columnID, err := d.InsertReturningID(ctx, db, query, projectID, name, holdsReady, holdsCompleted, holdsInProgress)
	if err != nil {
		tb.Fatalf("Failed to create column with flags: %v", err)
	}
	return columnID
}

// LinkColumns sets the next_id on fromColumn and prev_id on toColumn, creating a column ordering.
func LinkColumns(tb testing.TB, db *sql.DB, d Dialect, fromColumnID, toColumnID int) {
	tb.Helper()
	ctx := context.Background()
	query := fmt.Sprintf("UPDATE columns SET next_id = %s WHERE id = %s",
		d.Placeholder(1), d.Placeholder(2))
	_, err := db.ExecContext(ctx, query, toColumnID, fromColumnID)
	if err != nil {
		tb.Fatalf("Failed to set next_id on column %d: %v", fromColumnID, err)
	}
	query = fmt.Sprintf("UPDATE columns SET prev_id = %s WHERE id = %s",
		d.Placeholder(1), d.Placeholder(2))
	_, err = db.ExecContext(ctx, query, fromColumnID, toColumnID)
	if err != nil {
		tb.Fatalf("Failed to set prev_id on column %d: %v", toColumnID, err)
	}
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

// GetColumnIDByName retrieves a column ID by project ID and column name
func GetColumnIDByName(tb testing.TB, db *sql.DB, d Dialect, projectID int, name string) int {
	tb.Helper()
	query := fmt.Sprintf(
		"SELECT id FROM columns WHERE project_id = %s AND name = %s",
		d.Placeholder(1), d.Placeholder(2))
	var columnID int
	err := db.QueryRowContext(context.Background(), query, projectID, name).Scan(&columnID)
	if err != nil {
		tb.Fatalf("Failed to get column ID for project %d, name %q: %v", projectID, name, err)
	}
	return columnID
}

// AttachLabelToTask attaches a label to a task for test setup
func AttachLabelToTask(tb testing.TB, db *sql.DB, d Dialect, taskID, labelID int) {
	tb.Helper()
	query := fmt.Sprintf(
		"INSERT INTO task_labels (task_id, label_id) VALUES (%s, %s)",
		d.Placeholder(1), d.Placeholder(2))
	_, err := db.ExecContext(context.Background(), query, taskID, labelID)
	if err != nil {
		tb.Fatalf("Failed to attach label %d to task %d: %v", labelID, taskID, err)
	}
}

// AddTaskSubtask creates a parent-child or blocking relationship between tasks.
// Use relationTypeID=1 for parent/child, relationTypeID=2 for blocking.
func AddTaskSubtask(tb testing.TB, db *sql.DB, d Dialect, parentID, childID, relationTypeID int) {
	tb.Helper()
	query := fmt.Sprintf(
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, %s)",
		d.Placeholder(1), d.Placeholder(2), d.Placeholder(3))
	_, err := db.ExecContext(context.Background(), query, parentID, childID, relationTypeID)
	if err != nil {
		tb.Fatalf("Failed to add subtask relation (parent=%d, child=%d, type=%d): %v",
			parentID, childID, relationTypeID, err)
	}
}

// CreateTestComment creates a test comment on a task and returns its ID
func CreateTestComment(tb testing.TB, db *sql.DB, d Dialect, taskID int, content, author string) int {
	tb.Helper()
	query := fmt.Sprintf(
		"INSERT INTO task_comments (task_id, content, author) VALUES (%s, %s, %s)",
		d.Placeholder(1), d.Placeholder(2), d.Placeholder(3))
	commentID, err := d.InsertReturningID(context.Background(), db, query, taskID, content, author)
	if err != nil {
		tb.Fatalf("Failed to create comment on task %d: %v", taskID, err)
	}
	return commentID
}

// SetColumnHoldsCompletedTasks marks a column as holding completed tasks
func SetColumnHoldsCompletedTasks(tb testing.TB, db *sql.DB, d Dialect, columnID int) {
	tb.Helper()
	query := fmt.Sprintf(
		"UPDATE columns SET holds_completed_tasks = true WHERE id = %s",
		d.Placeholder(1))
	_, err := db.ExecContext(context.Background(), query, columnID)
	if err != nil {
		tb.Fatalf("Failed to set column %d as completed: %v", columnID, err)
	}
}

// SetColumnHoldsReadyTasks marks a column as holding ready tasks
func SetColumnHoldsReadyTasks(tb testing.TB, db *sql.DB, d Dialect, columnID int) {
	tb.Helper()
	query := fmt.Sprintf(
		"UPDATE columns SET holds_ready_tasks = true WHERE id = %s",
		d.Placeholder(1))
	_, err := db.ExecContext(context.Background(), query, columnID)
	if err != nil {
		tb.Fatalf("Failed to set column %d as ready: %v", columnID, err)
	}
}

// SetColumnHoldsInProgressTasks marks a column as holding in-progress tasks
func SetColumnHoldsInProgressTasks(tb testing.TB, db *sql.DB, d Dialect, columnID int) {
	tb.Helper()
	query := fmt.Sprintf(
		"UPDATE columns SET holds_in_progress_tasks = true WHERE id = %s",
		d.Placeholder(1))
	_, err := db.ExecContext(context.Background(), query, columnID)
	if err != nil {
		tb.Fatalf("Failed to set column %d as in-progress: %v", columnID, err)
	}
}

// UpdateTaskDescription sets the description for a task
func UpdateTaskDescription(tb testing.TB, db *sql.DB, d Dialect, taskID int, description string) {
	tb.Helper()
	query := fmt.Sprintf(
		"UPDATE tasks SET description = %s WHERE id = %s",
		d.Placeholder(1), d.Placeholder(2))
	_, err := db.ExecContext(context.Background(), query, description, taskID)
	if err != nil {
		tb.Fatalf("Failed to update description for task %d: %v", taskID, err)
	}
}

// UpdateTaskFields updates multiple fields on a task at once.
// Pass a map of column names to values (e.g., {"description": "new desc", "ticket_number": 5}).
func UpdateTaskFields(tb testing.TB, db *sql.DB, d Dialect, taskID int, fields map[string]any) {
	tb.Helper()
	if len(fields) == 0 {
		return
	}

	setClauses := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+1)
	i := 1
	for col, val := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", col, d.Placeholder(i)))
		args = append(args, val)
		i++
	}
	args = append(args, taskID)

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = %s",
		joinStrings(setClauses, ", "), d.Placeholder(i))
	_, err := db.ExecContext(context.Background(), query, args...)
	if err != nil {
		tb.Fatalf("Failed to update task %d fields: %v", taskID, err)
	}
}

// LinkProjectGitBranch associates a git branch with a project
func LinkProjectGitBranch(tb testing.TB, db *sql.DB, d Dialect, projectID int, branchName string) {
	tb.Helper()
	query := fmt.Sprintf(
		"INSERT INTO project_git_branches (project_id, branch_name) VALUES (%s, %s)",
		d.Placeholder(1), d.Placeholder(2))
	_, err := db.ExecContext(context.Background(), query, projectID, branchName)
	if err != nil {
		tb.Fatalf("Failed to link git branch %q to project %d: %v", branchName, projectID, err)
	}
}

func joinStrings(strs []string, sep string) string {
	var result strings.Builder
	for i, s := range strs {
		if i > 0 {
			result.WriteString(sep)
		}
		result.WriteString(s)
	}
	return result.String()
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
