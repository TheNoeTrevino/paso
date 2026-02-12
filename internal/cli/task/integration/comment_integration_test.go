package task_test

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestCommentTask(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	ctx := context.Background()

	// Create test project with default columns (Todo, In Progress, Done)
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	columnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("add basic comment - default output", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Test Task")

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "This is a test comment",
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")
		assert.Contains(t, output, "Comment ID:")
		assert.Contains(t, output, "This is a test comment")

		// Verify comment in database
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, "This is a test comment", taskDetail.Comments[0].Message)
	})

	t.Run("add comment with custom author", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Test Task")

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Comment from custom author",
			"--author", "john.doe",
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify author in database
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, "john.doe", taskDetail.Comments[0].Author)
		assert.Equal(t, "Comment from custom author", taskDetail.Comments[0].Message)
	})

	t.Run("add comment - JSON mode output", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "JSON Test Task")

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "JSON comment test",
			"--json",
		})

		require.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))

		comment := result["comment"].(map[string]any)
		assert.NotNil(t, comment["id"])
		assert.Equal(t, float64(taskID), comment["task_id"])
		assert.Equal(t, "JSON comment test", comment["message"])
		assert.NotEmpty(t, comment["author"])
		assert.NotEmpty(t, comment["created_at"])

		task := result["task"].(map[string]any)
		assert.Equal(t, float64(taskID), task["id"])
		assert.Equal(t, "JSON Test Task", task["title"])
		assert.NotNil(t, task["ticket_number"])
		assert.Equal(t, "Test Project", task["project"])
	})

	t.Run("add comment - quiet mode output", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Quiet Test Task")

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Quiet comment test",
			"--quiet",
		})

		require.NoError(t, err)

		// Quiet mode should only output the comment ID (numeric)
		commentIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, commentIDStr)

		// Verify comment was actually created
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, "Quiet comment test", taskDetail.Comments[0].Message)

		// Verify the comment ID matches
		commentID, err := strconv.Atoi(commentIDStr)
		require.NoError(t, err, "Failed to parse comment ID")
		assert.Equal(t, commentID, taskDetail.Comments[0].ID)
	})

	t.Run("add multiple comments to same task - sequential", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Multi-comment Task")

		// Add first comment
		cmd1 := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "First comment",
			"--quiet",
		})
		require.NoError(t, err)

		// Add second comment
		cmd2 := task.CommentCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Second comment",
			"--quiet",
		})
		require.NoError(t, err)

		// Add third comment
		cmd3 := task.CommentCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd3, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Third comment",
			"--quiet",
		})
		require.NoError(t, err)

		// Verify all comments exist
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 3)

		// Verify content of all comments
		messages := make([]string, 3)
		for i, comment := range taskDetail.Comments {
			messages[i] = comment.Message
		}
		assert.Contains(t, messages, "First comment")
		assert.Contains(t, messages, "Second comment")
		assert.Contains(t, messages, "Third comment")
	})

	t.Run("add comments to different tasks", func(t *testing.T) {
		task1ID := cli.CreateTestTask(t, db, columnID, "Task One")
		task2ID := cli.CreateTestTask(t, db, columnID, "Task Two")

		// Add comment to task 1
		cmd1 := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--id", strconv.Itoa(task1ID),
			"--message", "Comment on task 1",
			"--quiet",
		})
		require.NoError(t, err)

		// Add comment to task 2
		cmd2 := task.CommentCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--id", strconv.Itoa(task2ID),
			"--message", "Comment on task 2",
			"--quiet",
		})
		require.NoError(t, err)

		// Verify task 1 has only its comment
		task1Detail, err := app.TaskService.GetTaskDetail(ctx, task1ID)
		require.NoError(t, err)
		require.Len(t, task1Detail.Comments, 1)
		assert.Equal(t, "Comment on task 1", task1Detail.Comments[0].Message)

		// Verify task 2 has only its comment
		task2Detail, err := app.TaskService.GetTaskDetail(ctx, task2ID)
		require.NoError(t, err)
		require.Len(t, task2Detail.Comments, 1)
		assert.Equal(t, "Comment on task 2", task2Detail.Comments[0].Message)
	})

	// Edge cases (Task 61 requirements)

	t.Run("empty comment message - rejected by service", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Empty Comment Task")

		cmd := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "",
		})
		// Cobra may reject empty required flag or service rejects empty message
		assert.Error(t, err)
	})

	t.Run("very long comment - 999 characters", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Long Comment Task")

		// Create a message with exactly 999 characters
		longMessage := strings.Repeat("a", 999)

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", longMessage,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify it was saved correctly
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, 999, len(taskDetail.Comments[0].Message))
		assert.Equal(t, longMessage, taskDetail.Comments[0].Message)
	})

	t.Run("comment at 1000 character boundary", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Boundary Test Task")

		// Create a message with exactly 1000 characters (at limit)
		boundaryMessage := strings.Repeat("b", 1000)

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", boundaryMessage,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify it was saved correctly
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, 1000, len(taskDetail.Comments[0].Message))
		assert.Equal(t, boundaryMessage, taskDetail.Comments[0].Message)
	})

	t.Run("comment with special characters - emoji", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Emoji Task")

		message := "Great work! 🎉 ✅ 🚀 💯 ⭐"

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify emoji preserved
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("comment with Unicode - Chinese characters", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Chinese Task")

		message := "这是一个测试评论 (This is a test comment)"

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify Chinese characters preserved
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("comment with Unicode - Arabic characters", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Arabic Task")

		message := "هذا تعليق اختباري (This is a test comment)"

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify Arabic characters preserved
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("comment with newlines and special formatting", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Newline Task")

		message := "Line 1\nLine 2\nLine 3\n\nDouble newline above"

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify newlines preserved
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
		assert.Contains(t, taskDetail.Comments[0].Message, "\n")
	})

	t.Run("comment with single quotes", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Quotes Task")

		message := "It's a test with 'quoted' words"

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify single quotes preserved
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("comment with double quotes", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Double Quotes Task")

		message := `This is a "quoted" message`

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify double quotes preserved
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("comment with backslashes and escape sequences", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Backslash Task")

		message := `Path: C:\Users\test\file.txt and escaped: \n \t \r`

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify backslashes preserved
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("comment with only whitespace", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Whitespace Task")

		message := "   \t   \n   "

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify whitespace preserved
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("comment with HTML/XML-like content", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "HTML Task")

		message := "<div>Test HTML content</div> <script>alert('test')</script>"

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify HTML-like content preserved as plain text (no sanitization)
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("comment with SQL-like content", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "SQL Task")

		message := "SELECT * FROM tasks WHERE id = 1; DROP TABLE tasks; --"

		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify SQL-like content stored as plain text (no injection)
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)

		// Verify no SQL injection occurred - tasks table still exists
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks").Scan(&count)
		require.NoError(t, err)
		assert.Greater(t, count, 0, "Tasks table should still exist and have data")
	})

	t.Run("verify created_at timestamp is set", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Timestamp Task")

		cmd := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Test timestamp",
			"--quiet",
		})
		require.NoError(t, err)

		// Verify timestamp via direct database query
		// Note: GetTaskDetail may not populate timestamps from sql.NullTime correctly
		var createdAt, updatedAt string
		err = db.QueryRowContext(ctx,
			"SELECT created_at, updated_at FROM task_comments WHERE task_id = ?",
			taskID).Scan(&createdAt, &updatedAt)
		require.NoError(t, err)
		assert.NotEmpty(t, createdAt)
		assert.NotEmpty(t, updatedAt)
	})

	t.Run("default author uses current username", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Default Author Task")

		cmd := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Test default author",
			"--quiet",
		})
		require.NoError(t, err)

		// Verify author is set (not empty)
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)

		// Author should be set to something (depends on environment)
		assert.NotEmpty(t, taskDetail.Comments[0].Author)
	})
}

func TestCommentTask_Errors(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	ctx := context.Background()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	columnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("missing --id flag", func(t *testing.T) {
		cmd := task.CommentCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--message", "Test comment",
		})

		// Should fail because --id is required
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required flag")
	})

	t.Run("missing --message flag", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Test Task")

		cmd := task.CommentCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
		})

		// Should fail because --message is required
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required flag")
	})

	t.Run("invalid task ID - non-existent", func(t *testing.T) {
		cmd := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "999999",
			"--message", "test comment on non-existent task",
		})
		cli.AssertExitError(t, err, 1) // ExitError
		assert.Contains(t, err.Error(), "failed to get task detail")
	})

	t.Run("zero task ID", func(t *testing.T) {
		cmd := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "0",
			"--message", "test comment on zero task",
		})
		cli.AssertExitError(t, err, 1) // ExitError
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("negative task ID", func(t *testing.T) {
		cmd := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "-1",
			"--message", "test comment on negative task",
		})
		// Cobra may interpret "-1" as a flag
		assert.Error(t, err)
	})

	t.Run("message exceeds 1000 characters - 1001 chars", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Long Comment 1001 Task")

		cmd := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", strings.Repeat("x", 1001),
		})
		cli.AssertExitError(t, err, 5) // ExitValidation
		assert.Contains(t, err.Error(), "message exceeds 1000 character limit")
	})

	t.Run("message exceeds 1000 characters - 1500 chars", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Long Comment 1500 Task")

		cmd := task.CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", strings.Repeat("y", 1500),
		})
		cli.AssertExitError(t, err, 5) // ExitValidation
		assert.Contains(t, err.Error(), "message exceeds 1000 character limit")
	})

	t.Run("empty author string is valid", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Empty Author Task")

		cmd := task.CommentCmd()

		// Empty author should fall back to current user
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Test with empty author",
			"--author", "",
		})

		// Should succeed and use default author
		require.NoError(t, err)
		assert.Contains(t, output, "Comment added")

		// Verify default author was used (not empty)
		taskDetail, err := app.TaskService.GetTaskDetail(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.NotEmpty(t, taskDetail.Comments[0].Author)
	})

	t.Run("both JSON and quiet flags - quiet takes precedence", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Conflicting Flags Task")

		cmd := task.CommentCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Test conflicting output flags",
			"--json",
			"--quiet",
		})

		// Should succeed - OutputFormatter handles precedence
		require.NoError(t, err)

		// Quiet mode output (just comment ID)
		commentIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, commentIDStr)
	})
}
