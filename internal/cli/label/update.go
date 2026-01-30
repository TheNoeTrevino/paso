package label

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	labelservice "github.com/thenoetrevino/paso/internal/services/label"
)

// UpdateCmd returns the label update subcommand
func UpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a label",
		Long: `Update a label's name and/or color.

Examples:
  # Update both name and color (shorthand)
  paso label update -i 1 -n "critical-bug" -c "#FF0000"

  # Update only name (keeps existing color)
  paso label update -i 1 -n "critical-bug"

  # Update only color (keeps existing name)
  paso label update -i 1 -c "#FF0000"

  # JSON output
  paso label update -i 1 -n "urgent" -j

  # Long-form flags also supported
  paso label update --id=1 --name="urgent" --json
`,
		RunE: runUpdate,
	}

	cmd.Flags().IntP("id", "i", 0, "Label ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}
	cmd.Flags().StringP("name", "n", "", "New label name")
	cmd.Flags().StringP("color", "c", "", "New label color in hex format #RRGGBB")

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	labelID, _ := cmd.Flags().GetInt("id")
	labelName, _ := cmd.Flags().GetString("name")
	labelColor, _ := cmd.Flags().GetString("color")
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

	// Check if at least one update flag is provided
	nameProvided := cmd.Flags().Changed("name")
	colorProvided := cmd.Flags().Changed("color")

	if !nameProvided && !colorProvided {
		if fmtErr := formatter.Error("MISSING_FLAGS", "at least one of --name or --color must be provided"); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

	// Get existing label to fetch current values
	currentLabel, err := cli.GetLabelByID(ctx, cliInstance, labelID)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_NOT_FOUND", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Use existing values if not provided
	newName := currentLabel.Name
	if nameProvided {
		newName = labelName
	}

	newColor := currentLabel.Color
	if colorProvided {
		// Validate color format
		if err := cli.ValidateColorHex(labelColor); err != nil {
			if fmtErr := formatter.Error("INVALID_COLOR", err.Error()); fmtErr != nil {
				slog.Error("failed to formatting error message", "error", fmtErr)
			}
			os.Exit(cli.ExitValidation)
		}
		newColor = labelColor
	}

	// Update label
	req := labelservice.UpdateLabelRequest{
		ID: labelID,
	}
	if nameProvided {
		req.Name = &newName
	}
	if colorProvided {
		req.Color = &newColor
	}

	if err := cliInstance.App.LabelService.UpdateLabel(ctx, req); err != nil {
		if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"label": map[string]any{
				"id":        labelID,
				"name":      newName,
				"color":     newColor,
				"old_name":  currentLabel.Name,
				"old_color": currentLabel.Color,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Label %d updated successfully\n", labelID)
	if nameProvided {
		fmt.Printf("  Name: '%s' → '%s'\n", currentLabel.Name, newName)
	}
	if colorProvided {
		fmt.Printf("  Color: %s → %s\n", currentLabel.Color, newColor)
	}
	return nil
}
