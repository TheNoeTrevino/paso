// Package column holds all cli commands related to columns
//
// e.g., paso column ...
package column

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	columnservice "github.com/thenoetrevino/paso/internal/services/column"
)

// CreateCmd returns the column create subcommand
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new column",
		Long: `Create a new column in a project.

Examples:
  # Create column at end (human-readable output)
  paso column create --name="Review" --project=1

  # JSON output for agents
  paso column create --name="Review" --project=1 --json

  # Quiet mode for bash capture
  COLUMN_ID=$(paso column create --name="Review" --project=1 --quiet)

  # Create column after specific column
  paso column create --name="Done" --project=1 --after=3
`,
		RunE: runCreate,
	}

	// Required flags
	cmd.Flags().String("name", "", "Column name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Int("project", 0, "Project ID (uses PASO_PROJECT env var if not specified)")

	// Optional flags
	cmd.Flags().Int("after", 0, "Insert after column ID (0 = append to end)")
	cmd.Flags().Bool("ready", false, "Mark this column as holding ready tasks")
	cmd.Flags().Bool("completed", false, "Mark this column as holding completed tasks")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

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

	// Get project ID from flag or environment variable
	columnProject, err := cli.GetProjectID(cmd)
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
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, columnProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", columnProject)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Validate after column if specified
	var afterID *int
	if columnAfter > 0 {
		afterCol, err := cliInstance.App.ColumnService.GetColumnByID(ctx, columnAfter)
		if err != nil {
			if fmtErr := formatter.Error("COLUMN_NOT_FOUND", fmt.Sprintf("column %d not found", columnAfter)); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			os.Exit(cli.ExitNotFound)
		}
		// Verify column belongs to same project
		if afterCol.ProjectID != columnProject {
			if fmtErr := formatter.Error("INVALID_COLUMN", fmt.Sprintf("column %d does not belong to project %d", columnAfter, columnProject)); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			os.Exit(cli.ExitValidation)
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
			if fmtErr := formatter.Error("COMPLETED_COLUMN_EXISTS",
				fmt.Sprintf("%s\n\nUse the --force flag to change the done column.\nPaso uses the done column to move tasks with the {complete task command}.\nThis could lead to unexpected behavior, and this is not suggested.", err.Error())); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			os.Exit(cli.ExitValidation)
		}
		if fmtErr := formatter.Error("COLUMN_CREATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		fmt.Printf("%d\n", column.ID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"column": map[string]interface{}{
				"id":         column.ID,
				"name":       column.Name,
				"project_id": column.ProjectID,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Column '%s' created successfully (ID: %d)\n", columnName, column.ID)
	fmt.Printf("  Project: %s\n", project.Name)
	return nil
}
