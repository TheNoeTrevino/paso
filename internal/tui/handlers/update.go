package handlers

import (
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// Update is the main update dispatcher that handles all messages and updates the model.
// This implements the "Update" part of the Model-View-Update pattern.
func Update(m *tui.Model, msg tea.Msg) tea.Cmd {
	// Check if context is cancelled (graceful shutdown)
	select {
	case <-m.Ctx.Done():
		// Context cancelled, initiate graceful shutdown
		return tea.Quit
	default:
		// Continue normal processing
	}

	// Start listening for events on first update if not already started
	var cmd tea.Cmd
	if m.EventChan != nil && !m.SubscriptionStarted {
		m.SubscriptionStarted = true
		cmd = modelops.SubscribeToEvents(m)
	}

	// Handle form modes first - forms need ALL messages
	// TODO: Extract form update logic to handlers/forms.go
	if m.UiState.Mode() == state.TicketFormMode {
		_, formCmd := m.Update(msg)
		return formCmd
	}
	if m.UiState.Mode() == state.ProjectFormMode {
		_, formCmd := m.Update(msg)
		return formCmd
	}

	switch msg := msg.(type) {
	case tui.RefreshMsg:
		// Only refresh if event is for current project
		currentProject := m.AppState.GetCurrentProject()
		if currentProject != nil && msg.Event.ProjectID == currentProject.ID {
			modelops.ReloadCurrentProject(m)
		}

		// Continue listening for more events
		cmd = modelops.SubscribeToEvents(m)
		return cmd

	case events.NotificationMsg:
		// Handle user-facing notification from events client
		level := state.LevelInfo
		switch msg.Level {
		case "error":
			level = state.LevelError
		case "warning":
			level = state.LevelWarning
		}
		m.NotificationState.Add(level, msg.Message)

		// Update connection status based on notification message
		if strings.Contains(msg.Message, "Connection lost") || strings.Contains(msg.Message, "reconnecting") {
			m.ConnectionState.SetStatus(state.Reconnecting)
		} else if strings.Contains(msg.Message, "Reconnected") {
			m.ConnectionState.SetStatus(state.Connected)
		} else if strings.Contains(msg.Message, "Failed to reconnect") {
			m.ConnectionState.SetStatus(state.Disconnected)
		}

		// Continue listening for more notifications
		cmd = listenForNotifications(m)
		return cmd

	case tui.ConnectionEstablishedMsg:
		m.ConnectionState.SetStatus(state.Connected)
		return nil

	case tui.ConnectionLostMsg:
		m.ConnectionState.SetStatus(state.Disconnected)
		return nil

	case tui.ConnectionReconnectingMsg:
		m.ConnectionState.SetStatus(state.Reconnecting)
		return nil

	case tea.KeyMsg:
		return HandleKeyMsg(m, msg)

	case tea.WindowSizeMsg:
		return HandleWindowResize(m, msg)
	}

	return cmd
}

// HandleKeyMsg dispatches key messages to the appropriate mode handler.
func HandleKeyMsg(m *tui.Model, msg tea.KeyMsg) tea.Cmd {
	switch m.UiState.Mode() {
	case state.NormalMode:
		return HandleNormalMode(m, msg)
	case state.AddColumnMode, state.EditColumnMode:
		return HandleInputMode(m, msg)
	case state.DiscardConfirmMode:
		return HandleDiscardConfirm(m, msg)
	case state.DeleteConfirmMode:
		return HandleDeleteConfirm(m, msg)
	case state.DeleteColumnConfirmMode:
		return HandleDeleteColumnConfirm(m, msg)
	case state.HelpMode:
		return HandleHelpMode(m, msg)
	case state.LabelPickerMode:
		// TODO: Extract picker logic to handlers/pickers.go
		_, pickerCmd := m.Update(msg)
		return pickerCmd
	case state.ParentPickerMode:
		_, pickerCmd := m.Update(msg)
		return pickerCmd
	case state.ChildPickerMode:
		_, pickerCmd := m.Update(msg)
		return pickerCmd
	case state.PriorityPickerMode:
		_, pickerCmd := m.Update(msg)
		return pickerCmd
	case state.TypePickerMode:
		_, pickerCmd := m.Update(msg)
		return pickerCmd
	case state.RelationTypePickerMode:
		_, pickerCmd := m.Update(msg)
		return pickerCmd
	case state.SearchMode:
		return HandleSearchMode(m, msg)
	case state.StatusPickerMode:
		return HandleStatusPickerMode(m, msg)
	}
	return nil
}

// HandleWindowResize handles terminal resize events.
func HandleWindowResize(m *tui.Model, msg tea.WindowSizeMsg) tea.Cmd {
	m.UiState.SetWidth(msg.Width)
	m.UiState.SetHeight(msg.Height)

	// Update notification state with new window dimensions
	m.NotificationState.SetWindowSize(msg.Width, msg.Height)

	// Ensure viewport offset is still valid after resize
	if m.UiState.ViewportOffset()+m.UiState.ViewportSize() > len(m.AppState.Columns()) {
		m.UiState.SetViewportOffset(max(0, len(m.AppState.Columns())-m.UiState.ViewportSize()))
	}
	return nil
}

// listenForNotifications returns a command that waits for the next notification from the events client.
func listenForNotifications(m *tui.Model) tea.Cmd {
	if m.NotifyChan == nil {
		return nil
	}

	return func() tea.Msg {
		// Wait for notification from channel
		notification, ok := <-m.NotifyChan
		if !ok {
			// Channel closed
			slog.Info("notification channel closed")
			return nil
		}
		return notification
	}
}
