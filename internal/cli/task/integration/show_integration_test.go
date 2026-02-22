package task_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestShowTask_Integration(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("show task with ID flag", func(t *testing.T) {
		// Create a task with description
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Test Task")
		cli.UpdateTaskFields(t, db, taskID, map[string]any{
			"description":   "Test Description",
			"task_number": 1,
		})

		cmd := task.ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Test Task")
		assert.Contains(t, output, "Test Description")
		assert.Contains(t, output, "Test Project-1")
	})

	t.Run("show task with positional argument", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Another Task")
		cli.UpdateTaskFields(t, db, taskID, map[string]any{"task_number": 2})

		cmd := task.ShowCmd()

		// Pass task ID as positional argument (no --id flag)
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Another Task")
		assert.Contains(t, output, "Test Project-2")
	})

	t.Run("show task in quiet mode", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Task")

		cmd := task.ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should only output the task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)
	})

	t.Run("show task in JSON mode", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "JSON Task")
		cli.UpdateTaskFields(t, db, taskID, map[string]any{
			"description":   "JSON Description",
			"task_number": 3,
		})

		cmd := task.ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))
		taskData := result["task"].(map[string]any)
		assert.Equal(t, "JSON Task", taskData["title"])
		assert.Equal(t, "JSON Description", taskData["description"])
		assert.Equal(t, "Test Project", taskData["project_name"])
		assert.Equal(t, float64(3), taskData["task_number"])
	})

	t.Run("show task with labels", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Task with Labels")
		cli.UpdateTaskFields(t, db, taskID, map[string]any{"task_number": 4})

		// Create labels and attach to task
		labelID1 := cli.CreateTestLabel(t, db, projectID, "bug", "#EF4444")
		labelID2 := cli.CreateTestLabel(t, db, projectID, "urgent", "#F97316")
		cli.AttachLabelToTask(t, db, taskID, labelID1)
		cli.AttachLabelToTask(t, db, taskID, labelID2)

		cmd := task.ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Task with Labels")
		assert.Contains(t, output, "Labels")
		// Note: The actual rendering of labels depends on the styles package
		// We just verify the Labels section appears
	})

	t.Run("show task with parent relationship", func(t *testing.T) {
		// Create parent task
		parentID := cli.CreateTestTask(t, db, todoColumnID, "Parent Task")
		cli.UpdateTaskFields(t, db, parentID, map[string]any{"task_number": 5})

		// Create child task
		childID := cli.CreateTestTask(t, db, todoColumnID, "Child Task")
		cli.UpdateTaskFields(t, db, childID, map[string]any{"task_number": 6})

		// Create relationship (parent-child, non-blocking)
		cli.AddTaskSubtask(t, db, parentID, childID, 1)

		cmd := task.ShowCmd()

		// Show parent task - should display child
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", parentID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Parent Task")
		assert.Contains(t, output, "Child Tasks")
		assert.Contains(t, output, "Child Task")
	})

	t.Run("show task with blocking relationship", func(t *testing.T) {
		// Create blocker task
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "Blocker Task")
		cli.UpdateTaskFields(t, db, blockerID, map[string]any{"task_number": 7})

		// Create blocked task
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "Blocked Task")
		cli.UpdateTaskFields(t, db, blockedID, map[string]any{"task_number": 8})

		// Create blocking relationship (relation_type_id = 2 for blocking)
		cli.AddTaskSubtask(t, db, blockedID, blockerID, 2)

		cmd := task.ShowCmd()

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

	t.Run("show task with all metadata", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Full Metadata Task")
		cli.UpdateTaskFields(t, db, taskID, map[string]any{
			"description":   "Full description with details",
			"task_number": 9,
			"type_id":       2,
			"priority_id":   4,
		})

		cmd := task.ShowCmd()

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

	t.Run("show task in JSON mode with complete structure", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Complete JSON Task")
		cli.UpdateTaskFields(t, db, taskID, map[string]any{
			"description":   "Complete description",
			"task_number": 10,
			"type_id":       3,
			"priority_id":   5,
		})

		// Add a label
		labelID := cli.CreateTestLabel(t, db, projectID, "feature", "#3B82F6")
		cli.AttachLabelToTask(t, db, taskID, labelID)

		cmd := task.ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify complete JSON structure
		assert.True(t, result["success"].(bool))
		taskData := result["task"].(map[string]any)

		// Verify all fields
		assert.Equal(t, float64(taskID), taskData["id"])
		assert.Equal(t, "Complete JSON Task", taskData["title"])
		assert.Equal(t, "Complete description", taskData["description"])
		assert.Equal(t, "bug", taskData["type"])
		assert.Equal(t, float64(10), taskData["task_number"])
		assert.Equal(t, "Test Project", taskData["project_name"])

		// Verify priority structure
		priority := taskData["priority"].(map[string]any)
		assert.Equal(t, "critical", priority["name"])
		assert.Equal(t, "#EF4444", priority["color"])

		// Verify column structure
		column := taskData["column"].(map[string]any)
		assert.Equal(t, "Todo", column["name"])

		// Verify labels array
		labels := taskData["labels"].([]any)
		assert.Len(t, labels, 1)

		// Verify timestamps exist
		assert.NotNil(t, taskData["created_at"])
		assert.NotNil(t, taskData["updated_at"])
	})

	t.Run("show task with multi-line description", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Multi-line Task")

		multiLineDesc := `This is a multi-line description.
It spans multiple lines.
Each line should be properly displayed.`

		cli.UpdateTaskFields(t, db, taskID, map[string]any{
			"description":   multiLineDesc,
			"task_number": 11,
		})

		cmd := task.ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Multi-line Task")
		assert.Contains(t, output, "This is a multi-line description.")
		assert.Contains(t, output, "It spans multiple lines.")
		assert.Contains(t, output, "Each line should be properly displayed.")
	})

	t.Run("show task with position information", func(t *testing.T) {
		// Create multiple tasks to verify position
		task1ID := cli.CreateTestTask(t, db, todoColumnID, "Position Task 1")
		task2ID := cli.CreateTestTask(t, db, todoColumnID, "Position Task 2")

		cli.UpdateTaskFields(t, db, task1ID, map[string]any{"task_number": 12})
		cli.UpdateTaskFields(t, db, task2ID, map[string]any{"task_number": 13})

		cmd := task.ShowCmd()

		// Show second task and verify it's in JSON with position
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", task2ID),
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		taskData := result["task"].(map[string]any)
		// Position should be greater than 0 (second task in column)
		assert.NotNil(t, taskData["position"])
	})

	t.Run("show task with empty description", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "No Description Task")
		cli.UpdateTaskFields(t, db, taskID, map[string]any{"task_number": 14})

		cmd := task.ShowCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No Description Task")
		// Verify the task appears correctly (basic smoke test for empty description)
		assert.Contains(t, output, "Type:")
		assert.Contains(t, output, "Priority:")
	})

	t.Run("show task with both parent and child relationships", func(t *testing.T) {
		// Create a task with both parent and child relationships
		middleTaskID := cli.CreateTestTask(t, db, todoColumnID, "Middle Task")
		parentTaskID := cli.CreateTestTask(t, db, todoColumnID, "Parent of Middle")
		childTaskID := cli.CreateTestTask(t, db, todoColumnID, "Child of Middle")

		cli.UpdateTaskFields(t, db, middleTaskID, map[string]any{"task_number": 15})
		cli.UpdateTaskFields(t, db, parentTaskID, map[string]any{"task_number": 16})
		cli.UpdateTaskFields(t, db, childTaskID, map[string]any{"task_number": 17})

		// Middle task is child of parentTask
		cli.AddTaskSubtask(t, db, parentTaskID, middleTaskID, 1)
		// childTask is child of middleTask
		cli.AddTaskSubtask(t, db, middleTaskID, childTaskID, 1)

		cmd := task.ShowCmd()

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

func TestShowTask_Integration_Errors(t *testing.T) {
	t.Parallel()

	_, app := cli.SetupCLITest(t)

	t.Run("missing task ID", func(t *testing.T) {
		cmd := task.ShowCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		cli.AssertExitError(t, err, rootcli.ExitUsage)
		assert.Contains(t, err.Error(), "task ID must be a positive integer")
	})

	t.Run("invalid task ID 'abc'", func(t *testing.T) {
		cmd := task.ShowCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"abc"})
		cli.AssertExitError(t, err, rootcli.ExitUsage)
		assert.Contains(t, err.Error(), "task ID must be a positive integer")
	})

	t.Run("non-existent task ID 999999", func(t *testing.T) {
		cmd := task.ShowCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--id", "999999"})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "task 999999 not found")
	})

	t.Run("negative task ID -1", func(t *testing.T) {
		cmd := task.ShowCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"-1"})
		assert.Error(t, err)
	})
}
