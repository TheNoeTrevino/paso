package label_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestCreateLabel_Integration(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	projectID := cli.CreateTestProject(t, db, "Test Project")

	tests := []struct {
		name         string
		flags        []string
		expectError  bool
		verifyOutput func(t *testing.T, output string)
		verifyDB     func(t *testing.T)
	}{
		{
			name: "Create label with basic flags",
			flags: []string{
				"--name", "bug",
				"--color", "#FF0000",
				"--project", fmt.Sprintf("%d", projectID),
			},
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "bug")
				assert.Contains(t, output, "Test Project")
			},
			verifyDB: func(t *testing.T) {
				var name, color string
				err := db.QueryRowContext(context.Background(),
					"SELECT name, color FROM labels WHERE project_id = ? AND name = 'bug'",
					projectID).Scan(&name, &color)
				assert.NoError(t, err)
				assert.Equal(t, "bug", name)
				assert.Equal(t, "#FF0000", color)
			},
		},
		{
			name: "Create label with JSON output",
			flags: []string{
				"--name", "feature",
				"--color", "#00FF00",
				"--project", fmt.Sprintf("%d", projectID),
				"--json",
			},
			verifyOutput: func(t *testing.T, output string) {
				var result map[string]any
				err := json.Unmarshal([]byte(output), &result)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.True(t, result["success"].(bool))

				data := result["data"].(map[string]any)
				assert.Equal(t, "feature", data["Name"])
				assert.Equal(t, "#00FF00", data["Color"])
				assert.Equal(t, float64(projectID), data["ProjectID"])
			},
		},
		{
			name: "Create label with quiet mode",
			flags: []string{
				"--name", "urgent",
				"--color", "#FFFF00",
				"--project", fmt.Sprintf("%d", projectID),
				"--quiet",
			},
			verifyOutput: func(t *testing.T, output string) {
				trimmed := strings.TrimSpace(output)
				assert.Regexp(t, regexp.MustCompile(`^\d+$`), trimmed,
					"Quiet mode should return only label ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
		t.Parallel()
			cmd := label.CreateCmd()
			output, err := cli.ExecuteCLICommand(t, app, cmd, tt.flags)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.verifyOutput != nil {
				tt.verifyOutput(t, output)
			}
			if tt.verifyDB != nil {
				tt.verifyDB(t)
			}
		})
	}
}

func TestCreateLabel_ErrorCases(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	projectID := cli.CreateTestProject(t, db, "Test Project")

	t.Run("invalid project ID", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "test-label",
			"--color", "#FF0000",
			"--project", "99999",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "project 99999 not found")
	})

	t.Run("invalid color format", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "bad-color",
			"--color", "not-a-color",
			"--project", fmt.Sprintf("%d", projectID),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid color")
	})

	t.Run("short hex color rejected", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "short-color",
			"--color", "#FFF",
			"--project", fmt.Sprintf("%d", projectID),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid color")
	})
}
