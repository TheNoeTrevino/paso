package golden_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/assignee"
	"github.com/thenoetrevino/paso/internal/cli/column"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func assertGoldenOutput(t *testing.T, name, actual string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", name+".golden")

	if *updateGolden {
		err := os.MkdirAll("testdata", 0755)
		require.NoError(t, err, "failed to create testdata directory")
		err = os.WriteFile(goldenPath, []byte(actual), 0644)
		require.NoError(t, err, "failed to write golden file")
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Logf("Golden file does not exist, creating: %s\nRun with -update to regenerate", goldenPath)
			err := os.MkdirAll("testdata", 0755)
			require.NoError(t, err, "failed to create testdata directory")
			err = os.WriteFile(goldenPath, []byte(actual), 0644)
			require.NoError(t, err, "failed to write initial golden file")
			return
		}
		t.Fatalf("golden file not found: %s (run with -update to create)", goldenPath)
	}

	assert.Equal(t, string(expected), actual, "golden file mismatch for %s\n\nRun with -update to regenerate", name)
}

func captureHelpOutput(t *testing.T, cmd *cobra.Command) string {
	t.Helper()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err, "help command should not error")

	return fixtures.StripANSI(buf.String())
}

func TestGolden_HelpOutput(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"task-help", task.TaskCmd},
		{"project-help", project.ProjectCmd},
		{"label-help", label.LabelCmd},
		{"column-help", column.ColumnCmd},
		{"assignee-help", assignee.AssigneeCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()
			output := captureHelpOutput(t, cmd)
			assertGoldenOutput(t, tt.name, output)
		})
	}
}
