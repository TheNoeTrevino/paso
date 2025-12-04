package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type StatusBarProps struct {
	Width int
}

// RenderStatusBar renders a status bar with left and right aligned text
// Left side: "Paso - Task Management"
// Right side: "press ? for help"
func RenderStatusBar(props StatusBarProps) string {
	leftText := "Paso - Task Management"
	rightText := "press ? for help"

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	leftRendered := style.Render(leftText)
	rightRendered := style.Render(rightText)

	// Calculate space between left and right text
	leftWidth := lipgloss.Width(leftRendered)
	rightWidth := lipgloss.Width(rightRendered)
	gapWidth := props.Width - leftWidth - rightWidth
	if gapWidth < 1 {
		gapWidth = 1
	}

	gap := strings.Repeat(" ", gapWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}
