package task

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "Project creation failed - GetProjectByID returned error (projectID=%d)", projectID)
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
				require.NoError(t, err, "Expected numeric task ID, got: %s", output)
				assert.Positive(t, taskID)
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

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if !tt.shouldErr && tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}
		})
	}
}
