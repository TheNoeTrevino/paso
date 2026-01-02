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
		Use:   "update",
		Short: "Update a task",
		Long:  "Update task title, description, or priority.",
		RunE:  runUpdate,
	}

	// Required flags
	cmd.Flags().Int("id", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		slog.Error("Error marking flag as required", "error", err)
	}

	// Optional update flags
	cmd.Flags().String("title", "", "New task title")
	cmd.Flags().String("description", "", "New task description")
	cmd.Flags().String("priority", "", "New priority: trivial, low, medium, high, critical")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	taskID, _ := cmd.Flags().GetInt("id")
	taskTitle, _ := cmd.Flags().GetString("title")
	taskDescription, _ := cmd.Flags().GetString("description")
	taskPriority, _ := cmd.Flags().GetString("priority")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("Error formatting error message", "error", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("Error closing CLI", "error", err)
		}
	}()

	// At least one update field must be provided
	titleFlag := cmd.Flags().Lookup("title")
	descFlag := cmd.Flags().Lookup("description")
	priorityFlag := cmd.Flags().Lookup("priority")

	if !titleFlag.Changed && !descFlag.Changed && !priorityFlag.Changed {
		if fmtErr := formatter.Error("NO_UPDATES", "at least one of --title, --description, or --priority must be specified"); fmtErr != nil {
			slog.Error("Error formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
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
			if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
				slog.Error("Error formatting error message", "error", fmtErr)
			}
			return err
		}
	}

	// Update priority if provided
	if priorityFlag.Changed {
		priorityID, err := cli.ParsePriority(taskPriority)
		if err != nil {
			if fmtErr := formatter.Error("INVALID_PRIORITY", err.Error()); fmtErr != nil {
				slog.Error("Error formatting error message", "error", fmtErr)
			}
			os.Exit(cli.ExitValidation)
		}
		req := taskservice.UpdateTaskRequest{
			TaskID:     taskID,
			PriorityID: &priorityID,
		}
		if err := cliInstance.App.TaskService.UpdateTask(ctx, req); err != nil {
			if fmtErr := formatter.Error("PRIORITY_UPDATE_ERROR", err.Error()); fmtErr != nil {
				slog.Error("Error formatting error message", "error", fmtErr)
			}
			return err
		}
	}

	// Output success
	if quietMode {
		fmt.Printf("%d\n", taskID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"task_id": taskID,
		})
	}

	fmt.Printf("✓ Task %d updated successfully\n", taskID)
	return nil
}
