package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// TaskCardHeight is the fixed heght of the task card
const TaskCardHeight = 7

// RenderColumn renders a complete column with its title and tasks
// This is a pure, reusable component that composes individual task components
//
// Layout:
//
//	{Column Name} ({count})
//	▲ (if scrolled down)
//	{Task 1}
//	{Task 2}
//	...
//	▼ (if more tasks below)
//
// Parameters:
//   - column: The column to render
//   - tasks: Task summaries in this column
//   - selected: Whether this column is currently selected
//   - selectedTaskIdx: Index of selected task in this column (-1 if not this column)
//   - height: Fixed height for the column (0 for auto)
//   - scrollOffset: Index of first visible task
func RenderColumn(column *models.Column, tasks []*models.TaskSummary, selected bool, selectedTaskIdx int, height int, scrollOffset int) string {
	// Render column title with task count
	header := fmt.Sprintf("%s (%d)", column.Name, len(tasks))
	content := TitleStyle.Render(header) + "\n"

	// Render all tasks in the column or show empty state
	if len(tasks) == 0 {
		// Empty column - show helpful message
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Italic(true).
			Padding(1, 0)
		content += emptyStyle.Render("No tasks")
	} else {
		// Calculate how many tasks fit
		// Column overhead breakdown:
		// - Border + Padding: 3 lines (top border(1) + bottom padding(1) + bottom border(1))
		// - Header: 1 line (column name and count)
		// - Top indicator: 1 line (empty line or "▲ more above")
		// - Bottom indicator: 1 line ("▼ more below" when present)
		// Total: 6 lines
		const columnOverhead = 5
		availableHeight := height - columnOverhead
		maxVisibleTasks := max(availableHeight/TaskCardHeight, 1)

		// Style for indicators
		indicatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Subtle)).
			Align(lipgloss.Center)

		// Always reserve space for top indicator
		if scrollOffset > 0 {
			content += indicatorStyle.Render("▲ more above") + "\n"
		} else {
			content += "\n" // Empty line to maintain consistent spacing
		}

		// Calculate visible task range
		endIdx := min(scrollOffset+maxVisibleTasks, len(tasks))
		visibleTasks := tasks[scrollOffset:endIdx]

		// Render visible tasks (no separators - tasks are adjacent)
		for i, task := range visibleTasks {
			// Task is selected if this is the selected column and matches the actual index
			actualIdx := scrollOffset + i
			isTaskSelected := selected && actualIdx == selectedTaskIdx
			content += RenderTask(task, isTaskSelected)
		}

		// Calculate padding to push bottom indicator to the bottom.
		//
		// The height parameter is the TOTAL box height (including borders and padding).
		// ColumnStyle adds: TopBorder(1) + BottomPadding(1) + BottomBorder(1) = 3 lines
		// Therefore, available content height = height - 3
		//
		// Content lines used so far:
		// - Header: 1 line
		// - Top indicator: 1 line (empty or "▲ more above")
		// - Tasks: len(visibleTasks) * TaskCardHeight lines
		// - Bottom indicator: 1 line (if present) or 0 lines (if at end)
		//
		// We want to fill the remaining space with newlines to push the bottom
		// indicator flush to the bottom padding area.

		usedLines := 1 + 1 + (len(visibleTasks) * TaskCardHeight)

		// Determine if we need a bottom indicator
		hasBottomIndicator := endIdx < len(tasks)
		var bottomIndicatorLines int
		if hasBottomIndicator {
			bottomIndicatorLines = 2 // newline + indicator text
		} else {
			bottomIndicatorLines = 0
		}

		// Calculate remaining space.
		// Account for the 3 lines used by borders and padding (handled by lipgloss)
		contentHeight := height - 3
		remainingLines := contentHeight - usedLines - bottomIndicatorLines

		// Add padding newlines to fill space
		if remainingLines > 0 {
			content += strings.Repeat("\n", remainingLines)
		}

		// Add bottom indicator if needed (newline + indicator text = 2 lines)
		if hasBottomIndicator {
			content += "\n" + indicatorStyle.Render("▼ more below")
		}
	}

	// Apply column styling with selection highlight and fixed height
	style := ColumnStyle
	if selected {
		style = style.BorderForeground(lipgloss.Color(theme.SelectedBorder))
	}
	if height > 0 {
		// Subtract 2 for top and bottom borders since .Height() sets content area height
		style = style.Height(height - 2)
	}

	return style.Render(content)
}
