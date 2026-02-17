// Package components contains reusable TUI components
package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

type StatusBarProps struct {
	Width            int
	ConnectionStatus state.ConnectionStatus
	DatabaseName     string // Current database connection name (e.g., "Local", "Production")
	Tip              string
}

// RenderStatusBar renders a status bar with left and right aligned text
// Left side: connection status
// Middle: tip (if available)
// Right side: "press ? for help"
//
// Layout:
//
//	┌─────────────────────────────────────────────────────────┐
//	│ ● Connected            Tip: ...              ? for help  │
//	└─────────────────────────────────────────────────────────┘
func RenderStatusBar(props StatusBarProps) string {
	var leftText string
	var leftColor string

	// Always show connection status
	// Note: Connection status colors are intentionally hardcoded
	// as they follow standard conventions (green=connected, yellow=warning, gray=reconnecting)
	switch props.ConnectionStatus {
	case state.Connected:
		leftText = "● Connected"
		leftColor = "#00ff00" // Green
	case state.Reconnecting:
		leftText = "◌ Reconnecting to daemon"
		leftColor = "#888888" // Gray
	case state.Disconnected:
		leftText = "○ No Connection To Daemon"
		leftColor = "#ffff00" // Yellow
	default:
		leftText = "Paso - Task Management"
		leftColor = theme.Subtle
	}

	// Append database name if not using local SQLite (Local is the default, so we skip it)
	if props.DatabaseName != "" && props.DatabaseName != "Local" {
		leftText = "[" + props.DatabaseName + "] " + leftText
	}

	rightText := "? for help"

	leftStyle := StatusBarStyle.Foreground(lipgloss.Color(leftColor))
	rightStyle := StatusBarStyle
	tipStyle := StatusBarTipStyle

	leftRendered := leftStyle.Render(" " + leftText + " ")
	rightRendered := rightStyle.Render(" " + rightText + " ")

	leftWidth := lipgloss.Width(leftRendered)
	rightWidth := lipgloss.Width(rightRendered)

	var middleRendered string
	var middleWidth int
	if props.Tip != "" {
		middleRendered = tipStyle.Render(" Tip: " + props.Tip)
		middleWidth = lipgloss.Width(middleRendered)
	}

	gapWidth := max(props.Width-leftWidth-rightWidth-middleWidth, 1)

	gap := StatusBarTipStyle.Render(strings.Repeat(" ", gapWidth))

	if middleRendered != "" {
		return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, middleRendered, gap, rightRendered)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}
