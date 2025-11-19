package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/thenoetrevino/paso/internal/models"
)

// RenderTask renders a single task as a formatted string
// This is a pure, reusable component that displays task title and ID
//
// Format:
//
//	▸ {Task Title}
//	  PASO-{ID}
//
// When selected is true, the task is highlighted with:
//   - Bold text
//   - Purple foreground color
//   - Left border with thick style
func RenderTask(task *models.Task, selected bool) string {
	content := fmt.Sprintf("▸ %s\n  PASO-%d", task.Title, task.ID)

	// Apply selection styling if this task is selected
	style := TaskStyle
	if selected {
		style = style.
			Bold(true).
			Foreground(lipgloss.Color("170")).
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("170"))
	}

	return style.Render(content)
}

// RenderColumn renders a complete column with its title and tasks
// This is a pure, reusable component that composes individual task components
//
// Layout:
//
//	{Column Name}
//
//	{Task 1}
//	{Task 2}
//	...
//
// Parameters:
//   - column: The column to render
//   - tasks: Tasks in this column
//   - selected: Whether this column is currently selected
//   - selectedTaskIdx: Index of selected task in this column (-1 if not this column)
func RenderColumn(column *models.Column, tasks []*models.Task, selected bool, selectedTaskIdx int) string {
	// Render column title
	content := TitleStyle.Render(column.Name) + "\n\n"

	// Render all tasks in the column or show empty state
	if len(tasks) == 0 {
		// Empty column - no tasks to render
		content += ""
	} else {
		var taskViews []string
		for i, task := range tasks {
			// Task is selected if this is the selected column and matches the index
			isTaskSelected := selected && i == selectedTaskIdx
			taskViews = append(taskViews, RenderTask(task, isTaskSelected))
		}

		// Join tasks with newlines
		content += strings.Join(taskViews, "\n")
	}

	// Apply column styling with selection highlight
	style := ColumnStyle
	if selected {
		style = style.BorderForeground(lipgloss.Color("170"))
	}

	return style.Render(content)
}
