package label

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestAttachLabel_Positive(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column
	var todoColumnID int
	err := db.QueryRowContext(ctx,
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&todoColumnID)
	assert.NoError(t, err)

	t.Run("Attach label to task", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Test Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "attach-me", "#FF0000")

		cmd := AttachCmd()
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

	t.Run("Attach label JSON output", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "JSON Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "json-attach", "#00FF00")

		cmd := AttachCmd()
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

	t.Run("Attach label quiet mode", func(t *testing.T) {
		taskID := cli.CreateTestTask(t, db, todoColumnID, "Quiet Task")
		labelID := cli.CreateTestLabel(t, db, projectID, "quiet-attach", "#0000FF")

		cmd := AttachCmd()
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

func TestAttachLabel_Negative(t *testing.T) {
	_, app := cli.SetupCLITest(t)

	t.Run("Invalid task ID format", func(t *testing.T) {
		cmd := AttachCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", "not-a-number",
			"--label", "1",
		})
		assert.Error(t, err)
	})

	t.Run("Invalid label ID format", func(t *testing.T) {
		cmd := AttachCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", "1",
			"--label", "not-a-number",
		})
		assert.Error(t, err)
	})
}
