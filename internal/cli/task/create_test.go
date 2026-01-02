package task

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/testutil"
	testutilcli "github.com/thenoetrevino/paso/internal/testutil/cli"
)

// TestCreateTaskCommand tests the create command
func TestCreateTaskCommand(t *testing.T) {
	db, app := testutilcli.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Todo")

	// Validate that project exists in database
	_, err := app.ProjectService.GetProjectByID(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Project creation failed - GetProjectByID returned error: %v (projectID=%d)", err, projectID)
	}
	t.Logf("Setup successful: projectID=%d, columnID=%d", projectID, columnID)

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "create task with required flags",
			args:      []string{"--title", "Test Task", "--project", strconv.Itoa(projectID), "--column", "Todo", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				taskID, err := strconv.Atoi(strings.TrimSpace(output))
				if err != nil {
					t.Fatalf("Expected numeric task ID, got: %s", output)
				}
				if taskID <= 0 {
					t.Errorf("Expected positive task ID, got: %d", taskID)
				}
			},
		},
		{
			name:      "create task missing title",
			args:      []string{"--project", strconv.Itoa(projectID), "--column", "Todo"},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CreateCmd()
			output, err := testutilcli.ExecuteCLICommand(t, app, cmd, tt.args)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v, output: %s", err, output)
			}

			if !tt.shouldErr && tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}
		})
	}
}
