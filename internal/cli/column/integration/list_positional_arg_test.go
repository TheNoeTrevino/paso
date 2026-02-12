package column_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/column"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestListColumn_PositionalArgVsFlag(t *testing.T) {
	t.Parallel()
	t.Run("positional arg and flag produce same JSON output", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestColumn(t, db, projectID, "Column 1")
		cli.CreateTestColumn(t, db, projectID, "Column 2")

		// Execute with positional arg
		cmdPositional := column.ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1", "--json"})

		// Execute with flag
		cmdFlag := column.ListCmd()
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

	t.Run("positional arg and flag produce same quiet output", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col1 := cli.CreateTestColumn(t, db, projectID, "Column 1")
		col2 := cli.CreateTestColumn(t, db, projectID, "Column 2")

		cmdPositional := column.ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1", "--quiet"})

		cmdFlag := column.ListCmd()
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

	t.Run("positional arg and flag produce same human-readable output", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestColumn(t, db, projectID, "Column 1")
		cli.CreateTestColumn(t, db, projectID, "Column 2")

		cmdPositional := column.ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1"})

		cmdFlag := column.ListCmd()
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

	t.Run("positional arg takes precedence over flag", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
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
		cmd := column.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"2", "--project", "1", "--json"})

		assert.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &result))

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

	t.Run("invalid positional arg shows clear error", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		_, app := cli.SetupCLITest(t)

		cmd := column.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"not-a-number",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitUsage)
		assert.Contains(t, err.Error(), "Invalid project ID")
	})

	t.Run("positional arg works for valid project", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestColumn(t, db, projectID, "Column 1")

		cmd := column.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Columns in project 'Test Project':")
		assert.Contains(t, output, "Column 1")
	})

	t.Run("empty args falls back to flag or git detection", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestColumn(t, db, projectID, "Column 1")

		// No positional arg, use flag
		cmd := column.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"--project", "1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Columns in project 'Test Project':")
		assert.Contains(t, output, "Column 1")
	})
}

// Additional edge case tests
func TestListColumn_PositionalArgEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("multiple positional args not allowed", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := column.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1", "2"})

		// Should error due to Args: cobra.MaximumNArgs(1)
		assert.Error(t, err)
	})

	t.Run("positional arg with zero value", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := column.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"0",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "project 0 not found")
	})

	t.Run("positional arg with negative value", func(t *testing.T) {
		t.Parallel()
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		// Cobra may interpret "-1" as a flag, so just assert error
		cmd := column.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"-1",
			"--quiet",
		})
		assert.Error(t, err)
	})
}
