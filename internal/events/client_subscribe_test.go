package events

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestSubscribe_BeforeConnect(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "paso.sock")

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Try to subscribe before connecting
	err = client.Subscribe(1)
	if err == nil {
		t.Error("Expected Subscribe to fail before connecting")
	}

	t.Logf("✓ Subscribe correctly fails before connection")
}

func TestSubscribe_AfterConnect(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe message (project 0)
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" || msg.Subscribe == nil || msg.Subscribe.ProjectID != 0 {
			t.Fatalf("Expected initial subscribe for project 0, got: %+v", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Subscribe to project 5
	if err := client.Subscribe(5); err != nil {
		t.Fatalf("Expected Subscribe to succeed, got error: %v", err)
	}

	// Wait for subscribe message
	select {
	case msg := <-messages:
		if msg.Type != "subscribe" {
			t.Errorf("Expected subscribe message, got: %s", msg.Type)
		}
		if msg.Subscribe == nil {
			t.Fatal("Expected subscribe message to have Subscribe field")
		}
		if msg.Subscribe.ProjectID != 5 {
			t.Errorf("Expected project ID 5, got %d", msg.Subscribe.ProjectID)
		}
		t.Logf("✓ Subscribe message sent correctly for project 5")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for subscribe message")
	}

	// Verify client's currentProjectID is updated
	client.mu.Lock()
	currentProject := client.currentProjectID
	client.mu.Unlock()

	if currentProject != 5 {
		t.Errorf("Expected currentProjectID to be 5, got %d", currentProject)
	}
}

func TestSubscribe_MultipleProjects(t *testing.T) {
	socketPath, listener, messages := setupMockDaemon(t)
	defer func() { _ = listener.Close() }()

	client, err := NewClient(socketPath)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Drain initial subscribe message (project 0)
	select {
	case <-messages:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for initial subscribe")
	}

	// Subscribe to multiple projects
	projects := []int{1, 2, 3}
	for _, projectID := range projects {
		if err := client.Subscribe(projectID); err != nil {
			t.Fatalf("Failed to subscribe to project %d: %v", projectID, err)
		}
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
			t.Fatal("Timeout waiting for subscribe messages")
		}
	}

	for _, projectID := range projects {
		if !receivedProjects[projectID] {
			t.Errorf("Did not receive subscribe message for project %d", projectID)
		}
	}

	t.Logf("✓ Multiple subscribe messages sent correctly")
}
