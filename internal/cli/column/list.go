package column

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// ListCmd returns the column list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List columns in a project",
		Long: `List all columns in a project (in order).

Examples:
  # Human-readable list
  paso column list --project=1

  # JSON output for agents
  paso column list --project=1 --json

  # Quiet mode (one ID per line)
  paso column list --project=1 --quiet
`,
		RunE: runList,
	}

	// Flags
	cmd.Flags().Int("project", 0, "Project ID (uses PASO_PROJECT env var if not specified)")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (IDs only)")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Get project ID from flag or environment variable
	columnProject, err := cli.GetProjectID(cmd)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("NO_PROJECT",
			err.Error(),
			"Set project with: eval $(paso use project <project-id>)"); fmtErr != nil {
			slog.Error("Error formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("Error formatting error message", "error", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("Error closing CLI", "error", err)
		}
	}()

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, columnProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", columnProject)); fmtErr != nil {
			slog.Error("Error formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get columns
	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, columnProject)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("Error formatting error message", "error", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		for _, col := range columns {
			fmt.Printf("%d\n", col.ID)
		}
		return nil
	}

	if jsonOutput {
		columnList := make([]map[string]interface{}, len(columns))
		for i, col := range columns {
			columnList[i] = map[string]interface{}{
				"id":                    col.ID,
				"name":                  col.Name,
				"project_id":            col.ProjectID,
				"holds_ready_tasks":     col.HoldsReadyTasks,
				"holds_completed_tasks": col.HoldsCompletedTasks,
			}
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"columns": columnList,
		})
	}

	// Human-readable output
	if len(columns) == 0 {
		fmt.Printf("No columns found in project '%s'\n", project.Name)
		return nil
	}

	fmt.Printf("Columns in project '%s':\n", project.Name)
	for i, col := range columns {
		flags := ""
		if col.HoldsReadyTasks {
			flags += " [READY]"
		}
		if col.HoldsCompletedTasks {
			flags += " [COMPLETED]"
		}
		fmt.Printf("  %d. %s%s (ID: %d)\n", i+1, col.Name, flags, col.ID)
	}
	return nil
}
