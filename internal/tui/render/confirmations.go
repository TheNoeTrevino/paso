package render

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
)

// ViewDeleteTaskConfirm renders the task deletion confirmation dialog
func ViewDeleteTaskConfirm(m *tui.Model) string {
	task := modelops.GetCurrentTask(m)
	if task == nil {
		return ""
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(fmt.Sprintf("Delete '%s'?\n\n[y]es  [n]o", task.Title))

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		confirmBox,
	)
}

// ViewDeleteColumnConfirm renders the column deletion confirmation with task count warning
func ViewDeleteColumnConfirm(m *tui.Model) string {
	column := modelops.GetCurrentColumn(m)
	if column == nil {
		return ""
	}

	var content string
	taskCount := m.InputState.DeleteColumnTaskCount
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
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		confirmBox,
	)
}

// ViewDiscardConfirm renders the discard confirmation dialog with context-aware message
func ViewDiscardConfirm(m *tui.Model) string {
	ctx := m.UiState.DiscardContext()
	if ctx == nil {
		return ""
	}

	// Use context message for personalized prompt
	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(fmt.Sprintf("%s\n\n[y]es  [n]o", ctx.Message))

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		confirmBox,
	)
}
