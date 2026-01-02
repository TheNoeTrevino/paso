// Package project holds all cli commands related to projects
//
// e.g., paso project ...
package project

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/handler"
	projectservice "github.com/thenoetrevino/paso/internal/services/project"
)

// CreateCmd returns the project create subcommand
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Create a new project with specified attributes.

Examples:
  # Simple project (human-readable output)
  paso project create --title="Backend API"

  # JSON output for agents
  paso project create --title="Backend API" --json

  # Quiet mode for bash capture
  PROJECT_ID=$(paso project create --title="Backend API" --quiet)

  # With description
  paso project create \
    --title="Backend API" \
    --description="REST API for mobile app"
`,
		RunE: handler.Command(&createHandler{}, parseCreateFlags),
	}

	// Required flags
	cmd.Flags().String("title", "", "Project title (required)")
	if err := cmd.MarkFlagRequired("title"); err != nil {
		slog.Error("Error marking flag as required", "error", err)
	}

	// Optional flags
	cmd.Flags().String("description", "", "Project description")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

// createHandler implements handler.Handler for project creation
type createHandler struct{}

// Execute implements the Handler interface
func (h *createHandler) Execute(ctx context.Context, args *handler.Arguments) (interface{}, error) {
	// Get flag values from arguments
	projectTitle := args.MustGetString("title")
	projectDescription := args.GetString("description", "")

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialization error: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("Error closing CLI", "error", err)
		}
	}()

	// Create project
	project, err := cliInstance.App.ProjectService.CreateProject(ctx, projectservice.CreateProjectRequest{
		Name:        projectTitle,
		Description: projectDescription,
	})
	if err != nil {
		return nil, fmt.Errorf("project creation error: %w", err)
	}

	return &projectCreateResult{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.String(),
	}, nil
}

// projectCreateResult represents the result of project creation
type projectCreateResult struct {
	ID          int
	Name        string
	Description string
	CreatedAt   string
}

// GetID implements the GetID interface for quiet mode output
func (r *projectCreateResult) GetID() int {
	return r.ID
}

func parseCreateFlags(cmd *cobra.Command) error {
	// Validate required flags
	title, _ := cmd.Flags().GetString("title")
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("project title cannot be empty")
	}
	return nil
}
