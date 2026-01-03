package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestDeleteTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project with default columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the first column ID (Todo column)
	var columnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? LIMIT 1", projectID).Scan(&columnID)
	assert.NoError(t, err)

	t.Run("Delete task with force flag", func(t *testing.T) {
		// Create a task to delete
		taskID := cli.CreateTestTask(t, db, columnID, "Task to Delete")

		cmd := DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")
		assert.Contains(t, output, fmt.Sprintf("%d", taskID))

		// Verify task is gone from DB
		var count int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "Task should be deleted from database")
	})

	t.Run("Delete task with quiet flag", func(t *testing.T) {
		// Create a task to delete
		taskID := cli.CreateTestTask(t, db, columnID, "Task to Delete Quietly")

		cmd := DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, "", output, "Quiet mode should produce no output")

		// Verify task is gone from DB
		var count int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "Task should be deleted from database")
	})

	t.Run("Delete task with json flag", func(t *testing.T) {
		// Create a task to delete
		taskID := cli.CreateTestTask(t, db, columnID, "Task to Delete with JSON")

		cmd := DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--json",
			"--force",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(strings.TrimSpace(output)), &result)
		assert.NoError(t, err, "Output should be valid JSON")

		// Verify JSON structure
		assert.Equal(t, true, result["success"], "JSON should contain success=true")
		assert.Equal(t, float64(taskID), result["task_id"], "JSON should contain correct task_id")

		// Verify task is gone from DB
		var count int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "Task should be deleted from database")
	})

	t.Run("Delete task with parent-child relationships", func(t *testing.T) {
		// Create parent and child tasks
		parentID := cli.CreateTestTask(t, db, columnID, "Parent Task")
		childID := cli.CreateTestTask(t, db, columnID, "Child Task")

		// Create parent-child relationship (relation_type_id = 1 for Parent/Child)
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
			parentID, childID)
		assert.NoError(t, err)

		// Verify relationship exists
		var relCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentID, childID).Scan(&relCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, relCount, "Relationship should exist before deletion")

		// Delete parent task
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", parentID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify parent task is deleted
		var taskCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", parentID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Parent task should be deleted")

		// Verify relationship was cascade deleted
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? OR child_id = ?",
			parentID, parentID).Scan(&relCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, relCount, "Relationships should be cascade deleted")

		// Verify child task still exists (only relationship should be deleted)
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", childID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, taskCount, "Child task should still exist")
	})

	t.Run("Delete task with blocking relationships", func(t *testing.T) {
		// Create tasks with blocking relationship
		blockerID := cli.CreateTestTask(t, db, columnID, "Blocker Task")
		blockedID := cli.CreateTestTask(t, db, columnID, "Blocked Task")

		// Create blocking relationship (relation_type_id = 2 for Blocks/Blocked By)
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
			blockerID, blockedID)
		assert.NoError(t, err)

		// Delete blocker task
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", blockerID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify blocker task is deleted
		var taskCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", blockerID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Blocker task should be deleted")

		// Verify blocking relationship was cascade deleted
		var relCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? OR child_id = ?",
			blockerID, blockerID).Scan(&relCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, relCount, "Blocking relationships should be cascade deleted")

		// Verify blocked task still exists
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", blockedID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, taskCount, "Blocked task should still exist")
	})

	t.Run("Delete task with labels", func(t *testing.T) {
		// Create a task
		taskID := cli.CreateTestTask(t, db, columnID, "Task with Labels")

		// Create labels and attach to task
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO labels (project_id, name, color) VALUES (?, ?, ?)",
			projectID, "bug", "#FF0000")
		assert.NoError(t, err)

		var labelID int
		err = db.QueryRowContext(context.Background(),
			"SELECT id FROM labels WHERE name = ?", "bug").Scan(&labelID)
		assert.NoError(t, err)

		// Attach label to task
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)",
			taskID, labelID)
		assert.NoError(t, err)

		// Verify label attachment exists
		var labelCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", taskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, labelCount, "Label should be attached to task")

		// Delete task
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify task is deleted
		var taskCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Task should be deleted")

		// Verify label attachment was cascade deleted
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", taskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, labelCount, "Label attachments should be cascade deleted")

		// Verify label itself still exists
		var labelStillExists int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM labels WHERE id = ?", labelID).Scan(&labelStillExists)
		assert.NoError(t, err)
		assert.Equal(t, 1, labelStillExists, "Label should still exist after task deletion")
	})

	t.Run("Delete multiple tasks", func(t *testing.T) {
		// Create multiple tasks
		taskIDs := make([]int, 3)
		for i := 0; i < 3; i++ {
			taskIDs[i] = cli.CreateTestTask(t, db, columnID, fmt.Sprintf("Task to Delete %d", i+1))
		}

		// Delete all tasks
		cmd := DeleteCmd()
		for _, taskID := range taskIDs {
			output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				"--id", fmt.Sprintf("%d", taskID),
				"--quiet",
			})

			assert.NoError(t, err)
			assert.Equal(t, "", output, "Quiet mode should produce no output")
		}

		// Verify all tasks are deleted
		for _, taskID := range taskIDs {
			var count int
			err := db.QueryRowContext(context.Background(),
				"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, 0, count, fmt.Sprintf("Task %d should be deleted", taskID))
		}
	})

	t.Run("Delete task with comments", func(t *testing.T) {
		// Create a task
		taskID := cli.CreateTestTask(t, db, columnID, "Task with Comments")

		// Add comments to task
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_comments (task_id, content, author) VALUES (?, ?, ?)",
			taskID, "First comment", "testuser")
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_comments (task_id, content, author) VALUES (?, ?, ?)",
			taskID, "Second comment", "testuser")
		assert.NoError(t, err)

		// Verify comments exist
		var commentCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_comments WHERE task_id = ?", taskID).Scan(&commentCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, commentCount, "Task should have 2 comments")

		// Delete task
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify task is deleted
		var taskCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Task should be deleted")

		// Verify comments were cascade deleted
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_comments WHERE task_id = ?", taskID).Scan(&commentCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, commentCount, "Comments should be cascade deleted")
	})

	t.Run("Delete task with complex relationships", func(t *testing.T) {
		// Create a complex task structure:
		// - Task with labels, comments, and both parent and child relationships
		mainTaskID := cli.CreateTestTask(t, db, columnID, "Main Task")
		parentTaskID := cli.CreateTestTask(t, db, columnID, "Parent Task")
		childTaskID := cli.CreateTestTask(t, db, columnID, "Child Task")

		// Add parent relationship
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
			parentTaskID, mainTaskID)
		assert.NoError(t, err)

		// Add child relationship
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
			mainTaskID, childTaskID)
		assert.NoError(t, err)

		// Add label
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO labels (project_id, name, color) VALUES (?, ?, ?)",
			projectID, "feature", "#00FF00")
		assert.NoError(t, err)

		var labelID int
		err = db.QueryRowContext(context.Background(),
			"SELECT id FROM labels WHERE name = ?", "feature").Scan(&labelID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)",
			mainTaskID, labelID)
		assert.NoError(t, err)

		// Add comment
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_comments (task_id, content, author) VALUES (?, ?, ?)",
			mainTaskID, "Important comment", "testuser")
		assert.NoError(t, err)

		// Delete main task
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", mainTaskID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		// Verify main task is deleted
		var taskCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id = ?", mainTaskID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, taskCount, "Main task should be deleted")

		// Verify all relationships involving main task are deleted
		var relCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? OR child_id = ?",
			mainTaskID, mainTaskID).Scan(&relCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, relCount, "All relationships should be cascade deleted")

		// Verify label attachment is deleted
		var labelCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ?", mainTaskID).Scan(&labelCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, labelCount, "Label attachments should be cascade deleted")

		// Verify comments are deleted
		var commentCount int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_comments WHERE task_id = ?", mainTaskID).Scan(&commentCount)
		assert.NoError(t, err)
		assert.Equal(t, 0, commentCount, "Comments should be cascade deleted")

		// Verify parent and child tasks still exist
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM tasks WHERE id IN (?, ?)", parentTaskID, childTaskID).Scan(&taskCount)
		assert.NoError(t, err)
		assert.Equal(t, 2, taskCount, "Parent and child tasks should still exist")
	})
}
