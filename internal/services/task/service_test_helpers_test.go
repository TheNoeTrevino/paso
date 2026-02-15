package task

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/services/taskevent"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

type testEnv struct {
	DB        *sql.DB
	Dialect   fixtures.Dialect
	Svc       Service
	ProjectID int
	ColumnID  int
	Ctx       context.Context
}

func setupTestEnv(tb testing.TB, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) *testEnv {
	tb.Helper()
	projectID := fixtures.CreateBareProject(tb, db, d, "Test Project")
	columnID := fixtures.CreateTestColumn(tb, db, d, projectID, "To Do")
	svc, err := NewService(db, dbType, nil, nil)
	require.NoError(tb, err, "failed to create test service")
	return &testEnv{
		DB:        db,
		Dialect:   d,
		Svc:       svc,
		ProjectID: projectID,
		ColumnID:  columnID,
		Ctx:       context.Background(),
	}
}

func setupTestEnvWithEventMock(tb testing.TB, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) (*testEnv, *taskevent.MockService) {
	tb.Helper()
	mock := taskevent.NewMockService()
	svc, err := NewService(db, dbType, nil, mock)
	require.NoError(tb, err, "failed to create test service with mock")
	projectID := fixtures.CreateBareProject(tb, db, d, "Test Project")
	columnID := fixtures.CreateTestColumn(tb, db, d, projectID, "To Do")
	return &testEnv{
		DB:        db,
		Dialect:   d,
		Svc:       svc,
		ProjectID: projectID,
		ColumnID:  columnID,
		Ctx:       context.Background(),
	}, mock
}

// testDialect is a package-level SQLite dialect for benchmark tests that
// don't use RunDatabaseTests (benchmarks stay SQLite-only).
var testDialect = fixtures.SQLiteDialect()

func ptrString(s string) *string {
	return &s
}
