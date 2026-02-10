package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/models"
)

// BlockedCmd returns the task blocked subcommand
func BlockedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocked",
		Short: "List blocked tasks",
		Long: `List all tasks that are blocked by dependencies.

These are tasks that cannot be started until their blocking
dependencies are completed.

Examples:
  # Human-readable output (shorthand)
  paso task blocked -p 1

  # JSON output for agents
  paso task blocked -p 1 -j

  # Quiet mode for bash capture
  TASK_IDS=$(paso task blocked -p 1 -q)

  # Long-form flags also supported
  paso task blocked --project=1 --json
`,
		RunE: runBlocked,
	}

	// Flags
	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runBlocked(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI first
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

	// Get project ID from flag or git branch
	taskProject, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("NO_PROJECT",
			err.Error(),
			"Use --project flag or create a project associated with this git branch"); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

	// Validate project exists
	_, err = cliInstance.App.ProjectService.GetProjectByID(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("PROJECT_NOT_FOUND",
			fmt.Sprintf("project %d not found", taskProject),
			"Use 'paso project list' to see available projects"); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get all tasks for project (includes IsBlocked field)
	tasksByColumn, err := cliInstance.App.TaskService.GetTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}

	// Filter for blocked tasks (IsBlocked == true)
	var blockedTasks []*models.TaskSummary
	for _, columnTasks := range tasksByColumn {
		for _, task := range columnTasks {
			if task.IsBlocked {
				blockedTasks = append(blockedTasks, task)
			}
		}
	}

	// Output in appropriate format
	if quietMode {
		// Just print IDs
		for _, t := range blockedTasks {
			fmt.Printf("%d\n", t.ID)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"tasks":   blockedTasks,
			"count":   len(blockedTasks),
		})
	}

	// Human-readable output
	if len(blockedTasks) == 0 {
		fmt.Println("No blocked tasks found")
		return nil
	}

	fmt.Printf("Found %d blocked tasks:\n\n", len(blockedTasks))
	for _, t := range blockedTasks {
		// Include priority if set
		priorityInfo := ""
		if t.PriorityDescription != "" && t.PriorityDescription != "medium" {
			priorityInfo = fmt.Sprintf(" [%s]", t.PriorityDescription)
		}
		fmt.Printf("  [%d] %s%s (BLOCKED)\n", t.ID, t.Title, priorityInfo)
	}

	return nil
}
