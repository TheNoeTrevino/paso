package label

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

func PickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pick [project-id]",
		Short: "Interactively pick a label",
		Long: `Interactively pick a label from a filterable list and output its ID.

Designed for command chaining with other paso commands.

Examples:
  # Pick a label and attach it to a task
  paso label attach --task=42 --label=$(paso label pick)

  # Pick from a specific project
  paso label pick 2
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

	labels, err := cliInstance.App.LabelService.GetLabelsByProject(ctx, projectID)
	if err != nil {
		return formatter.Error(cli.ExitError, "LABEL_FETCH_ERROR", err.Error())
	}

	if len(labels) == 0 {
		return formatter.Error(cli.ExitError, "NO_LABELS", "No labels found in this project")
	}

	selected, err := cli.RunPick("Pick a label", labels)
	if err != nil {
		return formatter.Error(cli.ExitError, "PICK_CANCELLED", "Selection cancelled")
	}

	fmt.Println(selected)
	return nil
}
