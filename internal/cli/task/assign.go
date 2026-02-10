package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/user"
)

// AssignCmd returns the task assign subcommand
func AssignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign <task_id> [assignee_name]",
		Short: "Assign a task to an assignee",
		Long: `Assign a task to an assignee by name.

If no assignee name is given, defaults to the active assignee.
Use --clear to remove the assignee from a task.

Examples:
  # Assign to a specific person
  paso task assign 42 alice

  # Assign to active assignee (yourself)
  paso task assign 42

  # Remove assignee
  paso task assign 42 --clear

  # JSON output for agents
  paso task assign 42 alice -j
`,
		RunE: runAssign,
		Args: cobra.RangeArgs(1, 2),
	}

	cmd.Flags().Bool("clear", false, "Remove assignee from the task")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runAssign(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	clearAssignee, _ := cmd.Flags().GetBool("clear")
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
	_, err = cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID))
	}

	if clearAssignee {
		return assignTask(cmd, cliInstance, taskID, nil, "", formatter, jsonOutput, quietMode)
	}

	// Resolve assignee name
	assigneeName := ""
	if len(args) > 1 {
		assigneeName = args[1]
	}
	if assigneeName == "" {
		cfg, err := config.Load()
		if err == nil {
			assigneeName = cfg.GetActiveAssignee()
		}
		if assigneeName == "" {
			assigneeName = user.GetCurrentUsername()
		}
	}

	assignee, err := cliInstance.App.AssigneeService.GetOrCreate(ctx, assigneeName)
	if err != nil {
		return formatter.Error(cli.ExitError, "ASSIGNEE_ERROR", fmt.Sprintf("failed to resolve assignee '%s': %s", assigneeName, err))
	}

	assigneeID := &assignee.ID
	return assignTask(cmd, cliInstance, taskID, assigneeID, assigneeName, formatter, jsonOutput, quietMode)
}

func assignTask(cmd *cobra.Command, cliInstance *cli.CLI, taskID int, assigneeID *int, assigneeName string, formatter *cli.OutputFormatter, jsonOutput, quietMode bool) error {
	ctx := cmd.Context()

	err := cliInstance.App.TaskService.UpdateTaskAssignee(ctx, taskID, assigneeID)
	if err != nil {
		return formatter.Error(cli.ExitError, "ASSIGN_ERROR", err.Error())
	}

	if quietMode {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if jsonOutput {
		result := map[string]any{
			"success": true,
			"task_id": taskID,
		}
		if assigneeID != nil {
			result["assignee"] = assigneeName
		} else {
			result["assignee"] = nil
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	if assigneeID != nil {
		fmt.Printf("Task %d assigned to @%s\n", taskID, assigneeName)
	} else {
		fmt.Printf("Task %d assignee cleared\n", taskID)
	}
	return nil
}
