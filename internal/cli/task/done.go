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

// DoneCmd returns the task done subcommand
func DoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <task_id>",
		Short: "Move a task to the completed column",
		Long: `Move a task to the column designated as holding completed tasks.

The completed column is marked with holds_completed_tasks = true.
Use 'paso column update -i <column_id> -c' to designate a completed column.

Note: This command moves tasks to the column marked as completed (see: paso column update --completed)

Examples:
  # Move task to completed column
  paso task done 42

  # JSON output for agents
  paso task done 42 -j

  # Quiet mode for bash capture
  paso task done 42 -q

  # Long-form flags also supported
  paso task done 42 --json
`,
		RunE: runDone,
		Args: cobra.ExactArgs(1),
	}

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runDone(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Parse task ID from positional argument
	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		formatter := &cli.OutputFormatter{}
		return formatter.Error(cli.ExitValidation, "INVALID_ID", fmt.Sprintf("invalid task ID '%s': must be a number", args[0]))
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
	taskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID))
	}

	currentColumnName := taskDetail.ColumnName

	// Move task to completed column
	err = cliInstance.App.TaskService.MoveTaskToCompletedColumn(ctx, taskID)
	if err != nil {
		// Check for specific errors
		if err == taskservice.ErrTaskAlreadyInTargetColumn {
			// Write to stderr as per requirements
			fmt.Fprintf(os.Stderr, "Task %d is already in the completed column (%s)\n", taskID, currentColumnName)
			// Still exit successfully
			if quietMode {
				fmt.Printf("%d\n", taskID)
			}
			return nil
		}
		if strings.Contains(err.Error(), "no completed column configured") {
			return formatter.ErrorWithSuggestion(cli.ExitValidation, "NO_COMPLETED_COLUMN",
				"no completed column configured for this project",
				"Use 'paso column update --id=<column_id> --completed' to designate a completed column")
		}
		return formatter.Error(cli.ExitError, "MOVE_ERROR", err.Error())
	}

	// Get updated task detail for output
	updatedTaskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitError, "TASK_FETCH_ERROR", err.Error())
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
