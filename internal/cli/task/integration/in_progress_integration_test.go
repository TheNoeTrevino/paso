package task_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestInProgressTask(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()

	// Create test project with columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	// Create an "In Progress" column and mark it
	inProgressColumnID := cli.CreateTestColumn(t, db, projectID, "In Progress")
	cli.SetColumnHoldsInProgressTasks(t, db, inProgressColumnID)

	t.Run("mark task as in-progress", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task to Start")

		cli.UpdateTaskFields(t, db, taskID, map[string]any{"ticket_number": 1})

		cmd := task.InProgressCmd()

		// Note: InProgressCmd takes task ID as positional arg, not --id flag!
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved to in-progress column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, inProgressColumnID, columnID)
	})

	t.Run("mark task as in-progress - quiet mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Mode Task")

		cli.UpdateTaskFields(t, db, taskID, map[string]any{"ticket_number": 2})

		cmd := task.InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should only output task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify task moved to in-progress column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, inProgressColumnID, columnID)
	})

	t.Run("mark task as in-progress - JSON mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "JSON Mode Task")

		cli.UpdateTaskFields(t, db, taskID, map[string]any{"ticket_number": 3})

		cmd := task.InProgressCmd()

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
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(taskID), result["task_id"])
		assert.Equal(t, "Todo", result["from_column"])
		assert.Equal(t, "In Progress", result["to_column"])

		// Verify task moved to in-progress column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, inProgressColumnID, columnID)
	})

	t.Run("list in-progress tasks", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create task and move it to in-progress using the command
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Listed In Progress Task")
		cli.UpdateTaskFields(t, db, taskID, map[string]any{"ticket_number": 4})

		// Move task to in-progress using the command
		moveCmd := task.InProgressCmd()
		_, err := cli.ExecuteCLICommand(t, app, moveCmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})
		assert.NoError(t, err)

		// Now list in-progress tasks
		listCmd := task.InProgressCmd()
		output, err := cli.ExecuteCLICommand(t, app, listCmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Listed In Progress Task")
		assert.Contains(t, output, "Found")
		assert.Contains(t, output, "in-progress tasks")
	})

	t.Run("list in-progress tasks - quiet mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create and move tasks to in-progress using the command
		taskID1 := cli.CreateTestTask(t, db, todoColumnID, "Quiet List Task 1")
		taskID2 := cli.CreateTestTask(t, db, todoColumnID, "Quiet List Task 2")
		cli.UpdateTaskFields(t, db, taskID1, map[string]any{"ticket_number": 5})
		cli.UpdateTaskFields(t, db, taskID2, map[string]any{"ticket_number": 6})

		// Move tasks using the command
		moveCmd1 := task.InProgressCmd()
		_, err := cli.ExecuteCLICommand(t, app, moveCmd1, []string{
			fmt.Sprintf("%d", taskID1),
			"--quiet",
		})
		assert.NoError(t, err)

		moveCmd2 := task.InProgressCmd()
		_, err = cli.ExecuteCLICommand(t, app, moveCmd2, []string{
			fmt.Sprintf("%d", taskID2),
			"--quiet",
		})
		assert.NoError(t, err)

		// Now list in quiet mode
		listCmd := task.InProgressCmd()
		output, err := cli.ExecuteCLICommand(t, app, listCmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--quiet",
		})

		assert.NoError(t, err)

		// Quiet mode should only output task IDs, one per line
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.GreaterOrEqual(t, len(lines), 2, "Should have at least 2 task IDs")

		// Verify each line is a numeric task ID
		for _, line := range lines {
			assert.Regexp(t, `^\d+$`, line)
		}
	})

	t.Run("list in-progress tasks - JSON mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create and move a task to in-progress using the command
		taskID := cli.CreateTestTask(t, db, todoColumnID, "JSON List Task")
		cli.UpdateTaskFields(t, db, taskID, map[string]any{"ticket_number": 7})

		// Move task using the command
		moveCmd := task.InProgressCmd()
		_, err := cli.ExecuteCLICommand(t, app, moveCmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})
		assert.NoError(t, err)

		// Now list in JSON mode
		listCmd := task.InProgressCmd()
		output, err := cli.ExecuteCLICommand(t, app, listCmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["tasks"])
		assert.NotNil(t, result["count"])

		tasks := result["tasks"].([]any)
		assert.GreaterOrEqual(t, len(tasks), 1, "Should have at least 1 task")

		// Verify task structure
		foundTask := false
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]any)
			if int(taskData["id"].(float64)) == taskID {
				foundTask = true
				assert.Equal(t, "JSON List Task", taskData["title"])
				assert.Equal(t, float64(7), taskData["ticket_number"])
				break
			}
		}
		assert.True(t, foundTask, "Should find the created task in list")
	})

	t.Run("task already in in-progress column", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create task already in in-progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Already In Progress")

		cli.UpdateTaskFields(t, db, taskID, map[string]any{"ticket_number": 8})

		cmd := task.InProgressCmd()

		// Try to move to in-progress again (should handle gracefully)
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		// Should not error - command handles this gracefully
		assert.NoError(t, err)
		// Output should contain the informational message and task ID
		assert.Contains(t, output, "Task")
		assert.Contains(t, output, "already in the in-progress column")
		assert.Contains(t, output, fmt.Sprintf("%d", taskID))
	})

	t.Run("multiple tasks moved to in-progress", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create multiple tasks
		taskID1 := cli.CreateTestTask(t, db, todoColumnID, "Multi Task 1")
		taskID2 := cli.CreateTestTask(t, db, todoColumnID, "Multi Task 2")
		taskID3 := cli.CreateTestTask(t, db, todoColumnID, "Multi Task 3")

		cli.UpdateTaskFields(t, db, taskID1, map[string]any{"ticket_number": 9})
		cli.UpdateTaskFields(t, db, taskID2, map[string]any{"ticket_number": 10})
		cli.UpdateTaskFields(t, db, taskID3, map[string]any{"ticket_number": 11})

		cmd1 := task.InProgressCmd()
		cmd2 := task.InProgressCmd()
		cmd3 := task.InProgressCmd()

		// Move all tasks to in-progress
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			fmt.Sprintf("%d", taskID1),
			"--quiet",
		})
		assert.NoError(t, err)

		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			fmt.Sprintf("%d", taskID2),
			"--quiet",
		})
		assert.NoError(t, err)

		_, err = cli.ExecuteCLICommand(t, app, cmd3, []string{
			fmt.Sprintf("%d", taskID3),
			"--quiet",
		})
		assert.NoError(t, err)

		// Verify all tasks moved to in-progress column
		for _, taskID := range []int{taskID1, taskID2, taskID3} {
			columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
			assert.Equal(t, inProgressColumnID, columnID)
		}
	})

	t.Run("list in-progress tasks with different priorities", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create tasks with different priorities
		lowPriorityTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Low Priority Task")
		highPriorityTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "High Priority Task")
		criticalPriorityTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Critical Priority Task")

		// Assign ticket numbers and priorities
		cli.UpdateTaskFields(t, db, lowPriorityTaskID, map[string]any{"ticket_number": 12, "priority_id": 2})
		cli.UpdateTaskFields(t, db, highPriorityTaskID, map[string]any{"ticket_number": 13, "priority_id": 4})
		cli.UpdateTaskFields(t, db, criticalPriorityTaskID, map[string]any{"ticket_number": 14, "priority_id": 5})

		cmd := task.InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Low Priority Task")
		assert.Contains(t, output, "High Priority Task")
		assert.Contains(t, output, "Critical Priority Task")
	})

	t.Run("list in-progress tasks with blocked status", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create blocker and blocked tasks
		blockerTaskID := cli.CreateTestTask(t, db, todoColumnID, "Blocker Task")
		blockedTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Blocked In Progress Task")

		cli.UpdateTaskFields(t, db, blockerTaskID, map[string]any{"ticket_number": 15})
		cli.UpdateTaskFields(t, db, blockedTaskID, map[string]any{"ticket_number": 16})

		// Create blocking relationship
		cli.AddTaskSubtask(t, db, blockedTaskID, blockerTaskID, 2)

		cmd := task.InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Blocked In Progress Task")
		// Should show blocked indicator
		assert.Contains(t, output, "BLOCKED")
	})

	t.Run("list empty in-progress tasks", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a new project with no in-progress tasks
		newProjectID := cli.CreateTestProject(t, db, "Empty Project")
		emptyInProgressColumnID := cli.CreateTestColumn(t, db, newProjectID, "In Progress")

		// Mark as in-progress column
		cli.SetColumnHoldsInProgressTasks(t, db, emptyInProgressColumnID)

		cmd := task.InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", newProjectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No in-progress tasks found")
	})

	t.Run("move task with labels to in-progress", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create task with labels
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task With Labels")

		cli.UpdateTaskFields(t, db, taskID, map[string]any{"ticket_number": 17})

		// Create and attach labels
		labelID1 := cli.CreateTestLabel(t, db, projectID, "bug", "#EF4444")
		labelID2 := cli.CreateTestLabel(t, db, projectID, "urgent", "#F97316")
		cli.AttachLabelToTask(t, db, taskID, labelID1)
		cli.AttachLabelToTask(t, db, taskID, labelID2)

		cmd := task.InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved to in-progress column
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, inProgressColumnID, columnID)

		// Verify labels are still attached
		var labelCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", taskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, labelCount)
	})

	t.Run("move task from different column to in-progress", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a "Done" column
		doneColumnID := cli.CreateTestColumn(t, db, projectID, "Done")

		// Create task in Done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task From Done")

		cli.UpdateTaskFields(t, db, taskID, map[string]any{"ticket_number": 18})

		cmd := task.InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved from Done to In Progress
		columnID := fixtures.GetTaskColumnID(t, db, fixtures.SQLiteDialect(), taskID)
		assert.Equal(t, inProgressColumnID, columnID)
	})

	t.Run("move task with description to in-progress", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create task with description
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task With Description")

		description := "This is a detailed description of the task that needs to be completed."
		cli.UpdateTaskFields(t, db, taskID, map[string]any{"description": description, "ticket_number": 19})

		cmd := task.InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved and description is preserved
		var columnID int
		var savedDescription string
		err = db.QueryRowContext(ctx,
			"SELECT column_id, description FROM tasks WHERE id = ?", taskID).Scan(&columnID, &savedDescription)
		assert.NoError(t, err)
		assert.Equal(t, inProgressColumnID, columnID)
		assert.Equal(t, description, savedDescription)
	})

	t.Run("list in-progress tasks in JSON with complete structure", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a task with all metadata
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Complete Metadata Task")

		cli.UpdateTaskFields(t, db, taskID, map[string]any{
			"description":   "Complete description",
			"ticket_number": 20,
			"type_id":       2,
			"priority_id":   4,
		})

		cmd := task.InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify structure
		assert.True(t, result["success"].(bool))
		tasks := result["tasks"].([]any)

		// Find our task
		foundTask := false
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]any)
			if int(taskData["id"].(float64)) == taskID {
				foundTask = true
				assert.Equal(t, "Complete Metadata Task", taskData["title"])
				assert.Equal(t, float64(20), taskData["ticket_number"])
				assert.Equal(t, "feature", taskData["type_description"])
				assert.Equal(t, "high", taskData["priority_description"])
				assert.NotEmpty(t, taskData["priority_color"])
				break
			}
		}
		assert.True(t, foundTask, "Should find the created task")
	})
}
