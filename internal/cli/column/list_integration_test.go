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

func TestListColumns_Positive(t *testing.T) {
	// Setup test DB and App
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Set flags on default columns for testing
	_, err := db.ExecContext(context.Background(), `
		UPDATE columns
		SET holds_ready_tasks = 1
		WHERE project_id = ? AND name = 'Todo'`, projectID)
	if err != nil {
		t.Fatalf("Failed to set holds_ready_tasks flag: %v", err)
	}

	_, err = db.ExecContext(context.Background(), `
		UPDATE columns
		SET holds_in_progress_tasks = 1
		WHERE project_id = ? AND name = 'In Progress'`, projectID)
	if err != nil {
		t.Fatalf("Failed to set holds_in_progress_tasks flag: %v", err)
	}

	_, err = db.ExecContext(context.Background(), `
		UPDATE columns
		SET holds_completed_tasks = 1
		WHERE project_id = ? AND name = 'Done'`, projectID)
	if err != nil {
		t.Fatalf("Failed to set holds_completed_tasks flag: %v", err)
	}

	t.Run("List columns with project flag", func(t *testing.T) {
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Columns in project 'Test Project':")
		// Default columns created with project: Todo, In Progress, Done
		assert.Contains(t, output, "Todo")
		assert.Contains(t, output, "In Progress")
		assert.Contains(t, output, "Done")
	})

	t.Run("List columns with JSON output", func(t *testing.T) {
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--json",
		})

		assert.NoError(t, err)

		// Parse JSON output
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify JSON structure
		assert.True(t, result["success"].(bool))
		columns := result["columns"].([]interface{})
		assert.Len(t, columns, 3)

		// Verify first column (Todo)
		firstCol := columns[0].(map[string]interface{})
		assert.Equal(t, "Todo", firstCol["name"])
		assert.True(t, firstCol["holds_ready_tasks"].(bool))

		// Verify second column (In Progress)
		secondCol := columns[1].(map[string]interface{})
		assert.Equal(t, "In Progress", secondCol["name"])
		assert.True(t, secondCol["holds_in_progress_tasks"].(bool))

		// Verify third column (Done)
		thirdCol := columns[2].(map[string]interface{})
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
		assert.Equal(t, 3, len(lines))

		// Each line should be a numeric ID
		for _, line := range lines {
			assert.Regexp(t, `^\d+$`, line)
		}
	})

	t.Run("List columns when none exist (empty project)", func(t *testing.T) {
		// Create a project with no columns (we need to manually create one without default columns)
		// Since CreateTestProject creates default columns, we create a project directly
		var emptyProjectID int
		err := db.QueryRowContext(context.Background(),
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
		err := db.QueryRowContext(context.Background(),
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

		// Verify columns appear in order with position numbers
		firstPos := strings.Index(output, "1. First")
		secondPos := strings.Index(output, "2. Second")
		thirdPos := strings.Index(output, "3. Third")

		assert.Greater(t, firstPos, -1)
		assert.Greater(t, secondPos, -1)
		assert.Greater(t, thirdPos, -1)
		assert.Less(t, firstPos, secondPos)
		assert.Less(t, secondPos, thirdPos)

		// Verify column IDs appear in output
		assert.Contains(t, output, fmt.Sprintf("ID: %d", column1ID))
		assert.Contains(t, output, fmt.Sprintf("ID: %d", column2ID))
		assert.Contains(t, output, fmt.Sprintf("ID: %d", column3ID))
	})

	t.Run("List columns with column flags set", func(t *testing.T) {
		// Create a project and modify column flags
		var flagProjectID int
		err := db.QueryRowContext(context.Background(),
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"Flag Project").Scan(&flagProjectID)
		assert.NoError(t, err)

		// Create a column and set flags
		flagColumnID := cli.CreateTestColumn(t, db, flagProjectID, "Custom")
		_, err = db.ExecContext(context.Background(), `
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
		assert.Contains(t, output, "Custom")
		assert.Contains(t, output, "[READY]")
		assert.Contains(t, output, "[IN-PROGRESS]")
	})

	t.Run("List columns in JSON mode with complete structure", func(t *testing.T) {
		// Create a project with multiple columns
		var jsonProjectID int
		err := db.QueryRowContext(context.Background(),
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"JSON Project").Scan(&jsonProjectID)
		assert.NoError(t, err)

		// Create columns
		column1ID := cli.CreateTestColumn(t, db, jsonProjectID, "Backlog")
		column2ID := cli.CreateTestColumn(t, db, jsonProjectID, "Active")

		// Set flags on columns
		_, err = db.ExecContext(context.Background(), `
			UPDATE columns
			SET holds_ready_tasks = 1
			WHERE id = ?`,
			column1ID)
		assert.NoError(t, err)

		_, err = db.ExecContext(context.Background(), `
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
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Verify structure
		assert.True(t, result["success"].(bool))
		columns := result["columns"].([]interface{})
		assert.Len(t, columns, 2)

		// Verify first column structure
		col1 := columns[0].(map[string]interface{})
		assert.Equal(t, float64(column1ID), col1["id"])
		assert.Equal(t, "Backlog", col1["name"])
		assert.Equal(t, float64(jsonProjectID), col1["project_id"])
		assert.True(t, col1["holds_ready_tasks"].(bool))
		assert.False(t, col1["holds_completed_tasks"].(bool))

		// Verify second column structure
		col2 := columns[1].(map[string]interface{})
		assert.Equal(t, float64(column2ID), col2["id"])
		assert.Equal(t, "Active", col2["name"])
		assert.Equal(t, float64(jsonProjectID), col2["project_id"])
		assert.True(t, col2["holds_in_progress_tasks"].(bool))
	})

	t.Run("List columns with multiple projects", func(t *testing.T) {
		// Create another project
		var projectID2 int
		err := db.QueryRowContext(context.Background(),
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
		assert.Contains(t, output, "Columns in project 'Second Project':")
		assert.Contains(t, output, "Custom1")
		assert.Contains(t, output, "Custom2")

		// Verify first project's columns are NOT in the output
		assert.NotContains(t, output, "Todo")
		assert.NotContains(t, output, "In Progress")
		assert.NotContains(t, output, "Done")
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
	db, app := cli.SetupCLITest(t)
	defer func() {
		_ = db.Close()
	}()

	t.Run("Invalid project ID error handling", func(t *testing.T) {
		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "99999",
		})

		// Should have an error since project doesn't exist
		assert.Error(t, err)
		assert.Contains(t, output, "project 99999 not found")
	})

	t.Run("Missing project flag", func(t *testing.T) {
		cmd := ListCmd()

		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{})

		// Should have an error since --project is required
		assert.Error(t, err)
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
	defer func() {
		_ = db.Close()
	}()

	t.Run("List columns with special characters in names", func(t *testing.T) {
		var projectID int
		err := db.QueryRowContext(context.Background(),
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
		err := db.QueryRowContext(context.Background(),
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
		err := db.QueryRowContext(context.Background(),
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

		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		columns := result["columns"].([]interface{})
		assert.Len(t, columns, 5)

		// Verify order matches insertion order
		for i, colInterface := range columns {
			col := colInterface.(map[string]interface{})
			assert.Equal(t, names[i], col["name"])
		}
	})

	t.Run("List columns with no flags formatting", func(t *testing.T) {
		// Create a project with columns that have no flags set
		var projectID int
		err := db.QueryRowContext(context.Background(),
			"INSERT INTO projects (name) VALUES (?) RETURNING id",
			"No Flags Project").Scan(&projectID)
		assert.NoError(t, err)

		cli.CreateTestColumn(t, db, projectID, "Neutral")

		cmd := ListCmd()

		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Neutral")
		// Should not contain flag indicators
		assert.NotContains(t, output, "[READY]")
		assert.NotContains(t, output, "[IN-PROGRESS]")
		assert.NotContains(t, output, "[COMPLETED]")
	})

	t.Run("List columns quiet mode with single column", func(t *testing.T) {
		var projectID int
		err := db.QueryRowContext(context.Background(),
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
		err := db.QueryRowContext(context.Background(),
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
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)

		// Note: empty project returns "No columns found..." in human-readable
		// but still processes successfully
		columns := result["columns"].([]interface{})
		assert.Len(t, columns, 0)
	})
}
