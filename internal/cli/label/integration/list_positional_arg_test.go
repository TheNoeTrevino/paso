package label_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestListLabel_PositionalArgVsFlag(t *testing.T) {
	t.Parallel()
	t.Run("positional arg and flag produce same JSON output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")

		// Create labels
		cli.CreateTestLabel(t, db, projectID, "Label 1", "#FF0000")
		cli.CreateTestLabel(t, db, projectID, "Label 2", "#00FF00")

		// Execute with positional arg
		cmdPositional := label.ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1", "--json"})

		// Execute with flag
		cmdFlag := label.ListCmd()
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

	t.Run("positional arg and flag produce same quiet output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")

		lbl1 := cli.CreateTestLabel(t, db, projectID, "Label 1", "#FF0000")
		lbl2 := cli.CreateTestLabel(t, db, projectID, "Label 2", "#00FF00")

		cmdPositional := label.ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1", "--quiet"})

		cmdFlag := label.ListCmd()
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

	t.Run("positional arg and flag produce same human-readable output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")

		cli.CreateTestLabel(t, db, projectID, "Label 1", "#FF0000")
		cli.CreateTestLabel(t, db, projectID, "Label 2", "#00FF00")

		cmdPositional := label.ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1"})

		cmdFlag := label.ListCmd()
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

	t.Run("positional arg takes precedence over flag", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		// Create two projects
		projectID1 := cli.CreateTestProject(t, db, "Project 1")
		projectID2 := cli.CreateTestProject(t, db, "Project 2")

		cli.CreateTestLabel(t, db, projectID1, "Label in Project 1", "#FF0000")
		cli.CreateTestLabel(t, db, projectID2, "Label in Project 2", "#00FF00")

		// Use positional arg "2" with flag "--project 1"
		// Should use project 2 (positional takes precedence)
		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"2", "--project", "1", "--json"})

		assert.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &result))

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

	t.Run("invalid positional arg shows clear error", func(t *testing.T) {
		_, app := cli.SetupCLITest(t)

		cmd := label.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitUsage)
		assert.Contains(t, err.Error(), "Invalid project ID")
	})

	t.Run("positional arg works for valid project", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestLabel(t, db, projectID, "Label 1", "#FF0000")

		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Labels in project 'Test Project':")
		assert.Contains(t, output, "Label 1")
	})

	t.Run("empty args falls back to flag or git detection", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		cli.CreateTestLabel(t, db, projectID, "Label 1", "#FF0000")

		// No positional arg, use flag
		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"--project", "1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Labels in project 'Test Project':")
		assert.Contains(t, output, "Label 1")
	})
}

// Additional edge case tests
func TestListLabel_PositionalArgEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("multiple positional args not allowed", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := label.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1", "2"})

		// Should error due to Args: cobra.MaximumNArgs(1)
		assert.Error(t, err)
	})

	t.Run("positional arg with zero value", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := label.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"0",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "project 0 not found")
	})

	t.Run("positional arg with negative value", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := label.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"-1",
			"--quiet",
		})
		assert.Error(t, err)
	})
}
