package handlers

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// HELP MODE HANDLERS
// ============================================================================

// HandleHelpMode handles input in the help screen.
func (w *Wrapper) HandleHelpMode(msg tea.KeyMsg) (*Wrapper, tea.Cmd) {
	switch msg.String() {
	case w.Config.KeyMappings.ShowHelp, w.Config.KeyMappings.Quit, "esc", "enter", " ":
		w.UiState.SetMode(state.NormalMode)
		return w, nil
	}
	return w, nil
}
