package task

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

// ListInput represents the parsed input for listing tasks
type ListInput struct {
	ProjectID int
}

// ListResult represents the output of a successful list operation
type ListResult struct {
	Tasks []*models.TaskSummary
}

// TableRow represents a single row of task data for display
type TableRow struct {
	ID                  string
	Title               string
	PriorityDescription string
	TypeDescription     string
	Estimate            string
	Labels              string
	Blocked             string
	PriorityColor       string
}

// ParseListProjectID parses the project ID from arguments
func ParseListProjectID(args []string) (*ListInput, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("project ID is required")
	}

	projectID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %s", args[0])
	}

	return &ListInput{
		ProjectID: projectID,
	}, nil
}

// FlattenTasksByColumn flattens a map of tasks by column into a single slice.
// Columns are iterated in ascending ID order for deterministic output.
func FlattenTasksByColumn(tasksByColumn map[int][]*models.TaskSummary) []*models.TaskSummary {
	columnIDs := make([]int, 0, len(tasksByColumn))
	for colID := range tasksByColumn {
		columnIDs = append(columnIDs, colID)
	}
	sort.Ints(columnIDs)

	var allTasks []*models.TaskSummary
	for _, colID := range columnIDs {
		allTasks = append(allTasks, tasksByColumn[colID]...)
	}
	return allTasks
}

// FormatTasksQuiet formats tasks for quiet output (IDs only)
func FormatTasksQuiet(tasks []*models.TaskSummary) []string {
	if len(tasks) == 0 {
		return nil
	}
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = strconv.Itoa(t.ID)
	}
	return ids
}

// FormatTasksJSON generates the JSON output structure
func FormatTasksJSON(tasks []*models.TaskSummary) map[string]any {
	return map[string]any{
		"success": true,
		"tasks":   tasks,
	}
}

// BuildTableRows converts task summaries into table rows
func BuildTableRows(tasks []*models.TaskSummary) []TableRow {
	rows := make([]TableRow, len(tasks))

	for i, t := range tasks {
		// Build labels string
		labelNames := make([]string, len(t.Labels))
		for j, lbl := range t.Labels {
			labelNames[j] = lbl.Name
		}
		labelsStr := strings.Join(labelNames, ", ")
		if labelsStr == "" {
			labelsStr = "-"
		}

		// Estimate display
		estimateStr := DisplayOrDefault(t.Estimate, "-")

		// Blocked indicator
		blockedStr := ""
		if t.IsBlocked {
			blockedStr = "BLOCKED"
		}

		rows[i] = TableRow{
			ID:                  strconv.Itoa(t.ID),
			Title:               t.Title,
			PriorityDescription: t.PriorityDescription,
			TypeDescription:     t.TypeDescription,
			Estimate:            estimateStr,
			Labels:              labelsStr,
			Blocked:             blockedStr,
			PriorityColor:       t.PriorityColor,
		}
	}

	return rows
}

// RenderTasksTable creates a styled table from table rows
func RenderTasksTable(rows []TableRow, colorScheme colors.ColorScheme) string {
	// Define colors
	normalColor := lipgloss.Color(colorScheme.Normal)
	headerColor := lipgloss.Color(colorScheme.Accent)

	// Convert TableRow to string rows for lipgloss table
	stringRows := make([][]string, len(rows))
	for i, row := range rows {
		stringRows[i] = []string{
			row.ID,
			row.Title,
			row.PriorityDescription,
			row.TypeDescription,
			row.Estimate,
			row.Labels,
			row.Blocked,
		}
	}

	// Create row color map
	taskRowColors := make(map[int]string)
	for i, row := range rows {
		taskRowColors[i] = row.PriorityColor
	}

	// Create table with styling
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Bold(true).Foreground(headerColor)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers("ID", "TITLE", "PRIORITY", "TYPE", "EST", "LABELS", "").
		Rows(stringRows...).
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

			// Blocked column (col 6)
			if col == 6 && stringRows[row][6] != "" {
				return baseStyle.Bold(true).Foreground(lipgloss.Color("#FF5555"))
			}

			return baseStyle.Foreground(normalColor)
		})

	return t.String()
}
