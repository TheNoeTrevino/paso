package column

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
	"github.com/thenoetrevino/paso/internal/models"
)

// ListCmd returns the column list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List columns in a project",
		Long: `List all columns in a project (in order).

Examples:
  # Human-readable list
  paso column list --project=1

  # JSON output for agents
  paso column list --project=1 --json

  # Quiet mode (one ID per line)
  paso column list --project=1 --quiet
`,
		RunE: runList,
	}

	// Flags
	cmd.Flags().Int("project", 0, "Project ID (uses git branch association if not specified)")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (IDs only)")

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
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Get project ID from flag or git branch
	columnProject, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("NO_PROJECT",
			err.Error(),
			"Use --project flag or create a project associated with this git branch"); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		return err
	}

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, columnProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", columnProject)); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		return fmt.Errorf("project %d not found", columnProject)
	}

	// Get columns
	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, columnProject)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		return err
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

	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Bold(true).Foreground(lipgloss.Color("252"))

	var rows [][]string
	for _, col := range columns {
		rows = append(rows, []string{
			strconv.Itoa(col.ID),
			col.Name,
			renderTypeFlags(col),
		})
	}

	tbl := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("ID", "NAME", "TYPE").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			even := row%2 == 0
			if even {
				return baseStyle.Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Foreground(lipgloss.Color("252"))
		})

	fmt.Println(tbl)
	return nil
}

func renderTypeFlags(col *models.Column) string {
	var parts []string

	if col.HoldsReadyTasks {
		styled := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E2C7")).
			Render("[READY]")
		parts = append(parts, styled)
	}
	if col.HoldsInProgressTasks {
		styled := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFAB40")).
			Render("[IN-PROGRESS]")
		parts = append(parts, styled)
	}
	if col.HoldsCompletedTasks {
		styled := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#64B5F6")).
			Render("[COMPLETED]")
		parts = append(parts, styled)
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}
