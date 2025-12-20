package state

import "sync"

// ConnectionStatus represents the current connection state to the daemon
type ConnectionStatus int

const (
	Disconnected ConnectionStatus = iota
	Connected
	Reconnecting
)

// String returns a human-readable string representation of the connection status
func (cs ConnectionStatus) String() string {
	switch cs {
	case Connected:
		return "Connected"
	case Disconnected:
		return "Disconnected"
	case Reconnecting:
		return "Reconnecting"
	default:
		return "Unknown"
	}
}

// ConnectionState manages the connection status to the daemon
type ConnectionState struct {
	mu     sync.RWMutex
	status ConnectionStatus
}

// NewConnectionState creates a new ConnectionState with the given initial status
func NewConnectionState(initialStatus ConnectionStatus) *ConnectionState {
	return &ConnectionState{
		status: initialStatus,
	}
}

// Status returns the current connection status (thread-safe)
func (cs *ConnectionState) Status() ConnectionStatus {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.status
}

// SetStatus updates the connection status (thread-safe)
func (cs *ConnectionState) SetStatus(status ConnectionStatus) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.status = status
}
