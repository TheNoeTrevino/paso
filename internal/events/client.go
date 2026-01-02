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
	batcherDone    chan struct{}
	batcherRunning bool // Track if batcher is currently running

	// Notification callback for user-facing messages
	onNotify NotifyFunc

	// Write deadline duration (configurable for tests)
	writeDeadline time.Duration
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
		writeDeadline:    5 * time.Second, // Default: 5 seconds
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

// setWriteDeadlineForTest allows tests to use shorter deadlines.
// This is an unexported method for test use only.
func (c *Client) setWriteDeadlineForTest(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDeadline = d
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

	// Stop any existing batcher before starting a new one
	// This prevents multiple batchers from running simultaneously
	if c.batcherRunning {
		oldBatcherDone := c.batcherDone
		c.batcherRunning = false

		// Close eventQueue to signal batcher to exit
		// We'll recreate it after the old batcher stops
		close(c.eventQueue)

		// Wait for old batcher to finish (unlock while waiting)
		c.mu.Unlock()
		<-oldBatcherDone
		c.mu.Lock()

		// Recreate eventQueue for the new batcher
		c.eventQueue = make(chan Event, 100)
	}

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

	// Start the new batching goroutine
	c.batcherDone = make(chan struct{})
	c.batcherRunning = true
	go c.startBatcher()

	return nil
}

// SendEvent queues an event to be sent to the daemon with backpressure handling.
// Events are batched and sent in bursts within the debounce window.
// Implements retry with exponential backoff to handle queue saturation gracefully.
//
// Backpressure Strategy (Option B):
// - When queue is full, retries with exponential backoff instead of immediate failure
// - Max retries: 3 attempts (50ms, 100ms, 200ms delays)
// - This ensures critical events aren't silently dropped during high throughput
// - Logging at WARN level for queue saturation to aid debugging
//
// Returns error only if all retry attempts fail after max retries.
func (c *Client) SendEvent(event Event) error {
	if c == nil {
		return fmt.Errorf("event client is nil")
	}

	// First attempt
	select {
	case c.eventQueue <- event:
		return nil
	default:
		// Queue is full - implement backpressure with retry
		return c.sendEventWithRetry(event)
	}
}

// sendEventWithRetry attempts to queue an event with exponential backoff.
// This handles queue saturation gracefully by retrying instead of silently dropping.
func (c *Client) sendEventWithRetry(event Event) error {
	maxRetries := 3
	baseDelay := 50 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Log queue saturation at WARN level (only on first saturation)
		if attempt == 0 {
			slog.Warn("event queue saturated, applying backpressure",
				"event_type", event.Type,
				"project_id", event.ProjectID,
				"queue_capacity", 100) // Fixed queue size from line 73
		}

		// Don't sleep before the first attempt (we already tried above)
		if attempt > 0 {
			// Exponential backoff: 50ms, 100ms, 200ms
			delay := baseDelay * (1 << (attempt - 1))
			time.Sleep(delay)
		}

		// Try to send again
		select {
		case c.eventQueue <- event:
			if attempt > 0 {
				slog.Debug("event queued after retry",
					"attempt", attempt+1,
					"event_type", event.Type,
					"project_id", event.ProjectID)
			}
			return nil
		default:
			// Queue still full, will retry unless this was the last attempt
			if attempt < maxRetries-1 {
				delay := baseDelay * (1 << attempt)
				slog.Debug("event queue full, retrying",
					"attempt", attempt+1,
					"max_retries", maxRetries,
					"retry_delay", delay,
					"event_type", event.Type)
			}
		}
	}

	// All retry attempts failed
	slog.Error("event queue full after all retries, dropping event",
		"attempts", maxRetries,
		"event_type", event.Type,
		"project_id", event.ProjectID)

	return fmt.Errorf("event queue full (all %d retry attempts exhausted)", maxRetries)
}

// startBatcher runs in a goroutine and batches events from the queue.
// It sends a single event every debounce duration if any events are pending.
// If events from multiple projects are batched together, sends projectID 0 (all projects).
func (c *Client) startBatcher() {
	defer func() {
		close(c.batcherDone)
		c.mu.Lock()
		c.batcherRunning = false
		c.mu.Unlock()
	}()

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
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeDeadline)); err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	msg := Message{
		Version: ProtocolVersion,
		Type:    "event",
		Event:   &event,
	}
	err := c.encoder.Encode(msg)

	// Clear the deadline after writing to avoid affecting future operations
	if clearErr := c.conn.SetWriteDeadline(time.Time{}); clearErr != nil {
		slog.Debug("failed to clear write deadline", "error", clearErr)
	}

	return err
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

		// TOCTOU Fix: Use careful synchronization to prevent race conditions.
		// We hold the lock only to check state and set deadline, then release before
		// the potentially blocking Decode() call. This prevents deadlocks while
		// ensuring the decoder remains valid throughout its use.
		//
		// Strategy: Double-check pattern with state invariant checking
		// - conn and decoder are always set/cleared together (in Connect/Close)
		// - We verify conn != nil before Decode
		// - If Close() closes the connection, Decode() will detect EOF error
		// - We don't hold locks during long blocking operations

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
		// Copy reference while lock is held
		decoder := c.decoder
		c.mu.Unlock()

		// Decode with unlocked decoder reference
		// Safe because:
		// 1. decoder pointer value is immutable once copied (Go passes pointers by value)
		// 2. Close() invalidates the underlying connection, detected by Decode() -> EOF
		// 3. No data structures are modified during Decode
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
				// Restore previous subscription after reconnect
				if c.currentProjectID > 0 {
					slog.Info("restoring subscription after reconnect", "project_id", c.currentProjectID)
					if err := c.Subscribe(c.currentProjectID); err != nil {
						slog.Error("failed to restore subscription after reconnect", "project_id", c.currentProjectID, "error", err)
					}
				}
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

	// Set a write deadline for the subscription message
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeDeadline)); err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	err := c.encoder.Encode(msg)

	// Clear the deadline after writing
	if clearErr := c.conn.SetWriteDeadline(time.Time{}); clearErr != nil {
		slog.Debug("failed to clear write deadline after subscribe", "error", clearErr)
	}

	return err
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
			// Check if batcher is running before closing channels
			c.mu.Lock()
			batcherRunning := c.batcherRunning
			c.mu.Unlock()

			if batcherRunning {
				// Close the event queue to signal no more events coming
				// This allows batcher to flush pending events before exiting
				close(c.eventQueue)

				// Wait for batcher to finish (it will flush pending events)
				<-c.batcherDone
			}

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
