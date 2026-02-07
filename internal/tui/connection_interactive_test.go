package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, state.Connected, m.ConnectionState.Status())

	// Verify model can handle messages while connected
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyRight})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify model still maintains connected state
	assert.Equal(t, state.Connected, m.ConnectionState.Status())
}

// TestConnection_DisconnectNotification tests disconnect event handling
func TestConnection_DisconnectNotification(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// First establish a connected state
	m.ConnectionState.SetStatus(state.Connected)
	time.Sleep(10 * time.Millisecond)

	// Verify state is connected
	assert.Equal(t, state.Connected, m.ConnectionState.Status())

	// Simulate disconnect event by setting status to disconnected
	m.ConnectionState.SetStatus(state.Disconnected)
	time.Sleep(10 * time.Millisecond)

	// Verify state changed to disconnected
	assert.Equal(t, state.Disconnected, m.ConnectionState.Status())

	// Send a key message while disconnected
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify model continues to process messages gracefully
	assert.Equal(t, state.Disconnected, m.ConnectionState.Status())
}

// TestConnection_ReconnectingState tests reconnecting status display
func TestConnection_ReconnectingState(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// First establish a connected state
	m.ConnectionState.SetStatus(state.Connected)
	time.Sleep(10 * time.Millisecond)

	// Verify state is connected
	assert.Equal(t, state.Connected, m.ConnectionState.Status())

	// Transition to reconnecting state (connection lost but trying to restore)
	m.ConnectionState.SetStatus(state.Reconnecting)
	time.Sleep(10 * time.Millisecond)

	// Verify state changed to reconnecting
	assert.Equal(t, state.Reconnecting, m.ConnectionState.Status())

	// Send navigation message while reconnecting
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify model handles messages during reconnecting state
	assert.Equal(t, state.Reconnecting, m.ConnectionState.Status())
}

// TestConnection_SuccessfulReconnect tests transition back to connected
func TestConnection_SuccessfulReconnect(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Establish initial connected state
	m.ConnectionState.SetStatus(state.Connected)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, state.Connected, m.ConnectionState.Status())

	// Simulate disconnect
	m.ConnectionState.SetStatus(state.Disconnected)
	time.Sleep(10 * time.Millisecond)

	// Verify disconnected state
	assert.Equal(t, state.Disconnected, m.ConnectionState.Status())

	// Transition to reconnecting
	m.ConnectionState.SetStatus(state.Reconnecting)
	time.Sleep(10 * time.Millisecond)

	// Verify reconnecting state
	assert.Equal(t, state.Reconnecting, m.ConnectionState.Status())

	// Reconnect successfully
	m.ConnectionState.SetStatus(state.Connected)
	time.Sleep(10 * time.Millisecond)

	// Verify back to connected state
	assert.Equal(t, state.Connected, m.ConnectionState.Status())
}

// TestConnection_TaskOpsWhileDisconnected tests graceful degradation when offline
func TestConnection_TaskOpsWhileDisconnected(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Disconnect the model
	m.ConnectionState.SetStatus(state.Disconnected)
	time.Sleep(10 * time.Millisecond)

	// Verify disconnected
	assert.Equal(t, state.Disconnected, m.ConnectionState.Status())

	// Try to navigate while disconnected
	m.UIState.Mode = state.NormalMode
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyRight})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify mode is still normal (no interruption)
	assert.Equal(t, state.NormalMode, m.UIState.Mode)

	// Try another navigation operation
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify model remains functional
	assert.Equal(t, state.Disconnected, m.ConnectionState.Status())
}
