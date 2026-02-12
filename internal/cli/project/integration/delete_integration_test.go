package project_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestDeleteProject_Integration(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()

	t.Run("delete project with force flag", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "Force Delete Project")
		cmd := project.DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", projectID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("delete project with quiet flag", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "Quiet Delete Project")
		cmd := project.DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", projectID),
			"--quiet",
		})

		assert.NoError(t, err)

		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("delete project with JSON output", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "JSON Delete Project")
		cmd := project.DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", projectID),
			"--json",
			"--force",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(projectID), result["project_id"].(float64))
	})
}

func TestDeleteProject_Integration_Errors(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	t.Run("invalid project ID format", func(t *testing.T) {
		t.Parallel()
		cmd := project.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"abc",
			"--force",
			"--quiet",
		})
		cli.AssertExitError(t, err, 5)
		assert.Contains(t, err.Error(), "invalid ID 'abc'")
	})
}
