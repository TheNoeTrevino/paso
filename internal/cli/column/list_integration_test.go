package column

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestListColumns_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Set flags on default columns for testing
	_, err := db.ExecContext(ctx, `
		UPDATE columns
		SET holds_ready_tasks = 1
		WHERE project_id = ? AND name = 'Todo'`, projectID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		UPDATE columns
		SET holds_in_progress_tasks = 1
		WHERE project_id = ? AND name = 'In Progress'`, projectID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		UPDATE columns
		SET holds_completed_tasks = 1
		WHERE project_id = ? AND name = 'Done'`, projectID)
	require.NoError(t, err)

	t.Run("List columns with project flag", func(t *testing.T) {
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)

		// Strip ANSI codes for easier assertions
		cleanOutput := testutil.StripANSI(output)

		// Verify table headers are present (new format uses separate columns)
		assert.Contains(t, cleanOutput, "ID")
		assert.Contains(t, cleanOutput, "NAME")
		assert.Contains(t, cleanOutput, "READY")
		assert.Contains(t, cleanOutput, "IN-PROG")
		assert.Contains(t, cleanOutput, "DONE")

		// Default columns created with project: Todo, In Progress, Done
		assert.Contains(t, cleanOutput, "Todo")
		assert.Contains(t, cleanOutput, "In Progress")
		assert.Contains(t, cleanOutput, "Done")
	})

	t.Run("List columns with JSON output", func(t *testing.T) {
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))
		columns := result["columns"].([]any)
		assert.Len(t, columns, 3)

		// Verify first column (Todo)
		firstCol := columns[0].(map[string]any)
		assert.Equal(t, "Todo", firstCol["name"])
		assert.True(t, firstCol["holds_ready_tasks"].(bool))

		// Verify second column (In Progress)
		secondCol := columns[1].(map[string]any)
		assert.Equal(t, "In Progress", secondCol["name"])
		assert.True(t, secondCol["holds_in_progress_tasks"].(bool))

		// Verify third column (Done)
		thirdCol := columns[2].(map[string]any)
		assert.Equal(t, "Done", thirdCol["name"])
		assert.True(t, thirdCol["holds_completed_tasks"].(bool))
	})

	t.Run("List columns with quiet mode", func(t *testing.T) {
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--quiet",
		})

		assert.NoError(t, err)

		// In quiet mode, output should be just IDs, one per line
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Len(t, lines, 3)

		// Each line should be a numeric ID
		for _, line := range lines {
			assert.Regexp(t, `^\d+$`, line)
		}
	})

	t.Run("List columns when none exist (empty project)", func(t *testing.T) {
		// Create a project with no columns (we need to manually create one without default columns)
		// Since CreateTestProject creates default columns, we create a project directly
		var emptyProjectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Empty Project").Scan(&emptyProjectID)
		assert.NoError(t, err)

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", emptyProjectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "No columns found in project 'Empty Project'")
	})

	t.Run("List columns and verify sorting by position", func(t *testing.T) {
		// Create a project with custom columns in specific order
		var customProjectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Ordered Project").Scan(&customProjectID)
		assert.NoError(t, err)

		// Create columns with specific names
		column1ID := cli.CreateTestColumn(t, db, customProjectID, "First")
		column2ID := cli.CreateTestColumn(t, db, customProjectID, "Second")
		column3ID := cli.CreateTestColumn(t, db, customProjectID, "Third")

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", customProjectID),
		})

		assert.NoError(t, err)

		// Strip ANSI codes for easier assertions
		cleanOutput := testutil.StripANSI(output)

		// Verify columns appear in order in the table
		firstPos := strings.Index(cleanOutput, "First")
		secondPos := strings.Index(cleanOutput, "Second")
		thirdPos := strings.Index(cleanOutput, "Third")

		assert.Greater(t, firstPos, -1, "First column should appear in output")
		assert.Greater(t, secondPos, -1, "Second column should appear in output")
		assert.Greater(t, thirdPos, -1, "Third column should appear in output")
		assert.Less(t, firstPos, secondPos, "First should appear before Second")
		assert.Less(t, secondPos, thirdPos, "Second should appear before Third")

		// Verify column IDs appear in the table
		assert.Contains(t, cleanOutput, fmt.Sprintf("%d", column1ID))
		assert.Contains(t, cleanOutput, fmt.Sprintf("%d", column2ID))
		assert.Contains(t, cleanOutput, fmt.Sprintf("%d", column3ID))
	})

	t.Run("List columns with column flags set", func(t *testing.T) {
		// Create a project and modify column flags
		var flagProjectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Flag Project").Scan(&flagProjectID)
		assert.NoError(t, err)

		// Create a column and set flags
		flagColumnID := cli.CreateTestColumn(t, db, flagProjectID, "Custom")
		_, err = db.ExecContext(ctx, `
			UPDATE columns
			SET holds_ready_tasks = 1, holds_in_progress_tasks = 1
			WHERE id = ?`,
			flagColumnID)
		assert.NoError(t, err)

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", flagProjectID),
		})

		assert.NoError(t, err)
		cleanOutput := testutil.StripANSI(output)
		assert.Contains(t, cleanOutput, "Custom")
	})

	t.Run("List columns in JSON mode with complete structure", func(t *testing.T) {
		// Create a project with multiple columns
		var jsonProjectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"JSON Project").Scan(&jsonProjectID)
		assert.NoError(t, err)

		// Create columns
		column1ID := cli.CreateTestColumn(t, db, jsonProjectID, "Backlog")
		column2ID := cli.CreateTestColumn(t, db, jsonProjectID, "Active")

		// Set flags on columns
		_, err = db.ExecContext(ctx, `
			UPDATE columns
			SET holds_ready_tasks = 1
			WHERE id = ?`,
			column1ID)
		assert.NoError(t, err)

		_, err = db.ExecContext(ctx, `
			UPDATE columns
			SET holds_in_progress_tasks = 1
			WHERE id = ?`,
			column2ID)
		assert.NoError(t, err)

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", jsonProjectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify structure
		assert.True(t, result["success"].(bool))
		columns := result["columns"].([]any)
		assert.Len(t, columns, 2)

		// Verify first column structure
		col1 := columns[0].(map[string]any)
		assert.Equal(t, float64(column1ID), col1["id"])
		assert.Equal(t, "Backlog", col1["name"])
		assert.Equal(t, float64(jsonProjectID), col1["project_id"])
		assert.True(t, col1["holds_ready_tasks"].(bool))
		assert.False(t, col1["holds_completed_tasks"].(bool))

		// Verify second column structure
		col2 := columns[1].(map[string]any)
		assert.Equal(t, float64(column2ID), col2["id"])
		assert.Equal(t, "Active", col2["name"])
		assert.Equal(t, float64(jsonProjectID), col2["project_id"])
		assert.True(t, col2["holds_in_progress_tasks"].(bool))
	})

	t.Run("List columns with multiple projects", func(t *testing.T) {
		// Create another project
		var projectID2 int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Second Project").Scan(&projectID2)
		assert.NoError(t, err)

		// Create columns in second project
		cli.CreateTestColumn(t, db, projectID2, "Custom1")
		cli.CreateTestColumn(t, db, projectID2, "Custom2")

		// List columns for second project
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID2),
		})

		assert.NoError(t, err)

		// Strip ANSI codes for easier assertions
		cleanOutput := testutil.StripANSI(output)

		// Verify second project's columns appear in table
		assert.Contains(t, cleanOutput, "Custom1")
		assert.Contains(t, cleanOutput, "Custom2")

		// Verify first project's columns are NOT in the output (project isolation)
		assert.NotContains(t, cleanOutput, "Todo")
		assert.NotContains(t, cleanOutput, "In Progress")
		assert.NotContains(t, cleanOutput, "Done")
	})

	t.Run("List columns with all flag combinations", func(t *testing.T) {
		// Test quiet and JSON together (quiet should take precedence)
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--quiet",
			"--json",
		})

		assert.NoError(t, err)

		// Quiet mode should output just IDs
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Greater(t, len(lines), 0)

		// Each line should be a numeric ID (not JSON)
		for _, line := range lines {
			assert.Regexp(t, `^\d+$`, line)
		}
	})
}

func TestListColumns_Errors(t *testing.T) {
	// Setup test DB and App
	_, app := cli.SetupCLITest(t)

	t.Run("Invalid project ID error handling", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on project not found")
	})

	t.Run("Missing project flag", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() when project cannot be resolved")
	})

	t.Run("Invalid project flag value", func(t *testing.T) {
		cmd := ListCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "not-a-number",
		})

		// Should have an error since value is not numeric
		assert.Error(t, err)
	})
}

func TestListColumns_EdgeCases(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	ctx := context.Background()

	t.Run("List columns with special characters in names", func(t *testing.T) {
		var projectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Special Project").Scan(&projectID)
		assert.NoError(t, err)

		// Create columns with special characters
		cli.CreateTestColumn(t, db, projectID, "To-Do")
		cli.CreateTestColumn(t, db, projectID, "In_Progress")
		cli.CreateTestColumn(t, db, projectID, "Done/Shipped")

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "To-Do")
		assert.Contains(t, output, "In_Progress")
		assert.Contains(t, output, "Done/Shipped")
	})

	t.Run("List columns with long names", func(t *testing.T) {
		var projectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Long Names Project").Scan(&projectID)
		assert.NoError(t, err)

		longName := "This is a very long column name that should still display properly"
		cli.CreateTestColumn(t, db, projectID, longName)

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, longName)
	})

	t.Run("List columns preserves insertion order", func(t *testing.T) {
		var projectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Order Project").Scan(&projectID)
		assert.NoError(t, err)

		// Create columns and remember their IDs
		ids := make([]int, 5)
		names := []string{"A", "B", "C", "D", "E"}
		for i, name := range names {
			ids[i] = cli.CreateTestColumn(t, db, projectID, name)
		}

		cmd := ListCmd()

		// Get JSON output to verify order
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		columns := result["columns"].([]any)
		assert.Len(t, columns, 5)

		// Verify order matches insertion order
		for i, colInterface := range columns {
			col := colInterface.(map[string]any)
			assert.Equal(t, names[i], col["name"])
		}
	})

	t.Run("List columns with no flags formatting", func(t *testing.T) {
		// Create a project with columns that have no flags set
		var projectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"No Flags Project").Scan(&projectID)
		assert.NoError(t, err)

		cli.CreateTestColumn(t, db, projectID, "Neutral")

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		cleanOutput := testutil.StripANSI(output)
		assert.Contains(t, cleanOutput, "Neutral")
	})

	t.Run("List columns quiet mode with single column", func(t *testing.T) {
		var projectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Single Column Project").Scan(&projectID)
		assert.NoError(t, err)

		columnID := cli.CreateTestColumn(t, db, projectID, "OnlyColumn")

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--quiet",
		})

		assert.NoError(t, err)

		// Output should be just the column ID
		trimmedOutput := strings.TrimSpace(output)
		assert.Equal(t, fmt.Sprintf("%d", columnID), trimmedOutput)
	})

	t.Run("List columns JSON output with empty project", func(t *testing.T) {
		var projectID int
		err := db.QueryRowContext(ctx,
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Empty JSON Project").Scan(&projectID)
		assert.NoError(t, err)

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		// Should still return valid JSON with empty columns array
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Note: empty project returns "No columns found..." in human-readable
		// but still processes successfully
		columns := result["columns"].([]any)
		assert.Len(t, columns, 0)
	})
}
