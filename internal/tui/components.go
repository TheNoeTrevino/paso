package tui

import (
	"fmt"
	"strings"

	"github.com/thenoetrevino/paso/internal/models"
)

// RenderTask renders a single task as a formatted string
// This is a pure, reusable component that displays task title and ID
//
// Format:
//   ▸ {Task Title}
//     PASO-{ID}
func RenderTask(task *models.Task) string {
	content := fmt.Sprintf("▸ %s\n  PASO-%d", task.Title, task.ID)
	return TaskStyle.Render(content)
}

// RenderColumn renders a complete column with its title and tasks
// This is a pure, reusable component that composes individual task components
//
// Layout:
//   {Column Name}
//
//   {Task 1}
//   {Task 2}
//   ...
func RenderColumn(column *models.Column, tasks []*models.Task) string {
	// Render column title
	content := TitleStyle.Render(column.Name) + "\n\n"

	// Render all tasks in the column
	var taskViews []string
	for _, task := range tasks {
		taskViews = append(taskViews, RenderTask(task))
	}

	// Join tasks with newlines
	content += strings.Join(taskViews, "\n")

	// Apply column styling and return
	return ColumnStyle.Render(content)
}
