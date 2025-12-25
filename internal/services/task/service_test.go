package task

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"
)

// ============================================================================
// TEST HELPERS
// ============================================================================

// setupTestDB creates an in-memory database and runs migrations
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

	// Run migrations inline
	if err := createTestSchema(db); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// createTestSchema creates the minimal schema needed for task service tests
func createTestSchema(db *sql.DB) error {
	schema := `
	-- Projects table
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
		(2, 'Blocks', 'Blocked By', '#EF4444', 1);

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
		FOREIGN KEY (priority_id) REFERENCES priorities(id)
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
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	-- Task-labels join table
	CREATE TABLE IF NOT EXISTS task_labels (
		task_id INTEGER NOT NULL,
		label_id INTEGER NOT NULL,
		PRIMARY KEY (task_id, label_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
	);
	`

	_, err := db.Exec(schema)
	return err
}

// createTestProject creates a test project and returns its ID
func createTestProject(t *testing.T, db *sql.DB) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO projects (name, description) VALUES (?, ?)", "Test Project", "Description")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Initialize project counter
	id, _ := result.LastInsertId()
	_, err = db.Exec("INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)", id)
	if err != nil {
		t.Fatalf("Failed to initialize project counter: %v", err)
	}

	return int(id)
}

// createTestColumn creates a test column and returns its ID
func createTestColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, name)
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

// createTestLabel creates a test label and returns its ID
func createTestLabel(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO labels (project_id, name, color) VALUES (?, ?, ?)", projectID, name, "#FF5733")
	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

// ============================================================================
// TEST CASES - CREATE
// ============================================================================

func TestCreateTask(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	req := CreateTaskRequest{
		Title:       "Fix bug in login",
		Description: "Users can't log in",
		ColumnID:    columnID,
		Position:    0,
		PriorityID:  4, // high
		TypeID:      3, // bug
	}

	result, err := svc.CreateTask(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected task result, got nil")
	}

	if result.Title != "Fix bug in login" {
		t.Errorf("Expected title 'Fix bug in login', got '%s'", result.Title)
	}

	if result.Description != "Users can't log in" {
		t.Errorf("Expected description 'Users can't log in', got '%s'", result.Description)
	}

	if result.ColumnID != columnID {
		t.Errorf("Expected column ID %d, got %d", columnID, result.ColumnID)
	}

	if result.ID == 0 {
		t.Error("Expected task ID to be set")
	}

	// Note: models.Task doesn't include TicketNumber (only TaskDetail does)
	// We could verify it via GetTaskDetail if needed, but basic task creation is sufficient here
}

func TestCreateTask_EmptyTitle(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	req := CreateTaskRequest{
		Title:    "", // Empty title
		ColumnID: columnID,
		Position: 0,
	}

	_, err := svc.CreateTask(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for empty title")
	}

	if err != ErrEmptyTitle {
		t.Errorf("Expected ErrEmptyTitle, got %v", err)
	}
}

func TestCreateTask_TitleTooLong(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	longTitle := ""
	for i := 0; i < 256; i++ {
		longTitle += "a"
	}

	req := CreateTaskRequest{
		Title:    longTitle,
		ColumnID: columnID,
		Position: 0,
	}

	_, err := svc.CreateTask(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for long title")
	}

	if err != ErrTitleTooLong {
		t.Errorf("Expected ErrTitleTooLong, got %v", err)
	}
}

func TestCreateTask_InvalidColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	req := CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: 0, // Invalid
		Position: 0,
	}

	_, err := svc.CreateTask(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for invalid column ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestCreateTask_InvalidPosition(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	req := CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: -1, // Invalid
	}

	_, err := svc.CreateTask(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for invalid position")
	}

	if err != ErrInvalidPosition {
		t.Errorf("Expected ErrInvalidPosition, got %v", err)
	}
}

func TestCreateTask_WithLabels(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	label1ID := createTestLabel(t, db, projectID, "Bug")
	label2ID := createTestLabel(t, db, projectID, "Critical")
	svc := NewService(db, nil)

	req := CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
		LabelIDs: []int{label1ID, label2ID},
	}

	result, err := svc.CreateTask(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify labels are attached
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM task_labels WHERE task_id = ?", result.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task labels: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 labels attached, got %d", count)
	}
}

// ============================================================================
// TEST CASES - READ
// ============================================================================

func TestGetTaskDetail(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create a task
	created, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		ColumnID:    columnID,
		Position:    0,
		PriorityID:  4,
		TypeID:      3,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Get task detail
	result, err := svc.GetTaskDetail(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, result.ID)
	}

	if result.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", result.Title)
	}

	if result.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", result.Description)
	}
}

func TestGetTaskDetail_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	_, err := svc.GetTaskDetail(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}

	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetTaskDetail_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	_, err := svc.GetTaskDetail(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

func TestGetTaskSummariesByProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "To Do")
	col2ID := createTestColumn(t, db, projectID, "Done")
	svc := NewService(db, nil)

	// Create tasks in different columns
	_, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: col1ID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task 1: %v", err)
	}

	_, err = svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: col2ID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task 2: %v", err)
	}

	// Get summaries
	results, err := svc.GetTaskSummariesByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 columns with tasks, got %d", len(results))
	}

	if len(results[col1ID]) != 1 {
		t.Errorf("Expected 1 task in column 1, got %d", len(results[col1ID]))
	}

	if len(results[col2ID]) != 1 {
		t.Errorf("Expected 1 task in column 2, got %d", len(results[col2ID]))
	}
}

// ============================================================================
// TEST CASES - UPDATE
// ============================================================================

func TestUpdateTask(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create a task
	created, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Old Title",
		ColumnID: columnID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Update task
	newTitle := "New Title"
	newDesc := "New Description"
	req := UpdateTaskRequest{
		TaskID:      created.ID,
		Title:       &newTitle,
		Description: &newDesc,
	}

	err = svc.UpdateTask(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify update
	updated, err := svc.GetTaskDetail(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}

	if updated.Title != "New Title" {
		t.Errorf("Expected title 'New Title', got '%s'", updated.Title)
	}

	if updated.Description != "New Description" {
		t.Errorf("Expected description 'New Description', got '%s'", updated.Description)
	}
}

func TestUpdateTask_EmptyTitle(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create a task
	created, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Old Title",
		ColumnID: columnID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Try to update with empty title
	emptyTitle := ""
	req := UpdateTaskRequest{
		TaskID: created.ID,
		Title:  &emptyTitle,
	}

	err = svc.UpdateTask(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for empty title")
	}

	if err != ErrEmptyTitle {
		t.Errorf("Expected ErrEmptyTitle, got %v", err)
	}
}

func TestUpdateTask_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	newTitle := "New Title"
	req := UpdateTaskRequest{
		TaskID: 0,
		Title:  &newTitle,
	}

	err := svc.UpdateTask(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

// ============================================================================
// TEST CASES - DELETE
// ============================================================================

func TestDeleteTask(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create a task
	created, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Delete task
	err = svc.DeleteTask(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify task is deleted
	_, err = svc.GetTaskDetail(context.Background(), created.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Expected sql.ErrNoRows after deletion, got %v", err)
	}
}

func TestDeleteTask_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	err := svc.DeleteTask(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

// ============================================================================
// TEST CASES - LABELS
// ============================================================================

func TestAttachLabel(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	labelID := createTestLabel(t, db, projectID, "Bug")
	svc := NewService(db, nil)

	// Create a task
	created, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Attach label
	err = svc.AttachLabel(context.Background(), created.ID, labelID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify label is attached
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?", created.ID, labelID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task labels: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 label attached, got %d", count)
	}
}

func TestDetachLabel(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	labelID := createTestLabel(t, db, projectID, "Bug")
	svc := NewService(db, nil)

	// Create a task with label
	created, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
		LabelIDs: []int{labelID},
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Detach label
	err = svc.DetachLabel(context.Background(), created.ID, labelID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify label is detached
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?", created.ID, labelID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task labels: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 labels attached, got %d", count)
	}
}

// ============================================================================
// TEST CASES - TASK RELATIONSHIPS
// ============================================================================

func TestAddParentRelation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create two tasks
	task1, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})

	task2, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: columnID,
		Position: 1,
	})

	// Add parent relation (task1 is parent of task2)
	err := svc.AddParentRelation(context.Background(), task2.ID, task1.ID, 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify relationship exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task relationships: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 relationship, got %d", count)
	}
}

func TestAddChildRelation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create two tasks
	task1, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})

	task2, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: columnID,
		Position: 1,
	})

	// Add child relation (task2 is child of task1)
	err := svc.AddChildRelation(context.Background(), task1.ID, task2.ID, 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify relationship exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task relationships: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 relationship, got %d", count)
	}
}

func TestRemoveParentRelation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create two tasks with relationship
	task1, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})

	task2, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:     "Child Task",
		ColumnID:  columnID,
		Position:  1,
		ParentIDs: []int{task1.ID},
	})

	// Remove parent relation
	err := svc.RemoveParentRelation(context.Background(), task2.ID, task1.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify relationship is removed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task relationships: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 relationships, got %d", count)
	}
}
