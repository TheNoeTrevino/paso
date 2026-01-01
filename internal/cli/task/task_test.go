package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/testutil"
)

// TestCreateTaskCommand tests the create command with basic task creation
func TestCreateTaskCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "create task with required flags",
			args:      []string{"--title", "Test Task", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				taskID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric task ID, got: %s", output)
				}
				if taskID <= 0 {
					t.Errorf("Expected positive task ID, got: %d", taskID)
				}
			},
		},
		{
			name:      "create task with JSON output",
			args:      []string{"--title", "JSON Task", "--project", strconv.Itoa(projectID), "--json"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Failed to parse JSON output: %v", err)
				}
				if !result["success"].(bool) {
					t.Error("Expected success=true in JSON output")
				}
				if _, ok := result["task"]; !ok {
					t.Error("Expected 'task' key in JSON output")
				}
			},
		},
		{
			name:      "create task with description",
			args:      []string{"--title", "Task with Desc", "--description", "Test description", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				taskID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric task ID, got: %s", output)
				}
				if taskID <= 0 {
					t.Errorf("Expected positive task ID, got: %d", taskID)
				}
			},
		},
		{
			name:      "create task with priority",
			args:      []string{"--title", "High Priority Task", "--priority", "high", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				taskID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric task ID, got: %s", output)
				}
				if taskID <= 0 {
					t.Errorf("Expected positive task ID, got: %d", taskID)
				}
			},
		},
		{
			name:      "create task with type",
			args:      []string{"--title", "Feature Task", "--type", "feature", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				taskID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric task ID, got: %s", output)
				}
				if taskID <= 0 {
					t.Errorf("Expected positive task ID, got: %d", taskID)
				}
			},
		},
		{
			name:      "create task missing required title",
			args:      []string{"--project", strconv.Itoa(projectID)},
			shouldErr: true,
		},
		{
			name:      "create task invalid priority",
			args:      []string{"--title", "Task", "--priority", "invalid", "--project", strconv.Itoa(projectID)},
			shouldErr: true,
		},
		{
			name:      "create task invalid type",
			args:      []string{"--title", "Task", "--type", "invalid", "--project", strconv.Itoa(projectID)},
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

// TestListTaskCommand tests the list command
func TestListTaskCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Todo")

	// Create some tasks
	_ = testutil.CreateTestTask(t, db, columnID, "Task 1")
	_ = testutil.CreateTestTask(t, db, columnID, "Task 2")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "list tasks with quiet mode",
			args:      []string{"--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) < 2 {
					t.Errorf("Expected at least 2 task IDs, got %d", len(lines))
				}
			},
		},
		{
			name:      "list tasks with JSON output",
			args:      []string{"--project", strconv.Itoa(projectID), "--json"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Failed to parse JSON output: %v", err)
				}
				if !result["success"].(bool) {
					t.Error("Expected success=true in JSON output")
				}
				if _, ok := result["tasks"]; !ok {
					t.Error("Expected 'tasks' key in JSON output")
				}
			},
		},
		{
			name:      "list tasks with human-readable output",
			args:      []string{"--project", strconv.Itoa(projectID)},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "Found") || !strings.Contains(output, "Task 1") || !strings.Contains(output, "Task 2") {
					t.Errorf("Output missing expected task information")
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

// TestUpdateTaskCommand tests the update command
func TestUpdateTaskCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Todo")
	taskID := testutil.CreateTestTask(t, db, columnID, "Original Title")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "update task title",
			args:      []string{"--id", strconv.Itoa(taskID), "--title", "Updated Title", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != strconv.Itoa(taskID) {
					t.Errorf("Expected task ID %d in output, got: %s", taskID, output)
				}
			},
		},
		{
			name:      "update task description",
			args:      []string{"--id", strconv.Itoa(taskID), "--description", "New description", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != strconv.Itoa(taskID) {
					t.Errorf("Expected task ID %d in output, got: %s", taskID, output)
				}
			},
		},
		{
			name:      "update task priority",
			args:      []string{"--id", strconv.Itoa(taskID), "--priority", "critical", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != strconv.Itoa(taskID) {
					t.Errorf("Expected task ID %d in output, got: %s", taskID, output)
				}
			},
		},
		{
			name:      "update task with JSON output",
			args:      []string{"--id", strconv.Itoa(taskID), "--title", "JSON Updated", "--json"},
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
		{
			name:      "update task with no changes",
			args:      []string{"--id", strconv.Itoa(taskID)},
			shouldErr: true,
		},
		{
			name:      "update task with invalid priority",
			args:      []string{"--id", strconv.Itoa(taskID), "--priority", "invalid"},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := UpdateCmd()
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

// TestDeleteTaskCommand tests the delete command
func TestDeleteTaskCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Todo")
	taskID := testutil.CreateTestTask(t, db, columnID, "Task to Delete")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "delete task with force flag",
			args:      []string{"--id", strconv.Itoa(taskID), "--force", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				// Quiet mode produces no output for successful delete
				output = strings.TrimSpace(output)
				if output != "" {
					t.Errorf("Expected no output in quiet mode, got: %s", output)
				}
			},
		},
		{
			name:      "delete task with JSON output",
			args:      []string{"--id", strconv.Itoa(taskID), "--force", "--json"},
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

// TestMoveTaskCommand tests the move command
func TestMoveTaskCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	todoCol := testutil.CreateTestColumn(t, db, projectID, "Todo")
	inProgressCol := testutil.CreateTestColumn(t, db, projectID, "In Progress")
	doneCol := testutil.CreateTestColumn(t, db, projectID, "Done")

	// Link columns to form a chain
	updateColumnNextSQL(t, db, todoCol, inProgressCol)
	updateColumnNextSQL(t, db, inProgressCol, doneCol)

	taskID := testutil.CreateTestTask(t, db, todoCol, "Task to Move")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "move task to next column",
			args:      []string{"--id", strconv.Itoa(taskID), "next", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != strconv.Itoa(taskID) {
					t.Errorf("Expected task ID %d in output, got: %s", taskID, output)
				}
			},
		},
		{
			name:      "move task to specific column by name",
			args:      []string{"--id", strconv.Itoa(taskID), "Done", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != strconv.Itoa(taskID) {
					t.Errorf("Expected task ID %d in output, got: %s", taskID, output)
				}
			},
		},
		{
			name:      "move task with JSON output",
			args:      []string{"--id", strconv.Itoa(taskID), "next", "--json"},
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
			cmd := MoveCmd()
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

// Helper function to update column next_id for linking columns
func updateColumnNextSQL(t *testing.T, db *sql.DB, currentID, nextID int) {
	_, err := db.ExecContext(context.Background(), "UPDATE columns SET next_id = ? WHERE id = ?", nextID, currentID)
	if err != nil {
		t.Fatalf("Failed to update column: %v", err)
	}
	_, err = db.ExecContext(context.Background(), "UPDATE columns SET prev_id = ? WHERE id = ?", currentID, nextID)
	if err != nil {
		t.Fatalf("Failed to update column prev: %v", err)
	}
}

// TestCreateTaskWithRelationships tests task creation with parent/blocking relationships
func TestCreateTaskWithRelationships(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Todo")
	parentTaskID := testutil.CreateTestTask(t, db, columnID, "Parent Task")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
	}{
		{
			name:      "create task with parent",
			args:      []string{"--title", "Child Task", "--parent", strconv.Itoa(parentTaskID), "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
		},
		{
			name:      "create task with blocking relationship",
			args:      []string{"--title", "Task A", "--blocks", strconv.Itoa(parentTaskID), "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
		},
		{
			name:      "create task with blocked-by relationship",
			args:      []string{"--title", "Task B", "--blocked-by", strconv.Itoa(parentTaskID), "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
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

			if !tt.shouldErr {
				output = strings.TrimSpace(output)
				taskID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric task ID, got: %s", output)
				}
				if taskID <= 0 {
					t.Errorf("Expected positive task ID, got: %d", taskID)
				}
			}
		})
	}
}

// TestCreateTaskWithCustomColumn tests task creation with specified column
func TestCreateTaskWithCustomColumn(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	_ = testutil.CreateTestColumn(t, db, projectID, "Todo")
	_ = testutil.CreateTestColumn(t, db, projectID, "In Progress")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
	}{
		{
			name:      "create task in specific column by name",
			args:      []string{"--title", "Task in Progress", "--column", "In Progress", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
		},
		{
			name:      "create task in first column by default",
			args:      []string{"--title", "Task in Todo", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
		},
		{
			name:      "create task with invalid column",
			args:      []string{"--title", "Task", "--column", "Nonexistent", "--project", strconv.Itoa(projectID)},
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

			if !tt.shouldErr {
				output = strings.TrimSpace(output)
				taskID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric task ID, got: %s", output)
				}
				if taskID <= 0 {
					t.Errorf("Expected positive task ID, got: %d", taskID)
				}
			}
		})
	}

}

// TestTaskCommandEnvironmentVariable tests using PASO_PROJECT environment variable
func TestTaskCommandEnvironmentVariable(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")

	// Set environment variable
	oldEnv := os.Getenv("PASO_PROJECT")
	defer os.Setenv("PASO_PROJECT", oldEnv)
	os.Setenv("PASO_PROJECT", strconv.Itoa(projectID))

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
	}{
		{
			name:      "create task using PASO_PROJECT env var",
			args:      []string{"--title", "Task with Env Project", "--quiet"},
			shouldErr: false,
		},
		{
			name:      "flag overrides environment variable",
			args:      []string{"--title", "Task with Flag Override", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CreateCmd()
			cmd.SetArgs(tt.args)

			_, err := testutil.ExecuteCommand(t, cmd)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
