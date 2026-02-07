package assignee

import (
	"context"
	"testing"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testutil"
)

func TestCreateTrimsWhitespace(t *testing.T) {
	db := testutil.SetupTestDB(t)

	svc, err := NewService(db, database.SQLite)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create with whitespace-padded name
	a, err := svc.Create(ctx, "  alice  ")
	if err != nil {
		t.Fatalf("Create(' alice ') failed: %v", err)
	}
	if a.Name != "alice" {
		t.Errorf("expected trimmed name 'alice', got '%s'", a.Name)
	}

	// GetByName with trimmed name should find it
	found, err := svc.GetByName(ctx, "alice")
	if err != nil {
		t.Fatalf("GetByName('alice') failed: %v", err)
	}
	if found.ID != a.ID {
		t.Errorf("expected same ID %d, got %d", a.ID, found.ID)
	}

	// Creating again with 'alice' should fail (already exists)
	_, err = svc.Create(ctx, "alice")
	if err == nil {
		t.Error("expected error creating duplicate 'alice', got nil")
	}
}

func TestGetByNameTrimsWhitespace(t *testing.T) {
	db := testutil.SetupTestDB(t)

	svc, err := NewService(db, database.SQLite)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create normally
	a, err := svc.Create(ctx, "bob")
	if err != nil {
		t.Fatalf("Create('bob') failed: %v", err)
	}

	// GetByName with whitespace should still find it
	found, err := svc.GetByName(ctx, "  bob  ")
	if err != nil {
		t.Fatalf("GetByName(' bob ') failed: %v", err)
	}
	if found.ID != a.ID {
		t.Errorf("expected same ID %d, got %d", a.ID, found.ID)
	}
}
