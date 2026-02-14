package label

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

// testEnv provides a standard test environment for label service tests
type testEnv struct {
	DB        *sql.DB
	Dialect   fixtures.Dialect
	Svc       Service
	ProjectID int
	Ctx       context.Context
}

// setupTestEnv creates a standard test environment with:
// - Test project (bare, no default columns)
// - Label service
// - Background context
func setupTestEnv(tb testing.TB, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) *testEnv {
	tb.Helper()
	projectID := fixtures.CreateBareProject(tb, db, d, "Test Project")
	svc, err := NewService(db, dbType, nil)
	require.NoError(tb, err, "failed to create test service")
	return &testEnv{
		DB:        db,
		Dialect:   d,
		Svc:       svc,
		ProjectID: projectID,
		Ctx:       context.Background(),
	}
}
