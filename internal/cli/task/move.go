package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/models"
)

// MoveCmd returns the task move subcommand
func MoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move a task to another column",
		Long: `Move a task to another column by direction or column name.

Examples:
  # Move to next column
  paso task move --id 1 next

  # Move to previous column
  paso task move --id 1 prev

  # Move to specific column by name (case-insensitive)
  paso task move --id 1 "In Progress"
  paso task move --id 1 done

  # JSON output for agents
  paso task move --id 1 next --json

  # Quiet mode for bash capture
  paso task move --id 1 next --quiet
`,
		RunE: runMove,
		Args: cobra.ExactArgs(1),
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runMove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	taskID, _ := cmd.Flags().GetInt("id")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")
	target := args[0]

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to closing CLI", "error", err)
		}
	}()

	// Get task detail to find current column and project
	taskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get column to find project ID
	currentColumn, err := cliInstance.App.ColumnService.GetColumnByID(ctx, taskDetail.ColumnID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	// Get all columns for the project
	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, currentColumn.ProjectID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	currentColumnName := cli.GetCurrentColumnName(columns, taskDetail.ColumnID)
	var toColumnName string

	// Handle the target: next, prev, or column name
	switch strings.ToLower(target) {
	case "next":
		err = cliInstance.App.TaskService.MoveTaskToNextColumn(ctx, taskID)
		if err != nil {
			if strings.Contains(err.Error(), "no next column") {
				if fmtErr := formatter.Error("NO_NEXT_COLUMN",
					fmt.Sprintf("task is already in the last column (%s)", currentColumnName)); fmtErr != nil {
					slog.Error("failed to formatting error message", "error", fmtErr)
				}
				os.Exit(cli.ExitValidation)
			}
			if fmtErr := formatter.Error("MOVE_ERROR", err.Error()); fmtErr != nil {
				slog.Error("failed to formatting error message", "error", fmtErr)
			}
			return err
		}
		// Find next column name for output
		toColumnName = findNextColumnName(columns, taskDetail.ColumnID)

	case "prev":
		err = cliInstance.App.TaskService.MoveTaskToPrevColumn(ctx, taskID)
		if err != nil {
			if strings.Contains(err.Error(), "no previous column") {
				if fmtErr := formatter.Error("NO_PREV_COLUMN",
					fmt.Sprintf("task is already in the first column (%s)", currentColumnName)); fmtErr != nil {
					slog.Error("failed to formatting error message", "error", fmtErr)
				}
				os.Exit(cli.ExitValidation)
			}
			if fmtErr := formatter.Error("MOVE_ERROR", err.Error()); fmtErr != nil {
				slog.Error("failed to formatting error message", "error", fmtErr)
			}
			return err
		}
		// Find prev column name for output
		toColumnName = findPrevColumnName(columns, taskDetail.ColumnID)

	default:
		// Try to find column by name
		targetColumn, err := cli.FindColumnByName(columns, target)
		if err != nil {
			if fmtErr := formatter.ErrorWithSuggestion("COLUMN_NOT_FOUND",
				fmt.Sprintf("column '%s' not found", target),
				fmt.Sprintf("Task is currently in: %s\nAvailable columns: %s",
					currentColumnName, cli.FormatAvailableColumns(columns))); fmtErr != nil {
				slog.Error("failed to formatting error message", "error", fmtErr)
			}
			os.Exit(cli.ExitNotFound)
		}

		// Check if already in target column (silent success)
		if targetColumn.ID == taskDetail.ColumnID {
			toColumnName = targetColumn.Name
			// Skip the move, just output success
		} else {
			err = cliInstance.App.TaskService.MoveTaskToColumn(ctx, taskID, targetColumn.ID)
			if err != nil {
				if fmtErr := formatter.Error("MOVE_ERROR", err.Error()); fmtErr != nil {
					slog.Error("failed to formatting error message", "error", fmtErr)
				}
				return err
			}
			toColumnName = targetColumn.Name
		}
	}

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
	if currentColumnName == toColumnName {
		fmt.Printf("Task %d is already in '%s'\n", taskID, toColumnName)
	} else {
		fmt.Printf("Task %d moved to '%s'\n", taskID, toColumnName)
	}
	return nil
}

// findNextColumnName finds the name of the next column in the linked list
func findNextColumnName(columns []*models.Column, currentColumnID int) string {
	for _, col := range columns {
		if col.ID == currentColumnID && col.NextID != nil {
			for _, nextCol := range columns {
				if nextCol.ID == *col.NextID {
					return nextCol.Name
				}
			}
		}
	}
	return "Unknown"
}

// findPrevColumnName finds the name of the previous column in the linked list
func findPrevColumnName(columns []*models.Column, currentColumnID int) string {
	for _, col := range columns {
		if col.ID == currentColumnID && col.PrevID != nil {
			for _, prevCol := range columns {
				if prevCol.ID == *col.PrevID {
					return prevCol.Name
				}
			}
		}
	}
	return "Unknown"
}
