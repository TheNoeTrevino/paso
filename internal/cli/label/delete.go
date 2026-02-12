package label

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// DeleteCmd returns the label delete subcommand
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a label",
		Long: `Delete a label by ID (requires confirmation unless --force/-f or --quiet/-q).

Examples:
  # Delete with confirmation
  paso label delete 1

  # Skip confirmation
  paso label delete 1 -f

  # Quiet mode (no confirmation)
  paso label delete 1 -q

  # Long-form flags also supported
  paso label delete 1 --force
`,
		Args: cobra.ExactArgs(1),
		RunE: runDelete,
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

	force, _ := cmd.Flags().GetBool("force")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Parse ID from positional argument
	labelID, err := strconv.Atoi(args[0])
	if err != nil {
		return formatter.Error(cli.ExitValidation, "INVALID_ID", fmt.Sprintf("invalid ID '%s': must be a number", args[0]))
	}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Get label details for confirmation
	label, err := cli.GetLabelByID(ctx, cliInstance, labelID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "LABEL_NOT_FOUND", err.Error())
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Printf("Delete label #%d: '%s'? (y/N): ", labelID, label.Name)
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

	// Delete the label
	if err := cliInstance.App.LabelService.DeleteLabel(ctx, labelID); err != nil {
		return formatter.Error(cli.ExitError, "DELETE_ERROR", err.Error())
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

	message := fmt.Sprintf("Label %d deleted successfully", labelID)
	cli.PrintSuccess(message)
	return nil
}
