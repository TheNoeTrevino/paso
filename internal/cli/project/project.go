package project

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/git"
)

// ProjectCmd returns the project parent command
func ProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
		RunE:  runProject,
	}

	// Flags for --show-current
	cmd.Flags().Bool("show-current", false, "Show the project associated with the current git branch")
	cmd.Flags().Bool("json", false, "Output in JSON format (used with --show-current)")
	cmd.Flags().Bool("quiet", false, "Minimal output - ID only (used with --show-current)")

	cmd.AddCommand(CreateCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(DeleteCmd())
	cmd.AddCommand(TreeCmd())
	cmd.AddCommand(GitLinkCmd())
	cmd.AddCommand(GitUnlinkCmd())

	return cmd
}

func runProject(cmd *cobra.Command, args []string) error {
	showCurrent, _ := cmd.Flags().GetBool("show-current")

	if !showCurrent {
		// Default behavior: show help
		return cmd.Help()
	}

	ctx := cmd.Context()
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Detect git info
	gitInfo := git.DetectGitInfo(ctx)

	if !gitInfo.IsRepo {
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{
				"success": false,
				"error":   "NOT_IN_GIT_REPO",
				"message": "not in a git repository",
			})
		}
		fmt.Fprintln(os.Stderr, "error: not in a git repository")
		return fmt.Errorf("not in a git repository")
	}

	if gitInfo.IsDetached || gitInfo.CurrentBranch == "" {
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{
				"success": false,
				"error":   "DETACHED_HEAD",
				"message": "cannot determine current branch (detached HEAD)",
			})
		}
		fmt.Fprintln(os.Stderr, "error: cannot determine current branch (detached HEAD)")
		return fmt.Errorf("cannot determine current branch (detached HEAD)")
	}

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

	// Find project by git branch
	project, err := cliInstance.App.ProjectService.GetProjectByGitBranch(ctx, gitInfo.CurrentBranch)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		return err
	}

	if project == nil {
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{
				"success": true,
				"project": nil,
				"message": fmt.Sprintf("no project associated with branch '%s'", gitInfo.CurrentBranch),
			})
		}
		if quietMode {
			// Nothing to output in quiet mode when no project
			return nil
		}
		fmt.Printf("no project associated with branch '%s'\n", gitInfo.CurrentBranch)
		return nil
	}

	// Output the project
	if quietMode {
		fmt.Printf("%d\n", project.ID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"project": project,
		})
	}

	// Human-readable output
	fmt.Printf("ID:          %d\n", project.ID)
	fmt.Printf("Name:        %s\n", project.Name)
	if project.Description != "" {
		fmt.Printf("Description: %s\n", project.Description)
	}
	fmt.Printf("Branch:      %s\n", project.GitBranch)

	return nil
}
