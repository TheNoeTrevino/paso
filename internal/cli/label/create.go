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
  # Create label using shorthand flags
  paso label create -n "bug" -c "#FF0000" -p 1

  # JSON output for agents
  paso label create -n "bug" -c "#FF0000" -p 1 -j

  # Quiet mode for bash capture
  LABEL_ID=$(paso label create -n "bug" -c "#FF0000" -p 1 -q)

  # Long-form flags also supported
  paso label create --name="bug" --color="#FF0000" --project=1
`,
		RunE: handler.Command(&createHandler{}, parseCreateFlags),
	}

	// Required flags
	cmd.Flags().StringP("name", "n", "", "Label name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}

	cmd.Flags().StringP("color", "c", "", "Label color in hex format #RRGGBB (required)")
	if err := cmd.MarkFlagRequired("color"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}

	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

// createHandler implements handler.Handler for label creation
type createHandler struct{}

// Execute implements the Handler interface
func (h *createHandler) Execute(ctx context.Context, args *handler.Arguments) (any, error) {
	labelName := args.MustGetString("name")
	labelColor := args.MustGetString("color")
	if err := cli.ValidateColorHex(labelColor); err != nil {
		return nil, fmt.Errorf("invalid color: %w", err)
	}

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialization error: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to closing CLI", "error", err)
		}
	}()

	// Get project ID from flag or git branch
	cmd := args.GetCmd()
	labelProject, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		return nil, fmt.Errorf("no project specified: use --project flag or create a project associated with this branch")
	}

	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, labelProject)
	if err != nil {
		return nil, fmt.Errorf("project %d not found", labelProject)
	}

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
