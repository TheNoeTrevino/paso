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

func TestAttachLabel_Integration(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column
	todoColumnID := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("attach label to task", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Test Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "attach-me", "#FF0000")

		cmd := label.AttachCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", fmt.Sprintf("%d", taskID),
			"--label", fmt.Sprintf("%d", labelID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "attach-me")
		assert.Contains(t, output, fmt.Sprintf("#%d", taskID))

		// Verify in DB
		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?",
			taskID, labelID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("attach label JSON output", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "JSON Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "json-attach", "#00FF00")

		cmd := label.AttachCmd()
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

	t.Run("attach label quiet mode", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "quiet-attach", "#0000FF")

		cmd := label.AttachCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", fmt.Sprintf("%d", taskID),
			"--label", fmt.Sprintf("%d", labelID),
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify it was attached
		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM task_labels WHERE task_id = ? AND label_id = ?",
			taskID, labelID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestAttachLabel_Integration_Errors(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	t.Run("invalid task ID format", func(t *testing.T) {
		cmd := label.AttachCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", "not-a-number",
			"--label", "1",
		})
		assert.Error(t, err)
	})

	t.Run("invalid label ID format", func(t *testing.T) {
		cmd := label.AttachCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", "1",
			"--label", "not-a-number",
		})
		assert.Error(t, err)
	})
}
