package column

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestDeleteColumnIntegration_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project with default columns
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Get the default "Todo" column ID for later use
	var todoColumnID int
	err := db.QueryRowContext(context.Background(),
		"SELECT id FROM columns WHERE project_id = ? AND name = 'Todo'",
		projectID).Scan(&todoColumnID)
	assert.NoError(t, err)

	t.Run("Delete column by ID", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "DeleteMe")

		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", columnID),
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify column was deleted from database
		var count int
		dbErr := db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 0, count, "Column should be deleted from database")
	})

	t.Run("Delete column with --quiet flag", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "QuietDelete")

		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", columnID),
			"--quiet",
		})

		assert.NoError(t, err, "Command should succeed")

		// Verify column was deleted
		var count int
		dbErr := db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 0, count, "Column should be deleted from database")
	})

	t.Run("Delete column with --json flag", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "JSONDelete")

		cmd := DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", columnID),
			"--json",
			"--force",
		})

		assert.NoError(t, err, "Command should succeed")

		// Parse JSON output
		var result map[string]interface{}
		jsonErr := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, jsonErr, "Output should be valid JSON")

		// Verify JSON structure
		assert.True(t, result["success"].(bool), "success field should be true")
		assert.Equal(t, float64(columnID), result["column_id"].(float64), "column_id should match")

		// Verify column was deleted
		var count int
		dbErr := db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 0, count, "Column should be deleted from database")
	})

	t.Run("Delete column with --force flag (skip confirmation)", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "ForceDelete")

		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", columnID),
			"--force",
		})

		assert.NoError(t, err, "Command should succeed")

		// Verify column was deleted
		var count int
		dbErr := db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 0, count, "Column should be deleted from database")
	})

	t.Run("Delete column cannot delete if it contains tasks", func(t *testing.T) {
		// Create a new column and add a task to it
		columnID := cli.CreateTestColumn(t, db, projectID, "CannotDelete")
		_ = cli.CreateTestTask(t, db, columnID, "Task in Column")

		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", columnID),
			"--force",
		})

		// Should error because column has tasks
		assert.Error(t, err, "Should error when trying to delete column with tasks")

		// Verify column still exists
		var count int
		dbErr := db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID).Scan(&count)
		assert.NoError(t, dbErr)
		assert.Equal(t, 1, count, "Column should not be deleted if it contains tasks")
	})

	t.Run("Delete column - human readable output", func(t *testing.T) {
		// Create a new column to delete
		columnID := cli.CreateTestColumn(t, db, projectID, "HumanDelete")

		cmd := DeleteCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", columnID),
			"--force",
		})

		assert.NoError(t, err, "Command should succeed")

		// Verify column was deleted
		var count int
		dbErr := db.QueryRowContext(context.Background(),
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
			"--id", fmt.Sprintf("%d", columnID1),
			"--quiet",
		})
		assert.NoError(t, err1)

		// Delete second column
		_, err2 := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", columnID2),
			"--quiet",
		})
		assert.NoError(t, err2)

		// Verify both are deleted
		var count1, count2 int
		db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID1).Scan(&count1)
		db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM columns WHERE id = ?", columnID2).Scan(&count2)

		assert.Equal(t, 0, count1, "First column should be deleted")
		assert.Equal(t, 0, count2, "Second column should be deleted")
	})
}

func TestDeleteColumnIntegration_Negative(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer db.Close()

	// Create test project (unused in skipped tests, but kept for potential future use)
	_ = cli.CreateTestProject(t, db, "Test Project")

	t.Run("Delete non-existent column", func(t *testing.T) {
		// Note: This test case calls os.Exit() in delete.go which terminates the process
		// We cannot test os.Exit() calls in integration tests without special handling
		// Skipping this test as it would terminate the test process
		t.Skip("Skipping test that calls os.Exit() - cannot test exit behavior in integration tests")

		// Expected behavior:
		// - Exit code: cli.ExitNotFound (3)
		// - Error message: "COLUMN_NOT_FOUND" with column ID that doesn't exist
		cmd := DeleteCmd()
		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "99999",
			"--quiet",
		})
	})

	t.Run("Delete column - missing required --id flag", func(t *testing.T) {
		// Note: This test case results in flag parsing error which calls os.Exit()
		// with ExitValidation code. Skipping as it would terminate the test process
		t.Skip("Skipping test that calls os.Exit() on missing required flag - cannot test exit behavior in integration tests")

		// Expected behavior:
		// - Exit code: cli.ExitValidation (5)
		// - Error message indicating --id flag is required
		cmd := DeleteCmd()
		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--quiet",
		})
	})

	t.Run("Delete column - zero column ID", func(t *testing.T) {
		// Note: This will trigger os.Exit() for non-existent column
		t.Skip("Skipping: command calls os.Exit() on invalid column ID")

		// Expected behavior:
		// - Exit code: cli.ExitNotFound (3)
		// - Column 0 does not exist
		cmd := DeleteCmd()
		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "0",
			"--quiet",
		})
	})

	t.Run("Delete column - negative column ID", func(t *testing.T) {
		// Note: This will trigger os.Exit() for non-existent column
		t.Skip("Skipping: command calls os.Exit() on invalid column ID")

		// Expected behavior:
		// - Exit code: cli.ExitNotFound (3)
		// - Column -1 does not exist
		cmd := DeleteCmd()
		_, _ = cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "-1",
			"--quiet",
		})
	})

	t.Run("Column ID as string instead of int", func(t *testing.T) {
		// This will fail at flag parsing level with invalid value error
		// The cobra library will report that "invalid" is not a valid int
		cmd := DeleteCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "invalid",
			"--quiet",
		})

		// Error should occur during flag parsing
		assert.Error(t, err, "Should error on invalid column ID format")
		// Output may contain the error message about invalid integer
		assert.True(t, err != nil || strings.Contains(output, "invalid"), "Should indicate invalid input")
	})
}
