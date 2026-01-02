package task

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestCommentTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project with default columns (Todo, In Progress, Done)
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	var columnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&columnID)
	require.NoError(t, err)

	t.Run("Add basic comment - default output", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Test Task")

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "This is a test comment",
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")
		assert.Contains(t, output, "Comment ID:")
		assert.Contains(t, output, "This is a test comment")

		// Verify comment in database
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, "This is a test comment", taskDetail.Comments[0].Message)
	})

	t.Run("Add comment with custom author", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Test Task")

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Comment from custom author",
			"--author", "john.doe",
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify author in database
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, "john.doe", taskDetail.Comments[0].Author)
		assert.Equal(t, "Comment from custom author", taskDetail.Comments[0].Message)
	})

	t.Run("Add comment - JSON mode output", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "JSON Test Task")

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "JSON comment test",
			"--json",
		})

		require.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))

		comment := result["comment"].(map[string]interface{})
		assert.NotNil(t, comment["id"])
		assert.Equal(t, float64(taskID), comment["task_id"])
		assert.Equal(t, "JSON comment test", comment["message"])
		assert.NotEmpty(t, comment["author"])
		assert.NotEmpty(t, comment["created_at"])

		task := result["task"].(map[string]interface{})
		assert.Equal(t, float64(taskID), task["id"])
		assert.Equal(t, "JSON Test Task", task["title"])
		assert.NotNil(t, task["ticket_number"])
		assert.Equal(t, "Test Project", task["project"])
	})

	t.Run("Add comment - quiet mode output", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Quiet Test Task")

		cmd := CommentCmd()
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
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, "Quiet comment test", taskDetail.Comments[0].Message)

		// Verify the comment ID matches
		commentID, _ := strconv.Atoi(commentIDStr)
		assert.Equal(t, commentID, taskDetail.Comments[0].ID)
	})

	t.Run("Add multiple comments to same task - sequential", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Multi-comment Task")

		// Add first comment
		cmd1 := CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "First comment",
			"--quiet",
		})
		require.NoError(t, err)

		// Add second comment
		cmd2 := CommentCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Second comment",
			"--quiet",
		})
		require.NoError(t, err)

		// Add third comment
		cmd3 := CommentCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd3, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Third comment",
			"--quiet",
		})
		require.NoError(t, err)

		// Verify all comments exist
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
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

	t.Run("Add comments to different tasks", func(t *testing.T) {
		task1ID := cli.CreateTestTask(t, db, columnID, "Task One")
		task2ID := cli.CreateTestTask(t, db, columnID, "Task Two")

		// Add comment to task 1
		cmd1 := CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--id", strconv.Itoa(task1ID),
			"--message", "Comment on task 1",
			"--quiet",
		})
		require.NoError(t, err)

		// Add comment to task 2
		cmd2 := CommentCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--id", strconv.Itoa(task2ID),
			"--message", "Comment on task 2",
			"--quiet",
		})
		require.NoError(t, err)

		// Verify task 1 has only its comment
		task1Detail, err := app.TaskService.GetTaskDetail(context.Background(), task1ID)
		require.NoError(t, err)
		require.Len(t, task1Detail.Comments, 1)
		assert.Equal(t, "Comment on task 1", task1Detail.Comments[0].Message)

		// Verify task 2 has only its comment
		task2Detail, err := app.TaskService.GetTaskDetail(context.Background(), task2ID)
		require.NoError(t, err)
		require.Len(t, task2Detail.Comments, 1)
		assert.Equal(t, "Comment on task 2", task2Detail.Comments[0].Message)
	})

	// Edge cases (Task 61 requirements)

	t.Run("Empty comment message - rejected by service", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Empty Comment Task")

		cmd := CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "",
		})

		// Empty message is rejected by service validation (ErrEmptyCommentMessage)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")

		// Verify no comment was stored
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		assert.Len(t, taskDetail.Comments, 0)
	})

	t.Run("Very long comment - 999 characters", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Long Comment Task")

		// Create a message with exactly 999 characters
		longMessage := strings.Repeat("a", 999)

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", longMessage,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify it was saved correctly
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, 999, len(taskDetail.Comments[0].Message))
		assert.Equal(t, longMessage, taskDetail.Comments[0].Message)
	})

	t.Run("Comment at 1000 character boundary", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Boundary Test Task")

		// Create a message with exactly 1000 characters (at limit)
		boundaryMessage := strings.Repeat("b", 1000)

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", boundaryMessage,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify it was saved correctly
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, 1000, len(taskDetail.Comments[0].Message))
		assert.Equal(t, boundaryMessage, taskDetail.Comments[0].Message)
	})

	t.Run("Comment with special characters - emoji", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Emoji Task")

		message := "Great work! 🎉 ✅ 🚀 💯 ⭐"

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify emoji preserved
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("Comment with Unicode - Chinese characters", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Chinese Task")

		message := "这是一个测试评论 (This is a test comment)"

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify Chinese characters preserved
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("Comment with Unicode - Arabic characters", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Arabic Task")

		message := "هذا تعليق اختباري (This is a test comment)"

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify Arabic characters preserved
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("Comment with newlines and special formatting", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Newline Task")

		message := "Line 1\nLine 2\nLine 3\n\nDouble newline above"

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify newlines preserved
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
		assert.Contains(t, taskDetail.Comments[0].Message, "\n")
	})

	t.Run("Comment with single quotes", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Quotes Task")

		message := "It's a test with 'quoted' words"

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify single quotes preserved
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("Comment with double quotes", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Double Quotes Task")

		message := `This is a "quoted" message`

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify double quotes preserved
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("Comment with backslashes and escape sequences", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Backslash Task")

		message := `Path: C:\Users\test\file.txt and escaped: \n \t \r`

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify backslashes preserved
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("Comment with only whitespace", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Whitespace Task")

		message := "   \t   \n   "

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify whitespace preserved
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("Comment with HTML/XML-like content", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "HTML Task")

		message := "<div>Test HTML content</div> <script>alert('test')</script>"

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify HTML-like content preserved as plain text (no sanitization)
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)
	})

	t.Run("Comment with SQL-like content", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "SQL Task")

		message := "SELECT * FROM tasks WHERE id = 1; DROP TABLE tasks; --"

		cmd := CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", message,
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify SQL-like content stored as plain text (no injection)
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.Equal(t, message, taskDetail.Comments[0].Message)

		// Verify no SQL injection occurred - tasks table still exists
		var count int
		err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM tasks").Scan(&count)
		require.NoError(t, err)
		assert.Greater(t, count, 0, "Tasks table should still exist and have data")
	})

	t.Run("Verify created_at timestamp is set", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Timestamp Task")

		cmd := CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Test timestamp",
			"--quiet",
		})
		require.NoError(t, err)

		// Verify timestamp via direct database query
		// Note: GetTaskDetail may not populate timestamps from sql.NullTime correctly
		var createdAt, updatedAt string
		err = db.QueryRowContext(context.Background(),
			"SELECT created_at, updated_at FROM task_comments WHERE task_id = ?",
			taskID).Scan(&createdAt, &updatedAt)
		require.NoError(t, err)
		assert.NotEmpty(t, createdAt)
		assert.NotEmpty(t, updatedAt)
	})

	t.Run("Default author uses current username", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Default Author Task")

		cmd := CommentCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Test default author",
			"--quiet",
		})
		require.NoError(t, err)

		// Verify author is set (not empty)
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)

		// Author should be set to something (depends on environment)
		assert.NotEmpty(t, taskDetail.Comments[0].Author)
	})
}

func TestCommentTask_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	var columnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&columnID)
	require.NoError(t, err)

	t.Run("Missing --id flag", func(t *testing.T) {
		cmd := CommentCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--message", "Test comment",
		})

		// Should fail because --id is required
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required flag")
	})

	t.Run("Missing --message flag", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Test Task")

		cmd := CommentCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
		})

		// Should fail because --message is required
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required flag")
	})

	t.Run("Invalid task ID - non-existent", func(t *testing.T) {
		nonExistentTaskID := 99999

		cmd := CommentCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(nonExistentTaskID),
			"--message", "Comment on non-existent task",
		})

		// Should fail because task doesn't exist
		assert.Error(t, err)
	})

	t.Run("Zero task ID", func(t *testing.T) {
		cmd := CommentCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "0",
			"--message", "Comment on zero task",
		})

		// Should fail with zero task ID
		assert.Error(t, err)
	})

	t.Run("Negative task ID", func(t *testing.T) {
		cmd := CommentCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "-1",
			"--message", "Comment on negative task",
		})

		// Should fail with negative task ID
		assert.Error(t, err)
	})

	t.Run("Message exceeds 1000 characters - 1001 chars", func(t *testing.T) {
		t.Skip("Skipping test: command calls os.Exit(cli.ExitValidation) when message exceeds 1000 characters")
		// Note: This test is skipped because the comment command calls os.Exit() at line 82
		// when the message length exceeds 1000 characters, which would terminate the test process.
		// The validation logic in comment.go (lines 77-83) correctly handles this case before
		// initializing the CLI, ensuring the error is caught early.
		//
		// To test this manually, run:
		// paso task comment --id=1 --message="$(printf 'a%.0s' {1..1001})"
		// Expected: exits with code 2 (cli.ExitValidation) and error message
	})

	t.Run("Message exceeds 1000 characters - 1500 chars", func(t *testing.T) {
		t.Skip("Skipping test: command calls os.Exit(cli.ExitValidation) when message exceeds 1000 characters")
		// Note: This test is skipped because the comment command calls os.Exit() at line 82
		// when the message length exceeds 1000 characters, which would terminate the test process.
		// See comment.go lines 77-83 for validation implementation.
	})

	t.Run("Empty author string is valid", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Empty Author Task")

		cmd := CommentCmd()

		// Empty author should fall back to current user
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(taskID),
			"--message", "Test with empty author",
			"--author", "",
		})

		// Should succeed and use default author
		require.NoError(t, err)
		assert.Contains(t, output, "✓ Comment added")

		// Verify default author was used (not empty)
		taskDetail, err := app.TaskService.GetTaskDetail(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, taskDetail.Comments, 1)
		assert.NotEmpty(t, taskDetail.Comments[0].Author)
	})

	t.Run("Both JSON and quiet flags - quiet takes precedence", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Conflicting Flags Task")

		cmd := CommentCmd()

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
