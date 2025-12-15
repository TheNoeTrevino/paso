package daemon

import (
	"sync/atomic"
	"time"
)

// Metrics tracks daemon statistics using atomic operations for thread-safety
type Metrics struct {
	EventsSent       atomic.Int64
	EventsReceived   atomic.Int64
	Reconnections    atomic.Int64
	RefreshesTotal   atomic.Int64
	ConnectedClients atomic.Int32
	StartTime        time.Time
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// IncEventsSent increments the events sent counter
func (m *Metrics) IncEventsSent() {
	m.EventsSent.Add(1)
}

// IncEventsReceived increments the events received counter
func (m *Metrics) IncEventsReceived() {
	m.EventsReceived.Add(1)
}

// IncReconnections increments the reconnections counter
func (m *Metrics) IncReconnections() {
	m.Reconnections.Add(1)
}

// IncRefreshesTotal increments the refreshes total counter
func (m *Metrics) IncRefreshesTotal() {
	m.RefreshesTotal.Add(1)
}

// SetConnectedClients sets the current connected clients count
func (m *Metrics) SetConnectedClients(count int32) {
	m.ConnectedClients.Store(count)
}

// GetEventsSent returns the total events sent
func (m *Metrics) GetEventsSent() int64 {
	return m.EventsSent.Load()
}

// GetEventsReceived returns the total events received
func (m *Metrics) GetEventsReceived() int64 {
	return m.EventsReceived.Load()
}

// GetReconnections returns the total reconnections
func (m *Metrics) GetReconnections() int64 {
	return m.Reconnections.Load()
}

// GetRefreshesTotal returns the total refreshes
func (m *Metrics) GetRefreshesTotal() int64 {
	return m.RefreshesTotal.Load()
}

// GetConnectedClients returns the current connected clients count
func (m *Metrics) GetConnectedClients() int32 {
	return m.ConnectedClients.Load()
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	EventsSent       int64     `json:"events_sent"`
	EventsReceived   int64     `json:"events_received"`
	Reconnections    int64     `json:"reconnections"`
	RefreshesTotal   int64     `json:"refreshes_total"`
	ConnectedClients int32     `json:"connected_clients"`
	StartTime        time.Time `json:"start_time"`
	Uptime           string    `json:"uptime"`
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	return MetricsSnapshot{
		EventsSent:       m.GetEventsSent(),
		EventsReceived:   m.GetEventsReceived(),
		Reconnections:    m.GetReconnections(),
		RefreshesTotal:   m.GetRefreshesTotal(),
		ConnectedClients: m.GetConnectedClients(),
		StartTime:        m.StartTime,
		Uptime:           time.Since(m.StartTime).String(),
	}
}
