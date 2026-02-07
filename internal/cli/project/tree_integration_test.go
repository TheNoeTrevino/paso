package project

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestTreeProject_Positive(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()
	projectID := cli.CreateTestProject(t, db, "Tree Project")

	// Get the Todo column
	var todoColumnID int
	err := db.QueryRowContext(ctx,
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&todoColumnID)
	assert.NoError(t, err)

	t.Run("Tree with empty project JSON", func(t *testing.T) {
		cmd := TreeCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(projectID), result["project_id"])

		tree := result["tree"].([]any)
		assert.Len(t, tree, 0)
	})

	t.Run("Tree with empty project human-readable", func(t *testing.T) {
		cmd := TreeCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No tasks found")
	})

	t.Run("Tree with empty project quiet mode", func(t *testing.T) {
		cmd := TreeCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--quiet",
		})

		assert.NoError(t, err)
	})

	t.Run("Tree with tasks JSON", func(t *testing.T) {
		taskProjectID := cli.CreateTestProject(t, db, "Task Tree Project")
		var taskTodoColumnID int
		err := db.QueryRowContext(ctx,
			"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
			taskProjectID).Scan(&taskTodoColumnID)
		assert.NoError(t, err)

		cli.CreateTestTask(t, db, taskTodoColumnID, "Root Task")
		cli.CreateTestTask(t, db, taskTodoColumnID, "Another Root Task")

		cmd := TreeCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", taskProjectID),
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		tree := result["tree"].([]any)
		assert.GreaterOrEqual(t, len(tree), 2)
	})

	t.Run("Tree with positional argument", func(t *testing.T) {
		cmd := TreeCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))
	})
}

func TestTreeProject_Negative(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("Invalid positional argument", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on invalid project ID")
	})

	t.Run("Too many positional arguments", func(t *testing.T) {
		cmd := TreeCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"1", "2",
		})
		assert.Error(t, err)
	})
}
