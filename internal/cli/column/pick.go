package column

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

func PickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pick [project-id]",
		Short: "Interactively pick a column",
		Long: `Interactively pick a column from a filterable list and output its ID.

Designed for command chaining with other paso commands.

Examples:
  # Pick a column to move a task to
  paso task move -i 42 $(paso column pick)

  # Pick from a specific project
  paso column pick 2
`,
		Args: cobra.MaximumNArgs(1),
		RunE: runPick,
	}

	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

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

	var projectID int
	if len(args) > 0 {
		projectID, err = strconv.Atoi(args[0])
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "INVALID_PROJECT_ID",
				fmt.Sprintf("Invalid project ID: %s", args[0]),
				"Project ID must be a number")
		}
	} else {
		projectID, err = cli.GetProjectIDWithCLI(cmd, cliInstance)
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "NO_PROJECT",
				err.Error(),
				"Specify project ID as argument, use --project flag, or associate branch with project")
		}
	}

	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, projectID)
	if err != nil {
		return formatter.Error(cli.ExitError, "COLUMN_FETCH_ERROR", err.Error())
	}

	if len(columns) == 0 {
		return formatter.Error(cli.ExitError, "NO_COLUMNS", "No columns found in this project")
	}

	selected, err := cli.RunPick("Pick a column", columns)
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return formatter.Error(cli.ExitUsage, "PICK_CANCELLED", "Selection cancelled")
		}
		return formatter.Error(cli.ExitError, "PICK_ERROR", err.Error())
	}

	fmt.Println(selected)
	return nil
}
