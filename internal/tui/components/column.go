package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderColumn renders a complete column with its title and tasks
// Dynamically renders amount of tasks based on height and scroll offset
//
// Layout:
//
//	╭──────────────────────────╮
//	│ {Column Name} ({count})  │
//	│ ▲ (if scrolled down)     │
//	│ {Task 1}                 │
//	│ {Task 2}                 │
//	│ ▼ (if more tasks below)  │
//	╰──────────────────────────╯
//
// Parameters:
//   - column: The column to render
//   - tasks: Task summaries in this column
//   - selected: Whether this column is currently selected
//   - selectedTaskIdx: Index of selected task in this column (-1 if not this column)
//   - height: Fixed height for the column (0 for auto)
//   - scrollOffset: Index of first visible task
func RenderColumn(
	column *models.Column,
	tasks []*models.TaskSummary,
	selected bool,
	selectedTaskIdx int,
	height int,
	scrollOffset int,
) string {
	header := renderColumnHeader(column, len(tasks))

	if len(tasks) == 0 {
		content := renderEmptyColumnContent(header)
		return applyColumnStyle(content, selected, height)
	}

	content := renderColumnWithTasksContent(header, tasks, selected, selectedTaskIdx, height, scrollOffset)
	return applyColumnStyle(content, selected, height)
}

// renderColumnHeader formats the column title with task count
func renderColumnHeader(column *models.Column, taskCount int) string {
	header := fmt.Sprintf("%s (%d)", column.Name, taskCount)
	return TitleStyle.Render(header)
}

// renderScrollIndicator renders a centered scroll indicator or empty line for spacing consistently
// ▲ more above or ▼ more below in this case
func renderScrollIndicator(show bool, text string) string {
	if !show {
		return "\n" // Empty line to maintain consistent spacing
	}
	return IndicatorStyle.Render(text) + "\n"
}

// renderEmptyColumnContent renders the empty state
func renderEmptyColumnContent(header string) string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)

	content := header + "\n" + "\n"

	content += emptyStyle.Render("No tasks")

	return content
}

// renderColumnWithTasksContent renders tasks with scroll indicators and padding
func renderColumnWithTasksContent(
	header string,
	tasks []*models.TaskSummary,
	selected bool,
	selectedTaskIdx int,
	height int,
	scrollOffset int,
) string {
	content := header + "\n"

	columnOverhead := columnBorderOverhead + headerLines + topIndicatorLines
	availableHeight := height - columnOverhead
	maxVisibleTasks := max(availableHeight/TaskCardHeight, 1)

	content += renderScrollIndicator(scrollOffset > 0, "▲ more above")

	endIdx := min(scrollOffset+maxVisibleTasks, len(tasks))
	visibleTasks := tasks[scrollOffset:endIdx]

	for i, task := range visibleTasks {
		actualIdx := scrollOffset + i
		isTaskSelected := selected && actualIdx == selectedTaskIdx
		content += RenderTask(task, isTaskSelected)
	}

	showBottomIndicator := endIdx < len(tasks)
	content += strings.TrimRight(renderScrollIndicator(showBottomIndicator, "▼ more below"), "\n")

	return content
}

// applyColumnStyle applies border, selection highlighting, and height to content
func applyColumnStyle(content string, selected bool, height int) string {
	style := ColumnStyle

	if selected {
		style = style.BorderForeground(lipgloss.Color(theme.SelectedBorder))
	}

	style = style.Height(max(height, 0))

	return style.Render(content)
}
