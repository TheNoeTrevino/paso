package task

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestCreateTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Test creating a task with just title and project ID
	t.Run("Create task with title only", func(t *testing.T) {
		cmd := CreateCmd()

		// Capture output to verify task ID is returned
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"--project", "1", "--title", "Simple Task", "--quiet"})

		assert.NoError(t, err)

		// Verify output contains a task ID (numeric)
		taskIDStr := strings.TrimSpace(output)
		assert.Regexp(t, `^\d+$`, taskIDStr)

		// Verify task exists in DB
		var title string
		err = db.QueryRowContext(context.Background(),
			"SELECT title FROM tasks WHERE id = ?", taskIDStr).Scan(&title)
		assert.NoError(t, err)
		assert.Equal(t, "Simple Task", title)
	})

	// Test creating a task with all fields
	t.Run("Create task with all fields", func(t *testing.T) {
		cmd := CreateCmd()

		// Create a column to put the task in (other than default)
		columnID := cli.CreateTestColumn(t, db, projectID, "CustomColumn")

		// For detailed tests, use the handler logic directly since flag parsing might be skipping defaults in test
		// This is safer because cobra flags rely on pflag which can be tricky in tests

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "1",
			"--title", "Detailed Task",
			"--description", "This is a detailed description",
			"--priority", "high",
			"--type", "feature",
			"--column", "CustomColumn",
			"--quiet",
		})

		assert.NoError(t, err)

		taskIDStr := strings.TrimSpace(output)

		// Verify task fields in DB
		var title, description, priority, taskType string
		var dbColumnID int

		err = db.QueryRowContext(context.Background(), `
			SELECT t.title, t.description, p.description, ty.description, t.column_id
			FROM tasks t
			JOIN priorities p ON t.priority_id = p.id
			JOIN types ty ON t.type_id = ty.id
			WHERE t.id = ?`, taskIDStr).Scan(&title, &description, &priority, &taskType, &dbColumnID)

		assert.NoError(t, err)
		assert.Equal(t, "Detailed Task", title)
		assert.Equal(t, "This is a detailed description", description)
		assert.Equal(t, "high", priority)
		assert.Equal(t, "feature", taskType)
		assert.Equal(t, columnID, dbColumnID)
	})
}
