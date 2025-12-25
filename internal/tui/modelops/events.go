package modelops

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui"
)

// TODO: is this unused? if so, lets remove it
// SubscribeToEvents returns a command that listens for events from the daemon
// and sends RefreshMsg when data changes.
// Returns nil if EventChan is not initialized.
func SubscribeToEvents(m *tui.Model) tea.Cmd {
	if m.EventChan == nil {
		return nil
	}

	return func() tea.Msg {
		select {
		case event, ok := <-m.EventChan:
			if !ok {
				// Channel closed, connection lost
				return nil
			}
			return tui.RefreshMsg{Event: event}
		case <-m.Ctx.Done():
			return nil
		}
	}
}
