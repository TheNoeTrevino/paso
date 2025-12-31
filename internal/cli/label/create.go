// Package label holds all cli commands related to labels
// e.g., paso label ...
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

// CreateCmd returns the label create subcommand
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new label",
		Long: `Create a new label with a name and color.

Examples:
  # Create label (human-readable output)
  paso label create --name="bug" --color="#FF0000" --project=1

  # JSON output for agents
  paso label create --name="bug" --color="#FF0000" --project=1 --json

  # Quiet mode for bash capture
  LABEL_ID=$(paso label create --name="bug" --color="#FF0000" --project=1 --quiet)
`,
		RunE: runCreate,
	}

	// Required flags
	cmd.Flags().String("name", "", "Label name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().String("color", "", "Label color in hex format #RRGGBB (required)")
	if err := cmd.MarkFlagRequired("color"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Int("project", 0, "Project ID (uses PASO_PROJECT env var if not specified)")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get flag values
	labelName, _ := cmd.Flags().GetString("name")
	labelColor, _ := cmd.Flags().GetString("color")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Get project ID from flag or environment variable
	labelProject, err := cli.GetProjectID(cmd)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("NO_PROJECT",
			err.Error(),
			"Set project with: eval $(paso use project <project-id>)"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

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

	// Validate color format
	if err := cli.ValidateColorHex(labelColor); err != nil {
		if fmtErr := formatter.Error("INVALID_COLOR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitValidation)
	}

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", labelProject)); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Create label
	label, err := cliInstance.App.LabelService.CreateLabel(ctx, labelservice.CreateLabelRequest{
		ProjectID: labelProject,
		Name:      labelName,
		Color:     labelColor,
	})
	if err != nil {
		if fmtErr := formatter.Error("LABEL_CREATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		fmt.Printf("%d\n", label.ID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"label": map[string]interface{}{
				"id":         label.ID,
				"name":       label.Name,
				"color":      label.Color,
				"project_id": label.ProjectID,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Label '%s' created successfully (ID: %d)\n", labelName, label.ID)
	fmt.Printf("  Project: %s\n", project.Name)
	fmt.Printf("  Color: %s\n", labelColor)
	return nil
}
