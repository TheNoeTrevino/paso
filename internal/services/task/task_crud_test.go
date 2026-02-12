package task

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestCreateTask(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)

	req := CreateTaskRequest{
		Title:       "Fix bug in login",
		Description: "Users can't log in",
		ColumnID:    env.ColumnID,
		Position:    0,
		PriorityID:  4, // high
		TypeID:      3, // bug
	}

	result, err := env.Svc.CreateTask(env.Ctx, req)
	require.NoError(t, err, "Operation failed")

	require.NotNil(t, result, "Expected task result, got nil")

	assert.Equal(t, "Fix bug in login", result.Title)
	assert.Equal(t, "Users can't log in", result.Description)
	assert.Equal(t, env.ColumnID, result.ColumnID)
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
			t.Parallel()

			env := setupTestEnv(t)
			req := tt.args.req
			// Setup database if needed
			if tt.needsTest || tt.args.setupFn != nil {
				if tt.args.setupFn != nil {
					req = tt.args.setupFn(env.DB, env.ColumnID)
				}
			}
			_, err := env.Svc.CreateTask(env.Ctx, req)

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

	env := setupTestEnv(t)
	label1ID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "Bug", fixtures.DefaultTestLabelColor)
	label2ID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "Critical", fixtures.DefaultTestLabelColor)
	req := CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		Position: 0,
		LabelIDs: []int{label1ID, label2ID},
	}

	result, err := env.Svc.CreateTask(env.Ctx, req)
	require.NoError(t, err, "Operation failed")

	// Verify labels are attached
	var count int
	err = env.DB.QueryRowContext(env.Ctx, "SELECT COUNT(*) FROM task_labels WHERE task_id = ?", result.ID).Scan(&count)
	require.NoError(t, err, "failed to query task labels")

	assert.Equal(t, 2, count)
}

func TestGetTaskDetail(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create a task
	created, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		ColumnID:    env.ColumnID,
		Position:    0,
		PriorityID:  4,
		TypeID:      3,
	})
	require.NoError(t, err)

	// Get task detail
	result, err := env.Svc.GetTaskDetail(env.Ctx, created.ID)
	require.NoError(t, err, "Operation failed")

	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, "Test Task", result.Title)
	assert.Equal(t, "Test Description", result.Description)
}

func TestGetTaskDetail_NotFound(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	_, err := env.Svc.GetTaskDetail(env.Ctx, 999)

	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestGetTaskDetail_InvalidID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	_, err := env.Svc.GetTaskDetail(env.Ctx, 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestGetTaskSummariesByProject(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	col1ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "To Do")
	col2ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "Done")
	// Create tasks in different columns
	_, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: col1ID,
		Position: 0,
	})
	require.NoError(t, err)

	_, err = env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: col2ID,
		Position: 0,
	})
	require.NoError(t, err)

	// Get summaries
	results, err := env.Svc.GetTaskSummariesByProject(env.Ctx, env.ProjectID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, results, 2)
	assert.Len(t, results[col1ID], 1)
	assert.Len(t, results[col2ID], 1)
}

func TestUpdateTask(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create a task
	created, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Old Title",
		ColumnID: env.ColumnID,
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

	err = env.Svc.UpdateTask(env.Ctx, req)
	require.NoError(t, err, "Operation failed")

	// Verify update
	updated, err := env.Svc.GetTaskDetail(env.Ctx, created.ID)
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
			t.Parallel()

			env := setupTestEnv(t)
			taskID := tt.taskID
			if tt.setupFn != nil {
				taskID = tt.setupFn(env.DB)
			}
			req := UpdateTaskRequest{
				TaskID: taskID,
				Title:  tt.title,
			}

			err := env.Svc.UpdateTask(env.Ctx, req)

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

func TestDeleteTask(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create a task
	created, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Delete task
	err = env.Svc.DeleteTask(env.Ctx, created.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task is deleted
	_, err = env.Svc.GetTaskDetail(env.Ctx, created.ID)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeleteTask_InvalidID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	err := env.Svc.DeleteTask(env.Ctx, 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestAttachLabel(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	labelID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "Bug", fixtures.DefaultTestLabelColor)
	// Create a task
	created, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Attach label
	err = env.Svc.AttachLabel(env.Ctx, created.ID, labelID)
	require.NoError(t, err, "Operation failed")

	// Verify label is attached
	var count int
	err = env.DB.QueryRowContext(env.Ctx, "SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?", created.ID, labelID).Scan(&count)
	require.NoError(t, err, "failed to query task labels")

	assert.Equal(t, 1, count)
}

func TestDetachLabel(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	labelID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "Bug", fixtures.DefaultTestLabelColor)
	// Create a task with label
	created, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		Position: 0,
		LabelIDs: []int{labelID},
	})
	require.NoError(t, err)

	// Detach label
	err = env.Svc.DetachLabel(env.Ctx, created.ID, labelID)
	require.NoError(t, err, "Operation failed")

	// Verify label is detached
	var count int
	err = env.DB.QueryRowContext(env.Ctx, "SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?", created.ID, labelID).Scan(&count)
	require.NoError(t, err, "failed to query task labels")

	assert.Equal(t, 0, count)
}

func TestGetTaskDetail_IncludesComments(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create task
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Create comments
	fixtures.CreateTestComment(t, env.DB, env.Dialect, task.ID, "First comment", "user1")
	fixtures.CreateTestComment(t, env.DB, env.Dialect, task.ID, "Second comment", "user2")

	// Get task detail
	detail, err := env.Svc.GetTaskDetail(env.Ctx, task.ID)
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

func TestGetTaskDetail_NegativeID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	_, err := env.Svc.GetTaskDetail(env.Ctx, -1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestCreateTask_MaxLengthTitle(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Test exactly at max length (255 chars)
	maxTitle := strings.Repeat("a", 255)

	req := CreateTaskRequest{
		Title:    maxTitle,
		ColumnID: env.ColumnID,
		Position: 0,
	}

	result, err := env.Svc.CreateTask(env.Ctx, req)
	require.NoError(t, err)

	require.NotNil(t, result, "Expected task result, got nil")

	assert.Equal(t, 255, len(result.Title))
}
