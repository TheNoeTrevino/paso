package label_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestListLabels(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Create some labels
	cli.CreateTestLabel(t, db, projectID, "bug", "#FF0000")
	cli.CreateTestLabel(t, db, projectID, "feature", "#00FF00")
	cli.CreateTestLabel(t, db, projectID, "urgent", "#FFFF00")

	t.Run("list labels human-readable", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		cleanOutput := fixtures.StripANSI(output)

		assert.Contains(t, cleanOutput, "bug")
		assert.Contains(t, cleanOutput, "feature")
		assert.Contains(t, cleanOutput, "urgent")
		assert.Contains(t, cleanOutput, "ID")
		assert.Contains(t, cleanOutput, "NAME")
		assert.Contains(t, cleanOutput, "COLOR")
	})

	t.Run("list labels JSON output", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		labels := result["labels"].([]any)
		assert.Len(t, labels, 3)

		firstLabel := labels[0].(map[string]any)
		assert.Equal(t, "bug", firstLabel["name"])
		assert.Equal(t, "#FF0000", firstLabel["color"])
		assert.Equal(t, float64(projectID), firstLabel["project_id"])
	})

	t.Run("list labels quiet mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--quiet",
		})

		assert.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 3)
		for _, line := range lines {
			assert.Regexp(t, `^\d+$`, line)
		}
	})

	t.Run("list labels empty project", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		emptyProjectID := cli.CreateTestProject(t, db, "Empty Project")
		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", emptyProjectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No labels found")
	})

	t.Run("list labels empty project JSON", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		emptyProjectID := cli.CreateTestProject(t, db, "Empty JSON Project")
		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", emptyProjectID),
			"--json",
		})

		assert.NoError(t, err)
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		labels := result["labels"].([]any)
		assert.Len(t, labels, 0)
	})

	t.Run("list labels project isolation", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		otherProjectID := cli.CreateTestProject(t, db, "Other Project")
		cli.CreateTestLabel(t, db, otherProjectID, "other-label", "#0000FF")

		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", otherProjectID),
			"--json",
		})

		assert.NoError(t, err)
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		labels := result["labels"].([]any)
		assert.Len(t, labels, 1)
		assert.Equal(t, "other-label", labels[0].(map[string]any)["name"])
	})
}

func TestListLabels_Errors(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	t.Run("missing project flag calls os.Exit", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		cli.AssertExitError(t, err, rootcli.ExitUsage)
		assert.Contains(t, err.Error(), "no project specified")
	})

	t.Run("invalid project flag value", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		cmd := label.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "not-a-number",
		})
		assert.Error(t, err)
	})
}
