package project

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
)

// ListCmd returns the project list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Long:  "List all projects with their details.",
		RunE:  runList,
	}

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

	// Initialize CLI
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

	// Get all projects
	projects, err := cliInstance.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		if fmtErr := formatter.Error("PROJECT_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	// Output in appropriate format
	if quietMode {
		// Just print IDs (one per line)
		for _, p := range projects {
			fmt.Printf("%d\n", p.ID)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":  true,
			"projects": projects,
		})
	}

	// Human-readable output
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Bold(true).Foreground(lipgloss.Color("252"))

	var rows [][]string
	for _, p := range projects {
		desc := p.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		rows = append(rows, []string{
			strconv.Itoa(p.ID),
			p.Name,
			desc,
			p.GitBranch,
		})
	}

	tbl := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("ID", "NAME", "DESCRIPTION", "GIT BRANCH").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			project := projects[row]
			even := row%2 == 0

			if col == 3 {
				return baseStyle.Foreground(lipgloss.Color(branchColor(project.GitBranch)))
			}

			if even {
				return baseStyle.Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Foreground(lipgloss.Color("252"))
		})

	fmt.Println(tbl)
	return nil
}

func branchColor(branch string) string {
	switch {
	case strings.HasPrefix(branch, "feature/"):
		return "#00E2C7"
	case strings.HasPrefix(branch, "bugfix/") || strings.HasPrefix(branch, "fix/"):
		return "#FF7698"
	case strings.HasPrefix(branch, "release/"):
		return "#01BE85"
	case branch == "main" || branch == "master":
		return "#929292"
	default:
		return "252"
	}
}
