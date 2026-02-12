package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestDeleteAssigneeSetNullBehavior(t *testing.T) {
	t.Parallel()
	db := fixtures.SetupTestDB(t)

	ctx := context.Background()

	// Create a project and column for the task
	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "test-project")
	// CreateTestProject creates 3 columns, get the first one
	var columnID int
	err := db.QueryRowContext(ctx, "SELECT id FROM columns WHERE project_id = ? LIMIT 1", projectID).Scan(&columnID)
	require.NoError(t, err)

	// Create an assignee
	result, err := db.ExecContext(ctx, "INSERT INTO assignees (name) VALUES (?)", "test-user")
	require.NoError(t, err)
	assigneeID, err := result.LastInsertId()
	require.NoError(t, err)

	// Create a task assigned to this assignee
	taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "test-task")
	_, err = db.ExecContext(ctx, "UPDATE tasks SET assignee_id = ? WHERE id = ?", assigneeID, taskID)
	require.NoError(t, err)

	// Verify assignment
	var gotAssigneeID sql.NullInt64
	err = db.QueryRowContext(ctx, "SELECT assignee_id FROM tasks WHERE id = ?", taskID).Scan(&gotAssigneeID)
	require.NoError(t, err)
	require.True(t, gotAssigneeID.Valid && gotAssigneeID.Int64 == assigneeID)

	// Delete the assignee
	_, err = db.ExecContext(ctx, "DELETE FROM assignees WHERE id = ?", assigneeID)
	require.NoError(t, err)

	// Verify task's assignee_id is now NULL
	err = db.QueryRowContext(ctx, "SELECT assignee_id FROM tasks WHERE id = ?", taskID).Scan(&gotAssigneeID)
	require.NoError(t, err)
	assert.False(t, gotAssigneeID.Valid)
}
