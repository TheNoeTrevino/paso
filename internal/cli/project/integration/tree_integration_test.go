package project_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestTreeProject(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	projectID := cli.CreateTestProject(t, db, "Tree Project")

	t.Run("tree with empty project JSON", func(t *testing.T) {
		cmd := project.TreeCmd()
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

	t.Run("tree with empty project human-readable", func(t *testing.T) {
		cmd := project.TreeCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No tasks found")
	})

	t.Run("tree with empty project quiet mode", func(t *testing.T) {
		cmd := project.TreeCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--quiet",
		})

		assert.NoError(t, err)
	})

	t.Run("tree with tasks JSON", func(t *testing.T) {
		taskProjectID := cli.CreateTestProject(t, db, "Task Tree Project")
		taskTodoColumnID := cli.GetColumnIDByName(t, db, taskProjectID, "Todo")

		cli.CreateTestTask(t, db, taskTodoColumnID, "Root Task")
		cli.CreateTestTask(t, db, taskTodoColumnID, "Another Root Task")

		cmd := project.TreeCmd()
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

	t.Run("tree with positional argument", func(t *testing.T) {
		cmd := project.TreeCmd()
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

func TestTreeProject_Errors(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	t.Run("invalid positional argument", func(t *testing.T) {
		cmd := project.TreeCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid",
		})
		cli.AssertExitError(t, err, 2)
		assert.Contains(t, err.Error(), "project ID must be a positive integer")
	})

	t.Run("too many positional arguments", func(t *testing.T) {
		cmd := project.TreeCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"1", "2",
		})
		assert.Error(t, err)
	})
}
