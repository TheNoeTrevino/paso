package handlers

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// HELP MODE HANDLERS
// ============================================================================

// HandleHelpMode handles input in the help screen.
func HandleHelpMode(m *tui.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case m.Config.KeyMappings.ShowHelp, m.Config.KeyMappings.Quit, "esc", "enter", " ":
		m.UiState.SetMode(state.NormalMode)
		return nil
	}
	return nil
}
