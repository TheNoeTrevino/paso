package label

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
		Long: `Delete a label by ID (requires confirmation unless --force/-f or --quiet/-q).

Examples:
  # Delete with confirmation (shorthand)
  paso label delete -i 1

  # Skip confirmation
  paso label delete -i 1 -f

  # Quiet mode (no confirmation)
  paso label delete -i 1 -q

  # Long-form flags also supported
  paso label delete --id=1 --force
`,
		RunE: runDelete,
	}

	// Required flags
	cmd.Flags().IntP("id", "i", 0, "Label ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}

	// Optional flags
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	labelID, _ := cmd.Flags().GetInt("id")
	force, _ := cmd.Flags().GetBool("force")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to closing CLI", "error", err)
		}
	}()

	// Get label details for confirmation
	label, err := cli.GetLabelByID(ctx, cliInstance, labelID)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_NOT_FOUND", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Printf("Delete label #%d: '%s'? (y/N): ", labelID, label.Name)
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			slog.Error("failed to reading user input", "error", err)
			fmt.Println("Cancelled (failed to read input)")
			return nil
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete the label
	if err := cliInstance.App.LabelService.DeleteLabel(ctx, labelID); err != nil {
		if fmtErr := formatter.Error("DELETE_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":  true,
			"label_id": labelID,
		})
	}

	fmt.Printf("✓ Label %d deleted successfully\n", labelID)
	return nil
}
