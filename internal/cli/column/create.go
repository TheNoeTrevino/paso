// Package column holds all cli commands related to columns
//
// e.g., paso column ...
package column

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
	columnservice "github.com/thenoetrevino/paso/internal/services/column"
)

// CreateCmd returns the column create subcommand
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new column",
		Long: `Create a new column in a project.

Special Column Flags:
  Columns can be marked with special properties to control task workflow:
  
  --ready       Mark this column as holding ready tasks. When tasks are marked
                as ready, they will automatically move to this column.
  
  --completed   Mark this column as holding completed tasks. When tasks are 
                marked as done, they will automatically move to this column.
  
  --in-progress Mark this column as holding in-progress tasks. When tasks are
                started, they will automatically move to this column.
  
  Note: Only one column per project can have each special property. Setting a 
  special property on a new column will automatically remove it from other columns.

Examples:
  # Create column using shorthand flags
  paso column create -n "Review" -p 1

  # Create column with special property
  paso column create -n "Done" -p 1 -c

  # Create column after specific column
  paso column create -n "Done" -p 1 -a 3

  # Mark as ready or completed column
  paso column create -n "Backlog" -p 1 -r
  paso column create -n "Done" -p 1 -c

  # JSON output for agents
  paso column create -n "Review" -p 1 -j

  # Quiet mode for bash capture
  COLUMN_ID=$(paso column create -n "Review" -p 1 -q)

  # Long-form flags also supported
  paso column create --name="Review" --project=1 --json
`,
		RunE: runCreate,
	}

	// Required flags
	cmd.Flags().StringP("name", "n", "", "Column name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	// Optional flags
	cmd.Flags().IntP("after", "a", 0, "Insert after column ID (0 = append to end)")
	cmd.Flags().BoolP("ready", "r", false, "Mark this column as holding ready tasks")
	cmd.Flags().BoolP("completed", "c", false, "Mark this column as holding completed tasks")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	columnName, _ := cmd.Flags().GetString("name")
	columnAfter, _ := cmd.Flags().GetInt("after")
	holdsReady, _ := cmd.Flags().GetBool("ready")
	holdsCompleted, _ := cmd.Flags().GetBool("completed")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI first
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Get project ID from flag or git branch
	columnProject, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		return formatter.ErrorWithSuggestion(cli.ExitUsage, "NO_PROJECT",
			err.Error(),
			"Use --project flag or create a project associated with this git branch")
	}

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, columnProject)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", columnProject))
	}

	// Validate after column if specified
	var afterID *int
	if columnAfter > 0 {
		afterCol, err := cliInstance.App.ColumnService.GetColumnByID(ctx, columnAfter)
		if err != nil {
			return formatter.Error(cli.ExitNotFound, "COLUMN_NOT_FOUND", fmt.Sprintf("column %d not found", columnAfter))
		}
		// Verify column belongs to same project
		if afterCol.ProjectID != columnProject {
			return formatter.Error(cli.ExitValidation, "INVALID_COLUMN", fmt.Sprintf("column %d does not belong to project %d", columnAfter, columnProject))
		}
		afterID = &columnAfter
	}

	// Create column
	column, err := cliInstance.App.ColumnService.CreateColumn(ctx, columnservice.CreateColumnRequest{
		Name:                columnName,
		ProjectID:           columnProject,
		AfterID:             afterID,
		HoldsReadyTasks:     holdsReady,
		HoldsCompletedTasks: holdsCompleted,
	})
	if err != nil {
		// Check for specific error about completed column already existing
		if strings.Contains(err.Error(), "completed column already exists") {
			return formatter.Error(cli.ExitValidation, "COMPLETED_COLUMN_EXISTS",
				fmt.Sprintf("%s\n\nUse the --force flag to change the done column.\nPaso uses the done column to move tasks with the {complete task command}.\nThis could lead to unexpected behavior, and this is not suggested.", err.Error()))
		}
		return formatter.Error(cli.ExitError, "COLUMN_CREATE_ERROR", err.Error())
	}

	// Output based on mode
	if quietMode {
		fmt.Printf("%d\n", column.ID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"column": map[string]any{
				"id":         column.ID,
				"name":       column.Name,
				"project_id": column.ProjectID,
			},
		})
	}

	// Human-readable output
	details := []styles.Detail{
		{Key: "ID", Value: strconv.Itoa(column.ID)},
		{Key: "Name", Value: columnName},
		{Key: "Project", Value: project.Name},
	}
	fmt.Print(styles.RenderSuccessWithDetails("Column created successfully", details, cli.GetColorScheme()))
	return nil
}
