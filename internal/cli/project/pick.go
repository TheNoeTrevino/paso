package project

import (
	"errors"
	"fmt"
	"log/slog"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

func PickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pick",
		Short: "Interactively pick a project",
		Long: `Interactively pick a project from a filterable list and output its ID.

Designed for command chaining with other paso commands.

Examples:
  # Pick a project and list its tasks
  paso task list $(paso project pick)

  # Pick a project and view its tree
  paso project tree $(paso project pick)
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

	projects, err := cliInstance.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "PROJECT_FETCH_ERROR", err.Error())
	}

	if len(projects) == 0 {
		return formatter.Error(cli.ExitError, "NO_PROJECTS", "No projects found")
	}

	selected, err := cli.RunPick("Pick a project", projects)
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return formatter.Error(cli.ExitUsage, "PICK_CANCELLED", "Selection cancelled")
		}
		return formatter.Error(cli.ExitError, "PICK_ERROR", err.Error())
	}

	fmt.Println(selected)
	return nil
}
