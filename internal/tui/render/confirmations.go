package render

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
)

// ViewDeleteTaskConfirm renders the task deletion confirmation dialog
func (w *Wrapper) ViewDeleteTaskConfirm() string {
	ops := modelops.New(w.Model)
	task := ops.GetCurrentTask()
	if task == nil {
		return ""
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(fmt.Sprintf("Delete '%s'?\n\n[y]es  [n]o", task.Title))

	return lipgloss.Place(
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		confirmBox,
	)
}

// ViewDeleteColumnConfirm renders the column deletion confirmation with task count warning
func (w *Wrapper) ViewDeleteColumnConfirm() string {
	ops := modelops.New(w.Model)
	column := ops.GetCurrentColumn()
	if column == nil {
		return ""
	}

	var content string
	taskCount := w.InputState.DeleteColumnTaskCount
	if taskCount > 0 {
		content = fmt.Sprintf(
			"Delete column '%s'?\nThis will also delete %d task(s).\n\n[y]es  [n]o",
			column.Name,
			taskCount,
		)
	} else {
		content = fmt.Sprintf("Delete column '%s'?\n\n[y]es  [n]o", column.Name)
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(content)

	return lipgloss.Place(
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		confirmBox,
	)
}

// ViewDiscardConfirm renders the discard confirmation dialog with context-aware message
func (w *Wrapper) ViewDiscardConfirm() string {
	ctx := w.UiState.DiscardContext()
	if ctx == nil {
		return ""
	}

	// Use context message for personalized prompt
	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(fmt.Sprintf("%s\n\n[y]es  [n]o", ctx.Message))

	return lipgloss.Place(
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		confirmBox,
	)
}
