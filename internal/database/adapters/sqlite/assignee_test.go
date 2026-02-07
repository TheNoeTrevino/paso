package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/thenoetrevino/paso/internal/testutil"
)

func TestDeleteAssigneeSetNullBehavior(t *testing.T) {
	db := testutil.SetupTestDB(t)

	ctx := context.Background()

	// Create a project and column for the task
	projectID := testutil.CreateTestProject(t, db, "test-project")
	// CreateTestProject creates 3 columns, get the first one
	var columnID int
	err := db.QueryRowContext(ctx, "SELECT id FROM columns WHERE project_id = ? LIMIT 1", projectID).Scan(&columnID)
	if err != nil {
		t.Fatalf("failed to get column: %v", err)
	}

	// Create an assignee
	result, err := db.ExecContext(ctx, "INSERT INTO assignees (name) VALUES (?)", "test-user")
	if err != nil {
		t.Fatalf("failed to create assignee: %v", err)
	}
	assigneeID, _ := result.LastInsertId()

	// Create a task assigned to this assignee
	taskID := testutil.CreateTestTask(t, db, columnID, "test-task")
	_, err = db.ExecContext(ctx, "UPDATE tasks SET assignee_id = ? WHERE id = ?", assigneeID, taskID)
	if err != nil {
		t.Fatalf("failed to assign task: %v", err)
	}

	// Verify assignment
	var gotAssigneeID sql.NullInt64
	err = db.QueryRowContext(ctx, "SELECT assignee_id FROM tasks WHERE id = ?", taskID).Scan(&gotAssigneeID)
	if err != nil {
		t.Fatalf("failed to query task: %v", err)
	}
	if !gotAssigneeID.Valid || gotAssigneeID.Int64 != assigneeID {
		t.Fatalf("task assignee_id = %v, want %d", gotAssigneeID, assigneeID)
	}

	// Delete the assignee
	_, err = db.ExecContext(ctx, "DELETE FROM assignees WHERE id = ?", assigneeID)
	if err != nil {
		t.Fatalf("failed to delete assignee: %v", err)
	}

	// Verify task's assignee_id is now NULL
	err = db.QueryRowContext(ctx, "SELECT assignee_id FROM tasks WHERE id = ?", taskID).Scan(&gotAssigneeID)
	if err != nil {
		t.Fatalf("failed to query task after delete: %v", err)
	}
	if gotAssigneeID.Valid {
		t.Errorf("expected NULL assignee_id after delete, got %d", gotAssigneeID.Int64)
	}
}
