package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// ArchiveCmd returns the task archive subcommand
func ArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <task_id>",
		Short: "Toggle archived status of a task",
		Long: `Toggle the archived state of a task.

If the task is currently active, it will be archived.
If the task is already archived, it will be unarchived (restored to active).

Examples:
  # Archive or unarchive a task
  paso task archive 42

  # JSON output for agents
  paso task archive 42 -j

  # Quiet mode for bash capture
  paso task archive 42 -q

  # Long-form flags also supported
  paso task archive 42 --json
`,
		RunE: runArchive,
		Args: cobra.ExactArgs(1),
	}

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runArchive(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Parse task ID from positional argument
	input, err := ParseArchiveArgs(args)
	if err != nil {
		formatter := &cli.OutputFormatter{}
		return formatter.Error(cli.ExitValidation, "INVALID_ID", err.Error())
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

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

	// Toggle archived state
	err = cliInstance.App.TaskService.ArchiveTask(ctx, input.TaskID)
	if err != nil {
		return formatter.Error(cli.ExitError, "ARCHIVE_ERROR", err.Error())
	}

	// Get updated task detail to check new archived state
	updatedTask, err := cliInstance.App.TaskService.GetTaskDetail(ctx, input.TaskID)
	if err != nil {
		return formatter.Error(cli.ExitError, "TASK_FETCH_ERROR", err.Error())
	}

	result := &ArchiveResult{
		TaskID:      input.TaskID,
		WasArchived: updatedTask.Archived,
	}

	// Output success
	if quietMode {
		fmt.Printf("%d\n", input.TaskID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatArchiveJSON(result))
	}

	// Human-readable output
	message := FormatArchiveOutput(result)
	cli.PrintSuccess(message)
	return nil
}
