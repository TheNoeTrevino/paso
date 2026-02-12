package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtocolVersion(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 1, ProtocolVersion)
}

func TestEventTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventDatabaseChanged, "db_changed"},
		{EventPing, "ping"},
		{EventPong, "pong"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, string(tt.eventType))
	}
}

func TestEvent_Creation(t *testing.T) {
	t.Parallel()
	now := time.Now()
	event := Event{
		Type:       EventDatabaseChanged,
		ProjectID:  42,
		Timestamp:  now,
		SequenceID: 123,
	}

	assert.Equal(t, EventDatabaseChanged, event.Type)
	assert.Equal(t, 42, event.ProjectID)
	assert.Equal(t, now, event.Timestamp)
	assert.Equal(t, int64(123), event.SequenceID)
}

func TestSubscribeMessage_Creation(t *testing.T) {
	t.Parallel()
	// Test specific project subscription
	msg := SubscribeMessage{ProjectID: 5}
	assert.Equal(t, 5, msg.ProjectID)

	// Test all projects subscription
	allMsg := SubscribeMessage{ProjectID: 0}
	assert.Equal(t, 0, allMsg.ProjectID)
}

func TestMessage_EventMessage(t *testing.T) {
	t.Parallel()
	event := &Event{
		Type:      EventDatabaseChanged,
		ProjectID: 10,
	}

	msg := Message{
		Version: ProtocolVersion,
		Type:    "event",
		Event:   event,
	}

	assert.Equal(t, ProtocolVersion, msg.Version)
	assert.Equal(t, "event", msg.Type)
	require.NotNil(t, msg.Event)
	assert.Equal(t, 10, msg.Event.ProjectID)
	assert.Nil(t, msg.Subscribe)
}

func TestMessage_SubscribeMessage(t *testing.T) {
	t.Parallel()
	subscribe := &SubscribeMessage{ProjectID: 7}

	msg := Message{
		Version:   ProtocolVersion,
		Type:      "subscribe",
		Subscribe: subscribe,
	}

	assert.Equal(t, ProtocolVersion, msg.Version)
	assert.Equal(t, "subscribe", msg.Type)
	require.NotNil(t, msg.Subscribe)
	assert.Equal(t, 7, msg.Subscribe.ProjectID)
	assert.Nil(t, msg.Event)
}

func TestMessage_PingPong(t *testing.T) {
	t.Parallel()
	// Test ping message
	pingMsg := Message{
		Version: ProtocolVersion,
		Type:    "ping",
	}
	assert.Equal(t, "ping", pingMsg.Type)

	// Test pong message
	pongMsg := Message{
		Version: ProtocolVersion,
		Type:    "pong",
	}
	assert.Equal(t, "pong", pongMsg.Type)
}

func TestNotificationMsg_Levels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level   string
		message string
	}{
		{"info", "Information message"},
		{"warning", "Warning message"},
		{"error", "Error message"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			notif := NotificationMsg{
				Level:   tt.level,
				Message: tt.message,
			}

			assert.Equal(t, tt.level, notif.Level)
			assert.Equal(t, tt.message, notif.Message)
		})
	}
}
