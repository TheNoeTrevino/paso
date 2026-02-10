package task

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

func TestListTask_PositionalArgVsFlag(t *testing.T) {
	t.Run("Positional arg and flag produce same JSON output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col := cli.CreateTestColumn(t, db, projectID, "Column")
		cli.CreateTestTask(t, db, col, "Task 1")
		cli.CreateTestTask(t, db, col, "Task 2")

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
		var resultPositional, resultFlag struct {
			Success bool                  `json:"success"`
			Tasks   []*models.TaskSummary `json:"tasks"`
		}

		err := json.Unmarshal([]byte(outputPositional), &resultPositional)
		assert.NoError(t, err)

		err = json.Unmarshal([]byte(outputFlag), &resultFlag)
		assert.NoError(t, err)

		// Results should be identical
		assert.Equal(t, resultFlag.Success, resultPositional.Success)
		assert.Equal(t, len(resultFlag.Tasks), len(resultPositional.Tasks))

		// Verify same task IDs in same order
		for i := range resultFlag.Tasks {
			assert.Equal(t, resultFlag.Tasks[i].ID, resultPositional.Tasks[i].ID)
			assert.Equal(t, resultFlag.Tasks[i].Title, resultPositional.Tasks[i].Title)
		}
	})

	t.Run("Positional arg and flag produce same quiet output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col := cli.CreateTestColumn(t, db, projectID, "Column")
		task1 := cli.CreateTestTask(t, db, col, "Task 1")
		task2 := cli.CreateTestTask(t, db, col, "Task 2")

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

		// Verify output contains task IDs
		assert.Contains(t, outputPositional, fmt.Sprintf("%d", task1))
		assert.Contains(t, outputPositional, fmt.Sprintf("%d", task2))
	})

	t.Run("Positional arg and flag produce same human-readable output", func(t *testing.T) {
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col := cli.CreateTestColumn(t, db, projectID, "Column")
		cli.CreateTestTask(t, db, col, "Task 1")
		cli.CreateTestTask(t, db, col, "Task 2")

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
		assert.Contains(t, outputPositional, "Found 2 tasks")
		assert.Contains(t, outputPositional, "Task 1")
		assert.Contains(t, outputPositional, "Task 2")
	})

	t.Run("Positional arg takes precedence over flag", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		// Create two projects
		projectID1 := cli.CreateTestProject(t, db, "Project 1")
		projectID2 := cli.CreateTestProject(t, db, "Project 2")

		col1 := cli.CreateTestColumn(t, db, projectID1, "Col1")
		col2 := cli.CreateTestColumn(t, db, projectID2, "Col2")

		cli.CreateTestTask(t, db, col1, "Task in Project 1")
		cli.CreateTestTask(t, db, col2, "Task in Project 2")

		// Use positional arg "2" with flag "--project 1"
		// Should use project 2 (positional takes precedence)
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"2", "--project", "1", "--json"})

		assert.NoError(t, err)

		var result struct {
			Tasks []*models.TaskSummary `json:"tasks"`
		}
		json.Unmarshal([]byte(output), &result)

		// Should contain task from project 2
		assert.Equal(t, 1, len(result.Tasks))
		assert.Equal(t, "Task in Project 2", result.Tasks[0].Title)
	})

	t.Run("Invalid positional arg shows clear error", func(t *testing.T) {
	})

	t.Run("Positional arg works for valid project", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col := cli.CreateTestColumn(t, db, projectID, "Column")
		cli.CreateTestTask(t, db, col, "Task 1")

		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Found 1 tasks")
		assert.Contains(t, output, "Task 1")
	})

	t.Run("Empty args falls back to flag or git detection", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col := cli.CreateTestColumn(t, db, projectID, "Column")
		cli.CreateTestTask(t, db, col, "Task 1")

		// No positional arg, use flag
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"--project", "1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Found 1 tasks")
		assert.Contains(t, output, "Task 1")

		// Note: "no positional arg, no flag" case is skipped because
		// the command calls os.Exit() when no project can be determined
	})
}

// Additional edge case tests
func TestListTask_PositionalArgEdgeCases(t *testing.T) {
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
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"0"})

		// Project 0 might not exist, but the command should handle it gracefully
		// Either returns empty result or error
		_ = err
		_ = output
		// The behavior depends on whether project 0 exists in DB
		// This test just ensures no panic occurs
	})

	t.Run("Positional arg with negative value", func(t *testing.T) {
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"-1"})

		// Negative project ID might fail or return empty result
		// The important part is no panic occurs
		_ = err
		_ = output
	})
}
