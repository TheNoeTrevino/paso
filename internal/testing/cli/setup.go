package cli

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

// SetupCLITest creates an in-memory DB and returns both the DB and App instance.
// This function is only for CLI tests and is isolated in a separate package
// to avoid import cycles when service tests import testutil.
func SetupCLITest(tb testing.TB, opts ...app.Option) (*sql.DB, *app.App) {
	tb.Helper()
	db := fixtures.SetupTestDB(tb)

	// Create app instance with services.
	// Note: EventPublisher is nil — event publishing is tested elsewhere.
	appInstance, err := app.New(db, opts...)
	require.NoError(tb, err, "failed to create app instance")

	return db, appInstance
}

// SetupCLITestWithGit creates an in-memory DB and App with a mock git detector.
// Returns the DB, App, and mock so tests can configure git behavior.
func SetupCLITestWithGit(tb testing.TB) (*sql.DB, *app.App, *mocks.MockGitDetector) {
	tb.Helper()
	mock := mocks.NewMockGitDetector()
	db, appInstance := SetupCLITest(tb, app.WithGitDetector(mock))
	return db, appInstance, mock
}

// CreateTestProject wraps fixtures.CreateTestProject for CLI tests.
// Creates a test project with default columns (Todo, In Progress, Done).
func CreateTestProject(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	return fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), name)
}

// CreateTestProjectWithBranch wraps fixtures.CreateTestProjectWithBranch for CLI tests.
// Creates a project with default columns and a git branch association.
func CreateTestProjectWithBranch(t *testing.T, db *sql.DB, name, branch string) int {
	t.Helper()
	return fixtures.CreateTestProjectWithBranch(t, db, fixtures.SQLiteDialect(), name, branch)
}

// CreateBareProject wraps fixtures.CreateBareProject for CLI tests.
// Creates a project without any default columns.
func CreateBareProject(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	return fixtures.CreateBareProject(t, db, fixtures.SQLiteDialect(), name)
}

// CreateTestColumn wraps fixtures.CreateTestColumn for CLI tests.
// Creates a test column and returns its ID.
func CreateTestColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	return fixtures.CreateTestColumn(t, db, fixtures.SQLiteDialect(), projectID, name)
}

// CreateTestTask wraps fixtures.CreateTestTask for CLI tests.
// Creates a test task and returns its ID.
func CreateTestTask(t *testing.T, db *sql.DB, columnID int, title string) int {
	t.Helper()
	return fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, title)
}

// CreateTestLabel wraps fixtures.CreateTestLabel for CLI tests.
// Creates a test label and returns its ID.
func CreateTestLabel(t *testing.T, db *sql.DB, projectID int, name, color string) int {
	t.Helper()
	return fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, name, color)
}

// CreateTestAssignee wraps fixtures.CreateTestAssignee for CLI tests.
func CreateTestAssignee(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	return fixtures.CreateTestAssignee(t, db, fixtures.SQLiteDialect(), name)
}

// GetColumnIDByName wraps fixtures.GetColumnIDByName for CLI tests.
func GetColumnIDByName(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	return fixtures.GetColumnIDByName(t, db, fixtures.SQLiteDialect(), projectID, name)
}

// AttachLabelToTask wraps fixtures.AttachLabelToTask for CLI tests.
func AttachLabelToTask(t *testing.T, db *sql.DB, taskID, labelID int) {
	t.Helper()
	fixtures.AttachLabelToTask(t, db, fixtures.SQLiteDialect(), taskID, labelID)
}

// AddTaskSubtask wraps fixtures.AddTaskSubtask for CLI tests.
// Use relationTypeID=1 for parent/child, relationTypeID=2 for blocking.
func AddTaskSubtask(t *testing.T, db *sql.DB, parentID, childID, relationTypeID int) {
	t.Helper()
	fixtures.AddTaskSubtask(t, db, fixtures.SQLiteDialect(), parentID, childID, relationTypeID)
}

// CreateTestComment wraps fixtures.CreateTestComment for CLI tests.
func CreateTestComment(t *testing.T, db *sql.DB, taskID int, content, author string) int {
	t.Helper()
	return fixtures.CreateTestComment(t, db, fixtures.SQLiteDialect(), taskID, content, author)
}

// SetColumnHoldsCompletedTasks wraps fixtures.SetColumnHoldsCompletedTasks for CLI tests.
func SetColumnHoldsCompletedTasks(t *testing.T, db *sql.DB, columnID int) {
	t.Helper()
	fixtures.SetColumnHoldsCompletedTasks(t, db, fixtures.SQLiteDialect(), columnID)
}

// SetColumnHoldsReadyTasks wraps fixtures.SetColumnHoldsReadyTasks for CLI tests.
func SetColumnHoldsReadyTasks(t *testing.T, db *sql.DB, columnID int) {
	t.Helper()
	fixtures.SetColumnHoldsReadyTasks(t, db, fixtures.SQLiteDialect(), columnID)
}

// SetColumnHoldsInProgressTasks wraps fixtures.SetColumnHoldsInProgressTasks for CLI tests.
func SetColumnHoldsInProgressTasks(t *testing.T, db *sql.DB, columnID int) {
	t.Helper()
	fixtures.SetColumnHoldsInProgressTasks(t, db, fixtures.SQLiteDialect(), columnID)
}

// UpdateTaskDescription wraps fixtures.UpdateTaskDescription for CLI tests.
func UpdateTaskDescription(t *testing.T, db *sql.DB, taskID int, description string) {
	t.Helper()
	fixtures.UpdateTaskDescription(t, db, fixtures.SQLiteDialect(), taskID, description)
}

// UpdateTaskFields wraps fixtures.UpdateTaskFields for CLI tests.
func UpdateTaskFields(t *testing.T, db *sql.DB, taskID int, fields map[string]any) {
	t.Helper()
	fixtures.UpdateTaskFields(t, db, fixtures.SQLiteDialect(), taskID, fields)
}

// LinkProjectGitBranch wraps fixtures.LinkProjectGitBranch for CLI tests.
func LinkProjectGitBranch(t *testing.T, db *sql.DB, projectID int, branchName string) {
	t.Helper()
	fixtures.LinkProjectGitBranch(t, db, fixtures.SQLiteDialect(), projectID, branchName)
}
