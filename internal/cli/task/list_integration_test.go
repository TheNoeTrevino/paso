package task

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestListTask_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Create some tasks
	col1 := cli.CreateTestColumn(t, db, projectID, "Col1")
	task1 := cli.CreateTestTask(t, db, col1, "Task 1")
	task2 := cli.CreateTestTask(t, db, col1, "Task 2")

	// Create another column with a task
	col2 := cli.CreateTestColumn(t, db, projectID, "Col2")
	task3 := cli.CreateTestTask(t, db, col2, "Task 3")

	t.Run("List tasks human readable", func(t *testing.T) {
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "1",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Found 3 tasks")
		assert.Contains(t, output, "Task 1")
		assert.Contains(t, output, "Task 2")
		assert.Contains(t, output, "Task 3")
	})

	t.Run("List tasks quiet mode", func(t *testing.T) {
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "1",
			"--quiet",
		})

		assert.NoError(t, err)

		// Should contain just the IDs
		ids := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, ids, 3)
		assert.Contains(t, ids, convertIntToString(task1))
		assert.Contains(t, ids, convertIntToString(task2))
		assert.Contains(t, ids, convertIntToString(task3))
	})

	t.Run("List tasks JSON mode", func(t *testing.T) {
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "1",
			"--json",
		})

		assert.NoError(t, err)

		var result struct {
			Success bool                  `json:"success"`
			Tasks   []*models.TaskSummary `json:"tasks"`
		}

		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Len(t, result.Tasks, 3)

		// Verify tasks are present
		foundTask1 := false
		for _, task := range result.Tasks {
			if task.ID == task1 {
				foundTask1 = true
				assert.Equal(t, "Task 1", task.Title)
			}
		}
		assert.True(t, foundTask1)
	})
}

// Helper to convert int to string for assertions
func convertIntToString(i int) string {
	return fmt.Sprintf("%d", i)
}
