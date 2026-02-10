package label

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
)

// AttachCmd returns the label attach subcommand
func AttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach a label to a task",
		Long: `Attach a label to a task by their IDs.

Examples:
  # Attach label to task (shorthand)
  paso label attach -t 5 -l 2

  # JSON output
  paso label attach -t 5 -l 2 -j

  # Quiet mode
  paso label attach -t 5 -l 2 -q

  # Long-form flags also supported
  paso label attach --task=5 --label=2 --json
`,
		RunE: runAttach,
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

func runAttach(cmd *cobra.Command, args []string) error {
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

	// Validate task exists
	task, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID))
	}

	// Get task's project ID via column
	column, err := cliInstance.App.ColumnService.GetColumnByID(ctx, task.ColumnID)
	if err != nil {
		return formatter.Error(cli.ExitError, "COLUMN_FETCH_ERROR", err.Error())
	}
	taskProjectID := column.ProjectID

	// Validate label exists
	label, err := cli.GetLabelByID(ctx, cliInstance, labelID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "LABEL_NOT_FOUND", err.Error())
	}

	// Verify task and label belong to same project
	if taskProjectID != label.ProjectID {
		return formatter.Error(cli.ExitValidation, "PROJECT_MISMATCH", fmt.Sprintf("task %d and label %d do not belong to the same project", taskID, labelID))
	}

	// Attach label to task
	if err := cliInstance.App.TaskService.AttachLabel(ctx, taskID, labelID); err != nil {
		return formatter.Error(cli.ExitError, "ATTACH_ERROR", err.Error())
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

	colors := cli.GetColorScheme()
	message := fmt.Sprintf("Label '%s' attached to task #%d", label.Name, taskID)
	fmt.Print(styles.RenderSuccess(message, colors))
	return nil
}
