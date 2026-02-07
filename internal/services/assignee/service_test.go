package assignee

import (
	"context"
	"sync"
	"testing"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testutil"
)

func TestGetOrCreate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	svc, err := NewService(db, database.SQLite)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	ctx := context.Background()

	// First call should create
	a1, err := svc.GetOrCreate(ctx, "alice")
	if err != nil {
		t.Fatalf("first GetOrCreate failed: %v", err)
	}
	if a1.Name != "alice" {
		t.Errorf("expected name 'alice', got '%s'", a1.Name)
	}

	// Second call should return existing
	a2, err := svc.GetOrCreate(ctx, "alice")
	if err != nil {
		t.Fatalf("second GetOrCreate failed: %v", err)
	}
	if a2.ID != a1.ID {
		t.Errorf("expected same ID %d, got %d", a1.ID, a2.ID)
	}
}

func TestGetOrCreateConcurrent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	// SQLite in-memory DBs are per-connection; limit to one connection so all
	// goroutines share the same schema and data.
	db.SetMaxOpenConns(1)

	svc, err := NewService(db, database.SQLite)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	const goroutines = 10
	const name = "concurrent-test-user"

	var wg sync.WaitGroup
	results := make(chan int, goroutines)
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assignee, err := svc.GetOrCreate(context.Background(), name)
			if err != nil {
				errors <- err
				return
			}
			results <- assignee.ID
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	for err := range errors {
		t.Errorf("GetOrCreate returned error: %v", err)
	}

	var firstID int
	first := true
	for id := range results {
		if first {
			firstID = id
			first = false
			continue
		}
		if id != firstID {
			t.Errorf("got different IDs: %d and %d", firstID, id)
		}
	}

	if first {
		t.Fatal("no results received")
	}
}
