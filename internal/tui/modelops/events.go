package modelops

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui"
)

// SubscribeToEvents returns a command that listens for events from the daemon
// and sends RefreshMsg when data changes.
// Returns nil if EventChan is not initialized.
func (w *Wrapper) SubscribeToEvents() tea.Cmd {
	if w.EventChan == nil {
		return nil
	}

	return func() tea.Msg {
		select {
		case event, ok := <-w.EventChan:
			if !ok {
				// Channel closed, connection lost
				return nil
			}
			return tui.RefreshMsg{Event: event}
		case <-w.Ctx.Done():
			return nil
		}
	}
}
