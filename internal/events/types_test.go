package events

import (
	"testing"
	"time"
)

// ============================================================================
// Constants Tests
// ============================================================================

func TestProtocolVersion(t *testing.T) {
	if ProtocolVersion != 1 {
		t.Errorf("Expected ProtocolVersion to be 1, got %d", ProtocolVersion)
	}
}

func TestEventTypes(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventDatabaseChanged, "db_changed"},
		{EventPing, "ping"},
		{EventPong, "pong"},
	}

	for _, tt := range tests {
		if string(tt.eventType) != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, string(tt.eventType))
		}
	}
}

// ============================================================================
// Struct Tests
// ============================================================================

func TestEvent_Creation(t *testing.T) {
	now := time.Now()
	event := Event{
		Type:       EventDatabaseChanged,
		ProjectID:  42,
		Timestamp:  now,
		SequenceID: 123,
	}

	if event.Type != EventDatabaseChanged {
		t.Errorf("Expected type %s, got %s", EventDatabaseChanged, event.Type)
	}
	if event.ProjectID != 42 {
		t.Errorf("Expected ProjectID 42, got %d", event.ProjectID)
	}
	if !event.Timestamp.Equal(now) {
		t.Errorf("Expected timestamp %v, got %v", now, event.Timestamp)
	}
	if event.SequenceID != 123 {
		t.Errorf("Expected SequenceID 123, got %d", event.SequenceID)
	}
}

func TestSubscribeMessage_Creation(t *testing.T) {
	// Test specific project subscription
	msg := SubscribeMessage{ProjectID: 5}
	if msg.ProjectID != 5 {
		t.Errorf("Expected ProjectID 5, got %d", msg.ProjectID)
	}

	// Test all projects subscription
	allMsg := SubscribeMessage{ProjectID: 0}
	if allMsg.ProjectID != 0 {
		t.Errorf("Expected ProjectID 0 (all projects), got %d", allMsg.ProjectID)
	}
}

func TestMessage_EventMessage(t *testing.T) {
	event := &Event{
		Type:      EventDatabaseChanged,
		ProjectID: 10,
	}

	msg := Message{
		Version: ProtocolVersion,
		Type:    "event",
		Event:   event,
	}

	if msg.Version != ProtocolVersion {
		t.Errorf("Expected version %d, got %d", ProtocolVersion, msg.Version)
	}
	if msg.Type != "event" {
		t.Errorf("Expected type 'event', got '%s'", msg.Type)
	}
	if msg.Event == nil {
		t.Fatal("Expected Event to be set, got nil")
	}
	if msg.Event.ProjectID != 10 {
		t.Errorf("Expected Event ProjectID 10, got %d", msg.Event.ProjectID)
	}
	if msg.Subscribe != nil {
		t.Error("Expected Subscribe to be nil")
	}
}

func TestMessage_SubscribeMessage(t *testing.T) {
	subscribe := &SubscribeMessage{ProjectID: 7}

	msg := Message{
		Version:   ProtocolVersion,
		Type:      "subscribe",
		Subscribe: subscribe,
	}

	if msg.Version != ProtocolVersion {
		t.Errorf("Expected version %d, got %d", ProtocolVersion, msg.Version)
	}
	if msg.Type != "subscribe" {
		t.Errorf("Expected type 'subscribe', got '%s'", msg.Type)
	}
	if msg.Subscribe == nil {
		t.Fatal("Expected Subscribe to be set, got nil")
	}
	if msg.Subscribe.ProjectID != 7 {
		t.Errorf("Expected Subscribe ProjectID 7, got %d", msg.Subscribe.ProjectID)
	}
	if msg.Event != nil {
		t.Error("Expected Event to be nil")
	}
}

func TestMessage_PingPong(t *testing.T) {
	// Test ping message
	pingMsg := Message{
		Version: ProtocolVersion,
		Type:    "ping",
	}
	if pingMsg.Type != "ping" {
		t.Errorf("Expected type 'ping', got '%s'", pingMsg.Type)
	}

	// Test pong message
	pongMsg := Message{
		Version: ProtocolVersion,
		Type:    "pong",
	}
	if pongMsg.Type != "pong" {
		t.Errorf("Expected type 'pong', got '%s'", pongMsg.Type)
	}
}

func TestNotificationMsg_Levels(t *testing.T) {
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

			if notif.Level != tt.level {
				t.Errorf("Expected level '%s', got '%s'", tt.level, notif.Level)
			}
			if notif.Message != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, notif.Message)
			}
		})
	}
}
