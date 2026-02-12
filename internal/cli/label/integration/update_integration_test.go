package label_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestUpdateLabel(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()
	projectID := cli.CreateTestProject(t, db, "Test Project")

	t.Run("update label name only", func(t *testing.T) {
		t.Parallel()
		labelID := cli.CreateTestLabel(t, db, projectID, "original", "#FF0000")
		cmd := label.UpdateCmd()

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

	t.Run("update label color only", func(t *testing.T) {
		t.Parallel()
		labelID := cli.CreateTestLabel(t, db, projectID, "color-test", "#FF0000")
		cmd := label.UpdateCmd()

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

	t.Run("update both name and color", func(t *testing.T) {
		t.Parallel()
		labelID := cli.CreateTestLabel(t, db, projectID, "both-test", "#FF0000")
		cmd := label.UpdateCmd()

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

	t.Run("update label JSON output", func(t *testing.T) {
		t.Parallel()
		labelID := cli.CreateTestLabel(t, db, projectID, "json-update", "#AABBCC")
		cmd := label.UpdateCmd()

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

	t.Run("update label human-readable output", func(t *testing.T) {
		t.Parallel()
		labelID := cli.CreateTestLabel(t, db, projectID, "human-update", "#DDEEFF")
		cmd := label.UpdateCmd()

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
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	t.Run("invalid color format", func(t *testing.T) {
		t.Parallel()
		projectID := cli.CreateTestProject(t, db, "Color Test Project")
		labelID := cli.CreateTestLabel(t, db, projectID, "color-test", "#FF0000")
		cmd := label.UpdateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID),
			"--color", "not-a-color",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitValidation)
		assert.Contains(t, err.Error(), "color must be in hex format")
	})

	t.Run("invalid label ID format", func(t *testing.T) {
		t.Parallel()
		cmd := label.UpdateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid",
			"--name", "test",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitValidation)
		assert.Contains(t, err.Error(), "invalid ID")
	})
}
