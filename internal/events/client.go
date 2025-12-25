package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// NotifyFunc is a callback function for sending user-facing notifications to the TUI
type NotifyFunc func(level, message string)

// Client represents a connection to the Paso daemon for receiving live updates.
// It handles event sending, receiving, batching, reconnection, and subscriptions.
type Client struct {
	socketPath string
	conn       net.Conn
	encoder    *json.Encoder
	decoder    *json.Decoder
	mu         sync.Mutex

	// Batching configuration
	eventQueue chan Event
	debounce   time.Duration
	closeOnce  sync.Once // Ensure cleanup happens only once

	// Reconnection configuration
	maxRetries int
	baseDelay  time.Duration

	// Subscription state
	currentProjectID int

	// Event tracking
	lastSequence int64

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Batching goroutine
	batcherDone chan struct{}

	// Notification callback for user-facing messages
	onNotify NotifyFunc
}

// NewClient creates a new event client but does not connect.
// The socket path should be the full path to the Unix domain socket.
// The debounce duration controls event batching (default 100ms if not configured via env var).
func NewClient(socketPath string) (*Client, error) {
	// Read debounce duration from environment variable
	debounceMs := 100 // Default: 100ms
	if envVal := os.Getenv("PASO_EVENT_DEBOUNCE_MS"); envVal != "" {
		if parsed, err := strconv.Atoi(envVal); err == nil && parsed > 0 {
			debounceMs = parsed
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		socketPath:       socketPath,
		eventQueue:       make(chan Event, 100),
		debounce:         time.Duration(debounceMs) * time.Millisecond,
		maxRetries:       5,
		baseDelay:        1 * time.Second,
		currentProjectID: 0,
		lastSequence:     0,
		ctx:              ctx,
		cancel:           cancel,
		batcherDone:      make(chan struct{}),
	}, nil
}

// SetNotifyFunc sets the notification callback for user-facing messages
func (c *Client) SetNotifyFunc(fn NotifyFunc) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onNotify = fn
}

// notify calls the notification callback if it's set (thread-safe)
func (c *Client) notify(level, message string) {
	c.mu.Lock()
	fn := c.onNotify
	c.mu.Unlock()

	if fn != nil {
		fn(level, message)
	}
}

// Connect establishes a connection to the daemon socket.
// It sends an initial subscription message for all projects.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Dial the Unix domain socket
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to dial daemon socket: %w", err)
	}

	c.conn = conn
	c.encoder = json.NewEncoder(conn)
	c.decoder = json.NewDecoder(conn)

	// Send initial subscription for all projects (ProjectID = 0)
	msg := Message{
		Version: ProtocolVersion,
		Type:    "subscribe",
		Subscribe: &SubscribeMessage{
			ProjectID: 0,
		},
	}
	if err := c.encoder.Encode(msg); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Debug("error closing connection", "error", closeErr)
		}
		return fmt.Errorf("failed to send subscription: %w", err)
	}

	// Start the batching goroutine
	go c.startBatcher()

	return nil
}

// SendEvent queues an event to be sent to the daemon.
// Events are batched and sent in bursts within the debounce window.
// Returns error if the queue is full (non-blocking send).
func (c *Client) SendEvent(event Event) error {
	if c == nil {
		return fmt.Errorf("event client is nil")
	}
	select {
	case c.eventQueue <- event:
		return nil
	default:
		return fmt.Errorf("event queue full")
	}
}

// startBatcher runs in a goroutine and batches events from the queue.
// It sends a single event every debounce duration if any events are pending.
// If events from multiple projects are batched together, sends projectID 0 (all projects).
func (c *Client) startBatcher() {
	defer close(c.batcherDone)

	ticker := time.NewTicker(c.debounce)
	defer ticker.Stop()

	var pending bool
	var projectID int
	var hasMultipleProjects bool

	// Helper to flush pending events
	flushPending := func() {
		if pending {
			batchProjectID := projectID
			if hasMultipleProjects {
				batchProjectID = 0
			}

			if err := c.sendToSocket(Event{
				Type:      EventDatabaseChanged,
				ProjectID: batchProjectID,
				Timestamp: time.Now(),
			}); err != nil {
				if !isConnectionError(err) {
					slog.Error("failed to send batched event", "error", err)
				}
			}
			pending = false
		}
	}

	for {
		select {
		case <-c.ctx.Done():
			// Flush any pending events before exiting
			flushPending()
			return

		case event, ok := <-c.eventQueue:
			if !ok {
				// Channel closed - flush and exit
				flushPending()
				return
			}

			// Mark that we have pending events to send
			if !pending {
				pending = true
				projectID = event.ProjectID
				hasMultipleProjects = false
			} else if projectID != event.ProjectID && event.ProjectID != 0 {
				// Different project detected - use projectID 0 (all projects)
				hasMultipleProjects = true
			}

			// Continue draining the queue to batch multiple events together
			// This loop drains any other events queued during this batch window
		drainLoop:
			for {
				select {
				case evt, ok := <-c.eventQueue:
					if !ok {
						break drainLoop
					}
					// Check if we have events from different projects
					if projectID != evt.ProjectID && evt.ProjectID != 0 {
						hasMultipleProjects = true
					}
				default:
					break drainLoop
				}
			}

		case <-ticker.C:
			flushPending()
		}
	}
}

// sendToSocket sends an event to the daemon socket.
func (c *Client) sendToSocket(event Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected to daemon")
	}

	// Set a short write deadline to detect dead connections
	if err := c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	msg := Message{
		Version: ProtocolVersion,
		Type:    "event",
		Event:   &event,
	}
	return c.encoder.Encode(msg)
}

// Listen starts listening for events from the daemon.
// It returns a channel that receives events and handles reconnection automatically.
// The channel is closed when context is done or reconnection fails.
func (c *Client) Listen(ctx context.Context) (<-chan Event, error) {
	if c == nil {
		// Return a closed channel when client is nil
		eventChan := make(chan Event)
		close(eventChan)
		return eventChan, fmt.Errorf("event client is nil")
	}
	eventChan := make(chan Event, 10)
	go c.listenLoop(ctx, eventChan)
	return eventChan, nil
}

// listenLoop reads events from the daemon and handles reconnection.
func (c *Client) listenLoop(ctx context.Context, eventChan chan Event) {
	defer close(eventChan)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := c.readEvents(ctx, eventChan)
			if err != nil {
				slog.Info("connection lost, reconnecting", "error", err)
				c.notify("info", "Connection lost, reconnecting...")

				if c.reconnect(ctx) {
					slog.Info("reconnected to daemon")
					c.notify("info", "Reconnected to daemon")
					continue
				}

				slog.Error("failed to reconnect after max attempts", "attempts", c.maxRetries)
				c.notify("error", fmt.Sprintf("Failed to reconnect after %d attempts", c.maxRetries))
				return
			}
		}
	}
}

// readEvents reads messages from the socket and sends them to the event channel.
func (c *Client) readEvents(ctx context.Context, eventChan chan Event) error {
	for {
		var msg Message

		c.mu.Lock()
		if c.conn == nil {
			c.mu.Unlock()
			return fmt.Errorf("connection closed")
		}
		// Set read deadline to detect hung connections (60 seconds)
		if err := c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			c.mu.Unlock()
			return fmt.Errorf("failed to set read deadline: %w", err)
		}
		decoder := c.decoder // Keep reference to decoder
		c.mu.Unlock()

		if err := decoder.Decode(&msg); err != nil {
			return fmt.Errorf("failed to decode message: %w", err)
		}

		// Check protocol version - log warning if mismatch
		if msg.Version != 0 && msg.Version != ProtocolVersion {
			slog.Warn("protocol version mismatch", "received", msg.Version, "expected", ProtocolVersion)
		}

		switch msg.Type {
		case "event":
			if msg.Event != nil {
				// Check for event ordering (basic duplicate detection)
				if msg.Event.SequenceID > c.lastSequence {
					c.lastSequence = msg.Event.SequenceID
					select {
					case eventChan <- *msg.Event:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}

		case "ping":
			// Respond to daemon ping with pong
			if err := c.sendToSocket(Event{Type: EventPong}); err != nil {
				// Broken pipe/connection closed is expected during disconnection
				if !isConnectionError(err) {
					slog.Debug("failed to send pong", "error", err)
				}
			}
		}
	}
}

// isConnectionError checks if an error is a network connection error
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "use of closed network connection")
}

// reconnect attempts to reconnect to the daemon with exponential backoff.
// It tries up to maxRetries times, doubling the delay each time.
func (c *Client) reconnect(ctx context.Context) bool {
	delay := c.baseDelay

	for i := 0; i < c.maxRetries; i++ {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(delay):
			// Try to reconnect
			c.mu.Lock()
			if c.conn != nil {
				if err := c.conn.Close(); err != nil {
					slog.Debug("error closing connection during reconnect", "error", err)
				}
			}
			c.mu.Unlock()

			if err := c.Connect(ctx); err == nil {
				slog.Info("reconnected to daemon", "attempt", i+1, "max_retries", c.maxRetries)
				c.notify("info", fmt.Sprintf("Reconnected (attempt %d/%d)", i+1, c.maxRetries))
				return true
			}

			slog.Info("reconnection attempt failed", "attempt", i+1, "max_retries", c.maxRetries, "retry_delay", delay)
			delay *= 2 // Exponential backoff: 1s, 2s, 4s, 8s, 16s
		}
	}

	return false
}

// Subscribe changes the subscription to a specific project.
// ProjectID 0 means subscribe to all projects.
func (c *Client) Subscribe(projectID int) error {
	if c == nil {
		return fmt.Errorf("event client is nil")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentProjectID = projectID

	msg := Message{
		Version: ProtocolVersion,
		Type:    "subscribe",
		Subscribe: &SubscribeMessage{
			ProjectID: projectID,
		},
	}

	if c.conn == nil {
		return fmt.Errorf("not connected to daemon")
	}

	return c.encoder.Encode(msg)
}

// Close closes the connection to the daemon and stops all goroutines.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	var err error
	c.closeOnce.Do(func() {
		// Cancel context to stop other goroutines
		c.cancel()

		// Check if we ever connected (and thus started the batcher)
		c.mu.Lock()
		wasConnected := c.conn != nil
		c.mu.Unlock()

		if wasConnected {
			// Close the event queue to signal no more events coming
			// This allows batcher to flush pending events before exiting
			close(c.eventQueue)

			// Wait for batcher to finish (it will flush pending events)
			<-c.batcherDone

			// Close the connection
			c.mu.Lock()
			if c.conn != nil {
				err = c.conn.Close()
				c.conn = nil
			}
			c.mu.Unlock()
		}
	})
	return err
}
