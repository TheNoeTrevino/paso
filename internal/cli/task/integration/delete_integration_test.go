package task_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestDeleteTask_Integration(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	ctx := context.Background()

	// Create test project with default columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the first column ID (Todo column)
	columnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("delete task with force flag", func(t *testing.T) {
		// Create a task to delete
		taskID := cli.CreateTestTask(t, db, columnID, "Task to Delete")

		cmd := task.DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")
		assert.Contains(t, output, fmt.Sprintf("%d", taskID))

		// Verify task is gone from DB
		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "Task should be deleted from database")
	})

	t.Run("delete task with quiet flag", func(t *testing.T) {
		// Create a task to delete
		taskID := cli.CreateTestTask(t, db, columnID, "Task to Delete Quietly")

		cmd := task.DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, "", output, "Quiet mode should produce no output")

		// Verify task is gone from DB
		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "Task should be deleted from database")
	})

	t.Run("delete task with json flag", func(t *testing.T) {
		// Create a task to delete
		taskID := cli.CreateTestTask(t, db, columnID, "Task to Delete with JSON")

		cmd := task.DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--json",
			"--force",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(strings.TrimSpace(output)), &result)
		assert.NoError(t, err, "Output should be valid JSON")

		// Verify JSON structure
		assert.Equal(t, true, result["success"], "JSON should contain success=true")
		assert.Equal(t, float64(taskID), result["task_id"], "JSON should contain correct task_id")

		// Verify task is gone from DB
		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "Task should be deleted from database")
	})

	t.Run("delete task with parent-child relationships", func(t *testing.T) {
		// Create parent and child tasks
		parentID := cli.CreateTestTask(t, db, columnID, "Parent Task")
		childID := cli.CreateTestTask(t, db, columnID, "Child Task")

		// Create parent-child relationship (relation_type_id = 1 for Parent/Child)
		cli.AddTaskSubtask(t, db, parentID, childID, 1)

		// Verify relationship exists
		var relCount int
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentID, childID).Scan(&relCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, relCount, "Relationship should exist before deletion")

		// Delete parent task
		cmd := task.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", parentID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify parent task is deleted
		var taskCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", parentID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Parent task should be deleted")

		// Verify relationship was cascade deleted
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? OR child_id = ?",
			parentID, parentID).Scan(&relCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, relCount, "Relationships should be cascade deleted")

		// Verify child task still exists (only relationship should be deleted)
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", childID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, taskCount, "Child task should still exist")
	})

	t.Run("delete task with blocking relationships", func(t *testing.T) {
		// Create tasks with blocking relationship
		blockerID := cli.CreateTestTask(t, db, columnID, "Blocker Task")
		blockedID := cli.CreateTestTask(t, db, columnID, "Blocked Task")

		// Create blocking relationship (relation_type_id = 2 for Blocks/Blocked By)
		cli.AddTaskSubtask(t, db, blockerID, blockedID, 2)

		// Delete blocker task
		cmd := task.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", blockerID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify blocker task is deleted
		var taskCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", blockerID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Blocker task should be deleted")

		// Verify blocking relationship was cascade deleted
		var relCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? OR child_id = ?",
			blockerID, blockerID).Scan(&relCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, relCount, "Blocking relationships should be cascade deleted")

		// Verify blocked task still exists
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", blockedID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, taskCount, "Blocked task should still exist")
	})

	t.Run("delete task with labels", func(t *testing.T) {
		// Create a task
		taskID := cli.CreateTestTask(t, db, columnID, "Task with Labels")

		// Create label and attach to task
		labelID := cli.CreateTestLabel(t, db, projectID, "bug", "#FF0000")
		cli.AttachLabelToTask(t, db, taskID, labelID)

		// Verify label attachment exists
		var labelCount int
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", taskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, labelCount, "Label should be attached to task")

		// Delete task
		cmd := task.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify task is deleted
		var taskCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Task should be deleted")

		// Verify label attachment was cascade deleted
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", taskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, labelCount, "Label attachments should be cascade deleted")

		// Verify label itself still exists
		var labelStillExists int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM labels WHERE id = ?", labelID).Scan(&labelStillExists)
		assert.NoError(t, err)
		assert.Equal(t, 1, labelStillExists, "Label should still exist after task deletion")
	})

	t.Run("delete multiple tasks", func(t *testing.T) {
		// Create multiple tasks
		taskIDs := make([]int, 3)
		for i := 0; i < 3; i++ {
			taskIDs[i] = cli.CreateTestTask(t, db, columnID, fmt.Sprintf("Task to Delete %d", i+1))
		}

		// Delete all tasks
		cmd := task.DeleteCmd()
		for _, taskID := range taskIDs {
			output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				fmt.Sprintf("%d", taskID),
				"--quiet",
			})

			assert.NoError(t, err)
			assert.Equal(t, "", output, "Quiet mode should produce no output")
		}

		// Verify all tasks are deleted
		for _, taskID := range taskIDs {
			var count int
			err := db.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, 0, count, fmt.Sprintf("Task %d should be deleted", taskID))
		}
	})

	t.Run("delete task with comments", func(t *testing.T) {
		// Create a task
		taskID := cli.CreateTestTask(t, db, columnID, "Task with Comments")

		// Add comments to task
		cli.CreateTestComment(t, db, taskID, "First comment", "testuser")
		cli.CreateTestComment(t, db, taskID, "Second comment", "testuser")

		// Verify comments exist
		var commentCount int
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_comments WHERE task_id = ?", taskID).Scan(&commentCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, commentCount, "Task should have 2 comments")

		// Delete task
		cmd := task.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify task is deleted
		var taskCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Task should be deleted")

		// Verify comments were cascade deleted
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_comments WHERE task_id = ?", taskID).Scan(&commentCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, commentCount, "Comments should be cascade deleted")
	})

	t.Run("delete task with complex relationships", func(t *testing.T) {
		// Create a complex task structure:
		// - Task with labels, comments, and both parent and child relationships
		mainTaskID := cli.CreateTestTask(t, db, columnID, "Main Task")
		parentTaskID := cli.CreateTestTask(t, db, columnID, "Parent Task")
		childTaskID := cli.CreateTestTask(t, db, columnID, "Child Task")

		// Add parent relationship
		cli.AddTaskSubtask(t, db, parentTaskID, mainTaskID, 1)
		// Add child relationship
		cli.AddTaskSubtask(t, db, mainTaskID, childTaskID, 1)

		// Add label
		labelID := cli.CreateTestLabel(t, db, projectID, "feature", "#00FF00")
		cli.AttachLabelToTask(t, db, mainTaskID, labelID)

		// Add comment
		cli.CreateTestComment(t, db, mainTaskID, "Important comment", "testuser")

		// Delete main task
		cmd := task.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", mainTaskID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify main task is deleted
		var taskCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id = ?", mainTaskID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Main task should be deleted")

		// Verify all relationships involving main task are deleted
		var relCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? OR child_id = ?",
			mainTaskID, mainTaskID).Scan(&relCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, relCount, "All relationships should be cascade deleted")

		// Verify label attachment is deleted
		var labelCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", mainTaskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, labelCount, "Label attachments should be cascade deleted")

		// Verify comments are deleted
		var commentCount int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_comments WHERE task_id = ?", mainTaskID).Scan(&commentCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, commentCount, "Comments should be cascade deleted")

		// Verify parent and child tasks still exist
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE id IN (?, ?)", parentTaskID, childTaskID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, taskCount, "Parent and child tasks should still exist")
	})
}

func TestDeleteTask_Integration_Errors(t *testing.T) {
	t.Parallel()

	_, app := cli.SetupCLITest(t)

	t.Run("missing task ID", func(t *testing.T) {
		cmd := task.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--force"})
		assert.Error(t, err)
	})

	t.Run("invalid task ID 'abc'", func(t *testing.T) {
		cmd := task.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"abc", "--force"})
		cli.AssertExitError(t, err, rootcli.ExitValidation)
		assert.Contains(t, err.Error(), "invalid ID 'abc'")
	})

	t.Run("non-existent ID 999999", func(t *testing.T) {
		cmd := task.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"999999", "--force"})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "task 999999 not found")
	})

	t.Run("negative task ID -1", func(t *testing.T) {
		cmd := task.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"-1", "--force"})
		assert.Error(t, err)
	})
}
