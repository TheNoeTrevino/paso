// Package label holds all cli commands related to labels
// e.g., paso label ...
package label

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/handler"
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
		RunE: handler.Command(&createHandler{}, parseCreateFlags),
	}

	// Required flags
	cmd.Flags().String("name", "", "Label name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}

	cmd.Flags().String("color", "", "Label color in hex format #RRGGBB (required)")
	if err := cmd.MarkFlagRequired("color"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}

	cmd.Flags().Int("project", 0, "Project ID (uses PASO_PROJECT env var if not specified)")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

// createHandler implements handler.Handler for label creation
type createHandler struct{}

// Execute implements the Handler interface
func (h *createHandler) Execute(ctx context.Context, args *handler.Arguments) (interface{}, error) {
	// Get flag values from arguments
	labelName := args.MustGetString("name")
	labelColor := args.MustGetString("color")

	// Get project ID from flag or environment variable
	cmd := args.GetCmd()
	labelProject, err := cli.GetProjectID(cmd)
	if err != nil {
		return nil, fmt.Errorf("no project specified: use --project flag or set with 'eval $(paso use project <project-id>)'")
	}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialization error: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to closing CLI", "error", err)
		}
	}()

	// Validate color format
	if err := cli.ValidateColorHex(labelColor); err != nil {
		return nil, fmt.Errorf("invalid color: %w", err)
	}

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, labelProject)
	if err != nil {
		return nil, fmt.Errorf("project %d not found", labelProject)
	}

	// Create label
	label, err := cliInstance.App.LabelService.CreateLabel(ctx, labelservice.CreateLabelRequest{
		ProjectID: labelProject,
		Name:      labelName,
		Color:     labelColor,
	})
	if err != nil {
		return nil, fmt.Errorf("label creation error: %w", err)
	}

	return &labelCreateResult{
		ID:        label.ID,
		Name:      label.Name,
		Color:     label.Color,
		ProjectID: label.ProjectID,
		Project:   project.Name,
	}, nil
}

// labelCreateResult represents the result of label creation
type labelCreateResult struct {
	ID        int
	Name      string
	Color     string
	ProjectID int
	Project   string
}

// GetID implements the GetID interface for quiet mode output
func (r *labelCreateResult) GetID() int {
	return r.ID
}

func parseCreateFlags(cmd *cobra.Command) error {
	// Validate required flags
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return fmt.Errorf("label name is required")
	}

	color, _ := cmd.Flags().GetString("color")
	if color == "" {
		return fmt.Errorf("label color is required")
	}

	return nil
}
