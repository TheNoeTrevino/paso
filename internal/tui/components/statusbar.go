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
	SearchMode       bool
	SearchQuery      string
	ConnectionStatus state.ConnectionStatus
}

// RenderStatusBar renders a status bar with left and right aligned text
// Left side: connection status
// Middle: "/search-query" when searching (takes space from gap)
// Right side: "press ? for help"
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

	rightText := "? for help"

	leftStyle := StatusBarStyle.Foreground(lipgloss.Color(leftColor))
	rightStyle := StatusBarStyle
	searchStyle := StatusBarSearchStyle

	leftRendered := leftStyle.Render(" " + leftText + " ")
	rightRendered := rightStyle.Render(" " + rightText + " ")

	// Calculate space between left and right text
	leftWidth := lipgloss.Width(leftRendered)
	rightWidth := lipgloss.Width(rightRendered)

	// If searching, render search query and subtract its width from gap
	var searchRendered string
	var searchWidth int
	if props.SearchMode {
		searchText := "/" + props.SearchQuery
		searchRendered = searchStyle.Render(searchText)
		searchWidth = lipgloss.Width(searchRendered)
	}

	gapWidth := max(props.Width-leftWidth-rightWidth-searchWidth, 1)

	gap := StatusBarSearchStyle.Render(strings.Repeat(" ", gapWidth))

	if props.SearchMode {
		return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, searchRendered, gap, rightRendered)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}
