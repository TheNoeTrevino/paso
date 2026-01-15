package label

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
	"github.com/thenoetrevino/paso/internal/models"
)

// ListCmd returns the label list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List labels in a project",
		Long: `List all labels in a project.

Examples:
  # Human-readable list
  paso label list --project=1

  # JSON output for agents
  paso label list --project=1 --json

  # Quiet mode (one ID per line)
  paso label list --project=1 --quiet
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
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to closing CLI", "error", err)
		}
	}()

	// Get project ID from flag or git branch
	labelProject, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("NO_PROJECT",
			err.Error(),
			"Use --project flag or create a project associated with this git branch"); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitUsage)
	}

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", labelProject)); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get labels
	labels, err := cliInstance.App.LabelService.GetLabelsByProject(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	// Output based on mode
	if quietMode {
		for _, lbl := range labels {
			fmt.Printf("%d\n", lbl.ID)
		}
		return nil
	}

	if jsonOutput {
		labelList := make([]map[string]any, len(labels))
		for i, lbl := range labels {
			labelList[i] = map[string]any{
				"id":         lbl.ID,
				"name":       lbl.Name,
				"color":      lbl.Color,
				"project_id": lbl.ProjectID,
			}
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success": true,
			"labels":  labelList,
		})
	}

	if len(labels) == 0 {
		fmt.Printf("No labels found in project '%s'\n", project.Name)
		return nil
	}

	fmt.Printf("Labels in project '%s':\n", project.Name)
	fmt.Println(renderLabelTable(labels))
	return nil
}

func renderLabelTable(labels []*models.Label) string {
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Bold(true).Foreground(lipgloss.Color("252"))

	var rows [][]string
	for _, lbl := range labels {
		rows = append(rows, []string{
			strconv.Itoa(lbl.ID),
			lbl.Name,
			lbl.Color,
		})
	}

	tbl := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("ID", "NAME", "COLOR").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			label := labels[row]
			even := row%2 == 0

			if col == 2 && label.Color != "" {
				return baseStyle.Foreground(lipgloss.Color(label.Color))
			}

			if even {
				return baseStyle.Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Foreground(lipgloss.Color("252"))
		})

	return tbl.String()
}
