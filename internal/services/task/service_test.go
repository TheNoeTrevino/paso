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
		holds_ready_tasks BOOLEAN NOT NULL DEFAULT 0,
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

	_, err := db.ExecContext(context.Background(), schema)
	return err
}

// createTestProject creates a test project and returns its ID
func createTestProject(t *testing.T, db *sql.DB) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO projects (name, description) VALUES (?, ?)", "Test Project", "Description")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Initialize project counter
	id, _ := result.LastInsertId()
	_, err = db.ExecContext(context.Background(), "INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)", id)
	if err != nil {
		t.Fatalf("Failed to initialize project counter: %v", err)
	}

	return int(id)
}

// createTestColumn creates a test column and returns its ID
func createTestColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name, holds_ready_tasks) VALUES (?, ?, ?)", projectID, name, false)
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

// createTestReadyColumn creates a test column with holds_ready_tasks=true and returns its ID
func createTestReadyColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name, holds_ready_tasks) VALUES (?, ?, ?)", projectID, name, true)
	if err != nil {
		t.Fatalf("Failed to create test ready column: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

// createTestTask creates a test task and returns its ID
func createTestTask(t *testing.T, db *sql.DB, columnID int, title string) int {
	t.Helper()

	// Get the next position for this column
	var maxPos sql.NullInt64
	err := db.QueryRowContext(context.Background(),
		"SELECT MAX(position) FROM tasks WHERE column_id = ?", columnID).Scan(&maxPos)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Failed to get max position: %v", err)
	}

	nextPos := 0
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	result, err := db.ExecContext(context.Background(),
		"INSERT INTO tasks (column_id, title, position, type_id, priority_id) VALUES (?, ?, ?, 1, 3)",
		columnID, title, nextPos)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

// createTestLabel creates a test label and returns its ID
func createTestLabel(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO labels (project_id, name, color) VALUES (?, ?, ?)", projectID, name, "#FF5733")
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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM task_labels WHERE task_id = ?", result.ID).Scan(&count)
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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?", created.ID, labelID).Scan(&count)
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
	defer func() { _ = db.Close() }()

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
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?", created.ID, labelID).Scan(&count)
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
	defer func() { _ = db.Close() }()

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
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
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
	defer func() { _ = db.Close() }()

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
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
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
	defer func() { _ = db.Close() }()

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
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task relationships: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 relationships, got %d", count)
	}
}

func TestRemoveChildRelation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create two tasks with child relationship
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
		t.Fatalf("Failed to add child relation: %v", err)
	}

	// Remove child relation
	err = svc.RemoveChildRelation(context.Background(), task1.ID, task2.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify relationship is removed
	var count int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task relationships: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 relationships, got %d", count)
	}
}

// ============================================================================
// TEST CASES - TASK MOVEMENT OPERATIONS
// ============================================================================

func TestMoveTaskToNextColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "To Do")
	col2ID := createTestColumn(t, db, projectID, "In Progress")

	// Link columns
	_, err := db.ExecContext(context.Background(), "UPDATE columns SET next_id = ? WHERE id = ?", col2ID, col1ID)
	if err != nil {
		t.Fatalf("Failed to link columns: %v", err)
	}

	svc := NewService(db, nil)

	// Create task in first column
	task, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col1ID,
		Position: 0,
	})

	// Move to next column
	err = svc.MoveTaskToNextColumn(context.Background(), task.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify task moved to col2
	var columnID int
	err = db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task.ID).Scan(&columnID)
	if err != nil {
		t.Fatalf("Failed to query task: %v", err)
	}

	if columnID != col2ID {
		t.Errorf("Expected task in column %d, got %d", col2ID, columnID)
	}
}

func TestMoveTaskToNextColumn_LastColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Done")
	svc := NewService(db, nil)

	// Create task in last column (no next_id)
	task, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})

	// Try to move to next column (should fail)
	err := svc.MoveTaskToNextColumn(context.Background(), task.ID)

	if err == nil {
		t.Fatal("Expected error when moving from last column")
	}
}

func TestMoveTaskToNextColumn_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.MoveTaskToNextColumn(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for invalid task ID")
	}
}

func TestMoveTaskToPrevColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "To Do")
	col2ID := createTestColumn(t, db, projectID, "In Progress")

	// Link columns
	_, err := db.ExecContext(context.Background(), "UPDATE columns SET next_id = ?, prev_id = ? WHERE id = ?", col2ID, 0, col1ID)
	if err != nil {
		t.Fatalf("Failed to link columns: %v", err)
	}
	_, err = db.ExecContext(context.Background(), "UPDATE columns SET prev_id = ? WHERE id = ?", col1ID, col2ID)
	if err != nil {
		t.Fatalf("Failed to link columns: %v", err)
	}

	svc := NewService(db, nil)

	// Create task in second column
	task, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col2ID,
		Position: 0,
	})

	// Move to previous column
	err = svc.MoveTaskToPrevColumn(context.Background(), task.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify task moved to col1
	var columnID int
	err = db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task.ID).Scan(&columnID)
	if err != nil {
		t.Fatalf("Failed to query task: %v", err)
	}

	if columnID != col1ID {
		t.Errorf("Expected task in column %d, got %d", col1ID, columnID)
	}
}

func TestMoveTaskToPrevColumn_FirstColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create task in first column (no prev_id)
	task, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})

	// Try to move to previous column (should fail)
	err := svc.MoveTaskToPrevColumn(context.Background(), task.ID)

	if err == nil {
		t.Fatal("Expected error when moving from first column")
	}
}

func TestMoveTaskToPrevColumn_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.MoveTaskToPrevColumn(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for invalid task ID")
	}
}

func TestMoveTaskToColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "To Do")
	col2ID := createTestColumn(t, db, projectID, "Done")
	svc := NewService(db, nil)

	// Create task in first column
	task, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col1ID,
		Position: 0,
	})

	// Move to specific column
	err := svc.MoveTaskToColumn(context.Background(), task.ID, col2ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify task moved
	var columnID int
	err = db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task.ID).Scan(&columnID)
	if err != nil {
		t.Fatalf("Failed to query task: %v", err)
	}

	if columnID != col2ID {
		t.Errorf("Expected task in column %d, got %d", col2ID, columnID)
	}
}

func TestMoveTaskToColumn_InvalidColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create task
	task, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})

	// Try to move to invalid column
	err := svc.MoveTaskToColumn(context.Background(), task.ID, 999)

	if err == nil {
		t.Fatal("Expected error for invalid column ID")
	}
}

func TestMoveTaskToColumn_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Try to move invalid task
	err := svc.MoveTaskToColumn(context.Background(), 999, columnID)

	if err == nil {
		t.Fatal("Expected error for invalid task ID")
	}
}

func TestMoveTaskUp(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create two tasks
	task1, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})

	task2, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})

	// Move task2 up (should swap positions with task1)
	err := svc.MoveTaskUp(context.Background(), task2.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify positions swapped
	var pos1, pos2 int64
	err = db.QueryRowContext(context.Background(), "SELECT position FROM tasks WHERE id = ?", task1.ID).Scan(&pos1)
	if err != nil {
		t.Fatalf("Failed to query task1 position: %v", err)
	}

	err = db.QueryRowContext(context.Background(), "SELECT position FROM tasks WHERE id = ?", task2.ID).Scan(&pos2)
	if err != nil {
		t.Fatalf("Failed to query task2 position: %v", err)
	}

	if pos2 >= pos1 {
		t.Errorf("Expected task2 position (%d) to be less than task1 position (%d)", pos2, pos1)
	}
}

func TestMoveTaskUp_FirstPosition(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create task at first position
	task, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})

	// Try to move up (should fail - no task above)
	err := svc.MoveTaskUp(context.Background(), task.ID)

	if err == nil {
		t.Fatal("Expected error when moving up from first position")
	}
}

func TestMoveTaskUp_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.MoveTaskUp(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for invalid task ID")
	}
}

func TestMoveTaskDown(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create two tasks
	task1, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})

	task2, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})

	// Move task1 down (should swap positions with task2)
	err := svc.MoveTaskDown(context.Background(), task1.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify positions swapped
	var pos1, pos2 int64
	err = db.QueryRowContext(context.Background(), "SELECT position FROM tasks WHERE id = ?", task1.ID).Scan(&pos1)
	if err != nil {
		t.Fatalf("Failed to query task1 position: %v", err)
	}

	err = db.QueryRowContext(context.Background(), "SELECT position FROM tasks WHERE id = ?", task2.ID).Scan(&pos2)
	if err != nil {
		t.Fatalf("Failed to query task2 position: %v", err)
	}

	if pos1 <= pos2 {
		t.Errorf("Expected task1 position (%d) to be greater than task2 position (%d)", pos1, pos2)
	}
}

func TestMoveTaskDown_LastPosition(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create task at last position
	task, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})

	// Try to move down (should fail - no task below)
	err := svc.MoveTaskDown(context.Background(), task.ID)

	if err == nil {
		t.Fatal("Expected error when moving down from last position")
	}
}

func TestMoveTaskDown_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.MoveTaskDown(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for invalid task ID")
	}
}

// ============================================================================
// TEST CASES - TASK FILTERING AND REFERENCES
// ============================================================================

func TestGetTaskSummariesByProjectFiltered(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create tasks with different titles
	_, _ = svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Fix bug in login",
		ColumnID: columnID,
		Position: 0,
	})

	_, _ = svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Add new feature",
		ColumnID: columnID,
		Position: 1,
	})

	_, _ = svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Fix bug in signup",
		ColumnID: columnID,
		Position: 2,
	})

	// Filter by "bug"
	results, err := svc.GetTaskSummariesByProjectFiltered(context.Background(), projectID, "bug")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return 2 tasks with "bug" in title
	totalTasks := 0
	for _, tasks := range results {
		totalTasks += len(tasks)
	}

	if totalTasks != 2 {
		t.Errorf("Expected 2 tasks with 'bug' in title, got %d", totalTasks)
	}
}

func TestGetTaskSummariesByProjectFiltered_NoResults(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create task
	_, _ = svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})

	// Filter by non-existent term
	results, err := svc.GetTaskSummariesByProjectFiltered(context.Background(), projectID, "nonexistent")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return empty map
	totalTasks := 0
	for _, tasks := range results {
		totalTasks += len(tasks)
	}

	if totalTasks != 0 {
		t.Errorf("Expected 0 tasks, got %d", totalTasks)
	}
}

func TestGetTaskSummariesByProjectFiltered_EmptyQuery(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create tasks
	_, _ = svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})

	_, _ = svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})

	// Filter with empty query (should return all)
	results, err := svc.GetTaskSummariesByProjectFiltered(context.Background(), projectID, "")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return all tasks
	totalTasks := 0
	for _, tasks := range results {
		totalTasks += len(tasks)
	}

	if totalTasks != 2 {
		t.Errorf("Expected 2 tasks, got %d", totalTasks)
	}
}

func TestGetTaskReferencesForProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create tasks
	_, _ = svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})

	_, _ = svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})

	// Get task references
	refs, err := svc.GetTaskReferencesForProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(refs) != 2 {
		t.Errorf("Expected 2 task references, got %d", len(refs))
	}
}

func TestGetTaskReferencesForProject_EmptyProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Get task references for empty project
	refs, err := svc.GetTaskReferencesForProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(refs) != 0 {
		t.Errorf("Expected 0 task references, got %d", len(refs))
	}
}

// ============================================================================
// TEST CASES - READY TASKS (holds_ready_tasks feature)
// ============================================================================

func TestGetReadyTaskSummariesByProject_OnlyReadyColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create three columns: Todo (ready), In Progress, Done
	todoCol := createTestReadyColumn(t, db, projectID, "Todo")
	inProgressCol := createTestColumn(t, db, projectID, "In Progress")
	doneCol := createTestColumn(t, db, projectID, "Done")

	// Create tasks in each column
	task1 := createTestTask(t, db, todoCol, "Task in Todo")
	task2 := createTestTask(t, db, inProgressCol, "Task in In Progress")
	task3 := createTestTask(t, db, doneCol, "Task in Done")

	// Get ready tasks
	readyTasks, err := svc.GetReadyTaskSummariesByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(readyTasks) != 1 {
		t.Fatalf("Expected 1 ready task, got %d", len(readyTasks))
	}

	// Verify only task from Todo column is returned
	if readyTasks[0].ID != task1 {
		t.Errorf("Expected task ID %d, got %d", task1, readyTasks[0].ID)
	}

	// Verify tasks from other columns are not included
	for _, task := range readyTasks {
		if task.ID == task2 || task.ID == task3 {
			t.Errorf("Unexpected task ID %d in ready tasks (should only be from Todo column)", task.ID)
		}
	}
}

func TestGetReadyTaskSummariesByProject_ExcludesBlockedTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create ready column
	readyCol := createTestReadyColumn(t, db, projectID, "Todo")

	// Create two tasks
	task1 := createTestTask(t, db, readyCol, "Unblocked Task")
	task2 := createTestTask(t, db, readyCol, "Blocked Task")

	// Create a blocker relationship (task2 is blocked by task1)
	// relation_type_id = 2 is "Blocked By/Blocker" with is_blocking = 1
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, ?)",
		task2, task1, 2)
	if err != nil {
		t.Fatalf("Failed to create blocking relationship: %v", err)
	}

	// Get ready tasks
	readyTasks, err := svc.GetReadyTaskSummariesByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should only get task1 (task2 is blocked)
	if len(readyTasks) != 1 {
		t.Fatalf("Expected 1 unblocked task, got %d", len(readyTasks))
	}

	if readyTasks[0].ID != task1 {
		t.Errorf("Expected unblocked task ID %d, got %d", task1, readyTasks[0].ID)
	}
}

func TestGetReadyTaskSummariesByProject_EmptyWhenNoReadyColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create columns but none are ready
	col1 := createTestColumn(t, db, projectID, "Todo")
	createTestColumn(t, db, projectID, "Done")

	// Create tasks
	createTestTask(t, db, col1, "Task 1")

	// Get ready tasks
	readyTasks, err := svc.GetReadyTaskSummariesByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(readyTasks) != 0 {
		t.Errorf("Expected 0 ready tasks when no column is marked as ready, got %d", len(readyTasks))
	}
}

func TestGetReadyTaskSummariesByProject_EmptyReadyColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create ready column with no tasks
	createTestReadyColumn(t, db, projectID, "Todo")

	// Get ready tasks
	readyTasks, err := svc.GetReadyTaskSummariesByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(readyTasks) != 0 {
		t.Errorf("Expected 0 ready tasks for empty ready column, got %d", len(readyTasks))
	}
}

func TestGetReadyTaskSummariesByProject_IncludesTaskDetails(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create ready column
	readyCol := createTestReadyColumn(t, db, projectID, "Todo")

	// Create label
	labelID := createTestLabel(t, db, projectID, "bug")

	// Create task with label and high priority (priority_id = 4)
	taskID := createTestTask(t, db, readyCol, "Important Bug Fix")
	_, err := db.ExecContext(context.Background(),
		"UPDATE tasks SET priority_id = 4 WHERE id = ?", taskID)
	if err != nil {
		t.Fatalf("Failed to update task priority: %v", err)
	}

	// Attach label
	_, err = db.ExecContext(context.Background(),
		"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
	if err != nil {
		t.Fatalf("Failed to attach label: %v", err)
	}

	// Get ready tasks
	readyTasks, err := svc.GetReadyTaskSummariesByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(readyTasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(readyTasks))
	}

	task := readyTasks[0]

	// Verify task details
	if task.Title != "Important Bug Fix" {
		t.Errorf("Expected title 'Important Bug Fix', got '%s'", task.Title)
	}

	if task.PriorityDescription != "high" {
		t.Errorf("Expected priority 'high', got '%s'", task.PriorityDescription)
	}

	if len(task.Labels) != 1 {
		t.Fatalf("Expected 1 label, got %d", len(task.Labels))
	}

	if task.Labels[0].Name != "bug" {
		t.Errorf("Expected label name 'bug', got '%s'", task.Labels[0].Name)
	}
}
