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
//	│ [label1] [label2]   │
//	└─────────────────────┘
//
// When selected is true, the task is highlighted with:
//   - Bold text
//   - Purple border color
//   - Brighter background
func RenderTask(task *models.TaskSummary, selected bool) string {
	// Format task content with title (add leading space for padding)
	title := lipgloss.NewStyle().Bold(true).Render(" 󰗴 " + task.Title)
	text := lipgloss.NewStyle().Background(lipgloss.Color(theme.TaskBg)).Render(" ")

	if selected {
		text = lipgloss.NewStyle().Background(lipgloss.Color(theme.SelectedBg)).Render(" ")
	}

	// Render label chips
	var labelChips string
	if len(task.Labels) > 0 {
		var chips []string
		for _, label := range task.Labels {
			if selected {
				chips = append(chips, components.RenderLabelChip(label, theme.SelectedBg))
			} else {
				chips = append(chips, components.RenderLabelChip(label, theme.TaskBg))
			}
		}
		labelChips = "\n " + strings.Join(chips, text)
	}

	// Render type and priority on the same line, separated by │
	var typeDisplay string
	var priorityDisplay string

	// Type display
	if task.TypeDescription != "" {
		typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
		typeDisplay = typeStyle.Render(task.TypeDescription)
	} else {
		typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
		typeDisplay = typeStyle.Render("task")
	}

	// Priority display with color
	if task.PriorityDescription != "" && task.PriorityColor != "" {
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(task.PriorityColor))
		priorityDisplay = priorityStyle.Render(task.PriorityDescription)
	} else {
		// Default to medium priority if not set
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EAB308"))
		priorityDisplay = priorityStyle.Render("medium")
	}

	// Separator
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
	separator := separatorStyle.Render(" │ ")

	// Combine type and priority
	metadataLine := "\n " + typeDisplay + separator + priorityDisplay

	content := title + metadataLine + labelChips

	// HACK: Make the selected look look like its focused
	// Maybe we just make a selected task style instead?
	style := TaskStyle
	if selected {
		style = style.
			BorderForeground(lipgloss.Color(theme.SelectedBorder)).
			BorderBackground(lipgloss.Color(theme.SelectedBg)).
			Background(lipgloss.Color(theme.SelectedBg)).
			BorderStyle(lipgloss.ThickBorder())
	}

	return style.Render(content)
}

// TaskCardHeight is the fixed height of a task card in lines.
// Since labels are no longer shown in kanban view, all tasks have consistent height:
// - Top border (1) + bottom border (1) + padding (2) + title (1) + margin (1) = 6 lines
const TaskCardHeight = 6

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
		// Column overhead: header (3) + padding (2) + borders (2) + top indicator (1) + bottom indicator (1) = 11
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

		// Render visible tasks
		var taskViews []string
		for i, task := range visibleTasks {
			// Task is selected if this is the selected column and matches the actual index
			actualIdx := scrollOffset + i
			isTaskSelected := selected && actualIdx == selectedTaskIdx
			taskViews = append(taskViews, RenderTask(task, isTaskSelected))
		}

		// Join tasks with newlines
		content += strings.Join(taskViews, "\n")

		// Always reserve space for bottom indicator (sticky to bottom)
		content += "\n"
		if endIdx < len(tasks) {
			content += indicatorStyle.Render("▼ more below")
		}
		// Note: If no more tasks, we still added "\n" above, leaving empty space at bottom
	}

	// Apply column styling with selection highlight and fixed height
	style := ColumnStyle
	if selected {
		style = style.BorderForeground(lipgloss.Color(theme.SelectedBorder))
	}
	if height > 0 {
		style = style.Height(height)
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
