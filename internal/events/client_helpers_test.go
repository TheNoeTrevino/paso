package events

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// setupMockDaemon creates a simple mock daemon server for testing
func setupMockDaemon(t *testing.T) (string, net.Listener, chan Message) {
	t.Helper()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create Unix socket listener
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	})

	// Channel to send messages received from client
	messages := make(chan Message, 10)

	// Accept connections in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Listener closed
			}

			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				decoder := json.NewDecoder(c)
				encoder := json.NewEncoder(c)

				for {
					var msg Message
					if err := decoder.Decode(&msg); err != nil {
						return
					}

					// Echo message to test channel
					select {
					case messages <- msg:
					default:
					}

					// Send ack for subscribe messages
					if msg.Type == "subscribe" {
						ackMsg := Message{
							Version: ProtocolVersion,
							Type:    "ack",
						}
						_ = encoder.Encode(ackMsg)
					}
				}
			}(conn)
		}
	}()

	return socketPath, listener, messages
}

// setupMockDaemonWithControl creates a mock daemon server that can be stopped
// and restarted on demand for reconnection testing.
//
// It properly tracks and closes all active client connections when stopped, ensuring
// clients detect disconnection immediately rather than blocking indefinitely.
//
// Returns:
//   - socketPath: path to the Unix socket
//   - startFunc: function to start/restart the daemon (returns error if start fails)
//   - stopFunc: function to stop the daemon
//   - messages: channel to receive messages from client
//
// The daemon can be stopped and started multiple times. Each restart creates
// a new listener on the same socket path. Proper cleanup is handled via t.Cleanup().
func setupMockDaemonWithControl(t *testing.T) (string, func() error, func(), chan Message) {
	t.Helper()

	// Create temp directory and socket path
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Channel to send messages received from client
	messages := make(chan Message, 10)

	// Track the current listener and its cancel function
	var currentListener net.Listener
	var cancelAccept context.CancelFunc

	// Track active connections for proper cleanup when daemon stops
	var activeConns map[net.Conn]struct{}
	var connsMu sync.Mutex
	var listenerMu struct {
		mu       *testing.T // Prevents direct mutex usage, uses test helper pattern
		listener net.Listener
		cancel   context.CancelFunc
	}
	listenerMu.mu = t

	// Track cleanup state
	cleanedUp := false
	var cleanupMu struct {
		mu        *testing.T
		cleanedUp *bool
	}
	cleanupMu.mu = t
	cleanupMu.cleanedUp = &cleanedUp

	// startFunc creates a new listener and starts accepting connections
	startFunc := func() error {
		// Remove old socket file if it exists
		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old socket: %w", err)
		}

		// Create Unix socket listener
		listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to create mock daemon listener: %w", err)
		}

		// Create cancellable context for accept loop
		ctx, cancel := context.WithCancel(context.Background())

		// Update current listener and cancel function
		currentListener = listener
		cancelAccept = cancel

		// Initialize connection tracking map for this daemon instance
		connsMu.Lock()
		activeConns = make(map[net.Conn]struct{})
		connsMu.Unlock()

		// Accept connections in background
		go func() {
			for {
				// Check if context is cancelled
				select {
				case <-ctx.Done():
					return
				default:
				}

				conn, err := listener.Accept()
				if err != nil {
					// Check if this is due to listener being closed
					select {
					case <-ctx.Done():
						return // Normal shutdown
					default:
						// Unexpected error, but continue to allow for reconnection
						return
					}
				}

				// Track this connection for cleanup when daemon stops
				connsMu.Lock()
				activeConns[conn] = struct{}{}
				connsMu.Unlock()
				go func(c net.Conn) {
					// Remove from active connections when handler exits (any reason: error, normal, panic)
					defer func() {
						connsMu.Lock()
						delete(activeConns, c)
						connsMu.Unlock()
						_ = c.Close()
					}()

					decoder := json.NewDecoder(c)
					encoder := json.NewEncoder(c)

					for {
						var msg Message
						if err := decoder.Decode(&msg); err != nil {
							return
						}

						// Echo message to test channel
						select {
						case messages <- msg:
						default:
						}

						// Send ack for subscribe messages
						if msg.Type == "subscribe" {
							ackMsg := Message{
								Version: ProtocolVersion,
								Type:    "ack",
							}
							_ = encoder.Encode(ackMsg)
						}
					}
				}(conn)
			}
		}()

		return nil
	}

	// stopFunc stops the daemon by closing the listener AND all active connections
	stopFunc := func() {
		// Step 1: Stop accepting new connections
		if cancelAccept != nil {
			cancelAccept()
		}
		if currentListener != nil {
			_ = currentListener.Close()
		}

		// Step 2: Close all active connections (THE CRITICAL FIX)
		connsMu.Lock()
		// Copy connections to slice to avoid holding lock during Close()
		connsToClose := make([]net.Conn, 0, len(activeConns))
		for conn := range activeConns {
			connsToClose = append(connsToClose, conn)
		}
		// Clear map while we have the lock
		activeConns = make(map[net.Conn]struct{})
		connsMu.Unlock()

		// Close without holding mutex (prevents deadlock)
		for _, conn := range connsToClose {
			_ = conn.Close()
		}

		// Shorter sleep since we actively close connections
		time.Sleep(10 * time.Millisecond)
	}

	// Register cleanup to ensure resources are released
	t.Cleanup(func() {
		*cleanupMu.cleanedUp = true
		stopFunc()
		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to remove socket during cleanup: %v", err)
		}
	})

	return socketPath, startFunc, stopFunc, messages
}
