package task_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestAssignTask(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	ctx := context.Background()

	projectID := cli.CreateTestProject(t, db, "Test Project")

	columnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("assign user to task", func(t *testing.T) {
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, columnID, "Task to assign")

		cmd := task.AssignCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"testuser",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d assigned to @testuser", taskID))

		var assigneeID sql.NullInt64
		err = db.QueryRowContext(ctx,
			"SELECT assignee_id FROM tasks WHERE id = ?", taskID).Scan(&assigneeID)
		assert.NoError(t, err)
		assert.True(t, assigneeID.Valid, "assignee_id should be set")

		var assigneeName string
		err = db.QueryRowContext(ctx,
			"SELECT name FROM assignees WHERE id = ?", assigneeID.Int64).Scan(&assigneeName)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", assigneeName)
	})

	t.Run("assign with --quiet", func(t *testing.T) {
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, columnID, "Task for quiet assign")

		cmd := task.AssignCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"quietuser",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)
	})

	t.Run("assign with --json", func(t *testing.T) {
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, columnID, "Task for JSON assign")

		cmd := task.AssignCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"jsonuser",
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(taskID), result["task_id"])
		assert.Equal(t, "jsonuser", result["assignee"])
	})

	t.Run("clear assignee with --clear", func(t *testing.T) {
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, columnID, "Task to clear assignee")

		// First assign someone
		cmd := task.AssignCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"tempuser",
		})
		require.NoError(t, err)

		// Verify assignee is set
		var assigneeID sql.NullInt64
		err = db.QueryRowContext(ctx,
			"SELECT assignee_id FROM tasks WHERE id = ?", taskID).Scan(&assigneeID)
		require.NoError(t, err)
		require.True(t, assigneeID.Valid, "assignee should be set before clearing")

		// Now clear
		cmd = task.AssignCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--clear",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d assignee cleared", taskID))

		// Verify assignee is cleared in DB
		err = db.QueryRowContext(ctx,
			"SELECT assignee_id FROM tasks WHERE id = ?", taskID).Scan(&assigneeID)
		assert.NoError(t, err)
		assert.False(t, assigneeID.Valid, "assignee_id should be null after clearing")
	})

	t.Run("clear assignee with --json", func(t *testing.T) {
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, columnID, "Task to clear with JSON")

		// First assign someone
		cmd := task.AssignCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"tempuser2",
		})
		require.NoError(t, err)

		// Clear with JSON output
		cmd = task.AssignCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--clear",
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(taskID), result["task_id"])
		assert.Nil(t, result["assignee"])
	})

	t.Run("assign without name defaults to system user", func(t *testing.T) {
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, columnID, "Task for default assign")

		cmd := task.AssignCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		// Should contain "assigned to @<some_name>" (default from config or OS username)
		assert.Contains(t, output, fmt.Sprintf("Task %d assigned to @", taskID))

		// Verify assignee was set in DB
		var assigneeID sql.NullInt64
		err = db.QueryRowContext(ctx,
			"SELECT assignee_id FROM tasks WHERE id = ?", taskID).Scan(&assigneeID)
		assert.NoError(t, err)
		assert.True(t, assigneeID.Valid, "assignee_id should be set when using default")
	})

	t.Run("reassign task to different user", func(t *testing.T) {
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, columnID, "Task to reassign")

		// Assign to first user
		cmd := task.AssignCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"firstuser",
		})
		require.NoError(t, err)

		// Reassign to second user
		cmd = task.AssignCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"seconduser",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d assigned to @seconduser", taskID))

		// Verify in DB
		var assigneeName string
		err = db.QueryRowContext(ctx, `
			SELECT a.name FROM tasks t
			JOIN assignees a ON t.assignee_id = a.id
			WHERE t.id = ?`, taskID).Scan(&assigneeName)
		assert.NoError(t, err)
		assert.Equal(t, "seconduser", assigneeName)
	})
}

func TestAssignTask_Errors(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	projectID := cli.CreateTestProject(t, db, "Test Project")

	columnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("invalid task ID", func(t *testing.T) {
		t.Parallel()
		cmd := task.AssignCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"notanumber",
			"testuser",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("non-existent task ID", func(t *testing.T) {
		t.Parallel()
		cmd := task.AssignCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"999999",
			"testuser",
		})

		assert.Error(t, err)
		cli.AssertExitError(t, err, 3) // ExitNotFound
		assert.Contains(t, err.Error(), "task 999999 not found")
	})

	t.Run("missing task ID", func(t *testing.T) {
		t.Parallel()
		cmd := task.AssignCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})

		// cobra.RangeArgs(1, 2) should reject zero args
		assert.Error(t, err)
	})

	t.Run("too many arguments", func(t *testing.T) {
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, columnID, "Task for too many args")

		cmd := task.AssignCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"user1",
			"extra",
		})

		// cobra.RangeArgs(1, 2) should reject 3 args
		assert.Error(t, err)
	})
}
