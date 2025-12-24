package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

type StatusBarProps struct {
	Width            int
	SearchMode       bool
	SearchQuery      string
	ConnectionStatus state.ConnectionStatus
}

// RenderStatusBar renders a status bar with left and right aligned text
// Left side: "/search-query" when searching, or connection status otherwise
// Right side: "press ? for help"
func RenderStatusBar(props StatusBarProps) string {
	var leftText string
	var leftColor string

	if props.SearchMode {
		leftText = "/" + props.SearchQuery
		leftColor = theme.Subtle
	} else {
		// Show connection status
		switch props.ConnectionStatus {
		case state.Connected:
			leftText = "● Connected"
			leftColor = "#00ff00"
		case state.Reconnecting:
			leftText = "◌ Reconnecting to daemon"
			leftColor = "#888888"
		case state.Disconnected:
			leftText = "○ No Connection To Daemon"
			leftColor = "#ffff00"
		default:
			leftText = "Paso - Task Management"
			leftColor = theme.Subtle
		}
	}

	rightText := "press ? for help"

	leftStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(leftColor))
	rightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	leftRendered := leftStyle.Render(leftText)
	rightRendered := rightStyle.Render(rightText)

	// Calculate space between left and right text
	leftWidth := lipgloss.Width(leftRendered)
	rightWidth := lipgloss.Width(rightRendered)
	gapWidth := max(props.Width-leftWidth-rightWidth, 1)

	gap := strings.Repeat(" ", gapWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}
