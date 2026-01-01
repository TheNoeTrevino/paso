package column

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/testutil"
)

// TestCreateColumnCommand tests the column create command
func TestCreateColumnCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "create column with required flags",
			args:      []string{"--name", "Review", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				columnID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric column ID, got: %s", output)
				}
				if columnID <= 0 {
					t.Errorf("Expected positive column ID, got: %d", columnID)
				}
			},
		},
		{
			name:      "create column with JSON output",
			args:      []string{"--name", "Testing", "--project", strconv.Itoa(projectID), "--json"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Failed to parse JSON output: %v", err)
				}
				if !result["success"].(bool) {
					t.Error("Expected success=true in JSON output")
				}
				if _, ok := result["column"]; !ok {
					t.Error("Expected 'column' key in JSON output")
				}
			},
		},
		{
			name:      "create column with human-readable output",
			args:      []string{"--name", "Deploy", "--project", strconv.Itoa(projectID)},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "created successfully") {
					t.Errorf("Output missing success message")
				}
				if !strings.Contains(output, "Deploy") {
					t.Errorf("Output missing column name")
				}
			},
		},
		{
			name:      "create column missing name",
			args:      []string{"--project", strconv.Itoa(projectID)},
			shouldErr: true,
		},
		{
			name:      "create column missing project",
			args:      []string{"--name", "Test Column"},
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

// TestListColumnCommand tests the column list command
func TestListColumnCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")

	// Create additional columns
	testutil.CreateTestColumn(t, db, projectID, "Backlog")
	testutil.CreateTestColumn(t, db, projectID, "Ready")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "list columns with quiet mode",
			args:      []string{"--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) < 5 { // 3 default + 2 created = 5
					t.Errorf("Expected at least 5 columns, got %d", len(lines))
				}
			},
		},
		{
			name:      "list columns with JSON output",
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
				if _, ok := result["columns"]; !ok {
					t.Error("Expected 'columns' key in JSON output")
				}
			},
		},
		{
			name:      "list columns with human-readable output",
			args:      []string{"--project", strconv.Itoa(projectID)},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "Backlog") || !strings.Contains(output, "Ready") {
					t.Errorf("Output missing expected column names")
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

// TestUpdateColumnCommand tests the column update command
func TestUpdateColumnCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Original Name")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "update column name",
			args:      []string{"--id", strconv.Itoa(columnID), "--name", "Updated Name", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != strconv.Itoa(columnID) {
					t.Errorf("Expected column ID %d in output, got: %s", columnID, output)
				}
			},
		},
		{
			name:      "update column with JSON output",
			args:      []string{"--id", strconv.Itoa(columnID), "--name", "JSON Updated", "--json"},
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
			name:      "update column with human-readable output",
			args:      []string{"--id", strconv.Itoa(columnID), "--name", "Human Updated"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "updated successfully") {
					t.Errorf("Output missing success message")
				}
			},
		},
		{
			name:      "update column without changes",
			args:      []string{"--id", strconv.Itoa(columnID)},
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

// TestDeleteColumnCommand tests the column delete command
func TestDeleteColumnCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Column to Delete")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "delete column with force flag",
			args:      []string{"--id", strconv.Itoa(columnID), "--force", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != "" {
					t.Errorf("Expected no output in quiet mode, got: %s", output)
				}
			},
		},
		{
			name:      "delete column with JSON output",
			args:      []string{"--id", strconv.Itoa(columnID), "--force", "--json"},
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

// TestCreateColumnWithPositioning tests column creation with specific positioning
func TestCreateColumnWithPositioning(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	afterCol := testutil.CreateTestColumn(t, db, projectID, "After Column")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
	}{
		{
			name:      "create column after specific column",
			args:      []string{"--name", "Inserted Column", "--project", strconv.Itoa(projectID), "--after", strconv.Itoa(afterCol), "--quiet"},
			shouldErr: false,
		},
		{
			name:      "create column at end (default)",
			args:      []string{"--name", "End Column", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
		},
		{
			name:      "create column with invalid after ID",
			args:      []string{"--name", "Bad Column", "--project", strconv.Itoa(projectID), "--after", "99999"},
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
				columnID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric column ID, got: %s", output)
				}
				if columnID <= 0 {
					t.Errorf("Expected positive column ID, got: %d", columnID)
				}
			}
		})
	}
}

// TestColumnSpecialProperties tests creating columns with special properties
func TestColumnSpecialProperties(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
	}{
		{
			name:      "create ready column",
			args:      []string{"--name", "Ready Queue", "--project", strconv.Itoa(projectID), "--ready", "--quiet"},
			shouldErr: false,
		},
		{
			name:      "create completed column",
			args:      []string{"--name", "Completed Tasks", "--project", strconv.Itoa(projectID), "--completed", "--quiet"},
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
				columnID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric column ID, got: %s", output)
				}
				if columnID <= 0 {
					t.Errorf("Expected positive column ID, got: %d", columnID)
				}
			}
		})
	}
}
