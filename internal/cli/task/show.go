package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/config"
)

// ShowCmd returns the task show subcommand
func ShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [id]",
		Short: "Show task details",
		Long: `Display all details of a task including description, relationships, labels, and metadata.

Examples:
  paso task show 42
  paso task show -i 42
  paso task show -i 42 -j
  paso task show --id=42 --json
`,
		Args: cobra.MaximumNArgs(1),
		RunE: runShow,
	}

	// Flags
	cmd.Flags().IntP("id", "i", 0, "Task ID (can also be provided as positional argument)")
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Parse flags
	flagID, _ := cmd.Flags().GetInt("id")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Parse and validate task ID
	input, err := ParseShowTaskID(args, flagID)
	if err != nil {
		return formatter.ErrorWithSuggestion(cli.ExitUsage, "INVALID_TASK_ID",
			err.Error(),
			"Usage: paso task show <id> or paso task show --id=<id>")
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

	// Get task details
	task, err := cliInstance.App.TaskService.GetTaskDetail(ctx, input.TaskID)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "TASK_NOT_FOUND", fmt.Sprintf("task %d not found", input.TaskID))
	}

	// Output in appropriate format
	if quietMode {
		fmt.Printf("%d\n", task.ID)
		return nil
	}

	if jsonOutput {
		output := FormatShowJSON(task)
		return json.NewEncoder(os.Stdout).Encode(output)
	}

	// Load config for color scheme
	cfg, err := config.Load()
	if err != nil {
		// Fallback to default colors if config fails to load
		cfg = &config.Config{
			ColorScheme: config.DefaultColorScheme(),
		}
	}

	// Human-readable output with lipgloss
	output := FormatShowHuman(task, cfg.ColorScheme)
	fmt.Println(output)

	return nil
}
