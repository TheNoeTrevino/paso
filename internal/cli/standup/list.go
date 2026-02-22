package standup

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// ListCmd returns the standup list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List standup logs",
		Long: `List all standup logs for the current project.

Examples:
  # List logs for current project (via git branch)
  paso standup list

  # List logs for a specific project
  paso standup list -p 2

  # JSON output
  paso standup list -j`,
		RunE: runList,
	}

	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (IDs only)")

	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	projectID, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		return formatter.ErrorWithSuggestion(cli.ExitUsage, "NO_PROJECT",
			err.Error(),
			"Specify project ID with --project flag or associate branch with a project")
	}

	logs, err := cliInstance.App.StandupLogService.ListByProject(ctx, projectID)
	if err != nil {
		return formatter.Error(cli.ExitError, "LOG_FETCH_ERROR", err.Error())
	}

	if quietMode {
		for _, id := range FormatListQuiet(logs) {
			fmt.Println(id)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatListJSON(logs))
	}

	if len(logs) == 0 {
		fmt.Println("No standup logs found")
		return nil
	}

	fmt.Print(FormatListHuman(logs, cli.GetColorScheme()))
	return nil
}
