package column_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/column"
	"github.com/thenoetrevino/paso/internal/testing/cli"
)

// TestCreateColumn_MissingFlags tests error cases for the create command with missing required flags
func TestCreateColumn_MissingFlags(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Test missing --name flag
	t.Run("missing --name flag", func(t *testing.T) {
		cmd := column.CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--quiet",
		})
		assert.Error(t, err)
	})

	// Test missing --project flag
	t.Run("missing --project flag", func(t *testing.T) {
		cmd := column.CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Test Column",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitUsage)
		assert.Contains(t, err.Error(), "no project specified")
	})

	// Test invalid project ID
	t.Run("invalid project ID", func(t *testing.T) {
		cmd := column.CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Test Column",
			"--project", "99999",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "project 99999 not found")
	})

	// Test invalid --after column ID
	t.Run("invalid --after column ID", func(t *testing.T) {
		cmd := column.CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Test Column",
			"--project", fmt.Sprintf("%d", projectID),
			"--after", "99999",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "column 99999 not found")
	})

	// Test after column from different project
	t.Run("after column from different project", func(t *testing.T) {
		otherProjectID := cli.CreateTestProject(t, db, "Other Project")
		otherColumnID := cli.CreateTestColumn(t, db, otherProjectID, "Other Column")

		cmd := column.CreateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Test Column",
			"--project", fmt.Sprintf("%d", projectID),
			"--after", strconv.Itoa(otherColumnID),
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitValidation)
		assert.Contains(t, err.Error(), "does not belong to project")
	})
}

// TestUpdateColumn_MissingFlags tests error cases for the update command
func TestUpdateColumn_MissingFlags(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	// Test missing ID argument
	t.Run("missing ID argument", func(t *testing.T) {
		cmd := column.UpdateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Updated Name",
			"--quiet",
		})
		assert.Error(t, err)
	})

	// Test non-existent column ID
	t.Run("non-existent column ID", func(t *testing.T) {
		cmd := column.UpdateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"99999",
			"--name", "Updated Name",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "column 99999 not found")
	})

	// Test no update flags provided
	t.Run("no update flags provided", func(t *testing.T) {
		cmd := column.UpdateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"1",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitUsage)
		assert.Contains(t, err.Error(), "at least one of --name, --ready, --completed, or --in-progress must be provided")
	})

	// Test invalid ID value format
	t.Run("invalid ID value format", func(t *testing.T) {
		cmd := column.UpdateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid",
			"--name", "Updated Name",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitValidation)
		assert.Contains(t, err.Error(), "invalid ID 'invalid': must be a number")
	})

	// Test zero column ID
	t.Run("zero column ID", func(t *testing.T) {
		cmd := column.UpdateCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"0",
			"--name", "Updated Name",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "column 0 not found")
	})
}

// TestListColumn_MissingFlags tests error cases for the list command
func TestListColumn_MissingFlags(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	// Test missing --project flag
	t.Run("missing --project flag", func(t *testing.T) {
		cmd := column.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitUsage)
		assert.Contains(t, err.Error(), "no project specified")
	})

	// Test invalid project ID
	t.Run("invalid project ID", func(t *testing.T) {
		cmd := column.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "99999",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "project 99999 not found")
	})

	// Test invalid --project value format
	t.Run("invalid --project value format", func(t *testing.T) {
		cmd := column.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "invalid",
			"--quiet",
		})
		assert.Error(t, err)
	})

	// Test negative project ID
	t.Run("negative project ID", func(t *testing.T) {
		// Cobra may interpret "-1" as a flag, so just assert error
		cmd := column.ListCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "-1",
			"--quiet",
		})
		assert.Error(t, err)
	})
}

// TestDeleteColumn_MissingFlags tests error cases for the delete command
func TestDeleteColumn_MissingFlags(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	// Test missing ID argument
	t.Run("missing ID argument", func(t *testing.T) {
		cmd := column.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--force",
			"--quiet",
		})
		assert.Error(t, err)
	})

	// Test non-existent column ID
	t.Run("non-existent column ID", func(t *testing.T) {
		cmd := column.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"99999",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "column 99999 not found")
	})

	// Test invalid ID value format
	t.Run("invalid ID value format", func(t *testing.T) {
		cmd := column.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"invalid",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitValidation)
		assert.Contains(t, err.Error(), "invalid ID 'invalid': must be a number")
	})

	// Test zero column ID
	t.Run("zero column ID", func(t *testing.T) {
		cmd := column.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"0",
			"--quiet",
		})
		cli.AssertExitError(t, err, rootcli.ExitNotFound)
		assert.Contains(t, err.Error(), "column 0 not found")
	})

	// Test negative column ID
	t.Run("negative column ID", func(t *testing.T) {
		// Cobra may interpret "-1" as a flag, so just assert error
		cmd := column.DeleteCmd()
		_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"-1",
			"--quiet",
		})
		assert.Error(t, err)
	})
}

// TestCreateColumn_DuplicateName tests that duplicate regular column names are allowed
// (only special columns like ready, completed, in-progress have uniqueness constraints)
func TestCreateColumn_DuplicateName(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Create the first column
	cmd := column.CreateCmd()
	firstColumnName := "Regular Column"
	output1, err := cli.ExecuteCLICommand(t, app, cmd, []string{
		"--name", firstColumnName,
		"--project", fmt.Sprintf("%d", projectID),
		"--quiet",
	})
	require.NoError(t, err, "Failed to create first column")

	// Attempt to create a second column with the same name (should succeed)
	secondCmd := column.CreateCmd()
	output2, err := cli.ExecuteCLICommand(t, app, secondCmd, []string{
		"--name", firstColumnName,
		"--project", fmt.Sprintf("%d", projectID),
		"--quiet",
	})

	assert.NoError(t, err, "Creating duplicate column name should be allowed")
	// Parse IDs from quiet output
	id1 := strings.TrimSpace(output1)
	id2 := strings.TrimSpace(output2)
	assert.NotEqual(t, id1, id2, "Duplicate columns should have different IDs")
}

// TestUpdateColumn_InvalidTransition tests invalid flag transitions
func TestUpdateColumn_InvalidTransition(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	// Create test project and column
	projectID := cli.CreateTestProject(t, db, "Test Project")
	columnID := cli.CreateTestColumn(t, db, projectID, "Test Column")

	// Set the column as completed
	cmd := column.UpdateCmd()
	_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
		strconv.Itoa(columnID),
		"--completed",
		"--quiet",
	})
	require.NoError(t, err, "Failed to set column as completed")

	// Create another column and attempt to set it as completed (should fail without --force)
	otherColumnID := cli.CreateTestColumn(t, db, projectID, "Other Column")
	secondCmd := column.UpdateCmd()
	_, err = cli.ExecuteCLICommand(t, app, secondCmd, []string{
		strconv.Itoa(otherColumnID),
		"--completed",
		"--quiet",
	})

	// Should get an error about completed column already existing
	assert.Error(t, err, "Expected error when setting second completed column without --force")
	assert.Contains(t, err.Error(), "completed column already exists")
}

// TestCreateColumn_ProjectValidation tests project validation during creation
func TestCreateColumn_ProjectValidation(t *testing.T) {
	t.Parallel()
	_, app := cli.SetupCLITest(t)

	tests := []struct {
		name      string
		projectID int
		expectErr bool
	}{
		{
			name:      "non-existent project",
			projectID: 99999,
			expectErr: true,
		},
		{
			name:      "zero project ID",
			projectID: 0,
			expectErr: true,
		},
		{
			name:      "negative project ID",
			projectID: -1,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := column.CreateCmd()
			_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				"--name", "Test Column",
				"--project", fmt.Sprintf("%d", tt.projectID),
				"--quiet",
			})

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
