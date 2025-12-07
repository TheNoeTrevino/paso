// Package tui implements the terminal user interface for the application.
// This package contains rendering, input handling, and state management for
// the interactive kanban board interface.
package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderTaskPicker renders the task picker popup in a GitHub-style interface with checkboxes.
// This picker is used for both parent and child issue selection.
//
// The picker displays tasks in PROJ-123 format with titles, allows filtering by typing,
// and indicates selected tasks with checkboxes. The cursor position determines which
// task is currently highlighted.
//
// Parameters:
//   - filteredItems: Pre-filtered list of tasks to display. The caller must filter using
//     TaskPickerState.GetFilteredItems() before calling this function.
//   - cursorIdx: Zero-based index of the currently highlighted item in filteredItems.
//   - filterText: Current filter text entered by the user. If empty, shows placeholder text.
//   - pickerType: Title displayed at the top ("Parent Issues" or "Child Issues").
//   - width: Maximum width for the picker content in characters.
//   - height: Maximum height for the picker content in rows (currently unused but reserved for scrolling).
//
// Returns:
//   - A formatted string ready for rendering in the terminal.
func RenderTaskPicker(
	filteredItems []state.TaskPickerItem,
	cursorIdx int,
	filterText string,
	pickerType string, // "Parent Issues" or "Child Issues"
	width int,
	height int,
) string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	content.WriteString(titleStyle.Render(pickerType) + "\n\n")

	// Filter input
	filterStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Subtle)).
		Padding(0, 1).
		Width(width - 8)

	filterDisplay := filterText
	if filterDisplay == "" {
		filterDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("Filter tasks...")
	}
	content.WriteString(filterStyle.Render(filterDisplay) + "\n\n")

	// Task list
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	// Render each task item (already filtered by caller)
	for i, item := range filteredItems {
		// Checkbox indicator
		checkbox := "[ ]"
		if item.Selected {
			checkbox = "[x]"
		}

		// Task identifier (PROJ-123 format)
		taskID := fmt.Sprintf("%s-%d", item.TaskRef.ProjectName, item.TaskRef.TicketNumber)

		// Truncate title if needed to fit width
		maxLen := width - 12
		title := item.TaskRef.Title
		if len(title) > maxLen {
			title = title[:maxLen-3] + "..."
		}

		// Build line: [x] PROJ-12: Task title
		line := fmt.Sprintf("%s %s: %s", checkbox, taskID, title)

		// Apply cursor styling
		if i == cursorIdx {
			content.WriteString(selectedStyle.Render("> " + line) + "\n")
		} else {
			content.WriteString(normalStyle.Render("  " + line) + "\n")
		}
	}

	// Show "no results" if filter matched nothing
	if len(filteredItems) == 0 && filterText != "" {
		content.WriteString(dimStyle.Render("  No tasks match \"" + filterText + "\"") + "\n")
	}

	// Show "no tasks" if no items exist (and no filter)
	if len(filteredItems) == 0 && filterText == "" {
		content.WriteString(dimStyle.Render("  No other tasks in this project") + "\n")
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(dimStyle.Render("Enter: toggle  Esc: close") + "\n")

	return content.String()
}
