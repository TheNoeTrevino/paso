package project

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
)

// DeleteCmd returns the project delete subcommand
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project",
		Long: `Delete a project by ID (requires confirmation unless --force/-f or --quiet/-q).

Examples:
  # Delete with confirmation
  paso project delete 1

  # Skip confirmation
  paso project delete 1 -f

  # Long-form flags also supported
  paso project delete 1 --force
`,
		Args: cobra.ExactArgs(1),
		RunE: runDelete,
	}

	// Optional flags
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation")
	cmd.Flags().Bool("dry-run", false, "Show what would be deleted without actually deleting")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	force, _ := cmd.Flags().GetBool("force")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Parse ID from positional argument
	projectID, err := strconv.Atoi(args[0])
	if err != nil {
		return formatter.Error(cli.ExitValidation, "INVALID_ID", fmt.Sprintf("invalid ID '%s': must be a number", args[0]))
	}

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

	// Get project details for confirmation
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, projectID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", projectID))
	}

	// Fetch project statistics (columns and tasks)
	columnCount := 0
	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, projectID)
	if err != nil {
		slog.Error("failed to fetch columns for project", "error", err, "project_id", projectID)
	} else {
		columnCount = len(columns)
	}

	taskCount := 0
	tasksByColumn, err := cliInstance.App.TaskService.GetTaskSummariesByProject(ctx, projectID)
	if err != nil {
		slog.Error("failed to fetch tasks for project", "error", err, "project_id", projectID)
	} else {
		for _, tasks := range tasksByColumn {
			taskCount += len(tasks)
		}
	}

	// Handle dry-run mode
	if dryRun {
		if quietMode {
			return nil
		}

		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{
				"success":      true,
				"dry_run":      true,
				"project_id":   projectID,
				"project_name": project.Name,
				"columns":      columnCount,
				"tasks":        taskCount,
			})
		}

		fmt.Printf("Would delete project %d (%s) with %d column(s) and %d task(s)\n", projectID, project.Name, columnCount, taskCount)
		return nil
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Printf("Delete project #%d: %s? (%d columns, %d tasks) (y/N): ", projectID, project.Name, columnCount, taskCount)
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			slog.Error("failed to read user input", "error", err)
			fmt.Println("Cancelled (failed to read input)")
			return nil
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete the project
	if err := cliInstance.App.ProjectService.DeleteProject(ctx, projectID, force); err != nil {
		return formatter.Error(cli.ExitError, "DELETE_ERROR", err.Error())
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":    true,
			"project_id": projectID,
		})
	}

	colors := cli.GetColorScheme()
	message := fmt.Sprintf("Project %d deleted successfully", projectID)
	fmt.Print(styles.RenderSuccess(message, colors))
	return nil
}
