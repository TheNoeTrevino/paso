package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// DeleteCmd returns the task delete subcommand
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a task",
		Long:  "Delete a task by ID (requires confirmation unless --force or --quiet).",
		RunE:  runDelete,
	}

	cmd.Flags().Int("id", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		slog.Error("Error marking flag as required", "error", err)
	}

	cmd.Flags().Bool("force", false, "Skip confirmation")

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	taskID, _ := cmd.Flags().GetInt("id")
	force, _ := cmd.Flags().GetBool("force")
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

	// Get task details for confirmation
	task, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
			slog.Error("Error formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Printf("Delete task #%d: '%s'? (y/N): ", taskID, task.Title)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			slog.Error("Error reading user input", "error", err)
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete the task
	if err := cliInstance.App.TaskService.DeleteTask(ctx, taskID); err != nil {
		if fmtErr := formatter.Error("DELETE_ERROR", err.Error()); fmtErr != nil {
			slog.Error("Error formatting error message", "error", fmtErr)
		}
		return err
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"task_id": taskID,
		})
	}

	fmt.Printf("✓ Task %d deleted successfully\n", taskID)
	return nil
}
