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

// GitUnlinkCmd returns the git-unlink subcommand
func GitUnlinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git-unlink [project-id]",
		Short: "Unlink a project from its git branch",
		Args:  cobra.MaximumNArgs(1),
		Long: `Unlink a project from its git branch.

Positional argument and flag can be used interchangeably:
  [project-id] can be provided as a positional argument or via --id/-i flag

If project-id is not provided, defaults to the project associated with the current branch.

Examples:
  paso project git-unlink 5
  paso project git-unlink -i 5
  paso project git-unlink  # unlinks current project`,
		RunE: runGitUnlink,
	}

	cmd.Flags().IntP("id", "i", 0, "Project ID to unlink (can also be provided as positional argument)")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runGitUnlink(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

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

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Resolve project ID
	if projectID == 0 {
		// Default to current project (project associated with current branch)
		gitInfo := cliInstance.App.GitDetector.DetectGitInfo(ctx)

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
			return formatter.Error(cli.ExitError, "PROJECT_FETCH_ERROR", err.Error())
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
		return formatter.Error(cli.ExitNotFound, "PROJECT_NOT_FOUND", fmt.Sprintf("project with ID %d not found", projectID))
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
		return formatter.Error(cli.ExitError, "UNLINK_ERROR", err.Error())
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
