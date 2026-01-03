package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestUpdateTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project - this creates default columns (Todo, In Progress, Done)
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the first column ID (Todo column) to use for creating tasks
	var columnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? ORDER BY id LIMIT 1", projectID).Scan(&columnID)
	if err != nil {
		t.Fatalf("Failed to get column ID: %v", err)
	}

	t.Run("Update title only", func(t *testing.T) {
		// Create a task first
		taskID := cli.CreateTestTask(t, db, columnID, "Original Title")

		// First add a description so we can verify it doesn't change
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET description = ? WHERE id = ?", "Original Description", taskID)
		assert.NoError(t, err)

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--title", "Updated Title",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify in DB
		var title, description string
		err = db.QueryRowContext(context.Background(),
			"SELECT title, description FROM tasks WHERE id = ?", taskID).Scan(&title, &description)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", title)
		assert.Equal(t, "Original Description", description) // unchanged
	})

	t.Run("Update description only", func(t *testing.T) {
		// Create a task first
		taskID := cli.CreateTestTask(t, db, columnID, "Original Title")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--description", "Updated Description",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify in DB
		var title, description string
		err = db.QueryRowContext(context.Background(),
			"SELECT title, description FROM tasks WHERE id = ?", taskID).Scan(&title, &description)
		assert.NoError(t, err)
		assert.Equal(t, "Original Title", title) // unchanged
		assert.Equal(t, "Updated Description", description)
	})

	t.Run("Update priority to trivial", func(t *testing.T) {
		// Create a task first with default priority
		taskID := cli.CreateTestTask(t, db, columnID, "Task Title")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--priority", "trivial",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify in DB
		var priority string
		err = db.QueryRowContext(context.Background(), `
			SELECT p.description
			FROM tasks t
			JOIN priorities p ON t.priority_id = p.id
			WHERE t.id = ?`, taskID).Scan(&priority)
		assert.NoError(t, err)
		assert.Equal(t, "trivial", priority)
	})

	t.Run("Update priority to low", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task Title")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--priority", "low",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify in DB
		var priority string
		err = db.QueryRowContext(context.Background(), `
			SELECT p.description
			FROM tasks t
			JOIN priorities p ON t.priority_id = p.id
			WHERE t.id = ?`, taskID).Scan(&priority)
		assert.NoError(t, err)
		assert.Equal(t, "low", priority)
	})

	t.Run("Update priority to medium", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task Title")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--priority", "medium",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify in DB
		var priority string
		err = db.QueryRowContext(context.Background(), `
			SELECT p.description
			FROM tasks t
			JOIN priorities p ON t.priority_id = p.id
			WHERE t.id = ?`, taskID).Scan(&priority)
		assert.NoError(t, err)
		assert.Equal(t, "medium", priority)
	})

	t.Run("Update priority to high", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task Title")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--priority", "high",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify in DB
		var priority string
		err = db.QueryRowContext(context.Background(), `
			SELECT p.description
			FROM tasks t
			JOIN priorities p ON t.priority_id = p.id
			WHERE t.id = ?`, taskID).Scan(&priority)
		assert.NoError(t, err)
		assert.Equal(t, "high", priority)
	})

	t.Run("Update priority to critical", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, columnID, "Task Title")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--priority", "critical",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify in DB
		var priority string
		err = db.QueryRowContext(context.Background(), `
			SELECT p.description
			FROM tasks t
			JOIN priorities p ON t.priority_id = p.id
			WHERE t.id = ?`, taskID).Scan(&priority)
		assert.NoError(t, err)
		assert.Equal(t, "critical", priority)
	})

	t.Run("Update multiple fields at once", func(t *testing.T) {
		// Create a task first
		taskID := cli.CreateTestTask(t, db, columnID, "Original Title")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--title", "New Title",
			"--description", "New Description",
			"--priority", "high",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify all fields updated in DB
		var title, description, priority string
		err = db.QueryRowContext(context.Background(), `
			SELECT t.title, t.description, p.description
			FROM tasks t
			JOIN priorities p ON t.priority_id = p.id
			WHERE t.id = ?`, taskID).Scan(&title, &description, &priority)
		assert.NoError(t, err)
		assert.Equal(t, "New Title", title)
		assert.Equal(t, "New Description", description)
		assert.Equal(t, "high", priority)
	})

	t.Run("Quiet mode output", func(t *testing.T) {
		// Create a task first
		taskID := cli.CreateTestTask(t, db, columnID, "Test Task")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--title", "Updated in Quiet Mode",
			"--quiet",
		})

		assert.NoError(t, err)
		// Quiet mode should only output the task ID
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)
		// Should be numeric
		taskIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, taskIDStr)
	})

	t.Run("JSON mode output", func(t *testing.T) {
		// Create a task first
		taskID := cli.CreateTestTask(t, db, columnID, "Test Task")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--title", "Updated in JSON Mode",
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure
		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(taskID), result["task_id"])
	})

	t.Run("Default human-readable output", func(t *testing.T) {
		// Create a task first
		taskID := cli.CreateTestTask(t, db, columnID, "Test Task")

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--title", "Updated in Normal Mode",
		})

		assert.NoError(t, err)
		// Default mode should have success message
		assert.Contains(t, output, "✓ Task")
		assert.Contains(t, output, fmt.Sprintf("%d", taskID))
		assert.Contains(t, output, "updated successfully")
	})

	t.Run("Verify unchanged fields remain intact", func(t *testing.T) {
		// Create a task with specific values
		taskID := cli.CreateTestTask(t, db, columnID, "Original Title")

		// Set initial description
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET description = ? WHERE id = ?", "Original Description", taskID)
		assert.NoError(t, err)

		// First, set priority to high
		cmd1 := UpdateCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--priority", "high",
			"--quiet",
		})
		assert.NoError(t, err)

		// Now update only title, verify description and priority unchanged
		cmd2 := UpdateCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--title", "Only Title Changed",
			"--quiet",
		})
		assert.NoError(t, err)

		// Verify all fields
		var title, description, priority string
		err = db.QueryRowContext(context.Background(), `
			SELECT t.title, t.description, p.description
			FROM tasks t
			JOIN priorities p ON t.priority_id = p.id
			WHERE t.id = ?`, taskID).Scan(&title, &description, &priority)
		assert.NoError(t, err)
		assert.Equal(t, "Only Title Changed", title)
		assert.Equal(t, "Original Description", description) // unchanged
		assert.Equal(t, "high", priority)                    // unchanged
	})

	t.Run("Update empty description", func(t *testing.T) {
		// Create a task with a description
		taskID := cli.CreateTestTask(t, db, columnID, "Task Title")

		// Set initial description
		_, err := db.ExecContext(context.Background(),
			"UPDATE tasks SET description = ? WHERE id = ?", "Has Description", taskID)
		assert.NoError(t, err)

		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--description", "",
			"--quiet",
		})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d\n", taskID), output)

		// Verify description is now empty
		var description sql.NullString
		err = db.QueryRowContext(context.Background(),
			"SELECT description FROM tasks WHERE id = ?", taskID).Scan(&description)
		assert.NoError(t, err)
		assert.Equal(t, "", description.String)
	})

	t.Run("Update title and description separately preserves both", func(t *testing.T) {
		// Create a task
		taskID := cli.CreateTestTask(t, db, columnID, "Original Title")

		// Update title first
		cmd1 := UpdateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd1, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--title", "Updated Title",
			"--quiet",
		})
		assert.NoError(t, err)

		// Update description second
		cmd2 := UpdateCmd()
		_, err = cli.ExecuteCLICommand(t, app, cmd2, []string{
			"--id", fmt.Sprintf("%d", taskID),
			"--description", "Updated Description",
			"--quiet",
		})
		assert.NoError(t, err)

		// Verify both are updated
		var title, description string
		err = db.QueryRowContext(context.Background(),
			"SELECT title, description FROM tasks WHERE id = ?", taskID).Scan(&title, &description)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", title)
		assert.Equal(t, "Updated Description", description)
	})
}

func TestUpdateTask_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project
	_ = cli.CreateTestProject(t, db, "Test Project")

	t.Run("Error with non-existent task ID", func(t *testing.T) {
		cmd := UpdateCmd()

		// Use a task ID that doesn't exist
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "99999",
			"--title", "Updated Title",
		})

		// Should fail because task doesn't exist
		assert.Error(t, err)
	})

	t.Run("Error with negative task ID", func(t *testing.T) {
		cmd := UpdateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "-1",
			"--title", "Updated Title",
		})

		// Should fail with negative task ID
		assert.Error(t, err)
	})

	t.Run("Error with zero task ID", func(t *testing.T) {
		cmd := UpdateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "0",
			"--title", "Updated Title",
		})

		// Should fail with zero task ID
		assert.Error(t, err)
	})
}
