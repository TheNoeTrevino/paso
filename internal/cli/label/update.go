package label

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
  # Update both name and color
  paso label update --id=1 --name="critical-bug" --color="#FF0000"

  # Update only name (keeps existing color)
  paso label update --id=1 --name="critical-bug"

  # Update only color (keeps existing name)
  paso label update --id=1 --color="#FF0000"

  # JSON output
  paso label update --id=1 --name="urgent" --json
`,
		RunE: runUpdate,
	}

	cmd.Flags().Int("id", 0, "Label ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}
	cmd.Flags().String("name", "", "New label name")
	cmd.Flags().String("color", "", "New label color in hex format #RRGGBB")

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	labelID, _ := cmd.Flags().GetInt("id")
	labelName, _ := cmd.Flags().GetString("name")
	labelColor, _ := cmd.Flags().GetString("color")
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

	// Check if at least one update flag is provided
	nameProvided := cmd.Flags().Changed("name")
	colorProvided := cmd.Flags().Changed("color")

	if !nameProvided && !colorProvided {
		if fmtErr := formatter.Error("MISSING_FLAGS", "at least one of --name or --color must be provided"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

	// Get existing label to fetch current values
	currentLabel, err := cli.GetLabelByID(ctx, cliInstance, labelID)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_NOT_FOUND", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
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
				log.Printf("Error formatting error message: %v", fmtErr)
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
			"label": map[string]interface{}{
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
