package cli_test

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/assignee"
	"github.com/thenoetrevino/paso/internal/cli/column"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
	"github.com/thenoetrevino/paso/internal/testing/snapshots"
)

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
			helper := snapshots.NewHelper(t, "testdata/help")
			helper.Compare(tt.name, output)
		})
	}
}
