package task_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestBlockedTask(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	// Create test project with columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	// Create an "In Progress" column
	inProgressColumnID := cli.CreateTestColumn(t, db, projectID, "In Progress")

	t.Run("list blocked tasks with blocking relationships", func(t *testing.T) {
		t.Parallel()
		// Create parent task
		parentID := cli.CreateTestTask(t, db, todoColumnID, "Parent Task")
		cli.UpdateTaskFields(t, db, parentID, map[string]any{"ticket_number": 1})

		// Create child task (blocker)
		childID := cli.CreateTestTask(t, db, todoColumnID, "Child Task (Blocker)")
		cli.UpdateTaskFields(t, db, childID, map[string]any{"ticket_number": 2})

		// Create blocking relationship (parentID blocked by childID)
		linkCmd := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(parentID), "--child", strconv.Itoa(childID), "--blocker"})
		assert.NoError(t, err)

		// Now parentID should appear in blocked list
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Parent Task")
		assert.Contains(t, output, "BLOCKED")
		assert.Contains(t, output, "Found 1 blocked tasks")
	})

	t.Run("list blocked tasks with no blocked tasks", func(t *testing.T) {
		t.Parallel()
		// Create a new project with no blocking relationships
		newProjectID := cli.CreateTestProject(t, db, "Empty Project")

		// Get the default "Todo" column ID
		emptyTodoColumnID := cli.GetColumnIDByName(t, db, newProjectID, "Todo")

		// Create tasks but no blocking relationships
		cli.CreateTestTask(t, db, emptyTodoColumnID, "Regular Task 1")
		cli.CreateTestTask(t, db, emptyTodoColumnID, "Regular Task 2")

		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(newProjectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "No blocked tasks found")
	})

	t.Run("list multiple blocked tasks", func(t *testing.T) {
		t.Parallel()
		// Create multiple blocked tasks
		blocked1 := cli.CreateTestTask(t, db, todoColumnID, "Blocked Task 1")
		blocker1 := cli.CreateTestTask(t, db, todoColumnID, "Blocker Task 1")
		blocked2 := cli.CreateTestTask(t, db, todoColumnID, "Blocked Task 2")
		blocker2 := cli.CreateTestTask(t, db, todoColumnID, "Blocker Task 2")

		cli.UpdateTaskFields(t, db, blocked1, map[string]any{"ticket_number": 10})
		cli.UpdateTaskFields(t, db, blocker1, map[string]any{"ticket_number": 11})
		cli.UpdateTaskFields(t, db, blocked2, map[string]any{"ticket_number": 12})
		cli.UpdateTaskFields(t, db, blocker2, map[string]any{"ticket_number": 13})

		// Create blocking relationships
		linkCmd1 := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd1,
			[]string{"--parent", strconv.Itoa(blocked1), "--child", strconv.Itoa(blocker1), "--blocker"})
		assert.NoError(t, err)

		linkCmd2 := task.LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd2,
			[]string{"--parent", strconv.Itoa(blocked2), "--child", strconv.Itoa(blocker2), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Blocked Task 1")
		assert.Contains(t, output, "Blocked Task 2")
		assert.Contains(t, output, "BLOCKED")
	})

	t.Run("list blocked tasks in JSON mode", func(t *testing.T) {
		t.Parallel()
		// Create blocked task
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "JSON Blocked Task")
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "JSON Blocker Task")

		// Assign ticket numbers
		cli.UpdateTaskFields(t, db, blockedID, map[string]any{"ticket_number": 20})
		cli.UpdateTaskFields(t, db, blockerID, map[string]any{"ticket_number": 21})

		// Create blocking relationship
		linkCmd := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedID), "--child", strconv.Itoa(blockerID), "--blocker"})
		assert.NoError(t, err)

		// List in JSON mode
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID), "--json"})

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
		assert.GreaterOrEqual(t, len(tasks), 1, "Should have at least 1 blocked task")

		// Verify task structure
		foundTask := false
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]any)
			if taskData["ID"] != nil && int(taskData["ID"].(float64)) == blockedID {
				foundTask = true
				assert.Equal(t, "JSON Blocked Task", taskData["Title"])
				assert.True(t, taskData["IsBlocked"].(bool))
				break
			}
		}
		assert.True(t, foundTask, "Should find the blocked task in JSON list")
	})

	t.Run("list blocked tasks in quiet mode", func(t *testing.T) {
		t.Parallel()
		// Create blocked task
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Blocked Task")
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Blocker Task")

		// Assign ticket numbers
		cli.UpdateTaskFields(t, db, blockedID, map[string]any{"ticket_number": 30})
		cli.UpdateTaskFields(t, db, blockerID, map[string]any{"ticket_number": 31})

		// Create blocking relationship
		linkCmd := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedID), "--child", strconv.Itoa(blockerID), "--blocker"})
		assert.NoError(t, err)

		// List in quiet mode
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID), "--quiet"})

		assert.NoError(t, err)

		// Quiet mode should only output task IDs, one per line
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.GreaterOrEqual(t, len(lines), 1, "Should have at least 1 task ID")

		// Verify each line is a numeric task ID
		foundBlocked := false
		for _, line := range lines {
			assert.Regexp(t, `^\d+$`, line)
			if line == strconv.Itoa(blockedID) {
				foundBlocked = true
			}
		}
		assert.True(t, foundBlocked, "Should find the blocked task ID in quiet mode")
	})

	t.Run("verify only IsBlocked==true tasks returned", func(t *testing.T) {
		t.Parallel()
		// Create a mix of blocked and non-blocked tasks
		normalTask := cli.CreateTestTask(t, db, todoColumnID, "Normal Task")
		blockedTask := cli.CreateTestTask(t, db, todoColumnID, "Should Be Blocked")
		blockerTask := cli.CreateTestTask(t, db, todoColumnID, "Blocker For Test")

		// Assign ticket numbers
		cli.UpdateTaskFields(t, db, normalTask, map[string]any{"ticket_number": 40})
		cli.UpdateTaskFields(t, db, blockedTask, map[string]any{"ticket_number": 41})
		cli.UpdateTaskFields(t, db, blockerTask, map[string]any{"ticket_number": 42})

		// Create blocking relationship for only one task
		linkCmd := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedTask), "--child", strconv.Itoa(blockerTask), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks in quiet mode
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID), "--quiet"})

		assert.NoError(t, err)

		// Parse output
		lines := strings.Split(strings.TrimSpace(output), "\n")
		ids := make(map[int]bool)
		for _, line := range lines {
			if line != "" {
				id, err := strconv.Atoi(line)
				require.NoError(t, err, "Failed to parse task ID from output")
				ids[id] = true
			}
		}

		// Verify blockedTask is in the list
		assert.True(t, ids[blockedTask], "Blocked task should be in the list")
		// Verify normalTask is NOT in the list
		assert.False(t, ids[normalTask], "Normal task should NOT be in the list")
	})

	t.Run("test priority display in human-readable mode", func(t *testing.T) {
		t.Parallel()
		// Create blocked tasks with different priorities
		lowPriorityBlocked := cli.CreateTestTask(t, db, todoColumnID, "Low Priority Blocked")
		lowBlocker := cli.CreateTestTask(t, db, todoColumnID, "Low Blocker")
		highPriorityBlocked := cli.CreateTestTask(t, db, todoColumnID, "High Priority Blocked")
		highBlocker := cli.CreateTestTask(t, db, todoColumnID, "High Blocker")

		// Assign ticket numbers and priorities
		// priority_id: 2=low, 3=medium, 4=high, 5=critical
		cli.UpdateTaskFields(t, db, lowPriorityBlocked, map[string]any{"ticket_number": 50, "priority_id": 2})
		cli.UpdateTaskFields(t, db, lowBlocker, map[string]any{"ticket_number": 51})
		cli.UpdateTaskFields(t, db, highPriorityBlocked, map[string]any{"ticket_number": 52, "priority_id": 4})
		cli.UpdateTaskFields(t, db, highBlocker, map[string]any{"ticket_number": 53})

		// Create blocking relationships
		linkCmd1 := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd1,
			[]string{"--parent", strconv.Itoa(lowPriorityBlocked), "--child", strconv.Itoa(lowBlocker), "--blocker"})
		assert.NoError(t, err)

		linkCmd2 := task.LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd2,
			[]string{"--parent", strconv.Itoa(highPriorityBlocked), "--child", strconv.Itoa(highBlocker), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Low Priority Blocked")
		assert.Contains(t, output, "High Priority Blocked")
		// Priority should be displayed for non-medium priorities
		assert.Contains(t, output, "[low]")
		assert.Contains(t, output, "[high]")
	})

	t.Run("blocked tasks in different columns", func(t *testing.T) {
		t.Parallel()
		// Create blocked tasks in different columns
		todoBlocked := cli.CreateTestTask(t, db, todoColumnID, "Todo Blocked")
		todoBlocker := cli.CreateTestTask(t, db, todoColumnID, "Todo Blocker")
		inProgressBlocked := cli.CreateTestTask(t, db, inProgressColumnID, "In Progress Blocked")
		inProgressBlocker := cli.CreateTestTask(t, db, inProgressColumnID, "In Progress Blocker")

		// Assign ticket numbers
		cli.UpdateTaskFields(t, db, todoBlocked, map[string]any{"ticket_number": 70})
		cli.UpdateTaskFields(t, db, todoBlocker, map[string]any{"ticket_number": 71})
		cli.UpdateTaskFields(t, db, inProgressBlocked, map[string]any{"ticket_number": 72})
		cli.UpdateTaskFields(t, db, inProgressBlocker, map[string]any{"ticket_number": 73})

		// Create blocking relationships
		linkCmd1 := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd1,
			[]string{"--parent", strconv.Itoa(todoBlocked), "--child", strconv.Itoa(todoBlocker), "--blocker"})
		assert.NoError(t, err)

		linkCmd2 := task.LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd2,
			[]string{"--parent", strconv.Itoa(inProgressBlocked), "--child", strconv.Itoa(inProgressBlocker), "--blocker"})
		assert.NoError(t, err)

		// List all blocked tasks
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Todo Blocked")
		assert.Contains(t, output, "In Progress Blocked")
		// Both should appear regardless of column
	})

	t.Run("blocked task with labels", func(t *testing.T) {
		t.Parallel()
		// Create blocked task with labels
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "Blocked With Labels")
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "Label Blocker")

		// Assign ticket numbers
		cli.UpdateTaskFields(t, db, blockedID, map[string]any{"ticket_number": 80})
		cli.UpdateTaskFields(t, db, blockerID, map[string]any{"ticket_number": 81})

		// Create and attach labels
		labelID1 := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "blocked", "#EF4444")
		labelID2 := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "urgent", "#F97316")

		cli.AttachLabelToTask(t, db, blockedID, labelID1)
		cli.AttachLabelToTask(t, db, blockedID, labelID2)

		// Create blocking relationship
		linkCmd := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedID), "--child", strconv.Itoa(blockerID), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks in JSON mode to verify labels
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID), "--json"})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		tasks := result["tasks"].([]any)
		foundTask := false
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]any)
			if int(taskData["ID"].(float64)) == blockedID {
				foundTask = true
				assert.Equal(t, "Blocked With Labels", taskData["Title"])
				// Labels should be present
				assert.NotNil(t, taskData["Labels"])
				break
			}
		}
		assert.True(t, foundTask, "Should find the blocked task with labels")
	})

	t.Run("jSON output structure verification", func(t *testing.T) {
		t.Parallel()
		// Create a simple blocked task
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "Structure Test")
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "Structure Blocker")

		// Assign ticket numbers
		cli.UpdateTaskFields(t, db, blockedID, map[string]any{"ticket_number": 90})
		cli.UpdateTaskFields(t, db, blockerID, map[string]any{"ticket_number": 91})

		// Create blocking relationship
		linkCmd := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedID), "--child", strconv.Itoa(blockerID), "--blocker"})
		assert.NoError(t, err)

		// Get JSON output
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID), "--json"})

		assert.NoError(t, err)

		// Parse and verify structure
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify required fields
		assert.Contains(t, result, "success")
		assert.Contains(t, result, "tasks")
		assert.Contains(t, result, "count")
		assert.True(t, result["success"].(bool))

		// Verify count matches tasks array length
		tasks := result["tasks"].([]any)
		count := int(result["count"].(float64))
		assert.Equal(t, len(tasks), count, "Count should match tasks array length")
	})

	t.Run("empty project with no tasks", func(t *testing.T) {
		t.Parallel()
		// Create a completely empty project
		emptyProjectID := cli.CreateTestProject(t, db, "Completely Empty Project")

		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(emptyProjectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "No blocked tasks found")
	})

	t.Run("blocked task with critical priority", func(t *testing.T) {
		t.Parallel()
		// Create blocked task with critical priority
		criticalBlocked := cli.CreateTestTask(t, db, todoColumnID, "Critical Blocked Task")
		criticalBlocker := cli.CreateTestTask(t, db, todoColumnID, "Critical Blocker")

		// Assign ticket numbers and critical priority (priority_id = 5)
		cli.UpdateTaskFields(t, db, criticalBlocked, map[string]any{"ticket_number": 100, "priority_id": 5})
		cli.UpdateTaskFields(t, db, criticalBlocker, map[string]any{"ticket_number": 101})

		// Create blocking relationship
		linkCmd := task.LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(criticalBlocked), "--child", strconv.Itoa(criticalBlocker), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks
		blockedCmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Critical Blocked Task")
		assert.Contains(t, output, "[critical]")
		assert.Contains(t, output, "BLOCKED")
	})
}

func TestBlockedTask_Errors(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	_, app := cli.SetupCLITest(t)

	t.Run("missing project ID - no flag and no env var", func(t *testing.T) {
		t.Parallel()
		cmd := task.BlockedCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		cli.AssertExitError(t, err, 2) // ExitUsage
		assert.Contains(t, err.Error(), "no project specified")
	})

	t.Run("invalid project ID - non-existent", func(t *testing.T) {
		t.Parallel()
		cmd := task.BlockedCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--project", "999999"})
		cli.AssertExitError(t, err, 3) // ExitNotFound
		assert.Contains(t, err.Error(), "project 999999 not found")
	})

	t.Run("project ID as string instead of int", func(t *testing.T) {
		t.Parallel()
		// This will fail at flag parsing level
		blockedCmd := task.BlockedCmd()
		_, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", "not-a-number"})

		// Should get an error from cobra flag parsing
		assert.Error(t, err)
	})

	t.Run("negative project ID", func(t *testing.T) {
		t.Parallel()
		cmd := task.BlockedCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--project", "-1"})
		// Cobra may interpret -1 as a flag; just assert error
		assert.Error(t, err)
	})

	t.Run("zero project ID", func(t *testing.T) {
		t.Parallel()
		cmd := task.BlockedCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--project", "0"})
		cli.AssertExitError(t, err, 3) // ExitNotFound
		assert.Contains(t, err.Error(), "project 0 not found")
	})
}
