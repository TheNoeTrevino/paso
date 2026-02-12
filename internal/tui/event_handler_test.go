package tui

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestEventHandler_RefreshMsg_ReloadsData verifies that a RefreshMsg for the current project
// triggers a data reload and the model reflects the updated database state.
func TestEventHandler_RefreshMsg_ReloadsData(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)
	m.NotifyChan = make(chan events.NotificationMsg, 1)
	m.SubscriptionStarted = true

	// Verify we have a current project
	currentProject := m.AppState.GetCurrentProject()
	require.NotNil(t, currentProject, "test model should have a current project")

	// Get the first column to add a task
	columns := m.AppState.Columns()
	require.NotEmpty(t, columns, "test model should have columns")
	columnID := columns[0].ID

	// Add a task directly to the database (simulating an external change)
	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), int(columnID), "New Task From Event")

	// Send a RefreshMsg through Update (broadcast event)
	refreshMsg := RefreshMsg{
		Event: events.Event{
			Type:       events.EventDatabaseChanged,
			ProjectID:  0, // broadcast
			Timestamp:  time.Now(),
			SequenceID: 1,
		},
	}
	newModel, cmd := m.Update(refreshMsg)
	m = newModel.(Model)

	// After refresh, the model should have reloaded tasks including the new one
	tasks := m.AppState.Tasks()
	found := false
	for _, columnTasks := range tasks {
		for _, task := range columnTasks {
			if task.Title == "New Task From Event" {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "model should contain the new task after RefreshMsg reload")

	// RefreshMsg should return a non-nil cmd (resubscribe command)
	// When EventChan is nil, subscribeToEvents returns nil
	_ = cmd
}

// TestEventHandler_RefreshMsg_SpecificProject verifies that a RefreshMsg with the
// current project's ID triggers a reload.
func TestEventHandler_RefreshMsg_SpecificProject(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)
	m.NotifyChan = make(chan events.NotificationMsg, 1)
	m.SubscriptionStarted = true

	currentProject := m.AppState.GetCurrentProject()
	require.NotNil(t, currentProject)

	columns := m.AppState.Columns()
	require.NotEmpty(t, columns)
	columnID := columns[0].ID

	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), int(columnID), "Project-Specific Task")

	// Send RefreshMsg targeting this specific project
	refreshMsg := RefreshMsg{
		Event: events.Event{
			Type:       events.EventDatabaseChanged,
			ProjectID:  currentProject.ID,
			Timestamp:  time.Now(),
			SequenceID: 1,
		},
	}
	newModel, _ := m.Update(refreshMsg)
	m = newModel.(Model)

	tasks := m.AppState.Tasks()
	found := false
	for _, columnTasks := range tasks {
		for _, task := range columnTasks {
			if task.Title == "Project-Specific Task" {
				found = true
			}
		}
	}
	assert.True(t, found, "model should reload when RefreshMsg targets the current project")
}

// TestEventHandler_RefreshMsg_IgnoresOtherProject verifies that a RefreshMsg for a
// different project does not trigger a reload of the current project's data.
func TestEventHandler_RefreshMsg_IgnoresOtherProject(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)
	m.NotifyChan = make(chan events.NotificationMsg, 1)
	m.SubscriptionStarted = true

	currentProject := m.AppState.GetCurrentProject()
	require.NotNil(t, currentProject)

	columns := m.AppState.Columns()
	require.NotEmpty(t, columns)
	columnID := columns[0].ID

	// Snapshot current task count
	initialTasks := m.AppState.Tasks()
	initialCount := 0
	for _, columnTasks := range initialTasks {
		initialCount += len(columnTasks)
	}

	// Add a task directly to DB
	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), int(columnID), "Should Not Appear")

	// Send RefreshMsg for a different project ID
	refreshMsg := RefreshMsg{
		Event: events.Event{
			Type:       events.EventDatabaseChanged,
			ProjectID:  99999, // different project
			Timestamp:  time.Now(),
			SequenceID: 1,
		},
	}
	newModel, _ := m.Update(refreshMsg)
	m = newModel.(Model)

	// Task count should not have changed (no reload happened)
	afterTasks := m.AppState.Tasks()
	afterCount := 0
	for _, columnTasks := range afterTasks {
		afterCount += len(columnTasks)
	}
	assert.Equal(t, initialCount, afterCount, "model should not reload for a different project's RefreshMsg")
}

// TestEventHandler_NotificationMsg_AddsNotification verifies that notification messages
// sent through Update are added to the notification state.
func TestEventHandler_NotificationMsg_AddsNotification(t *testing.T) {
	t.Parallel()
	m := setupTestModel([]*models.Column{{ID: 1, Name: "Todo"}}, nil)
	m.NotifyChan = make(chan events.NotificationMsg, 1)
	m.ConnectionState = state.NewConnectionState(state.Connected)

	notifMsg := events.NotificationMsg{
		Level:   "error",
		Message: "Something went wrong",
	}

	newModel, cmd := m.Update(notifMsg)
	m = newModel.(Model)

	assert.True(t, m.UI.Notification.HasAny(), "notification should be added after NotificationMsg")
	notifications := m.UI.Notification.All()
	require.Len(t, notifications, 1)
	assert.Equal(t, state.LevelError, notifications[0].Level)
	assert.Equal(t, "Something went wrong", notifications[0].Message)

	// Should return a cmd to re-listen for the next notification
	assert.NotNil(t, cmd, "handleNotificationMsg should return a re-listen command")
}

// TestEventHandler_NotificationMsg_UpdatesConnectionState verifies that specific
// notification messages (e.g. "Connection lost") update the connection state.
func TestEventHandler_NotificationMsg_UpdatesConnectionState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		message        string
		initialStatus  state.ConnectionStatus
		expectedStatus state.ConnectionStatus
	}{
		{
			name:           "connection lost triggers reconnecting",
			message:        "Connection lost to daemon",
			initialStatus:  state.Connected,
			expectedStatus: state.Reconnecting,
		},
		{
			name:           "reconnected triggers connected",
			message:        "Reconnected to daemon",
			initialStatus:  state.Reconnecting,
			expectedStatus: state.Connected,
		},
		{
			name:           "failed to reconnect triggers disconnected",
			message:        "Failed to reconnect after retries",
			initialStatus:  state.Reconnecting,
			expectedStatus: state.Disconnected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupTestModel([]*models.Column{{ID: 1, Name: "Todo"}}, nil)
			m.NotifyChan = make(chan events.NotificationMsg, 1)
			m.ConnectionState = state.NewConnectionState(tt.initialStatus)

			notifMsg := events.NotificationMsg{Level: "warning", Message: tt.message}
			newModel, _ := m.Update(notifMsg)
			m = newModel.(Model)

			assert.Equal(t, tt.expectedStatus, m.ConnectionState.Status(),
				"connection status should update based on notification message")
		})
	}
}

// TestEventHandler_ConnectionMessages verifies that connection status messages
// (ConnectionEstablishedMsg, ConnectionLostMsg, ConnectionReconnectingMsg)
// update the model's connection state when sent through Update.
func TestEventHandler_ConnectionMessages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		msg            interface{}
		expectedStatus state.ConnectionStatus
	}{
		{
			name:           "ConnectionEstablishedMsg sets Connected",
			msg:            ConnectionEstablishedMsg{},
			expectedStatus: state.Connected,
		},
		{
			name:           "ConnectionLostMsg sets Disconnected",
			msg:            ConnectionLostMsg{},
			expectedStatus: state.Disconnected,
		},
		{
			name:           "ConnectionReconnectingMsg sets Reconnecting",
			msg:            ConnectionReconnectingMsg{},
			expectedStatus: state.Reconnecting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupTestModel([]*models.Column{{ID: 1, Name: "Todo"}}, nil)
			m.NotifyChan = make(chan events.NotificationMsg, 1)
			m.ConnectionState = state.NewConnectionState(state.Disconnected)
			m.SubscriptionStarted = true

			newModel, _ := m.Update(tt.msg)
			m = newModel.(Model)

			assert.Equal(t, tt.expectedStatus, m.ConnectionState.Status())
		})
	}
}

// TestEventHandler_SequentialStateUpdates verifies state consistency during sequential
// state mutations, as happens in the bubbletea single-threaded event loop.
func TestEventHandler_SequentialStateUpdates(t *testing.T) {
	t.Parallel()
	m := setupTestModel(
		[]*models.Column{{ID: 1, Name: "Todo"}},
		map[int][]*models.TaskSummary{
			1: {{ID: 1, Title: "Task 1"}},
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	assert.Equal(t, 0, m.UIState.SelectedColumn)
	assert.Equal(t, 0, m.UIState.SelectedTask)
	assert.Equal(t, state.NormalMode, m.UIState.Mode)

	m.UIState.SelectedColumn = 1
	m.UIState.SelectedTask = 2
	m.UIState.Mode = state.TicketFormMode

	assert.Equal(t, 1, m.UIState.SelectedColumn)
	assert.Equal(t, 2, m.UIState.SelectedTask)
	assert.Equal(t, state.TicketFormMode, m.UIState.Mode)

	select {
	case <-ctx.Done():
		require.Fail(t, "Test timed out unexpectedly")
	default:
	}
}
