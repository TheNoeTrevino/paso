package task

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestCreateComment(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		req := CreateCommentRequest{
			TaskID:  taskID,
			Message: "This is a test comment",
			Author:  "testuser",
		}

		result, err := env.Svc.CreateComment(env.Ctx, req)
		require.NoError(t, err, "Operation failed")

		require.NotNil(t, result, "Expected comment result, got nil")

		assert.NotZero(t, result.ID)
		assert.Equal(t, taskID, result.TaskID)
		assert.Equal(t, "This is a test comment", result.Message)
		assert.Equal(t, "testuser", result.Author)
		assert.False(t, result.CreatedAt.IsZero())
	})
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
		setupFn func(*sql.DB, fixtures.Dialect) int // Returns task ID if needed
	}{
		{
			name:    "empty message",
			message: "",
			author:  "testuser",
			wantErr: true,
			errType: ErrEmptyCommentMessage,
			setupFn: func(db *sql.DB, d fixtures.Dialect) int {
				projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, d, projectID, "To Do")
				return fixtures.CreateTestTask(t, db, d, columnID, "Test Task")
			},
		},
		{
			name: "message too long",
			setupFn: func(db *sql.DB, d fixtures.Dialect) int {
				projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, d, projectID, "To Do")
				return fixtures.CreateTestTask(t, db, d, columnID, "Test Task")
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
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				env := setupTestEnv(t, db, d, dbType)
				taskID := tt.taskID
				if tt.setupFn != nil {
					taskID = tt.setupFn(env.DB, d)
				}
				req := CreateCommentRequest{
					TaskID:  taskID,
					Message: tt.message,
					Author:  tt.author,
				}

				_, err := env.Svc.CreateComment(env.Ctx, req)

				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}

				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			})
		})
	}
}

func TestUpdateComment(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		commentID := fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Original message", "testuser")
		req := UpdateCommentRequest{
			CommentID: commentID,
			Message:   "Updated message",
		}

		err := env.Svc.UpdateComment(env.Ctx, req)
		require.NoError(t, err, "Operation failed")

		// Verify the comment was updated
		var updatedMessage string
		err = env.DB.QueryRowContext(env.Ctx,
			fmt.Sprintf("SELECT content FROM task_comments WHERE id = %s", d.Placeholder(1)), commentID).Scan(&updatedMessage)
		require.NoError(t, err, "failed to query updated comment")

		assert.Equal(t, "Updated message", updatedMessage)
	})
}

func TestUpdateComment_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		commentID int
		message   string
		wantErr   bool
		errType   error
		setupFn   func(*sql.DB, fixtures.Dialect) int // Returns comment ID if needed
	}{
		{
			name:    "empty message",
			message: "",
			wantErr: true,
			errType: ErrEmptyCommentMessage,
			setupFn: func(db *sql.DB, d fixtures.Dialect) int {
				projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, d, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, d, columnID, "Test Task")
				return fixtures.CreateTestComment(t, db, d, taskID, "Original message", "testuser")
			},
		},
		{
			name: "message too long",
			setupFn: func(db *sql.DB, d fixtures.Dialect) int {
				projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, d, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, d, columnID, "Test Task")
				return fixtures.CreateTestComment(t, db, d, taskID, "Original message", "testuser")
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
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				env := setupTestEnv(t, db, d, dbType)
				commentID := tt.commentID
				if tt.setupFn != nil {
					commentID = tt.setupFn(env.DB, d)
				}
				req := UpdateCommentRequest{
					CommentID: commentID,
					Message:   tt.message,
				}

				err := env.Svc.UpdateComment(env.Ctx, req)

				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}

				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			})
		})
	}
}

func TestDeleteComment(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		commentID := fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Test comment", "testuser")
		err := env.Svc.DeleteComment(env.Ctx, commentID)
		require.NoError(t, err, "Operation failed")

		// Verify the comment was deleted
		var count int
		err = env.DB.QueryRowContext(env.Ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM task_comments WHERE id = %s", d.Placeholder(1)), commentID).Scan(&count)
		require.NoError(t, err, "failed to query comment count")

		assert.Equal(t, 0, count)
	})
}

func TestDeleteComment_InvalidID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		err := env.Svc.DeleteComment(env.Ctx, 0) // Invalid ID

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCommentID)
		_ = env
	})
}

func TestDeleteComment_NonExistentComment(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		err := env.Svc.DeleteComment(env.Ctx, 999) // Non-existent comment

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCommentNotFound)
		_ = env
	})
}

func TestGetCommentsByTask(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create multiple comments
		comment1ID := fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "First comment", "user1")
		comment2ID := fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Second comment", "user2")
		comment3ID := fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Third comment", "user3")
		comments, err := env.Svc.GetCommentsByTask(env.Ctx, taskID)
		require.NoError(t, err, "Operation failed")

		require.Len(t, comments, 3)

		// Verify comments are returned (order by created_at DESC, so newest first)
		// Since we created them in quick succession, verify IDs are present
		foundIDs := make(map[int]bool)
		for _, c := range comments {
			foundIDs[c.ID] = true
		}

		assert.True(t, foundIDs[comment1ID] && foundIDs[comment2ID] && foundIDs[comment3ID], "Not all comments were returned")
	})
}

func TestGetCommentsByTask_NoComments(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		comments, err := env.Svc.GetCommentsByTask(env.Ctx, taskID)
		require.NoError(t, err, "Operation failed")

		assert.Len(t, comments, 0)
	})
}

func TestGetCommentsByTask_InvalidTaskID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		_, err := env.Svc.GetCommentsByTask(env.Ctx, 0) // Invalid ID

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTaskID)
		_ = env
	})
}

func TestGetCommentsByTask_OrderedByCreatedAt(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create comments with explicit timestamps using Go time values
		t1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		t2 := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
		t3 := time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC)
		_, err := env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_comments (task_id, content, author, created_at) VALUES (%s, 'Oldest comment', 'user1', %s), (%s, 'Middle comment', 'user2', %s), (%s, 'Newest comment', 'user3', %s)",
				d.Placeholder(1), d.Placeholder(2), d.Placeholder(3), d.Placeholder(4), d.Placeholder(5), d.Placeholder(6)),
			taskID, t1, taskID, t2, taskID, t3)
		require.NoError(t, err, "failed to create test comments")

		comments, err := env.Svc.GetCommentsByTask(env.Ctx, taskID)
		require.NoError(t, err, "Operation failed")

		require.Len(t, comments, 3)

		// Verify DESC order (newest first)
		assert.Equal(t, "Newest comment", comments[0].Message)
		assert.Equal(t, "Middle comment", comments[1].Message)
		assert.Equal(t, "Oldest comment", comments[2].Message)
	})
}

func TestDeleteTask_CascadesComments(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
		// Create comments
		comment1ID := fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "First comment", "user1")
		comment2ID := fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Second comment", "user2")
		// Delete the task
		err := env.Svc.DeleteTask(env.Ctx, taskID)
		require.NoError(t, err, "Operation failed")

		// Verify comments were cascade deleted
		var count int
		err = env.DB.QueryRowContext(env.Ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM task_comments WHERE id IN (%s, %s)", d.Placeholder(1), d.Placeholder(2)), comment1ID, comment2ID).Scan(&count)
		require.NoError(t, err, "failed to query comment count")

		assert.Equal(t, 0, count)
	})
}
