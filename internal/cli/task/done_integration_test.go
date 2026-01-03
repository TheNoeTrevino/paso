package task

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestDoneTask_Positive(t *testing.T) {
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
	assert.NoError(t, err)

	err = db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'In Progress'", projectID).Scan(&inProgressColumnID)
	assert.NoError(t, err)

	err = db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Done'", projectID).Scan(&doneColumnID)
	assert.NoError(t, err)

	// Mark "Done" column as completed column
	_, err = db.ExecContext(context.Background(),
		"UPDATE columns SET holds_completed_tasks = true WHERE id = ?", doneColumnID)
	assert.NoError(t, err)

	t.Run("Mark task as done - default output", func(t *testing.T) {
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task to Complete")

		cmd := DoneCmd()

		// Note: DoneCmd takes task ID as positional arg, not --id flag!
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'Done'", taskID))

		// Verify task moved to done column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("Mark task as done - quiet mode output", func(t *testing.T) {
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task for Quiet Mode")

		cmd := DoneCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should output only task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify task moved to done column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("Mark task as done - JSON mode output", func(t *testing.T) {
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task for JSON Mode")

		cmd := DoneCmd()

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
		assert.Equal(t, "Todo", result["from_column"])
		assert.Equal(t, "Done", result["to_column"])

		// Verify task moved to done column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("Verify column transition from In Progress to Done", func(t *testing.T) {
		// Create task in In Progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Task from In Progress")

		cmd := DoneCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output to verify transition
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		assert.Equal(t, "In Progress", result["from_column"])
		assert.Equal(t, "Done", result["to_column"])

		// Verify task moved to done column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("Task already in completed column", func(t *testing.T) {
		// Create task directly in done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Already Done Task")

		cmd := DoneCmd()

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
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, doneColumnID, columnID)
	})

	t.Run("Multiple tasks marked done", func(t *testing.T) {
		// Create multiple tasks in different columns
		taskID1 := cli.CreateTestTask(t, db, todoColumnID, "Task 1 to Complete")
		taskID2 := cli.CreateTestTask(t, db, todoColumnID, "Task 2 to Complete")
		taskID3 := cli.CreateTestTask(t, db, inProgressColumnID, "Task 3 to Complete")

		// Mark all tasks as done
		for _, taskID := range []int{taskID1, taskID2, taskID3} {
			cmd := DoneCmd()
			_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				fmt.Sprintf("%d", taskID),
				"--quiet",
			})
			assert.NoError(t, err)
		}

		// Verify all tasks moved to done column
		for _, taskID := range []int{taskID1, taskID2, taskID3} {
			var columnID int
			err = db.QueryRowContext(context.Background(),
				"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
			assert.NoError(t, err)
			assert.Equal(t, doneColumnID, columnID, "Task %d should be in Done column", taskID)
		}
	})
}

func TestDoneTask_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project with columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	var todoColumnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'", projectID).Scan(&todoColumnID)
	assert.NoError(t, err)

	// Note: Tests for "No completed column configured", "Invalid task ID", and "Task does not exist"
	// are not included here because they cause the command to call os.Exit(), which terminates
	// the test process. These error cases should be tested in unit tests of the service layer
	// or by capturing the exit code in a subprocess test.

	t.Run("Invalid task ID - non-numeric", func(t *testing.T) {
		// Mark done column for this test so we don't hit the os.Exit() path
		var doneColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Done'", projectID).Scan(&doneColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_completed_tasks = true WHERE id = ?", doneColumnID)
		assert.NoError(t, err)

		cmd := DoneCmd()

		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"not-a-number",
		})

		// Should error - invalid task ID
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})
}
