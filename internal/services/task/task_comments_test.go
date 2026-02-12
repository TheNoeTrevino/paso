package task

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestCreateComment(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				return fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
			},
		},
		{
			name: "message too long",
			setupFn: func(db *sql.DB) int {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				return fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
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
			env := setupTestEnv(t)
			taskID := tt.taskID
			if tt.setupFn != nil {
				taskID = tt.setupFn(env.DB)
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
	}
}

func TestUpdateComment(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
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
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return fixtures.CreateTestComment(t, db, testDialect, taskID, "Original message", "testuser")
			},
		},
		{
			name: "message too long",
			setupFn: func(db *sql.DB) int {
				projectID := fixtures.CreateBareProject(t, db, testDialect, "Test Project")
				columnID := fixtures.CreateTestColumn(t, db, testDialect, projectID, "To Do")
				taskID := fixtures.CreateTestTask(t, db, testDialect, columnID, "Test Task")
				return fixtures.CreateTestComment(t, db, testDialect, taskID, "Original message", "testuser")
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
			env := setupTestEnv(t)
			commentID := tt.commentID
			if tt.setupFn != nil {
				commentID = tt.setupFn(env.DB)
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
	}
}

func TestDeleteComment(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
	commentID := fixtures.CreateTestComment(t, env.DB, env.Dialect, taskID, "Test comment", "testuser")
	err := env.Svc.DeleteComment(env.Ctx, commentID)
	require.NoError(t, err, "Operation failed")

	// Verify the comment was deleted
	var count int
	err = env.DB.QueryRowContext(env.Ctx,
		"SELECT COUNT(*) FROM task_comments WHERE id = ?", commentID).Scan(&count)
	require.NoError(t, err, "failed to query comment count")

	assert.Equal(t, 0, count)
}

func TestDeleteComment_InvalidID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	err := env.Svc.DeleteComment(env.Ctx, 0) // Invalid ID

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCommentID)
}

func TestDeleteComment_NonExistentComment(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	err := env.Svc.DeleteComment(env.Ctx, 999) // Non-existent comment

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCommentNotFound)
}

func TestGetCommentsByTask(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
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
}

func TestGetCommentsByTask_NoComments(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
	comments, err := env.Svc.GetCommentsByTask(env.Ctx, taskID)
	require.NoError(t, err, "Operation failed")

	assert.Len(t, comments, 0)
}

func TestGetCommentsByTask_InvalidTaskID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	_, err := env.Svc.GetCommentsByTask(env.Ctx, 0) // Invalid ID

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestGetCommentsByTask_OrderedByCreatedAt(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
	// Create comments with explicit timestamps
	_, err := env.DB.ExecContext(env.Ctx,
		`INSERT INTO task_comments (task_id, content, author, created_at) VALUES 
		(?, 'Oldest comment', 'user1', datetime('2024-01-01 10:00:00')),
		(?, 'Middle comment', 'user2', datetime('2024-01-02 10:00:00')),
		(?, 'Newest comment', 'user3', datetime('2024-01-03 10:00:00'))`,
		taskID, taskID, taskID)
	require.NoError(t, err, "failed to create test comments")

	comments, err := env.Svc.GetCommentsByTask(env.Ctx, taskID)
	require.NoError(t, err, "Operation failed")

	require.Len(t, comments, 3)

	// Verify DESC order (newest first)
	assert.Equal(t, "Newest comment", comments[0].Message)
	assert.Equal(t, "Middle comment", comments[1].Message)
	assert.Equal(t, "Oldest comment", comments[2].Message)
}

func TestDeleteTask_CascadesComments(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
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
		"SELECT COUNT(*) FROM task_comments WHERE id IN (?, ?)", comment1ID, comment2ID).Scan(&count)
	require.NoError(t, err, "failed to query comment count")

	assert.Equal(t, 0, count)
}
