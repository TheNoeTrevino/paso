package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/thenoetrevino/paso/internal/models"
)

// RenderTask renders a single task as a card
// This is a pure, reusable component that displays task title and ID
//
// Format (as a card with border):
//
//	┌─────────────────────┐
//	│ {Task Title}        │
//	│ PASO-{ID}           │
//	└─────────────────────┘
//
// When selected is true, the task is highlighted with:
//   - Bold text
//   - Purple border color
//   - Brighter background
func RenderTask(task *models.Task, selected bool) string {
	// Format task content with title and ID
	title := lipgloss.NewStyle().Bold(true).Render(task.Title)
	id := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("PASO-%d", task.ID))

	content := fmt.Sprintf("%s\n%s", title, id)

	// Apply selection styling if this task is selected
	style := TaskStyle
	if selected {
		style = style.
			BorderForeground(lipgloss.Color("170")). // Purple border when selected
			Background(lipgloss.Color("237")).        // Lighter background when selected
			BorderStyle(lipgloss.ThickBorder())
	}

	return style.Render(content)
}

// RenderColumn renders a complete column with its title and tasks
// This is a pure, reusable component that composes individual task components
//
// Layout:
//
//	{Column Name} ({count})
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
	// Render column title with task count
	header := fmt.Sprintf("%s (%d)", column.Name, len(tasks))
	content := TitleStyle.Render(header) + "\n\n"

	// Render all tasks in the column or show empty state
	if len(tasks) == 0 {
		// Empty column - show helpful message
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Padding(1, 0)
		content += emptyStyle.Render("No tasks")
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
