package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

func (m Model) handleFilterBarMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.UI.Filter.IsActive = false
		m.UIState.Mode = state.NormalMode
		return m, nil

	case "enter":
		if m.UI.Filter.FocusedChip == state.FilterChipClearAll {
			m.UI.Filter.ClearAll()
			return m.executeSearch()
		}
		// TODO: Open picker for focused chip (EPIC 4-7)
		return m, nil

	case "h", "left":
		m.UI.Filter.MoveFocusLeft()
		return m, nil

	case "l", "right":
		m.UI.Filter.MoveFocusRight()
		return m, nil

	case "x", "delete", "backspace":
		if m.UI.Filter.ClearFocusedChip() {
			return m.executeSearch()
		}
		return m, nil
	}

	return m, nil
}
