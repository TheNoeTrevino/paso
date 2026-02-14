package assignee

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestGetOrCreate(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)

		ctx := context.Background()

		// First call should create
		a1, err := svc.GetOrCreate(ctx, "alice")
		require.NoError(t, err)
		assert.Equal(t, "alice", a1.Name)

		// Second call should return existing
		a2, err := svc.GetOrCreate(ctx, "alice")
		require.NoError(t, err)
		assert.Equal(t, a1.ID, a2.ID)
	})
}

func TestGetOrCreateConcurrent(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		// SQLite in-memory DBs are per-connection; limit to one connection so all
		// goroutines share the same schema and data.
		if dbType == database.SQLite {
			db.SetMaxOpenConns(1)
		}

		svc, err := NewService(db, dbType)
		require.NoError(t, err)

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
			assert.NoError(t, err, "GetOrCreate returned error: %v", err)
		}

		var firstID int
		first := true
		for id := range results {
			if first {
				firstID = id
				first = false
				continue
			}
			assert.Equal(t, firstID, id)
		}

		require.False(t, first, "no results received")
	})
}
