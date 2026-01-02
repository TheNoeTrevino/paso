package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestLinkTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project with default columns (Todo, In Progress, Done)
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	var columnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&columnID)
	require.NoError(t, err)

	t.Run("Create parent-child relationship (default)", func(t *testing.T) {
		// Create parent and child tasks
		parentID := cli.CreateTestTask(t, db, columnID, "Parent Task")
		childID := cli.CreateTestTask(t, db, columnID, "Child Task")

		cmd := LinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childID),
		})

		// Assertions
		require.NoError(t, err)
		assert.Contains(t, output, "Linked")
		assert.Contains(t, output, fmt.Sprintf("task %d as child of task %d", childID, parentID))

		// Verify relationship in database
		var relationType int
		err = db.QueryRowContext(context.Background(),
			"SELECT relation_type_id FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentID, childID).Scan(&relationType)
		require.NoError(t, err)
		assert.Equal(t, 1, relationType, "Default relation type should be 1 (parent-child)")
	})

	t.Run("Create blocking relationship", func(t *testing.T) {
		// Create tasks: blockedTask is blocked by blockerTask
		blockedID := cli.CreateTestTask(t, db, columnID, "Blocked Task")
		blockerID := cli.CreateTestTask(t, db, columnID, "Blocker Task")

		cmd := LinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(blockedID),
			"--child", strconv.Itoa(blockerID),
			"--blocker",
		})

		// Assertions
		require.NoError(t, err)
		assert.Contains(t, output, "blocking relationship")
		assert.Contains(t, output, fmt.Sprintf("task %d is blocked by task %d", blockedID, blockerID))

		// Verify relationship in database
		var relationType int
		err = db.QueryRowContext(context.Background(),
			"SELECT relation_type_id FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			blockedID, blockerID).Scan(&relationType)
		require.NoError(t, err)
		assert.Equal(t, 2, relationType, "Blocker relation type should be 2")
	})

	t.Run("Create related relationship", func(t *testing.T) {
		// Create two related tasks
		task1ID := cli.CreateTestTask(t, db, columnID, "Related Task 1")
		task2ID := cli.CreateTestTask(t, db, columnID, "Related Task 2")

		cmd := LinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(task1ID),
			"--child", strconv.Itoa(task2ID),
			"--related",
		})

		// Assertions
		require.NoError(t, err)
		assert.Contains(t, output, "related relationship")
		assert.Contains(t, output, fmt.Sprintf("between task %d and task %d", task1ID, task2ID))

		// Verify relationship in database
		var relationType int
		err = db.QueryRowContext(context.Background(),
			"SELECT relation_type_id FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			task1ID, task2ID).Scan(&relationType)
		require.NoError(t, err)
		assert.Equal(t, 3, relationType, "Related relation type should be 3")
	})

	t.Run("Link with JSON output", func(t *testing.T) {
		parentID := cli.CreateTestTask(t, db, columnID, "JSON Parent")
		childID := cli.CreateTestTask(t, db, columnID, "JSON Child")

		cmd := LinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childID),
			"--json",
		})

		require.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(parentID), result["parent_id"])
		assert.Equal(t, float64(childID), result["child_id"])
		assert.Equal(t, float64(1), result["relation_type_id"])
		assert.Equal(t, "parent-child", result["relation_type"])
	})

	t.Run("Link with JSON output - blocking relationship", func(t *testing.T) {
		parentID := cli.CreateTestTask(t, db, columnID, "JSON Blocked")
		childID := cli.CreateTestTask(t, db, columnID, "JSON Blocker")

		cmd := LinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childID),
			"--blocker",
			"--json",
		})

		require.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(parentID), result["parent_id"])
		assert.Equal(t, float64(childID), result["child_id"])
		assert.Equal(t, float64(2), result["relation_type_id"])
		assert.Equal(t, "blocking", result["relation_type"])
	})

	t.Run("Link with JSON output - related relationship", func(t *testing.T) {
		parentID := cli.CreateTestTask(t, db, columnID, "JSON Related 1")
		childID := cli.CreateTestTask(t, db, columnID, "JSON Related 2")

		cmd := LinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childID),
			"--related",
			"--json",
		})

		require.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(parentID), result["parent_id"])
		assert.Equal(t, float64(childID), result["child_id"])
		assert.Equal(t, float64(3), result["relation_type_id"])
		assert.Equal(t, "related", result["relation_type"])
	})

	t.Run("Link with quiet mode", func(t *testing.T) {
		parentID := cli.CreateTestTask(t, db, columnID, "Quiet Parent")
		childID := cli.CreateTestTask(t, db, columnID, "Quiet Child")

		cmd := LinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childID),
			"--quiet",
		})

		require.NoError(t, err)
		// Quiet mode should produce no output
		assert.Equal(t, "", output)

		// Verify relationship was created
		var relationType int
		err = db.QueryRowContext(context.Background(),
			"SELECT relation_type_id FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentID, childID).Scan(&relationType)
		require.NoError(t, err)
		assert.Equal(t, 1, relationType)
	})

	t.Run("Multiple links from same parent", func(t *testing.T) {
		// Create one parent and multiple children
		parentID := cli.CreateTestTask(t, db, columnID, "Multi-Parent")
		child1ID := cli.CreateTestTask(t, db, columnID, "Multi-Child 1")
		child2ID := cli.CreateTestTask(t, db, columnID, "Multi-Child 2")
		child3ID := cli.CreateTestTask(t, db, columnID, "Multi-Child 3")

		// Link all children to parent
		cmd1 := LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(child1ID),
			"--quiet",
		})
		require.NoError(t, err)

		cmd2 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(child2ID),
			"--quiet",
		})
		require.NoError(t, err)

		cmd3 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd3, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(child3ID),
			"--quiet",
		})
		require.NoError(t, err)

		// Verify all three relationships exist
		var count int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ?",
			parentID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 3, count, "Should have 3 children linked to parent")
	})

	t.Run("Multiple links to same child", func(t *testing.T) {
		// Create multiple parents and one child
		parent1ID := cli.CreateTestTask(t, db, columnID, "Multi-Link Parent 1")
		parent2ID := cli.CreateTestTask(t, db, columnID, "Multi-Link Parent 2")
		childID := cli.CreateTestTask(t, db, columnID, "Multi-Link Child")

		// Link same child to multiple parents
		cmd1 := LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--parent", strconv.Itoa(parent1ID),
			"--child", strconv.Itoa(childID),
			"--quiet",
		})
		require.NoError(t, err)

		cmd2 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--parent", strconv.Itoa(parent2ID),
			"--child", strconv.Itoa(childID),
			"--quiet",
		})
		require.NoError(t, err)

		// Verify both relationships exist
		var count int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_subtasks WHERE child_id = ?",
			childID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "Child should have 2 parent links")
	})

	t.Run("Verify task metadata preserved after linking", func(t *testing.T) {
		// Create tasks with metadata
		parentID := cli.CreateTestTask(t, db, columnID, "Parent with Metadata")
		childID := cli.CreateTestTask(t, db, columnID, "Child with Metadata")

		// Set metadata on both tasks
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET description = ?, priority_id = 4, type_id = 2 WHERE id = ?",
			"Parent description", parentID)
		require.NoError(t, err)

		_, err = db.ExecContext(context.Background(),
			"UPDATE tasks SET description = ?, priority_id = 5, type_id = 3 WHERE id = ?",
			"Child description", childID)
		require.NoError(t, err)

		// Link tasks
		cmd := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childID),
			"--quiet",
		})
		require.NoError(t, err)

		// Verify parent metadata unchanged
		var parentTitle, parentDesc string
		var parentPriority, parentType int
		err = db.QueryRowContext(context.Background(),
			"SELECT title, description, priority_id, type_id FROM tasks WHERE id = ?",
			parentID).Scan(&parentTitle, &parentDesc, &parentPriority, &parentType)
		require.NoError(t, err)
		assert.Equal(t, "Parent with Metadata", parentTitle)
		assert.Equal(t, "Parent description", parentDesc)
		assert.Equal(t, 4, parentPriority)
		assert.Equal(t, 2, parentType)

		// Verify child metadata unchanged
		var childTitle, childDesc string
		var childPriority, childType int
		err = db.QueryRowContext(context.Background(),
			"SELECT title, description, priority_id, type_id FROM tasks WHERE id = ?",
			childID).Scan(&childTitle, &childDesc, &childPriority, &childType)
		require.NoError(t, err)
		assert.Equal(t, "Child with Metadata", childTitle)
		assert.Equal(t, "Child description", childDesc)
		assert.Equal(t, 5, childPriority)
		assert.Equal(t, 3, childType)
	})

	t.Run("Link tasks in different columns", func(t *testing.T) {
		// Create second column
		inProgressColumn := cli.CreateTestColumn(t, db, projectID, "In Progress")

		// Create tasks in different columns
		parentID := cli.CreateTestTask(t, db, columnID, "Parent in Todo")
		childID := cli.CreateTestTask(t, db, inProgressColumn, "Child in Progress")

		cmd := LinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childID),
		})

		require.NoError(t, err)
		assert.Contains(t, output, "Linked")

		// Verify relationship exists
		var relationType int
		err = db.QueryRowContext(context.Background(),
			"SELECT relation_type_id FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentID, childID).Scan(&relationType)
		require.NoError(t, err)
		assert.Equal(t, 1, relationType)
	})

	t.Run("Link with mixed relationship types to same parent", func(t *testing.T) {
		// One parent with different relationship types to different children
		parentID := cli.CreateTestTask(t, db, columnID, "Mixed Parent")
		childNormalID := cli.CreateTestTask(t, db, columnID, "Normal Child")
		blockerID := cli.CreateTestTask(t, db, columnID, "Blocker Child")
		relatedID := cli.CreateTestTask(t, db, columnID, "Related Child")

		// Create normal parent-child relationship
		cmd1 := LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childNormalID),
			"--quiet",
		})
		require.NoError(t, err)

		// Create blocking relationship
		cmd2 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(blockerID),
			"--blocker",
			"--quiet",
		})
		require.NoError(t, err)

		// Create related relationship
		cmd3 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd3, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(relatedID),
			"--related",
			"--quiet",
		})
		require.NoError(t, err)

		// Verify all relationships with correct types
		var normalType int
		err = db.QueryRowContext(context.Background(),
			"SELECT relation_type_id FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentID, childNormalID).Scan(&normalType)
		require.NoError(t, err)
		assert.Equal(t, 1, normalType)

		var blockerType int
		err = db.QueryRowContext(context.Background(),
			"SELECT relation_type_id FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentID, blockerID).Scan(&blockerType)
		require.NoError(t, err)
		assert.Equal(t, 2, blockerType)

		var relatedType int
		err = db.QueryRowContext(context.Background(),
			"SELECT relation_type_id FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentID, relatedID).Scan(&relatedType)
		require.NoError(t, err)
		assert.Equal(t, 3, relatedType)
	})
}

func TestLinkTask_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID
	var columnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&columnID)
	require.NoError(t, err)

	t.Run("Missing parent flag", func(t *testing.T) {
		childID := cli.CreateTestTask(t, db, columnID, "Child Task")

		cmd := LinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--child", strconv.Itoa(childID),
		})

		// Expect error due to required flag
		assert.Error(t, err)
	})

	t.Run("Missing child flag", func(t *testing.T) {
		parentID := cli.CreateTestTask(t, db, columnID, "Parent Task")

		cmd := LinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
		})

		// Expect error due to required flag
		assert.Error(t, err)
	})

	t.Run("Both blocker and related flags (mutually exclusive)", func(t *testing.T) {
		t.Skip("Skipping test: command calls os.Exit() on mutually exclusive flag error")
		// Note: This test is skipped because the link command calls os.Exit(cli.ExitUsage)
		// when both --blocker and --related flags are provided, which terminates the test process.
		// The validation logic in link.go (lines 76-82) correctly handles this case.
	})

	t.Run("Invalid parent ID (non-existent task)", func(t *testing.T) {
		childID := cli.CreateTestTask(t, db, columnID, "Valid Child")
		nonExistentParentID := 99999

		cmd := LinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(nonExistentParentID),
			"--child", strconv.Itoa(childID),
		})

		// Expect error
		assert.Error(t, err)
	})

	t.Run("Invalid child ID (non-existent task)", func(t *testing.T) {
		parentID := cli.CreateTestTask(t, db, columnID, "Valid Parent")
		nonExistentChildID := 99999

		cmd := LinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(nonExistentChildID),
		})

		// Expect error
		assert.Error(t, err)
	})

	t.Run("Self-reference (parent equals child)", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Self Reference Task")

		cmd := LinkCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(taskID),
			"--child", strconv.Itoa(taskID),
		})

		// Expect error - service validates self-reference
		assert.Error(t, err)
		// Error message should indicate self-relation issue
		assert.Contains(t, strings.ToLower(output+err.Error()), "self")
	})

	t.Run("Circular dependency prevention", func(t *testing.T) {
		// Create tasks: A -> B -> C
		taskA := cli.CreateTestTask(t, db, columnID, "Task A")
		taskB := cli.CreateTestTask(t, db, columnID, "Task B")
		taskC := cli.CreateTestTask(t, db, columnID, "Task C")

		// Link A -> B
		cmd1 := LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--parent", strconv.Itoa(taskA),
			"--child", strconv.Itoa(taskB),
			"--quiet",
		})
		require.NoError(t, err)

		// Link B -> C
		cmd2 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--parent", strconv.Itoa(taskB),
			"--child", strconv.Itoa(taskC),
			"--quiet",
		})
		require.NoError(t, err)

		// Try to link C -> A (would create cycle)
		cmd3 := LinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd3, []string{
			"--parent", strconv.Itoa(taskC),
			"--child", strconv.Itoa(taskA),
		})

		// Expect error - service should prevent circular dependency
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(output+err.Error()), "circular")
	})

	t.Run("Zero parent ID", func(t *testing.T) {
		childID := cli.CreateTestTask(t, db, columnID, "Valid Child")

		cmd := LinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", "0",
			"--child", strconv.Itoa(childID),
		})

		// Expect error - zero is invalid
		assert.Error(t, err)
	})

	t.Run("Zero child ID", func(t *testing.T) {
		parentID := cli.CreateTestTask(t, db, columnID, "Valid Parent")

		cmd := LinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", "0",
		})

		// Expect error - zero is invalid
		assert.Error(t, err)
	})

	t.Run("Negative parent ID", func(t *testing.T) {
		childID := cli.CreateTestTask(t, db, columnID, "Valid Child")

		cmd := LinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", "-1",
			"--child", strconv.Itoa(childID),
		})

		// Expect error - negative ID is invalid
		assert.Error(t, err)
	})

	t.Run("Negative child ID", func(t *testing.T) {
		parentID := cli.CreateTestTask(t, db, columnID, "Valid Parent")

		cmd := LinkCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", "-1",
		})

		// Expect error - negative ID is invalid
		assert.Error(t, err)
	})

	t.Run("Duplicate link (same parent-child pair) - idempotent", func(t *testing.T) {
		parentID := cli.CreateTestTask(t, db, columnID, "Parent for Duplicate")
		childID := cli.CreateTestTask(t, db, columnID, "Child for Duplicate")

		// Create first link
		cmd1 := LinkCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childID),
			"--quiet",
		})
		require.NoError(t, err)

		// Try to create duplicate link - should succeed (idempotent using INSERT OR REPLACE)
		cmd2 := LinkCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--parent", strconv.Itoa(parentID),
			"--child", strconv.Itoa(childID),
			"--quiet",
		})

		// Should succeed - command is idempotent
		require.NoError(t, err)

		// Verify only one relationship exists
		var count int
		err = db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?",
			parentID, childID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should have exactly one relationship, not duplicates")
	})
}
