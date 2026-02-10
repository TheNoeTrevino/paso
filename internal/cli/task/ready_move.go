package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// ReadyMoveCmd returns the task to-ready subcommand
func ReadyMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to-ready <task_id>",
		Short: "Move a task to the ready column",
		Long: `Move a task to the column designated as holding ready tasks.

The ready column is marked with holds_ready_tasks = true.
Use 'paso column update -i <column_id> -r' to designate a ready column.

Note: This command moves tasks to the column marked as ready (see: paso column update --ready)

Examples:
  # Move task to ready column
  paso task to-ready 42

  # JSON output for agents
  paso task to-ready 42 -j

  # Quiet mode for bash capture
  paso task to-ready 42 -q

  # Long-form flags also supported
  paso task to-ready 42 --json
`,
		RunE: runReadyMove,
		Args: cobra.ExactArgs(1),
	}

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runReadyMove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Parse task ID from positional argument
	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		formatter := &cli.OutputFormatter{}
		if fmtErr := formatter.Error("INVALID_ID", fmt.Sprintf("invalid task ID '%s': must be a number", args[0])); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitValidation)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Get task detail before move for output
	taskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	currentColumnName := taskDetail.ColumnName

	// Move task to ready column
	err = cliInstance.App.TaskService.MoveTaskToReadyColumn(ctx, taskID)
	if err != nil {
		// Check for specific errors
		if err == taskservice.ErrTaskAlreadyInTargetColumn {
			// Write to stderr as per requirements
			fmt.Fprintf(os.Stderr, "Task %d is already in the ready column (%s)\n", taskID, currentColumnName)
			// Still exit successfully
			if quietMode {
				fmt.Printf("%d\n", taskID)
			}
			return nil
		}
		if strings.Contains(err.Error(), "no ready column configured") {
			if fmtErr := formatter.ErrorWithSuggestion("NO_READY_COLUMN",
				"no ready column configured for this project",
				"Use 'paso column update --id=<column_id> --ready' to designate a ready column"); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitValidation)
		}
		if fmtErr := formatter.Error("MOVE_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}

	// Get updated task detail for output
	updatedTaskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}

	toColumnName := updatedTaskDetail.ColumnName

	// Output success
	if quietMode {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":     true,
			"task_id":     taskID,
			"from_column": currentColumnName,
			"to_column":   toColumnName,
		})
	}

	// Human-readable output
	colors := cli.GetColorScheme()
	message := fmt.Sprintf("Task %d moved to '%s'", taskID, toColumnName)
	fmt.Print(styles.RenderSuccess(message, colors))
	return nil
}
