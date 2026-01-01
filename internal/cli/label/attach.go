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

// AttachCmd returns the label attach subcommand
func AttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach a label to a task",
		Long: `Attach a label to a task by their IDs.

Examples:
  # Attach label to task
  paso label attach --task=5 --label=2

  # JSON output
  paso label attach --task=5 --label=2 --json

  # Quiet mode
  paso label attach --task=5 --label=2 --quiet
`,
		RunE: runAttach,
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

func runAttach(cmd *cobra.Command, args []string) error {
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

	// Validate task exists
	task, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get task's project ID via column
	column, err := cliInstance.App.ColumnService.GetColumnByID(ctx, task.ColumnID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	taskProjectID := column.ProjectID

	// Validate label exists
	label, err := cli.GetLabelByID(ctx, cliInstance, labelID)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_NOT_FOUND", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Verify task and label belong to same project
	if taskProjectID != label.ProjectID {
		if fmtErr := formatter.Error("PROJECT_MISMATCH", fmt.Sprintf("task %d and label %d do not belong to the same project", taskID, labelID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitValidation)
	}

	// Attach label to task
	if err := cliInstance.App.TaskService.AttachLabel(ctx, taskID, labelID); err != nil {
		if fmtErr := formatter.Error("ATTACH_ERROR", err.Error()); fmtErr != nil {
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

	fmt.Printf("✓ Label '%s' attached to task #%d\n", label.Name, taskID)
	return nil
}
