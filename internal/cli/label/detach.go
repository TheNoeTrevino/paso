package label

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// DetachCmd returns the label detach subcommand
func DetachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach a label from a task",
		Long: `Detach a label from a task by their IDs.

Examples:
  # Detach label from task (shorthand)
  paso label detach -t 5 -l 2

  # JSON output
  paso label detach -t 5 -l 2 -j

  # Quiet mode
  paso label detach -t 5 -l 2 -q

  # Long-form flags also supported
  paso label detach --task=5 --label=2 --json
`,
		RunE: runDetach,
	}

	// Required flags
	cmd.Flags().IntP("task", "t", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("task"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	cmd.Flags().IntP("label", "l", 0, "Label ID (required)")
	if err := cmd.MarkFlagRequired("label"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runDetach(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	taskID, _ := cmd.Flags().GetInt("task")
	labelID, _ := cmd.Flags().GetInt("label")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Detach label from task (no validation needed - removing non-existent association is not an error)
	if err := cliInstance.App.TaskService.DetachLabel(ctx, taskID, labelID); err != nil {
		return formatter.Error(cli.ExitError, "DETACH_ERROR", err.Error())
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":  true,
			"task_id":  taskID,
			"label_id": labelID,
		})
	}

	message := fmt.Sprintf("Label #%d detached from task #%d", labelID, taskID)
	cli.PrintSuccess(message)
	return nil
}
