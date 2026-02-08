package assignee

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testutil"
)

func TestCreateTrimsWhitespace(t *testing.T) {
	db := testutil.SetupTestDB(t)

	svc, err := NewService(db, database.SQLite)
	require.NoError(t, err)

	ctx := context.Background()

	// Create with whitespace-padded name
	a, err := svc.Create(ctx, "  alice  ")
	require.NoError(t, err)
	assert.Equal(t, "alice", a.Name)

	// GetByName with trimmed name should find it
	found, err := svc.GetByName(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, a.ID, found.ID)

	// Creating again with 'alice' should fail (already exists)
	_, err = svc.Create(ctx, "alice")
	assert.Error(t, err)
}

func TestGetByNameTrimsWhitespace(t *testing.T) {
	db := testutil.SetupTestDB(t)

	svc, err := NewService(db, database.SQLite)
	require.NoError(t, err)

	ctx := context.Background()

	// Create normally
	a, err := svc.Create(ctx, "bob")
	require.NoError(t, err)

	// GetByName with whitespace should still find it
	found, err := svc.GetByName(ctx, "  bob  ")
	require.NoError(t, err)
	assert.Equal(t, a.ID, found.ID)
}
