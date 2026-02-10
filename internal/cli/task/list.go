package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/models"
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
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
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
			if fmtErr := formatter.ErrorWithSuggestion("INVALID_PROJECT_ID",
				fmt.Sprintf("Invalid project ID: %s", args[0]),
				"Project ID must be a number"); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitUsage)
		}
	} else {
		// Priority 2: Flag or git branch detection
		taskProject, err = cli.GetProjectIDWithCLI(cmd, cliInstance)
		if err != nil {
			if fmtErr := formatter.ErrorWithSuggestion("NO_PROJECT",
				err.Error(),
				"Specify project ID as argument, use --project flag, or associate branch with project"); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitUsage)
		}
	}

	// Get tasks (returns map[columnID][]*TaskSummary)
	tasksByColumn, err := cliInstance.App.TaskService.GetTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
	}

	// Flatten tasks from all columns
	var allTasks []*models.TaskSummary
	for _, columnTasks := range tasksByColumn {
		allTasks = append(allTasks, columnTasks...)
	}

	// Output in appropriate format
	if quietMode {
		// Just print IDs
		for _, t := range allTasks {
			fmt.Printf("%d\n", t.ID)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"tasks":   allTasks,
		})
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

	// Define colors
	normalColor := lipgloss.Color(cfg.ColorScheme.Normal)
	headerColor := lipgloss.Color(cfg.ColorScheme.Accent)

	// Build table data
	var rows [][]string
	taskRowColors := make(map[int]string) // row index -> priority color

	for i, t := range allTasks {
		// Build labels string
		labelNames := make([]string, len(t.Labels))
		for j, lbl := range t.Labels {
			labelNames[j] = lbl.Name
		}
		labelsStr := strings.Join(labelNames, ", ")
		if labelsStr == "" {
			labelsStr = "-"
		}

		// Blocked indicator
		blockedStr := ""
		if t.IsBlocked {
			blockedStr = "BLOCKED"
		}

		rows = append(rows, []string{
			strconv.Itoa(t.ID),
			t.Title,
			t.PriorityDescription,
			t.TypeDescription,
			labelsStr,
			blockedStr,
		})
		taskRowColors[i] = t.PriorityColor
	}

	// Create table with styling
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Bold(true).Foreground(headerColor)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("ID", "TITLE", "PRIORITY", "TYPE", "LABELS", "").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			// Priority column (col 2) - use task priority color
			if col == 2 {
				priorityColor := taskRowColors[row]
				if priorityColor != "" {
					return baseStyle.Foreground(lipgloss.Color(priorityColor))
				}
			}

			// Blocked column (col 5)
			if col == 5 && rows[row][5] != "" {
				return baseStyle.Bold(true).Foreground(lipgloss.Color("#FF5555"))
			}

			return baseStyle.Foreground(normalColor)
		})

	fmt.Printf("Found %d tasks:\n", len(allTasks))
	fmt.Println(t)

	return nil
}

func renderLabels(labels []*models.Label) string {
	if len(labels) == 0 {
		return ""
	}
	var parts []string
	for _, l := range labels {
		styled := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(l.Color)).
			Render("[" + l.Name + "]")
		parts = append(parts, styled)
	}
	return strings.Join(parts, " ")
}
