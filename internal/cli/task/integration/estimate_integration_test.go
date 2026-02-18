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

func TestEstimateTask(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	ctx := context.Background()

	projectID := cli.CreateTestProject(t, db, "Test Project")
	columnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("set estimate on task", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for estimate")

		cmd := task.EstimateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"2d",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d estimate set to 2d", taskID))

		var estimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
		assert.NoError(t, err)
		assert.True(t, estimate.Valid, "estimate should be set")
		assert.Equal(t, "2d", estimate.String)
	})

	t.Run("set compound estimate", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for compound estimate")

		cmd := task.EstimateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"1w2d3h",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d estimate set to 1w2d3h", taskID))

		var estimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
		assert.NoError(t, err)
		assert.True(t, estimate.Valid)
		assert.Equal(t, "1w2d3h", estimate.String)
	})

	t.Run("set estimate with --quiet flag", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for quiet estimate")

		cmd := task.EstimateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"2d",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		var estimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
		assert.NoError(t, err)
		assert.True(t, estimate.Valid)
		assert.Equal(t, "2d", estimate.String)
	})

	t.Run("set estimate with --json flag", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for JSON estimate")

		cmd := task.EstimateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"2d",
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(taskID), result["task_id"])
		assert.Equal(t, "2d", result["estimate"])

		var estimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
		assert.NoError(t, err)
		assert.True(t, estimate.Valid)
		assert.Equal(t, "2d", estimate.String)
	})

	t.Run("clear estimate with --clear flag", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task to clear estimate")

		// First set an estimate
		cmd := task.EstimateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"2d",
		})
		require.NoError(t, err)

		// Verify estimate is set
		var estimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
		require.NoError(t, err)
		require.True(t, estimate.Valid, "estimate should be set before clearing")

		// Now clear
		cmd = task.EstimateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--clear",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d estimate cleared", taskID))

		// Verify estimate is cleared in DB
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
		assert.NoError(t, err)
		assert.False(t, estimate.Valid, "estimate should be null after clearing")
	})

	t.Run("clear estimate with --json", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task to clear with JSON")

		// First set an estimate
		cmd := task.EstimateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"3d",
		})
		require.NoError(t, err)

		// Clear with JSON output
		cmd = task.EstimateCmd()
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
		assert.Nil(t, result["estimate"])

		// Verify in DB
		var estimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
		assert.NoError(t, err)
		assert.False(t, estimate.Valid)
	})

	t.Run("update existing estimate", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task to update estimate")

		// Set initial estimate
		cmd := task.EstimateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"1d",
		})
		require.NoError(t, err)

		// Verify initial estimate
		var estimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
		require.NoError(t, err)
		require.Equal(t, "1d", estimate.String)

		// Update to new estimate
		cmd = task.EstimateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"2d",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d estimate set to 2d", taskID))

		// Verify estimate was updated
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
		assert.NoError(t, err)
		assert.Equal(t, "2d", estimate.String)
	})

	t.Run("various valid estimate formats", func(t *testing.T) {
		formats := []string{"4h", "30m", "1w", "2d", "1w2d", "3m"}

		for _, format := range formats {
			t.Run(format, func(t *testing.T) {
				taskID := cli.CreateTestTask(t, db, columnID, fmt.Sprintf("Task for %s", format))

				cmd := task.EstimateCmd()

				output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
					fmt.Sprintf("%d", taskID),
					format,
				})

				assert.NoError(t, err)
				assert.Contains(t, output, fmt.Sprintf("Task %d estimate set to %s", taskID, format))

				var estimate sql.NullString
				err = db.QueryRowContext(ctx,
					"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&estimate)
				assert.NoError(t, err)
				assert.True(t, estimate.Valid)
				assert.Equal(t, format, estimate.String)
			})
		}
	})
}

func TestEstimateTask_Errors(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	projectID := cli.CreateTestProject(t, db, "Test Project")
	columnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("invalid task ID", func(t *testing.T) {
		cmd := task.EstimateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"notanumber",
			"2d",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("non-existent task ID", func(t *testing.T) {
		cmd := task.EstimateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"999999",
			"2d",
		})

		assert.Error(t, err)
		cli.AssertExitError(t, err, 3) // ExitNotFound
		assert.Contains(t, err.Error(), "task 999999 not found")
	})

	t.Run("missing task ID", func(t *testing.T) {
		cmd := task.EstimateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})

		// cobra.RangeArgs(1, 2) should reject zero args
		assert.Error(t, err)
	})

	t.Run("missing estimate value without --clear", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for missing estimate")

		cmd := task.EstimateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.Error(t, err)
		cli.AssertExitError(t, err, 5) // ExitValidation
		assert.Contains(t, err.Error(), "estimate value is required")
	})

	t.Run("invalid estimate format - invalid unit", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for invalid unit")

		cmd := task.EstimateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"2x",
		})

		assert.Error(t, err)
		cli.AssertExitError(t, err, 5) // ExitValidation
	})

	t.Run("invalid estimate format - no unit", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for no unit")

		cmd := task.EstimateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"5",
		})

		assert.Error(t, err)
		cli.AssertExitError(t, err, 5) // ExitValidation
	})

	t.Run("invalid estimate format - duplicate unit", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for duplicate unit")

		cmd := task.EstimateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"1d2d",
		})

		assert.Error(t, err)
		cli.AssertExitError(t, err, 5) // ExitValidation
	})

	t.Run("invalid estimate format - spaces", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for spaces")

		cmd := task.EstimateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"1d 2h",
		})

		assert.Error(t, err)
		cli.AssertExitError(t, err, 5) // ExitValidation
	})

	t.Run("too many arguments", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task for too many args")

		cmd := task.EstimateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"2d",
			"extra",
		})

		// cobra.RangeArgs(1, 2) should reject 3 args
		assert.Error(t, err)
	})
}
