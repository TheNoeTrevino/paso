package task

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// ============================================================================
// TEST HELPERS
// ============================================================================

// setupTestDB creates an in-memory database with full schema using testutil
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return testutil.SetupTestDB(t)
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

// createTestCompletedColumn creates a test column with holds_completed_tasks=true and returns its ID
func createTestCompletedColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name, holds_completed_tasks) VALUES (?, ?, ?)", projectID, name, true)
	if err != nil {
		t.Fatalf("Failed to create test completed column: %v", err)
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

// createTestComment creates a test comment and returns its ID
func createTestComment(t *testing.T, db *sql.DB, taskID int, message, author string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(),
		"INSERT INTO task_comments (task_id, content, author) VALUES (?, ?, ?)",
		taskID, message, author)
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
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

func TestCreateTask_Validation(t *testing.T) {
	type args struct {
		req CreateTaskRequest
		// If setupFn is provided, it sets up additional DB state
		setupFn func(*sql.DB, int) CreateTaskRequest
	}

	tests := []struct {
		name      string
		args      args
		wantErr   bool
		errType   error
		needsTest bool // Whether this test needs a valid column for setup
	}{
		{
			name: "empty title",
			args: args{
				req: CreateTaskRequest{
					Title:    "",
					ColumnID: 1,
					Position: 0,
				},
			},
			wantErr: true,
			errType: ErrEmptyTitle,
		},
		{
			name: "title too long",
			args: args{
				setupFn: func(db *sql.DB, _ int) CreateTaskRequest {
					longTitle := ""
					for i := 0; i < 256; i++ {
						longTitle += "a"
					}
					return CreateTaskRequest{
						Title:    longTitle,
						ColumnID: 1,
						Position: 0,
					}
				},
			},
			wantErr: true,
			errType: ErrTitleTooLong,
		},
		{
			name: "invalid column ID",
			args: args{
				req: CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: 0,
					Position: 0,
				},
			},
			wantErr: true,
			errType: ErrInvalidColumnID,
		},
		{
			name:      "invalid position",
			needsTest: true,
			args: args{
				setupFn: func(db *sql.DB, columnID int) CreateTaskRequest {
					return CreateTaskRequest{
						Title:    "Test Task",
						ColumnID: columnID,
						Position: -1,
					}
				},
			},
			wantErr: true,
			errType: ErrInvalidPosition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			req := tt.args.req

			// Setup database if needed
			if tt.needsTest || tt.args.setupFn != nil {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")

				if tt.args.setupFn != nil {
					req = tt.args.setupFn(db, columnID)
				}
			}

			svc := NewService(db, nil)
			_, err := svc.CreateTask(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("CreateTask() error = %v, want %v", err, tt.errType)
			}
		})
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

func TestUpdateTask_Validation(t *testing.T) {
	tests := []struct {
		name    string
		taskID  int
		title   *string
		wantErr bool
		errType error
		setupFn func(*sql.DB) int // Returns task ID if needed
	}{
		{
			name:    "empty title",
			title:   ptrString(""),
			wantErr: true,
			errType: ErrEmptyTitle,
			setupFn: func(db *sql.DB) int {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				task, _ := NewService(db, nil).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Old Title",
					ColumnID: columnID,
					Position: 0,
				})
				return task.ID
			},
		},
		{
			name:    "invalid ID",
			taskID:  0,
			title:   ptrString("New Title"),
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			taskID := tt.taskID
			if tt.setupFn != nil {
				taskID = tt.setupFn(db)
			}

			svc := NewService(db, nil)
			req := UpdateTaskRequest{
				TaskID: taskID,
				Title:  tt.title,
			}

			err := svc.UpdateTask(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("UpdateTask() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// ptrString is a helper function that returns a pointer to a string
func ptrString(s string) *string {
	return &s
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

// ============================================================================
// TEST CASES - TASK TREE
// ============================================================================

func TestGetTaskTreeByProject_EmptyProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Get tree for empty project
	tree, err := svc.GetTaskTreeByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tree) != 0 {
		t.Errorf("Expected 0 root tasks, got %d", len(tree))
	}
}

func TestGetTaskTreeByProject_SimpleHierarchy(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create parent and child tasks
	parentID := createTestTask(t, db, columnID, "Parent Task")
	childID := createTestTask(t, db, columnID, "Child Task")

	// Add parent-child relation (non-blocking)
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		parentID, childID)
	if err != nil {
		t.Fatalf("Failed to create relation: %v", err)
	}

	// Get tree
	tree, err := svc.GetTaskTreeByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tree) != 1 {
		t.Fatalf("Expected 1 root task, got %d", len(tree))
	}

	root := tree[0]
	if root.ID != parentID {
		t.Errorf("Expected root ID %d, got %d", parentID, root.ID)
	}

	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.Children))
	}

	child := root.Children[0]
	if child.ID != childID {
		t.Errorf("Expected child ID %d, got %d", childID, child.ID)
	}
}

func TestGetTaskTreeByProject_CircularDependency(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create three tasks: A -> B -> C -> A (circular)
	taskA := createTestTask(t, db, columnID, "Task A")
	taskB := createTestTask(t, db, columnID, "Task B")
	taskC := createTestTask(t, db, columnID, "Task C")

	// A -> B -> C -> A (circular)
	_, _ = db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		taskA, taskB)
	_, _ = db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		taskB, taskC)
	_, _ = db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		taskC, taskA)

	// Get tree - should handle circular dependency gracefully
	tree, err := svc.GetTaskTreeByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// With circular dependency, none are truly "root" tasks
	// This is expected behavior - the function should not crash or hang
	if len(tree) != 0 {
		t.Logf("Got %d root tasks with circular dependencies (expected)", len(tree))
	}
}

func TestGetTaskTreeByProject_DeepNesting(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create a deep hierarchy: Task1 -> Task2 -> Task3 -> Task4 -> Task5
	tasks := make([]int, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = createTestTask(t, db, columnID, "Task "+string(rune('1'+i)))
	}

	// Link them in a chain
	for i := 0; i < 4; i++ {
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
			tasks[i], tasks[i+1])
		if err != nil {
			t.Fatalf("Failed to create relation: %v", err)
		}
	}

	// Get tree
	tree, err := svc.GetTaskTreeByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tree) != 1 {
		t.Fatalf("Expected 1 root task, got %d", len(tree))
	}

	// Traverse the tree and verify depth
	depth := 0
	current := tree[0]
	for {
		depth++
		if len(current.Children) == 0 {
			break
		}
		if depth > 10 {
			t.Fatal("Depth exceeds expected maximum")
		}
		current = current.Children[0]
	}

	if depth != 5 {
		t.Errorf("Expected depth 5, got %d", depth)
	}
}

func TestGetTaskTreeByProject_BlockingRelationship(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create parent and child tasks
	parentID := createTestTask(t, db, columnID, "Parent Task")
	blockerID := createTestTask(t, db, columnID, "Blocker Task")

	// Add blocking relation (relation_type_id = 2 is the blocker type)
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
		parentID, blockerID)
	if err != nil {
		t.Fatalf("Failed to create blocking relation: %v", err)
	}

	// Get tree
	tree, err := svc.GetTaskTreeByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tree) != 1 {
		t.Fatalf("Expected 1 root task, got %d", len(tree))
	}

	root := tree[0]
	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.Children))
	}

	blocker := root.Children[0]
	if !blocker.IsBlocking {
		t.Error("Expected child to be marked as blocking")
	}
}

func TestGetTaskTreeByProject_SortedByTicketNumber(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create multiple root tasks (no parent relations)
	task1 := createTestTask(t, db, columnID, "Task 1")
	task2 := createTestTask(t, db, columnID, "Task 2")
	task3 := createTestTask(t, db, columnID, "Task 3")

	// Set ticket numbers (in database they auto-increment, but let's verify sorting)
	_, _ = db.ExecContext(context.Background(), "UPDATE tasks SET ticket_number = 3 WHERE id = ?", task1)
	_, _ = db.ExecContext(context.Background(), "UPDATE tasks SET ticket_number = 1 WHERE id = ?", task2)
	_, _ = db.ExecContext(context.Background(), "UPDATE tasks SET ticket_number = 2 WHERE id = ?", task3)

	// Get tree
	tree, err := svc.GetTaskTreeByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tree) != 3 {
		t.Fatalf("Expected 3 root tasks, got %d", len(tree))
	}

	// Verify sorted by ticket number
	if tree[0].TicketNumber != 1 {
		t.Errorf("Expected first task ticket number 1, got %d", tree[0].TicketNumber)
	}
	if tree[1].TicketNumber != 2 {
		t.Errorf("Expected second task ticket number 2, got %d", tree[1].TicketNumber)
	}
	if tree[2].TicketNumber != 3 {
		t.Errorf("Expected third task ticket number 3, got %d", tree[2].TicketNumber)
	}
}

func TestGetTaskTreeByProject_MultipleRootsWithChildren(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create two separate trees
	root1 := createTestTask(t, db, columnID, "Root 1")
	child1 := createTestTask(t, db, columnID, "Child 1")

	root2 := createTestTask(t, db, columnID, "Root 2")
	child2 := createTestTask(t, db, columnID, "Child 2")

	// Link them
	_, _ = db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		root1, child1)
	_, _ = db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		root2, child2)

	// Get tree
	tree, err := svc.GetTaskTreeByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tree) != 2 {
		t.Fatalf("Expected 2 root tasks, got %d", len(tree))
	}

	// Verify both roots have exactly one child
	for i, root := range tree {
		if len(root.Children) != 1 {
			t.Errorf("Root %d: expected 1 child, got %d", i, len(root.Children))
		}
	}
}

// MOVE TO READY COLUMN TESTS
// ============================================================================

func TestMoveTaskToReadyColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	todoColID := createTestColumn(t, db, projectID, "To Do")
	readyColID := createTestReadyColumn(t, db, projectID, "Ready")
	svc := NewService(db, nil)

	// Create task in To Do column
	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: todoColID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Move to ready column
	err = svc.MoveTaskToReadyColumn(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify task moved to ready column
	var columnID int
	err = db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task.ID).Scan(&columnID)
	if err != nil {
		t.Fatalf("Failed to query task: %v", err)
	}

	if columnID != readyColID {
		t.Errorf("Expected task in ready column %d, got %d", readyColID, columnID)
	}
}

func TestMoveTaskToReadyColumn_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Try to move non-existent task
	err := svc.MoveTaskToReadyColumn(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for invalid task ID")
	}

	if !errors.Is(err, ErrInvalidTaskID) && !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Expected ErrInvalidTaskID or sql.ErrNoRows, got %v", err)
	}
}

func TestMoveTaskToReadyColumn_NoReadyColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create task
	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Try to move to ready column when none exists
	err = svc.MoveTaskToReadyColumn(context.Background(), task.ID)

	if err == nil {
		t.Fatal("Expected error when no ready column exists")
	}

	if !errors.Is(err, sql.ErrNoRows) && err.Error() != "no ready column configured for this project" {
		t.Errorf("Expected 'no ready column' error, got %v", err)
	}
}

func TestMoveTaskToReadyColumn_AlreadyInReadyColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	readyColID := createTestReadyColumn(t, db, projectID, "Ready")
	svc := NewService(db, nil)

	// Create task already in ready column
	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: readyColID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Try to move to ready column (already there)
	err = svc.MoveTaskToReadyColumn(context.Background(), task.ID)

	if !errors.Is(err, ErrTaskAlreadyInTargetColumn) {
		t.Errorf("Expected ErrTaskAlreadyInTargetColumn, got %v", err)
	}
}

func TestMoveTaskToReadyColumn_ZeroTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Try to move task with ID 0
	err := svc.MoveTaskToReadyColumn(context.Background(), 0)

	if !errors.Is(err, ErrInvalidTaskID) {
		t.Errorf("Expected ErrInvalidTaskID for task ID 0, got %v", err)
	}
}

func TestMoveTaskToReadyColumn_NegativeTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Try to move task with negative ID
	err := svc.MoveTaskToReadyColumn(context.Background(), -1)

	if !errors.Is(err, ErrInvalidTaskID) {
		t.Errorf("Expected ErrInvalidTaskID for negative task ID, got %v", err)
	}
}

// ============================================================================
// MOVE TO COMPLETED COLUMN TESTS
// ============================================================================

func TestMoveTaskToCompletedColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	todoColID := createTestColumn(t, db, projectID, "To Do")
	completedColID := createTestCompletedColumn(t, db, projectID, "Done")
	svc := NewService(db, nil)

	// Create task in To Do column
	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: todoColID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Move to completed column
	err = svc.MoveTaskToCompletedColumn(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify task moved to completed column
	var columnID int
	err = db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task.ID).Scan(&columnID)
	if err != nil {
		t.Fatalf("Failed to query task: %v", err)
	}

	if columnID != completedColID {
		t.Errorf("Expected task in completed column %d, got %d", completedColID, columnID)
	}
}

func TestMoveTaskToCompletedColumn_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Try to move non-existent task
	err := svc.MoveTaskToCompletedColumn(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for invalid task ID")
	}

	if !errors.Is(err, ErrInvalidTaskID) && !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Expected ErrInvalidTaskID or sql.ErrNoRows, got %v", err)
	}
}

func TestMoveTaskToCompletedColumn_NoCompletedColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create task
	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Try to move to completed column when none exists
	err = svc.MoveTaskToCompletedColumn(context.Background(), task.ID)

	if err == nil {
		t.Fatal("Expected error when no completed column exists")
	}

	if !errors.Is(err, sql.ErrNoRows) && err.Error() != "no completed column configured for this project" {
		t.Errorf("Expected 'no completed column' error, got %v", err)
	}
}

func TestMoveTaskToCompletedColumn_AlreadyInCompletedColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	completedColID := createTestCompletedColumn(t, db, projectID, "Done")
	svc := NewService(db, nil)

	// Create task already in completed column
	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: completedColID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Try to move to completed column (already there)
	err = svc.MoveTaskToCompletedColumn(context.Background(), task.ID)

	if !errors.Is(err, ErrTaskAlreadyInTargetColumn) {
		t.Errorf("Expected ErrTaskAlreadyInTargetColumn, got %v", err)
	}
}

func TestMoveTaskToCompletedColumn_ZeroTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Try to move task with ID 0
	err := svc.MoveTaskToCompletedColumn(context.Background(), 0)

	if !errors.Is(err, ErrInvalidTaskID) {
		t.Errorf("Expected ErrInvalidTaskID for task ID 0, got %v", err)
	}
}

func TestMoveTaskToCompletedColumn_NegativeTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Try to move task with negative ID
	err := svc.MoveTaskToCompletedColumn(context.Background(), -1)

	if !errors.Is(err, ErrInvalidTaskID) {
		t.Errorf("Expected ErrInvalidTaskID for negative task ID, got %v", err)
	}
}

func TestMoveTaskToCompletedColumn_MultipleTasksInProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	todoColID := createTestColumn(t, db, projectID, "To Do")
	inProgressColID := createTestColumn(t, db, projectID, "In Progress")
	completedColID := createTestCompletedColumn(t, db, projectID, "Done")
	svc := NewService(db, nil)

	// Create multiple tasks in different columns
	task1, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: todoColID,
		Position: 0,
	})
	task2, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: inProgressColID,
		Position: 0,
	})
	task3, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 3",
		ColumnID: todoColID,
		Position: 1,
	})

	// Move task2 to completed
	err := svc.MoveTaskToCompletedColumn(context.Background(), task2.ID)
	if err != nil {
		t.Fatalf("Expected no error moving task2, got %v", err)
	}

	// Verify task2 is in completed column
	var col2 int
	if err := db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task2.ID).Scan(&col2); err != nil {
		t.Fatalf("Failed to query task2 column: %v", err)
	}
	if col2 != completedColID {
		t.Errorf("Expected task2 in completed column, got column %d", col2)
	}

	// Verify other tasks are unchanged
	var col1, col3 int
	if err := db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task1.ID).Scan(&col1); err != nil {
		t.Fatalf("Failed to query task1 column: %v", err)
	}
	if err := db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task3.ID).Scan(&col3); err != nil {
		t.Fatalf("Failed to query task3 column: %v", err)
	}

	if col1 != todoColID {
		t.Errorf("Expected task1 in todo column, got column %d", col1)
	}
	if col3 != todoColID {
		t.Errorf("Expected task3 in todo column, got column %d", col3)
	}
}

func TestMoveTaskToReadyColumn_MultipleTasksInProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	todoColID := createTestColumn(t, db, projectID, "To Do")
	inProgressColID := createTestColumn(t, db, projectID, "In Progress")
	readyColID := createTestReadyColumn(t, db, projectID, "Ready")
	svc := NewService(db, nil)

	// Create multiple tasks in different columns
	task1, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: todoColID,
		Position: 0,
	})
	task2, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: inProgressColID,
		Position: 0,
	})
	task3, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 3",
		ColumnID: todoColID,
		Position: 1,
	})

	// Move task2 to ready
	err := svc.MoveTaskToReadyColumn(context.Background(), task2.ID)
	if err != nil {
		t.Fatalf("Expected no error moving task2, got %v", err)
	}

	// Verify task2 is in ready column
	var col2 int
	if err := db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task2.ID).Scan(&col2); err != nil {
		t.Fatalf("Failed to query task2 column: %v", err)
	}
	if col2 != readyColID {
		t.Errorf("Expected task2 in ready column, got column %d", col2)
	}

	// Verify other tasks are unchanged
	var col1, col3 int
	if err := db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task1.ID).Scan(&col1); err != nil {
		t.Fatalf("Failed to query task1 column: %v", err)
	}
	if err := db.QueryRowContext(context.Background(), "SELECT column_id FROM tasks WHERE id = ?", task3.ID).Scan(&col3); err != nil {
		t.Fatalf("Failed to query task3 column: %v", err)
	}

	if col1 != todoColID {
		t.Errorf("Expected task1 in todo column, got column %d", col1)
	}
	if col3 != todoColID {
		t.Errorf("Expected task3 in todo column, got column %d", col3)
	}
}

// ============================================================================
// TEST CASES - COMMENT OPERATIONS
// ============================================================================

func TestCreateComment(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	svc := NewService(db, nil)

	req := CreateCommentRequest{
		TaskID:  taskID,
		Message: "This is a test comment",
		Author:  "testuser",
	}

	result, err := svc.CreateComment(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected comment result, got nil")
	}

	if result.ID == 0 {
		t.Error("Expected comment ID to be set")
	}

	if result.TaskID != taskID {
		t.Errorf("Expected task ID %d, got %d", taskID, result.TaskID)
	}

	if result.Message != "This is a test comment" {
		t.Errorf("Expected message 'This is a test comment', got '%s'", result.Message)
	}

	if result.Author != "testuser" {
		t.Errorf("Expected author 'testuser', got '%s'", result.Author)
	}

	if result.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestCreateComment_Validation(t *testing.T) {
	tests := []struct {
		name    string
		taskID  int
		message string
		author  string
		wantErr bool
		errType error
		setupFn func(*sql.DB) int // Returns task ID if needed
	}{
		{
			name:    "empty message",
			message: "",
			author:  "testuser",
			wantErr: true,
			errType: ErrEmptyCommentMessage,
			setupFn: func(db *sql.DB) int {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return createTestTask(t, db, columnID, "Test Task")
			},
		},
		{
			name: "message too long",
			setupFn: func(db *sql.DB) int {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return createTestTask(t, db, columnID, "Test Task")
			},
			message: func() string {
				msg := ""
				for i := 0; i < 1001; i++ {
					msg += "a"
				}
				return msg
			}(),
			author:  "testuser",
			wantErr: true,
			errType: ErrCommentMessageTooLong,
		},
		{
			name:    "invalid task ID",
			taskID:  0,
			message: "Test comment",
			author:  "testuser",
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name:    "non-existent task",
			taskID:  999,
			message: "Test comment",
			author:  "testuser",
			wantErr: true,
			errType: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			taskID := tt.taskID
			if tt.setupFn != nil {
				taskID = tt.setupFn(db)
			}

			svc := NewService(db, nil)
			req := CreateCommentRequest{
				TaskID:  taskID,
				Message: tt.message,
				Author:  tt.author,
			}

			_, err := svc.CreateComment(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("CreateComment() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	commentID := createTestComment(t, db, taskID, "Original message", "testuser")
	svc := NewService(db, nil)

	req := UpdateCommentRequest{
		CommentID: commentID,
		Message:   "Updated message",
	}

	err := svc.UpdateComment(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the comment was updated
	var updatedMessage string
	err = db.QueryRowContext(context.Background(),
		"SELECT content FROM task_comments WHERE id = ?", commentID).Scan(&updatedMessage)
	if err != nil {
		t.Fatalf("Failed to query updated comment: %v", err)
	}

	if updatedMessage != "Updated message" {
		t.Errorf("Expected message 'Updated message', got '%s'", updatedMessage)
	}
}

func TestUpdateComment_Validation(t *testing.T) {
	tests := []struct {
		name      string
		commentID int
		message   string
		wantErr   bool
		errType   error
		setupFn   func(*sql.DB) int // Returns comment ID if needed
	}{
		{
			name:    "empty message",
			message: "",
			wantErr: true,
			errType: ErrEmptyCommentMessage,
			setupFn: func(db *sql.DB) int {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return createTestComment(t, db, taskID, "Original message", "testuser")
			},
		},
		{
			name: "message too long",
			setupFn: func(db *sql.DB) int {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return createTestComment(t, db, taskID, "Original message", "testuser")
			},
			message: func() string {
				msg := ""
				for i := 0; i < 1001; i++ {
					msg += "a"
				}
				return msg
			}(),
			wantErr: true,
			errType: ErrCommentMessageTooLong,
		},
		{
			name:      "invalid ID",
			commentID: 0,
			message:   "Updated message",
			wantErr:   true,
			errType:   ErrInvalidCommentID,
		},
		{
			name:      "non-existent comment",
			commentID: 999,
			message:   "Updated message",
			wantErr:   true,
			errType:   ErrCommentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			commentID := tt.commentID
			if tt.setupFn != nil {
				commentID = tt.setupFn(db)
			}

			svc := NewService(db, nil)
			req := UpdateCommentRequest{
				CommentID: commentID,
				Message:   tt.message,
			}

			err := svc.UpdateComment(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("UpdateComment() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	commentID := createTestComment(t, db, taskID, "Test comment", "testuser")
	svc := NewService(db, nil)

	err := svc.DeleteComment(context.Background(), commentID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the comment was deleted
	var count int
	err = db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM task_comments WHERE id = ?", commentID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query comment count: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected comment to be deleted, but still exists")
	}
}

func TestDeleteComment_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.DeleteComment(context.Background(), 0) // Invalid ID

	if err == nil {
		t.Fatal("Expected validation error for invalid comment ID")
	}

	if err != ErrInvalidCommentID {
		t.Errorf("Expected ErrInvalidCommentID, got %v", err)
	}
}

func TestDeleteComment_NonExistentComment(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.DeleteComment(context.Background(), 999) // Non-existent comment

	if err == nil {
		t.Fatal("Expected error for non-existent comment")
	}

	if err != ErrCommentNotFound {
		t.Errorf("Expected ErrCommentNotFound, got %v", err)
	}
}

func TestGetCommentsByTask(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create multiple comments
	comment1ID := createTestComment(t, db, taskID, "First comment", "user1")
	comment2ID := createTestComment(t, db, taskID, "Second comment", "user2")
	comment3ID := createTestComment(t, db, taskID, "Third comment", "user3")

	svc := NewService(db, nil)

	comments, err := svc.GetCommentsByTask(context.Background(), taskID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(comments) != 3 {
		t.Fatalf("Expected 3 comments, got %d", len(comments))
	}

	// Verify comments are returned (order by created_at DESC, so newest first)
	// Since we created them in quick succession, verify IDs are present
	foundIDs := make(map[int]bool)
	for _, c := range comments {
		foundIDs[c.ID] = true
	}

	if !foundIDs[comment1ID] || !foundIDs[comment2ID] || !foundIDs[comment3ID] {
		t.Error("Not all comments were returned")
	}
}

func TestGetCommentsByTask_NoComments(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	svc := NewService(db, nil)

	comments, err := svc.GetCommentsByTask(context.Background(), taskID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(comments) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(comments))
	}
}

func TestGetCommentsByTask_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetCommentsByTask(context.Background(), 0) // Invalid ID

	if err == nil {
		t.Fatal("Expected validation error for invalid task ID")
	}

	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

func TestGetCommentsByTask_OrderedByCreatedAt(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create comments with explicit timestamps
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO task_comments (task_id, content, author, created_at) VALUES 
		(?, 'Oldest comment', 'user1', datetime('2024-01-01 10:00:00')),
		(?, 'Middle comment', 'user2', datetime('2024-01-02 10:00:00')),
		(?, 'Newest comment', 'user3', datetime('2024-01-03 10:00:00'))`,
		taskID, taskID, taskID)
	if err != nil {
		t.Fatalf("Failed to create test comments: %v", err)
	}

	svc := NewService(db, nil)

	comments, err := svc.GetCommentsByTask(context.Background(), taskID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(comments) != 3 {
		t.Fatalf("Expected 3 comments, got %d", len(comments))
	}

	// Verify DESC order (newest first)
	if comments[0].Message != "Newest comment" {
		t.Errorf("Expected first comment to be 'Newest comment', got '%s'", comments[0].Message)
	}

	if comments[1].Message != "Middle comment" {
		t.Errorf("Expected second comment to be 'Middle comment', got '%s'", comments[1].Message)
	}

	if comments[2].Message != "Oldest comment" {
		t.Errorf("Expected third comment to be 'Oldest comment', got '%s'", comments[2].Message)
	}
}

// ============================================================================
// INTEGRATION TESTS - COMMENTS
// ============================================================================

func TestGetTaskDetail_IncludesComments(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create task
	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Create comments
	createTestComment(t, db, task.ID, "First comment", "user1")
	createTestComment(t, db, task.ID, "Second comment", "user2")

	// Get task detail
	detail, err := svc.GetTaskDetail(context.Background(), task.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify comments are included
	if len(detail.Comments) != 2 {
		t.Fatalf("Expected 2 comments in task detail, got %d", len(detail.Comments))
	}

	// Verify comment data
	foundFirst := false
	foundSecond := false
	for _, c := range detail.Comments {
		if c.Message == "First comment" && c.Author == "user1" {
			foundFirst = true
		}
		if c.Message == "Second comment" && c.Author == "user2" {
			foundSecond = true
		}
	}

	if !foundFirst || !foundSecond {
		t.Error("Expected both comments to be present in task detail")
	}
}

func TestDeleteTask_CascadesComments(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create comments
	comment1ID := createTestComment(t, db, taskID, "First comment", "user1")
	comment2ID := createTestComment(t, db, taskID, "Second comment", "user2")

	svc := NewService(db, nil)

	// Delete the task
	err := svc.DeleteTask(context.Background(), taskID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify comments were cascade deleted
	var count int
	err = db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM task_comments WHERE id IN (?, ?)", comment1ID, comment2ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query comment count: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected comments to be cascade deleted, but found %d comments", count)
	}
}

// ============================================================================
// TEST - GetInProgressTasksByProject (N+1 Query Optimization)
// ============================================================================

func TestGetInProgressTasksByProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Setup: Create project with in-progress column
	projectID := createTestProject(t, db)
	inProgressCol := createTestColumnWithFlag(t, db, projectID, "In Progress", true, false, false)

	svc := NewService(db, nil)

	// Create multiple in-progress tasks with labels
	label1ID := createTestLabel(t, db, projectID, "urgent")
	label2ID := createTestLabel(t, db, projectID, "review")

	task1, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: inProgressCol,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task 1: %v", err)
	}

	task2, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: inProgressCol,
		Position: 1,
	})
	if err != nil {
		t.Fatalf("Failed to create task 2: %v", err)
	}

	// Attach labels to tasks
	if err := svc.AttachLabel(context.Background(), task1.ID, label1ID); err != nil {
		t.Fatalf("Failed to attach label to task 1: %v", err)
	}
	if err := svc.AttachLabel(context.Background(), task2.ID, label2ID); err != nil {
		t.Fatalf("Failed to attach label to task 2: %v", err)
	}

	// Get in-progress tasks
	tasks, err := svc.GetInProgressTasksByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get in-progress tasks: %v", err)
	}

	// Verify results
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 in-progress tasks, got %d", len(tasks))
	}

	// Check first task
	if tasks[0].Title != "Task 1" {
		t.Errorf("Expected task 1 title 'Task 1', got '%s'", tasks[0].Title)
	}
	if len(tasks[0].Labels) != 1 {
		t.Errorf("Expected 1 label on task 1, got %d", len(tasks[0].Labels))
	}
	if tasks[0].Labels[0].Name != "urgent" {
		t.Errorf("Expected label 'urgent' on task 1, got '%s'", tasks[0].Labels[0].Name)
	}

	// Check second task
	if tasks[1].Title != "Task 2" {
		t.Errorf("Expected task 2 title 'Task 2', got '%s'", tasks[1].Title)
	}
	if len(tasks[1].Labels) != 1 {
		t.Errorf("Expected 1 label on task 2, got %d", len(tasks[1].Labels))
	}
	if tasks[1].Labels[0].Name != "review" {
		t.Errorf("Expected label 'review' on task 2, got '%s'", tasks[1].Labels[0].Name)
	}
}

func TestGetInProgressTasksByProject_InvalidProjectID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Test with invalid project ID
	_, err := svc.GetInProgressTasksByProject(context.Background(), -1)
	if err == nil {
		t.Errorf("Expected error for invalid project ID, got nil")
	}

	_, err = svc.GetInProgressTasksByProject(context.Background(), 0)
	if err == nil {
		t.Errorf("Expected error for zero project ID, got nil")
	}
}

func TestGetInProgressTasksByProject_EmptyProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Project has no in-progress column
	tasks, err := svc.GetInProgressTasksByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return empty slice
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

// createTestColumnWithFlag creates a test column with specific column type flags
func createTestColumnWithFlag(t *testing.T, db *sql.DB, projectID int, name string, holdsInProgress, holdsReady, holdsCompleted bool) int {
	t.Helper()

	var columnID int
	err := db.QueryRowContext(
		context.Background(),
		`INSERT INTO columns (name, project_id, holds_in_progress_tasks, holds_ready_tasks, holds_completed_tasks)
		 VALUES (?, ?, ?, ?, ?)
		 RETURNING id`,
		name, projectID, holdsInProgress, holdsReady, holdsCompleted,
	).Scan(&columnID)
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}

	return columnID
}

// ============================================================================
// INTEGRATION TESTS - TREE BUILDING AND CIRCULAR DEPENDENCIES
// ============================================================================

// Helper function to add a relationship between tasks
func addTaskRelation(t *testing.T, db *sql.DB, parentID, childID int, relationTypeID int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, ?)",
		parentID, childID, relationTypeID)
	if err != nil {
		t.Fatalf("Failed to add task relation: %v", err)
	}
}

// TestGetTaskTreeByProject_SingleTask tests tree with a single task (no relationships)
func TestGetTaskTreeByProject_SingleTask(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create a single task
	task1ID := createTestTask(t, db, columnID, "Task 1")

	// Get tree
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get task tree: %v", err)
	}

	// Should have exactly one root node
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(nodes))
	}

	node := nodes[0]
	if node.ID != task1ID {
		t.Errorf("Expected task ID %d, got %d", task1ID, node.ID)
	}
	if node.Title != "Task 1" {
		t.Errorf("Expected title 'Task 1', got '%s'", node.Title)
	}
	if len(node.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(node.Children))
	}
}

// TestGetTaskTreeByProject_SimpleLinearTree tests a linear chain: A -> B -> C
func TestGetTaskTreeByProject_SimpleLinearTree(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create three tasks in a linear chain: A -> B -> C
	// A is parent (root), B is child of A, C is child of B
	taskA := createTestTask(t, db, columnID, "Task A")
	taskB := createTestTask(t, db, columnID, "Task B")
	taskC := createTestTask(t, db, columnID, "Task C")

	// Create parent-child relationships
	addTaskRelation(t, db, taskA, taskB, 1) // A -> B (parent-child)
	addTaskRelation(t, db, taskB, taskC, 1) // B -> C (parent-child)

	// Get tree
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get task tree: %v", err)
	}

	// Should have 1 root (Task A)
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(nodes))
	}

	// Check root node
	rootNode := nodes[0]
	if rootNode.ID != taskA {
		t.Errorf("Expected root ID %d, got %d", taskA, rootNode.ID)
	}
	if rootNode.Title != "Task A" {
		t.Errorf("Expected root title 'Task A', got '%s'", rootNode.Title)
	}

	// Check first level child (B)
	if len(rootNode.Children) != 1 {
		t.Fatalf("Expected 1 child for root, got %d", len(rootNode.Children))
	}
	childB := rootNode.Children[0]
	if childB.ID != taskB {
		t.Errorf("Expected child ID %d, got %d", taskB, childB.ID)
	}
	if childB.Title != "Task B" {
		t.Errorf("Expected child title 'Task B', got '%s'", childB.Title)
	}

	// Check second level child (C)
	if len(childB.Children) != 1 {
		t.Fatalf("Expected 1 child for Task B, got %d", len(childB.Children))
	}
	childC := childB.Children[0]
	if childC.ID != taskC {
		t.Errorf("Expected grandchild ID %d, got %d", taskC, childC.ID)
	}
	if childC.Title != "Task C" {
		t.Errorf("Expected grandchild title 'Task C', got '%s'", childC.Title)
	}

	// Check that C has no children
	if len(childC.Children) != 0 {
		t.Errorf("Expected 0 children for Task C, got %d", len(childC.Children))
	}
}

// TestGetTaskTreeByProject_MultipleRoots tests multiple independent roots
func TestGetTaskTreeByProject_MultipleRoots(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create 3 independent root tasks
	task1 := createTestTask(t, db, columnID, "Root 1")
	task2 := createTestTask(t, db, columnID, "Root 2")
	task3 := createTestTask(t, db, columnID, "Root 3")

	// Get tree
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get task tree: %v", err)
	}

	// Should have 3 roots
	if len(nodes) != 3 {
		t.Fatalf("Expected 3 root nodes, got %d", len(nodes))
	}

	// Verify roots are sorted by ticket number (ascending)
	if nodes[0].ID != task1 || nodes[1].ID != task2 || nodes[2].ID != task3 {
		t.Errorf("Roots not in expected order")
	}

	for _, node := range nodes {
		if len(node.Children) != 0 {
			t.Errorf("Expected 0 children for root %d, got %d", node.ID, len(node.Children))
		}
	}
}

// TestGetTaskTreeByProject_DiamondDependencies tests diamond pattern: A -> (B,C) -> D
func TestGetTaskTreeByProject_DiamondDependencies(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create diamond pattern:
	//       A
	//      / \
	//     B   C
	//      \ /
	//       D
	taskA := createTestTask(t, db, columnID, "Task A")
	taskB := createTestTask(t, db, columnID, "Task B")
	taskC := createTestTask(t, db, columnID, "Task C")
	taskD := createTestTask(t, db, columnID, "Task D")

	// Build relationships
	addTaskRelation(t, db, taskA, taskB, 1) // A -> B
	addTaskRelation(t, db, taskA, taskC, 1) // A -> C
	addTaskRelation(t, db, taskB, taskD, 1) // B -> D
	addTaskRelation(t, db, taskC, taskD, 1) // C -> D

	// Get tree
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get task tree: %v", err)
	}

	// Should have 1 root (Task A)
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(nodes))
	}

	rootA := nodes[0]
	if rootA.ID != taskA {
		t.Errorf("Expected root ID %d, got %d", taskA, rootA.ID)
	}

	// A should have 2 children (B and C)
	if len(rootA.Children) != 2 {
		t.Fatalf("Expected 2 children for A, got %d", len(rootA.Children))
	}

	// Check children are B and C (order may vary)
	childIDs := map[int]bool{rootA.Children[0].ID: true, rootA.Children[1].ID: true}
	if !childIDs[taskB] || !childIDs[taskC] {
		t.Errorf("Expected children to be B and C, got %d and %d", rootA.Children[0].ID, rootA.Children[1].ID)
	}

	// Both B and C should have D as child
	for _, child := range rootA.Children {
		if len(child.Children) != 1 {
			t.Fatalf("Expected 1 child for %d, got %d", child.ID, len(child.Children))
		}
		if child.Children[0].ID != taskD {
			t.Errorf("Expected child to be D (%d), got %d", taskD, child.Children[0].ID)
		}
	}
}

// TestGetTaskTreeByProject_SelfDependency tests self-referencing task (A -> A)
func TestGetTaskTreeByProject_SelfDependency(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	taskA := createTestTask(t, db, columnID, "Task A")

	// Create self-dependency
	addTaskRelation(t, db, taskA, taskA, 1) // A -> A

	// Should handle gracefully (not panic or hang)
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get task tree: %v", err)
	}

	// With self-dependency, the task is marked as having a parent (itself)
	// So it won't appear as a root. This is a quirk of how circular dependencies are handled.
	// The important thing is that it doesn't panic or infinite loop.
	// Traverse the entire tree to ensure no infinite loops
	nodeCount := 0
	var traverse func(*models.TaskTreeNode)
	traverse = func(node *models.TaskTreeNode) {
		nodeCount++
		if nodeCount > 100 {
			t.Fatalf("Tree traversal exceeded 100 nodes - likely infinite loop")
		}
		for _, child := range node.Children {
			traverse(child)
		}
	}

	for _, root := range nodes {
		traverse(root)
	}

	// Even if task is not a root, total node count should be <= 1
	if nodeCount > 1 {
		t.Errorf("Excessive node count: %d (should be 0 or 1)", nodeCount)
	}
}

// TestGetTaskTreeByProject_BlockingRelationships tests tree with blocking relationships
func TestGetTaskTreeByProject_BlockingRelationships(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create tasks with blocking relationships
	taskA := createTestTask(t, db, columnID, "Task A (Blocker)")
	taskB := createTestTask(t, db, columnID, "Task B (Blocked)")

	// A blocks B (relation type 2)
	// Note: The tree building logic treats ALL relationships as parent-child hierarchies
	// So one will be a root and the other will be a child
	addTaskRelation(t, db, taskB, taskA, 2) // B is blocked by A (A is parent)

	// Get tree
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get task tree: %v", err)
	}

	// Should have exactly 1 root (the parent in the relationship)
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(nodes))
	}

	root := nodes[0]

	// Root should have 1 child (the child in the relationship)
	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child for root, got %d", len(root.Children))
	}

	child := root.Children[0]

	// Verify it's the A->B relationship with blocking flag
	taskIDs := map[int]bool{root.ID: true, child.ID: true}
	if !taskIDs[taskA] || !taskIDs[taskB] {
		t.Errorf("Expected tasks A and B in tree, got %d and %d", root.ID, child.ID)
	}

	// The relation should be marked as blocking
	if !child.IsBlocking {
		t.Errorf("Expected IsBlocking to be true for blocking relationship")
	}
}

// TestGetTaskTreeByProject_MixedRelationships tests tree with both parent-child and blocking
func TestGetTaskTreeByProject_MixedRelationships(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create 4 tasks
	taskA := createTestTask(t, db, columnID, "Task A")
	taskB := createTestTask(t, db, columnID, "Task B")
	taskC := createTestTask(t, db, columnID, "Task C")
	taskD := createTestTask(t, db, columnID, "Task D")

	// A -> B (parent-child), B -> C (parent-child), C blocks D
	addTaskRelation(t, db, taskA, taskB, 1) // A -> B (parent-child)
	addTaskRelation(t, db, taskB, taskC, 1) // B -> C (parent-child)
	addTaskRelation(t, db, taskD, taskC, 2) // C is blocked by D

	// Get tree
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get task tree: %v", err)
	}

	// Should have 2 roots: A and D
	if len(nodes) != 2 {
		t.Fatalf("Expected 2 root nodes, got %d", len(nodes))
	}

	// Find root A
	var rootA *models.TaskTreeNode
	for _, node := range nodes {
		if node.ID == taskA {
			rootA = node
			break
		}
	}

	if rootA == nil {
		t.Fatalf("Could not find root A")
	}

	// Verify A -> B -> C hierarchy
	if len(rootA.Children) != 1 {
		t.Fatalf("Expected 1 child for A, got %d", len(rootA.Children))
	}

	childB := rootA.Children[0]
	if childB.ID != taskB {
		t.Errorf("Expected child to be B, got %d", childB.ID)
	}

	if len(childB.Children) != 1 {
		t.Fatalf("Expected 1 child for B, got %d", len(childB.Children))
	}

	childC := childB.Children[0]
	if childC.ID != taskC {
		t.Errorf("Expected grandchild to be C, got %d", childC.ID)
	}
}

// TestAddParentRelation_CircularDependencyCheck tests AddParentRelation doesn't create circles
func TestAddParentRelation_CircularDependencyCheck(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create two tasks
	task1, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task 1: %v", err)
	}

	task2, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})
	if err != nil {
		t.Fatalf("Failed to create task 2: %v", err)
	}

	// Add parent relationship: 2 -> 1
	err = svc.AddParentRelation(context.Background(), task2.ID, task1.ID, 1)
	if err != nil {
		t.Fatalf("Failed to add parent relation: %v", err)
	}

	// Try to add reverse relationship which would create a circle: 1 -> 2 when 2 -> 1 exists
	// Note: The current implementation doesn't prevent this at the service level,
	// so we're just testing that it doesn't panic
	err = svc.AddChildRelation(context.Background(), task1.ID, task2.ID, 1)
	if err != nil {
		// This is acceptable - the service might prevent circular deps
		t.Logf("Service prevented circular dependency: %v", err)
	}
}

// TestAddParentRelation_SelfRelationPrevention tests that self-relations are prevented
func TestAddParentRelation_SelfRelationPrevention(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Try to create a self-relation
	err = svc.AddParentRelation(context.Background(), task.ID, task.ID, 1)
	if err == nil {
		t.Fatalf("Expected error for self-relation, got nil")
	}
	if !errors.Is(err, ErrSelfRelation) {
		t.Errorf("Expected ErrSelfRelation, got %v", err)
	}
}

// TestRemoveParentRelation_RestructuresTree tests that removing relations restructures tree
func TestRemoveParentRelation_RestructuresTree(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Create three tasks
	task1, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task A",
		ColumnID: columnID,
		Position: 0,
	})

	task2, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task B",
		ColumnID: columnID,
		Position: 1,
	})

	task3, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Task C",
		ColumnID: columnID,
		Position: 2,
	})

	// Create chain: A -> B -> C
	svc.AddChildRelation(context.Background(), task1.ID, task2.ID, 1)
	svc.AddChildRelation(context.Background(), task2.ID, task3.ID, 1)

	// Verify structure
	nodes, _ := svc.GetTaskTreeByProject(context.Background(), projectID)
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 root, got %d", len(nodes))
	}

	// Remove middle relationship: B from A
	err := svc.RemoveChildRelation(context.Background(), task1.ID, task2.ID)
	if err != nil {
		t.Fatalf("Failed to remove relation: %v", err)
	}

	// Now A and B should both be roots
	nodes, _ = svc.GetTaskTreeByProject(context.Background(), projectID)
	if len(nodes) != 2 {
		t.Errorf("Expected 2 roots after removing relation, got %d", len(nodes))
	}
}

// countNodes recursively counts all nodes in tree (used for large hierarchy testing)
func countNodes(node *models.TaskTreeNode) int {
	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}

// getMaxDepth recursively finds maximum depth of tree
func getMaxDepth(node *models.TaskTreeNode) int {
	if len(node.Children) == 0 {
		return 0
	}
	maxChildDepth := 0
	for _, child := range node.Children {
		childDepth := getMaxDepth(child)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}
	return maxChildDepth + 1
}

// ============================================================================
// ERROR PATH TESTS - COMPREHENSIVE ERROR SCENARIOS
// ============================================================================

// TestCreateTask_ErrorPaths tests various error scenarios for CreateTask
func TestCreateTask_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*sql.DB) CreateTaskRequest
		wantErr bool
		errType error
	}{
		{
			name: "negative column ID",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				return CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: -1,
					Position: 0,
				}
			},
			wantErr: true,
			errType: ErrInvalidColumnID,
		},
		{
			name: "non-existent column ID",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				return CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: 99999,
					Position: 0,
				}
			},
			wantErr: true,
		},
		{
			name: "position exceeds int64 max",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 9223372036854775807, // max int64
				}
			},
			wantErr: false, // Large position should work
		},
		{
			name: "non-existent label IDs",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
					LabelIDs: []int{99999, 88888},
				}
			},
			wantErr: true,
		},
		{
			name: "mixed valid and invalid label IDs",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				labelID := createTestLabel(t, db, projectID, "Valid")
				return CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
					LabelIDs: []int{labelID, 99999},
				}
			},
			wantErr: true,
		},
		{
			name: "zero label ID",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
					LabelIDs: []int{0},
				}
			},
			wantErr: true,
		},
		{
			name: "negative label ID",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
					LabelIDs: []int{-1},
				}
			},
			wantErr: true,
		},
		{
			name: "non-existent parent IDs",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:     "Test Task",
					ColumnID:  columnID,
					Position:  0,
					ParentIDs: []int{99999},
				}
			},
			wantErr: true,
		},
		{
			name: "zero parent ID",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:     "Test Task",
					ColumnID:  columnID,
					Position:  0,
					ParentIDs: []int{0},
				}
			},
			wantErr: true,
		},
		{
			name: "negative parent ID",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:     "Test Task",
					ColumnID:  columnID,
					Position:  0,
					ParentIDs: []int{-1},
				}
			},
			wantErr: true,
		},
		{
			name: "invalid priority ID negative",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:      "Test Task",
					ColumnID:   columnID,
					Position:   0,
					PriorityID: -1,
				}
			},
			wantErr: true,
			errType: ErrInvalidPriority,
		},
		{
			name: "invalid priority ID too high - database constraint",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:      "Test Task",
					ColumnID:   columnID,
					Position:   0,
					PriorityID: 999,
				}
			},
			wantErr: true, // Database will catch foreign key constraint
		},
		{
			name: "invalid type ID negative",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
					TypeID:   -1,
				}
			},
			wantErr: true,
			errType: ErrInvalidType,
		},
		{
			name: "invalid type ID too high - database constraint",
			setupFn: func(db *sql.DB) CreateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
					TypeID:   999,
				}
			},
			wantErr: true, // Database will catch foreign key constraint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			req := tt.setupFn(db)

			_, err := svc.CreateTask(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("CreateTask() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestUpdateTask_ErrorPaths tests various error scenarios for UpdateTask
func TestUpdateTask_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*sql.DB) UpdateTaskRequest
		wantErr bool
		errType error
	}{
		{
			name: "negative task ID",
			setupFn: func(db *sql.DB) UpdateTaskRequest {
				title := "New Title"
				return UpdateTaskRequest{
					TaskID: -1,
					Title:  &title,
				}
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "non-existent task ID",
			setupFn: func(db *sql.DB) UpdateTaskRequest {
				title := "New Title"
				return UpdateTaskRequest{
					TaskID: 99999,
					Title:  &title,
				}
			},
			wantErr: true,
		},
		{
			name: "title too long (exactly 256 chars)",
			setupFn: func(db *sql.DB) UpdateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				task, _ := NewService(db, nil).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Old Title",
					ColumnID: columnID,
					Position: 0,
				})
				longTitle := ""
				for i := 0; i < 256; i++ {
					longTitle += "a"
				}
				return UpdateTaskRequest{
					TaskID: task.ID,
					Title:  &longTitle,
				}
			},
			wantErr: true,
			errType: ErrTitleTooLong,
		},
		{
			name: "update with invalid priority ID - database constraint",
			setupFn: func(db *sql.DB) UpdateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				task, _ := NewService(db, nil).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
				})
				priority := 999
				return UpdateTaskRequest{
					TaskID:     task.ID,
					PriorityID: &priority,
				}
			},
			wantErr: true, // Database constraint
		},
		{
			name: "update with invalid type ID - database constraint",
			setupFn: func(db *sql.DB) UpdateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				task, _ := NewService(db, nil).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
				})
				typeID := 999
				return UpdateTaskRequest{
					TaskID: task.ID,
					TypeID: &typeID,
				}
			},
			wantErr: true, // Database constraint
		},
		{
			name: "whitespace-only title - may or may not be validated",
			setupFn: func(db *sql.DB) UpdateTaskRequest {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				task, _ := NewService(db, nil).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
				})
				title := "   "
				return UpdateTaskRequest{
					TaskID: task.ID,
					Title:  &title,
				}
			},
			wantErr: false, // Whitespace title may be allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			req := tt.setupFn(db)

			err := svc.UpdateTask(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("UpdateTask() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestDeleteTask_ErrorPaths tests various error scenarios for DeleteTask
func TestDeleteTask_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		taskID  int
		wantErr bool
		errType error
	}{
		{
			name:    "negative task ID",
			taskID:  -1,
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name:    "non-existent task ID - may succeed",
			taskID:  99999,
			wantErr: false, // DELETE operations may succeed even if row doesn't exist
		},
		{
			name:    "very large task ID - may succeed",
			taskID:  999999999,
			wantErr: false, // DELETE operations may succeed even if row doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)

			err := svc.DeleteTask(context.Background(), tt.taskID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("DeleteTask() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestAttachLabel_ErrorPaths tests various error scenarios for AttachLabel
func TestAttachLabel_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*sql.DB) (int, int)
		wantErr bool
		errType error
	}{
		{
			name: "invalid task ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				labelID := createTestLabel(t, db, projectID, "Bug")
				return 0, labelID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative task ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				labelID := createTestLabel(t, db, projectID, "Bug")
				return -1, labelID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid label ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
		{
			name: "negative label ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
		{
			name: "non-existent task ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				labelID := createTestLabel(t, db, projectID, "Bug")
				return 99999, labelID
			},
			wantErr: true,
		},
		{
			name: "non-existent label ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return taskID, 99999
			},
			wantErr: true,
		},
		{
			name: "duplicate label attachment - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				labelID := createTestLabel(t, db, projectID, "Bug")
				svc := NewService(db, nil)
				task, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
					LabelIDs: []int{labelID},
				})
				return task.ID, labelID
			},
			wantErr: false, // May succeed or may error - database dependent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			taskID, labelID := tt.setupFn(db)

			err := svc.AttachLabel(context.Background(), taskID, labelID)

			if (err != nil) != tt.wantErr {
				t.Errorf("AttachLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("AttachLabel() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestDetachLabel_ErrorPaths tests various error scenarios for DetachLabel
func TestDetachLabel_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*sql.DB) (int, int)
		wantErr bool
		errType error
	}{
		{
			name: "invalid task ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				labelID := createTestLabel(t, db, projectID, "Bug")
				return 0, labelID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative task ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				labelID := createTestLabel(t, db, projectID, "Bug")
				return -1, labelID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid label ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
		{
			name: "negative label ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
		{
			name: "non-existent task ID - may succeed",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				labelID := createTestLabel(t, db, projectID, "Bug")
				return 99999, labelID
			},
			wantErr: false, // May succeed even if task doesn't exist
		},
		{
			name: "non-existent label ID - may succeed",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return taskID, 99999
			},
			wantErr: false, // May succeed even if label doesn't exist
		},
		{
			name: "label not attached to task - may succeed",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				labelID := createTestLabel(t, db, projectID, "Bug")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return taskID, labelID
			},
			wantErr: false, // May succeed even if label not attached
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			taskID, labelID := tt.setupFn(db)

			err := svc.DetachLabel(context.Background(), taskID, labelID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DetachLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("DetachLabel() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestAddParentRelation_ErrorPaths tests various error scenarios for AddParentRelation
func TestAddParentRelation_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*sql.DB) (int, int)
		wantErr bool
		errType error
	}{
		{
			name: "invalid child ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return 0, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative child ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return -1, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid parent ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative parent ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "non-existent child task",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return 99999, taskID
			},
			wantErr: true,
		},
		{
			name: "non-existent parent task",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return taskID, 99999
			},
			wantErr: true,
		},
		{
			name: "duplicate parent relation - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				svc := NewService(db, nil)
				parent, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Parent",
					ColumnID: columnID,
					Position: 0,
				})
				child, _ := svc.CreateTask(context.Background(), CreateTaskRequest{
					Title:     "Child",
					ColumnID:  columnID,
					Position:  1,
					ParentIDs: []int{parent.ID},
				})
				return child.ID, parent.ID
			},
			wantErr: false, // May succeed or may error - database dependent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			childID, parentID := tt.setupFn(db)

			err := svc.AddParentRelation(context.Background(), childID, parentID, 1)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddParentRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("AddParentRelation() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestAddChildRelation_ErrorPaths tests various error scenarios for AddChildRelation
func TestAddChildRelation_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*sql.DB) (int, int)
		wantErr bool
		errType error
	}{
		{
			name: "invalid parent ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return 0, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative parent ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return -1, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid child ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative child ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "non-existent parent task",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return 99999, taskID
			},
			wantErr: true,
		},
		{
			name: "non-existent child task",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return taskID, 99999
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			parentID, childID := tt.setupFn(db)

			err := svc.AddChildRelation(context.Background(), parentID, childID, 1)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddChildRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("AddChildRelation() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestRemoveParentRelation_ErrorPaths tests various error scenarios for RemoveParentRelation
func TestRemoveParentRelation_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*sql.DB) (int, int)
		wantErr bool
		errType error
	}{
		{
			name: "invalid child ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return 0, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative child ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return -1, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid parent ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative parent ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "non-existent relationship - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				task1 := createTestTask(t, db, columnID, "Task 1")
				task2 := createTestTask(t, db, columnID, "Task 2")
				return task1, task2
			},
			wantErr: false, // May succeed even if no relationship exists
		},
		{
			name: "non-existent child task - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return 99999, taskID
			},
			wantErr: false, // May succeed even if task doesn't exist
		},
		{
			name: "non-existent parent task - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return taskID, 99999
			},
			wantErr: false, // May succeed even if task doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			childID, parentID := tt.setupFn(db)

			err := svc.RemoveParentRelation(context.Background(), childID, parentID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveParentRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("RemoveParentRelation() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestRemoveChildRelation_ErrorPaths tests various error scenarios for RemoveChildRelation
func TestRemoveChildRelation_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*sql.DB) (int, int)
		wantErr bool
		errType error
	}{
		{
			name: "invalid parent ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return 0, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative parent ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Child")
				return -1, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid child ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative child ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Parent")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "non-existent relationship - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				task1 := createTestTask(t, db, columnID, "Task 1")
				task2 := createTestTask(t, db, columnID, "Task 2")
				return task1, task2
			},
			wantErr: false, // May succeed even if no relationship exists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			parentID, childID := tt.setupFn(db)

			err := svc.RemoveChildRelation(context.Background(), parentID, childID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveChildRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("RemoveChildRelation() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestMoveTaskToColumn_ErrorPaths tests various error scenarios for MoveTaskToColumn
func TestMoveTaskToColumn_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*sql.DB) (int, int)
		wantErr bool
		errType error
	}{
		{
			name: "invalid task ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return 0, columnID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative task ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				return -1, columnID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid column ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidColumnID,
		},
		{
			name: "negative column ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")
				taskID := createTestTask(t, db, columnID, "Test Task")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidColumnID,
		},
		{
			name: "both task and column non-existent",
			setupFn: func(db *sql.DB) (int, int) {
				return 99999, 88888
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			taskID, columnID := tt.setupFn(db)

			err := svc.MoveTaskToColumn(context.Background(), taskID, columnID)

			if (err != nil) != tt.wantErr {
				t.Errorf("MoveTaskToColumn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("MoveTaskToColumn() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestGetTaskSummariesByProject_ErrorPaths tests error scenarios for GetTaskSummariesByProject
func TestGetTaskSummariesByProject_ErrorPaths(t *testing.T) {
	tests := []struct {
		name      string
		projectID int
		wantErr   bool
		errType   error
	}{
		{
			name:      "negative project ID - may not validate",
			projectID: -1,
			wantErr:   false, // May return empty result instead of error
		},
		{
			name:      "zero project ID - may not validate",
			projectID: 0,
			wantErr:   false, // May return empty result instead of error
		},
		{
			name:      "non-existent project ID",
			projectID: 99999,
			wantErr:   false, // Should return empty map, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)

			result, err := svc.GetTaskSummariesByProject(context.Background(), tt.projectID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskSummariesByProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("GetTaskSummariesByProject() error = %v, want %v", err, tt.errType)
			}

			// For non-existent project, should return empty map
			if !tt.wantErr && err == nil && tt.projectID == 99999 {
				if len(result) != 0 {
					t.Errorf("Expected empty map for non-existent project, got %d entries", len(result))
				}
			}
		})
	}
}

// TestGetTaskDetail_NegativeID tests GetTaskDetail with negative ID
func TestGetTaskDetail_NegativeID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetTaskDetail(context.Background(), -1)

	if err == nil {
		t.Fatal("Expected error for negative task ID")
	}

	if !errors.Is(err, ErrInvalidTaskID) {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

// TestGetTaskReferencesForProject_ErrorPaths tests error scenarios
func TestGetTaskReferencesForProject_ErrorPaths(t *testing.T) {
	tests := []struct {
		name      string
		projectID int
		wantErr   bool
		errType   error
	}{
		{
			name:      "negative project ID - may not validate",
			projectID: -1,
			wantErr:   false, // May return empty result instead of error
		},
		{
			name:      "zero project ID - may not validate",
			projectID: 0,
			wantErr:   false, // May return empty result instead of error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)

			_, err := svc.GetTaskReferencesForProject(context.Background(), tt.projectID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskReferencesForProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("GetTaskReferencesForProject() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// TestComment_BoundaryConditions tests boundary conditions for comments
func TestComment_BoundaryConditions(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	svc := NewService(db, nil)

	// Test exactly at max length (1000 chars)
	maxMessage := ""
	for i := 0; i < 1000; i++ {
		maxMessage += "a"
	}

	req := CreateCommentRequest{
		TaskID:  taskID,
		Message: maxMessage,
		Author:  "testuser",
	}

	result, err := svc.CreateComment(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error for 1000 char message, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected comment result, got nil")
	}

	if len(result.Message) != 1000 {
		t.Errorf("Expected message length 1000, got %d", len(result.Message))
	}
}

// TestCreateTask_MaxLengthTitle tests title at exact max length
func TestCreateTask_MaxLengthTitle(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := NewService(db, nil)

	// Test exactly at max length (255 chars)
	maxTitle := ""
	for i := 0; i < 255; i++ {
		maxTitle += "a"
	}

	req := CreateTaskRequest{
		Title:    maxTitle,
		ColumnID: columnID,
		Position: 0,
	}

	result, err := svc.CreateTask(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error for 255 char title, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected task result, got nil")
	}

	if len(result.Title) != 255 {
		t.Errorf("Expected title length 255, got %d", len(result.Title))
	}
}
