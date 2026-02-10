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
)

// UpdateCmd returns the column update subcommand
func UpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a column",
		Long: `Update a column's name and special properties.

Special column flags control which columns are used by task movement commands:
  --ready: Sets this column for the 'paso task to-ready' command
  --in-progress: Sets this column for the 'paso task in-progress move' command
  --completed: Sets this column for the 'paso task done' command

Note: Only one column per project can have each special property. Setting a flag
on one column automatically removes it from any other column that had it.

Examples:
  # Update column name
  paso column update 1 -n "Completed"

  # Set a column as the ready column
  paso column update 2 -r

  # Set a column as the in-progress column
  paso column update 3 -I

  # Set a column as the completed column
  paso column update 4 -c

  # Force completed column override
  paso column update 3 -c -f

  # JSON output for agents
  paso column update 1 -n "Completed" -j

  # Long-form flags also supported
  paso column update 1 --name="Completed" --json
`,
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}

	// Optional flags
	cmd.Flags().StringP("name", "n", "", "New column name")
	cmd.Flags().BoolP("ready", "r", false, "Set this column as holding ready tasks")
	cmd.Flags().BoolP("completed", "c", false, "Set this column as holding completed tasks")
	cmd.Flags().BoolP("in-progress", "I", false, "Set this column as holding in-progress tasks")
	cmd.Flags().BoolP("force", "f", false, "Force setting completed column even if one already exists")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	columnName, _ := cmd.Flags().GetString("name")
	setReady, _ := cmd.Flags().GetBool("ready")
	setCompleted, _ := cmd.Flags().GetBool("completed")
	setInProgress, _ := cmd.Flags().GetBool("in-progress")
	force, _ := cmd.Flags().GetBool("force")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Parse ID from positional argument
	columnID, err := strconv.Atoi(args[0])
	if err != nil {
		if fmtErr := formatter.Error("INVALID_ID", fmt.Sprintf("invalid ID '%s': must be a number", args[0])); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitValidation)
	}

	// Validate at least one update flag is provided
	if columnName == "" && !setReady && !setCompleted && !setInProgress {
		if fmtErr := formatter.Error("INVALID_INPUT", "at least one of --name, --ready, --completed, or --in-progress must be provided"); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

	// Initialize CLI
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

	// Validate column exists
	column, err := cliInstance.App.ColumnService.GetColumnByID(ctx, columnID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_NOT_FOUND", fmt.Sprintf("column %d not found", columnID)); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	oldName := column.Name
	updatedColumn := column

	// Update column name if provided
	if columnName != "" {
		if err := cliInstance.App.ColumnService.UpdateColumnName(ctx, columnID, columnName); err != nil {
			if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitError)
		}
	}

	// Update ready status if flag is set
	if setReady {
		updatedColumn, err = cliInstance.App.ColumnService.SetHoldsReadyTasks(ctx, columnID)
		if err != nil {
			if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitError)
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
					slog.Error("failed to format error message", "error", fmtErr)
				}
				os.Exit(cli.ExitError)
			}
			if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitError)
		}
	}

	// Update in-progress status if flag is set
	if setInProgress {
		updatedColumn, err = cliInstance.App.ColumnService.SetHoldsInProgressTasks(ctx, columnID)
		if err != nil {
			if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitError)
		}
	}

	// Output based on mode
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"column": map[string]any{
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
	colors := cli.GetColorScheme()
	details := []styles.Detail{
		{Key: "ID", Value: strconv.Itoa(columnID)},
		{Key: "Name", Value: updatedColumn.Name},
	}
	if setReady {
		details = append(details, styles.Detail{Key: "Ready tasks", Value: fmt.Sprintf("%v", updatedColumn.HoldsReadyTasks)})
	}
	if setInProgress {
		details = append(details, styles.Detail{Key: "In-progress tasks", Value: fmt.Sprintf("%v", updatedColumn.HoldsInProgressTasks)})
	}
	if setCompleted {
		details = append(details, styles.Detail{Key: "Completed tasks", Value: fmt.Sprintf("%v", updatedColumn.HoldsCompletedTasks)})
	}
	fmt.Print(styles.RenderSuccessWithDetails("Column updated successfully", details, colors))
	return nil
}
