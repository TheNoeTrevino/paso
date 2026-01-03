package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
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
Use 'paso column update --id=<column_id> --in-progress' to designate an in-progress column.

Examples:
  # Move task to in-progress column
  paso task in-progress 42

  # List all in-progress tasks for a project
  paso task in-progress --project=1

  # JSON output for agents
  paso task in-progress --project=1 --json

  # Quiet mode for bash capture
  paso task in-progress 42 --quiet
`,
		RunE: runInProgress,
		Args: cobra.MaximumNArgs(1),
	}

	// Flags
	cmd.Flags().Int("project", 0, "Project ID (for listing in-progress tasks)")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runInProgress(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	projectID, _ := cmd.Flags().GetInt("project")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Determine mode: list or move
	if projectID > 0 {
		// List mode
		return listInProgressTasks(ctx, projectID, formatter)
	}

	// Move mode - require task ID
	if len(args) == 0 {
		if fmtErr := formatter.Error("INVALID_INPUT", "either provide a task ID or use --project flag to list tasks"); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return fmt.Errorf("either provide a task ID or use --project flag to list tasks")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	return moveTaskToInProgress(ctx, taskID, formatter)
}

func listInProgressTasks(ctx context.Context, projectID int, formatter *cli.OutputFormatter) error {
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

	// Validate project exists
	_, err = cliInstance.App.ProjectService.GetProjectByID(ctx, projectID)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("PROJECT_NOT_FOUND",
			fmt.Sprintf("project %d not found", projectID),
			"Use 'paso project list' to see available projects"); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get in-progress tasks
	tasks, err := cliInstance.App.TaskService.GetInProgressTasksByProject(ctx, projectID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	// Convert TaskDetail to simpler format for display
	type TaskDisplay struct {
		ID                  int    `json:"id"`
		TicketNumber        int    `json:"ticket_number"`
		Title               string `json:"title"`
		TypeDescription     string `json:"type_description"`
		PriorityDescription string `json:"priority_description"`
		PriorityColor       string `json:"priority_color"`
		IsBlocked           bool   `json:"is_blocked"`
	}

	displayTasks := make([]TaskDisplay, len(tasks))
	for i, t := range tasks {
		displayTasks[i] = TaskDisplay{
			ID:                  t.ID,
			TicketNumber:        t.TicketNumber,
			Title:               t.Title,
			TypeDescription:     t.TypeDescription,
			PriorityDescription: t.PriorityDescription,
			PriorityColor:       t.PriorityColor,
			IsBlocked:           t.IsBlocked,
		}
	}

	// Output in appropriate format
	if formatter.Quiet {
		// Just print IDs
		for _, t := range displayTasks {
			fmt.Printf("%d\n", t.ID)
		}
		return nil
	}

	if formatter.JSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"tasks":   displayTasks,
			"count":   len(displayTasks),
		})
	}

	// Human-readable output
	if len(displayTasks) == 0 {
		fmt.Println("No in-progress tasks found")
		return nil
	}

	fmt.Printf("Found %d in-progress tasks:\n\n", len(displayTasks))
	for _, t := range displayTasks {
		// Include priority if set
		priorityInfo := ""
		if t.PriorityDescription != "" && t.PriorityDescription != "medium" {
			priorityInfo = fmt.Sprintf(" [%s]", t.PriorityDescription)
		}

		// Include blocked indicator
		blockedInfo := ""
		if t.IsBlocked {
			blockedInfo = " ⚠️ BLOCKED"
		}

		fmt.Printf("  [%d] %s%s%s\n", t.ID, t.Title, priorityInfo, blockedInfo)
	}

	return nil
}

func moveTaskToInProgress(ctx context.Context, taskID int, formatter *cli.OutputFormatter) error {
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

	// Get task detail before move for output
	taskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
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
			if fmtErr := formatter.ErrorWithSuggestion("NO_IN_PROGRESS_COLUMN",
				"no in-progress column configured for this project",
				"Use 'paso column update --id=<column_id> --in-progress' to designate an in-progress column"); fmtErr != nil {
				slog.Error("failed to formatting error message", "error", fmtErr)
			}
			os.Exit(cli.ExitValidation)
		}
		if fmtErr := formatter.Error("MOVE_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	// Get updated task detail for output
	updatedTaskDetail, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	toColumnName := updatedTaskDetail.ColumnName

	// Output success
	if formatter.Quiet {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if formatter.JSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
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
