package column

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

	cmd.Flags().String("name", "", "New column name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	columnID, _ := cmd.Flags().GetInt("id")
	columnName, _ := cmd.Flags().GetString("name")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cliInstance, err := cli.NewCLI(ctx)
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

	// Update column
	if err := cliInstance.App.ColumnService.UpdateColumnName(ctx, columnID, columnName); err != nil {
		if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"column": map[string]interface{}{
				"id":       columnID,
				"name":     columnName,
				"old_name": oldName,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Column %d updated successfully\n", columnID)
	fmt.Printf("  '%s' → '%s'\n", oldName, columnName)
	return nil
}
