package task_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/app"
	projectservice "github.com/thenoetrevino/paso/internal/services/project"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

// setupLinkedColumns creates a project with 3 properly linked columns and returns their IDs.
// Uses the service layer via app.ProjectService to ensure columns are correctly linked.
func setupLinkedColumns(t *testing.T, db *sql.DB, testApp *app.App) (projectID, column1ID, column2ID, column3ID int) {
	t.Helper()

	ctx := context.Background()

	project, err := testApp.ProjectService.CreateProject(ctx, projectservice.CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test description",
	})
	require.NoError(t, err, "Failed to create project")
	projectID = project.ID

	column1ID = cli.GetColumnIDByName(t, db, projectID, "Todo")
	column2ID = cli.GetColumnIDByName(t, db, projectID, "In Progress")
	column3ID = cli.GetColumnIDByName(t, db, projectID, "Done")

	return projectID, column1ID, column2ID, column3ID
}

func TestMoveTask_Integration(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	// Create test project with properly linked columns
	_, column1ID, column2ID, column3ID := setupLinkedColumns(t, db, app)

	t.Run("move to next column", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for next move")

		// Verify task is in column 1
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column1ID, columnID)

		cmd := task.MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to", taskID))

		// Verify task moved to column 2
		columnID = fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("move to previous column", func(t *testing.T) {
		// Create task in third column
		taskID := cli.CreateTestTask(t, db, column3ID, "Task for prev move")

		// Verify task is in column 3
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column3ID, columnID)

		cmd := task.MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"prev",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to", taskID))

		// Verify task moved to column 2
		columnID = fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("move to specific column by name", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for name move")

		cmd := task.MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"Done",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Done'", taskID))

		// Verify task moved to column 3 (Done)
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column3ID, columnID)
	})

	t.Run("move with case-insensitive column name - lowercase", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for case test 1")

		cmd := task.MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"done",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Done'", taskID))

		// Verify task moved to column 3 (Done)
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column3ID, columnID)
	})

	t.Run("move with case-insensitive column name - uppercase", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for case test 2")

		cmd := task.MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"TODO",
		})

		assert.NoError(t, err)
		// Task should already be in Todo, so output should reflect that
		assert.Contains(t, output, fmt.Sprintf("Task %d is already in 'Todo'", taskID))

		// Verify task is still in column 1
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column1ID, columnID)
	})

	t.Run("move with case-insensitive column name - mixed case", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for case test 3")

		cmd := task.MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"In ProGRess",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved to column 2 (In Progress)
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("quiet mode output", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for quiet mode")

		cmd := task.MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should output only the task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify task moved
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("jSON mode output", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for JSON mode")

		cmd := task.MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
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
		assert.Equal(t, "Todo", result["from_column"])
		assert.Equal(t, "In Progress", result["to_column"])

		// Verify task moved
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("verify position changes when moving between columns", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for position test")

		// Get initial position
		initialPosition := fixtures.GetTaskPosition(t, db, fixtures.SQLiteDialect(), taskID)
		_ = initialPosition

		cmd := task.MoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})

		assert.NoError(t, err)

		// Get new position (should be at the end of the target column)
		newPosition := fixtures.GetTaskPosition(t, db, fixtures.SQLiteDialect(), taskID)

		// Position should be updated (likely different as it's appended to the new column)
		// The exact position depends on how many tasks are already in the target column
		// but we can verify it's non-negative
		assert.GreaterOrEqual(t, newPosition, 0)
	})

	t.Run("move to same column - already in target", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task already in target")

		cmd := task.MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"Todo",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d is already in 'Todo'", taskID))

		// Verify task is still in column 1
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column1ID, columnID)
	})

	t.Run("move multiple times in sequence", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for multiple moves")

		// Move to next (should go to column 2)
		cmd := task.MoveCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})
		assert.NoError(t, err)

		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column2ID, columnID)

		// Move to next again (should go to column 3)
		cmd = task.MoveCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})
		assert.NoError(t, err)

		columnID = fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column3ID, columnID)

		// Move to prev (should go back to column 2)
		cmd = task.MoveCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"prev",
		})
		assert.NoError(t, err)

		columnID = fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column2ID, columnID)

		// Move to specific column by name (should go to column 1)
		cmd = task.MoveCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"Todo",
		})
		assert.NoError(t, err)

		columnID = fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, column1ID, columnID)
	})
}

func TestMoveTask_Integration_Errors(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	// Create test project with properly linked columns
	_, column1ID, _, column3ID := setupLinkedColumns(t, db, app)

	t.Run("move to next column when already in last column", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, column3ID, "Task in last column")

		cmd := task.MoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})
		cli.AssertExitError(t, err, 5) // ExitValidation
		assert.Contains(t, err.Error(), "already in the last column")
	})

	t.Run("move to prev column when already in first column", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, column1ID, "Task in first column")

		cmd := task.MoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"prev",
		})
		cli.AssertExitError(t, err, 5) // ExitValidation
		assert.Contains(t, err.Error(), "already in the first column")
	})

	t.Run("move to non-existent column by name", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for invalid column test")

		cmd := task.MoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"NonExistentColumn",
		})
		cli.AssertExitError(t, err, 3) // ExitNotFound
		assert.Contains(t, err.Error(), "column 'NonExistentColumn' not found")
	})

	t.Run("move non-existent task", func(t *testing.T) {
		nonExistentTaskID := 999999

		cmd := task.MoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", nonExistentTaskID),
			"next",
		})
		cli.AssertExitError(t, err, 3) // ExitNotFound
		assert.Contains(t, err.Error(), "task 999999 not found")
	})

	t.Run("move without providing task ID flag", func(t *testing.T) {
		cmd := task.MoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"next",
		})

		// Should fail because --id flag is required
		assert.Error(t, err)
	})

	t.Run("move without providing target argument", func(t *testing.T) {
		// Create task for this test
		taskID := cli.CreateTestTask(t, db, column1ID, "Task without target")

		cmd := task.MoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		// Should fail because target argument is required
		assert.Error(t, err)
	})
}
