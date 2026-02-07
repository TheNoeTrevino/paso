package cli

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// SetupCLITest creates an in-memory DB and returns both the DB and App instance
// This function is only for CLI tests and is isolated in a separate package
// to avoid import cycles when service tests import testutil
func SetupCLITest(tb testing.TB) (*sql.DB, *app.App) {
	tb.Helper()
	db := testutil.SetupTestDB(tb)

	// Create app instance with services
	// Note: EventPublisher is nil - event publishing is tested elsewhere
	appInstance, err := app.New(db)
	require.NoError(tb, err, "failed to create app instance")

	return db, appInstance
}

// CreateTestProject wraps testutil.CreateTestProject for CLI tests
// Creates a test project with default columns (Todo, In Progress, Done)
func CreateTestProject(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	return testutil.CreateTestProject(t, db, testutil.SQLiteDialect(), name)
}

// CreateTestColumn wraps testutil.CreateTestColumn for CLI tests
// Creates a test column and returns its ID
func CreateTestColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	return testutil.CreateTestColumn(t, db, testutil.SQLiteDialect(), projectID, name)
}

// CreateTestTask wraps testutil.CreateTestTask for CLI tests
// Creates a test task and returns its ID
func CreateTestTask(t *testing.T, db *sql.DB, columnID int, title string) int {
	t.Helper()
	return testutil.CreateTestTask(t, db, testutil.SQLiteDialect(), columnID, title)
}

// CreateTestLabel wraps testutil.CreateTestLabel for CLI tests
// Creates a test label and returns its ID
func CreateTestLabel(t *testing.T, db *sql.DB, projectID int, name, color string) int {
	t.Helper()
	return testutil.CreateTestLabel(t, db, testutil.SQLiteDialect(), projectID, name, color)
}
