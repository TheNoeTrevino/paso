package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thenoetrevino/paso/internal/events"
)

// client represents a connected client to the daemon
type client struct {
	conn         net.Conn
	send         chan events.Message
	subscription events.SubscribeMessage
	lastPong     time.Time
	mu           sync.Mutex // Protects subscription and lastPong
	closeOnce    sync.Once  // Ensures send channel is closed only once
}

// Server represents the Paso event daemon
type Server struct {
	socketPath       string
	listener         net.Listener
	clients          map[*client]bool
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	broadcast        chan events.Event
	metrics          *Metrics
	sequenceCounter  atomic.Int64
	clientBufferSize int // Configurable client send queue size
	shutdownOnce     sync.Once
}

// getEnvInt reads an integer from an environment variable, returning defaultVal if not set or invalid
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultVal
}

// NewServer creates a new daemon server
func NewServer(socketPath string) (*Server, error) {
	// Ensure the directory exists
	dir := filepath.Dir(socketPath)

	if dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create socket directory: %w", err)
		}
	}

	// Remove stale socket file if it exists
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return nil, fmt.Errorf("failed to remove stale socket: %w", err)
		}
	}

	// Create Unix domain socket listener
	lc := net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket listener: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Read buffer sizes from environment variables (configurable for performance tuning)
	broadcastBuffer := getEnvInt("PASO_DAEMON_BROADCAST_BUFFER", 100)
	clientBuffer := getEnvInt("PASO_DAEMON_CLIENT_BUFFER", 10)

	return &Server{
		socketPath:       socketPath,
		listener:         listener,
		clients:          make(map[*client]bool),
		ctx:              ctx,
		cancel:           cancel,
		broadcast:        make(chan events.Event, broadcastBuffer),
		metrics:          NewMetrics(),
		sequenceCounter:  atomic.Int64{},
		clientBufferSize: clientBuffer,
	}, nil
}

// Start runs the daemon server
// It starts three main goroutines: accept, broadcast, and health monitoring
func (s *Server) Start(ctx context.Context) error {
	log.Printf("Daemon starting, listening on %s", s.socketPath)

	// Create a combined context that cancels when either the daemon context or caller context is done
	combinedCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-s.ctx.Done()
		cancel()
	}()

	// Start accept loop
	acceptErr := make(chan error, 1)
	go func() {
		acceptErr <- s.acceptLoop(combinedCtx)
	}()

	// Start broadcast loop
	go s.broadcastLoop(combinedCtx)

	// Start health monitor
	go s.monitorHealth(combinedCtx)

	// Wait for context or accept error
	select {
	case <-combinedCtx.Done():
		log.Println("Daemon context cancelled, shutting down")
	case err := <-acceptErr:
		if err != nil {
			log.Printf("Accept loop error: %v", err)
		}
	}

	// Graceful shutdown
	return s.Shutdown()
}

// acceptLoop accepts incoming client connections
func (s *Server) acceptLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Set a read deadline so we can check for context cancellation
		if err := s.listener.(*net.UnixListener).SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
			log.Printf("Error setting listener deadline: %v", err)
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return fmt.Errorf("accept error: %w", err)
		}

		// Create new client
		c := &client{
			conn:     conn,
			send:     make(chan events.Message, s.clientBufferSize),
			lastPong: time.Now(),
		}

		// Register client
		s.mu.Lock()
		s.clients[c] = true
		s.mu.Unlock()

		// Update metrics
		s.updateClientCount()

		log.Printf("Client connected, total clients: %d", s.getClientCount())

		// Start client handler goroutines
		go s.handleClient(c)
		go s.clientWriter(c)
	}
}

// broadcastLoop distributes events to subscribed clients
func (s *Server) broadcastLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case event := <-s.broadcast:
			// Add sequence number to event
			event.SequenceID = s.sequenceCounter.Add(1)

			s.metrics.IncRefreshesTotal()

			// Send to all subscribed clients
			s.mu.RLock()
			for c := range s.clients {
				// Check if client is subscribed to this project (protected by client mutex)
				c.mu.Lock()
				// Send event if: event is for all projects (0), OR client subscribed to all (0), OR client subscribed to specific project
				isSubscribed := event.ProjectID == 0 || c.subscription.ProjectID == 0 || c.subscription.ProjectID == event.ProjectID
				c.mu.Unlock()

				if isSubscribed {
					msg := events.Message{
						Version: events.ProtocolVersion,
						Type:    "event",
						Event:   &event,
					}

					// Non-blocking send - if client is slow, skip
					if !s.sendToClient(c, msg) {
						log.Printf("Client send queue full, event dropped")
					}
				}
			}
			s.mu.RUnlock()
		}
	}
}

// handleClient reads messages from a connected client
func (s *Server) handleClient(c *client) {
	defer func() {
		s.removeClient(c)
		log.Printf("Client disconnected, total clients: %d", s.getClientCount())
	}()

	decoder := json.NewDecoder(c.conn)

	for {
		var msg events.Message

		if err := decoder.Decode(&msg); err != nil {
			return
		}

		// Check protocol version - log warning if mismatch
		if msg.Version != 0 && msg.Version != events.ProtocolVersion {
			log.Printf("Warning: received message with protocol version %d, expected %d", msg.Version, events.ProtocolVersion)
		}

		switch msg.Type {
		case "event":
			if msg.Event != nil {
				s.metrics.IncEventsReceived()
				// Broadcast event to other clients
				select {
				case s.broadcast <- *msg.Event:
				default:
					log.Printf("Broadcast channel full")
				}
			}

		case "subscribe":
			if msg.Subscribe != nil {
				c.mu.Lock()
				c.subscription = *msg.Subscribe
				c.mu.Unlock()
				log.Printf("Client subscribed to project %d", msg.Subscribe.ProjectID)
			}

		case "pong":
			c.mu.Lock()
			c.lastPong = time.Now()
			c.mu.Unlock()
		}
	}
}

// clientWriter sends messages to a client
func (s *Server) clientWriter(c *client) {
	encoder := json.NewEncoder(c.conn)

	for msg := range c.send {
		if err := encoder.Encode(msg); err != nil {
			return
		}
	}
}

// monitorHealth sends ping messages and removes stale clients
func (s *Server) monitorHealth(ctx context.Context) {
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	healthTicker := time.NewTicker(60 * time.Second)
	defer healthTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-pingTicker.C:
			// Send ping to all clients
			s.mu.RLock()
			clients := make([]*client, 0, len(s.clients))
			for c := range s.clients {
				clients = append(clients, c)
			}
			s.mu.RUnlock()

			pingMsg := events.Message{
				Version: events.ProtocolVersion,
				Type:    "ping",
				Event: &events.Event{
					Type: events.EventPing,
				},
			}

			for _, c := range clients {
				if !s.sendToClient(c, pingMsg) {
					log.Printf("Failed to send ping to client (queue full)")
				}
			}

		case <-healthTicker.C:
			// Remove stale clients that haven't responded to ping in 90s
			// Use two-phase locking to avoid deadlock: collect clients, then process
			s.mu.RLock()
			staleClients := make([]*client, 0)
			now := time.Now()
			for c := range s.clients {
				c.mu.Lock()
				lastPong := c.lastPong
				c.mu.Unlock()

				if now.Sub(lastPong) > 90*time.Second {
					staleClients = append(staleClients, c)
				}
			}
			s.mu.RUnlock()

			// Remove stale clients (outside of server lock to avoid deadlock)
			for _, c := range staleClients {
				log.Printf("Removing stale client (last pong: %v ago)", now.Sub(c.lastPong))
				s.removeClient(c)
			}
		}
	}
}

// Broadcast sends an event to the broadcast channel (non-blocking)
func (s *Server) Broadcast(event events.Event) error {
	select {
	case s.broadcast <- event:
		return nil
	default:
		return fmt.Errorf("broadcast channel full")
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	var err error
	s.shutdownOnce.Do(func() {
		log.Println("Shutting down daemon...")

		s.cancel()

		// Close listener
		if s.listener != nil {
			if closeErr := s.listener.Close(); closeErr != nil {
				log.Printf("Error closing listener: %v", closeErr)
			}
		}

		// Close all client connections
		s.mu.Lock()
		for c := range s.clients {
			if closeErr := c.conn.Close(); closeErr != nil {
				log.Printf("Error closing client connection: %v", closeErr)
			}
			c.closeOnce.Do(func() {
				close(c.send)
			})
		}
		s.clients = make(map[*client]bool)
		s.mu.Unlock()

		// Remove socket file
		if removeErr := os.Remove(s.socketPath); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Printf("Warning: failed to remove socket file: %v", removeErr)
		}

		// Close broadcast channel
		close(s.broadcast)
	})

	return err
}

// Helper methods

func (s *Server) getClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

func (s *Server) updateClientCount() {
	count := s.getClientCount()
	s.metrics.SetConnectedClients(int32(count))
}

// removeClient safely removes a client from the server
func (s *Server) removeClient(c *client) {
	s.mu.Lock()
	delete(s.clients, c)
	s.mu.Unlock()

	if err := c.conn.Close(); err != nil {
		log.Printf("Error closing client connection: %v", err)
	}
	c.closeOnce.Do(func() {
		close(c.send)
	})

	s.updateClientCount()
}

// sendToClient attempts to send a message to a client (non-blocking)
// Returns true if successful, false if the queue is full
func (s *Server) sendToClient(c *client, msg events.Message) bool {
	select {
	case c.send <- msg:
		s.metrics.IncEventsSent()
		return true
	default:
		return false
	}
}
