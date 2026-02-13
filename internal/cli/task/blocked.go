package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// BlockedCmd returns the task blocked subcommand
func BlockedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocked",
		Short: "List blocked tasks",
		Long: `List all tasks that are blocked by dependencies.

These are tasks that cannot be started until their blocking
dependencies are completed.

Examples:
  # Human-readable output (shorthand)
  paso task blocked -p 1

  # JSON output for agents
  paso task blocked -p 1 -j

  # Quiet mode for bash capture
  TASK_IDS=$(paso task blocked -p 1 -q)

  # Long-form flags also supported
  paso task blocked --project=1 --json
`,
		RunE: runBlocked,
	}

	// Flags
	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runBlocked(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI first
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Get project ID from flag or git branch
	taskProject, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		return formatter.ErrorWithSuggestion(cli.ExitUsage, "NO_PROJECT",
			err.Error(),
			"Use --project flag or create a project associated with this git branch")
	}

	// Validate project exists
	_, err = cliInstance.App.ProjectService.GetProjectByID(ctx, taskProject)
	if err != nil {
		return formatter.ErrorWithSuggestion(cli.ExitNotFound, "PROJECT_NOT_FOUND",
			fmt.Sprintf("project %d not found", taskProject),
			"Use 'paso project list' to see available projects")
	}

	// Get all tasks for project (includes IsBlocked field)
	tasksByColumn, err := cliInstance.App.TaskService.GetTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		return formatter.Error(cli.ExitError, "TASK_FETCH_ERROR", err.Error())
	}

	// Filter for blocked tasks (IsBlocked == true)
	blockedTasks := FilterBlockedTasks(tasksByColumn)
	result := &BlockedResult{
		Tasks: blockedTasks,
		Count: len(blockedTasks),
	}

	// Output in appropriate format
	if quietMode {
		fmt.Print(FormatBlockedQuiet(result))
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatBlockedJSON(result))
	}

	// Human-readable output
	fmt.Print(FormatBlockedOutput(result))
	return nil
}
