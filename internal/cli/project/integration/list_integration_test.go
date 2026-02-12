package project_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestListProjects(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	t.Run("list projects when none exist", func(t *testing.T) {
		t.Parallel()
		cmd := project.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{})

		assert.NoError(t, err)
		assert.Contains(t, output, "No projects found")
	})

	t.Run("list projects JSON when none exist", func(t *testing.T) {
		t.Parallel()
		cmd := project.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--json"})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		projects := result["projects"].([]any)
		assert.Len(t, projects, 0)
	})

	t.Run("list projects quiet when none exist", func(t *testing.T) {
		t.Parallel()
		cmd := project.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--quiet"})

		assert.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(output))
	})
}

func TestListProjects_WithData(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	cli.CreateTestProject(t, db, "Alpha Project")
	cli.CreateTestProject(t, db, "Beta Project")
	cli.CreateTestProject(t, db, "Gamma Project")

	t.Run("list projects JSON with data", func(t *testing.T) {
		t.Parallel()
		cmd := project.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--json"})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		projects := result["projects"].([]any)
		assert.Len(t, projects, 3)
	})

	t.Run("list projects quiet with data", func(t *testing.T) {
		t.Parallel()
		cmd := project.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--quiet"})

		assert.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 3)
		for _, line := range lines {
			assert.Regexp(t, `^\d+$`, line)
		}
	})

	t.Run("list projects human-readable with data", func(t *testing.T) {
		t.Parallel()
		cmd := project.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{})

		assert.NoError(t, err)
		assert.Contains(t, output, "Alpha Project")
		assert.Contains(t, output, "Beta Project")
		assert.Contains(t, output, "Gamma Project")
	})
}
