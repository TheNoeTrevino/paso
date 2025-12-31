// Package use holds all cli commands related to setting contextual information
// e.g., paso use ...
package use

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// ProjectCmd returns the use project subcommand
func ProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [project-id]",
		Short: "Set project context for current shell session",
		Long: `Set the current project context using environment variables.
This command outputs shell commands that should be evaluated:

  eval $(paso use project 3)              # Use project 3
  eval $(paso use project --clear)        # Clear project context
  paso use project --show                 # Show current project

The PASO_PROJECT environment variable will be set in your current shell
session only. The --project flag on other commands takes precedence over
this environment variable.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runUseProject,
	}

	cmd.Flags().Bool("clear", false, "Clear the current project context")
	cmd.Flags().Bool("show", false, "Show the current project context")
	cmd.Flags().Bool("dry-run", false, "Show what would be exported without outputting shell commands")

	return cmd
}

func runUseProject(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	clearFlag, _ := cmd.Flags().GetBool("clear")
	showFlag, _ := cmd.Flags().GetBool("show")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Handle --show flag
	if showFlag {
		return showCurrentProject()
	}

	// Handle --clear flag
	if clearFlag {
		if dryRun {
			fmt.Fprintf(os.Stderr, "Would clear PASO_PROJECT\n")
			return nil
		}
		fmt.Println("unset PASO_PROJECT")
		fmt.Fprintf(os.Stderr, "Cleared project context\n")
		return nil
	}

	// Validate project ID provided
	if len(args) == 0 {
		return fmt.Errorf("project ID required\nUsage: eval $(paso use project <project-id>)")
	}

	var projectID int
	if _, err := fmt.Sscanf(args[0], "%d", &projectID); err != nil {
		return fmt.Errorf("invalid project ID: %s", args[0])
	}

	// Initialize CLI
	cliInstance, err := cli.NewCLI(ctx)
	if err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: project %d not found\n", projectID)
		fmt.Fprintf(os.Stderr, "Suggestion: Use 'paso project list' to see available projects\n")
		os.Exit(cli.ExitNotFound)
	}

	// Output shell export command (to stdout for eval)
	if dryRun {
		fmt.Fprintf(os.Stderr, "Would set PASO_PROJECT=%d (%s)\n", projectID, project.Name)
		return nil
	}

	fmt.Printf("export PASO_PROJECT=%d\n", projectID)
	fmt.Fprintf(os.Stderr, "Now using project %d: %s\n", projectID, project.Name)

	return nil
}

func showCurrentProject() error {
	currentProject := os.Getenv("PASO_PROJECT")
	if currentProject == "" {
		fmt.Println("No project context set")
		fmt.Println("Use 'eval $(paso use project <project-id>)' to set one")
		return nil
	}

	ctx := context.Background()
	var projectID int
	if _, err := fmt.Sscanf(currentProject, "%d", &projectID); err != nil {
		fmt.Printf("Invalid project context: %s\n", currentProject)
		return nil
	}

	// Initialize CLI
	cliInstance, err := cli.NewCLI(ctx)
	if err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Get project details
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, projectID)
	if err != nil {
		fmt.Printf("Current project: %s (project not found)\n", currentProject)
		return nil
	}

	fmt.Printf("Current project: %d (%s)\n", projectID, project.Name)
	return nil
}
