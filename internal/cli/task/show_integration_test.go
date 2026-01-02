package task

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestShowTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	var todoColumnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&todoColumnID)
	assert.NoError(t, err)

	t.Run("Show task with ID flag", func(t *testing.T) {
		// Create a task with description
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Test Task")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET description = ? WHERE id = ?",
			"Test Description", taskID)
		assert.NoError(t, err)

		// Assign ticket number
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 1 WHERE id = ?", taskID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Test Task")
		assert.Contains(t, output, "Test Description")
		assert.Contains(t, output, "Test Project-1")
	})

	t.Run("Show task with positional argument", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Another Task")

		// Assign ticket number
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 2 WHERE id = ?", taskID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		// Pass task ID as positional argument (no --id flag)
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Another Task")
		assert.Contains(t, output, "Test Project-2")
	})

	t.Run("Show task in quiet mode", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Task")

		cmd := ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should only output the task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)
	})

	t.Run("Show task in JSON mode", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "JSON Task")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET description = ?, ticket_number = 3 WHERE id = ?",
			"JSON Description", taskID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))
		taskData := result["task"].(map[string]interface{})
		assert.Equal(t, "JSON Task", taskData["title"])
		assert.Equal(t, "JSON Description", taskData["description"])
		assert.Equal(t, "Test Project", taskData["project_name"])
		assert.Equal(t, float64(3), taskData["ticket_number"])
	})

	t.Run("Show task with labels", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task with Labels")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 4 WHERE id = ?", taskID)
		assert.NoError(t, err)

		// Create labels
		labelID1 := testutil.CreateTestLabel(t, db, projectID, "bug", "#EF4444")
		labelID2 := testutil.CreateTestLabel(t, db, projectID, "urgent", "#F97316")

		// Attach labels to task
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID1)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID2)
		assert.NoError(t, err)

		cmd := ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Task with Labels")
		assert.Contains(t, output, "Labels")
		// Note: The actual rendering of labels depends on the styles package
		// We just verify the Labels section appears
	})

	t.Run("Show task with parent relationship", func(t *testing.T) {
		// Create parent task
		parentID := cli.CreateTestTask(t, db, todoColumnID, "Parent Task")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 5 WHERE id = ?", parentID)
		assert.NoError(t, err)

		// Create child task
		childID := cli.CreateTestTask(t, db, todoColumnID, "Child Task")
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 6 WHERE id = ?", childID)
		assert.NoError(t, err)

		// Create relationship (parent-child, non-blocking)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
			parentID, childID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		// Show parent task - should display child
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", parentID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Parent Task")
		assert.Contains(t, output, "Child Tasks")
		assert.Contains(t, output, "Child Task")
	})

	t.Run("Show task with blocking relationship", func(t *testing.T) {
		// Create blocker task
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "Blocker Task")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 7 WHERE id = ?", blockerID)
		assert.NoError(t, err)

		// Create blocked task
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "Blocked Task")
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 8 WHERE id = ?", blockedID)
		assert.NoError(t, err)

		// Create blocking relationship (relation_type_id = 2 for blocking)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
			blockedID, blockerID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		// Show blocked task - should show BLOCKED indicator and "Blocked By" section
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", blockedID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Blocked Task")
		assert.Contains(t, output, "BLOCKED")
		assert.Contains(t, output, "Blocked By")
		assert.Contains(t, output, "Blocker Task")
	})

	t.Run("Show task with all metadata", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Full Metadata Task")

		// Update with all metadata fields
		_, err := db.ExecContext(context.Background(), `
			UPDATE tasks 
			SET description = ?, 
			    ticket_number = ?,
			    type_id = 2, 
			    priority_id = 4
			WHERE id = ?`,
			"Full description with details", 9, taskID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		// Verify task title
		assert.Contains(t, output, "Full Metadata Task")
		// Verify description
		assert.Contains(t, output, "Full description with details")
		// Verify type (feature)
		assert.Contains(t, output, "feature")
		// Verify priority (high)
		assert.Contains(t, output, "high")
		// Verify column
		assert.Contains(t, output, "Todo")
		// Verify timestamps are present
		assert.Contains(t, output, "Created:")
		assert.Contains(t, output, "Updated:")
	})

	t.Run("Show task in JSON mode with complete structure", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Complete JSON Task")
		_, err := db.ExecContext(context.Background(), `
			UPDATE tasks 
			SET description = ?, 
			    ticket_number = ?,
			    type_id = 3, 
			    priority_id = 5
			WHERE id = ?`,
			"Complete description", 10, taskID)
		assert.NoError(t, err)

		// Add a label
		labelID := testutil.CreateTestLabel(t, db, projectID, "feature", "#3B82F6")
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify complete JSON structure
		assert.True(t, result["success"].(bool))
		taskData := result["task"].(map[string]interface{})

		// Verify all fields
		assert.Equal(t, float64(taskID), taskData["id"])
		assert.Equal(t, "Complete JSON Task", taskData["title"])
		assert.Equal(t, "Complete description", taskData["description"])
		assert.Equal(t, "bug", taskData["type"])
		assert.Equal(t, float64(10), taskData["ticket_number"])
		assert.Equal(t, "Test Project", taskData["project_name"])

		// Verify priority structure
		priority := taskData["priority"].(map[string]interface{})
		assert.Equal(t, "critical", priority["name"])
		assert.Equal(t, "#EF4444", priority["color"])

		// Verify column structure
		column := taskData["column"].(map[string]interface{})
		assert.Equal(t, "Todo", column["name"])

		// Verify labels array
		labels := taskData["labels"].([]interface{})
		assert.Len(t, labels, 1)

		// Verify timestamps exist
		assert.NotNil(t, taskData["created_at"])
		assert.NotNil(t, taskData["updated_at"])
	})

	t.Run("Show task with multi-line description", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Multi-line Task")

		multiLineDesc := `This is a multi-line description.
It spans multiple lines.
Each line should be properly displayed.`

		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET description = ?, ticket_number = 11 WHERE id = ?",
			multiLineDesc, taskID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Multi-line Task")
		assert.Contains(t, output, "This is a multi-line description.")
		assert.Contains(t, output, "It spans multiple lines.")
		assert.Contains(t, output, "Each line should be properly displayed.")
	})

	t.Run("Show task with position information", func(t *testing.T) {
		// Create multiple tasks to verify position
		task1ID := cli.CreateTestTask(t, db, todoColumnID, "Position Task 1")
		task2ID := cli.CreateTestTask(t, db, todoColumnID, "Position Task 2")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 12 WHERE id = ?", task1ID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 13 WHERE id = ?", task2ID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		// Show second task and verify it's in JSON with position
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", task2ID),
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		taskData := result["task"].(map[string]interface{})
		// Position should be greater than 0 (second task in column)
		assert.NotNil(t, taskData["position"])
	})

	t.Run("Show task with empty description", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "No Description Task")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 14 WHERE id = ?", taskID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No Description Task")
		// Verify the task appears correctly (basic smoke test for empty description)
		assert.Contains(t, output, "Type:")
		assert.Contains(t, output, "Priority:")
	})

	t.Run("Show task with both parent and child relationships", func(t *testing.T) {
		// Create a task with both parent and child relationships
		middleTaskID := cli.CreateTestTask(t, db, todoColumnID, "Middle Task")
		parentTaskID := cli.CreateTestTask(t, db, todoColumnID, "Parent of Middle")
		childTaskID := cli.CreateTestTask(t, db, todoColumnID, "Child of Middle")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 15 WHERE id = ?", middleTaskID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 16 WHERE id = ?", parentTaskID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 17 WHERE id = ?", childTaskID)
		assert.NoError(t, err)

		// Middle task is child of parentTask
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
			parentTaskID, middleTaskID)
		assert.NoError(t, err)

		// childTask is child of middleTask
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
			middleTaskID, childTaskID)
		assert.NoError(t, err)

		cmd := ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", middleTaskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Middle Task")
		assert.Contains(t, output, "Parent Tasks")
		assert.Contains(t, output, "Parent of Middle")
		assert.Contains(t, output, "Child Tasks")
		assert.Contains(t, output, "Child of Middle")
	})
}
