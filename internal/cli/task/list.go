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
)

// ListCmd returns the task list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project-id]",
		Short: "List tasks",
		Long: `List all tasks in a project.

Examples:
  # Using positional argument (recommended)
  paso task list 2

  # Using shorthand flag
  paso task list -p 2

  # Using git branch association
  paso task list

  # JSON output
  paso task list 2 -j

  # Long-form flags also supported
  paso task list --project=2 --json
`,
		Args: cobra.MaximumNArgs(1),
		RunE: runList,
	}

	// Flags
	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI first
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return formatter.Error(cli.ExitError, "INITIALIZATION_ERROR", err.Error())
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Get project ID with precedence: positional arg > flag > git detection
	var taskProject int

	if len(args) > 0 {
		// Priority 1: Positional argument
		taskProject, err = strconv.Atoi(args[0])
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "INVALID_PROJECT_ID",
				fmt.Sprintf("Invalid project ID: %s", args[0]),
				"Project ID must be a number")
		}
	} else {
		// Priority 2: Flag or git branch detection
		taskProject, err = cli.GetProjectIDWithCLI(cmd, cliInstance)
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "NO_PROJECT",
				err.Error(),
				"Specify project ID as argument, use --project flag, or associate branch with project")
		}
	}

	// Get tasks (returns map[columnID][]*TaskSummary)
	tasksByColumn, err := cliInstance.App.TaskService.GetTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		return formatter.Error(cli.ExitError, "TASK_FETCH_ERROR", err.Error())
	}

	// Flatten tasks from all columns
	allTasks := FlattenTasksByColumn(tasksByColumn)

	// Output in appropriate format
	if quietMode {
		// Just print IDs
		for _, id := range FormatTasksQuiet(allTasks) {
			fmt.Println(id)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(FormatTasksJSON(allTasks))
	}

	// Human-readable output
	if len(allTasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	// Load config for color scheme
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{
			ColorScheme: config.DefaultColorScheme(),
		}
	}

	// Build and render table
	tableRows := BuildTableRows(allTasks)
	renderedTable := RenderTasksTable(tableRows, cfg.ColorScheme)

	fmt.Printf("Found %d tasks:\n", len(allTasks))
	fmt.Println(renderedTable)

	return nil
}
