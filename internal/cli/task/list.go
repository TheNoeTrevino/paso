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

// ListCmd returns the task list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long:  "List all tasks in a project.",
		RunE:  runList,
	}

	// Flags
	cmd.Flags().Int("project", 0, "Project ID (uses PASO_PROJECT env var if not specified)")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Get tasks (returns map[columnID][]*TaskSummary)
	tasksByColumn, err := cliInstance.App.TaskService.GetTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Flatten tasks from all columns
	var allTasks []*models.TaskSummary
	for _, columnTasks := range tasksByColumn {
		allTasks = append(allTasks, columnTasks...)
	}

	// Output in appropriate format
	if quietMode {
		// Just print IDs
		for _, t := range allTasks {
			fmt.Printf("%d\n", t.ID)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"tasks":   allTasks,
		})
	}

	// Human-readable output
	if len(allTasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	fmt.Printf("Found %d tasks:\n\n", len(allTasks))
	for _, t := range allTasks {
		fmt.Printf("  [%d] %s\n", t.ID, t.Title)
	}

	return nil
}
