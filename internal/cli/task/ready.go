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

// ReadyCmd returns the task ready subcommand
func ReadyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ready",
		Short: "List tasks ready to work on",
		Long: `List all tasks that have no blocking dependencies.

These are tasks that can be started immediately as they are not
waiting on any other tasks to be completed.

Examples:
  # Human-readable output (shorthand)
  paso task ready -p 1

  # JSON output for agents
  paso task ready -p 1 -j

  # Quiet mode for bash capture
  TASK_IDS=$(paso task ready -p 1 -q)

  # Long-form flags also supported
  paso task ready --project=1 --json
`,
		RunE: runReady,
	}

	// Flags
	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runReady(cmd *cobra.Command, args []string) error {
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

	// Get ready tasks (tasks in ready columns and not blocked)
	var readyTasks []*models.TaskSummary
	readyTasks, err = cliInstance.App.TaskService.GetReadyTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}

	// Output in appropriate format
	if quietMode {
		// Just print IDs
		for _, t := range readyTasks {
			fmt.Printf("%d\n", t.ID)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"tasks":   readyTasks,
			"count":   len(readyTasks),
		})
	}

	// Human-readable output
	if len(readyTasks) == 0 {
		fmt.Println("No ready tasks found")
		return nil
	}

	fmt.Printf("Found %d ready tasks:\n\n", len(readyTasks))
	for _, t := range readyTasks {
		// Include priority if set
		priorityInfo := ""
		if t.PriorityDescription != "" && t.PriorityDescription != "medium" {
			priorityInfo = fmt.Sprintf(" [%s]", t.PriorityDescription)
		}
		fmt.Printf("  [%d] %s%s\n", t.ID, t.Title, priorityInfo)
	}

	return nil
}
