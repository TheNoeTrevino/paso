package components

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// RenderTabs renders a tab bar with the given tab names
// selectedIdx indicates which tab is active (0-indexed)
// width is the total width to fill with the tab gap
//
// Layout:
//
//	╭──────╮ ╭──────╮                      [Notification]
//	│ Tab1 │ │ Tab2 │──────────────────────
//	      active    inactive
func RenderTabs(tabs []string, selectedIdx int, width int, notificationContent string) string {
	var renderedTabs []string

	for i, tabName := range tabs {
		if i == selectedIdx {
			renderedTabs = append(renderedTabs, ActiveTabStyle.Render(tabName))
		} else {
			renderedTabs = append(renderedTabs, TabStyle.Render(tabName))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Calculate gap width accounting for notification if present
	notificationWidth := lipgloss.Width(notificationContent)
	gapWidth := max(width-lipgloss.Width(row)-notificationWidth-2, 0)
	gap := TabGapStyle.Render(strings.Repeat(" ", gapWidth))

	// Join: tabs + gap + notification
	if notificationContent != "" {
		return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap, notificationContent)
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
}
