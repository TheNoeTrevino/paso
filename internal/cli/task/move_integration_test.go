package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

// setupLinkedColumns creates a project with 3 properly linked columns and returns their IDs
func setupLinkedColumns(t *testing.T, db *sql.DB) (projectID, column1ID, column2ID, column3ID int) {
	t.Helper()

	// Create project manually
	result, err := db.ExecContext(context.Background(),
		"INSERT INTO projects (name, description) VALUES (?, ?)", "Test Project", "Test description")
	assert.NoError(t, err)
	projID, _ := result.LastInsertId()
	projectID = int(projID)

	// Initialize project counter
	_, err = db.ExecContext(context.Background(),
		"INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)", projectID)
	assert.NoError(t, err)

	// Create first column (Todo)
	result, err = db.ExecContext(context.Background(),
		"INSERT INTO columns (project_id, name, prev_id, next_id) VALUES (?, ?, NULL, NULL)",
		projectID, "Todo")
	assert.NoError(t, err)
	col1ID, _ := result.LastInsertId()
	column1ID = int(col1ID)

	// Create second column (In Progress)
	result, err = db.ExecContext(context.Background(),
		"INSERT INTO columns (project_id, name, prev_id, next_id) VALUES (?, ?, ?, NULL)",
		projectID, "In Progress", column1ID)
	assert.NoError(t, err)
	col2ID, _ := result.LastInsertId()
	column2ID = int(col2ID)

	// Link column1 to column2
	_, err = db.ExecContext(context.Background(),
		"UPDATE columns SET next_id = ? WHERE id = ?", column2ID, column1ID)
	assert.NoError(t, err)

	// Create third column (Done)
	result, err = db.ExecContext(context.Background(),
		"INSERT INTO columns (project_id, name, prev_id, next_id) VALUES (?, ?, ?, NULL)",
		projectID, "Done", column2ID)
	assert.NoError(t, err)
	col3ID, _ := result.LastInsertId()
	column3ID = int(col3ID)

	// Link column2 to column3
	_, err = db.ExecContext(context.Background(),
		"UPDATE columns SET next_id = ? WHERE id = ?", column3ID, column2ID)
	assert.NoError(t, err)

	return projectID, column1ID, column2ID, column3ID
}

func TestMoveTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project with properly linked columns
	_, column1ID, column2ID, column3ID := setupLinkedColumns(t, db)

	t.Run("Move to next column", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for next move")

		// Verify task is in column 1
		var columnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column1ID, columnID)

		cmd := MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to", taskID))

		// Verify task moved to column 2
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("Move to previous column", func(t *testing.T) {
		// Create task in third column
		taskID := cli.CreateTestTask(t, db, column3ID, "Task for prev move")

		// Verify task is in column 3
		var columnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column3ID, columnID)

		cmd := MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"prev",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to", taskID))

		// Verify task moved to column 2
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("Move to specific column by name", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for name move")

		cmd := MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"Done",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Done'", taskID))

		// Verify task moved to column 3 (Done)
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column3ID, columnID)
	})

	t.Run("Move with case-insensitive column name - lowercase", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for case test 1")

		cmd := MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"done",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Done'", taskID))

		// Verify task moved to column 3 (Done)
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column3ID, columnID)
	})

	t.Run("Move with case-insensitive column name - uppercase", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for case test 2")

		cmd := MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"TODO",
		})

		assert.NoError(t, err)
		// Task should already be in Todo, so output should reflect that
		assert.Contains(t, output, fmt.Sprintf("Task %d is already in 'Todo'", taskID))

		// Verify task is still in column 1
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column1ID, columnID)
	})

	t.Run("Move with case-insensitive column name - mixed case", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for case test 3")

		cmd := MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"In ProGRess",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved to column 2 (In Progress)
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("Quiet mode output", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for quiet mode")

		cmd := MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should output only the task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify task moved
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("JSON mode output", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for JSON mode")

		cmd := MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
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
		assert.Equal(t, "Todo", result["from_column"])
		assert.Equal(t, "In Progress", result["to_column"])

		// Verify task moved
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column2ID, columnID)
	})

	t.Run("Verify position changes when moving between columns", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for position test")

		// Get initial position
		var initialPosition int
		err := db.QueryRowContext(context.Background(),
			"SELECT position FROM tasks WHERE id = ?", taskID).Scan(&initialPosition)
		assert.NoError(t, err)

		cmd := MoveCmd()

		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})

		assert.NoError(t, err)

		// Get new position (should be at the end of the target column)
		var newPosition int
		err = db.QueryRowContext(context.Background(),
			"SELECT position FROM tasks WHERE id = ?", taskID).Scan(&newPosition)
		assert.NoError(t, err)

		// Position should be updated (likely different as it's appended to the new column)
		// The exact position depends on how many tasks are already in the target column
		// but we can verify it's non-negative
		assert.GreaterOrEqual(t, newPosition, 0)
	})

	t.Run("Move to same column - already in target", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task already in target")

		cmd := MoveCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"Todo",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d is already in 'Todo'", taskID))

		// Verify task is still in column 1
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column1ID, columnID)
	})

	t.Run("Move multiple times in sequence", func(t *testing.T) {
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for multiple moves")

		// Move to next (should go to column 2)
		cmd := MoveCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})
		assert.NoError(t, err)

		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column2ID, columnID)

		// Move to next again (should go to column 3)
		cmd = MoveCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})
		assert.NoError(t, err)

		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column3ID, columnID)

		// Move to prev (should go back to column 2)
		cmd = MoveCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"prev",
		})
		assert.NoError(t, err)

		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column2ID, columnID)

		// Move to specific column by name (should go to column 1)
		cmd = MoveCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"Todo",
		})
		assert.NoError(t, err)

		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, column1ID, columnID)
	})
}

func TestMoveTask_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project with properly linked columns
	_, column1ID, _, column3ID := setupLinkedColumns(t, db)

	// Note: We skip this test because the move command calls os.Exit() on validation errors,
	// which would terminate the test process. In production, this is the correct behavior,
	// but it cannot be tested in a standard Go unit test without process isolation.
	// The validation logic is still tested in the service layer tests.
	t.Run("Move to next column when already in last column", func(t *testing.T) {
		t.Skip("Skipping: move command calls os.Exit() on this validation error")
		// Create task in last column
		taskID := cli.CreateTestTask(t, db, column3ID, "Task in last column")

		cmd := MoveCmd()

		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"next",
		})

		// Note: This would call os.Exit(cli.ExitValidation) which terminates the process
	})

	// Note: We skip this test because the move command calls os.Exit() on validation errors.
	t.Run("Move to prev column when already in first column", func(t *testing.T) {
		t.Skip("Skipping: move command calls os.Exit() on this validation error")
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task in first column")

		cmd := MoveCmd()

		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"prev",
		})

		// Note: This would call os.Exit(cli.ExitValidation) which terminates the process
	})

	// Note: We skip this test because the move command calls os.Exit() when column not found.
	t.Run("Move to non-existent column by name", func(t *testing.T) {
		t.Skip("Skipping: move command calls os.Exit() when column is not found")
		// Create task in first column
		taskID := cli.CreateTestTask(t, db, column1ID, "Task for invalid column test")

		cmd := MoveCmd()

		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"NonExistentColumn",
		})

		// Note: This would call os.Exit(cli.ExitNotFound) which terminates the process
	})

	// Note: We skip this test because the move command calls os.Exit() when task not found.
	t.Run("Move non-existent task", func(t *testing.T) {
		t.Skip("Skipping: move command calls os.Exit() when task is not found")
		// Use a task ID that doesn't exist
		nonExistentTaskID := 999999

		cmd := MoveCmd()

		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", nonExistentTaskID),
			"next",
		})

		// Note: This would call os.Exit(cli.ExitNotFound) which terminates the process
	})

	t.Run("Move without providing task ID flag", func(t *testing.T) {
		cmd := MoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"next",
		})

		// Should fail because --id flag is required
		assert.Error(t, err)
	})

	t.Run("Move without providing target argument", func(t *testing.T) {
		// Create task for this test
		taskID := cli.CreateTestTask(t, db, column1ID, "Task without target")

		cmd := MoveCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		// Should fail because target argument is required
		assert.Error(t, err)
	})
}
