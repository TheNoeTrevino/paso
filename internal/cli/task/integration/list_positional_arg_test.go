package task_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

func TestListTask_PositionalArgVsFlag(t *testing.T) {
	t.Parallel()
	t.Run("positional arg and flag produce same JSON output", func(t *testing.T) {
		t.Parallel()
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
		cmdPositional := task.ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1", "--json"})

		// Execute with flag
		cmdFlag := task.ListCmd()
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

	t.Run("positional arg and flag produce same quiet output", func(t *testing.T) {
		t.Parallel()
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col := cli.CreateTestColumn(t, db, projectID, "Column")
		task1 := cli.CreateTestTask(t, db, col, "Task 1")
		task2 := cli.CreateTestTask(t, db, col, "Task 2")

		cmdPositional := task.ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1", "--quiet"})

		cmdFlag := task.ListCmd()
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

	t.Run("positional arg and flag produce same human-readable output", func(t *testing.T) {
		t.Parallel()
		// Setup
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col := cli.CreateTestColumn(t, db, projectID, "Column")
		cli.CreateTestTask(t, db, col, "Task 1")
		cli.CreateTestTask(t, db, col, "Task 2")

		cmdPositional := task.ListCmd()
		outputPositional, errPositional := cli.ExecuteCLICommand(t, app, cmdPositional,
			[]string{"1"})

		cmdFlag := task.ListCmd()
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

	t.Run("positional arg takes precedence over flag", func(t *testing.T) {
		t.Parallel()
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
		cmd := task.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"2", "--project", "1", "--json"})

		assert.NoError(t, err)

		var result struct {
			Tasks []*models.TaskSummary `json:"tasks"`
		}
		require.NoError(t, json.Unmarshal([]byte(output), &result))

		// Should contain task from project 2
		assert.Equal(t, 1, len(result.Tasks))
		assert.Equal(t, "Task in Project 2", result.Tasks[0].Title)
	})

	t.Run("invalid positional arg shows clear error", func(t *testing.T) {
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := task.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"not-a-number"})
		cli.AssertExitError(t, err, 2) // ExitUsage
		assert.Contains(t, err.Error(), "Invalid project ID: not-a-number")
	})

	t.Run("positional arg works for valid project", func(t *testing.T) {
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col := cli.CreateTestColumn(t, db, projectID, "Column")
		cli.CreateTestTask(t, db, col, "Task 1")

		cmd := task.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Found 1 tasks")
		assert.Contains(t, output, "Task 1")
	})

	t.Run("empty args falls back to flag or git detection", func(t *testing.T) {
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		projectID := cli.CreateTestProject(t, db, "Test Project")
		col := cli.CreateTestColumn(t, db, projectID, "Column")
		cli.CreateTestTask(t, db, col, "Task 1")

		// No positional arg, use flag
		cmd := task.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd,
			[]string{"--project", "1"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Found 1 tasks")
		assert.Contains(t, output, "Task 1")
	})
}

// Additional edge case tests
func TestListTask_PositionalArgEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("multiple positional args not allowed", func(t *testing.T) {
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := task.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"1", "2"})

		// Should error due to Args: cobra.MaximumNArgs(1)
		assert.Error(t, err)
	})

	t.Run("positional arg with zero value", func(t *testing.T) {
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := task.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"0"})

		// Project 0 doesn't exist but the query returns empty results gracefully
		assert.NoError(t, err)
		assert.Contains(t, output, "No tasks found")
	})

	t.Run("positional arg with negative value", func(t *testing.T) {
		t.Parallel()
		db, app := cli.SetupCLITest(t)
		defer func() {
			_ = db.Close()
		}()

		cmd := task.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{"-1"})

		// Cobra interprets "-1" as an unknown flag, not a positional arg
		assert.Error(t, err)
	})
}
