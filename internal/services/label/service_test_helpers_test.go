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
// - Test database
// - Test project (bare, no default columns)
// - Label service
// - Background context
func setupTestEnv(tb testing.TB) *testEnv {
	tb.Helper()
	db := fixtures.SetupTestDB(tb)
	dialect := fixtures.SQLiteDialect()
	projectID := fixtures.CreateBareProject(tb, db, dialect, "Test Project")
	svc, err := NewService(db, database.SQLite, nil)
	require.NoError(tb, err, "failed to create test service")
	return &testEnv{
		DB:        db,
		Dialect:   dialect,
		Svc:       svc,
		ProjectID: projectID,
		Ctx:       context.Background(),
	}
}
