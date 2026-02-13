package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
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
	input, err := ParseReadyMoveArgs(args)
	if err != nil {
		formatter := &cli.OutputFormatter{}
		return formatter.Error(cli.ExitValidation, "INVALID_ID", err.Error())
	}

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

	// Get task detail before move for output
	taskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, input.TaskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", input.TaskID))
	}

	currentColumnName := taskDetail.ColumnName

	// Move task to ready column
	err = cliInstance.App.TaskService.MoveTaskToReadyColumn(ctx, input.TaskID)
	if err != nil {
		// Check for specific errors
		if err == taskservice.ErrTaskAlreadyInTargetColumn {
			// Write to stderr as per requirements
			fmt.Fprintf(os.Stderr, "Task %d is already in the ready column (%s)\n", input.TaskID, currentColumnName)
			// Still exit successfully
			if quietMode {
				fmt.Printf("%s\n", FormatReadyMoveQuiet(&ReadyMoveResult{TaskID: input.TaskID}))
			}
			return nil
		}
		if strings.Contains(err.Error(), "no ready column configured") {
			return formatter.ErrorWithSuggestion(cli.ExitValidation, "NO_READY_COLUMN",
				"no ready column configured for this project",
				"Use 'paso column update --id=<column_id> --ready' to designate a ready column")
		}
		return formatter.Error(cli.ExitError, "MOVE_ERROR", err.Error())
	}

	// Get updated task detail for output
	updatedTaskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, input.TaskID)
	if err != nil {
		return formatter.Error(cli.ExitError, "TASK_FETCH_ERROR", err.Error())
	}

	toColumnName := updatedTaskDetail.ColumnName

	result := &ReadyMoveResult{
		TaskID:     input.TaskID,
		FromColumn: currentColumnName,
		ToColumn:   toColumnName,
	}

	// Output success
	if quietMode {
		fmt.Print(FormatReadyMoveQuiet(result))
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatReadyMoveJSON(result))
	}

	message := FormatReadyMoveOutput(result)
	cli.PrintSuccess(message)
	return nil
}
