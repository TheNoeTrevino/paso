package project

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	projectsvc "github.com/thenoetrevino/paso/internal/services/project"
)

// GitLinkCmd returns the git-link subcommand
func GitLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-link [project-id] [branch]",
		Short: "Link a project to a git branch",
		Args:  cobra.MaximumNArgs(2),
		Long: `Link a project to a git branch.

Positional arguments and flags can be used interchangeably:
  [project-id] can be provided as a positional argument or via --id/-i flag
  [branch] can be provided as a positional argument or via --branch/-b flag

If project-id is not provided, defaults to the project associated with the current branch.
If branch is not provided, defaults to the current git branch.

Examples:
  paso project git-link 5 feature/auth
  paso project git-link 5                     # uses current branch
  paso project git-link -i 5 -b feature/auth
  paso project git-link -b feature/new        # uses current project
  paso project git-link -i 5                  # uses current branch
  paso project git-link                       # uses current project and branch

  # Force transfer branch from another project
  paso project git-link 5 feature/auth -f`,
		RunE: runGitLink,
	}

	cmd.Flags().IntP("id", "i", 0, "Project ID to link (can also be provided as positional argument)")
	cmd.Flags().StringP("branch", "b", "", "Branch name to link (can also be provided as positional argument)")
	cmd.Flags().BoolP("force", "f", false, "Transfer branch from another project if already linked")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runGitLink(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	force, _ := cmd.Flags().GetBool("force")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Parse project ID from positional arg or flag
	var projectID int
	if len(args) > 0 {
		if id, err := strconv.Atoi(args[0]); err == nil {
			projectID = id
		} else {
			return outputError(formatter, jsonOutput, "INVALID_ID",
				fmt.Sprintf("invalid project ID '%s': must be a number", args[0]))
		}
	} else {
		projectID, _ = cmd.Flags().GetInt("id")
	}

	// Parse branch name from positional arg or flag
	var branchName string
	if len(args) > 1 {
		branchName = args[1]
	} else {
		branchName, _ = cmd.Flags().GetString("branch")
	}

	// Initialize CLI first so we can use the injectable git detector
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Detect git info for defaults
	gitInfo := cliInstance.App.GitDetector.DetectGitInfo(ctx)

	// Resolve branch name
	if branchName == "" {
		if !gitInfo.IsRepo {
			return outputError(formatter, jsonOutput, "NOT_IN_GIT_REPO",
				"not in a git repository; specify --branch")
		}
		if gitInfo.IsDetached || gitInfo.CurrentBranch == "" {
			return outputError(formatter, jsonOutput, "DETACHED_HEAD",
				"cannot determine current branch (detached HEAD); specify --branch")
		}
		branchName = gitInfo.CurrentBranch
	}

	// Resolve project ID
	if projectID == 0 {
		// Default to current project (project associated with current branch)
		if !gitInfo.IsRepo || gitInfo.IsDetached || gitInfo.CurrentBranch == "" {
			return outputError(formatter, jsonOutput, "NO_CURRENT_PROJECT",
				"no project associated with current branch; specify --id")
		}

		currentProject, err := cliInstance.App.ProjectService.GetProjectByGitBranch(ctx, gitInfo.CurrentBranch)
		if err != nil {
			if fmtErr := formatter.Error("PROJECT_FETCH_ERROR", err.Error()); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitError)
		}
		if currentProject == nil {
			return outputError(formatter, jsonOutput, "NO_CURRENT_PROJECT",
				"no project associated with current branch; specify --id")
		}
		projectID = currentProject.ID
	}

	// Get the project to display its name
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, projectID)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project with ID %d not found", projectID)); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Check if branch is already linked to another project
	existingProject, err := cliInstance.App.ProjectService.GetProjectByGitBranch(ctx, branchName)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}

	if existingProject != nil && existingProject.ID != projectID {
		if !force {
			msg := fmt.Sprintf("branch \"%s\" is already linked to project \"%s\" (id: %d)",
				branchName, existingProject.Name, existingProject.ID)
			hint := "hint: use --force to transfer the branch to this project"

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"success": false,
					"error":   "BRANCH_ALREADY_LINKED",
					"message": msg,
					"hint":    hint,
					"existing_project": map[string]any{
						"id":   existingProject.ID,
						"name": existingProject.Name,
					},
				})
			}
			fmt.Fprintf(os.Stderr, "error: %s\n%s\n", msg, hint)
			return fmt.Errorf("%s", msg)
		}

		// Force mode: unlink from existing project first
		emptyBranch := ""
		if err := cliInstance.App.ProjectService.UpdateProject(ctx, projectsvc.UpdateProjectRequest{
			ID:        existingProject.ID,
			GitBranch: &emptyBranch,
		}); err != nil {
			if fmtErr := formatter.Error("UNLINK_ERROR", err.Error()); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitError)
		}

		if !quietMode && !jsonOutput {
			fmt.Printf("Unlinked branch \"%s\" from project \"%s\" (id: %d)\n",
				branchName, existingProject.Name, existingProject.ID)
		}
	}

	// Link the branch to the project
	if err := cliInstance.App.ProjectService.UpdateProject(ctx, projectsvc.UpdateProjectRequest{
		ID:        projectID,
		GitBranch: &branchName,
	}); err != nil {
		if fmtErr := formatter.Error("LINK_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}

	// Output success
	if quietMode {
		fmt.Printf("%d\n", projectID)
		return nil
	}

	if jsonOutput {
		result := map[string]any{
			"success": true,
			"project": map[string]any{
				"id":     project.ID,
				"name":   project.Name,
				"branch": branchName,
			},
			"message": fmt.Sprintf("Linked project \"%s\" (id: %d) to branch \"%s\"",
				project.Name, project.ID, branchName),
		}
		if existingProject != nil && existingProject.ID != projectID {
			result["unlinked_from"] = map[string]any{
				"id":   existingProject.ID,
				"name": existingProject.Name,
			}
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	fmt.Printf("Linked project \"%s\" (id: %d) to branch \"%s\"\n",
		project.Name, project.ID, branchName)

	return nil
}

func outputError(formatter *cli.OutputFormatter, jsonOutput bool, errorCode, message string) error {
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": false,
			"error":   errorCode,
			"message": message,
		})
	}
	fmt.Fprintf(os.Stderr, "error: %s\n", message)
	return fmt.Errorf("%s", message)
}
