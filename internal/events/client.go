package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client represents a connection to the Paso daemon for receiving live updates.
// It handles event sending, receiving, batching, reconnection, and subscriptions.
type Client struct {
	socketPath       string
	conn             net.Conn
	encoder          *json.Encoder
	decoder          *json.Decoder
	mu               sync.Mutex

	// Batching configuration
	eventQueue chan Event
	debounce   time.Duration
	closed     bool // Prevent double-close panics

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
		Type: "subscribe",
		Subscribe: &SubscribeMessage{
			ProjectID: 0,
		},
	}
	if err := c.encoder.Encode(msg); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Error closing connection: %v", closeErr)
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
					log.Printf("Failed to send batched event: %v", err)
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
		Type:  "event",
		Event: &event,
	}
	return c.encoder.Encode(msg)
}

// Listen starts listening for events from the daemon.
// It returns a channel that receives events and handles reconnection automatically.
// The channel is closed when context is done or reconnection fails.
func (c *Client) Listen(ctx context.Context) (<-chan Event, error) {
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
				log.Printf("Connection lost: %v, reconnecting...", err)

				if c.reconnect(ctx) {
					log.Printf("Reconnected to daemon")
					continue
				}

				log.Printf("Failed to reconnect after %d attempts, giving up", c.maxRetries)
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
					log.Printf("Failed to send pong: %v", err)
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
					log.Printf("Error closing connection during reconnect: %v", err)
				}
			}
			c.mu.Unlock()

			if err := c.Connect(ctx); err == nil {
				log.Printf("Reconnected to daemon (attempt %d/%d)", i+1, c.maxRetries)
				return true
			}

			log.Printf("Reconnection attempt %d/%d failed, retrying in %v", i+1, c.maxRetries, delay)
			delay *= 2 // Exponential backoff: 1s, 2s, 4s, 8s, 16s
		}
	}

	return false
}

// Subscribe changes the subscription to a specific project.
// ProjectID 0 means subscribe to all projects.
func (c *Client) Subscribe(projectID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentProjectID = projectID

	msg := Message{
		Type: "subscribe",
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
	// Check if already closed
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true

	// Close the event queue to signal no more events coming
	// This allows batcher to flush pending events before exiting
	if c.eventQueue != nil {
		close(c.eventQueue)
	}
	c.mu.Unlock()

	// Cancel context to stop other goroutines
	c.cancel()

	// Wait for batcher to finish (it will flush pending events)
	<-c.batcherDone

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
