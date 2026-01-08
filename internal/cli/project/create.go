// Package project holds all cli commands related to projects
//
// e.g., paso project ...
package project

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/handler"
	"github.com/thenoetrevino/paso/internal/git"
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
		slog.Error("failed to mark flag as required", "error", err)
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
func (h *createHandler) Execute(ctx context.Context, args *handler.Arguments) (any, error) {
	// Get flag values from arguments
	projectTitle := args.MustGetString("title")
	projectDescription := args.GetString("description", "")

	// Check if quiet mode is enabled
	quietMode, _ := args.GetCmd().Flags().GetBool("quiet")

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialization error: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Detect git repository information and try to associate with branch
	gitInfo := git.DetectGitInfo(ctx)
	var gitBranch string
	if gitInfo.IsValidForAssociation() {
		gitBranch = gitInfo.CurrentBranch
		if !quietMode {
			fmt.Printf("ℹ️  Associating project with git branch: %s\n", gitBranch)
		}
	}

	// Create project - service layer will handle duplicate branch validation
	project, err := cliInstance.App.ProjectService.CreateProject(ctx, projectservice.CreateProjectRequest{
		Name:        projectTitle,
		Description: projectDescription,
		GitBranch:   gitBranch,
	})

	// Handle git branch conflict by retrying without branch association
	if err != nil && errors.Is(err, projectservice.ErrGitBranchAlreadyAssociated) {
		if !quietMode {
			existingProject, _ := cliInstance.App.ProjectService.GetProjectByGitBranch(ctx, gitBranch)
			if existingProject != nil {
				fmt.Printf("⚠️  Warning: Branch '%s' is already associated with project '%s' (ID: %d)\n",
					gitBranch, existingProject.Name, existingProject.ID)
			}
			fmt.Println("Creating new project without branch association...")
		}

		// Retry without branch association
		project, err = cliInstance.App.ProjectService.CreateProject(ctx, projectservice.CreateProjectRequest{
			Name:        projectTitle,
			Description: projectDescription,
			GitBranch:   "", // Don't associate with branch
		})
	}

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

// String provides human-readable output for the project creation result
func (r *projectCreateResult) String() string {
	if r.Description != "" {
		return fmt.Sprintf("✓ Project created: %s (ID: %d)\n  Description: %s\n  Created: %s",
			r.Name, r.ID, r.Description, r.CreatedAt)
	}
	return fmt.Sprintf("✓ Project created: %s (ID: %d)\n  Created: %s",
		r.Name, r.ID, r.CreatedAt)
}

func parseCreateFlags(cmd *cobra.Command) error {
	// Validate required flags
	title, _ := cmd.Flags().GetString("title")
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("project title cannot be empty")
	}
	return nil
}
