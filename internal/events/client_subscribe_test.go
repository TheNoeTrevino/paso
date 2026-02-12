package events

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribe_BeforeConnect(t *testing.T) {
	t.Parallel()
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Try to subscribe before connecting
	err = client.Subscribe(1)
	assert.Error(t, err)

	t.Logf("Subscribe correctly fails before connection")
}

func TestSubscribe_AfterConnect(t *testing.T) {
	t.Parallel()
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Drain initial subscribe message (project 0)
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" || msg.Subscribe == nil || msg.Subscribe.ProjectID != 0 {
			require.Fail(t, "Expected initial subscribe for project 0", "got: %+v", msg)
		}
	case <-time.After(1 * time.Second):
		require.Fail(t, "Timeout waiting for initial subscribe")
	}

	// Subscribe to project 5
	err = client.Subscribe(5)
	require.NoError(t, err)

	// Wait for subscribe message
	select {
	case msg := <-messages:
		assert.Equal(t, "subscribe", msg.Type)
		require.NotNil(t, msg.Subscribe)
		assert.Equal(t, 5, msg.Subscribe.ProjectID)
		t.Logf("Subscribe message sent correctly for project 5")
	case <-time.After(1 * time.Second):
		require.Fail(t, "Timeout waiting for subscribe message")
	}

	// Verify client's currentProjectID is updated
	client.mu.Lock()
	currentProject := client.currentProjectID
	client.mu.Unlock()

	assert.Equal(t, 5, currentProject)
}

func TestSubscribe_MultipleProjects(t *testing.T) {
	t.Parallel()
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	require.NoError(t, err)

	// Drain initial subscribe message (project 0)
	select {
	case <-messages:
	case <-time.After(1 * time.Second):
		require.Fail(t, "Timeout waiting for initial subscribe")
	}

	// Subscribe to multiple projects
	projects := []int{1, 2, 3}
	for _, projectID := range projects {
		err := client.Subscribe(projectID)
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)
	}

	// Verify we received all subscribe messages
	receivedProjects := make(map[int]bool)
	timeout := time.After(2 * time.Second)

	for i := 0; i < len(projects); i++ {
		select {
		case msg := <-messages:
			if msg.Type == "subscribe" && msg.Subscribe != nil {
				receivedProjects[msg.Subscribe.ProjectID] = true
			}
		case <-timeout:
			require.Fail(t, "Timeout waiting for subscribe messages")
		}
	}

	for _, projectID := range projects {
		assert.True(t, receivedProjects[projectID])
	}

	t.Logf("Multiple subscribe messages sent correctly")
}
