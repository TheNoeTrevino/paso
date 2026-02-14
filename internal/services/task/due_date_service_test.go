package task

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestUpdateTaskDueDate(t *testing.T) {
	t.Parallel()
	db := fixtures.SetupTestDB(t)

	svc, err := NewService(db, database.SQLite, nil, nil)
	require.NoError(t, err)

	ctx := context.Background()

	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	columnID := fixtures.GetColumnIDByName(t, db, fixtures.SQLiteDialect(), projectID, "Todo")

	t.Run("set due date on task without due date", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task without due date")

		dueDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
		err := svc.UpdateTaskDueDate(ctx, taskID, &dueDate)
		assert.NoError(t, err)

		var dbDueDate sql.NullTime
		err = db.QueryRowContext(ctx,
			"SELECT due_date FROM tasks WHERE id = ?", taskID).Scan(&dbDueDate)
		require.NoError(t, err)
		assert.True(t, dbDueDate.Valid, "due date should be set")
		assert.Equal(t, 2026, dbDueDate.Time.Year())
		assert.Equal(t, time.March, dbDueDate.Time.Month())
		assert.Equal(t, 15, dbDueDate.Time.Day())
	})

	t.Run("update existing due date", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task with due date")

		dueDate1 := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
		err := svc.UpdateTaskDueDate(ctx, taskID, &dueDate1)
		require.NoError(t, err)

		var dbDueDate sql.NullTime
		err = db.QueryRowContext(ctx,
			"SELECT due_date FROM tasks WHERE id = ?", taskID).Scan(&dbDueDate)
		require.NoError(t, err)
		require.Equal(t, 15, dbDueDate.Time.Day())

		dueDate2 := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
		err = svc.UpdateTaskDueDate(ctx, taskID, &dueDate2)
		assert.NoError(t, err)

		err = db.QueryRowContext(ctx,
			"SELECT due_date FROM tasks WHERE id = ?", taskID).Scan(&dbDueDate)
		require.NoError(t, err)
		assert.True(t, dbDueDate.Valid)
		assert.Equal(t, 2026, dbDueDate.Time.Year())
		assert.Equal(t, time.June, dbDueDate.Time.Month())
		assert.Equal(t, 20, dbDueDate.Time.Day())
	})

	t.Run("clear due date with nil", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task to clear")

		dueDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
		err := svc.UpdateTaskDueDate(ctx, taskID, &dueDate)
		require.NoError(t, err)

		err = svc.UpdateTaskDueDate(ctx, taskID, nil)
		assert.NoError(t, err)

		var dbDueDate sql.NullTime
		err = db.QueryRowContext(ctx,
			"SELECT due_date FROM tasks WHERE id = ?", taskID).Scan(&dbDueDate)
		require.NoError(t, err)
		assert.False(t, dbDueDate.Valid, "due date should be NULL")
	})

	t.Run("set due date in the past", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task with past due date")

		pastDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		err := svc.UpdateTaskDueDate(ctx, taskID, &pastDate)
		assert.NoError(t, err)

		var dbDueDate sql.NullTime
		err = db.QueryRowContext(ctx,
			"SELECT due_date FROM tasks WHERE id = ?", taskID).Scan(&dbDueDate)
		require.NoError(t, err)
		assert.True(t, dbDueDate.Valid)
		assert.Equal(t, 2020, dbDueDate.Time.Year())
	})

	t.Run("set due date far in the future", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task with far future date")

		futureDate := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
		err := svc.UpdateTaskDueDate(ctx, taskID, &futureDate)
		assert.NoError(t, err)

		var dbDueDate sql.NullTime
		err = db.QueryRowContext(ctx,
			"SELECT due_date FROM tasks WHERE id = ?", taskID).Scan(&dbDueDate)
		require.NoError(t, err)
		assert.True(t, dbDueDate.Valid)
		assert.Equal(t, 2099, dbDueDate.Time.Year())
		assert.Equal(t, time.December, dbDueDate.Time.Month())
		assert.Equal(t, 31, dbDueDate.Time.Day())
	})

	t.Run("set and clear due date multiple times", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for toggling")

		dueDate := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
		err := svc.UpdateTaskDueDate(ctx, taskID, &dueDate)
		require.NoError(t, err)

		err = svc.UpdateTaskDueDate(ctx, taskID, nil)
		require.NoError(t, err)

		newDueDate := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
		err = svc.UpdateTaskDueDate(ctx, taskID, &newDueDate)
		assert.NoError(t, err)

		var dbDueDate sql.NullTime
		err = db.QueryRowContext(ctx,
			"SELECT due_date FROM tasks WHERE id = ?", taskID).Scan(&dbDueDate)
		require.NoError(t, err)
		assert.True(t, dbDueDate.Valid)
		assert.Equal(t, 2026, dbDueDate.Time.Year())
		assert.Equal(t, time.August, dbDueDate.Time.Month())
		assert.Equal(t, 20, dbDueDate.Time.Day())
	})
}

func TestUpdateTaskDueDate_Errors(t *testing.T) {
	t.Parallel()
	db := fixtures.SetupTestDB(t)

	svc, err := NewService(db, database.SQLite, nil, nil)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("invalid task ID - zero", func(t *testing.T) {
		dueDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
		err := svc.UpdateTaskDueDate(ctx, 0, &dueDate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("invalid task ID - negative", func(t *testing.T) {
		dueDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
		err := svc.UpdateTaskDueDate(ctx, -1, &dueDate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("non-existent task", func(t *testing.T) {
		dueDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
		err := svc.UpdateTaskDueDate(ctx, 999999, &dueDate)
		// SQLite UPDATE doesn't error on non-existent rows - it succeeds with 0 rows affected
		assert.NoError(t, err)
	})

	t.Run("clear due date on non-existent task", func(t *testing.T) {
		err := svc.UpdateTaskDueDate(ctx, 999999, nil)
		// Same behavior as above - UPDATE with 0 rows
		assert.NoError(t, err)
	})
}
