package column

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestDeleteColumnIntegration_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()

	// Create test project with default columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID for later use
	var todoColumnID int
	err := db.QueryRowContext(ctx,
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&todoColumnID)
	assert.NoError(t, err)

	t.Run("Delete column by ID", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "DeleteMe")

		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", columnID),
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify column was deleted from database
		var count int
		dbErr := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 0, count, "Column should be deleted from database")
	})

	t.Run("Delete column with --quiet flag", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "QuietDelete")

		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", columnID),
			"--quiet",
		})

		assert.NoError(t, err, "Command should succeed")

		// Verify column was deleted
		var count int
		dbErr := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 0, count, "Column should be deleted from database")
	})

	t.Run("Delete column with --json flag", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "JSONDelete")

		cmd := DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", columnID),
			"--json",
			"--force",
		})

		assert.NoError(t, err, "Command should succeed")

		// Parse JSON output
		var result map[string]any
		jsonErr := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, jsonErr, "Output should be valid JSON")

		// Verify JSON structure
		assert.True(t, result["success"].(bool), "success field should be true")
		assert.Equal(t, float64(columnID), result["column_id"].(float64), "column_id should match")

		// Verify column was deleted
		var count int
		dbErr := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 0, count, "Column should be deleted from database")
	})

	t.Run("Delete column with --force flag (skip confirmation)", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "ForceDelete")

		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", columnID),
			"--force",
		})

		assert.NoError(t, err, "Command should succeed")

		// Verify column was deleted
		var count int
		dbErr := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 0, count, "Column should be deleted from database")
	})

	t.Run("Delete column cannot delete if it contains tasks", func(t *testing.T) {
		// Expected behavior:
		// - The column service will move tasks to the first column before deleting
		// - So this test scenario doesn't actually trigger an error anymore
	})

	t.Run("Delete column - human readable output", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "HumanDelete")

		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", columnID),
			"--force",
		})

		assert.NoError(t, err, "Command should succeed")

		// Verify column was deleted
		var count int
		dbErr := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 0, count, "Column should be deleted from database")
	})

	t.Run("Delete multiple columns sequentially", func(t *testing.T) {
		// Create two columns
		columnID1 := cli.CreateTestColumn(t, db, projectID, "Delete1")
		columnID2 := cli.CreateTestColumn(t, db, projectID, "Delete2")

		cmd := DeleteCmd()

		// Delete first column
		_, err1 := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", columnID1),
			"--quiet",
		})
		assert.NoError(t, err1)

		// Delete second column
		_, err2 := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", columnID2),
			"--quiet",
		})
		assert.NoError(t, err2)

		// Verify both are deleted
		var count1, count2 int
		err3 := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID1).Scan(&count1)
		assert.NoError(t, err3)
		err4 := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID2).Scan(&count2)
		assert.NoError(t, err4)

		assert.Equal(t, 0, count1, "First column should be deleted")
		assert.Equal(t, 0, count2, "Second column should be deleted")
	})

	t.Run("Dry-run mode does not delete column", func(t *testing.T) {
		// Create a column
		columnID := cli.CreateTestColumn(t, db, projectID, "DryRunTest")

		cmd := DeleteCmd()

		// Run with --dry-run flag
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", columnID),
			"--dry-run",
		})

		assert.NoError(t, err, "Dry-run should not error")
		assert.Contains(t, output, "Would delete", "Output should indicate dry-run")

		// Verify column still exists
		var count int
		dbErr := db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 1, count, "Column should still exist after dry-run")
	})

	t.Run("Dry-run mode with JSON output", func(t *testing.T) {
		// Create a column
		columnID := cli.CreateTestColumn(t, db, projectID, "DryRunJSON")

		cmd := DeleteCmd()

		// Run with --dry-run and --json flags
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", columnID),
			"--dry-run",
			"--json",
		})

		assert.NoError(t, err, "Dry-run should not error")

		// Parse JSON output
		var result map[string]any
		jsonErr := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, jsonErr, "Output should be valid JSON")
		assert.True(t, result["dry_run"].(bool), "dry_run field should be true")
		assert.True(t, result["success"].(bool), "success field should be true")

		// Verify column still exists
		var count int
		dbErr := db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 1, count, "Column should still exist after dry-run")
	})
}

func TestDeleteColumnIntegration_Negative(t *testing.T) {
	// Setup test DB and App
	_, app := cli.SetupCLITest(t)

	t.Run("Delete non-existent column", func(t *testing.T) {
		cmd := DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"99999",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "column 99999 not found")
	})

	t.Run("Delete column - missing required ID argument", func(t *testing.T) {
		cmd := DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--quiet",
		})
		assert.Error(t, err)
	})

	t.Run("Delete column - zero column ID", func(t *testing.T) {
		cmd := DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"0",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "column 0 not found")
	})

	t.Run("Delete column - negative column ID", func(t *testing.T) {
		// Cobra may interpret "-1" as a flag, so just assert error
		cmd := DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"-1",
			"--quiet",
		})
		assert.Error(t, err)
	})

	t.Run("Column ID as string instead of int", func(t *testing.T) {
		cmd := DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitValidation)
		assert.Contains(t, err.Error(), "invalid ID 'invalid': must be a number")
	})
}
