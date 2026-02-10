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
	"github.com/thenoetrevino/paso/internal/config"
)

// ListCmd returns the label list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project-id]",
		Short: "List labels in a project",
		Long: `List all labels in a project.

Examples:
  # Using positional argument (recommended)
  paso label list 1

  # Using shorthand flag
  paso label list -p 1

  # Using git branch association
  paso label list

  # JSON output for agents
  paso label list 1 -j

  # Quiet mode (one ID per line)
  paso label list 1 -q

  # Long-form flags also supported
  paso label list --project=1 --json
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
	var labelProject int

	if len(args) > 0 {
		// Priority 1: Positional argument
		labelProject, err = strconv.Atoi(args[0])
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
		labelProject, err = cli.GetProjectIDWithCLI(cmd, cliInstance)
		if err != nil {
			if fmtErr := formatter.ErrorWithSuggestion("NO_PROJECT",
				err.Error(),
				"Specify project ID as argument, use --project flag, or associate branch with project"); fmtErr != nil {
				slog.Error("failed to format error message", "error", fmtErr)
			}
			os.Exit(cli.ExitUsage)
		}
	}

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", labelProject)); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get labels
	labels, err := cliInstance.App.LabelService.GetLabelsByProject(ctx, labelProject)
	if err != nil {
		if fmtErr := formatter.Error("LABEL_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to format error message", "error", fmtErr)
		}
		os.Exit(cli.ExitError)
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

	// Build table data and track colors per row
	var rows [][]string
	rowColors := make(map[int]string)

	for i, lbl := range labels {
		rows = append(rows, []string{
			strconv.Itoa(lbl.ID),
			lbl.Name,
			lbl.Color,
		})
		rowColors[i] = lbl.Color
	}

	// Create table with styling
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Bold(true).Foreground(headerColor)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("ID", "NAME", "COLOR").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			// Color column (col 2) - use the label's color
			if col == 2 {
				return baseStyle.Foreground(lipgloss.Color(rowColors[row]))
			}

			return baseStyle.Foreground(normalColor)
		})

	fmt.Printf("Labels in project '%s':\n", project.Name)
	fmt.Println(t)

	return nil
}
