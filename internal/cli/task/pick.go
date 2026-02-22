package task

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// PickCmd returns the task pick subcommand
func PickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pick [project-id]",
		Short: "Interactively pick a task",
		Long: `Interactively pick a task from a filterable list and output its ID.

Designed for command chaining with other paso commands.

Examples:
  # Pick a task and mark it in-progress
  paso task in-progress $(paso task pick)

  # Pick a task from a specific project
  paso task done $(paso task pick 2)

  # Pick a task and assign someone
  paso task assign $(paso task pick) alice
`,
		Args: cobra.MaximumNArgs(1),
		RunE: runPick,
	}

	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	return cmd
}

func runPick(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	formatter := &cli.OutputFormatter{}

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	var taskProject int
	if len(args) > 0 {
		taskProject, err = strconv.Atoi(args[0])
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "INVALID_PROJECT_ID",
				fmt.Sprintf("Invalid project ID: %s", args[0]),
				"Project ID must be a number")
		}
	} else {
		taskProject, err = cli.GetProjectIDWithCLI(cmd, cliInstance)
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "NO_PROJECT",
				err.Error(),
				"Specify project ID as argument, use --project flag, or associate branch with project")
		}
	}

	tasksByColumn, err := cliInstance.App.TaskService.GetTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		return formatter.Error(cli.ExitError, "TASK_FETCH_ERROR", err.Error())
	}

	allTasks := FlattenTasksByColumn(tasksByColumn)

	if len(allTasks) == 0 {
		return formatter.Error(cli.ExitError, "NO_TASKS", "No tasks found in this project")
	}

	selected, err := cli.RunPick("Pick a task", allTasks)
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return formatter.Error(cli.ExitUsage, "PICK_CANCELLED", "Selection cancelled")
		}
		return formatter.Error(cli.ExitError, "PICK_ERROR", err.Error())
	}

	fmt.Println(selected)
	return nil
}
