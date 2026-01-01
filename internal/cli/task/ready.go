package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
  # Human-readable output
  paso task ready --project=1

  # JSON output for agents
  paso task ready --project=1 --json

  # Quiet mode for bash capture
  TASK_IDS=$(paso task ready --project=1 --quiet)
`,
		RunE: runReady,
	}

	// Flags
	cmd.Flags().Int("project", 0, "Project ID (uses PASO_PROJECT env var if not specified)")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runReady(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Get project ID from flag or environment variable
	taskProject, err := cli.GetProjectID(cmd)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("NO_PROJECT",
			err.Error(),
			"Set project with: eval $(paso use project <project-id>)"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

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

	// Validate project exists
	_, err = cliInstance.App.ProjectService.GetProjectByID(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("PROJECT_NOT_FOUND",
			fmt.Sprintf("project %d not found", taskProject),
			"Use 'paso project list' to see available projects"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get ready tasks (tasks in ready columns and not blocked)
	var readyTasks []*models.TaskSummary
	readyTasks, err = cliInstance.App.TaskService.GetReadyTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
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
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
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
