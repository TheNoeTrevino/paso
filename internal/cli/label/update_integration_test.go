package label

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestUpdateLabel_Positive(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()
	projectID := cli.CreateTestProject(t, db, "Test Project")

	t.Run("Update label name only", func(t *testing.T) {
		labelID := cli.CreateTestLabel(t, db, projectID, "original", "#FF0000")
		cmd := UpdateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID),
			"--name", "updated-name",
			"--quiet",
		})

		assert.NoError(t, err)

		var name, color string
		err = db.QueryRowContext(ctx,
			"SELECT name, color FROM labels WHERE id = ?", labelID).Scan(&name, &color)
		assert.NoError(t, err)
		assert.Equal(t, "updated-name", name)
		assert.Equal(t, "#FF0000", color, "Color should remain unchanged")
	})

	t.Run("Update label color only", func(t *testing.T) {
		labelID := cli.CreateTestLabel(t, db, projectID, "color-test", "#FF0000")
		cmd := UpdateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID),
			"--color", "#00FF00",
			"--quiet",
		})

		assert.NoError(t, err)

		var name, color string
		err = db.QueryRowContext(ctx,
			"SELECT name, color FROM labels WHERE id = ?", labelID).Scan(&name, &color)
		assert.NoError(t, err)
		assert.Equal(t, "color-test", name, "Name should remain unchanged")
		assert.Equal(t, "#00FF00", color)
	})

	t.Run("Update both name and color", func(t *testing.T) {
		labelID := cli.CreateTestLabel(t, db, projectID, "both-test", "#FF0000")
		cmd := UpdateCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID),
			"--name", "new-name",
			"--color", "#0000FF",
			"--quiet",
		})

		assert.NoError(t, err)

		var name, color string
		err = db.QueryRowContext(ctx,
			"SELECT name, color FROM labels WHERE id = ?", labelID).Scan(&name, &color)
		assert.NoError(t, err)
		assert.Equal(t, "new-name", name)
		assert.Equal(t, "#0000FF", color)
	})

	t.Run("Update label JSON output", func(t *testing.T) {
		labelID := cli.CreateTestLabel(t, db, projectID, "json-update", "#AABBCC")
		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID),
			"--name", "json-updated",
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))

		labelData := result["label"].(map[string]any)
		assert.Equal(t, float64(labelID), labelData["id"])
		assert.Equal(t, "json-updated", labelData["name"])
		assert.Equal(t, "json-update", labelData["old_name"])
	})

	t.Run("Update label human-readable output", func(t *testing.T) {
		labelID := cli.CreateTestLabel(t, db, projectID, "human-update", "#DDEEFF")
		cmd := UpdateCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID),
			"--name", "renamed",
			"--color", "#112233",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "updated successfully")
	})
}

func TestUpdateLabel_ErrorCases(t *testing.T) {
	_, _ = cli.SetupCLITest(t)

	t.Run("Invalid color format calls os.Exit", func(t *testing.T) {
		// This calls os.Exit via ExitValidation, which we cannot capture in-process.
		// Skipping as it would terminate the test process.
		t.Skip("Skipping: command calls os.Exit() on invalid color format")
	})

	t.Run("Invalid label ID format", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on invalid ID format")
	})
}
