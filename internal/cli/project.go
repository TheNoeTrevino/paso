package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func ProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}

	cmd.AddCommand(projectCreateCmd())
	cmd.AddCommand(projectListCmd())
	cmd.AddCommand(projectDeleteCmd())

	return cmd
}

func projectCreateCmd() *cobra.Command {
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
		RunE: runProjectCreate,
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

func runProjectCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Validate title is not empty
	if strings.TrimSpace(projectTitle) == "" {
		if fmtErr := formatter.Error("VALIDATION_ERROR", "project title cannot be empty"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(ExitValidation)
	}

	// Create project
	project, err := cli.Repo.CreateProject(ctx, projectTitle, projectDescription)
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

func projectListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Long:  "List all projects with their details.",
		RunE:  runProjectList,
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (IDs only)")

	return cmd
}

func runProjectList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Get all projects
	projects, err := cli.Repo.GetAllProjects(ctx)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Output in appropriate format
	if quietMode {
		// Just print IDs (one per line)
		for _, p := range projects {
			fmt.Printf("%d\n", p.ID)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success":  true,
			"projects": projects,
		})
	}

	// Human-readable output
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	fmt.Printf("Found %d projects:\n\n", len(projects))
	for _, p := range projects {
		fmt.Printf("  [%d] %s", p.ID, p.Name)
		if p.Description != "" {
			fmt.Printf(" - %s", p.Description)
		}
		fmt.Println()
	}

	return nil
}

func projectDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a project",
		Long:  "Delete a project by ID (requires confirmation unless --force or --quiet).",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

			// Initialize CLI
			cli, err := NewCLI(ctx)
			if err != nil {
				if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				return err
			}
			defer func() {
				if err := cli.Close(); err != nil {
					log.Printf("Error closing CLI: %v", err)
				}
			}()

			projectID, _ = cmd.Flags().GetInt("id")
			force, _ = cmd.Flags().GetBool("force")

			// Get project details for confirmation
			project, err := cli.Repo.GetProjectByID(ctx, projectID)
			if err != nil {
				if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", projectID)); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				os.Exit(ExitNotFound)
			}

			// Ask for confirmation unless force or quiet mode
			if !force && !quietMode {
				fmt.Printf("Delete project #%d: '%s'? (y/N): ", projectID, project.Name)
				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					log.Printf("Error reading user input: %v", err)
				}
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			// Delete the project
			if err := cli.Repo.DeleteProject(ctx, projectID); err != nil {
				if fmtErr := formatter.Error("DELETE_ERROR", err.Error()); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				return err
			}

			// Output success
			if quietMode {
				return nil
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"success":    true,
					"project_id": projectID,
				})
			}

			fmt.Printf("✓ Project %d deleted successfully\n", projectID)
			return nil
		},
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional flags
	cmd.Flags().Bool("force", false, "Skip confirmation")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}
