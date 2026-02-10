package column

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestListColumn_PositionalArgVsFlag(t *testing.T) {
	t.Run("Positional arg and flag produce same JSON output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestColumn(t, db, projectID, "Column 1")
		cli.CreateTestColumn(t, db, projectID, "Column 2")

		// Execute with positional arg
		cmdPositional := ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1", "--json"})

		// Execute with flag
		cmdFlag := ListCmd()
		outputFlag, errFlag := cli.ExecuteCLICommand(t, app, cmdFlag,
			[]string{"--project", "1", "--json"})

		// Both should succeed
		assert.NoError(t, errPositional)
		assert.NoError(t, errFlag)

		// Parse JSON from both
		var resultPositional, resultFlag map[string]any

		err := json.Unmarshal([]byte(outputPositional), &resultPositional)
		assert.NoError(t, err)

		err = json.Unmarshal([]byte(outputFlag), &resultFlag)
		assert.NoError(t, err)

		// Results should be identical
		assert.Equal(t, resultFlag["success"], resultPositional["success"])

		columnsFlag := resultFlag["columns"].([]any)
		columnsPositional := resultPositional["columns"].([]any)
		assert.Equal(t, len(columnsFlag), len(columnsPositional))

		// Verify same column IDs in same order
		for i := range columnsFlag {
			colFlag := columnsFlag[i].(map[string]any)
			colPositional := columnsPositional[i].(map[string]any)
			assert.Equal(t, colFlag["id"], colPositional["id"])
			assert.Equal(t, colFlag["name"], colPositional["name"])
		}
	})

	t.Run("Positional arg and flag produce same quiet output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col1 := cli.CreateTestColumn(t, db, projectID, "Column 1")
		col2 := cli.CreateTestColumn(t, db, projectID, "Column 2")

		cmdPositional := ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1", "--quiet"})

		cmdFlag := ListCmd()
		outputFlag, errFlag := cli.ExecuteCLICommand(t, app, cmdFlag,
			[]string{"--project", "1", "--quiet"})

		assert.NoError(t, errPositional)
		assert.NoError(t, errFlag)

		// Outputs should be identical
		assert.Equal(t, outputFlag, outputPositional)

		// Verify output contains column IDs
		assert.Contains(t, outputPositional, fmt.Sprintf("%d", col1))
		assert.Contains(t, outputPositional, fmt.Sprintf("%d", col2))
	})

	t.Run("Positional arg and flag produce same human-readable output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestColumn(t, db, projectID, "Column 1")
		cli.CreateTestColumn(t, db, projectID, "Column 2")

		cmdPositional := ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1"})

		cmdFlag := ListCmd()
		outputFlag, errFlag := cli.ExecuteCLICommand(t, app, cmdFlag,
			[]string{"--project", "1"})

		assert.NoError(t, errPositional)
		assert.NoError(t, errFlag)

		// Outputs should be identical
		assert.Equal(t, outputFlag, outputPositional)

		// Verify standard output format
		assert.Contains(t, outputPositional, "Columns in project 'Test Project':")
		assert.Contains(t, outputPositional, "Column 1")
		assert.Contains(t, outputPositional, "Column 2")
	})

	t.Run("Positional arg takes precedence over flag", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		// Create two projects
		projectID1 := cli.CreateTestProject(t, db, "Project 1")
		projectID2 := cli.CreateTestProject(t, db, "Project 2")

		cli.CreateTestColumn(t, db, projectID1, "Column in Project 1")
		cli.CreateTestColumn(t, db, projectID2, "Column in Project 2")

		// Use positional arg "2" with flag "--project 1"
		// Should use project 2 (positional takes precedence)
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"2", "--project", "1", "--json"})

		assert.NoError(t, err)

		var result map[string]any
		json.Unmarshal([]byte(output), &result)

		columns := result["columns"].([]any)
		// Should contain column from project 2
		found := false
		for _, colInterface := range columns {
			col := colInterface.(map[string]any)
			if col["name"] == "Column in Project 2" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find column from project 2")
	})

	t.Run("Invalid positional arg shows clear error", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on invalid positional arg")
	})

	t.Run("Positional arg works for valid project", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestColumn(t, db, projectID, "Column 1")

		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Columns in project 'Test Project':")
		assert.Contains(t, output, "Column 1")
	})

	t.Run("Empty args falls back to flag or git detection", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestColumn(t, db, projectID, "Column 1")

		// No positional arg, use flag
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"--project", "1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Columns in project 'Test Project':")
		assert.Contains(t, output, "Column 1")

		// Note: "No positional arg, no flag" case calls os.Exit() and cannot be tested here
	})
}

// Additional edge case tests
func TestListColumn_PositionalArgEdgeCases(t *testing.T) {
	t.Run("Multiple positional args not allowed", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1", "2"})

		// Should error due to Args: cobra.MaximumNArgs(1)
		assert.Error(t, err)
	})

	t.Run("Positional arg with zero value", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on project not found")
	})

	t.Run("Positional arg with negative value", func(t *testing.T) {
		t.Skip("Skipping: command calls os.Exit() on project not found")
	})
}
