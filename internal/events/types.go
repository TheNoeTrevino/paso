package events

import "time"

// EventType indicates what kind of change occurred
type EventType string

const (
	EventDatabaseChanged EventType = "db_changed"
	EventPing            EventType = "ping"
	EventPong            EventType = "pong"
)

// Event represents a database change notification
type Event struct {
	Type       EventType
	ProjectID  int       // For filtering - which project was modified
	Timestamp  time.Time // When the event occurred
	SequenceID int64     // Monotonically increasing sequence number for ordering
}

// SubscribeMessage is sent by clients to subscribe to specific project updates
type SubscribeMessage struct {
	ProjectID int // 0 = all projects, >0 = specific project
}

// Message wraps events and control messages for wire protocol
type Message struct {
	Type      string             // "event", "subscribe", "ping", "pong"
	Event     *Event             `json:",omitempty"`
	Subscribe *SubscribeMessage  `json:",omitempty"`
}
