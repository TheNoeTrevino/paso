package column

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil/cli"
)

// TestCreateColumn_MissingFlags tests error cases for the create command with missing required flags
func TestCreateColumn_MissingFlags(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Test missing --name flag
	t.Run("missing --name flag", func(t *testing.T) {
		cmd := CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test missing --project flag
	t.Run("missing --project flag", func(t *testing.T) {
		// Clear env var to test missing flag
		originalEnv := os.Getenv("PASO_PROJECT")
		_ = os.Unsetenv("PASO_PROJECT")
		defer func() {
			if originalEnv != "" {
				_ = os.Setenv("PASO_PROJECT", originalEnv)
			}
		}()

		cmd := CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Test Column",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test invalid project ID
	t.Run("invalid project ID", func(t *testing.T) {
		cmd := CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Test Column",
			"--project", "99999",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test invalid --after column ID
	t.Run("invalid --after column ID", func(t *testing.T) {
		cmd := CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Test Column",
			"--project", fmt.Sprintf("%d", projectID),
			"--after", "99999",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test after column from different project
	t.Run("after column from different project", func(t *testing.T) {
		// Create another project
		otherProjectID := cli.CreateTestProject(t, db, "Other Project")
		otherColumnID := cli.CreateTestColumn(t, db, otherProjectID, "Other Column")

		cmd := CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Test Column",
			"--project", fmt.Sprintf("%d", projectID),
			"--after", fmt.Sprintf("%d", otherColumnID),
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})
}

// TestUpdateColumn_MissingFlags tests error cases for the update command
func TestUpdateColumn_MissingFlags(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	// Create test project and column
	projectID := cli.CreateTestProject(t, db, "Test Project")
	columnID := cli.CreateTestColumn(t, db, projectID, "Test Column")

	// Test missing --id flag
	t.Run("missing --id flag", func(t *testing.T) {
		cmd := UpdateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Updated Name",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test non-existent column ID
	t.Run("non-existent column ID", func(t *testing.T) {
		cmd := UpdateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "99999",
			"--name", "Updated Name",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test no update flags provided
	t.Run("no update flags provided", func(t *testing.T) {
		cmd := UpdateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", strconv.Itoa(columnID),
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test invalid --id value format
	t.Run("invalid --id value format", func(t *testing.T) {
		cmd := UpdateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "invalid",
			"--name", "Updated Name",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test zero column ID
	t.Run("zero column ID", func(t *testing.T) {
		cmd := UpdateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "0",
			"--name", "Updated Name",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})
}

// TestListColumn_MissingFlags tests error cases for the list command
func TestListColumn_MissingFlags(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	// Test missing --project flag with no env var
	t.Run("missing --project flag with no env var", func(t *testing.T) {
		// Clear any PASO_PROJECT env var to ensure test isolation
		originalEnv := os.Getenv("PASO_PROJECT")
		_ = os.Unsetenv("PASO_PROJECT")
		defer func() {
			if originalEnv != "" {
				_ = os.Setenv("PASO_PROJECT", originalEnv)
			}
		}()

		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{"--quiet"})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test invalid project ID
	t.Run("invalid project ID", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "99999",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test invalid --project value format
	t.Run("invalid --project value format", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "invalid",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test negative project ID
	t.Run("negative project ID", func(t *testing.T) {
		cmd := ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", "-1",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})
}

// TestDeleteColumn_MissingFlags tests error cases for the delete command
func TestDeleteColumn_MissingFlags(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	// Test missing --id flag
	t.Run("missing --id flag", func(t *testing.T) {
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--force",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test non-existent column ID
	t.Run("non-existent column ID", func(t *testing.T) {
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "99999",
			"--force",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test invalid --id value format
	t.Run("invalid --id value format", func(t *testing.T) {
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "invalid",
			"--force",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test zero column ID
	t.Run("zero column ID", func(t *testing.T) {
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "0",
			"--force",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})

	// Test negative column ID
	t.Run("negative column ID", func(t *testing.T) {
		cmd := DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", "-1",
			"--force",
			"--quiet",
		})
		if err == nil {
			t.Errorf("Expected error but got none; output: %s", output)
		}
	})
}

// TestCreateColumn_DuplicateName tests that duplicate regular column names are allowed
// (only special columns like ready, completed, in-progress have uniqueness constraints)
func TestCreateColumn_DuplicateName(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	// Create test project
	projectID := cli.CreateTestProject(t, db, "Test Project")

	// Create the first column
	cmd := CreateCmd()
	firstColumnName := "Regular Column"
	output1, err := cli.ExecuteCLICommand(t, app, cmd, []string{
		"--name", firstColumnName,
		"--project", fmt.Sprintf("%d", projectID),
		"--quiet",
	})
	if err != nil {
		t.Fatalf("Failed to create first column: %v", err)
	}

	// Attempt to create a second column with the same name (should succeed)
	secondCmd := CreateCmd()
	output2, err := cli.ExecuteCLICommand(t, app, secondCmd, []string{
		"--name", firstColumnName,
		"--project", fmt.Sprintf("%d", projectID),
		"--quiet",
	})

	if err != nil {
		t.Errorf("Creating duplicate column name should be allowed, but got error: %v", err)
	}
	// Parse IDs from quiet output
	id1 := strings.TrimSpace(output1)
	id2 := strings.TrimSpace(output2)
	if id1 == id2 {
		t.Errorf("Duplicate columns should have different IDs, got same ID: %s", id1)
	}
}

// TestUpdateColumn_InvalidTransition tests invalid flag transitions
func TestUpdateColumn_InvalidTransition(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

	// Create test project and column
	projectID := cli.CreateTestProject(t, db, "Test Project")
	columnID := cli.CreateTestColumn(t, db, projectID, "Test Column")

	// Set the column as completed
	cmd := UpdateCmd()
	_, err := cli.ExecuteCLICommand(t, app, cmd, []string{
		"--id", strconv.Itoa(columnID),
		"--completed",
		"--quiet",
	})
	if err != nil {
		t.Fatalf("Failed to set column as completed: %v", err)
	}

	// Create another column and attempt to set it as completed (should fail without --force)
	otherColumnID := cli.CreateTestColumn(t, db, projectID, "Other Column")
	secondCmd := UpdateCmd()
	output, err := cli.ExecuteCLICommand(t, app, secondCmd, []string{
		"--id", strconv.Itoa(otherColumnID),
		"--completed",
		"--quiet",
	})

	// Should get an error about completed column already existing
	if err == nil {
		t.Errorf("Expected error when setting second completed column without --force, but got none; output: %s", output)
	}
}

// TestCreateColumn_ProjectValidation tests project validation during creation
func TestCreateColumn_ProjectValidation(t *testing.T) {
	db, app := cli.SetupCLITest(t)
	defer func() {
		err := db.Close()
		assert.NoError(t, err)
	}()

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
			cmd := CreateCmd()
			output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
				"--name", "Test Column",
				"--project", fmt.Sprintf("%d", tt.projectID),
				"--quiet",
			})

			if tt.expectErr && err == nil {
				t.Errorf("Expected error but got none; output: %s", output)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v; output: %s", err, output)
			}
		})
	}
}
