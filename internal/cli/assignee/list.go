package assignee

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

// ListCmd returns the assignee list subcommand
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all assignees",
		Long: `List all assignees in the database.

Examples:
  # Human-readable list
  paso assignee list

  # JSON output for agents
  paso assignee list -j

  # Quiet mode (one ID per line)
  paso assignee list -q
`,
		RunE: runList,
	}

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

	assignees, err := cliInstance.App.AssigneeService.List(ctx)
	if err != nil {
		if fmtErr := formatter.Error("ASSIGNEE_FETCH_ERROR", err.Error()); fmtErr != nil {
			slog.Error("failed to formatting error message", "error", fmtErr)
		}
		return err
	}

	if quietMode {
		for _, assignee := range assignees {
			fmt.Printf("%d\n", assignee.ID)
		}
		return nil
	}

	if jsonOutput {
		// TODO: can we type this more strongly?
		assigneeList := make([]map[string]any, len(assignees))
		for i, assignee := range assignees {
			// TODO: this would be the type of above
			assigneeList[i] = map[string]any{
				"id":   assignee.ID,
				"name": assignee.Name,
			}
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"success":   true,
			"assignees": assigneeList,
		})
	}

	if len(assignees) == 0 {
		fmt.Println("No assignees found")
		return nil
	}

	fmt.Println("Assignees:")
	// QUESTION: Cant we run .String() directly? Do we need to fmt print this?
	fmt.Println(renderAssigneeTable(assignees))
	return nil
}

func renderAssigneeTable(assignees []*models.Assignee) string {
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Bold(true).Foreground(lipgloss.Color("252"))

	var rows [][]string
	for _, assignee := range assignees {
		rows = append(rows, []string{
			strconv.Itoa(assignee.ID),
			assignee.Name,
		})
	}

	tbl := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("ID", "NAME").
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			even := row%2 == 0
			if even {
				return baseStyle.Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Foreground(lipgloss.Color("252"))
		})

	return tbl.String()
}
