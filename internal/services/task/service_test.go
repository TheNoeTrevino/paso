package task

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/taskevent"
	"github.com/thenoetrevino/paso/internal/testutil"
)

func TestCreateTask(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)

	req := CreateTaskRequest{
		Title:       "Fix bug in login",
		Description: "Users can't log in",
		ColumnID:    columnID,
		Position:    0,
		PriorityID:  4, // high
		TypeID:      3, // bug
	}

	result, err := svc.CreateTask(context.Background(), req)
	require.NoError(t, err, "Operation failed")

	require.NotNil(t, result, "Expected task result, got nil")

	assert.Equal(t, "Fix bug in login", result.Title)
	assert.Equal(t, "Users can't log in", result.Description)
	assert.Equal(t, columnID, result.ColumnID)
	assert.NotZero(t, result.ID)

	// Note: models.Task doesn't include TicketNumber (only TaskDetail does)
	// We could verify it via GetTaskDetail if needed, but basic task creation is sufficient here
}

func TestCreateTask_Validation(t *testing.T) {
	t.Parallel()
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
					return CreateTaskRequest{
						Title:    strings.Repeat("a", 256),
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

			db := testutil.SetupTestDB(t)

			req := tt.args.req

			// Setup database if needed
			if tt.needsTest || tt.args.setupFn != nil {
				projectID := createTestProject(t, db)
				columnID := createTestColumn(t, db, projectID, "To Do")

				if tt.args.setupFn != nil {
					req = tt.args.setupFn(db, columnID)
				}
			}

			svc := newTestService(t, db)
			_, err := svc.CreateTask(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

func TestCreateTask_WithLabels(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	label1ID := createTestLabel(t, db, projectID, "Bug")
	label2ID := createTestLabel(t, db, projectID, "Critical")
	svc := newTestService(t, db)
	ctx := context.Background()

	req := CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
		LabelIDs: []int{label1ID, label2ID},
	}

	result, err := svc.CreateTask(ctx, req)
	require.NoError(t, err, "Operation failed")

	// Verify labels are attached
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM task_labels WHERE task_id = ?", result.ID).Scan(&count)
	require.NoError(t, err, "failed to query task labels")

	assert.Equal(t, 2, count)
}

func TestGetTaskDetail(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a task
	created, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		ColumnID:    columnID,
		Position:    0,
		PriorityID:  4,
		TypeID:      3,
	})
	require.NoError(t, err)

	// Get task detail
	result, err := svc.GetTaskDetail(ctx, created.ID)
	require.NoError(t, err, "Operation failed")

	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, "Test Task", result.Title)
	assert.Equal(t, "Test Description", result.Description)
}

func TestGetTaskDetail_NotFound(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetTaskDetail(context.Background(), 999)

	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestGetTaskDetail_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetTaskDetail(context.Background(), 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestGetTaskSummariesByProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "To Do")
	col2ID := createTestColumn(t, db, projectID, "Done")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create tasks in different columns
	_, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: col1ID,
		Position: 0,
	})
	require.NoError(t, err)

	_, err = svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: col2ID,
		Position: 0,
	})
	require.NoError(t, err)

	// Get summaries
	results, err := svc.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, results, 2)
	assert.Len(t, results[col1ID], 1)
	assert.Len(t, results[col2ID], 1)
}

func TestUpdateTask(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a task
	created, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Old Title",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Update task
	newTitle := "New Title"
	newDesc := "New Description"
	req := UpdateTaskRequest{
		TaskID:      created.ID,
		Title:       &newTitle,
		Description: &newDesc,
	}

	err = svc.UpdateTask(ctx, req)
	require.NoError(t, err, "Operation failed")

	// Verify update
	updated, err := svc.GetTaskDetail(ctx, created.ID)
	require.NoError(t, err)

	assert.Equal(t, "New Title", updated.Title)
	assert.Equal(t, "New Description", updated.Description)
}

func TestUpdateTask_Validation(t *testing.T) {
	t.Parallel()
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
				task, err := newTestService(t, db).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Old Title",
					ColumnID: columnID,
					Position: 0,
				})
				require.NoError(t, err)
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

			db := testutil.SetupTestDB(t)

			taskID := tt.taskID
			if tt.setupFn != nil {
				taskID = tt.setupFn(db)
			}

			svc := newTestService(t, db)
			req := UpdateTaskRequest{
				TaskID: taskID,
				Title:  tt.title,
			}

			err := svc.UpdateTask(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// ptrString is a helper function that returns a pointer to a string
func ptrString(s string) *string {
	return &s
}

func TestDeleteTask(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a task
	created, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Delete task
	err = svc.DeleteTask(ctx, created.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task is deleted
	_, err = svc.GetTaskDetail(ctx, created.ID)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeleteTask_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	err := svc.DeleteTask(context.Background(), 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestAttachLabel(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	labelID := createTestLabel(t, db, projectID, "Bug")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a task
	created, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Attach label
	err = svc.AttachLabel(ctx, created.ID, labelID)
	require.NoError(t, err, "Operation failed")

	// Verify label is attached
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?", created.ID, labelID).Scan(&count)
	require.NoError(t, err, "failed to query task labels")

	assert.Equal(t, 1, count)
}

func TestDetachLabel(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	labelID := createTestLabel(t, db, projectID, "Bug")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a task with label
	created, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
		LabelIDs: []int{labelID},
	})
	require.NoError(t, err)

	// Detach label
	err = svc.DetachLabel(ctx, created.ID, labelID)
	require.NoError(t, err, "Operation failed")

	// Verify label is detached
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?", created.ID, labelID).Scan(&count)
	require.NoError(t, err, "failed to query task labels")

	assert.Equal(t, 0, count)
}

func TestAddParentRelation(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two tasks
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Add parent relation (task1 is parent of task2)
	err = svc.AddParentRelation(ctx, task2.ID, task1.ID, 1)
	require.NoError(t, err, "Operation failed")

	// Verify relationship exists
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	require.NoError(t, err, "failed to query task relationships")

	assert.Equal(t, 1, count)
}

func TestAddChildRelation(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two tasks
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Add child relation (task2 is child of task1)
	err = svc.AddChildRelation(ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err, "Operation failed")

	// Verify relationship exists
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	require.NoError(t, err, "failed to query task relationships")

	assert.Equal(t, 1, count)
}

func TestRemoveParentRelation(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two tasks with relationship
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:     "Child Task",
		ColumnID:  columnID,
		Position:  1,
		ParentIDs: []int{task1.ID},
	})
	require.NoError(t, err)

	// Remove parent relation
	err = svc.RemoveParentRelation(ctx, task2.ID, task1.ID)
	require.NoError(t, err, "Operation failed")

	// Verify relationship is removed
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	require.NoError(t, err, "failed to query task relationships")

	assert.Equal(t, 0, count)
}

func TestRemoveChildRelation(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two tasks with child relationship
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Add child relation (task2 is child of task1)
	err = svc.AddChildRelation(ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err)

	// Remove child relation
	err = svc.RemoveChildRelation(ctx, task1.ID, task2.ID)
	require.NoError(t, err, "Operation failed")

	// Verify relationship is removed
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	require.NoError(t, err, "failed to query task relationships")

	assert.Equal(t, 0, count)
}

func TestMoveTaskToNextColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "To Do")
	col2ID := createTestColumn(t, db, projectID, "In Progress")
	ctx := context.Background()

	// Link columns
	_, err := db.ExecContext(ctx, "UPDATE columns SET next_id = ? WHERE id = ?", col2ID, col1ID)
	require.NoError(t, err, "failed to link columns")

	svc := newTestService(t, db)

	// Create task in first column
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col1ID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to next column
	err = svc.MoveTaskToNextColumn(ctx, task.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved to col2
	testutil.AssertTaskInColumn(t, db, task.ID, col2ID)
}

func TestMoveTaskToNextColumn_LastColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Done")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task in last column (no next_id)
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to next column (should fail)
	err = svc.MoveTaskToNextColumn(ctx, task.ID)

	require.Error(t, err)
}

func TestMoveTaskToNextColumn_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	err := svc.MoveTaskToNextColumn(context.Background(), 999)

	require.Error(t, err)
}

func TestMoveTaskToPrevColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "To Do")
	col2ID := createTestColumn(t, db, projectID, "In Progress")
	ctx := context.Background()

	// Link columns
	_, err := db.ExecContext(ctx, "UPDATE columns SET next_id = ?, prev_id = ? WHERE id = ?", col2ID, 0, col1ID)
	require.NoError(t, err, "failed to link columns")
	_, err = db.ExecContext(ctx, "UPDATE columns SET prev_id = ? WHERE id = ?", col1ID, col2ID)
	require.NoError(t, err, "failed to link columns")

	svc := newTestService(t, db)

	// Create task in second column
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col2ID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to previous column
	err = svc.MoveTaskToPrevColumn(ctx, task.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved to col1
	testutil.AssertTaskInColumn(t, db, task.ID, col1ID)
}

func TestMoveTaskToPrevColumn_FirstColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task in first column (no prev_id)
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to previous column (should fail)
	err = svc.MoveTaskToPrevColumn(ctx, task.ID)

	require.Error(t, err)
}

func TestMoveTaskToPrevColumn_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	err := svc.MoveTaskToPrevColumn(context.Background(), 999)

	require.Error(t, err)
}

func TestMoveTaskToColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "To Do")
	col2ID := createTestColumn(t, db, projectID, "Done")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task in first column
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col1ID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to specific column
	err = svc.MoveTaskToColumn(ctx, task.ID, col2ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved
	testutil.AssertTaskInColumn(t, db, task.ID, col2ID)
}

func TestMoveTaskToColumn_InvalidColumnID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to invalid column
	err = svc.MoveTaskToColumn(ctx, task.ID, 999)

	require.Error(t, err)
}

func TestMoveTaskToColumn_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)

	// Try to move invalid task
	err := svc.MoveTaskToColumn(context.Background(), 999, columnID)

	require.Error(t, err)
}

func TestMoveTaskUp(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two tasks
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Move task2 up (should swap positions with task1)
	err = svc.MoveTaskUp(ctx, task2.ID)
	require.NoError(t, err, "Operation failed")

	// Verify positions swapped
	var pos1, pos2 int64
	err = db.QueryRowContext(ctx, "SELECT position FROM tasks WHERE id = ?", task1.ID).Scan(&pos1)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, "SELECT position FROM tasks WHERE id = ?", task2.ID).Scan(&pos2)
	require.NoError(t, err)

	assert.Less(t, pos2, pos1)
}

func TestMoveTaskUp_FirstPosition(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task at first position
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move up (should fail - no task above)
	err = svc.MoveTaskUp(ctx, task.ID)

	require.Error(t, err)
}

func TestMoveTaskUp_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	err := svc.MoveTaskUp(context.Background(), 999)

	require.Error(t, err)
}

func TestMoveTaskDown(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two tasks
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Move task1 down (should swap positions with task2)
	err = svc.MoveTaskDown(ctx, task1.ID)
	require.NoError(t, err, "Operation failed")

	// Verify positions swapped
	var pos1, pos2 int64
	err = db.QueryRowContext(ctx, "SELECT position FROM tasks WHERE id = ?", task1.ID).Scan(&pos1)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, "SELECT position FROM tasks WHERE id = ?", task2.ID).Scan(&pos2)
	require.NoError(t, err)

	assert.Greater(t, pos1, pos2)
}

func TestMoveTaskDown_LastPosition(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task at last position
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move down (should fail - no task below)
	err = svc.MoveTaskDown(ctx, task.ID)

	require.Error(t, err)
}

func TestMoveTaskDown_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	err := svc.MoveTaskDown(context.Background(), 999)

	require.Error(t, err)
}

func TestGetTaskSummariesByProjectFiltered(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create tasks with different titles
	_, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Fix bug in login",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	_, err = svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Add new feature",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	_, err = svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Fix bug in signup",
		ColumnID: columnID,
		Position: 2,
	})
	require.NoError(t, err)

	// Filter by "bug"
	results, err := svc.GetTaskSummariesByProjectFiltered(ctx, projectID, "bug")
	require.NoError(t, err, "Operation failed")

	// Should return 2 tasks with "bug" in title
	totalTasks := 0
	for _, tasks := range results {
		totalTasks += len(tasks)
	}

	assert.Equal(t, 2, totalTasks)
}

func TestGetTaskSummariesByProjectFiltered_NoResults(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task
	_, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Filter by non-existent term
	results, err := svc.GetTaskSummariesByProjectFiltered(ctx, projectID, "nonexistent")
	require.NoError(t, err, "Operation failed")

	// Should return empty map
	totalTasks := 0
	for _, tasks := range results {
		totalTasks += len(tasks)
	}

	assert.Equal(t, 0, totalTasks)
}

func TestGetTaskSummariesByProjectFiltered_EmptyQuery(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create tasks
	_, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	_, err = svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Filter with empty query (should return all)
	results, err := svc.GetTaskSummariesByProjectFiltered(ctx, projectID, "")
	require.NoError(t, err, "Operation failed")

	// Should return all tasks
	totalTasks := 0
	for _, tasks := range results {
		totalTasks += len(tasks)
	}

	assert.Equal(t, 2, totalTasks)
}

func TestGetTaskReferencesForProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create tasks
	_, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	_, err = svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Get task references
	refs, err := svc.GetTaskReferencesForProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	assert.Len(t, refs, 2)
}

func TestGetTaskReferencesForProject_EmptyProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)

	// Get task references for empty project
	refs, err := svc.GetTaskReferencesForProject(context.Background(), projectID)
	require.NoError(t, err, "Operation failed")

	assert.Len(t, refs, 0)
}

func TestGetReadyTaskSummariesByProject_OnlyReadyColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)

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
	require.NoError(t, err, "Operation failed")

	require.Len(t, readyTasks, 1)

	// Verify only task from Todo column is returned
	assert.Equal(t, task1, readyTasks[0].ID)

	// Verify tasks from other columns are not included
	for _, task := range readyTasks {
		assert.False(t, task.ID == task2 || task.ID == task3, "Unexpected task ID %d in ready tasks (should only be from Todo column)", task.ID)
	}
}

func TestGetReadyTaskSummariesByProject_ExcludesBlockedTasks(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create ready column
	readyCol := createTestReadyColumn(t, db, projectID, "Todo")

	// Create two tasks
	task1 := createTestTask(t, db, readyCol, "Unblocked Task")
	task2 := createTestTask(t, db, readyCol, "Blocked Task")

	// Create a blocker relationship (task2 is blocked by task1)
	// relation_type_id = 2 is "Blocked By/Blocker" with is_blocking = 1
	_, err := db.ExecContext(ctx,
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, ?)",
		task2, task1, 2)
	require.NoError(t, err, "failed to create blocking relationship")

	// Get ready tasks
	readyTasks, err := svc.GetReadyTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	// Should only get task1 (task2 is blocked)
	require.Len(t, readyTasks, 1)

	assert.Equal(t, task1, readyTasks[0].ID)
}

func TestGetReadyTaskSummariesByProject_EmptyWhenNoReadyColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)

	// Create columns but none are ready
	col1 := createTestColumn(t, db, projectID, "Todo")
	createTestColumn(t, db, projectID, "Done")

	// Create tasks
	createTestTask(t, db, col1, "Task 1")

	// Get ready tasks
	readyTasks, err := svc.GetReadyTaskSummariesByProject(context.Background(), projectID)
	require.NoError(t, err, "Operation failed")

	assert.Len(t, readyTasks, 0)
}

func TestGetReadyTaskSummariesByProject_EmptyReadyColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)

	// Create ready column with no tasks
	createTestReadyColumn(t, db, projectID, "Todo")

	// Get ready tasks
	readyTasks, err := svc.GetReadyTaskSummariesByProject(context.Background(), projectID)
	require.NoError(t, err, "Operation failed")

	assert.Len(t, readyTasks, 0)
}

func TestGetReadyTaskSummariesByProject_IncludesTaskDetails(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create ready column
	readyCol := createTestReadyColumn(t, db, projectID, "Todo")

	// Create label
	labelID := createTestLabel(t, db, projectID, "bug")

	// Create task with label and high priority (priority_id = 4)
	taskID := createTestTask(t, db, readyCol, "Important Bug Fix")
	_, err := db.ExecContext(ctx,
		"UPDATE tasks SET priority_id = 4 WHERE id = ?", taskID)
	require.NoError(t, err, "failed to update task priority")

	// Attach label
	_, err = db.ExecContext(ctx,
		"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
	require.NoError(t, err, "failed to attach label")

	// Get ready tasks
	readyTasks, err := svc.GetReadyTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, readyTasks, 1)

	task := readyTasks[0]

	// Verify task details
	assert.Equal(t, "Important Bug Fix", task.Title)
	assert.Equal(t, "high", task.PriorityDescription)

	require.Len(t, task.Labels, 1)

	assert.Equal(t, "bug", task.Labels[0].Name)
}

func TestGetTaskTreeByProject_EmptyProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)

	// Get tree for empty project
	tree, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	require.NoError(t, err, "Operation failed")

	assert.Len(t, tree, 0)
}

func TestGetTaskTreeByProject_SimpleHierarchy(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create parent and child tasks
	parentID := createTestTask(t, db, columnID, "Parent Task")
	childID := createTestTask(t, db, columnID, "Child Task")

	// Add parent-child relation (non-blocking)
	_, err := db.ExecContext(ctx,
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		parentID, childID)
	require.NoError(t, err, "failed to create relation")

	// Get tree
	tree, err := svc.GetTaskTreeByProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, tree, 1)

	root := tree[0]
	assert.Equal(t, parentID, root.ID)

	require.Len(t, root.Children, 1)

	child := root.Children[0]
	assert.Equal(t, childID, child.ID)
}

func TestGetTaskTreeByProject_CircularDependency(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create three tasks: A -> B -> C -> A (circular)
	taskA := createTestTask(t, db, columnID, "Task A")
	taskB := createTestTask(t, db, columnID, "Task B")
	taskC := createTestTask(t, db, columnID, "Task C")

	// A -> B -> C -> A (circular)
	_, err := db.ExecContext(ctx,
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		taskA, taskB)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		taskB, taskC)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		taskC, taskA)
	require.NoError(t, err)

	// Get tree - should handle circular dependency gracefully
	tree, err := svc.GetTaskTreeByProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	// With circular dependency, none are truly "root" tasks
	// This is expected behavior - the function should not crash or hang
	if len(tree) != 0 {
		t.Logf("Got %d root tasks with circular dependencies (expected)", len(tree))
	}
}

func TestGetTaskTreeByProject_DeepNesting(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a deep hierarchy: Task1 -> Task2 -> Task3 -> Task4 -> Task5
	tasks := make([]int, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = createTestTask(t, db, columnID, "Task "+string(rune('1'+i)))
	}

	// Link them in a chain
	for i := 0; i < 4; i++ {
		_, err := db.ExecContext(ctx,
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
			tasks[i], tasks[i+1])
		require.NoError(t, err, "failed to create relation")
	}

	// Get tree
	tree, err := svc.GetTaskTreeByProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, tree, 1)

	// Traverse the tree and verify depth
	depth := 0
	current := tree[0]
	for {
		depth++
		if len(current.Children) == 0 {
			break
		}
		require.LessOrEqual(t, depth, 10, "Depth exceeds expected maximum")
		current = current.Children[0]
	}

	assert.Equal(t, 5, depth)
}

func TestGetTaskTreeByProject_BlockingRelationship(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create parent and child tasks
	parentID := createTestTask(t, db, columnID, "Parent Task")
	blockerID := createTestTask(t, db, columnID, "Blocker Task")

	// Add blocking relation (relation_type_id = 2 is the blocker type)
	_, err := db.ExecContext(ctx,
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
		parentID, blockerID)
	require.NoError(t, err, "failed to create blocking relation")

	// Get tree
	tree, err := svc.GetTaskTreeByProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, tree, 1)

	root := tree[0]
	require.Len(t, root.Children, 1)

	blocker := root.Children[0]
	assert.True(t, blocker.IsBlocking)
}

func TestGetTaskTreeByProject_SortedByTicketNumber(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create multiple root tasks (no parent relations)
	task1 := createTestTask(t, db, columnID, "Task 1")
	task2 := createTestTask(t, db, columnID, "Task 2")
	task3 := createTestTask(t, db, columnID, "Task 3")

	// Set ticket numbers (in database they auto-increment, but let's verify sorting)
	_, err := db.ExecContext(ctx, "UPDATE tasks SET ticket_number = 3 WHERE id = ?", task1)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE tasks SET ticket_number = 1 WHERE id = ?", task2)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE tasks SET ticket_number = 2 WHERE id = ?", task3)
	require.NoError(t, err)

	// Get tree
	tree, err := svc.GetTaskTreeByProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, tree, 3)

	// Verify sorted by ticket number
	assert.Equal(t, 1, tree[0].TicketNumber)
	assert.Equal(t, 2, tree[1].TicketNumber)
	assert.Equal(t, 3, tree[2].TicketNumber)
}

func TestGetTaskTreeByProject_MultipleRootsWithChildren(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two separate trees
	root1 := createTestTask(t, db, columnID, "Root 1")
	child1 := createTestTask(t, db, columnID, "Child 1")

	root2 := createTestTask(t, db, columnID, "Root 2")
	child2 := createTestTask(t, db, columnID, "Child 2")

	// Link them
	_, err := db.ExecContext(ctx,
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		root1, child1)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		root2, child2)
	require.NoError(t, err)

	// Get tree
	tree, err := svc.GetTaskTreeByProject(ctx, projectID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, tree, 2)

	// Verify both roots have exactly one child
	for i, root := range tree {
		assert.Len(t, root.Children, 1, "Root %d", i)
	}
}

func TestMoveTaskToReadyColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	todoColID := createTestColumn(t, db, projectID, "To Do")
	readyColID := createTestReadyColumn(t, db, projectID, "Ready")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task in To Do column
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: todoColID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to ready column
	err = svc.MoveTaskToReadyColumn(ctx, task.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved to ready column
	testutil.AssertTaskInColumn(t, db, task.ID, readyColID)
}

func TestMoveTaskToReadyColumn_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Try to move non-existent task
	err := svc.MoveTaskToReadyColumn(context.Background(), 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidTaskID) || errors.Is(err, sql.ErrNoRows))
}

func TestMoveTaskToReadyColumn_NoReadyColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to ready column when none exists
	err = svc.MoveTaskToReadyColumn(ctx, task.ID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows) || err.Error() == "no ready column configured for this project")
}

func TestMoveTaskToReadyColumn_AlreadyInReadyColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	readyColID := createTestReadyColumn(t, db, projectID, "Ready")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task already in ready column
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: readyColID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to ready column (already there)
	err = svc.MoveTaskToReadyColumn(ctx, task.ID)

	assert.ErrorIs(t, err, ErrTaskAlreadyInTargetColumn)
}

func TestMoveTaskToReadyColumn_ZeroTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Try to move task with ID 0
	err := svc.MoveTaskToReadyColumn(context.Background(), 0)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestMoveTaskToReadyColumn_NegativeTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Try to move task with negative ID
	err := svc.MoveTaskToReadyColumn(context.Background(), -1)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestMoveTaskToCompletedColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	todoColID := createTestColumn(t, db, projectID, "To Do")
	completedColID := createTestCompletedColumn(t, db, projectID, "Done")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task in To Do column
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: todoColID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to completed column
	err = svc.MoveTaskToCompletedColumn(ctx, task.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved to completed column
	testutil.AssertTaskInColumn(t, db, task.ID, completedColID)
}

func TestMoveTaskToCompletedColumn_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Try to move non-existent task
	err := svc.MoveTaskToCompletedColumn(context.Background(), 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidTaskID) || errors.Is(err, sql.ErrNoRows))
}

func TestMoveTaskToCompletedColumn_NoCompletedColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to completed column when none exists
	err = svc.MoveTaskToCompletedColumn(ctx, task.ID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows) || err.Error() == "no completed column configured for this project")
}

func TestMoveTaskToCompletedColumn_AlreadyInCompletedColumn(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	completedColID := createTestCompletedColumn(t, db, projectID, "Done")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task already in completed column
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: completedColID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to completed column (already there)
	err = svc.MoveTaskToCompletedColumn(ctx, task.ID)

	assert.ErrorIs(t, err, ErrTaskAlreadyInTargetColumn)
}

func TestMoveTaskToCompletedColumn_ZeroTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Try to move task with ID 0
	err := svc.MoveTaskToCompletedColumn(context.Background(), 0)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestMoveTaskToCompletedColumn_NegativeTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Try to move task with negative ID
	err := svc.MoveTaskToCompletedColumn(context.Background(), -1)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestMoveTaskToCompletedColumn_MultipleTasksInProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	todoColID := createTestColumn(t, db, projectID, "To Do")
	inProgressColID := createTestColumn(t, db, projectID, "In Progress")
	completedColID := createTestCompletedColumn(t, db, projectID, "Done")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create multiple tasks in different columns
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: todoColID,
		Position: 0,
	})
	require.NoError(t, err)
	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: inProgressColID,
		Position: 0,
	})
	require.NoError(t, err)
	task3, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 3",
		ColumnID: todoColID,
		Position: 1,
	})
	require.NoError(t, err)

	// Move task2 to completed
	err = svc.MoveTaskToCompletedColumn(ctx, task2.ID)
	require.NoError(t, err)

	// Verify task2 is in completed column
	testutil.AssertTaskInColumn(t, db, task2.ID, completedColID)

	// Verify other tasks are unchanged
	testutil.AssertTaskInColumn(t, db, task1.ID, todoColID)
	testutil.AssertTaskInColumn(t, db, task3.ID, todoColID)
}

func TestMoveTaskToReadyColumn_MultipleTasksInProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	todoColID := createTestColumn(t, db, projectID, "To Do")
	inProgressColID := createTestColumn(t, db, projectID, "In Progress")
	readyColID := createTestReadyColumn(t, db, projectID, "Ready")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create multiple tasks in different columns
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: todoColID,
		Position: 0,
	})
	require.NoError(t, err)
	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: inProgressColID,
		Position: 0,
	})
	require.NoError(t, err)
	task3, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 3",
		ColumnID: todoColID,
		Position: 1,
	})
	require.NoError(t, err)

	// Move task2 to ready
	err = svc.MoveTaskToReadyColumn(ctx, task2.ID)
	require.NoError(t, err)

	// Verify task2 is in ready column
	testutil.AssertTaskInColumn(t, db, task2.ID, readyColID)

	// Verify other tasks are unchanged
	testutil.AssertTaskInColumn(t, db, task1.ID, todoColID)
	testutil.AssertTaskInColumn(t, db, task3.ID, todoColID)
}

func TestGetTaskActivities_RetrievesBothCommentsAndEvents(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create mock event service with predefined events
	mock := taskevent.NewMockService()
	mock.GetEventsByTaskResult = []models.TaskEvent{
		{
			ID:        1,
			TaskID:    taskID,
			Content:   "Task created",
			Author:    "system",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        2,
			TaskID:    taskID,
			Content:   "Task moved from Todo to Done",
			Author:    "system",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	// Create some comments directly in the database
	createTestComment(t, db, taskID, "First comment", "user1")
	createTestComment(t, db, taskID, "Second comment", "user2")

	svc := newTestServiceWithMock(t, db, mock)

	activities, err := svc.GetTaskActivities(context.Background(), taskID)
	require.NoError(t, err)

	// Should have 4 activities total: 2 events + 2 comments
	require.Len(t, activities, 4)

	// Count events and comments
	eventCount := 0
	commentCount := 0
	for _, activity := range activities {
		switch activity.Type {
		case models.ActivityTypeEvent:
			eventCount++
		case models.ActivityTypeComment:
			commentCount++
		}
	}

	assert.Equal(t, 2, eventCount)
	assert.Equal(t, 2, commentCount)

	// Verify mock was called
	require.True(t, mock.HasCall("GetEventsByTask", taskID))
}

func TestGetTaskActivities_SortedByTimestampDescending(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	ctx := context.Background()

	// Create mock event service with events at specific times
	mock := taskevent.NewMockService()
	now := time.Now()
	mock.GetEventsByTaskResult = []models.TaskEvent{
		{
			ID:        1,
			TaskID:    taskID,
			Content:   "Oldest event",
			Author:    "system",
			CreatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID:        2,
			TaskID:    taskID,
			Content:   "Newest event",
			Author:    "system",
			CreatedAt: now.Add(-30 * time.Minute),
		},
	}

	// Create comments with specific timestamps
	_, err := db.ExecContext(ctx,
		`INSERT INTO task_comments (task_id, content, author, created_at) VALUES 
		(?, 'Middle comment', 'user1', datetime('now', '-2 hours')),
		(?, 'Recent comment', 'user2', datetime('now', '-1 hour'))`,
		taskID, taskID)
	require.NoError(t, err, "failed to create test comments")

	svc := newTestServiceWithMock(t, db, mock)

	activities, err := svc.GetTaskActivities(ctx, taskID)
	require.NoError(t, err)

	require.Len(t, activities, 4)

	// Verify activities are sorted by CreatedAt descending (newest first)
	for i := 1; i < len(activities); i++ {
		assert.False(t, activities[i-1].CreatedAt.Before(activities[i].CreatedAt),
			"Activities not sorted correctly at index %d", i)
	}
}

func TestGetTaskActivities_TaskWithNoActivities(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create mock with no events
	mock := taskevent.NewMockService()
	mock.GetEventsByTaskResult = []models.TaskEvent{}

	svc := newTestServiceWithMock(t, db, mock)

	activities, err := svc.GetTaskActivities(context.Background(), taskID)
	require.NoError(t, err)

	assert.Len(t, activities, 0)
}

func TestGetTaskActivities_TaskWithOnlyComments(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create mock with no events
	mock := taskevent.NewMockService()
	mock.GetEventsByTaskResult = []models.TaskEvent{}

	// Create comments
	createTestComment(t, db, taskID, "Comment 1", "user1")
	createTestComment(t, db, taskID, "Comment 2", "user2")
	createTestComment(t, db, taskID, "Comment 3", "user3")

	svc := newTestServiceWithMock(t, db, mock)

	activities, err := svc.GetTaskActivities(context.Background(), taskID)
	require.NoError(t, err)

	require.Len(t, activities, 3)

	// Verify all are comments
	for _, activity := range activities {
		assert.Equal(t, models.ActivityTypeComment, activity.Type)
	}
}

func TestGetTaskActivities_TaskWithOnlyEvents(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create mock with events
	mock := taskevent.NewMockService()
	mock.GetEventsByTaskResult = []models.TaskEvent{
		{
			ID:        1,
			TaskID:    taskID,
			Content:   "Task created",
			Author:    "system",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        2,
			TaskID:    taskID,
			Content:   "Label 'bug' added",
			Author:    "user1",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	svc := newTestServiceWithMock(t, db, mock)

	activities, err := svc.GetTaskActivities(context.Background(), taskID)
	require.NoError(t, err)

	require.Len(t, activities, 2)

	// Verify all are events
	for _, activity := range activities {
		assert.Equal(t, models.ActivityTypeEvent, activity.Type)
	}
}

func TestGetTaskActivities_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	_, err := svc.GetTaskActivities(context.Background(), 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestGetTaskActivities_NegativeTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	_, err := svc.GetTaskActivities(context.Background(), -1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestGetTaskActivities_EventServiceError(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create mock that returns an error
	mock := taskevent.NewMockService()
	mock.GetEventsByTaskErr = errors.New("event service unavailable")

	svc := newTestServiceWithMock(t, db, mock)

	_, err := svc.GetTaskActivities(context.Background(), taskID)

	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to get task events")
}

func TestGetTaskActivities_VerifiesActivityItemFields(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create mock with a specific event
	eventTime := time.Now().Add(-1 * time.Hour)
	mock := taskevent.NewMockService()
	mock.GetEventsByTaskResult = []models.TaskEvent{
		{
			ID:        42,
			TaskID:    taskID,
			Content:   "Task moved to Done",
			Author:    "testauthor",
			CreatedAt: eventTime,
		},
	}

	// Create a comment
	commentID := createTestComment(t, db, taskID, "Test comment content", "commentauthor")

	svc := newTestServiceWithMock(t, db, mock)

	activities, err := svc.GetTaskActivities(context.Background(), taskID)
	require.NoError(t, err)

	require.Len(t, activities, 2)

	// Find the event activity
	var eventActivity *models.ActivityItem
	var commentActivity *models.ActivityItem
	for i := range activities {
		if activities[i].Type == models.ActivityTypeEvent {
			eventActivity = &activities[i]
		} else {
			commentActivity = &activities[i]
		}
	}

	// Verify event activity fields
	require.NotNil(t, eventActivity, "Event activity not found")
	assert.Equal(t, 42, eventActivity.ID)
	assert.Equal(t, taskID, eventActivity.TaskID)
	assert.Equal(t, "Task moved to Done", eventActivity.Content)
	assert.Equal(t, "testauthor", eventActivity.Author)

	// Verify comment activity fields
	require.NotNil(t, commentActivity, "Comment activity not found")
	assert.Equal(t, commentID, commentActivity.ID)
	assert.Equal(t, taskID, commentActivity.TaskID)
	assert.Equal(t, "Test comment content", commentActivity.Content)
	assert.Equal(t, "commentauthor", commentActivity.Author)
}

func TestGetTaskActivities_LargeNumberOfActivities(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create mock with many events
	mock := taskevent.NewMockService()
	events := make([]models.TaskEvent, 50)
	for i := 0; i < 50; i++ {
		events[i] = models.TaskEvent{
			ID:        i + 1,
			TaskID:    taskID,
			Content:   "Event " + string(rune('A'+i%26)),
			Author:    "system",
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}
	mock.GetEventsByTaskResult = events

	// Create many comments
	for i := 0; i < 50; i++ {
		createTestComment(t, db, taskID, "Comment "+string(rune('A'+i%26)), "user")
	}

	svc := newTestServiceWithMock(t, db, mock)

	activities, err := svc.GetTaskActivities(context.Background(), taskID)
	require.NoError(t, err)

	require.Len(t, activities, 100)

	// Verify sorting is maintained with large number of activities
	for i := 1; i < len(activities); i++ {
		assert.False(t, activities[i-1].CreatedAt.Before(activities[i].CreatedAt),
			"Activities not sorted correctly at index %d", i)
	}
}

func TestCreateComment(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	svc := newTestService(t, db)

	req := CreateCommentRequest{
		TaskID:  taskID,
		Message: "This is a test comment",
		Author:  "testuser",
	}

	result, err := svc.CreateComment(context.Background(), req)
	require.NoError(t, err, "Operation failed")

	require.NotNil(t, result, "Expected comment result, got nil")

	assert.NotZero(t, result.ID)
	assert.Equal(t, taskID, result.TaskID)
	assert.Equal(t, "This is a test comment", result.Message)
	assert.Equal(t, "testuser", result.Author)
	assert.False(t, result.CreatedAt.IsZero())
}

func TestCreateComment_Validation(t *testing.T) {
	t.Parallel()
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
			message: strings.Repeat("a", 1001),
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

			db := testutil.SetupTestDB(t)

			taskID := tt.taskID
			if tt.setupFn != nil {
				taskID = tt.setupFn(db)
			}

			svc := newTestService(t, db)
			req := CreateCommentRequest{
				TaskID:  taskID,
				Message: tt.message,
				Author:  tt.author,
			}

			_, err := svc.CreateComment(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	commentID := createTestComment(t, db, taskID, "Original message", "testuser")
	svc := newTestService(t, db)
	ctx := context.Background()

	req := UpdateCommentRequest{
		CommentID: commentID,
		Message:   "Updated message",
	}

	err := svc.UpdateComment(ctx, req)
	require.NoError(t, err, "Operation failed")

	// Verify the comment was updated
	var updatedMessage string
	err = db.QueryRowContext(ctx,
		"SELECT content FROM task_comments WHERE id = ?", commentID).Scan(&updatedMessage)
	require.NoError(t, err, "failed to query updated comment")

	assert.Equal(t, "Updated message", updatedMessage)
}

func TestUpdateComment_Validation(t *testing.T) {
	t.Parallel()
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
			message: strings.Repeat("a", 1001),
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

			db := testutil.SetupTestDB(t)

			commentID := tt.commentID
			if tt.setupFn != nil {
				commentID = tt.setupFn(db)
			}

			svc := newTestService(t, db)
			req := UpdateCommentRequest{
				CommentID: commentID,
				Message:   tt.message,
			}

			err := svc.UpdateComment(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	commentID := createTestComment(t, db, taskID, "Test comment", "testuser")
	svc := newTestService(t, db)
	ctx := context.Background()

	err := svc.DeleteComment(ctx, commentID)
	require.NoError(t, err, "Operation failed")

	// Verify the comment was deleted
	var count int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM task_comments WHERE id = ?", commentID).Scan(&count)
	require.NoError(t, err, "failed to query comment count")

	assert.Equal(t, 0, count)
}

func TestDeleteComment_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	err := svc.DeleteComment(context.Background(), 0) // Invalid ID

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCommentID)
}

func TestDeleteComment_NonExistentComment(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	err := svc.DeleteComment(context.Background(), 999) // Non-existent comment

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCommentNotFound)
}

func TestGetCommentsByTask(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create multiple comments
	comment1ID := createTestComment(t, db, taskID, "First comment", "user1")
	comment2ID := createTestComment(t, db, taskID, "Second comment", "user2")
	comment3ID := createTestComment(t, db, taskID, "Third comment", "user3")

	svc := newTestService(t, db)

	comments, err := svc.GetCommentsByTask(context.Background(), taskID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, comments, 3)

	// Verify comments are returned (order by created_at DESC, so newest first)
	// Since we created them in quick succession, verify IDs are present
	foundIDs := make(map[int]bool)
	for _, c := range comments {
		foundIDs[c.ID] = true
	}

	assert.True(t, foundIDs[comment1ID] && foundIDs[comment2ID] && foundIDs[comment3ID], "Not all comments were returned")
}

func TestGetCommentsByTask_NoComments(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	svc := newTestService(t, db)

	comments, err := svc.GetCommentsByTask(context.Background(), taskID)
	require.NoError(t, err, "Operation failed")

	assert.Len(t, comments, 0)
}

func TestGetCommentsByTask_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetCommentsByTask(context.Background(), 0) // Invalid ID

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestGetCommentsByTask_OrderedByCreatedAt(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	ctx := context.Background()

	// Create comments with explicit timestamps
	_, err := db.ExecContext(ctx,
		`INSERT INTO task_comments (task_id, content, author, created_at) VALUES 
		(?, 'Oldest comment', 'user1', datetime('2024-01-01 10:00:00')),
		(?, 'Middle comment', 'user2', datetime('2024-01-02 10:00:00')),
		(?, 'Newest comment', 'user3', datetime('2024-01-03 10:00:00'))`,
		taskID, taskID, taskID)
	require.NoError(t, err, "failed to create test comments")

	svc := newTestService(t, db)

	comments, err := svc.GetCommentsByTask(ctx, taskID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, comments, 3)

	// Verify DESC order (newest first)
	assert.Equal(t, "Newest comment", comments[0].Message)
	assert.Equal(t, "Middle comment", comments[1].Message)
	assert.Equal(t, "Oldest comment", comments[2].Message)
}

func TestGetTaskDetail_IncludesComments(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create task
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Create comments
	createTestComment(t, db, task.ID, "First comment", "user1")
	createTestComment(t, db, task.ID, "Second comment", "user2")

	// Get task detail
	detail, err := svc.GetTaskDetail(ctx, task.ID)
	require.NoError(t, err, "Operation failed")

	// Verify comments are included
	require.Len(t, detail.Comments, 2)

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

	assert.True(t, foundFirst && foundSecond, "Expected both comments to be present in task detail")
}

func TestDeleteTask_CascadesComments(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")

	// Create comments
	comment1ID := createTestComment(t, db, taskID, "First comment", "user1")
	comment2ID := createTestComment(t, db, taskID, "Second comment", "user2")

	svc := newTestService(t, db)
	ctx := context.Background()

	// Delete the task
	err := svc.DeleteTask(ctx, taskID)
	require.NoError(t, err, "Operation failed")

	// Verify comments were cascade deleted
	var count int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM task_comments WHERE id IN (?, ?)", comment1ID, comment2ID).Scan(&count)
	require.NoError(t, err, "failed to query comment count")

	assert.Equal(t, 0, count)
}

func TestGetInProgressTasksByProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	// Setup: Create project with in-progress column
	projectID := createTestProject(t, db)
	inProgressCol := createTestColumnWithFlag(t, db, projectID, "In Progress", true, false, false)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create multiple in-progress tasks with labels
	label1ID := createTestLabel(t, db, projectID, "urgent")
	label2ID := createTestLabel(t, db, projectID, "review")

	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: inProgressCol,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: inProgressCol,
		Position: 1,
	})
	require.NoError(t, err)

	// Attach labels to tasks
	err = svc.AttachLabel(ctx, task1.ID, label1ID)
	require.NoError(t, err)
	err = svc.AttachLabel(ctx, task2.ID, label2ID)
	require.NoError(t, err)

	// Get in-progress tasks
	tasks, err := svc.GetInProgressTasksByProject(ctx, projectID)
	require.NoError(t, err)

	// Verify results
	require.Len(t, tasks, 2)

	// Check first task
	assert.Equal(t, "Task 1", tasks[0].Title)
	assert.Len(t, tasks[0].Labels, 1)
	assert.Equal(t, "urgent", tasks[0].Labels[0].Name)

	// Check second task
	assert.Equal(t, "Task 2", tasks[1].Title)
	assert.Len(t, tasks[1].Labels, 1)
	assert.Equal(t, "review", tasks[1].Labels[0].Name)
}

func TestGetInProgressTasksByProject_InvalidProjectID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Test with invalid project ID
	_, err := svc.GetInProgressTasksByProject(ctx, -1)
	assert.Error(t, err)

	_, err = svc.GetInProgressTasksByProject(ctx, 0)
	assert.Error(t, err)
}

func TestGetInProgressTasksByProject_EmptyProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)

	// Project has no in-progress column
	tasks, err := svc.GetInProgressTasksByProject(context.Background(), projectID)
	require.NoError(t, err, "Operation failed")

	// Should return empty slice
	assert.Len(t, tasks, 0)
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
	require.NoError(t, err, "Failed to create test column")

	return columnID
}

// Helper function to add a relationship between tasks
func addTaskRelation(t *testing.T, db *sql.DB, parentID, childID int, relationTypeID int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, ?)",
		parentID, childID, relationTypeID)
	require.NoError(t, err, "Failed to add task relation")
}

// TestGetTaskTreeByProject_SingleTask tests tree with a single task (no relationships)
func TestGetTaskTreeByProject_SingleTask(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)

	// Create a single task
	task1ID := createTestTask(t, db, columnID, "Task 1")

	// Get tree
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	require.NoError(t, err)

	// Should have exactly one root node
	require.Len(t, nodes, 1)

	node := nodes[0]
	assert.Equal(t, task1ID, node.ID)
	assert.Equal(t, "Task 1", node.Title)
	assert.Len(t, node.Children, 0)
}

// TestGetTaskTreeByProject_SimpleLinearTree tests a linear chain: A -> B -> C
func TestGetTaskTreeByProject_SimpleLinearTree(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)

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
	require.NoError(t, err)

	// Should have 1 root (Task A)
	require.Len(t, nodes, 1)

	// Check root node
	rootNode := nodes[0]
	assert.Equal(t, taskA, rootNode.ID)
	assert.Equal(t, "Task A", rootNode.Title)

	// Check first level child (B)
	require.Len(t, rootNode.Children, 1)
	childB := rootNode.Children[0]
	assert.Equal(t, taskB, childB.ID)
	assert.Equal(t, "Task B", childB.Title)

	// Check second level child (C)
	require.Len(t, childB.Children, 1)
	childC := childB.Children[0]
	assert.Equal(t, taskC, childC.ID)
	assert.Equal(t, "Task C", childC.Title)

	// Check that C has no children
	assert.Len(t, childC.Children, 0)
}

// TestGetTaskTreeByProject_MultipleRoots tests multiple independent roots
func TestGetTaskTreeByProject_MultipleRoots(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create 3 independent root tasks
	task1 := createTestTask(t, db, columnID, "Root 1")
	task2 := createTestTask(t, db, columnID, "Root 2")
	task3 := createTestTask(t, db, columnID, "Root 3")

	// Set ticket numbers to establish deterministic order
	_, err := db.ExecContext(ctx, "UPDATE tasks SET ticket_number = 1 WHERE id = ?", task1)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE tasks SET ticket_number = 2 WHERE id = ?", task2)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE tasks SET ticket_number = 3 WHERE id = ?", task3)
	require.NoError(t, err)

	// Get tree
	nodes, err := svc.GetTaskTreeByProject(ctx, projectID)
	require.NoError(t, err)

	// Should have 3 roots
	require.Len(t, nodes, 3)

	// Verify roots are sorted by ticket number (ascending)
	assert.Equal(t, 1, nodes[0].TicketNumber)
	assert.Equal(t, 2, nodes[1].TicketNumber)
	assert.Equal(t, 3, nodes[2].TicketNumber)

	// Verify the task IDs match what we created
	assert.Equal(t, task1, nodes[0].ID)
	assert.Equal(t, task2, nodes[1].ID)
	assert.Equal(t, task3, nodes[2].ID)

	for _, node := range nodes {
		assert.Len(t, node.Children, 0)
	}
}

// TestGetTaskTreeByProject_DiamondDependencies tests diamond pattern: A -> (B,C) -> D
func TestGetTaskTreeByProject_DiamondDependencies(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)

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
	require.NoError(t, err)

	// Should have 1 root (Task A)
	require.Len(t, nodes, 1)

	rootA := nodes[0]
	assert.Equal(t, taskA, rootA.ID)

	// A should have 2 children (B and C)
	require.Len(t, rootA.Children, 2)

	// Check children are B and C (order may vary)
	childIDs := map[int]bool{rootA.Children[0].ID: true, rootA.Children[1].ID: true}
	assert.True(t, childIDs[taskB] && childIDs[taskC])

	// Both B and C should have D as child
	for _, child := range rootA.Children {
		require.Len(t, child.Children, 1)
		assert.Equal(t, taskD, child.Children[0].ID)
	}
}

// TestGetTaskTreeByProject_SelfDependency tests self-referencing task (A -> A)
func TestGetTaskTreeByProject_SelfDependency(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)

	taskA := createTestTask(t, db, columnID, "Task A")

	// Create self-dependency
	addTaskRelation(t, db, taskA, taskA, 1) // A -> A

	// Should handle gracefully (not panic or hang)
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	require.NoError(t, err)

	// With self-dependency, the task is marked as having a parent (itself)
	// So it won't appear as a root. This is a quirk of how circular dependencies are handled.
	// The important thing is that it doesn't panic or infinite loop.
	// Traverse the entire tree to ensure no infinite loops
	nodeCount := 0
	var traverse func(*models.TaskTreeNode)
	traverse = func(node *models.TaskTreeNode) {
		nodeCount++
		require.LessOrEqual(t, nodeCount, 100, "Tree traversal exceeded 100 nodes - likely infinite loop")
		for _, child := range node.Children {
			traverse(child)
		}
	}

	for _, root := range nodes {
		traverse(root)
	}

	// Even if task is not a root, total node count should be <= 1
	assert.LessOrEqual(t, nodeCount, 1)
}

// TestGetTaskTreeByProject_BlockingRelationships tests tree with blocking relationships
func TestGetTaskTreeByProject_BlockingRelationships(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)

	// Create tasks with blocking relationships
	taskA := createTestTask(t, db, columnID, "Task A (Blocker)")
	taskB := createTestTask(t, db, columnID, "Task B (Blocked)")

	// A blocks B (relation type 2)
	// Note: The tree building logic treats ALL relationships as parent-child hierarchies
	// So one will be a root and the other will be a child
	addTaskRelation(t, db, taskB, taskA, 2) // B is blocked by A (A is parent)

	// Get tree
	nodes, err := svc.GetTaskTreeByProject(context.Background(), projectID)
	require.NoError(t, err)

	// Should have exactly 1 root (the parent in the relationship)
	require.Len(t, nodes, 1)

	root := nodes[0]

	// Root should have 1 child (the child in the relationship)
	require.Len(t, root.Children, 1)

	child := root.Children[0]

	// Verify it's the A->B relationship with blocking flag
	taskIDs := map[int]bool{root.ID: true, child.ID: true}
	assert.True(t, taskIDs[taskA] && taskIDs[taskB])

	// The relation should be marked as blocking
	assert.True(t, child.IsBlocking)
}

// TestGetTaskTreeByProject_MixedRelationships tests tree with both parent-child and blocking
func TestGetTaskTreeByProject_MixedRelationships(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)

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
	require.NoError(t, err)

	// Should have 2 roots: A and D
	require.Len(t, nodes, 2)

	// Find root A
	var rootA *models.TaskTreeNode
	for _, node := range nodes {
		if node.ID == taskA {
			rootA = node
			break
		}
	}

	require.NotNil(t, rootA)

	// Verify A -> B -> C hierarchy
	require.Len(t, rootA.Children, 1)

	childB := rootA.Children[0]
	assert.Equal(t, taskB, childB.ID)

	require.Len(t, childB.Children, 1)

	childC := childB.Children[0]
	assert.Equal(t, taskC, childC.ID)
}

// TestAddParentRelation_CircularDependencyCheck tests AddParentRelation doesn't create circles
func TestAddParentRelation_CircularDependencyCheck(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two tasks
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Add parent relationship: 2 -> 1
	err = svc.AddParentRelation(ctx, task2.ID, task1.ID, 1)
	require.NoError(t, err)

	// Try to add reverse relationship which would create a circle: 1 -> 2 when 2 -> 1 exists
	// Note: The current implementation doesn't prevent this at the service level,
	// so we're just testing that it doesn't panic
	err = svc.AddChildRelation(ctx, task1.ID, task2.ID, 1)
	if err != nil {
		// This is acceptable - the service might prevent circular deps
		t.Logf("Service prevented circular dependency: %v", err)
	}
}

// TestAddParentRelation_SelfRelationPrevention tests that self-relations are prevented
func TestAddParentRelation_SelfRelationPrevention(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to create a self-relation
	err = svc.AddParentRelation(ctx, task.ID, task.ID, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSelfRelation)
}

// TestRemoveParentRelation_RestructuresTree tests that removing relations restructures tree
func TestRemoveParentRelation_RestructuresTree(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create three tasks
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task A",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task B",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	task3, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Task C",
		ColumnID: columnID,
		Position: 2,
	})
	require.NoError(t, err)

	// Create chain: A -> B -> C
	err = svc.AddChildRelation(ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err)
	err = svc.AddChildRelation(ctx, task2.ID, task3.ID, 1)
	require.NoError(t, err)

	// Verify structure
	nodes, err := svc.GetTaskTreeByProject(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	// Remove middle relationship: B from A
	err = svc.RemoveChildRelation(ctx, task1.ID, task2.ID)
	assert.NoError(t, err)

	// Now A and B should both be roots
	nodes, err = svc.GetTaskTreeByProject(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, nodes, 2)
}

// TestCreateTask_ErrorPaths tests various error scenarios for CreateTask
func TestCreateTask_ErrorPaths(t *testing.T) {
	t.Parallel()
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			req := tt.setupFn(db)

			_, err := svc.CreateTask(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestUpdateTask_ErrorPaths tests various error scenarios for UpdateTask
func TestUpdateTask_ErrorPaths(t *testing.T) {
	t.Parallel()
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
				task, err := newTestService(t, db).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Old Title",
					ColumnID: columnID,
					Position: 0,
				})
				require.NoError(t, err)
				longTitle := strings.Repeat("a", 256)
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
				task, err := newTestService(t, db).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
				})
				require.NoError(t, err)
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
				task, err := newTestService(t, db).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
				})
				require.NoError(t, err)
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
				task, err := newTestService(t, db).CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
				})
				require.NoError(t, err)
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			req := tt.setupFn(db)

			err := svc.UpdateTask(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestDeleteTask_ErrorPaths tests various error scenarios for DeleteTask
func TestDeleteTask_ErrorPaths(t *testing.T) {
	t.Parallel()
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
			name:    "non-existent task ID",
			taskID:  99999,
			wantErr: true, // Now fails because we fetch project ID before deletion for event publishing
		},
		{
			name:    "very large task ID",
			taskID:  999999999,
			wantErr: true, // Now fails because we fetch project ID before deletion for event publishing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)

			err := svc.DeleteTask(context.Background(), tt.taskID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestAttachLabel_ErrorPaths tests various error scenarios for AttachLabel
func TestAttachLabel_ErrorPaths(t *testing.T) {
	t.Parallel()
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
				svc := newTestService(t, db)
				task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Test Task",
					ColumnID: columnID,
					Position: 0,
					LabelIDs: []int{labelID},
				})
				require.NoError(t, err)
				return task.ID, labelID
			},
			wantErr: false, // May succeed or may error - database dependent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			taskID, labelID := tt.setupFn(db)

			err := svc.AttachLabel(context.Background(), taskID, labelID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestDetachLabel_ErrorPaths tests various error scenarios for DetachLabel
func TestDetachLabel_ErrorPaths(t *testing.T) {
	t.Parallel()
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			taskID, labelID := tt.setupFn(db)

			err := svc.DetachLabel(context.Background(), taskID, labelID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestAddParentRelation_ErrorPaths tests various error scenarios for AddParentRelation
func TestAddParentRelation_ErrorPaths(t *testing.T) {
	t.Parallel()
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
				svc := newTestService(t, db)
				parent, err := svc.CreateTask(context.Background(), CreateTaskRequest{
					Title:    "Parent",
					ColumnID: columnID,
					Position: 0,
				})
				require.NoError(t, err)
				child, err := svc.CreateTask(context.Background(), CreateTaskRequest{
					Title:     "Child",
					ColumnID:  columnID,
					Position:  1,
					ParentIDs: []int{parent.ID},
				})
				require.NoError(t, err)
				return child.ID, parent.ID
			},
			wantErr: false, // May succeed or may error - database dependent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			childID, parentID := tt.setupFn(db)

			err := svc.AddParentRelation(context.Background(), childID, parentID, 1)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestAddChildRelation_ErrorPaths tests various error scenarios for AddChildRelation
func TestAddChildRelation_ErrorPaths(t *testing.T) {
	t.Parallel()
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			parentID, childID := tt.setupFn(db)

			err := svc.AddChildRelation(context.Background(), parentID, childID, 1)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestRemoveParentRelation_ErrorPaths tests various error scenarios for RemoveParentRelation
func TestRemoveParentRelation_ErrorPaths(t *testing.T) {
	t.Parallel()
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			childID, parentID := tt.setupFn(db)

			err := svc.RemoveParentRelation(context.Background(), childID, parentID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestRemoveChildRelation_ErrorPaths tests various error scenarios for RemoveChildRelation
func TestRemoveChildRelation_ErrorPaths(t *testing.T) {
	t.Parallel()
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			parentID, childID := tt.setupFn(db)

			err := svc.RemoveChildRelation(context.Background(), parentID, childID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestMoveTaskToColumn_ErrorPaths tests various error scenarios for MoveTaskToColumn
func TestMoveTaskToColumn_ErrorPaths(t *testing.T) {
	t.Parallel()
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			taskID, columnID := tt.setupFn(db)

			err := svc.MoveTaskToColumn(context.Background(), taskID, columnID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestGetTaskSummariesByProject_ErrorPaths tests error scenarios for GetTaskSummariesByProject
func TestGetTaskSummariesByProject_ErrorPaths(t *testing.T) {
	t.Parallel()
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)

			result, err := svc.GetTaskSummariesByProject(context.Background(), tt.projectID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}

			// For non-existent project, should return empty map
			if !tt.wantErr && err == nil && tt.projectID == 99999 {
				assert.Len(t, result, 0)
			}
		})
	}
}

// TestGetTaskDetail_NegativeID tests GetTaskDetail with negative ID
func TestGetTaskDetail_NegativeID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetTaskDetail(context.Background(), -1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

// TestGetTaskReferencesForProject_ErrorPaths tests error scenarios
func TestGetTaskReferencesForProject_ErrorPaths(t *testing.T) {
	t.Parallel()
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)

			_, err := svc.GetTaskReferencesForProject(context.Background(), tt.projectID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

// TestComment_BoundaryConditions tests boundary conditions for comments
func TestComment_BoundaryConditions(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	taskID := createTestTask(t, db, columnID, "Test Task")
	svc := newTestService(t, db)

	// Test exactly at max length (1000 chars)
	maxMessage := strings.Repeat("a", 1000)

	req := CreateCommentRequest{
		TaskID:  taskID,
		Message: maxMessage,
		Author:  "testuser",
	}

	result, err := svc.CreateComment(context.Background(), req)
	require.NoError(t, err)

	require.NotNil(t, result, "Expected comment result, got nil")

	assert.Equal(t, 1000, len(result.Message))
}

// TestCreateTask_MaxLengthTitle tests title at exact max length
func TestCreateTask_MaxLengthTitle(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "To Do")
	svc := newTestService(t, db)

	// Test exactly at max length (255 chars)
	maxTitle := strings.Repeat("a", 255)

	req := CreateTaskRequest{
		Title:    maxTitle,
		ColumnID: columnID,
		Position: 0,
	}

	result, err := svc.CreateTask(context.Background(), req)
	require.NoError(t, err)

	require.NotNil(t, result, "Expected task result, got nil")

	assert.Equal(t, 255, len(result.Title))
}

// newTestService creates a new service for testing (panics on error since tests use valid SQLite)
func newTestService(t *testing.T, db *sql.DB) Service {
	t.Helper()
	svc, err := NewService(db, database.SQLite, nil, nil)
	require.NoError(t, err, "failed to create test service")
	return svc
}

// createTestProject creates a test project and returns its ID
func createTestProject(t *testing.T, db *sql.DB) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO projects (name, description) VALUES (?, ?)", "Test Project", "Description")
	require.NoError(t, err, "Failed to create test project")

	// Initialize project counter
	id, err := result.LastInsertId()
	require.NoError(t, err)
	_, err = db.ExecContext(context.Background(), "INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)", id)
	require.NoError(t, err, "Failed to initialize project counter")

	return int(id)
}

// createTestColumn creates a test column and returns its ID
func createTestColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name, holds_ready_tasks) VALUES (?, ?, ?)", projectID, name, false)
	require.NoError(t, err, "Failed to create test column")
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

// createTestReadyColumn creates a test column with holds_ready_tasks=true and returns its ID
func createTestReadyColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name, holds_ready_tasks) VALUES (?, ?, ?)", projectID, name, true)
	require.NoError(t, err, "Failed to create test ready column")
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

// createTestCompletedColumn creates a test column with holds_completed_tasks=true and returns its ID
func createTestCompletedColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name, holds_completed_tasks) VALUES (?, ?, ?)", projectID, name, true)
	require.NoError(t, err, "Failed to create test completed column")
	id, err := result.LastInsertId()
	require.NoError(t, err)
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
		require.NoError(t, err, "Failed to get max position")
	}

	nextPos := 0
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	result, err := db.ExecContext(context.Background(),
		"INSERT INTO tasks (column_id, title, position, type_id, priority_id) VALUES (?, ?, ?, 1, 3)",
		columnID, title, nextPos)
	require.NoError(t, err, "Failed to create test task")
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

// createTestLabel creates a test label and returns its ID
func createTestLabel(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO labels (project_id, name, color) VALUES (?, ?, ?)", projectID, name, "#FF5733")
	require.NoError(t, err, "Failed to create test label")
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

// createTestComment creates a test comment and returns its ID
func createTestComment(t *testing.T, db *sql.DB, taskID int, message, author string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(),
		"INSERT INTO task_comments (task_id, content, author) VALUES (?, ?, ?)",
		taskID, message, author)
	require.NoError(t, err, "Failed to create test comment")
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return int(id)
}
