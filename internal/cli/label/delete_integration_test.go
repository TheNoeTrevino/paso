package label

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestDeleteLabel_Positive(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()
	projectID := cli.CreateTestProject(t, db, "Test Project")

	t.Run("Delete label with force flag", func(t *testing.T) {
		labelID := cli.CreateTestLabel(t, db, projectID, "to-delete", "#FF0000")
		cmd := DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID),
			"--force",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "deleted successfully")

		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM labels WHERE id = ?", labelID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Delete label with quiet flag", func(t *testing.T) {
		labelID := cli.CreateTestLabel(t, db, projectID, "quiet-delete", "#00FF00")
		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID),
			"--quiet",
		})

		assert.NoError(t, err)

		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM labels WHERE id = ?", labelID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Delete label with JSON output", func(t *testing.T) {
		labelID := cli.CreateTestLabel(t, db, projectID, "json-delete", "#0000FF")
		cmd := DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID),
			"--json",
			"--force",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"].(bool))
		assert.Equal(t, float64(labelID), result["label_id"].(float64))

		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM labels WHERE id = ?", labelID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Delete multiple labels sequentially", func(t *testing.T) {
		labelID1 := cli.CreateTestLabel(t, db, projectID, "delete-1", "#111111")
		labelID2 := cli.CreateTestLabel(t, db, projectID, "delete-2", "#222222")

		cmd := DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID1),
			"--quiet",
		})
		assert.NoError(t, err)

		_, err = cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", labelID2),
			"--quiet",
		})
		assert.NoError(t, err)

		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM labels WHERE id IN (?, ?)", labelID1, labelID2).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestDeleteLabel_Negative(t *testing.T) {
	_, _ = cli.SetupCLITest(t)

	t.Run("Invalid label ID format", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on invalid ID format")
	})
}
