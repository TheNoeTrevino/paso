package task_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestDoneTask_Integration(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	// Create test project with columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default columns created by CreateTestProject
	todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")
	inProgressColumnID := cli.GetColumnIDByName(t, db, projectID, "In Progress")
	doneColumnID := cli.GetColumnIDByName(t, db, projectID, "Done")

	// Mark "Done" column as completed column
	cli.SetColumnHoldsCompletedTasks(t, db, doneColumnID)

	t.Run("mark task as done - default output", func(t *testing.T) {
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task to Complete")

		cmd := task.DoneCmd()

		// Note: DoneCmd takes task ID as positional arg, not --id flag!
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved from Todo to Done", taskID))

		// Verify task moved to done column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("mark task as done - quiet mode output", func(t *testing.T) {
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task for Quiet Mode")

		cmd := task.DoneCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should output only task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify task moved to done column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("mark task as done - JSON mode output", func(t *testing.T) {
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task for JSON Mode")

		cmd := task.DoneCmd()

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
		assert.Equal(t, "Todo", result["from_column"])
		assert.Equal(t, "Done", result["to_column"])

		// Verify task moved to done column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("verify column transition from In Progress to Done", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task from In Progress")

		cmd := task.DoneCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output to verify transition
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		assert.Equal(t, "In Progress", result["from_column"])
		assert.Equal(t, "Done", result["to_column"])

		// Verify task moved to done column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("task already in completed column", func(t *testing.T) {
		// Create task directly in done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Already Done Task")

		cmd := task.DoneCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		// Should not error - exits successfully
		assert.NoError(t, err)
		// Output should contain the informational message and task ID
		assert.Contains(t, output, "Task")
		assert.Contains(t, output, "already in the completed column")
		assert.Contains(t, output, fmt.Sprintf("%d", taskID))

		// Verify task is still in done column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("multiple tasks marked done", func(t *testing.T) {
		// Create multiple tasks in different columns
		taskID1 := cli.CreateTestTask(t, db, todoColumnID, "Task 1 to Complete")
		taskID2 := cli.CreateTestTask(t, db, todoColumnID, "Task 2 to Complete")
		taskID3 := cli.CreateTestTask(t, db, inProgressColumnID, "Task 3 to Complete")

		// Mark all tasks as done
		for _, taskID := range []int{taskID1, taskID2, taskID3} {
			cmd := task.DoneCmd()
			_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				fmt.Sprintf("%d", taskID),
				"--quiet",
			})
			assert.NoError(t, err)
		}

		// Verify all tasks moved to done column
		for _, taskID := range []int{taskID1, taskID2, taskID3} {
			columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
			assert.Equal(t, doneColumnID, columnID, "Task %d should be in Done column", taskID)
		}
	})
}

func TestDoneTask_Integration_Errors(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	t.Run("invalid task ID - non-numeric", func(t *testing.T) {
		cmd := task.DoneCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"not-a-number"})
		cli.AssertExitError(t, err, 5) // ExitValidation
		assert.Contains(t, err.Error(), "invalid task ID: not-a-number")
	})
}
