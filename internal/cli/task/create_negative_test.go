package task

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestCreateTask_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Task 29: Test Invalid Column ID
	t.Run("Create task with invalid column", func(t *testing.T) {
		cmd := CreateCmd()

		// Attempt to create a task in a non-existent column
		// Note: The CLI takes column NAME, not ID. So we test with a non-existent name.
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--title", "Invalid Column Task",
			"--column", "NonExistentColumn",
			"--quiet",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "column 'NonExistentColumn' not found")
	})

	// Create a column so we can test missing title with valid project/column
	cli.CreateTestColumn(t, db, projectID, "Todo")

	t.Run("Create task missing title", func(t *testing.T) {
		cmd := CreateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--column", "Todo",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})
}
