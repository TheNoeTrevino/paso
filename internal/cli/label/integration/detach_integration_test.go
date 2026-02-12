package label_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestDetachLabel_Integration(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()
	projectID := cli.CreateTestProject(t, db, "Test Project")

	todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("detach label from task", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Detach Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "detach-me", "#FF0000")

		// Attach first
		cli.AttachLabelToTask(t, db, taskID, labelID)

		cmd := label.DetachCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", fmt.Sprintf("%d", taskID),
			"--label", fmt.Sprintf("%d", labelID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "detached")

		// Verify removal
		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?",
			taskID, labelID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("detach label JSON output", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, todoColumnID, "JSON Detach Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "json-detach", "#00FF00")

		cli.AttachLabelToTask(t, db, taskID, labelID)

		cmd := label.DetachCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", fmt.Sprintf("%d", taskID),
			"--label", fmt.Sprintf("%d", labelID),
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(taskID), result["task_id"])
		assert.Equal(t, float64(labelID), result["label_id"])
	})

	t.Run("detach label quiet mode", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Detach Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "quiet-detach", "#0000FF")

		cli.AttachLabelToTask(t, db, taskID, labelID)

		cmd := label.DetachCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", fmt.Sprintf("%d", taskID),
			"--label", fmt.Sprintf("%d", labelID),
			"--quiet",
		})

		assert.NoError(t, err)

		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?",
			taskID, labelID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("detach non-existent association is not an error", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		taskID := cli.CreateTestTask(t, db, todoColumnID, "No Association Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "no-association", "#AABBCC")

		cmd := label.DetachCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", fmt.Sprintf("%d", taskID),
			"--label", fmt.Sprintf("%d", labelID),
			"--quiet",
		})

		assert.NoError(t, err)
	})
}
