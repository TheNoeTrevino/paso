package assignee

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

func PickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pick",
		Short: "Interactively pick an assignee",
		Long: `Interactively pick an assignee from a filterable list and output its ID.

Designed for command chaining with other paso commands.

Examples:
  # Pick an assignee and assign them to a task
  paso task assign 42 $(paso assignee pick)

  # Pick an assignee and set as active
  paso assignee set -n $(paso assignee pick)
`,
		RunE: runPick,
	}
	return cmd
}

func runPick(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	formatter := &cli.OutputFormatter{}

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	assignees, err := cliInstance.App.AssigneeService.List(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "ASSIGNEE_FETCH_ERROR", err.Error())
	}

	if len(assignees) == 0 {
		return formatter.Error(cli.ExitError, "NO_ASSIGNEES", "No assignees found")
	}

	selected, err := cli.RunPick("Pick an assignee", assignees)
	if err != nil {
		return formatter.Error(cli.ExitError, "PICK_CANCELLED", "Selection cancelled")
	}

	fmt.Println(selected)
	return nil
}
