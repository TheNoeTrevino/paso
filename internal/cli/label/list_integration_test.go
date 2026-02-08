package label

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestListLabels_Positive(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Create some labels
	cli.CreateTestLabel(t, db, projectID, "bug", "#FF0000")
	cli.CreateTestLabel(t, db, projectID, "feature", "#00FF00")
	cli.CreateTestLabel(t, db, projectID, "urgent", "#FFFF00")

	t.Run("List labels human-readable", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		cleanOutput := testutil.StripANSI(output)

		assert.Contains(t, cleanOutput, "bug")
		assert.Contains(t, cleanOutput, "feature")
		assert.Contains(t, cleanOutput, "urgent")
		assert.Contains(t, cleanOutput, "ID")
		assert.Contains(t, cleanOutput, "NAME")
		assert.Contains(t, cleanOutput, "COLOR")
	})

	t.Run("List labels JSON output", func(t *testing.T) {
		cmd := ListCmd()
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

	t.Run("List labels quiet mode", func(t *testing.T) {
		cmd := ListCmd()
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

	t.Run("List labels empty project", func(t *testing.T) {
		emptyProjectID := cli.CreateTestProject(t, db, "Empty Project")
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", emptyProjectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No labels found")
	})

	t.Run("List labels empty project JSON", func(t *testing.T) {
		emptyProjectID := cli.CreateTestProject(t, db, "Empty JSON Project")
		cmd := ListCmd()
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

	t.Run("List labels project isolation", func(t *testing.T) {
		otherProjectID := cli.CreateTestProject(t, db, "Other Project")
		cli.CreateTestLabel(t, db, otherProjectID, "other-label", "#0000FF")

		cmd := ListCmd()
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
	_, app := cli.SetupCLITest(t)

	t.Run("Missing project flag calls os.Exit", func(t *testing.T) {
		// The list command calls os.Exit(ExitUsage) when no project is specified
		// and git branch detection fails, so we cannot test this in-process.
		t.Skip("Skipping: command calls os.Exit() when no project is specified")
	})

	t.Run("Invalid project flag value", func(t *testing.T) {
		cmd := ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "not-a-number",
		})
		assert.Error(t, err)
	})
}
