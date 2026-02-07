package project

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testutilcli "github.com/thenoetrevino/paso/internal/testutil/cli"
)

// TestCreateProjectCommand tests the create command
func TestCreateProjectCommand(t *testing.T) {
	_, app := testutilcli.SetupCLITest(t)

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name:      "create project with name",
			args:      []string{"--title", "New Project", "--quiet"},
			shouldErr: false,
			checkFunc: func(t *testing.T, output string) {
				projectID, err := strconv.Atoi(strings.TrimSpace(output))
				require.NoError(t, err, "Expected numeric project ID, got: %s", output)
				assert.Positive(t, projectID)
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
