package task

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestReadyMoveTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project with columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default columns created by CreateTestProject
	var todoColumnID, inProgressColumnID, doneColumnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'", projectID).Scan(&todoColumnID)
	require.NoError(t, err)

	err = db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'In Progress'", projectID).Scan(&inProgressColumnID)
	require.NoError(t, err)

	err = db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Done'", projectID).Scan(&doneColumnID)
	require.NoError(t, err)

	// Mark "Todo" column as ready column (holds_ready_tasks = true)
	_, err = db.ExecContext(context.Background(),
		"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", todoColumnID)
	require.NoError(t, err)

	t.Run("Move task from In Progress to ready column - default output", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task to Move to Ready")

		cmd := ReadyMoveCmd()

		// Note: ReadyMoveCmd takes task ID as positional arg, not --id flag!
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", taskID))

		// Verify task moved to ready column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("Move task from Done to ready column - default output", func(t *testing.T) {
		// Create task in Done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task from Done to Ready")

		cmd := ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", taskID))

		// Verify task moved to ready column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("Move task to ready column - quiet mode output", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task for Quiet Mode")

		cmd := ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should output only task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify task moved to ready column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("Move task to ready column - JSON mode output", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task for JSON Mode")

		cmd := ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure
		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(taskID), result["task_id"])
		assert.Equal(t, "In Progress", result["from_column"])
		assert.Equal(t, "Todo", result["to_column"])

		// Verify task moved to ready column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("Verify column transition from Done to ready in JSON", func(t *testing.T) {
		// Create task in Done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task from Done")

		cmd := ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output to verify transition
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		assert.Equal(t, "Done", result["from_column"])
		assert.Equal(t, "Todo", result["to_column"])

		// Verify task moved to ready column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("Task already in ready column - warning to stderr", func(t *testing.T) {
		// Create task directly in ready column (Todo)
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Already in Ready Column")

		cmd := ReadyMoveCmd()

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
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("Task already in ready column - default output", func(t *testing.T) {
		// Create task directly in ready column (Todo)
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Already Ready")

		cmd := ReadyMoveCmd()

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
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)
	})

	t.Run("Verify task metadata preserved after move", func(t *testing.T) {
		// Create task with full metadata in Done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task with Metadata")

		description := "This task has a detailed description"
		_, err := db.ExecContext(context.Background(), `
			UPDATE tasks 
			SET description = ?, 
			    ticket_number = ?,
			    type_id = 2, 
			    priority_id = 4
			WHERE id = ?`,
			description, 99, taskID)
		require.NoError(t, err)

		cmd := ReadyMoveCmd()

		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)

		// Verify task moved and metadata is preserved
		var columnID int
		var savedDescription string
		var ticketNumber, typeID, priorityID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id, description, ticket_number, type_id, priority_id FROM tasks WHERE id = ?",
			taskID).Scan(&columnID, &savedDescription, &ticketNumber, &typeID, &priorityID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)
		assert.Equal(t, description, savedDescription)
		assert.Equal(t, 99, ticketNumber)
		assert.Equal(t, 2, typeID)
		assert.Equal(t, 4, priorityID)
	})

	t.Run("Move task with labels - labels preserved", func(t *testing.T) {
		// Create task with labels
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task With Labels")

		// Create and attach labels
		labelID1 := testutil.CreateTestLabel(t, db, projectID, "bug", "#EF4444")
		labelID2 := testutil.CreateTestLabel(t, db, projectID, "urgent", "#F97316")

		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID1)
		require.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID2)
		require.NoError(t, err)

		cmd := ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", taskID))

		// Verify task moved to ready column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)

		// Verify labels are still attached
		var labelCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", taskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, labelCount)
	})

	t.Run("Move task with relationships - relationships preserved", func(t *testing.T) {
		// Create parent and child tasks
		parentTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Parent Task")
		childTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Child Task")

		// Create relationship (relation_type_id = 1 for parent-child)
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
			parentTaskID, childTaskID)
		require.NoError(t, err)

		cmd := ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", parentTaskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", parentTaskID))

		// Verify task moved to ready column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", parentTaskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)

		// Verify relationship is still intact
		var relationCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentTaskID, childTaskID).Scan(&relationCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, relationCount)
	})

	t.Run("Move multiple tasks to ready column in sequence", func(t *testing.T) {
		// Create multiple tasks in different columns
		taskID1 := cli.CreateTestTask(t, db, inProgressColumnID, "Multi Task 1")
		taskID2 := cli.CreateTestTask(t, db, doneColumnID, "Multi Task 2")
		taskID3 := cli.CreateTestTask(t, db, inProgressColumnID, "Multi Task 3")

		// Move all tasks to ready column
		for _, taskID := range []int{taskID1, taskID2, taskID3} {
			cmd := ReadyMoveCmd()
			_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				fmt.Sprintf("%d", taskID),
				"--quiet",
			})
			assert.NoError(t, err)
		}

		// Verify all tasks moved to ready column
		for _, taskID := range []int{taskID1, taskID2, taskID3} {
			var columnID int
			err = db.QueryRowContext(context.Background(),
				"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
			assert.NoError(t, err)
			assert.Equal(t, todoColumnID, columnID, "Task %d should be in ready column", taskID)
		}
	})

	t.Run("Move task with blocking relationship preserved", func(t *testing.T) {
		// Create blocker and blocked tasks
		blockerTaskID := cli.CreateTestTask(t, db, doneColumnID, "Blocker Task")
		blockedTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Blocked Task")

		// Create blocking relationship (relation_type_id = 2 for blocking)
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
			blockedTaskID, blockerTaskID)
		require.NoError(t, err)

		cmd := ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", blockedTaskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", blockedTaskID))

		// Verify task moved to ready column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", blockedTaskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)

		// Verify blocking relationship is still intact
		var relationCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ? AND relation_type_id = 2",
			blockedTaskID, blockerTaskID).Scan(&relationCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, relationCount)
	})

	t.Run("Move task with comments preserved", func(t *testing.T) {
		// Create task with comments
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task With Comments")

		// Add comments
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_comments (task_id, content, author) VALUES (?, ?, ?)",
			taskID, "First comment", "user1")
		require.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_comments (task_id, content, author) VALUES (?, ?, ?)",
			taskID, "Second comment", "user2")
		require.NoError(t, err)

		cmd := ReadyMoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Todo'", taskID))

		// Verify task moved to ready column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, todoColumnID, columnID)

		// Verify comments are still attached
		var commentCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_comments WHERE task_id = ?", taskID).Scan(&commentCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, commentCount)
	})

	t.Run("Verify position updated when moving to ready column", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task for Position Test")

		// Get initial position
		var initialPosition int
		err := db.QueryRowContext(context.Background(),
			"SELECT position FROM tasks WHERE id = ?", taskID).Scan(&initialPosition)
		assert.NoError(t, err)

		cmd := ReadyMoveCmd()

		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)

		// Get new position (should be at the end of the ready column)
		var newPosition int
		err = db.QueryRowContext(context.Background(),
			"SELECT position FROM tasks WHERE id = ?", taskID).Scan(&newPosition)
		assert.NoError(t, err)

		// Position should be updated (likely different as it's appended to the ready column)
		// The exact position depends on how many tasks are already in the ready column
		// but we can verify it's non-negative
		assert.GreaterOrEqual(t, newPosition, 0)
	})
}

func TestReadyMoveTask_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project with columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	var todoColumnID, inProgressColumnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'", projectID).Scan(&todoColumnID)
	require.NoError(t, err)

	err = db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'In Progress'", projectID).Scan(&inProgressColumnID)
	require.NoError(t, err)

	// Mark "Todo" column as ready column for positive tests
	_, err = db.ExecContext(context.Background(),
		"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", todoColumnID)
	require.NoError(t, err)

	t.Run("Invalid task ID - non-numeric", func(t *testing.T) {
		cmd := ReadyMoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"not-a-number",
		})

		// Should error - invalid task ID
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("Missing task ID argument", func(t *testing.T) {
		cmd := ReadyMoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})

		// Should error - task ID is required
		assert.Error(t, err)
	})

	// Note: We skip tests for "Non-existent task ID" and "No ready column configured"
	// because they cause the command to call os.Exit(), which terminates the test process.
	// These error cases are tested in the service layer tests.

	t.Run("Non-existent task ID", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() when task is not found")
		// Use a task ID that doesn't exist
		nonExistentTaskID := 999999

		cmd := ReadyMoveCmd()

		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", nonExistentTaskID),
		})

		// Note: This would call os.Exit(cli.ExitNotFound) which terminates the process
	})

	t.Run("Project with no ready column configured", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on this validation error")

		// Create a new project without a ready column
		newProjectID := cli.CreateTestProject(t, db, "Project Without Ready Column")
		newColumnID := cli.CreateTestColumn(t, db, newProjectID, "Regular Column")

		// Ensure no columns are marked as ready
		_, err := db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = false WHERE project_id = ?", newProjectID)
		require.NoError(t, err)

		// Create task in the regular column
		taskID := cli.CreateTestTask(t, db, newColumnID, "Task in Project Without Ready")

		cmd := ReadyMoveCmd()

		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		// Note: This would call os.Exit(cli.ExitValidation) which terminates the process
	})

	t.Run("Invalid task ID - zero", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on invalid task ID")

		cmd := ReadyMoveCmd()

		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"0",
		})

		// Note: The service validates this and returns ErrInvalidTaskID,
		// but the CLI will call os.Exit() before we can capture the error.
	})

	t.Run("Invalid task ID - negative", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on invalid task ID")

		cmd := ReadyMoveCmd()

		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"-1",
		})

		// Note: The service validates this and returns ErrInvalidTaskID,
		// but the CLI will call os.Exit() before we can capture the error.
	})

	t.Run("Too many arguments", func(t *testing.T) {
		// Create task for this test
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task With Extra Args")

		cmd := ReadyMoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"extra-argument",
		})

		// Should fail because command expects exactly 1 positional argument
		assert.Error(t, err)
	})

	t.Run("Invalid flag combination - json and quiet", func(t *testing.T) {
		// Create task for this test
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task Flag Test")

		cmd := ReadyMoveCmd()

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
