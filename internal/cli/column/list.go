package column

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/config"
)

// ListCmd returns the column list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project-id]",
		Short: "List columns in a project",
		Long: `List all columns in a project (in order).

Examples:
  # Using positional argument (recommended)
  paso column list 1

  # Using shorthand flag
  paso column list -p 1

  # Using git branch association
  paso column list

  # JSON output for agents
  paso column list 1 -j

  # Quiet mode (one ID per line)
  paso column list 1 -q

  # Long-form flags also supported
  paso column list --project=1 --json
`,
		Args: cobra.MaximumNArgs(1),
		RunE: runList,
	}

	// Flags
	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (IDs only)")

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
	var columnProject int

	if len(args) > 0 {
		// Priority 1: Positional argument
		columnProject, err = strconv.Atoi(args[0])
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "INVALID_PROJECT_ID",
				fmt.Sprintf("Invalid project ID: %s", args[0]),
				"Project ID must be a number")
		}
	} else {
		// Priority 2: Flag or git branch detection
		columnProject, err = cli.GetProjectIDWithCLI(cmd, cliInstance)
		if err != nil {
			return formatter.ErrorWithSuggestion(cli.ExitUsage, "NO_PROJECT",
				err.Error(),
				"Specify project ID as argument, use --project flag, or associate branch with project")
		}
	}

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, columnProject)
	if err != nil {
		return formatter.Error(cli.ExitNotFound, "PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", columnProject))
	}

	// Get columns
	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, columnProject)
	if err != nil {
		return formatter.Error(cli.ExitError, "COLUMN_FETCH_ERROR", err.Error())
	}

	// Output based on mode
	if quietMode {
		for _, col := range columns {
			fmt.Printf("%d\n", col.ID)
		}
		return nil
	}

	if jsonOutput {
		columnList := make([]map[string]any, len(columns))
		for i, col := range columns {
			columnList[i] = map[string]any{
				"id":                      col.ID,
				"name":                    col.Name,
				"project_id":              col.ProjectID,
				"holds_ready_tasks":       col.HoldsReadyTasks,
				"holds_in_progress_tasks": col.HoldsInProgressTasks,
				"holds_completed_tasks":   col.HoldsCompletedTasks,
			}
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"columns": columnList,
		})
	}

	// Human-readable output
	if len(columns) == 0 {
		fmt.Printf("No columns found in project '%s'\n", project.Name)
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
	successColor := lipgloss.Color(cfg.ColorScheme.Create)

	// Build table data
	var rows [][]string

	for i, col := range columns {
		readyMark := "-"
		if col.HoldsReadyTasks {
			readyMark = "✓"
		}
		inProgressMark := "-"
		if col.HoldsInProgressTasks {
			inProgressMark = "✓"
		}
		completedMark := "-"
		if col.HoldsCompletedTasks {
			completedMark = "✓"
		}

		rows = append(rows, []string{
			strconv.Itoa(i + 1),
			strconv.Itoa(col.ID),
			col.Name,
			readyMark,
			inProgressMark,
			completedMark,
		})
	}

	// Create table with styling
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Bold(true).Foreground(headerColor)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("#", "ID", "NAME", "READY", "IN-PROG", "DONE").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			// Checkmark columns (cols 3, 4, 5) - highlight checkmarks
			if col >= 3 && col <= 5 {
				if rows[row][col] == "✓" {
					return baseStyle.Foreground(successColor)
				}
				return baseStyle.Foreground(lipgloss.Color("240"))
			}

			return baseStyle.Foreground(normalColor)
		})

	fmt.Printf("Columns in project '%s':\n", project.Name)
	fmt.Println(t)

	return nil
}
