package renderers

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// ListViewRow represents a single row in the list view
type ListViewRow struct {
	Task       *models.TaskSummary
	ColumnName string
	ColumnID   int
}

// RenderListView renders the task list as a table
// Parameters:
//   - rows: list of tasks with their column names
//   - selectedIdx: index of selected row
//   - scrollOffset: vertical scroll offset
//   - sortField: current sort field
//   - sortOrder: current sort order
//   - width: available width
//   - height: available height
func RenderListView(
	rows []ListViewRow,
	selectedIdx int,
	scrollOffset int,
	sortField state.SortField,
	sortOrder state.SortOrder,
	width int,
	height int,
) string {
	var output strings.Builder

	// Calculate column widths (70% title, 30% status)
	// Reserve space for selection indicator (2 chars), padding (4 chars), and borders
	const reservedWidth = 8
	availableWidth := width - reservedWidth
	titleWidth := int(float64(availableWidth) * 0.7)
	statusWidth := availableWidth - titleWidth

	// Ensure minimum widths
	if titleWidth < 10 {
		titleWidth = 10
	}
	if statusWidth < 10 {
		statusWidth = 10
	}

	// Create header style
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))

	// Build header with sort indicators
	titleHeader := "Title" + getSortIndicator(state.SortByTitle, sortField, sortOrder)
	statusHeader := "Status" + getSortIndicator(state.SortByStatus, sortField, sortOrder)

	titleHeaderPadded := truncateString(titleHeader, titleWidth)
	statusHeaderPadded := truncateString(statusHeader, statusWidth)

	// Pad headers to full width
	titleHeaderPadded = titleHeaderPadded + strings.Repeat(" ", titleWidth-len(titleHeaderPadded))
	statusHeaderPadded = statusHeaderPadded + strings.Repeat(" ", statusWidth-len(statusHeaderPadded))

	header := fmt.Sprintf("  %s  %s", titleHeaderPadded, statusHeaderPadded)
	output.WriteString(headerStyle.Render(header))
	output.WriteString("\n")

	// Add separator line
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
	separator := strings.Repeat("─", width-2)
	output.WriteString(separatorStyle.Render("  " + separator))
	output.WriteString("\n")

	// Add up indicator if scrolled down
	if scrollOffset > 0 {
		upIndicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Render("  ▲ more above")
		output.WriteString(upIndicator)
		output.WriteString("\n")
	}

	// Calculate visible rows
	// Reserve space for header (2 lines), help text (2 lines), scroll indicators (up to 2 lines)
	const reservedHeight = 6
	visibleHeight := max(height-reservedHeight, 1)

	// Calculate which rows to display
	endIdx := min(scrollOffset+visibleHeight, len(rows))

	// Render visible rows
	if len(rows) == 0 {
		// Empty state
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Italic(true)
		emptyMsg := "  No tasks to display"
		output.WriteString(emptyStyle.Render(emptyMsg))
		output.WriteString("\n")
	} else {
		for i := scrollOffset; i < endIdx; i++ {
			row := rows[i]
			isSelected := i == selectedIdx

			// Truncate title and status to fit columns
			title := truncateString(row.Task.Title, titleWidth)
			status := truncateString(row.ColumnName, statusWidth)

			// Pad to full width
			title = title + strings.Repeat(" ", titleWidth-len(title))
			status = status + strings.Repeat(" ", statusWidth-len(status))

			// Build row content
			var rowContent string
			if isSelected {
				rowContent = fmt.Sprintf("> %s  %s", title, status)
			} else {
				rowContent = fmt.Sprintf("  %s  %s", title, status)
			}

			// Apply styling
			var rowStyle lipgloss.Style
			if isSelected {
				rowStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(theme.Highlight)).
					Bold(true)
			} else {
				rowStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(theme.Normal))
			}

			output.WriteString(rowStyle.Render(rowContent))
			output.WriteString("\n")
		}
	}

	// Add down indicator if more rows below
	if len(rows) > 0 && endIdx < len(rows) {
		output.WriteString("\n")
		downIndicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Render("  ▼ more below")
		output.WriteString(downIndicator)
	}

	// Add help text at bottom
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)
	helpText := "v: kanban  s: status  S: sort  j/k: navigate"
	output.WriteString("\n")
	output.WriteString(helpStyle.Render("  " + helpText))

	return output.String()
}

// truncateString truncates a string to maxLen with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// getSortIndicator returns the sort arrow for the given field
func getSortIndicator(field state.SortField, currentField state.SortField, order state.SortOrder) string {
	if field != currentField {
		return ""
	}
	if order == state.SortAsc {
		return " ▲"
	}
	return " ▼"
}
