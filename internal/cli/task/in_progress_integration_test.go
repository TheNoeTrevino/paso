package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestInProgressTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project with columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	var todoColumnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&todoColumnID)
	assert.NoError(t, err)

	// Create an "In Progress" column
	inProgressColumnID := cli.CreateTestColumn(t, db, projectID, "In Progress")

	// Mark "In Progress" column as in-progress column
	_, err = db.ExecContext(context.Background(),
		"UPDATE columns SET holds_in_progress_tasks = true WHERE id = ?", inProgressColumnID)
	assert.NoError(t, err)

	t.Run("Mark task as in-progress", func(t *testing.T) {
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task to Start")

		// Assign ticket number
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 1 WHERE id = ?", taskID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		// Note: InProgressCmd takes task ID as positional arg, not --id flag!
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved to in-progress column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, inProgressColumnID, columnID)
	})

	t.Run("Mark task as in-progress - quiet mode", func(t *testing.T) {
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Mode Task")

		// Assign ticket number
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 2 WHERE id = ?", taskID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should only output task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify task moved to in-progress column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, inProgressColumnID, columnID)
	})

	t.Run("Mark task as in-progress - JSON mode", func(t *testing.T) {
		// Create task in todo column
		taskID := cli.CreateTestTask(t, db, todoColumnID, "JSON Mode Task")

		// Assign ticket number
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 3 WHERE id = ?", taskID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

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
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(taskID), result["task_id"])
		assert.Equal(t, "Todo", result["from_column"])
		assert.Equal(t, "In Progress", result["to_column"])

		// Verify task moved to in-progress column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, inProgressColumnID, columnID)
	})

	t.Run("List in-progress tasks", func(t *testing.T) {
		// Create task and move it to in-progress using the command
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Listed In Progress Task")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 4 WHERE id = ?", taskID)
		assert.NoError(t, err)

		// Move task to in-progress using the command
		moveCmd := InProgressCmd()
		_, err = cli.ExecuteCLICommand(t, app, moveCmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})
		assert.NoError(t, err)

		// Now list in-progress tasks
		listCmd := InProgressCmd()
		output, err := cli.ExecuteCLICommand(t, app, listCmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Listed In Progress Task")
		assert.Contains(t, output, "Found")
		assert.Contains(t, output, "in-progress tasks")
	})

	t.Run("List in-progress tasks - quiet mode", func(t *testing.T) {
		// Create and move tasks to in-progress using the command
		taskID1 := cli.CreateTestTask(t, db, todoColumnID, "Quiet List Task 1")
		taskID2 := cli.CreateTestTask(t, db, todoColumnID, "Quiet List Task 2")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 5 WHERE id = ?", taskID1)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 6 WHERE id = ?", taskID2)
		assert.NoError(t, err)

		// Move tasks using the command
		moveCmd1 := InProgressCmd()
		_, err = cli.ExecuteCLICommand(t, app, moveCmd1, []string{
			fmt.Sprintf("%d", taskID1),
			"--quiet",
		})
		assert.NoError(t, err)

		moveCmd2 := InProgressCmd()
		_, err = cli.ExecuteCLICommand(t, app, moveCmd2, []string{
			fmt.Sprintf("%d", taskID2),
			"--quiet",
		})
		assert.NoError(t, err)

		// Now list in quiet mode
		listCmd := InProgressCmd()
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

	t.Run("List in-progress tasks - JSON mode", func(t *testing.T) {
		// Create and move a task to in-progress using the command
		taskID := cli.CreateTestTask(t, db, todoColumnID, "JSON List Task")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 7 WHERE id = ?", taskID)
		assert.NoError(t, err)

		// Move task using the command
		moveCmd := InProgressCmd()
		_, err = cli.ExecuteCLICommand(t, app, moveCmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})
		assert.NoError(t, err)

		// Now list in JSON mode
		listCmd := InProgressCmd()
		output, err := cli.ExecuteCLICommand(t, app, listCmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["tasks"])
		assert.NotNil(t, result["count"])

		tasks := result["tasks"].([]interface{})
		assert.GreaterOrEqual(t, len(tasks), 1, "Should have at least 1 task")

		// Verify task structure
		foundTask := false
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]interface{})
			if int(taskData["id"].(float64)) == taskID {
				foundTask = true
				assert.Equal(t, "JSON List Task", taskData["title"])
				assert.Equal(t, float64(7), taskData["ticket_number"])
				break
			}
		}
		assert.True(t, foundTask, "Should find the created task in list")
	})

	t.Run("Task already in in-progress column", func(t *testing.T) {
		// Create task already in in-progress column
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Already In Progress")

		// Assign ticket number
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 8 WHERE id = ?", taskID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		// Try to move to in-progress again (should handle gracefully)
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		// Should not error - command handles this gracefully
		assert.NoError(t, err)
		// In quiet mode, still outputs the task ID even if already in target column
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)
	})

	t.Run("Multiple tasks moved to in-progress", func(t *testing.T) {
		// Create multiple tasks
		taskID1 := cli.CreateTestTask(t, db, todoColumnID, "Multi Task 1")
		taskID2 := cli.CreateTestTask(t, db, todoColumnID, "Multi Task 2")
		taskID3 := cli.CreateTestTask(t, db, todoColumnID, "Multi Task 3")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 9 WHERE id = ?", taskID1)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 10 WHERE id = ?", taskID2)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 11 WHERE id = ?", taskID3)
		assert.NoError(t, err)

		cmd1 := InProgressCmd()
		cmd2 := InProgressCmd()
		cmd3 := InProgressCmd()

		// Move all tasks to in-progress
		_, err = cli.ExecuteCLICommand(t, app, cmd1, []string{
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
			var columnID int
			err = db.QueryRowContext(context.Background(),
				"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
			assert.NoError(t, err)
			assert.Equal(t, inProgressColumnID, columnID)
		}
	})

	t.Run("List in-progress tasks with different priorities", func(t *testing.T) {
		// Create tasks with different priorities
		lowPriorityTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Low Priority Task")
		highPriorityTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "High Priority Task")
		criticalPriorityTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Critical Priority Task")

		// Assign ticket numbers and priorities
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 12, priority_id = 2 WHERE id = ?", lowPriorityTaskID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 13, priority_id = 4 WHERE id = ?", highPriorityTaskID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 14, priority_id = 5 WHERE id = ?", criticalPriorityTaskID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Low Priority Task")
		assert.Contains(t, output, "High Priority Task")
		assert.Contains(t, output, "Critical Priority Task")
	})

	t.Run("List in-progress tasks with blocked status", func(t *testing.T) {
		// Create blocker and blocked tasks
		blockerTaskID := cli.CreateTestTask(t, db, todoColumnID, "Blocker Task")
		blockedTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Blocked In Progress Task")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 15 WHERE id = ?", blockerTaskID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 16 WHERE id = ?", blockedTaskID)
		assert.NoError(t, err)

		// Create blocking relationship (relation_type_id = 2 for blocking)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
			blockedTaskID, blockerTaskID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Blocked In Progress Task")
		// Should show blocked indicator
		assert.Contains(t, output, "BLOCKED")
	})

	t.Run("List empty in-progress tasks", func(t *testing.T) {
		// Create a new project with no in-progress tasks
		newProjectID := cli.CreateTestProject(t, db, "Empty Project")
		emptyInProgressColumnID := cli.CreateTestColumn(t, db, newProjectID, "In Progress")

		// Mark as in-progress column
		_, err := db.ExecContext(context.Background(),
			"UPDATE columns SET holds_in_progress_tasks = true WHERE id = ?", emptyInProgressColumnID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", newProjectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No in-progress tasks found")
	})

	t.Run("Move task with labels to in-progress", func(t *testing.T) {
		// Create task with labels
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task With Labels")

		// Assign ticket number
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 17 WHERE id = ?", taskID)
		assert.NoError(t, err)

		// Create and attach labels
		labelID1 := testutil.CreateTestLabel(t, db, projectID, "bug", "#EF4444")
		labelID2 := testutil.CreateTestLabel(t, db, projectID, "urgent", "#F97316")

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID1)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID2)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved to in-progress column
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, inProgressColumnID, columnID)

		// Verify labels are still attached
		var labelCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", taskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, labelCount)
	})

	t.Run("Move task from different column to in-progress", func(t *testing.T) {
		// Create a "Done" column
		doneColumnID := cli.CreateTestColumn(t, db, projectID, "Done")

		// Create task in Done column
		taskID := cli.CreateTestTask(t, db, doneColumnID, "Task From Done")

		// Assign ticket number
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 18 WHERE id = ?", taskID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved from Done to In Progress
		var columnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id FROM tasks WHERE id = ?", taskID).Scan(&columnID)
		assert.NoError(t, err)
		assert.Equal(t, inProgressColumnID, columnID)
	})

	t.Run("Move task with description to in-progress", func(t *testing.T) {
		// Create task with description
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task With Description")

		description := "This is a detailed description of the task that needs to be completed."
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET description = ?, ticket_number = 19 WHERE id = ?", description, taskID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Task %d moved to 'In Progress'", taskID))

		// Verify task moved and description is preserved
		var columnID int
		var savedDescription string
		err = db.QueryRowContext(context.Background(),
			"SELECT column_id, description FROM tasks WHERE id = ?", taskID).Scan(&columnID, &savedDescription)
		assert.NoError(t, err)
		assert.Equal(t, inProgressColumnID, columnID)
		assert.Equal(t, description, savedDescription)
	})

	t.Run("List in-progress tasks in JSON with complete structure", func(t *testing.T) {
		// Create a task with all metadata
		taskID := cli.CreateTestTask(t, db, inProgressColumnID, "Complete Metadata Task")

		_, err := db.ExecContext(context.Background(), `
			UPDATE tasks 
			SET description = ?, 
			    ticket_number = ?,
			    type_id = 2, 
			    priority_id = 4
			WHERE id = ?`,
			"Complete description", 20, taskID)
		assert.NoError(t, err)

		cmd := InProgressCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify structure
		assert.True(t, result["success"].(bool))
		tasks := result["tasks"].([]interface{})

		// Find our task
		foundTask := false
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]interface{})
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
