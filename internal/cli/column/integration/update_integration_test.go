package column_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/cli/column"
	cliutil "github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestUpdateColumn_Integration(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cliutil.SetupCLITest(t)
	ctx := context.Background()

	// Create bare project (no default columns with flags) to avoid conflicts
	// when testing flag assignment on columns
	projectID := cliutil.CreateBareProject(t, db, "Test Project")

	t.Run("update column name only", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Original Name")

		cmd := column.UpdateCmd()

		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Updated Name",
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify in DB
		var name string
		err = db.QueryRowContext(ctx,
			"SELECT name FROM columns WHERE id = ?", testColumnID).Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", name)
	})

	t.Run("enable ready flag (holds_ready_tasks)", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Test Column")

		cmd := column.UpdateCmd()

		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--ready",
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify in DB
		var holdsReadyTasks bool
		err = db.QueryRowContext(ctx,
			"SELECT holds_ready_tasks FROM columns WHERE id = ?", testColumnID).Scan(&holdsReadyTasks)
		assert.NoError(t, err)
		assert.True(t, holdsReadyTasks)
	})

	t.Run("enable completed flag (holds_completed_tasks)", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Test Column")

		cmd := column.UpdateCmd()

		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--completed",
			"--force",
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify in DB
		var holdsCompletedTasks bool
		err = db.QueryRowContext(ctx,
			"SELECT holds_completed_tasks FROM columns WHERE id = ?", testColumnID).Scan(&holdsCompletedTasks)
		assert.NoError(t, err)
		assert.True(t, holdsCompletedTasks)
	})

	t.Run("enable in-progress flag (holds_in_progress_tasks)", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Test Column")

		cmd := column.UpdateCmd()

		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--in-progress",
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify in DB
		var holdsInProgressTasks bool
		err = db.QueryRowContext(ctx,
			"SELECT holds_in_progress_tasks FROM columns WHERE id = ?", testColumnID).Scan(&holdsInProgressTasks)
		assert.NoError(t, err)
		assert.True(t, holdsInProgressTasks)
	})

	t.Run("update name and in-progress flag together", func(t *testing.T) {
		// Create a separate bare project to avoid conflicts with other tests' in-progress flags
		testProjectID := cliutil.CreateBareProject(t, db, "Test Project for In-Progress")
		testColumnID := cliutil.CreateTestColumn(t, db, testProjectID, "Original Name")

		cmd := column.UpdateCmd()

		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Updated Name",
			"--in-progress",
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify in DB
		var name string
		var holdsInProgressTasks bool
		err = db.QueryRowContext(ctx,
			"SELECT name, holds_in_progress_tasks FROM columns WHERE id = ?", testColumnID).Scan(&name, &holdsInProgressTasks)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", name)
		assert.True(t, holdsInProgressTasks)
	})

	t.Run("update name and completed flag together with force", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Original Name")

		cmd := column.UpdateCmd()

		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Completed Column",
			"--completed",
			"--force",
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify in DB
		var name string
		var holdsCompletedTasks bool
		err = db.QueryRowContext(ctx,
			"SELECT name, holds_completed_tasks FROM columns WHERE id = ?", testColumnID).Scan(&name, &holdsCompletedTasks)
		assert.NoError(t, err)
		assert.Equal(t, "Completed Column", name)
		assert.True(t, holdsCompletedTasks)
	})

	t.Run("quiet mode output (no output)", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Test Column")

		cmd := column.UpdateCmd()

		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Updated in Quiet Mode",
			"--quiet",
		})

		assert.NoError(t, err)
	})

	t.Run("jSON mode output", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Test Column")

		cmd := column.UpdateCmd()

		output, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Updated in JSON Mode",
			"--json",
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, output)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure contains success
		assert.Equal(t, true, result["success"])
		assert.NotNil(t, result["column"])
	})

	t.Run("default human-readable output", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Test Column")

		cmd := column.UpdateCmd()

		output, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Updated in Normal Mode",
		})

		assert.NoError(t, err)
		// Default mode should have success message
		assert.Contains(t, output, "Column updated successfully")
		assert.Contains(t, output, fmt.Sprintf("%d", testColumnID))
	})

	t.Run("verify unchanged fields remain intact", func(t *testing.T) {
		// Create a separate bare project to avoid conflicts
		testProjectID := cliutil.CreateBareProject(t, db, "Test Project for Unchanged Fields")
		testColumnID := cliutil.CreateTestColumn(t, db, testProjectID, "Original Name")

		// First, enable in-progress flag
		cmd1 := column.UpdateCmd()
		_, err := cliutil.ExecuteCLICommand(t, app, cmd1, []string{
			fmt.Sprintf("%d", testColumnID),
			"--in-progress",
			"--quiet",
		})
		assert.NoError(t, err)

		// Now update only name, verify in-progress flag unchanged
		cmd2 := column.UpdateCmd()
		_, err = cliutil.ExecuteCLICommand(t, app, cmd2, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Only Name Changed",
			"--quiet",
		})
		assert.NoError(t, err)

		// Verify all fields
		var name string
		var holdsInProgressTasks bool
		err = db.QueryRowContext(ctx,
			"SELECT name, holds_in_progress_tasks FROM columns WHERE id = ?", testColumnID).Scan(&name, &holdsInProgressTasks)
		assert.NoError(t, err)
		assert.Equal(t, "Only Name Changed", name)
		assert.True(t, holdsInProgressTasks) // Should remain true
	})

	t.Run("jSON output includes column data", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Original Name")

		cmd := column.UpdateCmd()

		output, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "New Name",
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify column data is in response
		if columnData, ok := result["column"].(map[string]any); ok {
			assert.NotNil(t, columnData["old_name"])
			assert.NotNil(t, columnData["id"])
		}
	})

	t.Run("update same column multiple times sequentially", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Initial Name")

		// First update
		cmd1 := column.UpdateCmd()
		_, err := cliutil.ExecuteCLICommand(t, app, cmd1, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Second Name",
			"--quiet",
		})
		assert.NoError(t, err)

		// Second update
		cmd2 := column.UpdateCmd()
		_, err = cliutil.ExecuteCLICommand(t, app, cmd2, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Third Name",
			"--quiet",
		})
		assert.NoError(t, err)

		// Verify final state
		var name string
		err = db.QueryRowContext(ctx,
			"SELECT name FROM columns WHERE id = ?", testColumnID).Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "Third Name", name)
	})
}

func TestUpdateColumn_EdgeCases(t *testing.T) {
	t.Parallel()
	// Setup test DB and App
	db, app := cliutil.SetupCLITest(t)
	ctx := context.Background()

	// Create bare project (no default columns with flags) to avoid conflicts
	// when testing flag assignment on columns
	projectID := cliutil.CreateBareProject(t, db, "Test Project")

	t.Run("update column name to same name", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Original Name")

		cmd := column.UpdateCmd()

		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "Original Name",
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify name is still the same
		var name string
		err = db.QueryRowContext(ctx,
			"SELECT name FROM columns WHERE id = ?", testColumnID).Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "Original Name", name)
	})

	t.Run("update column name with special characters", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Original Name")

		cmd := column.UpdateCmd()

		specialName := "Todo & Done (Special!)"
		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", specialName,
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify name with special characters is preserved
		var name string
		err = db.QueryRowContext(ctx,
			"SELECT name FROM columns WHERE id = ?", testColumnID).Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, specialName, name)
	})

	t.Run("update column name with unicode characters", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Original Name")

		cmd := column.UpdateCmd()

		unicodeName := "待機中 (Waiting) 🚀"
		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", unicodeName,
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify unicode name is preserved
		var name string
		err = db.QueryRowContext(ctx,
			"SELECT name FROM columns WHERE id = ?", testColumnID).Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, unicodeName, name)
	})

	t.Run("update column name with maximum length", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Original Name")

		cmd := column.UpdateCmd()

		// Create a long name (50 characters - the max allowed)
		longName := strings.Repeat("A", 50)
		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", longName,
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify long name is preserved
		var name string
		err = db.QueryRowContext(ctx,
			"SELECT name FROM columns WHERE id = ?", testColumnID).Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, longName, name)
	})

	t.Run("ready flag toggles state appropriately", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Test Column")

		// First, enable ready flag
		cmd1 := column.UpdateCmd()
		_, err := cliutil.ExecuteCLICommand(t, app, cmd1, []string{
			fmt.Sprintf("%d", testColumnID),
			"--ready",
			"--quiet",
		})
		assert.NoError(t, err)

		// Verify ready flag is true
		var holdsReadyTasks bool
		err = db.QueryRowContext(ctx,
			"SELECT holds_ready_tasks FROM columns WHERE id = ?", testColumnID).Scan(&holdsReadyTasks)
		assert.NoError(t, err)
		assert.True(t, holdsReadyTasks)
	})

	t.Run("combined name and in-progress flag updates", func(t *testing.T) {
		// Create a test column
		testColumnID := cliutil.CreateTestColumn(t, db, projectID, "Original")

		cmd := column.UpdateCmd()

		_, err := cliutil.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", testColumnID),
			"--name", "In Progress Column",
			"--in-progress",
			"--quiet",
		})

		assert.NoError(t, err)

		// Verify both changes
		var name string
		var holdsInProgressTasks bool
		err = db.QueryRowContext(ctx,
			"SELECT name, holds_in_progress_tasks FROM columns WHERE id = ?", testColumnID).Scan(&name, &holdsInProgressTasks)
		assert.NoError(t, err)
		assert.Equal(t, "In Progress Column", name)
		assert.True(t, holdsInProgressTasks)
	})
}
