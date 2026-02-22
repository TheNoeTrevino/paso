package standup

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
)

// DeleteCmd returns the standup delete subcommand
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a standup log",
		Long: `Delete a standup log entry by ID.

Examples:
  # Delete a log entry
  paso standup delete -i 5

  # Quiet mode
  paso standup delete -i 5 -q`,
		RunE: runDelete,
	}

	cmd.Flags().IntP("id", "i", 0, "Standup log ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runDelete(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	logID, _ := cmd.Flags().GetInt("id")
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

	if err := cliInstance.App.StandupLogService.Delete(ctx, logID); err != nil {
		return formatter.Error(cli.ExitError, "LOG_DELETE_ERROR", err.Error())
	}

	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"deleted": logID,
		})
	}

	fmt.Print(styles.RenderSuccess(fmt.Sprintf("Standup log %d deleted", logID), cli.GetColorScheme()))
	return nil
}
