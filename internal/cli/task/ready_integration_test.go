package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestReadyTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project with default columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID (which we'll mark as ready column)
	var todoColumnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&todoColumnID)
	assert.NoError(t, err)

	// Mark "Todo" column as ready column
	_, err = db.ExecContext(context.Background(),
		"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", todoColumnID)
	assert.NoError(t, err)

	t.Run("List ready tasks in project - human-readable mode", func(t *testing.T) {
		// Create tasks in ready column
		taskID1 := cli.CreateTestTask(t, db, todoColumnID, "Ready Task 1")
		taskID2 := cli.CreateTestTask(t, db, todoColumnID, "Ready Task 2")

		cmd := ReadyCmd()

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

	t.Run("List ready tasks - no ready tasks found", func(t *testing.T) {
		// Create a new project with no tasks in ready column
		newProjectID := cli.CreateTestProject(t, db, "Empty Project")

		// Get Todo column and mark as ready
		var emptyTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			newProjectID).Scan(&emptyTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", emptyTodoColumnID)
		assert.NoError(t, err)

		cmd := ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(newProjectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No ready tasks found")
	})

	t.Run("List ready tasks - JSON mode", func(t *testing.T) {
		// Create a fresh project for JSON test
		jsonProjectID := cli.CreateTestProject(t, db, "JSON Project")

		var jsonTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			jsonProjectID).Scan(&jsonTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", jsonTodoColumnID)
		assert.NoError(t, err)

		// Create tasks
		taskID1 := cli.CreateTestTask(t, db, jsonTodoColumnID, "JSON Task 1")
		taskID2 := cli.CreateTestTask(t, db, jsonTodoColumnID, "JSON Task 2")

		cmd := ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(jsonProjectID),
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

		tasks := result["tasks"].([]interface{})
		assert.Equal(t, 2, len(tasks))

		// Verify task IDs are in results
		taskIDs := []int{taskID1, taskID2}
		foundCount := 0
		for _, taskItem := range tasks {
			taskData := taskItem.(map[string]interface{})
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

	t.Run("List ready tasks - quiet mode", func(t *testing.T) {
		// Create a fresh project for quiet test
		quietProjectID := cli.CreateTestProject(t, db, "Quiet Project")

		var quietTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			quietProjectID).Scan(&quietTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", quietTodoColumnID)
		assert.NoError(t, err)

		// Create tasks
		taskID1 := cli.CreateTestTask(t, db, quietTodoColumnID, "Quiet Task 1")
		taskID2 := cli.CreateTestTask(t, db, quietTodoColumnID, "Quiet Task 2")

		cmd := ReadyCmd()

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
			fmt.Sscanf(line, "%d", &id)
			foundIDs[id] = true
		}

		assert.True(t, foundIDs[taskID1], "Should find taskID1 in output")
		assert.True(t, foundIDs[taskID2], "Should find taskID2 in output")
	})

	t.Run("Blocked tasks do NOT appear in ready list", func(t *testing.T) {
		// Create a fresh project for blocking test
		blockProjectID := cli.CreateTestProject(t, db, "Block Project")

		var blockTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			blockProjectID).Scan(&blockTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", blockTodoColumnID)
		assert.NoError(t, err)

		// Create tasks
		blockedTaskID := cli.CreateTestTask(t, db, blockTodoColumnID, "Blocked Task")
		blockerTaskID := cli.CreateTestTask(t, db, blockTodoColumnID, "Blocker Task")
		unblockedTaskID := cli.CreateTestTask(t, db, blockTodoColumnID, "Unblocked Task")

		// Create blocking relationship (blockedTask is blocked by blockerTask)
		// relation_type_id = 2 for blocking relationship
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
			blockedTaskID, blockerTaskID)
		assert.NoError(t, err)

		cmd := ReadyCmd()

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

	t.Run("Priority display in human-readable mode", func(t *testing.T) {
		// Create a fresh project for priority test
		priorityProjectID := cli.CreateTestProject(t, db, "Priority Project")

		var priorityTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			priorityProjectID).Scan(&priorityTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", priorityTodoColumnID)
		assert.NoError(t, err)

		// Create tasks with different priorities
		lowTaskID := cli.CreateTestTask(t, db, priorityTodoColumnID, "Low Priority Task")
		highTaskID := cli.CreateTestTask(t, db, priorityTodoColumnID, "High Priority Task")
		criticalTaskID := cli.CreateTestTask(t, db, priorityTodoColumnID, "Critical Priority Task")
		mediumTaskID := cli.CreateTestTask(t, db, priorityTodoColumnID, "Medium Priority Task")

		// Set priorities (1=trivial, 2=low, 3=medium, 4=high, 5=critical)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET priority_id = 2 WHERE id = ?", lowTaskID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET priority_id = 4 WHERE id = ?", highTaskID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET priority_id = 5 WHERE id = ?", criticalTaskID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET priority_id = 3 WHERE id = ?", mediumTaskID)
		assert.NoError(t, err)

		cmd := ReadyCmd()

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

	t.Run("Ready tasks must be in columns with holds_ready_tasks=true", func(t *testing.T) {
		// Create a fresh project
		flagProjectID := cli.CreateTestProject(t, db, "Flag Project")

		// Get Todo and In Progress columns
		var flagTodoColumnID, flagInProgressColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			flagProjectID).Scan(&flagTodoColumnID)
		assert.NoError(t, err)

		err = db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'In Progress'",
			flagProjectID).Scan(&flagInProgressColumnID)
		assert.NoError(t, err)

		// Only mark Todo as ready column
		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", flagTodoColumnID)
		assert.NoError(t, err)

		// Create tasks in both columns
		readyTaskID := cli.CreateTestTask(t, db, flagTodoColumnID, "Task in Ready Column")
		nonReadyTaskID := cli.CreateTestTask(t, db, flagInProgressColumnID, "Task in Non-Ready Column")

		cmd := ReadyCmd()

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

	t.Run("List ready tasks with PASO_PROJECT env var", func(t *testing.T) {
		// Create a fresh project for env var test
		envProjectID := cli.CreateTestProject(t, db, "Env Project")

		var envTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			envProjectID).Scan(&envTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", envTodoColumnID)
		assert.NoError(t, err)

		// Create task
		taskID := cli.CreateTestTask(t, db, envTodoColumnID, "Env Var Task")

		// Set PASO_PROJECT environment variable
		t.Setenv("PASO_PROJECT", strconv.Itoa(envProjectID))

		cmd := ReadyCmd()

		// Don't pass --project flag, should use env var
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{})

		assert.NoError(t, err)
		assert.Contains(t, output, "Env Var Task")
		assert.Contains(t, output, fmt.Sprintf("[%d]", taskID))
	})

	t.Run("Multiple ready tasks with labels", func(t *testing.T) {
		// Create a fresh project for labels test
		labelProjectID := cli.CreateTestProject(t, db, "Label Project")

		var labelTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			labelProjectID).Scan(&labelTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", labelTodoColumnID)
		assert.NoError(t, err)

		// Create tasks
		taskID1 := cli.CreateTestTask(t, db, labelTodoColumnID, "Task With Labels 1")
		cli.CreateTestTask(t, db, labelTodoColumnID, "Task With Labels 2")

		// Create labels
		labelID1 := testutil.CreateTestLabel(t, db, labelProjectID, "bug", "#EF4444")
		labelID2 := testutil.CreateTestLabel(t, db, labelProjectID, "urgent", "#F97316")

		// Attach labels to task1
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID1, labelID1)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID1, labelID2)
		assert.NoError(t, err)

		cmd := ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(labelProjectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Task With Labels 1")
		assert.Contains(t, output, "Task With Labels 2")
		assert.Contains(t, output, "Found 2 ready tasks")
	})

	t.Run("JSON output contains complete task structure", func(t *testing.T) {
		// Create a fresh project for complete JSON test
		completeProjectID := cli.CreateTestProject(t, db, "Complete Project")

		var completeTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			completeProjectID).Scan(&completeTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", completeTodoColumnID)
		assert.NoError(t, err)

		// Create task with all metadata
		taskID := cli.CreateTestTask(t, db, completeTodoColumnID, "Complete Task")

		_, err = db.ExecContext(context.Background(), `
			UPDATE tasks 
			SET type_id = 2, 
			    priority_id = 4,
			    description = 'Full description'
			WHERE id = ?`, taskID)
		assert.NoError(t, err)

		// Add label
		labelID := testutil.CreateTestLabel(t, db, completeProjectID, "feature", "#3B82F6")
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
		assert.NoError(t, err)

		cmd := ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(completeProjectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify structure
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(1), result["count"])

		tasks := result["tasks"].([]interface{})
		assert.Equal(t, 1, len(tasks))

		taskData := tasks[0].(map[string]interface{})
		// Note: JSON fields are capitalized
		assert.Equal(t, float64(taskID), taskData["ID"])
		assert.Equal(t, "Complete Task", taskData["Title"])
		assert.Equal(t, "feature", taskData["TypeDescription"]) // type_id=2 is "feature"
		assert.Equal(t, "high", taskData["PriorityDescription"])
		assert.NotEmpty(t, taskData["PriorityColor"])

		// Labels are embedded as an array
		labels := taskData["Labels"].([]interface{})
		assert.Equal(t, 1, len(labels), "Should have 1 label")
	})

	t.Run("Empty JSON response with no ready tasks", func(t *testing.T) {
		// Create a fresh project with no ready tasks
		emptyJSONProjectID := cli.CreateTestProject(t, db, "Empty JSON Project")

		var emptyJSONTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			emptyJSONProjectID).Scan(&emptyJSONTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", emptyJSONTodoColumnID)
		assert.NoError(t, err)

		cmd := ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(emptyJSONProjectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify structure
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(0), result["count"])

		tasks := result["tasks"].([]interface{})
		assert.Equal(t, 0, len(tasks))
	})

	t.Run("Empty quiet response with no ready tasks", func(t *testing.T) {
		// Create a fresh project with no ready tasks
		emptyQuietProjectID := cli.CreateTestProject(t, db, "Empty Quiet Project")

		var emptyQuietTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			emptyQuietProjectID).Scan(&emptyQuietTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", emptyQuietTodoColumnID)
		assert.NoError(t, err)

		cmd := ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(emptyQuietProjectID),
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode with no tasks should output nothing
		assert.Empty(t, strings.TrimSpace(output))
	})

	t.Run("Multiple ready tasks ordered by position", func(t *testing.T) {
		// Create a fresh project for ordering test
		orderProjectID := cli.CreateTestProject(t, db, "Order Project")

		var orderTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			orderProjectID).Scan(&orderTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", orderTodoColumnID)
		assert.NoError(t, err)

		// Create tasks (they should be created with incrementing positions)
		taskID1 := cli.CreateTestTask(t, db, orderTodoColumnID, "First Task")
		taskID2 := cli.CreateTestTask(t, db, orderTodoColumnID, "Second Task")
		taskID3 := cli.CreateTestTask(t, db, orderTodoColumnID, "Third Task")

		cmd := ReadyCmd()

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

	t.Run("Complex blocking scenario - multiple blockers", func(t *testing.T) {
		// Create a fresh project for complex blocking test
		complexProjectID := cli.CreateTestProject(t, db, "Complex Block Project")

		var complexTodoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			complexProjectID).Scan(&complexTodoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", complexTodoColumnID)
		assert.NoError(t, err)

		// Create tasks
		blockedTaskID := cli.CreateTestTask(t, db, complexTodoColumnID, "Task Blocked By Two")
		blocker1ID := cli.CreateTestTask(t, db, complexTodoColumnID, "Blocker 1")
		blocker2ID := cli.CreateTestTask(t, db, complexTodoColumnID, "Blocker 2")
		cli.CreateTestTask(t, db, complexTodoColumnID, "Independent Task")

		// Create two blocking relationships for the same task
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
			blockedTaskID, blocker1ID)
		assert.NoError(t, err)
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
			blockedTaskID, blocker2ID)
		assert.NoError(t, err)

		cmd := ReadyCmd()

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

func TestReadyTask_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	t.Run("Missing project ID - no flag and no env var", func(t *testing.T) {
		// Note: This test case calls os.Exit() in ready.go:63 (ExitUsage)
		// We cannot test os.Exit() calls in integration tests without special handling
		// Skipping this test as it would terminate the test process
		t.Skip("Skipping test that calls os.Exit() - cannot test exit behavior in integration tests")
	})

	t.Run("Invalid project ID - non-existent project", func(t *testing.T) {
		// Note: This test case calls os.Exit() in ready.go:88 (ExitNotFound)
		// We cannot test os.Exit() calls in integration tests without special handling
		// Skipping this test as it would terminate the test process
		t.Skip("Skipping test that calls os.Exit() - cannot test exit behavior in integration tests")
	})

	t.Run("Project with no ready column", func(t *testing.T) {
		// Create project but don't mark any column as ready column
		projectID := cli.CreateTestProject(t, db, "No Ready Column Project")

		// Get Todo column and create a task (but don't mark column as ready)
		var todoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			projectID).Scan(&todoColumnID)
		assert.NoError(t, err)

		// Create task in Todo column (but it's not marked as ready column)
		cli.CreateTestTask(t, db, todoColumnID, "Task in Non-Ready Column")

		cmd := ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(projectID),
		})

		assert.NoError(t, err)
		// Should return no ready tasks since no column is marked as ready
		assert.Contains(t, output, "No ready tasks found")
	})

	t.Run("All tasks in ready column are blocked", func(t *testing.T) {
		// Create project
		projectID := cli.CreateTestProject(t, db, "All Blocked Project")

		var todoColumnID int
		err := db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			projectID).Scan(&todoColumnID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE columns SET holds_ready_tasks = true WHERE id = ?", todoColumnID)
		assert.NoError(t, err)

		// Create a blocked task and its blocker (blocker is in a different column)
		var inProgressColumnID int
		err = db.QueryRowContext(context.Background(),
			"SELECT id FROM columns WHERE project_id = ? AND name = 'In Progress'",
			projectID).Scan(&inProgressColumnID)
		assert.NoError(t, err)

		blockedTaskID := cli.CreateTestTask(t, db, todoColumnID, "Blocked Task")
		blockerTaskID := cli.CreateTestTask(t, db, inProgressColumnID, "Blocker in Different Column")

		// Create blocking relationship
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 2)",
			blockedTaskID, blockerTaskID)
		assert.NoError(t, err)

		cmd := ReadyCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", strconv.Itoa(projectID),
		})

		assert.NoError(t, err)
		// Should return no ready tasks since the only task in ready column is blocked
		assert.Contains(t, output, "No ready tasks found")
	})
}
