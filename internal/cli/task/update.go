package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// UpdateCmd returns the task update subcommand
func UpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a task",
		Long: `Update task title, description, or priority.

Examples:
  # Update title and priority
  paso task update 42 -t "New title" -r high

  # Update description
  paso task update 42 -d "Updated description"

  # Long-form flags also supported
  paso task update 42 --title="New title" --priority=high
`,
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}

	// Optional update flags
	cmd.Flags().StringP("title", "t", "", "New task title")
	cmd.Flags().StringP("description", "d", "", "New task description")
	cmd.Flags().StringP("priority", "r", "", "New priority: trivial, low, medium, high, critical")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	taskTitle, _ := cmd.Flags().GetString("title")
	taskDescription, _ := cmd.Flags().GetString("description")
	taskPriority, _ := cmd.Flags().GetString("priority")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Parse ID from positional argument
	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return formatter.Error(cli.ExitValidation, "INVALID_ID", fmt.Sprintf("invalid ID '%s': must be a number", args[0]))
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

	// At least one update field must be provided
	titleFlag := cmd.Flags().Lookup("title")
	descFlag := cmd.Flags().Lookup("description")
	priorityFlag := cmd.Flags().Lookup("priority")

	if !titleFlag.Changed && !descFlag.Changed && !priorityFlag.Changed {
		return formatter.Error(cli.ExitUsage, "NO_UPDATES", "at least one of --title, --description, or --priority must be specified")
	}

	// Update title/description if provided
	if titleFlag.Changed || descFlag.Changed {
		req := taskservice.UpdateTaskRequest{
			TaskID: taskID,
		}
		if titleFlag.Changed {
			req.Title = &taskTitle
		}
		if descFlag.Changed {
			req.Description = &taskDescription
		}
		if err := cliInstance.App.TaskService.UpdateTask(ctx, req); err != nil {
			return formatter.Error(cli.ExitError, "UPDATE_ERROR", err.Error())
		}
	}

	// Update priority if provided
	if priorityFlag.Changed {
		priorityID, err := cli.ParsePriority(taskPriority)
		if err != nil {
			return formatter.Error(cli.ExitValidation, "INVALID_PRIORITY", err.Error())
		}
		req := taskservice.UpdateTaskRequest{
			TaskID:     taskID,
			PriorityID: &priorityID,
		}
		if err := cliInstance.App.TaskService.UpdateTask(ctx, req); err != nil {
			return formatter.Error(cli.ExitError, "PRIORITY_UPDATE_ERROR", err.Error())
		}
	}

	// Output success
	if quietMode {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"task_id": taskID,
		})
	}

	colors := cli.GetColorScheme()
	message := fmt.Sprintf("Task %d updated successfully", taskID)
	fmt.Print(styles.RenderSuccess(message, colors))
	return nil
}
