package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
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

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	clearEstimate, _ := cmd.Flags().GetBool("clear")
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

	// Verify task exists
	_, err = cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID))
	}

	if clearEstimate {
		return updateEstimate(cmd, cliInstance, taskID, nil, formatter, jsonOutput, quietMode)
	}

	if len(args) < 2 {
		return formatter.Error(cli.ExitValidation, "MISSING_ESTIMATE", "estimate value is required (or use --clear to remove)")
	}

	estimate := args[1]

	// Validate estimate format before calling service
	if err := taskservice.ValidateEstimate(&estimate); err != nil {
		return formatter.Error(cli.ExitValidation, "INVALID_ESTIMATE", err.Error())
	}

	return updateEstimate(cmd, cliInstance, taskID, &estimate, formatter, jsonOutput, quietMode)
}

func updateEstimate(cmd *cobra.Command, cliInstance *cli.CLI, taskID int, estimate *string, formatter *cli.OutputFormatter, jsonOutput, quietMode bool) error {
	ctx := cmd.Context()

	err := cliInstance.App.TaskService.UpdateTaskEstimate(ctx, taskID, estimate)
	if err != nil {
		return formatter.Error(cli.ExitError, "ESTIMATE_ERROR", err.Error())
	}

	if quietMode {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if jsonOutput {
		result := map[string]any{
			"success": true,
			"task_id": taskID,
		}
		if estimate != nil {
			result["estimate"] = *estimate
		} else {
			result["estimate"] = nil
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	colorScheme := cli.GetColorScheme()
	if estimate != nil {
		fmt.Print(styles.RenderSuccess(fmt.Sprintf("Task %d estimate set to %s", taskID, *estimate), colorScheme))
	} else {
		fmt.Print(styles.RenderSuccess(fmt.Sprintf("Task %d estimate cleared", taskID), colorScheme))
	}
	return nil
}
