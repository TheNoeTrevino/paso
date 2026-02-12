package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// EstimateCmd returns the task estimate subcommand
func EstimateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate <task_id> [estimate]",
		Short: "Set a time estimate on a task",
		Long: `Set a time estimate on a task.

Estimates use a compact format: combinations of numbers followed by units.
Valid units: h (hours), d (days), w (weeks), m (months).

Use --clear to remove the estimate from a task.

Examples:
  # Set a 2-day estimate
  paso task estimate 42 2d

  # Set a compound estimate
  paso task estimate 42 1w2d

  # Remove estimate
  paso task estimate 42 --clear

  # JSON output for agents
  paso task estimate 42 2d -j
`,
		RunE: runEstimate,
		Args: cobra.RangeArgs(1, 2),
	}

	cmd.Flags().Bool("clear", false, "Remove estimate from the task")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runEstimate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	clearEstimate, _ := cmd.Flags().GetBool("clear")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	// Parse input
	input, err := ParseEstimateArgs(args, clearEstimate)
	if err != nil {
		return err
	}

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

	// Verify task exists
	_, err = cliInstance.App.TaskService.GetTaskDetail(ctx, input.TaskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", input.TaskID))
	}

	if input.Clear {
		return updateEstimate(cmd, cliInstance, input.TaskID, nil, formatter, jsonOutput, quietMode)
	}

	if input.Estimate == "" {
		return formatter.Error(cli.ExitValidation, "MISSING_ESTIMATE", "estimate value is required (or use --clear to remove)")
	}

	// Validate estimate format before calling service
	if err := taskservice.ValidateEstimate(&input.Estimate); err != nil {
		return formatter.Error(cli.ExitValidation, "INVALID_ESTIMATE", err.Error())
	}

	return updateEstimate(cmd, cliInstance, input.TaskID, &input.Estimate, formatter, jsonOutput, quietMode)
}

func updateEstimate(cmd *cobra.Command, cliInstance *cli.CLI, taskID int, estimate *string, formatter *cli.OutputFormatter, jsonOutput, quietMode bool) error {
	ctx := cmd.Context()

	err := cliInstance.App.TaskService.UpdateTaskEstimate(ctx, taskID, estimate)
	if err != nil {
		return formatter.Error(cli.ExitError, "ESTIMATE_ERROR", err.Error())
	}

	result := &EstimateResult{
		TaskID:  taskID,
		Cleared: estimate == nil,
	}
	if estimate != nil {
		result.Estimate = *estimate
	}

	if quietMode {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatEstimateJSON(result))
	}

	cli.PrintSuccess(FormatEstimateOutput(result))
	return nil
}
