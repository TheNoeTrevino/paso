package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// SwitchCmd returns the task switch subcommand
func SwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <task_id> <project_id>",
		Short: "Move a task to a different project",
		Long: `Move a task to a different project's ready column.

The task will be placed into the target project's ready column
(the column marked with holds_created_tasks = true).

Examples:
  # Move task 42 to project 5
  paso task switch 42 5

  # JSON output for agents
  paso task switch 42 5 -j

  # Quiet mode for bash capture
  paso task switch 42 5 -q
`,
		RunE: runSwitch,
		Args: cobra.ExactArgs(2),
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runSwitch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	input, err := ParseSwitchArgs(args)
	if err != nil {
		formatter := &cli.OutputFormatter{}
		return formatter.Error(cli.ExitValidation, "INVALID_INPUT", err.Error())
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Verify task exists
	_, err = cliInstance.App.TaskService.GetTaskDetail(ctx, input.TaskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", input.TaskID))
	}

	// Verify project exists and get its name
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, input.ProjectID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", input.ProjectID))
	}

	// Move task to the target project
	err = cliInstance.App.TaskService.MoveTaskToProject(ctx, input.TaskID, input.ProjectID)
	if err != nil {
		if errors.Is(err, taskservice.ErrTaskAlreadyInTargetProject) {
			return formatter.Error(cli.ExitValidation, "ALREADY_IN_PROJECT",
				fmt.Sprintf("task %d is already in project %s", input.TaskID, project.Name))
		}
		return formatter.Error(cli.ExitError, "SWITCH_ERROR", err.Error())
	}

	result := &SwitchResult{
		TaskID:      input.TaskID,
		ProjectID:   input.ProjectID,
		ProjectName: project.Name,
	}

	if quietMode {
		fmt.Printf("%d\n", input.TaskID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatSwitchJSON(result))
	}

	cli.PrintSuccess(FormatSwitchOutput(result))
	return nil
}
