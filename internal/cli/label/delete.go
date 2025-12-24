package label

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

// DeleteCmd returns the label delete subcommand
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a label",
		Long: `Delete a label by ID (requires confirmation unless --force or --quiet).

Examples:
  # Delete with confirmation
  paso label delete --id=1

  # Skip confirmation
  paso label delete --id=1 --force

  # Quiet mode (no confirmation)
  paso label delete --id=1 --quiet
`,
		RunE: runDelete,
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Label ID (required)")
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

	labelID, _ := cmd.Flags().GetInt("id")
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

	// Get label details for confirmation
	label, err := cli.GetLabelByID(ctx, cliInstance, labelID)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_NOT_FOUND", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Printf("Delete label #%d: '%s'? (y/N): ", labelID, label.Name)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			log.Printf("Error reading user input: %v", err)
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete the label
	if err := cliInstance.Repo().DeleteLabel(ctx, labelID); err != nil {
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
			"success":  true,
			"label_id": labelID,
		})
	}

	fmt.Printf("✓ Label %d deleted successfully\n", labelID)
	return nil
}
