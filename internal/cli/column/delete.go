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

// DeleteCmd returns the column delete subcommand
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a column",
		Long: `Delete a column by ID (requires confirmation unless --force/-f or --quiet/-q).

Warning: Deleting a column will move all tasks in that column to the project's first column.

Examples:
  # Delete with confirmation
  paso column delete 1

  # Skip confirmation
  paso column delete 1 -f

  # Quiet mode (no confirmation)
  paso column delete 1 -q

  # Long-form flags also supported
  paso column delete 1 --force
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
	columnID, err := strconv.Atoi(args[0])
	if err != nil {
		if fmtErr := formatter.Error("INVALID_ID", fmt.Sprintf("invalid ID '%s': must be a number", args[0])); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitValidation)
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

	// Get column details for confirmation
	column, err := cliInstance.App.ColumnService.GetColumnByID(ctx, columnID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_NOT_FOUND", fmt.Sprintf("column %d not found", columnID)); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Fetch task count in the column
	taskCount := 0
	tasksByColumn, err := cliInstance.App.TaskService.GetTaskSummariesByProject(ctx, column.ProjectID)
	if err != nil {
		slog.Error("failed to fetch tasks for column", "error", err, "column_id", columnID)
	} else {
		if tasks, ok := tasksByColumn[columnID]; ok {
			taskCount = len(tasks)
		}
	}

	// Handle dry-run mode
	if dryRun {
		if quietMode {
			return nil
		}

		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{
				"success":        true,
				"dry_run":        true,
				"column_id":      columnID,
				"column_name":    column.Name,
				"tasks_affected": taskCount,
			})
		}

		fmt.Printf("Would delete column %d (%s) and move %d task(s) to first column\n", columnID, column.Name, taskCount)
		return nil
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Printf("Warning: Deleting column will move %d task(s) to the project's first column\n", taskCount)
		fmt.Printf("Delete column #%d: '%s'? (y/N): ", columnID, column.Name)
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

	// Delete the column
	if err := cliInstance.App.ColumnService.DeleteColumn(ctx, columnID); err != nil {
		if fmtErr := formatter.Error("DELETE_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":   true,
			"column_id": columnID,
		})
	}

	colors := cli.GetColorScheme()
	message := fmt.Sprintf("Column %d deleted successfully", columnID)
	fmt.Print(styles.RenderSuccess(message, colors))
	return nil
}
