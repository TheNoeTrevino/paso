package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
)

// viewDeleteTaskConfirm renders the task deletion confirmation dialog
func (m Model) viewDeleteTaskConfirm() string {
	task := m.getCurrentTask()
	if task == nil {
		return ""
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(fmt.Sprintf("Delete '%s'?\n\n[y]es  [n]o", task.Title))

	return lipgloss.Place(
		m.UIState.Width(), m.UIState.Height,
		lipgloss.Center, lipgloss.Center,
		confirmBox,
	)
}
