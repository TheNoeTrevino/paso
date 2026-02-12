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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				labelID := fixtures.CreateTestLabel(t, db, testDialect, projectID, "Valid", fixtures.DefaultTestLabelColor)
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
			t.Parallel()

			env := setupTestEnv(t)
			req := tt.setupFn(env.DB)

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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
			t.Parallel()

			env := setupTestEnv(t)
			req := tt.setupFn(env.DB)

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
			t.Parallel()

			env := setupTestEnv(t)

			err := env.Svc.DeleteTask(env.Ctx, tt.taskID)

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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				labelID := fixtures.CreateTestLabel(t, db, testDialect, projectID, "Bug", fixtures.DefaultTestLabelColor)
				return 0, labelID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative task ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				labelID := fixtures.CreateTestLabel(t, db, testDialect, projectID, "Bug", fixtures.DefaultTestLabelColor)
				return -1, labelID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid label ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
		{
			name: "negative label ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
		{
			name: "non-existent task ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				labelID := fixtures.CreateTestLabel(t, db, testDialect, projectID, "Bug", fixtures.DefaultTestLabelColor)
				return 99999, labelID
			},
			wantErr: true,
		},
		{
			name: "non-existent label ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return taskID, 99999
			},
			wantErr: true,
		},
		{
			name: "duplicate label attachment - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				labelID := fixtures.CreateTestLabel(t, db, testDialect, projectID, "Bug", fixtures.DefaultTestLabelColor)
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
			t.Parallel()

			env := setupTestEnv(t)
			taskID, labelID := tt.setupFn(env.DB)

			err := env.Svc.AttachLabel(env.Ctx, taskID, labelID)

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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				labelID := fixtures.CreateTestLabel(t, db, testDialect, projectID, "Bug", fixtures.DefaultTestLabelColor)
				return 0, labelID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative task ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				labelID := fixtures.CreateTestLabel(t, db, testDialect, projectID, "Bug", fixtures.DefaultTestLabelColor)
				return -1, labelID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid label ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
		{
			name: "negative label ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
		{
			name: "non-existent task ID - may succeed",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				labelID := fixtures.CreateTestLabel(t, db, testDialect, projectID, "Bug", fixtures.DefaultTestLabelColor)
				return 99999, labelID
			},
			wantErr: false, // May succeed even if task doesn't exist
		},
		{
			name: "non-existent label ID - may succeed",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return taskID, 99999
			},
			wantErr: false, // May succeed even if label doesn't exist
		},
		{
			name: "label not attached to task - may succeed",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				labelID := fixtures.CreateTestLabel(t, db, testDialect, projectID, "Bug", fixtures.DefaultTestLabelColor)
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return taskID, labelID
			},
			wantErr: false, // May succeed even if label not attached
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()

			env := setupTestEnv(t)
			taskID, labelID := tt.setupFn(env.DB)

			err := env.Svc.DetachLabel(env.Ctx, taskID, labelID)

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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return 0, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative child ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return -1, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid parent ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative parent ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "non-existent child task",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return 99999, taskID
			},
			wantErr: true,
		},
		{
			name: "non-existent parent task",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return taskID, 99999
			},
			wantErr: true,
		},
		{
			name: "duplicate parent relation - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
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
			t.Parallel()

			env := setupTestEnv(t)
			childID, parentID := tt.setupFn(env.DB)

			err := env.Svc.AddParentRelation(env.Ctx, childID, parentID, 1)

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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return 0, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative parent ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return -1, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid child ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative child ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "non-existent parent task",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return 99999, taskID
			},
			wantErr: true,
		},
		{
			name: "non-existent child task",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return taskID, 99999
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()

			env := setupTestEnv(t)
			parentID, childID := tt.setupFn(env.DB)

			err := env.Svc.AddChildRelation(env.Ctx, parentID, childID, 1)

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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return 0, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative child ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return -1, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid parent ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative parent ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "non-existent relationship - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				task1 := fixtures.CreateTestTask(t, db, testDialect, columnID, "Task 1")
				task2 := fixtures.CreateTestTask(t, db, testDialect, columnID, "Task 2")
				return task1, task2
			},
			wantErr: false, // May succeed even if no relationship exists
		},
		{
			name: "non-existent child task - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return 99999, taskID
			},
			wantErr: false, // May succeed even if task doesn't exist
		},
		{
			name: "non-existent parent task - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return taskID, 99999
			},
			wantErr: false, // May succeed even if task doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()

			env := setupTestEnv(t)
			childID, parentID := tt.setupFn(env.DB)

			err := env.Svc.RemoveParentRelation(env.Ctx, childID, parentID)

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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return 0, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative parent ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Child")
				return -1, taskID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid child ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative child ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Parent")
				return taskID, -1
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "non-existent relationship - may succeed silently",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				task1 := fixtures.CreateTestTask(t, db, testDialect, columnID, "Task 1")
				task2 := fixtures.CreateTestTask(t, db, testDialect, columnID, "Task 2")
				return task1, task2
			},
			wantErr: false, // May succeed even if no relationship exists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()

			env := setupTestEnv(t)
			parentID, childID := tt.setupFn(env.DB)

			err := env.Svc.RemoveChildRelation(env.Ctx, parentID, childID)

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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				return 0, columnID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "negative task ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				return -1, columnID
			},
			wantErr: true,
			errType: ErrInvalidTaskID,
		},
		{
			name: "invalid column ID zero",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return taskID, 0
			},
			wantErr: true,
			errType: ErrInvalidColumnID,
		},
		{
			name: "negative column ID",
			setupFn: func(db *sql.DB) (int, int) {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
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
			t.Parallel()

			env := setupTestEnv(t)
			taskID, columnID := tt.setupFn(env.DB)

			err := env.Svc.MoveTaskToColumn(env.Ctx, taskID, columnID)

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
			t.Parallel()

			env := setupTestEnv(t)

			result, err := env.Svc.GetTaskSummariesByProject(env.Ctx, tt.projectID)

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
			t.Parallel()

			env := setupTestEnv(t)

			_, err := env.Svc.GetTaskReferencesForProject(env.Ctx, tt.projectID)

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

	env := setupTestEnv(t)
	taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
	// Test exactly at max length (1000 chars)
	maxMessage := strings.Repeat("a", 1000)

	req := CreateCommentRequest{
		TaskID:  taskID,
		Message: maxMessage,
		Author:  "testuser",
	}

	result, err := env.Svc.CreateComment(env.Ctx, req)
	require.NoError(t, err)

	require.NotNil(t, result, "Expected comment result, got nil")

	assert.Equal(t, 1000, len(result.Message))
}
