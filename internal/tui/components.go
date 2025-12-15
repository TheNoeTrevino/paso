package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderTabs renders a tab bar with the given tab names
// selectedIdx indicates which tab is active (0-indexed)
// width is the total width to fill with the tab gap
func RenderTabs(tabs []string, selectedIdx int, width int) string {
	var renderedTabs []string

	for i, tabName := range tabs {
		if i == selectedIdx {
			renderedTabs = append(renderedTabs, ActiveTabStyle.Render(tabName))
		} else {
			renderedTabs = append(renderedTabs, TabStyle.Render(tabName))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Fill remaining space with gap
	gapWidth := max(width-lipgloss.Width(row)-2, 0)
	gap := TabGapStyle.Render(strings.Repeat(" ", gapWidth))

	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
}

// RenderTask renders a single task as a card
// This is a pure, reusable component that displays task title and labels
//
// Format (as a card with border):
//
//	┌─────────────────────┐
//	│ {Task Title}        │
//	│ type | priority     │
//	│ [label1] [label2]   │
//	└─────────────────────┘
//
// All three content lines are ALWAYS displayed to maintain consistent card height.
// When selected is true, the task is highlighted with:
//   - Bold text
//   - Purple border color
//   - Brighter background
func RenderTask(task *models.TaskSummary, selected bool) string {
	var bg string
	if selected {
		bg = theme.SelectedBg
	} else {
		bg = theme.TaskBg
	}

	// Format task content with title (add leading space for padding)
	title := lipgloss.NewStyle().Bold(true).Render(" 󰗴 " + task.Title)
	text := lipgloss.NewStyle().Background(lipgloss.Color(bg)).Render(" ")

	// Render type and priority on the same line, separated by │
	var typeDisplay string
	var priorityDisplay string

	// Type display - always show, use placeholder if missing
	if task.TypeDescription != "" {
		typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
		typeDisplay = typeStyle.Render(task.TypeDescription)
	} else {
		typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Italic(true)
		typeDisplay = typeStyle.Render("no type")
	}

	// Priority display with color - always show, use placeholder if missing
	if task.PriorityDescription != "" && task.PriorityColor != "" {
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(task.PriorityColor)).Background(lipgloss.Color(bg))
		priorityDisplay = priorityStyle.Render(task.PriorityDescription)
	} else {
		// Default placeholder if not set
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg)).Italic(true)
		priorityDisplay = priorityStyle.Render("no priority")
	}

	// Separator
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg))
	separator := separatorStyle.Render(" │ ")

	// Combine type and priority - always include this line
	metadataLine := "\n " + typeDisplay + separator + priorityDisplay

	// Render label chips - ALWAYS include this line even if empty to maintain fixed height
	var labelChips string
	if len(task.Labels) > 0 {
		var chips []string
		for _, label := range task.Labels {
			chips = append(chips, components.RenderLabelChip(label, bg))
		}
		labelChips = "\n " + strings.Join(chips, text)
	} else {
		// place holder for no labels
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Background(lipgloss.Color(bg)).Italic(true)
		labelChips = "\n " + emptyStyle.Render("no labels")
	}

	content := title + metadataLine + labelChips

	style := TaskStyle.
		BorderForeground(lipgloss.Color(theme.SelectedBorder)).
		BorderBackground(lipgloss.Color(bg)).
		Background(lipgloss.Color(bg)).
		BorderStyle(lipgloss.ThickBorder())

	return style.Render(content)
}

// TaskCardHeight is the fixed height of a task card in lines.
// With the new consistent 3-line content (title, type|priority, labels),
// all tasks have consistent height:
// - Top border (1) + bottom border (1) + padding (2) + title (1) + metadata (1) + labels (1) = 7 lines
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

// ScrollIndicators holds the left and right scroll arrow indicators
type ScrollIndicators struct {
	Left  string
	Right string
}

// GetScrollIndicators returns the appropriate scroll arrows based on viewport position
func GetScrollIndicators(viewportOffset, viewportSize, columnCount int) ScrollIndicators {
	return ScrollIndicators{
		Left:  getLeftArrow(viewportOffset),
		Right: getRightArrow(viewportOffset, viewportSize, columnCount),
	}
}

// getLeftArrow returns "◀" if there are columns to the left, otherwise space
func getLeftArrow(viewportOffset int) string {
	if viewportOffset > 0 {
		return "◀"
	}
	return " "
}

// getRightArrow returns "▶" if there are columns to the right, otherwise space
func getRightArrow(viewportOffset, viewportSize, columnCount int) string {
	if viewportOffset+viewportSize < columnCount {
		return "▶"
	}
	return " "
}
