package task_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestReadyTask_Integration(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	// Create test project with default columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID (which we'll mark as ready column)
	todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	// Mark "Todo" column as ready column
	cli.SetColumnHoldsReadyTasks(t, db, todoColumnID)

	t.Run("list ready tasks in project - human-readable mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create tasks in ready column
		taskID1 := cli.CreateTestTask(t, db, todoColumnID, "Ready Task 1")
		taskID2 := cli.CreateTestTask(t, db, todoColumnID, "Ready Task 2")

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Found 2 ready tasks")
		assert.Contains(t, output, "Ready Task 1")
		assert.Contains(t, output, "Ready Task 2")
		assert.Contains(t, output, fmt.Sprintf("[%d]", taskID1))
		assert.Contains(t, output, fmt.Sprintf("[%d]", taskID2))
	})

	t.Run("list ready tasks - no ready tasks found", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a new project with no tasks in ready column
		newProjectID := cli.CreateTestProject(t, db, "Empty Project")

		// Get Todo column and mark as ready
		emptyTodoColumnID := cli.GetColumnIDByName(t, db, newProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, emptyTodoColumnID)

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(newProjectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No ready tasks found")
	})

	t.Run("list ready tasks - JSON mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project for JSON test
		jsonProjectID := cli.CreateTestProject(t, db, "JSON Project")

		jsonTodoColumnID := cli.GetColumnIDByName(t, db, jsonProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, jsonTodoColumnID)

		// Create tasks
		taskID1 := cli.CreateTestTask(t, db, jsonTodoColumnID, "JSON Task 1")
		taskID2 := cli.CreateTestTask(t, db, jsonTodoColumnID, "JSON Task 2")

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(jsonProjectID),
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

		// Handle count which may be float64 or int
		count := result["count"]
		var countValue float64
		switch v := count.(type) {
		case float64:
			countValue = v
		case int:
			countValue = float64(v)
		}
		assert.Equal(t, float64(2), countValue)

		tasks := result["tasks"].([]any)
		assert.Equal(t, 2, len(tasks))

		// Verify task IDs are in results
		taskIDs := []int{taskID1, taskID2}
		foundCount := 0
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]any)
			// Note: JSON fields are capitalized (ID, Title, etc.)
			taskID := int(taskData["ID"].(float64))
			for _, expectedID := range taskIDs {
				if taskID == expectedID {
					foundCount++
					break
				}
			}
		}
		assert.Equal(t, 2, foundCount, "Should find both tasks in JSON output")
	})

	t.Run("list ready tasks - quiet mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project for quiet test
		quietProjectID := cli.CreateTestProject(t, db, "Quiet Project")

		quietTodoColumnID := cli.GetColumnIDByName(t, db, quietProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, quietTodoColumnID)

		// Create tasks
		taskID1 := cli.CreateTestTask(t, db, quietTodoColumnID, "Quiet Task 1")
		taskID2 := cli.CreateTestTask(t, db, quietTodoColumnID, "Quiet Task 2")

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(quietProjectID),
			"--quiet",
		})

		assert.NoError(t, err)

		// Quiet mode should only output task IDs, one per line
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Equal(t, 2, len(lines), "Should have exactly 2 lines")

		// Verify each line is a numeric task ID
		foundIDs := make(map[int]bool)
		for _, line := range lines {
			assert.Regexp(t, `^\d+$`, line, "Each line should be a numeric ID")
			var id int
			_, err := fmt.Sscanf(line, "%d", &id)
			assert.NoError(t, err)
			foundIDs[id] = true
		}

		assert.True(t, foundIDs[taskID1], "Should find taskID1 in output")
		assert.True(t, foundIDs[taskID2], "Should find taskID2 in output")
	})

	t.Run("blocked tasks do NOT appear in ready list", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project for blocking test
		blockProjectID := cli.CreateTestProject(t, db, "Block Project")

		blockTodoColumnID := cli.GetColumnIDByName(t, db, blockProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, blockTodoColumnID)

		// Create tasks
		blockedTaskID := cli.CreateTestTask(t, db, blockTodoColumnID, "Blocked Task")
		blockerTaskID := cli.CreateTestTask(t, db, blockTodoColumnID, "Blocker Task")
		unblockedTaskID := cli.CreateTestTask(t, db, blockTodoColumnID, "Unblocked Task")

		// Create blocking relationship (blockedTask is blocked by blockerTask)
		// relation_type_id = 2 for blocking relationship
		cli.AddTaskSubtask(t, db, blockedTaskID, blockerTaskID, 2)

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(blockProjectID),
		})

		assert.NoError(t, err)

		// blockedTask should NOT appear
		assert.NotContains(t, output, "Blocked Task")
		assert.NotContains(t, output, fmt.Sprintf("[%d]", blockedTaskID))

		// blockerTask SHOULD appear (it's not blocked)
		assert.Contains(t, output, "Blocker Task")
		assert.Contains(t, output, fmt.Sprintf("[%d]", blockerTaskID))

		// unblockedTask SHOULD appear
		assert.Contains(t, output, "Unblocked Task")
		assert.Contains(t, output, fmt.Sprintf("[%d]", unblockedTaskID))

		// Should report 2 ready tasks (blocker and unblocked)
		assert.Contains(t, output, "Found 2 ready tasks")
	})

	t.Run("priority display in human-readable mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project for priority test
		priorityProjectID := cli.CreateTestProject(t, db, "Priority Project")

		priorityTodoColumnID := cli.GetColumnIDByName(t, db, priorityProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, priorityTodoColumnID)

		// Create tasks with different priorities
		lowTaskID := cli.CreateTestTask(t, db, priorityTodoColumnID, "Low Priority Task")
		highTaskID := cli.CreateTestTask(t, db, priorityTodoColumnID, "High Priority Task")
		criticalTaskID := cli.CreateTestTask(t, db, priorityTodoColumnID, "Critical Priority Task")
		mediumTaskID := cli.CreateTestTask(t, db, priorityTodoColumnID, "Medium Priority Task")

		// Set priorities (1=trivial, 2=low, 3=medium, 4=high, 5=critical)
		cli.UpdateTaskFields(t, db, lowTaskID, map[string]any{"priority_id": 2})
		cli.UpdateTaskFields(t, db, highTaskID, map[string]any{"priority_id": 4})
		cli.UpdateTaskFields(t, db, criticalTaskID, map[string]any{"priority_id": 5})
		cli.UpdateTaskFields(t, db, mediumTaskID, map[string]any{"priority_id": 3})

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(priorityProjectID),
		})

		assert.NoError(t, err)

		// According to ready.go:128-129, priority is shown if not medium
		assert.Contains(t, output, "[low]")
		assert.Contains(t, output, "[high]")
		assert.Contains(t, output, "[critical]")

		// Medium priority should NOT have priority label (it's the default)
		// Verify the medium task line doesn't have a priority tag
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Medium Priority Task") {
				assert.NotContains(t, line, "[medium]")
			}
		}
	})

	t.Run("ready tasks must be in columns with holds_ready_tasks=true", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project
		flagProjectID := cli.CreateTestProject(t, db, "Flag Project")

		// Get Todo and In Progress columns
		flagTodoColumnID := cli.GetColumnIDByName(t, db, flagProjectID, "Todo")
		flagInProgressColumnID := cli.GetColumnIDByName(t, db, flagProjectID, "In Progress")

		// Only mark Todo as ready column
		cli.SetColumnHoldsReadyTasks(t, db, flagTodoColumnID)

		// Create tasks in both columns
		readyTaskID := cli.CreateTestTask(t, db, flagTodoColumnID, "Task in Ready Column")
		nonReadyTaskID := cli.CreateTestTask(t, db, flagInProgressColumnID, "Task in Non-Ready Column")

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(flagProjectID),
		})

		assert.NoError(t, err)

		// Only task in ready column should appear
		assert.Contains(t, output, "Task in Ready Column")
		assert.Contains(t, output, fmt.Sprintf("[%d]", readyTaskID))

		// Task in non-ready column should NOT appear
		assert.NotContains(t, output, "Task in Non-Ready Column")
		assert.NotContains(t, output, fmt.Sprintf("[%d]", nonReadyTaskID))

		assert.Contains(t, output, "Found 1 ready tasks")
	})

	t.Run("multiple ready tasks with labels", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project for labels test
		labelProjectID := cli.CreateTestProject(t, db, "Label Project")

		labelTodoColumnID := cli.GetColumnIDByName(t, db, labelProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, labelTodoColumnID)

		// Create tasks
		taskID1 := cli.CreateTestTask(t, db, labelTodoColumnID, "Task With Labels 1")
		cli.CreateTestTask(t, db, labelTodoColumnID, "Task With Labels 2")

		// Create labels
		labelID1 := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), labelProjectID, "bug", "#EF4444")
		labelID2 := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), labelProjectID, "urgent", "#F97316")

		// Attach labels to task1
		cli.AttachLabelToTask(t, db, taskID1, labelID1)
		cli.AttachLabelToTask(t, db, taskID1, labelID2)

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(labelProjectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Task With Labels 1")
		assert.Contains(t, output, "Task With Labels 2")
		assert.Contains(t, output, "Found 2 ready tasks")
	})

	t.Run("jSON output contains complete task structure", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project for complete JSON test
		completeProjectID := cli.CreateTestProject(t, db, "Complete Project")

		completeTodoColumnID := cli.GetColumnIDByName(t, db, completeProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, completeTodoColumnID)

		// Create task with all metadata
		taskID := cli.CreateTestTask(t, db, completeTodoColumnID, "Complete Task")

		cli.UpdateTaskFields(t, db, taskID, map[string]any{
			"type_id":     2,
			"priority_id": 4,
			"description": "Full description",
		})

		// Add label
		labelID := fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), completeProjectID, "feature", "#3B82F6")
		cli.AttachLabelToTask(t, db, taskID, labelID)

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(completeProjectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify structure
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(1), result["count"])

		tasks := result["tasks"].([]any)
		assert.Equal(t, 1, len(tasks))

		taskData := tasks[0].(map[string]any)
		// Note: JSON fields are capitalized
		assert.Equal(t, float64(taskID), taskData["ID"])
		assert.Equal(t, "Complete Task", taskData["Title"])
		assert.Equal(t, "feature", taskData["TypeDescription"]) // type_id=2 is "feature"
		assert.Equal(t, "high", taskData["PriorityDescription"])
		assert.NotEmpty(t, taskData["PriorityColor"])

		// Labels are embedded as an array
		labels := taskData["Labels"].([]any)
		assert.Equal(t, 1, len(labels), "Should have 1 label")
	})

	t.Run("empty JSON response with no ready tasks", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project with no ready tasks
		emptyJSONProjectID := cli.CreateTestProject(t, db, "Empty JSON Project")

		emptyJSONTodoColumnID := cli.GetColumnIDByName(t, db, emptyJSONProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, emptyJSONTodoColumnID)

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(emptyJSONProjectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify structure
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(0), result["count"])

		tasks := result["tasks"].([]any)
		assert.Equal(t, 0, len(tasks))
	})

	t.Run("empty quiet response with no ready tasks", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project with no ready tasks
		emptyQuietProjectID := cli.CreateTestProject(t, db, "Empty Quiet Project")

		emptyQuietTodoColumnID := cli.GetColumnIDByName(t, db, emptyQuietProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, emptyQuietTodoColumnID)

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(emptyQuietProjectID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode with no tasks should output nothing
		assert.Empty(t, strings.TrimSpace(output))
	})

	t.Run("multiple ready tasks ordered by position", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project for ordering test
		orderProjectID := cli.CreateTestProject(t, db, "Order Project")

		orderTodoColumnID := cli.GetColumnIDByName(t, db, orderProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, orderTodoColumnID)

		// Create tasks (they should be created with incrementing positions)
		taskID1 := cli.CreateTestTask(t, db, orderTodoColumnID, "First Task")
		taskID2 := cli.CreateTestTask(t, db, orderTodoColumnID, "Second Task")
		taskID3 := cli.CreateTestTask(t, db, orderTodoColumnID, "Third Task")

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(orderProjectID),
		})

		assert.NoError(t, err)

		// Verify output contains all tasks
		assert.Contains(t, output, "First Task")
		assert.Contains(t, output, "Second Task")
		assert.Contains(t, output, "Third Task")

		// Verify order in output (First should appear before Second, Second before Third)
		firstIdx := strings.Index(output, fmt.Sprintf("[%d]", taskID1))
		secondIdx := strings.Index(output, fmt.Sprintf("[%d]", taskID2))
		thirdIdx := strings.Index(output, fmt.Sprintf("[%d]", taskID3))

		assert.True(t, firstIdx < secondIdx, "First task should appear before second")
		assert.True(t, secondIdx < thirdIdx, "Second task should appear before third")
	})

	t.Run("complex blocking scenario - multiple blockers", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create a fresh project for complex blocking test
		complexProjectID := cli.CreateTestProject(t, db, "Complex Block Project")

		complexTodoColumnID := cli.GetColumnIDByName(t, db, complexProjectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, complexTodoColumnID)

		// Create tasks
		blockedTaskID := cli.CreateTestTask(t, db, complexTodoColumnID, "Task Blocked By Two")
		blocker1ID := cli.CreateTestTask(t, db, complexTodoColumnID, "Blocker 1")
		blocker2ID := cli.CreateTestTask(t, db, complexTodoColumnID, "Blocker 2")
		cli.CreateTestTask(t, db, complexTodoColumnID, "Independent Task")

		// Create two blocking relationships for the same task
		cli.AddTaskSubtask(t, db, blockedTaskID, blocker1ID, 2)
		cli.AddTaskSubtask(t, db, blockedTaskID, blocker2ID, 2)

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(complexProjectID),
		})

		assert.NoError(t, err)

		// Blocked task should NOT appear
		assert.NotContains(t, output, "Task Blocked By Two")

		// Both blockers SHOULD appear (they're not blocked)
		assert.Contains(t, output, "Blocker 1")
		assert.Contains(t, output, "Blocker 2")

		// Independent task SHOULD appear
		assert.Contains(t, output, "Independent Task")

		// Should report 3 ready tasks
		assert.Contains(t, output, "Found 3 ready tasks")
	})
}

func TestReadyTask_Integration_Errors(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	t.Run("missing project ID - no flag and no env var", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := task.ReadyCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		cli.AssertExitError(t, err, 2) // ExitUsage
		assert.Contains(t, err.Error(), "no project specified")
	})

	t.Run("invalid project ID - non-existent project", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := task.ReadyCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--project", "999999"})
		cli.AssertExitError(t, err, 3) // ExitNotFound
		assert.Contains(t, err.Error(), "project 999999 not found")
	})

	t.Run("project with no ready column", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create project but don't mark any column as ready column
		projectID := cli.CreateTestProject(t, db, "No Ready Column Project")

		// Get Todo column and create a task (but don't mark column as ready)
		todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

		// Create task in Todo column (but it's not marked as ready column)
		cli.CreateTestTask(t, db, todoColumnID, "Task in Non-Ready Column")

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(projectID),
		})

		assert.NoError(t, err)
		// Should return no ready tasks since no column is marked as ready
		assert.Contains(t, output, "No ready tasks found")
	})

	t.Run("all tasks in ready column are blocked", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Create project
		projectID := cli.CreateTestProject(t, db, "All Blocked Project")

		todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")
		cli.SetColumnHoldsReadyTasks(t, db, todoColumnID)

		// Create a blocked task and its blocker (blocker is in a different column)
		inProgressColumnID := cli.GetColumnIDByName(t, db, projectID, "In Progress")

		blockedTaskID := cli.CreateTestTask(t, db, todoColumnID, "Blocked Task")
		blockerTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Blocker in Different Column")

		// Create blocking relationship
		cli.AddTaskSubtask(t, db, blockedTaskID, blockerTaskID, 2)

		cmd := task.ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(projectID),
		})

		assert.NoError(t, err)
		// Should return no ready tasks since the only task in ready column is blocked
		assert.Contains(t, output, "No ready tasks found")
	})
}
