package project

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/git"
	projectsvc "github.com/thenoetrevino/paso/internal/services/project"
)

// GitUnlinkCmd returns the git-unlink subcommand
func GitUnlinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-unlink",
		Short: "Unlink a project from its git branch",
		Long: `Unlink a project from its git branch.

If --id/-i is not provided, defaults to the project associated with the current branch.

Examples:
  # Shorthand flags
  paso project git-unlink -i 5
  paso project git-unlink          # unlinks current project

  # Long-form flags also supported
  paso project git-unlink --id 5`,
		RunE: runGitUnlink,
	}

	cmd.Flags().IntP("id", "i", 0, "Project ID to unlink (defaults to current project)")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runGitUnlink(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	projectID, _ := cmd.Flags().GetInt("id")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
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

	// Resolve project ID
	if projectID == 0 {
		// Default to current project (project associated with current branch)
		gitInfo := git.DetectGitInfo(ctx)

		if !gitInfo.IsRepo {
			return outputError(formatter, jsonOutput, "NOT_IN_GIT_REPO",
				"not in a git repository; specify --id")
		}
		if gitInfo.IsDetached || gitInfo.CurrentBranch == "" {
			return outputError(formatter, jsonOutput, "DETACHED_HEAD",
				"cannot determine current branch (detached HEAD); specify --id")
		}

		currentProject, err := cliInstance.App.ProjectService.GetProjectByGitBranch(ctx, gitInfo.CurrentBranch)
		if err != nil {
			if fmtErr := formatter.Error("PROJECT_FETCH_ERROR", err.Error()); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			return err
		}
		if currentProject == nil {
			return outputError(formatter, jsonOutput, "NO_CURRENT_PROJECT",
				"no project associated with current branch; specify --id")
		}
		projectID = currentProject.ID
	}

	// Get the project
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, projectID)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project with ID %d not found", projectID)); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		return err
	}

	// Check if project has a branch linked
	if project.GitBranch == "" {
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{
				"success": true,
				"warning": true,
				"message": fmt.Sprintf("project \"%s\" (id: %d) has no branch linked", project.Name, project.ID),
			})
		}
		fmt.Printf("warning: project \"%s\" (id: %d) has no branch linked\n", project.Name, project.ID)
		return nil
	}

	oldBranch := project.GitBranch

	// Unlink the branch
	emptyBranch := ""
	if err := cliInstance.App.ProjectService.UpdateProject(ctx, projectsvc.UpdateProjectRequest{
		ID:        projectID,
		GitBranch: &emptyBranch,
	}); err != nil {
		if fmtErr := formatter.Error("UNLINK_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		return err
	}

	// Output success
	if quietMode {
		fmt.Printf("%d\n", projectID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"project": map[string]any{
				"id":   project.ID,
				"name": project.Name,
			},
			"unlinked_branch": oldBranch,
			"message": fmt.Sprintf("Unlinked branch \"%s\" from project \"%s\" (id: %d)",
				oldBranch, project.Name, project.ID),
		})
	}

	fmt.Printf("Unlinked branch \"%s\" from project \"%s\" (id: %d)\n",
		oldBranch, project.Name, project.ID)

	return nil
}
