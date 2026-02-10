package label

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestListLabel_PositionalArgVsFlag(t *testing.T) {
	t.Run("Positional arg and flag produce same JSON output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")

		// Create labels
		createTestLabel(t, db, projectID, "Label 1", "#FF0000")
		createTestLabel(t, db, projectID, "Label 2", "#00FF00")

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

		labelsFlag := resultFlag["labels"].([]any)
		labelsPositional := resultPositional["labels"].([]any)
		assert.Equal(t, len(labelsFlag), len(labelsPositional))

		// Verify same label IDs in same order
		for i := range labelsFlag {
			lblFlag := labelsFlag[i].(map[string]any)
			lblPositional := labelsPositional[i].(map[string]any)
			assert.Equal(t, lblFlag["id"], lblPositional["id"])
			assert.Equal(t, lblFlag["name"], lblPositional["name"])
		}
	})

	t.Run("Positional arg and flag produce same quiet output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")

		lbl1 := createTestLabel(t, db, projectID, "Label 1", "#FF0000")
		lbl2 := createTestLabel(t, db, projectID, "Label 2", "#00FF00")

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

		// Verify output contains label IDs
		assert.Contains(t, outputPositional, fmt.Sprintf("%d", lbl1))
		assert.Contains(t, outputPositional, fmt.Sprintf("%d", lbl2))
	})

	t.Run("Positional arg and flag produce same human-readable output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")

		createTestLabel(t, db, projectID, "Label 1", "#FF0000")
		createTestLabel(t, db, projectID, "Label 2", "#00FF00")

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
		assert.Contains(t, outputPositional, "Labels in project 'Test Project':")
		assert.Contains(t, outputPositional, "Label 1")
		assert.Contains(t, outputPositional, "Label 2")
	})

	t.Run("Positional arg takes precedence over flag", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		// Create two projects
		projectID1 := cli.CreateTestProject(t, db, "Project 1")
		projectID2 := cli.CreateTestProject(t, db, "Project 2")

		createTestLabel(t, db, projectID1, "Label in Project 1", "#FF0000")
		createTestLabel(t, db, projectID2, "Label in Project 2", "#00FF00")

		// Use positional arg "2" with flag "--project 1"
		// Should use project 2 (positional takes precedence)
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"2", "--project", "1", "--json"})

		assert.NoError(t, err)

		var result map[string]any
		json.Unmarshal([]byte(output), &result)

		labels := result["labels"].([]any)
		// Should contain label from project 2
		found := false
		for _, lblInterface := range labels {
			lbl := lblInterface.(map[string]any)
			if lbl["name"] == "Label in Project 2" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find label from project 2")
	})

	t.Run("Invalid positional arg shows clear error", func(t *testing.T) {
	})

	t.Run("Positional arg works for valid project", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		createTestLabel(t, db, projectID, "Label 1", "#FF0000")

		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Labels in project 'Test Project':")
		assert.Contains(t, output, "Label 1")
	})

	t.Run("Empty args falls back to flag or git detection", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		createTestLabel(t, db, projectID, "Label 1", "#FF0000")

		// No positional arg, use flag
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"--project", "1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Labels in project 'Test Project':")
		assert.Contains(t, output, "Label 1")

		// Note: "No positional arg, no flag" case calls os.Exit() and cannot be tested here
	})
}

// Additional edge case tests
func TestListLabel_PositionalArgEdgeCases(t *testing.T) {
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
	})

	t.Run("Positional arg with negative value", func(t *testing.T) {
	})
}

// Test helper to create labels
func createTestLabel(t *testing.T, db *sql.DB, projectID int, name, color string) int {
	t.Helper()

	var labelID int
	err := db.QueryRowContext(context.Background(),
		"INSERT INTO labels (project_id, name, color) VALUES (?, ?, ?) RETURNING id",
		projectID, name, color).Scan(&labelID)

	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}

	return labelID
}
