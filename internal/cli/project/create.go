// Package project holds all cli commands related to projects
//
// e.g., paso project ...
package project

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
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
		RunE: runCreate,
	}

	// Required flags
	cmd.Flags().String("title", "", "Project title (required)")
	if err := cmd.MarkFlagRequired("title"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional flags
	cmd.Flags().String("description", "", "Project description")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	projectTitle, _ := cmd.Flags().GetString("title")
	projectDescription, _ := cmd.Flags().GetString("description")
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

	// Validate title is not empty
	if strings.TrimSpace(projectTitle) == "" {
		if fmtErr := formatter.Error("VALIDATION_ERROR", "project title cannot be empty"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitValidation)
	}

	// Create project
	project, err := cliInstance.App.ProjectService.CreateProject(ctx, projectservice.CreateProjectRequest{
		Name:        projectTitle,
		Description: projectDescription,
	})
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_CREATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output based on mode (JSON/Quiet/Human)
	if quietMode {
		fmt.Printf("%d\n", project.ID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"project": map[string]interface{}{
				"id":          project.ID,
				"name":        project.Name,
				"description": project.Description,
				"created_at":  project.CreatedAt,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Project '%s' created successfully (ID: %d)\n", projectTitle, project.ID)
	if projectDescription != "" {
		fmt.Printf("  Description: %s\n", projectDescription)
	}

	return nil
}
