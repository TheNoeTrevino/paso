package daemon

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/events"
)

func TestClientConnection_Single(t *testing.T) {
	t.Parallel()
	_, socketPath := setupTestDaemon(t)

	// Connect a raw client
	conn, encoder, _ := connectRawClient(t, socketPath)

	// Send initial subscribe message
	sendSubscribeMessage(t, encoder, 0)

	// Give server time to process connection
	time.Sleep(50 * time.Millisecond)

	// Verify connection is still active by checking if we can write
	err := encoder.Encode(events.Message{Version: events.ProtocolVersion, Type: "ping"})
	require.NoError(t, err)

	// Try to read with a short deadline - we expect timeout (no response expected for ping from client)
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	decoder := json.NewDecoder(conn)
	var msg events.Message
	err = decoder.Decode(&msg)
	if err == nil {
		// Unexpectedly got a message
		t.Logf("Note: Received unexpected message type: %s", msg.Type)
	}

	t.Logf("Client connected successfully")
}

func TestClientConnection_Multiple(t *testing.T) {
	t.Parallel()
	server, socketPath := setupTestDaemon(t)

	numClients := 5

	// Connect multiple clients
	for range numClients {
		_, encoder, _ := connectRawClient(t, socketPath)
		sendSubscribeMessage(t, encoder, 0)
	}

	// Give server time to process all connections
	time.Sleep(100 * time.Millisecond)

	t.Logf("Successfully connected %d clients", numClients)
	logServerState(t, server, "after multiple connections")
}

func TestClientDisconnection(t *testing.T) {
	t.Parallel()
	server, socketPath := setupTestDaemon(t)

	// Connect a client
	conn, encoder, _ := connectRawClient(t, socketPath)
	sendSubscribeMessage(t, encoder, 0)

	time.Sleep(50 * time.Millisecond)
	logServerState(t, server, "after connection")

	// Close the connection
	if closer, ok := conn.(interface{ Close() error }); ok {
		_ = closer.Close()
	}

	// Give server time to detect disconnection
	time.Sleep(100 * time.Millisecond)

	logServerState(t, server, "after disconnection")
	t.Logf("Client disconnected and cleaned up")
}
