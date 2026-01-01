package label

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
  # Detach label from task
  paso label detach --task=5 --label=2

  # JSON output
  paso label detach --task=5 --label=2 --json

  # Quiet mode
  paso label detach --task=5 --label=2 --quiet
`,
		RunE: runDetach,
	}

	// Required flags
	cmd.Flags().Int("task", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("task"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Int("label", 0, "Label ID (required)")
	if err := cmd.MarkFlagRequired("label"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runDetach(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	taskID, _ := cmd.Flags().GetInt("task")
	labelID, _ := cmd.Flags().GetInt("label")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Detach label from task (no validation needed - removing non-existent association is not an error)
	if err := cliInstance.App.TaskService.DetachLabel(ctx, taskID, labelID); err != nil {
		if fmtErr := formatter.Error("DETACH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success":  true,
			"task_id":  taskID,
			"label_id": labelID,
		})
	}

	fmt.Printf("✓ Label #%d detached from task #%d\n", labelID, taskID)
	return nil
}
