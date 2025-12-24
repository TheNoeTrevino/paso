package column

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// DeleteCmd returns the column delete subcommand
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a column",
		Long: `Delete a column by ID (requires confirmation unless --force or --quiet).

Warning: Deleting a column will move all tasks in that column to the project's first column.

Examples:
  # Delete with confirmation
  paso column delete --id=1

  # Skip confirmation
  paso column delete --id=1 --force

  # Quiet mode (no confirmation)
  paso column delete --id=1 --quiet
`,
		RunE: runDelete,
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Column ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional flags
	cmd.Flags().Bool("force", false, "Skip confirmation")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	columnID, _ := cmd.Flags().GetInt("id")
	force, _ := cmd.Flags().GetBool("force")
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

	// Get column details for confirmation
	column, err := cliInstance.Repo().GetColumnByID(ctx, columnID)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_NOT_FOUND", fmt.Sprintf("column %d not found", columnID)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Println("⚠ Warning: Deleting column will move all tasks to the project's first column")
		fmt.Printf("Delete column #%d: '%s'? (y/N): ", columnID, column.Name)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			log.Printf("Error reading user input: %v", err)
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete the column
	if err := cliInstance.Repo().DeleteColumn(ctx, columnID); err != nil {
		if fmtErr := formatter.Error("DELETE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success":   true,
			"column_id": columnID,
		})
	}

	fmt.Printf("✓ Column %d deleted successfully\n", columnID)
	return nil
}
