package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// HELP MODE HANDLERS
// ============================================================================

// handleHelpMode handles input in the help screen.
func (m Model) handleHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case m.Config.KeyMappings.ShowHelp, m.Config.KeyMappings.Quit, "esc", "enter", " ":
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}
