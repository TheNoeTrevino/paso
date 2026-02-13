package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// MoveCmd returns the task move subcommand
func MoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move a task to another column",
		Long: `Move a task to another column by direction or column name.

Examples:
  # Move to next/previous column (shorthand)
  paso task move -i 1 next
  paso task move -i 1 prev

  # Move to specific column by name (case-insensitive)
  paso task move -i 1 "In Progress"
  paso task move -i 1 done

  # JSON output for agents
  paso task move -i 1 next -j

  # Quiet mode for bash capture
  paso task move -i 1 next -q

  # Long-form flags also supported
  paso task move --id 1 next --json
`,
		RunE: runMove,
		Args: cobra.ExactArgs(1),
	}

	// Required flags
	cmd.Flags().IntP("id", "i", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")
	cmd.Flags().Bool("dry-run", false, "Show what would change without moving the task")

	return cmd
}

func runMove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	taskID, _ := cmd.Flags().GetInt("id")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	target := args[0]

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

	// Get task detail to find current column and project
	taskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID))
	}

	// Get column to find project ID
	currentColumn, err := cliInstance.App.ColumnService.GetColumnByID(ctx, taskDetail.ColumnID)
	if err != nil {
		return formatter.Error(cli.ExitError, "COLUMN_FETCH_ERROR", err.Error())
	}

	// Get all columns for the project
	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, currentColumn.ProjectID)
	if err != nil {
		return formatter.Error(cli.ExitError, "COLUMN_FETCH_ERROR", err.Error())
	}

	currentColumnName := cli.GetCurrentColumnName(columns, taskDetail.ColumnID)
	var toColumnName string

	// Handle the target: next, prev, or column name
	switch strings.ToLower(target) {
	case "next":
		// Find next column name for output
		toColumnName = findNextColumnName(columns, taskDetail.ColumnID)
		if toColumnName == unknownColumnName {
			return formatter.Error(cli.ExitValidation, "NO_NEXT_COLUMN",
				fmt.Sprintf("task is already in the last column (%s)", currentColumnName))
		}

		if !dryRun {
			err = cliInstance.App.TaskService.MoveTaskToNextColumn(ctx, taskID)
			if err != nil {
				if strings.Contains(err.Error(), "no next column") {
					return formatter.Error(cli.ExitValidation, "NO_NEXT_COLUMN",
						fmt.Sprintf("task is already in the last column (%s)", currentColumnName))
				}
				return formatter.Error(cli.ExitError, "MOVE_ERROR", err.Error())
			}
		}

	case "prev":
		// Find prev column name for output
		toColumnName = findPrevColumnName(columns, taskDetail.ColumnID)
		if toColumnName == unknownColumnName {
			return formatter.Error(cli.ExitValidation, "NO_PREV_COLUMN",
				fmt.Sprintf("task is already in the first column (%s)", currentColumnName))
		}

		if !dryRun {
			err = cliInstance.App.TaskService.MoveTaskToPrevColumn(ctx, taskID)
			if err != nil {
				if strings.Contains(err.Error(), "no previous column") {
					return formatter.Error(cli.ExitValidation, "NO_PREV_COLUMN",
						fmt.Sprintf("task is already in the first column (%s)", currentColumnName))
				}
				return formatter.Error(cli.ExitError, "MOVE_ERROR", err.Error())
			}
		}

	default:
		// Try to find column by name
		targetColumn, err := cli.FindColumnByName(columns, target)
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitNotFound, "COLUMN_NOT_FOUND",
				fmt.Sprintf("column '%s' not found", target),
				fmt.Sprintf("Task is currently in: %s\nAvailable columns: %s",
					currentColumnName, cli.FormatAvailableColumns(columns)))
		}

		toColumnName = targetColumn.Name

		// Check if already in target column (silent success)
		if targetColumn.ID != taskDetail.ColumnID && !dryRun {
			err = cliInstance.App.TaskService.MoveTaskToColumn(ctx, taskID, targetColumn.ID)
			if err != nil {
				return formatter.Error(cli.ExitError, "MOVE_ERROR", err.Error())
			}
		}
	}

	// Output success
	if quietMode {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if jsonOutput {
		result := createMoveResult(taskID, currentColumnName, toColumnName, dryRun)
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	// Human-readable output
	message := formatMoveMessage(taskID, currentColumnName, toColumnName, dryRun)
	cli.PrintSuccess(message)
	return nil
}
