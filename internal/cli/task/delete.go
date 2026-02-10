package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
)

// DeleteCmd returns the task delete subcommand
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Long: `Delete a task by ID (requires confirmation unless --force/-f or --quiet/-q).

Examples:
  paso task delete 42
  paso task delete 42 -f
  paso task delete 42 --force --json
`,
		Args: cobra.ExactArgs(1),
		RunE: runDelete,
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation")

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	force, _ := cmd.Flags().GetBool("force")
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

	// Get task details for confirmation
	task, err := cliInstance.App.TaskService.GetTaskDetail(ctx, taskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID))
	}

	// Ask for confirmation unless force or quiet mode
	if !force && !quietMode {
		fmt.Printf("Delete task #%d: '%s'? (y/N): ", taskID, task.Title)
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			slog.Error("failed to read user input", "error", err)
			fmt.Println("Cancelled (failed to read input)")
			return nil
		}
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete the task
	if err := cliInstance.App.TaskService.DeleteTask(ctx, taskID); err != nil {
		return formatter.Error(cli.ExitError, "DELETE_ERROR", err.Error())
	}

	// Output success
	if quietMode {
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"task_id": taskID,
		})
	}

	colors := cli.GetColorScheme()
	message := fmt.Sprintf("Task %d deleted successfully", taskID)
	fmt.Print(styles.RenderSuccess(message, colors))
	return nil
}
