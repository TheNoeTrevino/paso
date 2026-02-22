package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// InProgressCmd returns the task in-progress subcommand
func InProgressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "in-progress [<task_id>]",
		Short: "Move a task to in-progress or list in-progress tasks",
		Long: `Move a task to the column designated as holding in-progress tasks,
or list all in-progress tasks for a project.

The in-progress column is marked with holds_in_progress_tasks = true.
Use 'paso column update -i <column_id> -I' to designate an in-progress column.

Note: When moving tasks, this command uses the column marked as in-progress 
(see: paso column update --in-progress).

Examples:
  # Move task to in-progress column
  paso task in-progress 42

  # List all in-progress tasks for a project (shorthand)
  paso task in-progress -p 1

  # JSON output for agents
  paso task in-progress -p 1 -j

  # Quiet mode for bash capture
  paso task in-progress 42 -q

  # Long-form flags also supported
  paso task in-progress --project=1 --json
`,
		RunE: runInProgress,
		Args: cobra.MaximumNArgs(1),
	}

	// Flags
	cmd.Flags().IntP("project", "p", 0, "Project ID (for listing in-progress tasks)")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runInProgress(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	projectID, _ := cmd.Flags().GetInt("project")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Parse input
	input, err := ParseInProgressArgs(args, projectID)
	if err != nil {
		return formatter.Error(cli.ExitValidation, "INVALID_INPUT", err.Error())
	}

	// Execute based on mode
	if input.Mode == InProgressModeList {
		return listInProgressTasks(ctx, input.ProjectID, formatter)
	}

	return moveTaskToInProgress(ctx, input.TaskID, formatter)
}

func listInProgressTasks(ctx context.Context, projectID int, formatter *cli.OutputFormatter) error {
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

	// Validate project exists
	_, err = cliInstance.App.ProjectService.GetProjectByID(ctx, projectID)
	if err != nil {
		return formatter.ErrorWithSuggestion(cli.ExitNotFound, "PROJECT_NOT_FOUND",
			fmt.Sprintf("project %d not found", projectID),
			"Use 'paso project list' to see available projects")
	}

	// Get in-progress tasks
	tasks, err := cliInstance.App.TaskService.GetInProgressTasksByProject(ctx, projectID)
	if err != nil {
		return formatter.Error(cli.ExitError, "TASK_FETCH_ERROR", err.Error())
	}

	// Convert TaskDetail to display format
	displayTasks := make([]TaskDisplay, len(tasks))
	for i, t := range tasks {
		displayTasks[i] = TaskDisplay{
			ID:                  t.ID,
			TaskNumber:        t.TaskNumber,
			Title:               t.Title,
			TypeDescription:     t.TypeDescription,
			PriorityDescription: t.PriorityDescription,
			PriorityColor:       t.PriorityColor,
			IsBlocked:           t.IsBlocked,
		}
	}

	result := &ListInProgressResult{
		Tasks: displayTasks,
		Count: len(displayTasks),
	}

	// Output in appropriate format
	if formatter.Quiet {
		fmt.Print(FormatListQuiet(result))
		return nil
	}

	if formatter.JSON {
		return json.NewEncoder(os.Stdout).Encode(FormatListJSON(result))
	}

	// Human-readable output
	fmt.Print(FormatListHuman(result))
	return nil
}

func moveTaskToInProgress(ctx context.Context, taskID int, formatter *cli.OutputFormatter) error {
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

	// Move task to in-progress column
	err = cliInstance.App.TaskService.MoveTaskToInProgressColumn(ctx, taskID)
	if err != nil {
		// Check for specific errors
		if err == taskservice.ErrTaskAlreadyInTargetColumn {
			// Write to stderr as per requirements
			fmt.Fprintf(os.Stderr, "Task %d is already in the in-progress column (%s)\n", taskID, currentColumnName)
			// Still exit successfully
			if formatter.Quiet {
				fmt.Printf("%d\n", taskID)
			}
			return nil
		}
		if strings.Contains(err.Error(), "no in-progress column configured") {
			return formatter.ErrorWithSuggestion(cli.ExitValidation, "NO_IN_PROGRESS_COLUMN",
				"no in-progress column configured for this project",
				"Use 'paso column update --id=<column_id> --in-progress' to designate an in-progress column")
		}
		return formatter.Error(cli.ExitError, "MOVE_ERROR", err.Error())
	}

	// Get updated task detail for output
	updatedTaskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitError, "TASK_FETCH_ERROR", err.Error())
	}

	toColumnName := updatedTaskDetail.ColumnName

	result := &MoveInProgressResult{
		TaskID:     taskID,
		FromColumn: currentColumnName,
		ToColumn:   toColumnName,
	}

	// Output success
	if formatter.Quiet {
		fmt.Print(FormatMoveQuiet(result))
		return nil
	}

	if formatter.JSON {
		return json.NewEncoder(os.Stdout).Encode(FormatMoveJSON(result))
	}

	// Human-readable output
	cli.PrintSuccess(FormatMoveHuman(result))
	return nil
}
