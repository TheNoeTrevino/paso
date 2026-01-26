package project

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
	"github.com/thenoetrevino/paso/internal/git"
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

	// Get git branch info for markers
	gitInfo := git.DetectGitInfo(ctx)
	branchMarkers := make(map[string]string)

	if gitInfo.IsRepo {
		branches, err := git.ListBranches(ctx)
		if err == nil {
			for _, branch := range branches {
				branchMarkers[branch.Name] = branch.Marker
			}
		}
	}

	// Load config for color scheme
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{
			ColorScheme: config.DefaultColorScheme(),
		}
	}

	// Define colors
	greenColor := lipgloss.Color(cfg.ColorScheme.Create)  // Green for current branch
	cyanColor := lipgloss.Color("#00CED1")                // Cyan for worktree
	normalColor := lipgloss.Color(cfg.ColorScheme.Normal) // Normal text
	subtleColor := lipgloss.Color(cfg.ColorScheme.Subtle) // For dash when no branch
	headerColor := lipgloss.Color(cfg.ColorScheme.Accent) // Header color

	// Build table data
	var rows [][]string
	projectRowMarkers := make(map[int]string) // row index -> marker type

	for i, p := range projects {
		marker := "  "
		branch := "-"

		if p.GitBranch != "" {
			branch = p.GitBranch
			if m, ok := branchMarkers[p.GitBranch]; ok {
				marker = m
			}
		}

		rows = append(rows, []string{
			marker,
			strconv.Itoa(p.ID),
			p.Name,
			branch,
		})
		projectRowMarkers[i] = marker
	}

	// Create table with styling
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Bold(true).Foreground(headerColor)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("", "ID", "NAME", "BRANCH").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			// Header row
			if row == table.HeaderRow {
				return headerStyle
			}

			marker := projectRowMarkers[row]

			// Marker column (col 0) and Branch column (col 3)
			if col == 0 || col == 3 {
				switch marker {
				case "* ": // Current branch - green
					return baseStyle.Foreground(greenColor)
				case "+ ": // Worktree - cyan
					return baseStyle.Foreground(cyanColor)
				default:
					// For branch column with "-", use subtle color
					if col == 3 {
						rowData := rows[row]
						if rowData[3] == "-" {
							return baseStyle.Foreground(subtleColor)
						}
					}
					return baseStyle.Foreground(normalColor)
				}
			}

			// Other columns - normal color
			return baseStyle.Foreground(normalColor)
		})

	fmt.Println(t)

	return nil
}
