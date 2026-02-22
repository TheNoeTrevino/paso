package standup

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// LogCmd returns the standup log subcommand
func LogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log a standup entry",
		Long: `Log what you've been working on.

If -m is not provided, your $EDITOR will be opened to compose the message.

Examples:
  # Quick log with message flag
  paso standup log -m "Fixed the auth bug in login flow"

  # Open editor to write log
  paso standup log

  # Specify project explicitly
  paso standup log -m "Added unit tests" -p 2

  # JSON output for agents
  paso standup log -m "Refactored service layer" -j

  # Quiet mode (returns log ID)
  LOG_ID=$(paso standup log -m "Done for today" -q)`,
		RunE: runLog,
	}

	cmd.Flags().StringP("message", "m", "", "Log message (opens $EDITOR if not provided)")
	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (log ID only)")

	return cmd
}

func runLog(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	message, _ := cmd.Flags().GetString("message")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// If no message flag, open editor
	if message == "" {
		var err error
		message, err = openEditor(ctx)
		if err != nil {
			return formatter.Error(cli.ExitError, "EDITOR_ERROR", err.Error())
		}
		if message == "" {
			return formatter.Error(cli.ExitValidation, "EMPTY_MESSAGE", "aborting: empty log message")
		}
	}

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

	log, err := cliInstance.App.StandupLogService.Create(ctx, projectID, message)
	if err != nil {
		return formatter.Error(cli.ExitError, "LOG_CREATE_ERROR", err.Error())
	}

	result := &LogResult{
		ID:        log.ID,
		ProjectID: log.ProjectID,
		Content:   log.Content,
		CreatedAt: log.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if quietMode {
		fmt.Print(FormatLogQuiet(result.ID))
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatLogJSON(result))
	}

	fmt.Print(FormatLogHuman(result, cli.GetColorScheme()))
	return nil
}
