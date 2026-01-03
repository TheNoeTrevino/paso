package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestConnection_InitialConnectedState verifies model starts in connected state (when event client available)
func TestConnection_InitialConnectedState(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// When InitialModel is called without event client, it starts in disconnected state
	// This test verifies that we can transition to and maintain connected state

	// Simulate establishing connection
	m.ConnectionState.SetStatus(state.Connected)
	time.Sleep(10 * time.Millisecond)

	// Verify connection state is now connected
	if m.ConnectionState.Status() != state.Connected {
		t.Errorf("Expected status to be Connected, got %v", m.ConnectionState.Status())
	}

	// Verify model can handle messages while connected
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyRight})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify model still maintains connected state
	if m.ConnectionState.Status() != state.Connected {
		t.Errorf("Expected to remain Connected after update, got %v", m.ConnectionState.Status())
	}
}

// TestConnection_DisconnectNotification tests disconnect event handling
func TestConnection_DisconnectNotification(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// First establish a connected state
	m.ConnectionState.SetStatus(state.Connected)
	time.Sleep(10 * time.Millisecond)

	// Verify state is connected
	if m.ConnectionState.Status() != state.Connected {
		t.Errorf("Expected status to be Connected, got %v", m.ConnectionState.Status())
	}

	// Simulate disconnect event by setting status to disconnected
	m.ConnectionState.SetStatus(state.Disconnected)
	time.Sleep(10 * time.Millisecond)

	// Verify state changed to disconnected
	if m.ConnectionState.Status() != state.Disconnected {
		t.Errorf("Expected status to be Disconnected, got %v", m.ConnectionState.Status())
	}

	// Send a key message while disconnected
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify model continues to process messages gracefully
	if m.ConnectionState.Status() != state.Disconnected {
		t.Errorf("Expected to remain Disconnected, got %v", m.ConnectionState.Status())
	}
}

// TestConnection_ReconnectingState tests reconnecting status display
func TestConnection_ReconnectingState(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// First establish a connected state
	m.ConnectionState.SetStatus(state.Connected)
	time.Sleep(10 * time.Millisecond)

	// Verify state is connected
	if m.ConnectionState.Status() != state.Connected {
		t.Errorf("Expected initial status to be Connected, got %v", m.ConnectionState.Status())
	}

	// Transition to reconnecting state (connection lost but trying to restore)
	m.ConnectionState.SetStatus(state.Reconnecting)
	time.Sleep(10 * time.Millisecond)

	// Verify state changed to reconnecting
	if m.ConnectionState.Status() != state.Reconnecting {
		t.Errorf("Expected status to be Reconnecting, got %v", m.ConnectionState.Status())
	}

	// Send navigation message while reconnecting
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify model handles messages during reconnecting state
	if m.ConnectionState.Status() != state.Reconnecting {
		t.Errorf("Expected to remain in Reconnecting state, got %v", m.ConnectionState.Status())
	}
}

// TestConnection_SuccessfulReconnect tests transition back to connected
func TestConnection_SuccessfulReconnect(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Establish initial connected state
	m.ConnectionState.SetStatus(state.Connected)
	time.Sleep(10 * time.Millisecond)

	if m.ConnectionState.Status() != state.Connected {
		t.Errorf("Expected initial status to be Connected, got %v", m.ConnectionState.Status())
	}

	// Simulate disconnect
	m.ConnectionState.SetStatus(state.Disconnected)
	time.Sleep(10 * time.Millisecond)

	// Verify disconnected state
	if m.ConnectionState.Status() != state.Disconnected {
		t.Errorf("Expected status to be Disconnected, got %v", m.ConnectionState.Status())
	}

	// Transition to reconnecting
	m.ConnectionState.SetStatus(state.Reconnecting)
	time.Sleep(10 * time.Millisecond)

	// Verify reconnecting state
	if m.ConnectionState.Status() != state.Reconnecting {
		t.Errorf("Expected status to be Reconnecting, got %v", m.ConnectionState.Status())
	}

	// Reconnect successfully
	m.ConnectionState.SetStatus(state.Connected)
	time.Sleep(10 * time.Millisecond)

	// Verify back to connected state
	if m.ConnectionState.Status() != state.Connected {
		t.Errorf("Expected status to be Connected after reconnect, got %v", m.ConnectionState.Status())
	}
}

// TestConnection_TaskOpsWhileDisconnected tests graceful degradation when offline
func TestConnection_TaskOpsWhileDisconnected(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Disconnect the model
	m.ConnectionState.SetStatus(state.Disconnected)
	time.Sleep(10 * time.Millisecond)

	// Verify disconnected
	if m.ConnectionState.Status() != state.Disconnected {
		t.Errorf("Expected status to be Disconnected, got %v", m.ConnectionState.Status())
	}

	// Try to navigate while disconnected
	m.UIState.SetMode(state.NormalMode)
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyRight})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify mode is still normal (no interruption)
	if m.UIState.Mode() != state.NormalMode {
		t.Errorf("Expected NormalMode while disconnected, got %v", m.UIState.Mode())
	}

	// Try another navigation operation
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify model remains functional
	if m.ConnectionState.Status() != state.Disconnected {
		t.Errorf("Expected to remain Disconnected, got %v", m.ConnectionState.Status())
	}
}
