package label

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	labelservice "github.com/thenoetrevino/paso/internal/services/label"
)

// UpdateCmd returns the label update subcommand
func UpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a label",
		Long: `Update a label's name and/or color.

Examples:
  # Update both name and color
  paso label update 1 -n "critical-bug" -c "#FF0000"

  # Update only name (keeps existing color)
  paso label update 1 -n "critical-bug"

  # Update only color (keeps existing name)
  paso label update 1 -c "#FF0000"

  # JSON output
  paso label update 1 -n "urgent" -j

  # Long-form flags also supported
  paso label update 1 --name="urgent" --json
`,
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}

	cmd.Flags().StringP("name", "n", "", "New label name")
	cmd.Flags().StringP("color", "c", "", "New label color in hex format #RRGGBB")

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	labelName, _ := cmd.Flags().GetString("name")
	labelColor, _ := cmd.Flags().GetString("color")
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

	// Check if at least one update flag is provided
	nameProvided := cmd.Flags().Changed("name")
	colorProvided := cmd.Flags().Changed("color")

	if !nameProvided && !colorProvided {
		return formatter.Error(cli.ExitUsage, "MISSING_FLAGS", "at least one of --name or --color must be provided")
	}

	// Get existing label to fetch current values
	currentLabel, err := cli.GetLabelByID(ctx, cliInstance, labelID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "LABEL_NOT_FOUND", err.Error())
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
			return formatter.Error(cli.ExitValidation, "INVALID_COLOR", err.Error())
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
		return formatter.Error(cli.ExitError, "UPDATE_ERROR", err.Error())
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

	message := fmt.Sprintf("Label %d updated successfully", labelID)
	cli.PrintSuccess(message)
	return nil
}
