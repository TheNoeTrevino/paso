package label

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/testutil"
)

// TestCreateLabelCommand tests the label create command
func TestCreateLabelCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "create label with required flags",
			args:      []string{"--name", "bug", "--color", "#FF0000", "--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				labelID, err := strconv.Atoi(output)
				if err != nil {
					t.Fatalf("Expected numeric label ID, got: %s", output)
				}
				if labelID <= 0 {
					t.Errorf("Expected positive label ID, got: %d", labelID)
				}
			},
		},
		{
			name:      "create label with JSON output",
			args:      []string{"--name", "feature", "--color", "#00FF00", "--project", strconv.Itoa(projectID), "--json"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Failed to parse JSON output: %v", err)
				}
				if !result["success"].(bool) {
					t.Error("Expected success=true in JSON output")
				}
				if _, ok := result["label"]; !ok {
					t.Error("Expected 'label' key in JSON output")
				}
			},
		},
		{
			name:      "create label with human-readable output",
			args:      []string{"--name", "enhancement", "--color", "#0000FF", "--project", strconv.Itoa(projectID)},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "created successfully") {
					t.Errorf("Output missing success message")
				}
				if !strings.Contains(output, "enhancement") {
					t.Errorf("Output missing label name")
				}
				if !strings.Contains(output, "#0000FF") {
					t.Errorf("Output missing color")
				}
			},
		},
		{
			name:      "create label missing name",
			args:      []string{"--color", "#FF0000", "--project", strconv.Itoa(projectID)},
			shouldErr: true,
		},
		{
			name:      "create label missing color",
			args:      []string{"--name", "test", "--project", strconv.Itoa(projectID)},
			shouldErr: true,
		},
		{
			name:      "create label missing project",
			args:      []string{"--name", "test", "--color", "#FF0000"},
			shouldErr: true,
		},
		{
			name:      "create label with invalid color format",
			args:      []string{"--name", "test", "--color", "FF0000", "--project", strconv.Itoa(projectID)},
			shouldErr: true,
		},
		{
			name:      "create label with invalid hex color",
			args:      []string{"--name", "test", "--color", "#GGGGGG", "--project", strconv.Itoa(projectID)},
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

// TestListLabelCommand tests the label list command
func TestListLabelCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")

	// Create some labels
	testutil.CreateTestLabel(t, db, projectID, "bug", "#FF0000")
	testutil.CreateTestLabel(t, db, projectID, "feature", "#00FF00")
	testutil.CreateTestLabel(t, db, projectID, "documentation", "#0000FF")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "list labels with quiet mode",
			args:      []string{"--project", strconv.Itoa(projectID), "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) < 3 {
					t.Errorf("Expected at least 3 labels, got %d", len(lines))
				}
			},
		},
		{
			name:      "list labels with JSON output",
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
				if _, ok := result["labels"]; !ok {
					t.Error("Expected 'labels' key in JSON output")
				}
			},
		},
		{
			name:      "list labels with human-readable output",
			args:      []string{"--project", strconv.Itoa(projectID)},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "bug") || !strings.Contains(output, "feature") || !strings.Contains(output, "documentation") {
					t.Errorf("Output missing expected label names")
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

// TestUpdateLabelCommand tests the label update command
func TestUpdateLabelCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	labelID := testutil.CreateTestLabel(t, db, projectID, "original", "#FF0000")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "update label name",
			args:      []string{"--id", strconv.Itoa(labelID), "--name", "updated", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != strconv.Itoa(labelID) {
					t.Errorf("Expected label ID %d in output, got: %s", labelID, output)
				}
			},
		},
		{
			name:      "update label color",
			args:      []string{"--id", strconv.Itoa(labelID), "--color", "#00FF00", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != strconv.Itoa(labelID) {
					t.Errorf("Expected label ID %d in output, got: %s", labelID, output)
				}
			},
		},
		{
			name:      "update label with JSON output",
			args:      []string{"--id", strconv.Itoa(labelID), "--name", "json-updated", "--json"},
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
			name:      "update label with invalid color",
			args:      []string{"--id", strconv.Itoa(labelID), "--color", "INVALID"},
			shouldErr: true,
		},
		{
			name:      "update label without changes",
			args:      []string{"--id", strconv.Itoa(labelID)},
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

// TestDeleteLabelCommand tests the label delete command
func TestDeleteLabelCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	labelID := testutil.CreateTestLabel(t, db, projectID, "to-delete", "#FF0000")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "delete label with force flag",
			args:      []string{"--id", strconv.Itoa(labelID), "--force", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output != "" {
					t.Errorf("Expected no output in quiet mode, got: %s", output)
				}
			},
		},
		{
			name:      "delete label with JSON output",
			args:      []string{"--id", strconv.Itoa(labelID), "--force", "--json"},
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

// TestAttachLabelCommand tests attaching labels to tasks
func TestAttachLabelCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Todo")
	taskID := testutil.CreateTestTask(t, db, columnID, "Task")
	labelID := testutil.CreateTestLabel(t, db, projectID, "label1", "#FF0000")

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
	}{
		{
			name:      "attach label to task",
			args:      []string{"--task", strconv.Itoa(taskID), "--label", strconv.Itoa(labelID), "--quiet"},
			shouldErr: false,
		},
		{
			name:      "attach label with JSON output",
			args:      []string{"--task", strconv.Itoa(taskID), "--label", strconv.Itoa(labelID), "--json"},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := AttachCmd()
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

// TestDetachLabelCommand tests detaching labels from tasks
func TestDetachLabelCommand(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Todo")
	taskID := testutil.CreateTestTask(t, db, columnID, "Task")
	labelID := testutil.CreateTestLabel(t, db, projectID, "label1", "#FF0000")

	// Attach label first
	_, err := db.ExecContext(context.Background(), "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
	if err != nil {
		t.Fatalf("Failed to attach label: %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
	}{
		{
			name:      "detach label from task",
			args:      []string{"--task", strconv.Itoa(taskID), "--label", strconv.Itoa(labelID), "--quiet"},
			shouldErr: false,
		},
		{
			name:      "detach label with JSON output",
			args:      []string{"--task", strconv.Itoa(taskID), "--label", strconv.Itoa(labelID), "--json"},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := DetachCmd()
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

// TestColorValidation tests various valid and invalid color formats
func TestColorValidation(t *testing.T) {
	db, _ := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")

	tests := []struct {
		name      string
		color     string
		shouldErr bool
	}{
		{
			name:      "valid uppercase hex",
			color:     "#FF0000",
			shouldErr: false,
		},
		{
			name:      "valid lowercase hex",
			color:     "#ff0000",
			shouldErr: false,
		},
		{
			name:      "valid mixed case hex",
			color:     "#Ff00Ff",
			shouldErr: false,
		},
		{
			name:      "missing hash",
			color:     "FF0000",
			shouldErr: true,
		},
		{
			name:      "invalid hex characters",
			color:     "#GG0000",
			shouldErr: true,
		},
		{
			name:      "short hex",
			color:     "#FF0",
			shouldErr: true,
		},
		{
			name:      "long hex",
			color:     "#FF00000",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CreateCmd()
			cmd.SetArgs([]string{"--name", "test", "--color", tt.color, "--project", strconv.Itoa(projectID)})

			_, err := testutil.ExecuteCommand(t, cmd)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for color %s but got none", tt.color)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for valid color %s: %v", tt.color, err)
			}
		})
	}
}
