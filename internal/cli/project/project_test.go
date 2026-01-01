package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/testutil"
)

// TestCreateProjectCommand tests the project create command
func TestCreateProjectCommand(t *testing.T) {
	_, _ = testutil.SetupCLITest(t)

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "create project with title only",
			args:      []string{"--title", "My Project", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				projectID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric project ID, got: %s", output)
				}
				if projectID <= 0 {
					t.Errorf("Expected positive project ID, got: %d", projectID)
				}
			},
		},
		{
			name:      "create project with title and description",
			args:      []string{"--title", "Project with Description", "--description", "This is a test project", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				projectID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric project ID, got: %s", output)
				}
				if projectID <= 0 {
					t.Errorf("Expected positive project ID, got: %d", projectID)
				}
			},
		},
		{
			name:      "create project with JSON output",
			args:      []string{"--title", "JSON Project", "--json"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Failed to parse JSON output: %v", err)
				}
				if !result["success"].(bool) {
					t.Error("Expected success=true in JSON output")
				}
				if _, ok := result["project"]; !ok {
					t.Error("Expected 'project' key in JSON output")
				}
			},
		},
		{
			name:      "create project with human-readable output",
			args:      []string{"--title", "Human Project", "--description", "Test"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "created successfully") {
					t.Errorf("Output missing success message")
				}
				if !strings.Contains(output, "Human Project") {
					t.Errorf("Output missing project name")
				}
			},
		},
		{
			name:      "create project missing title",
			args:      []string{},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CreateCmd()
			cmd.SetArgs(tt.args)

			output, err := testutil.ExecuteCommand(t, cmd)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.shouldErr && tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}
		})
	}
}

// TestListProjectCommand tests the project list command
func TestListProjectCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)

	// Create some projects
	testutil.CreateTestProject(t, db, "Project 1")
	testutil.CreateTestProject(t, db, "Project 2")
	testutil.CreateTestProject(t, db, "Project 3")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "list projects with quiet mode",
			args:      []string{"--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) < 3 {
					t.Errorf("Expected at least 3 projects, got %d", len(lines))
				}
			},
		},
		{
			name:      "list projects with JSON output",
			args:      []string{"--json"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Failed to parse JSON output: %v", err)
				}
				if !result["success"].(bool) {
					t.Error("Expected success=true in JSON output")
				}
				if _, ok := result["projects"]; !ok {
					t.Error("Expected 'projects' key in JSON output")
				}
			},
		},
		{
			name:      "list projects with human-readable output",
			args:      []string{},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "Found") {
					t.Errorf("Output missing 'Found' prefix")
				}
				if !strings.Contains(output, "Project") {
					t.Errorf("Output missing project information")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ListCmd()
			cmd.SetArgs(tt.args)

			output, err := testutil.ExecuteCommand(t, cmd)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.shouldErr && tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}
		})
	}
}

// TestDeleteProjectCommand tests the project delete command
func TestDeleteProjectCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Project to Delete")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "delete project with force flag",
			args:      []string{"--id", strconv.Itoa(projectID), "--force", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != "" {
					t.Errorf("Expected no output in quiet mode, got: %s", output)
				}
			},
		},
		{
			name:      "delete project with JSON output",
			args:      []string{"--id", strconv.Itoa(projectID), "--force", "--json"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Failed to parse JSON output: %v", err)
				}
				if !result["success"].(bool) {
					t.Error("Expected success=true in JSON output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := DeleteCmd()
			cmd.SetArgs(tt.args)

			output, err := testutil.ExecuteCommand(t, cmd)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.shouldErr && tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}
		})
	}
}

// TestProjectTreeCommand tests the project tree command for displaying project structure
func TestProjectTreeCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Todo")

	// Create a task with subtasks to test tree display
	taskID := testutil.CreateTestTask(t, db, columnID, "Root Task")
	childTaskID := testutil.CreateTestTask(t, db, columnID, "Child Task")

	// Create parent-child relationship
	_, err := db.ExecContext(
		context.Background(),
		"INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (?, ?, 1)",
		taskID, childTaskID,
	)
	if err != nil {
		t.Fatalf("Failed to create relationship: %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "display project tree with default output",
			args:      []string{"--id", strconv.Itoa(projectID)},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "Root Task") {
					t.Errorf("Output missing expected root task")
				}
			},
		},
		{
			name:      "display project tree with JSON output",
			args:      []string{"--id", strconv.Itoa(projectID), "--json"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				// JSON tree output should be parseable
				var result interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Failed to parse JSON output: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := TreeCmd()
			cmd.SetArgs(tt.args)

			output, err := testutil.ExecuteCommand(t, cmd)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.shouldErr && tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}
		})
	}
}

// TestCreateProjectIntegration tests project creation creates default columns
func TestCreateProjectIntegration(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)

	cmd := CreateCmd()
	cmd.SetArgs([]string{"--title", "Integration Test Project", "--quiet"})

	output, err := testutil.ExecuteCommand(t, cmd)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	output = strings.TrimSpace(output)
	projectID, err := strconv.Atoi(output)
	if err != nil {
		t.Fatalf("Expected numeric project ID, got: %s", output)
	}

	// Verify project was created with default columns
	var count int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM columns WHERE project_id = ?", projectID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check columns: %v", err)
	}

	// Default columns should be created (Todo, In Progress, Done)
	if count < 3 {
		t.Errorf("Expected at least 3 default columns, got %d", count)
	}
}

// Helper to create columns (similar to testutil but for SQL context)
func createTestColumnWithSQL(t *testing.T, db *sql.DB, projectID int, name string) int {
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, name)
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}
	columnID, _ := result.LastInsertId()
	return int(columnID)
}
