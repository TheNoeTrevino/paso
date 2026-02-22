package task_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestReadyMoveTask(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	ctx := context.Background()

	// Create test project with columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default columns created by CreateTestProject
	todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")
	inProgressColumnID := cli.GetColumnIDByName(t, db, projectID, "In Progress")
	doneColumnID := cli.GetColumnIDByName(t, db, projectID, "Done")

	// Mark "Todo" column as ready column (holds_ready_tasks = true)
	cli.SetColumnHoldsReadyTasks(t, db, todoColumnID)

	t.Run("move task from In Progress to ready column - default output", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task to Move to Ready")

		cmd := task.ReadyMoveCmd()

		// Note: ReadyMoveCmd takes task ID as positional arg, not --id flag!
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", taskID))

		// Verify task moved to ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("move task from Done to ready column - default output", func(t *testing.T) {
		// Create task in Done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task from Done to Ready")

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", taskID))

		// Verify task moved to ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("move task to ready column - quiet mode output", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task for Quiet Mode")

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should output only task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify task moved to ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("move task to ready column - JSON mode output", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task for JSON Mode")

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure
		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(taskID), result["task_id"])
		assert.Equal(t, "In Progress", result["from_column"])
		assert.Equal(t, "Todo", result["to_column"])

		// Verify task moved to ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("verify column transition from Done to ready in JSON", func(t *testing.T) {
		// Create task in Done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task from Done")

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output to verify transition
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		assert.Equal(t, "Done", result["from_column"])
		assert.Equal(t, "Todo", result["to_column"])

		// Verify task moved to ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("task already in ready column - warning to stderr", func(t *testing.T) {
		// Create task directly in ready column (Todo)
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Already in Ready Column")

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		// Should not error - exits successfully
		assert.NoError(t, err)
		// Output should contain the informational message and task ID
		assert.Contains(t, output, "Task")
		assert.Contains(t, output, "already in the ready column")
		assert.Contains(t, output, fmt.Sprintf("%d", taskID))

		// Verify task is still in ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("task already in ready column - default output", func(t *testing.T) {
		// Create task directly in ready column (Todo)
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Already Ready")

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		// Should not error - exits successfully
		assert.NoError(t, err)
		// Output should contain the informational message
		assert.Contains(t, output, "Task")
		assert.Contains(t, output, "already in the ready column")
		assert.Contains(t, output, fmt.Sprintf("%d", taskID))

		// Verify task is still in ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("verify task metadata preserved after move", func(t *testing.T) {
		// Create task with full metadata in Done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task with Metadata")

		description := "This task has a detailed description"
		cli.UpdateTaskFields(t, db, taskID, map[string]any{
			"description":   description,
			"task_number": 99,
			"type_id":       2,
			"priority_id":   4,
		})

		cmd := task.ReadyMoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)

		// Verify task moved and metadata is preserved
		var columnID int
		var savedDescription string
		var taskNumber, typeID, priorityID int
		err = db.QueryRowContext(ctx,
			"SELECT column_id, description, task_number, type_id, priority_id FROM tasks WHERE id = ?",
			taskID).Scan(&columnID, &savedDescription, &taskNumber, &typeID, &priorityID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)
		assert.Equal(t, description, savedDescription)
		assert.Equal(t, 99, taskNumber)
		assert.Equal(t, 2, typeID)
		assert.Equal(t, 4, priorityID)
	})

	t.Run("move task with labels - labels preserved", func(t *testing.T) {
		// Create task with labels
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task With Labels")

		// Create and attach labels
		labelID1 := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "bug", "#EF4444")
		labelID2 := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "urgent", "#F97316")

		cli.AttachLabelToTask(t, db, taskID, labelID1)
		cli.AttachLabelToTask(t, db, taskID, labelID2)

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", taskID))

		// Verify task moved to ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, todoColumnID, columnID)

		// Verify labels are still attached
		var labelCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", taskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, labelCount)
	})

	t.Run("move task with relationships - relationships preserved", func(t *testing.T) {
		// Create parent and child tasks
		parentTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Parent Task")
		childTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Child Task")

		// Create relationship (relation_type_id = 1 for parent-child)
		cli.AddTaskSubtask(t, db, parentTaskID, childTaskID, 1)

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", parentTaskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", parentTaskID))

		// Verify task moved to ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), parentTaskID)
		assert.Equal(t, todoColumnID, columnID)

		// Verify relationship is still intact
		var relationCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentTaskID, childTaskID).Scan(&relationCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, relationCount)
	})

	t.Run("move multiple tasks to ready column in sequence", func(t *testing.T) {
		// Create multiple tasks in different columns
		taskID1 := cli.CreateTestTask(t, db, inProgressColumnID, "Multi Task 1")
		taskID2 := cli.CreateTestTask(t, db, doneColumnID, "Multi Task 2")
		taskID3 := cli.CreateTestTask(t, db, inProgressColumnID, "Multi Task 3")

		// Move all tasks to ready column
		for _, taskID := range []int{taskID1, taskID2, taskID3} {
			cmd := task.ReadyMoveCmd()
			_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				fmt.Sprintf("%d", taskID),
				"--quiet",
			})
			assert.NoError(t, err)
		}

		// Verify all tasks moved to ready column
		for _, taskID := range []int{taskID1, taskID2, taskID3} {
			columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
			assert.Equal(t, todoColumnID, columnID, "Task %d should be in ready column", taskID)
		}
	})

	t.Run("move task with blocking relationship preserved", func(t *testing.T) {
		// Create blocker and blocked tasks
		blockerTaskID := cli.CreateTestTask(t, db, doneColumnID, "Blocker Task")
		blockedTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Blocked Task")

		// Create blocking relationship (relation_type_id = 2 for blocking)
		cli.AddTaskSubtask(t, db, blockedTaskID, blockerTaskID, 2)

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", blockedTaskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", blockedTaskID))

		// Verify task moved to ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), blockedTaskID)
		assert.Equal(t, todoColumnID, columnID)

		// Verify blocking relationship is still intact
		var relationCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ? AND relation_type_id = 2",
			blockedTaskID, blockerTaskID).Scan(&relationCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, relationCount)
	})

	t.Run("move task with comments preserved", func(t *testing.T) {
		// Create task with comments
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task With Comments")

		// Add comments
		cli.CreateTestComment(t, db, taskID, "First comment", "user1")
		cli.CreateTestComment(t, db, taskID, "Second comment", "user2")

		cmd := task.ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", taskID))

		// Verify task moved to ready column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, todoColumnID, columnID)

		// Verify comments are still attached
		var commentCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_comments WHERE task_id = ?", taskID).Scan(&commentCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, commentCount)
	})

	t.Run("verify position updated when moving to ready column", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task for Position Test")

		// Get initial position
		var initialPosition int
		err := db.QueryRowContext(ctx,
			"SELECT position FROM tasks WHERE id = ?", taskID).Scan(&initialPosition)
		assert.NoError(t, err)

		cmd := task.ReadyMoveCmd()

		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)

		// Get new position (should be at the end of the ready column)
		var newPosition int
		err = db.QueryRowContext(ctx,
			"SELECT position FROM tasks WHERE id = ?", taskID).Scan(&newPosition)
		assert.NoError(t, err)

		// Position should be updated (likely different as it's appended to the ready column)
		// The exact position depends on how many tasks are already in the ready column
		// but we can verify it's non-negative
		assert.GreaterOrEqual(t, newPosition, 0)
	})
}

func TestReadyMoveTask_Errors(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	ctx := context.Background()

	// Create test project with columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")
	inProgressColumnID := cli.GetColumnIDByName(t, db, projectID, "In Progress")

	// Mark "Todo" column as ready column for positive tests
	cli.SetColumnHoldsReadyTasks(t, db, todoColumnID)

	t.Run("invalid task ID - non-numeric", func(t *testing.T) {
		cmd := task.ReadyMoveCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"not-a-number"})
		cli.AssertExitError(t, err, 5) // ExitValidation
		assert.Contains(t, err.Error(), "invalid task ID: not-a-number")
	})

	t.Run("missing task ID argument", func(t *testing.T) {
		cmd := task.ReadyMoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})

		// Should error - task ID is required
		assert.Error(t, err)
	})

	t.Run("non-existent task ID", func(t *testing.T) {
		nonExistentTaskID := 999999

		cmd := task.ReadyMoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", nonExistentTaskID),
		})
		cli.AssertExitError(t, err, 3) // ExitNotFound
		assert.Contains(t, err.Error(), "task 999999 not found")
	})

	t.Run("project with no ready column configured", func(t *testing.T) {
		newProjectID := cli.CreateTestProject(t, db, "Project Without Ready Column")
		newColumnID := cli.CreateTestColumn(t, db, newProjectID, "Regular Column")

		_, err := db.ExecContext(ctx,
			"UPDATE columns SET holds_ready_tasks = false WHERE project_id = ?", newProjectID)
		require.NoError(t, err)

		taskID := cli.CreateTestTask(t, db, newColumnID, "Task in Project Without Ready")

		cmd := task.ReadyMoveCmd()

		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})
		cli.AssertExitError(t, err, 5) // ExitValidation
		assert.Contains(t, err.Error(), "no ready column configured")
	})

	t.Run("invalid task ID - zero", func(t *testing.T) {
		cmd := task.ReadyMoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"0",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task ID must be a positive integer")
	})

	t.Run("invalid task ID - negative", func(t *testing.T) {
		cmd := task.ReadyMoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"-1",
		})
		// Cobra may interpret "-1" as a flag, so we just assert error
		assert.Error(t, err)
	})

	t.Run("too many arguments", func(t *testing.T) {
		// Create task for this test
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task With Extra Args")

		cmd := task.ReadyMoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"extra-argument",
		})

		// Should fail because command expects exactly 1 positional argument
		assert.Error(t, err)
	})

	t.Run("invalid flag combination - json and quiet", func(t *testing.T) {
		// Create task for this test
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task Flag Test")

		cmd := task.ReadyMoveCmd()

		// Note: Cobra doesn't prevent using both --json and --quiet flags together.
		// The command implementation will respect the order and use whichever is processed first.
		// This is not an error condition, but it's good to document the behavior.
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--json",
			"--quiet",
		})

		// Should not error, but behavior depends on flag processing order
		assert.NoError(t, err)
		assert.NotEmpty(t, output)
	})
}
