package column

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// UpdateCmd returns the column update subcommand
func UpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a column",
		Long: `Update a column's name.

Examples:
  # Update column name (human-readable output)
  paso column update --id=1 --name="Completed"

  # JSON output for agents
  paso column update --id=1 --name="Completed" --json

  # Quiet mode
  paso column update --id=1 --name="Completed" --quiet
`,
		RunE: runUpdate,
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Column ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional flags
	cmd.Flags().String("name", "", "New column name")
	cmd.Flags().Bool("ready", false, "Set this column as holding ready tasks")
	cmd.Flags().Bool("completed", false, "Set this column as holding completed tasks")
	cmd.Flags().Bool("in-progress", false, "Set this column as holding in-progress tasks")
	cmd.Flags().Bool("force", false, "Force setting completed column even if one already exists")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	columnID, _ := cmd.Flags().GetInt("id")
	columnName, _ := cmd.Flags().GetString("name")
	setReady, _ := cmd.Flags().GetBool("ready")
	setCompleted, _ := cmd.Flags().GetBool("completed")
	setInProgress, _ := cmd.Flags().GetBool("in-progress")
	force, _ := cmd.Flags().GetBool("force")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Validate at least one update flag is provided
	if columnName == "" && !setReady && !setCompleted && !setInProgress {
		if fmtErr := formatter.Error("INVALID_INPUT", "at least one of --name, --ready, --completed, or --in-progress must be provided"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return fmt.Errorf("at least one of --name, --ready, --completed, or --in-progress must be provided")
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

	// Validate column exists
	column, err := cliInstance.App.ColumnService.GetColumnByID(ctx, columnID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_NOT_FOUND", fmt.Sprintf("column %d not found", columnID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	oldName := column.Name
	updatedColumn := column

	// Update column name if provided
	if columnName != "" {
		if err := cliInstance.App.ColumnService.UpdateColumnName(ctx, columnID, columnName); err != nil {
			if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			return err
		}
	}

	// Update ready status if flag is set
	if setReady {
		updatedColumn, err = cliInstance.App.ColumnService.SetHoldsReadyTasks(ctx, columnID)
		if err != nil {
			if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			return err
		}
	}

	// Update completed status if flag is set
	if setCompleted {
		updatedColumn, err = cliInstance.App.ColumnService.SetHoldsCompletedTasks(ctx, columnID, force)
		if err != nil {
			// Check for specific error about completed column already existing
			if strings.Contains(err.Error(), "completed column already exists") {
				if fmtErr := formatter.Error("COMPLETED_COLUMN_EXISTS",
					fmt.Sprintf("%s\n\nUse the --force flag to change the done column.\nPaso uses the done column to move tasks with the {complete task command}.\nThis could lead to unexpected behavior, and this is not suggested.", err.Error())); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				os.Exit(cli.ExitValidation)
			}
			if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			return err
		}
	}

	// Update in-progress status if flag is set
	if setInProgress {
		updatedColumn, err = cliInstance.App.ColumnService.SetHoldsInProgressTasks(ctx, columnID)
		if err != nil {
			if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			return err
		}
	}

	// Output based on mode
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"column": map[string]interface{}{
				"id":                      columnID,
				"name":                    updatedColumn.Name,
				"old_name":                oldName,
				"holds_ready_tasks":       updatedColumn.HoldsReadyTasks,
				"holds_completed_tasks":   updatedColumn.HoldsCompletedTasks,
				"holds_in_progress_tasks": updatedColumn.HoldsInProgressTasks,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Column %d updated successfully\n", columnID)
	if columnName != "" && columnName != oldName {
		fmt.Printf("  Name: '%s' → '%s'\n", oldName, columnName)
	}
	if setReady {
		fmt.Printf("  Now holds ready tasks: %v\n", updatedColumn.HoldsReadyTasks)
	}
	if setCompleted {
		fmt.Printf("  Now holds completed tasks: %v\n", updatedColumn.HoldsCompletedTasks)
	}
	if setInProgress {
		fmt.Printf("  Now holds in-progress tasks: %v\n", updatedColumn.HoldsInProgressTasks)
	}
	return nil
}
