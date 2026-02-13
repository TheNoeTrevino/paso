package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
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

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")
	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Parse ID from positional argument
	taskID, err := ParseUpdateID(args)
	if err != nil {
		return formatter.Error(cli.ExitValidation, "INVALID_ID", err.Error())
	}

	// Parse update flags
	input, err := ParseUpdateFlags(cmd)
	if err != nil {
		return formatter.Error(cli.ExitUsage, "NO_UPDATES", err.Error())
	}

	input.TaskID = taskID

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

	// Update title/description if provided
	if HasContentUpdate(input) {
		req := taskservice.UpdateTaskRequest{
			TaskID:      input.TaskID,
			Title:       input.Title,
			Description: input.Description,
		}
		if err := cliInstance.App.TaskService.UpdateTask(ctx, req); err != nil {
			return formatter.Error(cli.ExitError, "UPDATE_ERROR", err.Error())
		}
	}

	// Update priority if provided
	if HasPriorityUpdate(input) {
		priorityID, err := cli.ParsePriority(*input.Priority)
		if err != nil {
			return formatter.Error(cli.ExitValidation, "INVALID_PRIORITY", err.Error())
		}
		req := taskservice.UpdateTaskRequest{
			TaskID:     input.TaskID,
			PriorityID: &priorityID,
		}
		if err := cliInstance.App.TaskService.UpdateTask(ctx, req); err != nil {
			return formatter.Error(cli.ExitError, "PRIORITY_UPDATE_ERROR", err.Error())
		}
	}

	// Output success
	result := &UpdateResult{TaskID: taskID}

	if quietMode {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatUpdateJSON(result))
	}

	cli.PrintSuccess(FormatUpdateOutput(result))
	return nil
}
