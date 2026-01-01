package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// DoneCmd returns the task done subcommand
func DoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <task_id>",
		Short: "Move a task to the completed column",
		Long: `Move a task to the column designated as holding completed tasks.

The completed column is marked with holds_completed_tasks = true.
Use 'paso column update --id=<column_id> --completed' to designate a completed column.

Examples:
  # Move task to completed column
  paso task done 42

  # JSON output for agents
  paso task done 42 --json

  # Quiet mode for bash capture
  paso task done 42 --quiet
`,
		RunE: runDone,
		Args: cobra.ExactArgs(1),
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runDone(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse task ID from positional argument
	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

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

	// Get task detail before move for output
	taskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
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
			if fmtErr := formatter.ErrorWithSuggestion("NO_COMPLETED_COLUMN",
				"no completed column configured for this project",
				"Use 'paso column update --id=<column_id> --completed' to designate a completed column"); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			os.Exit(cli.ExitValidation)
		}
		if fmtErr := formatter.Error("MOVE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Get updated task detail for output
	updatedTaskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	toColumnName := updatedTaskDetail.ColumnName

	// Output success
	if quietMode {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success":     true,
			"task_id":     taskID,
			"from_column": currentColumnName,
			"to_column":   toColumnName,
		})
	}

	// Human-readable output
	fmt.Printf("Task %d moved to '%s'\n", taskID, toColumnName)
	return nil
}
