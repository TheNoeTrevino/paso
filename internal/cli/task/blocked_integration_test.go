package task

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestBlockedTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

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

	t.Run("List blocked tasks with blocking relationships", func(t *testing.T) {
		// Create parent task
		parentID := cli.CreateTestTask(t, db, todoColumnID, "Parent Task")
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 1 WHERE id = ?", parentID)
		assert.NoError(t, err)

		// Create child task (blocker)
		childID := cli.CreateTestTask(t, db, todoColumnID, "Child Task (Blocker)")
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 2 WHERE id = ?", childID)
		assert.NoError(t, err)

		// Create blocking relationship (parentID blocked by childID)
		linkCmd := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(parentID), "--child", strconv.Itoa(childID), "--blocker"})
		assert.NoError(t, err)

		// Now parentID should appear in blocked list
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Parent Task")
		assert.Contains(t, output, "BLOCKED")
		assert.Contains(t, output, "Found 1 blocked tasks")
	})

	t.Run("List blocked tasks with no blocked tasks", func(t *testing.T) {
		// Create a new project with no blocking relationships
		newProjectID := cli.CreateTestProject(t, db, "Empty Project")

		// Get the default "Todo" column ID
		var emptyTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			newProjectID).Scan(&emptyTodoColumnID)
		assert.NoError(t, err)

		// Create tasks but no blocking relationships
		_ = cli.CreateTestTask(t, db, emptyTodoColumnID, "Regular Task 1")
		_ = cli.CreateTestTask(t, db, emptyTodoColumnID, "Regular Task 2")

		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(newProjectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "No blocked tasks found")
	})

	t.Run("List multiple blocked tasks", func(t *testing.T) {
		// Create multiple blocked tasks
		blocked1 := cli.CreateTestTask(t, db, todoColumnID, "Blocked Task 1")
		blocker1 := cli.CreateTestTask(t, db, todoColumnID, "Blocker Task 1")
		blocked2 := cli.CreateTestTask(t, db, todoColumnID, "Blocked Task 2")
		blocker2 := cli.CreateTestTask(t, db, todoColumnID, "Blocker Task 2")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 10 WHERE id = ?", blocked1)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 11 WHERE id = ?", blocker1)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 12 WHERE id = ?", blocked2)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 13 WHERE id = ?", blocker2)
		assert.NoError(t, err)

		// Create blocking relationships
		linkCmd1 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd1,
			[]string{"--parent", strconv.Itoa(blocked1), "--child", strconv.Itoa(blocker1), "--blocker"})
		assert.NoError(t, err)

		linkCmd2 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd2,
			[]string{"--parent", strconv.Itoa(blocked2), "--child", strconv.Itoa(blocker2), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Blocked Task 1")
		assert.Contains(t, output, "Blocked Task 2")
		assert.Contains(t, output, "BLOCKED")
	})

	t.Run("List blocked tasks in JSON mode", func(t *testing.T) {
		// Create blocked task
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "JSON Blocked Task")
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "JSON Blocker Task")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 20 WHERE id = ?", blockedID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 21 WHERE id = ?", blockerID)
		assert.NoError(t, err)

		// Create blocking relationship
		linkCmd := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedID), "--child", strconv.Itoa(blockerID), "--blocker"})
		assert.NoError(t, err)

		// List in JSON mode
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID), "--json"})

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
		assert.GreaterOrEqual(t, len(tasks), 1, "Should have at least 1 blocked task")

		// Verify task structure
		foundTask := false
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]interface{})
			if taskData["ID"] != nil && int(taskData["ID"].(float64)) == blockedID {
				foundTask = true
				assert.Equal(t, "JSON Blocked Task", taskData["Title"])
				assert.True(t, taskData["IsBlocked"].(bool))
				break
			}
		}
		assert.True(t, foundTask, "Should find the blocked task in JSON list")
	})

	t.Run("List blocked tasks in quiet mode", func(t *testing.T) {
		// Create blocked task
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Blocked Task")
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Blocker Task")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 30 WHERE id = ?", blockedID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 31 WHERE id = ?", blockerID)
		assert.NoError(t, err)

		// Create blocking relationship
		linkCmd := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedID), "--child", strconv.Itoa(blockerID), "--blocker"})
		assert.NoError(t, err)

		// List in quiet mode
		blockedCmd := BlockedCmd()
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

	t.Run("Verify only IsBlocked==true tasks returned", func(t *testing.T) {
		// Create a mix of blocked and non-blocked tasks
		normalTask := cli.CreateTestTask(t, db, todoColumnID, "Normal Task")
		blockedTask := cli.CreateTestTask(t, db, todoColumnID, "Should Be Blocked")
		blockerTask := cli.CreateTestTask(t, db, todoColumnID, "Blocker For Test")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 40 WHERE id = ?", normalTask)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 41 WHERE id = ?", blockedTask)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 42 WHERE id = ?", blockerTask)
		assert.NoError(t, err)

		// Create blocking relationship for only one task
		linkCmd := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedTask), "--child", strconv.Itoa(blockerTask), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks in quiet mode
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID), "--quiet"})

		assert.NoError(t, err)

		// Parse output
		lines := strings.Split(strings.TrimSpace(output), "\n")
		ids := make(map[int]bool)
		for _, line := range lines {
			if line != "" {
				id, _ := strconv.Atoi(line)
				ids[id] = true
			}
		}

		// Verify blockedTask is in the list
		assert.True(t, ids[blockedTask], "Blocked task should be in the list")
		// Verify normalTask is NOT in the list
		assert.False(t, ids[normalTask], "Normal task should NOT be in the list")
	})

	t.Run("Test priority display in human-readable mode", func(t *testing.T) {
		// Create blocked tasks with different priorities
		lowPriorityBlocked := cli.CreateTestTask(t, db, todoColumnID, "Low Priority Blocked")
		lowBlocker := cli.CreateTestTask(t, db, todoColumnID, "Low Blocker")
		highPriorityBlocked := cli.CreateTestTask(t, db, todoColumnID, "High Priority Blocked")
		highBlocker := cli.CreateTestTask(t, db, todoColumnID, "High Blocker")

		// Assign ticket numbers and priorities
		// priority_id: 2=low, 3=medium, 4=high, 5=critical
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 50, priority_id = 2 WHERE id = ?", lowPriorityBlocked)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 51 WHERE id = ?", lowBlocker)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 52, priority_id = 4 WHERE id = ?", highPriorityBlocked)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 53 WHERE id = ?", highBlocker)
		assert.NoError(t, err)

		// Create blocking relationships
		linkCmd1 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd1,
			[]string{"--parent", strconv.Itoa(lowPriorityBlocked), "--child", strconv.Itoa(lowBlocker), "--blocker"})
		assert.NoError(t, err)

		linkCmd2 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd2,
			[]string{"--parent", strconv.Itoa(highPriorityBlocked), "--child", strconv.Itoa(highBlocker), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Low Priority Blocked")
		assert.Contains(t, output, "High Priority Blocked")
		// Priority should be displayed for non-medium priorities
		assert.Contains(t, output, "[low]")
		assert.Contains(t, output, "[high]")
	})

	t.Run("Test with PASO_PROJECT env var", func(t *testing.T) {
		// Create blocked task
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "Env Var Blocked Task")
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "Env Var Blocker Task")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 60 WHERE id = ?", blockedID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 61 WHERE id = ?", blockerID)
		assert.NoError(t, err)

		// Create blocking relationship
		linkCmd := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedID), "--child", strconv.Itoa(blockerID), "--blocker"})
		assert.NoError(t, err)

		// Set PASO_PROJECT environment variable
		t.Setenv("PASO_PROJECT", strconv.Itoa(projectID))

		// List blocked tasks without --project flag
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd, []string{})

		assert.NoError(t, err)
		assert.Contains(t, output, "Env Var Blocked Task")
		assert.Contains(t, output, "BLOCKED")
	})

	t.Run("Blocked tasks in different columns", func(t *testing.T) {
		// Create blocked tasks in different columns
		todoBlocked := cli.CreateTestTask(t, db, todoColumnID, "Todo Blocked")
		todoBlocker := cli.CreateTestTask(t, db, todoColumnID, "Todo Blocker")
		inProgressBlocked := cli.CreateTestTask(t, db, inProgressColumnID, "In Progress Blocked")
		inProgressBlocker := cli.CreateTestTask(t, db, inProgressColumnID, "In Progress Blocker")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 70 WHERE id = ?", todoBlocked)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 71 WHERE id = ?", todoBlocker)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 72 WHERE id = ?", inProgressBlocked)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 73 WHERE id = ?", inProgressBlocker)
		assert.NoError(t, err)

		// Create blocking relationships
		linkCmd1 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd1,
			[]string{"--parent", strconv.Itoa(todoBlocked), "--child", strconv.Itoa(todoBlocker), "--blocker"})
		assert.NoError(t, err)

		linkCmd2 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd2,
			[]string{"--parent", strconv.Itoa(inProgressBlocked), "--child", strconv.Itoa(inProgressBlocker), "--blocker"})
		assert.NoError(t, err)

		// List all blocked tasks
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Todo Blocked")
		assert.Contains(t, output, "In Progress Blocked")
		// Both should appear regardless of column
	})

	t.Run("Blocked task with labels", func(t *testing.T) {
		// Create blocked task with labels
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "Blocked With Labels")
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "Label Blocker")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 80 WHERE id = ?", blockedID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 81 WHERE id = ?", blockerID)
		assert.NoError(t, err)

		// Create and attach labels
		labelID1 := testutil.CreateTestLabel(t, db, projectID, "blocked", "#EF4444")
		labelID2 := testutil.CreateTestLabel(t, db, projectID, "urgent", "#F97316")

		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", blockedID, labelID1)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", blockedID, labelID2)
		assert.NoError(t, err)

		// Create blocking relationship
		linkCmd := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedID), "--child", strconv.Itoa(blockerID), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks in JSON mode to verify labels
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID), "--json"})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		tasks := result["tasks"].([]interface{})
		foundTask := false
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]interface{})
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

	t.Run("JSON output structure verification", func(t *testing.T) {
		// Create a simple blocked task
		blockedID := cli.CreateTestTask(t, db, todoColumnID, "Structure Test")
		blockerID := cli.CreateTestTask(t, db, todoColumnID, "Structure Blocker")

		// Assign ticket numbers
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 90 WHERE id = ?", blockedID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 91 WHERE id = ?", blockerID)
		assert.NoError(t, err)

		// Create blocking relationship
		linkCmd := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(blockedID), "--child", strconv.Itoa(blockerID), "--blocker"})
		assert.NoError(t, err)

		// Get JSON output
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID), "--json"})

		assert.NoError(t, err)

		// Parse and verify structure
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify required fields
		assert.Contains(t, result, "success")
		assert.Contains(t, result, "tasks")
		assert.Contains(t, result, "count")
		assert.True(t, result["success"].(bool))

		// Verify count matches tasks array length
		tasks := result["tasks"].([]interface{})
		count := int(result["count"].(float64))
		assert.Equal(t, len(tasks), count, "Count should match tasks array length")
	})

	t.Run("Empty project with no tasks", func(t *testing.T) {
		// Create a completely empty project
		emptyProjectID := cli.CreateTestProject(t, db, "Completely Empty Project")

		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(emptyProjectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "No blocked tasks found")
	})

	t.Run("Blocked task with critical priority", func(t *testing.T) {
		// Create blocked task with critical priority
		criticalBlocked := cli.CreateTestTask(t, db, todoColumnID, "Critical Blocked Task")
		criticalBlocker := cli.CreateTestTask(t, db, todoColumnID, "Critical Blocker")

		// Assign ticket numbers and critical priority (priority_id = 5)
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 100, priority_id = 5 WHERE id = ?", criticalBlocked)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET ticket_number = 101 WHERE id = ?", criticalBlocker)
		assert.NoError(t, err)

		// Create blocking relationship
		linkCmd := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, linkCmd,
			[]string{"--parent", strconv.Itoa(criticalBlocked), "--child", strconv.Itoa(criticalBlocker), "--blocker"})
		assert.NoError(t, err)

		// List blocked tasks
		blockedCmd := BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", strconv.Itoa(projectID)})

		assert.NoError(t, err)
		assert.Contains(t, output, "Critical Blocked Task")
		assert.Contains(t, output, "[critical]")
		assert.Contains(t, output, "BLOCKED")
	})
}

func TestBlockedTask_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	t.Run("Missing project ID - no flag and no env var", func(t *testing.T) {
		// Note: This test will trigger os.Exit() call in the command
		// Since os.Exit() terminates the process, we skip this test
		// and document the expected behavior
		t.Skip("Command calls os.Exit() on missing project ID - cannot test in integration test")

		// Expected behavior:
		// - Error code: cli.ExitUsage
		// - Error message: "NO_PROJECT" with suggestion to use "paso use project"
	})

	t.Run("Invalid project ID - non-existent", func(t *testing.T) {
		// Note: This test will trigger os.Exit() call in the command
		// Since os.Exit() terminates the process, we skip this test
		// and document the expected behavior
		t.Skip("Command calls os.Exit() on invalid project ID - cannot test in integration test")

		// Expected behavior:
		// - Error code: cli.ExitNotFound
		// - Error message: "PROJECT_NOT_FOUND" with suggestion to use "paso project list"
	})

	t.Run("Project ID as string instead of int", func(t *testing.T) {
		// This will fail at flag parsing level
		blockedCmd := BlockedCmd()
		_, err := cli.ExecuteCLICommand(t, app, blockedCmd,
			[]string{"--project", "not-a-number"})

		// Should get an error from cobra flag parsing
		assert.Error(t, err)
	})

	t.Run("Negative project ID", func(t *testing.T) {
		// Note: This will likely trigger os.Exit() for non-existent project
		t.Skip("Command calls os.Exit() on invalid project ID - cannot test in integration test")

		// Expected behavior:
		// - Error code: cli.ExitNotFound
		// - Project -1 does not exist
	})

	t.Run("Zero project ID", func(t *testing.T) {
		// Note: This will likely trigger os.Exit() for non-existent project
		t.Skip("Command calls os.Exit() on invalid project ID - cannot test in integration test")

		// Expected behavior:
		// - Error code: cli.ExitNotFound
		// - Project 0 does not exist
	})
}
