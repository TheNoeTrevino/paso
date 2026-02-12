package task

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestUpdateTaskEstimate(t *testing.T) {
	t.Parallel()
	db := fixtures.SetupTestDB(t)

	svc, err := NewService(db, database.SQLite, nil, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Helper: create test project, column, and task
	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	columnID := fixtures.GetColumnIDByName(t, db, fixtures.SQLiteDialect(), projectID, "Todo")

	t.Run("set estimate on task without estimate", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task without estimate")

		estimate := "2d"
		err := svc.UpdateTaskEstimate(ctx, taskID, &estimate)
		assert.NoError(t, err)

		var dbEstimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&dbEstimate)
		require.NoError(t, err)
		assert.True(t, dbEstimate.Valid, "estimate should be set")
		assert.Equal(t, "2d", dbEstimate.String)
	})

	t.Run("update existing estimate", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task with estimate")

		// Set initial estimate
		estimate1 := "1d"
		err := svc.UpdateTaskEstimate(ctx, taskID, &estimate1)
		require.NoError(t, err)

		// Verify initial estimate
		var dbEstimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&dbEstimate)
		require.NoError(t, err)
		require.Equal(t, "1d", dbEstimate.String)

		// Update to new estimate
		estimate2 := "3d"
		err = svc.UpdateTaskEstimate(ctx, taskID, &estimate2)
		assert.NoError(t, err)

		// Verify updated estimate
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&dbEstimate)
		require.NoError(t, err)
		assert.Equal(t, "3d", dbEstimate.String)
	})

	t.Run("clear estimate with nil", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task to clear")

		// Set initial estimate
		estimate := "2d"
		err := svc.UpdateTaskEstimate(ctx, taskID, &estimate)
		require.NoError(t, err)

		// Clear estimate with nil
		err = svc.UpdateTaskEstimate(ctx, taskID, nil)
		assert.NoError(t, err)

		// Verify estimate is NULL
		var dbEstimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&dbEstimate)
		require.NoError(t, err)
		assert.False(t, dbEstimate.Valid, "estimate should be NULL")
	})

	t.Run("set empty string estimate", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for empty string")

		// Set initial estimate
		estimate := "2d"
		err := svc.UpdateTaskEstimate(ctx, taskID, &estimate)
		require.NoError(t, err)

		// Set empty string (validation allows it)
		emptyEstimate := ""
		err = svc.UpdateTaskEstimate(ctx, taskID, &emptyEstimate)
		assert.NoError(t, err)

		// Verify estimate is stored as empty string
		var dbEstimate sql.NullString
		err = db.QueryRowContext(ctx,
			"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&dbEstimate)
		require.NoError(t, err)
		assert.True(t, dbEstimate.Valid, "empty string is a valid value")
		assert.Equal(t, "", dbEstimate.String)
	})

	t.Run("set compound estimates", func(t *testing.T) {
		compoundFormats := []string{"1w2d", "2d3h", "1w2d3h", "1w2d3h4m"}

		for _, format := range compoundFormats {
			format := format // capture loop variable
			t.Run(format, func(t *testing.T) {
				taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for "+format)

				err := svc.UpdateTaskEstimate(ctx, taskID, &format)
				assert.NoError(t, err)

				var dbEstimate sql.NullString
				err = db.QueryRowContext(ctx,
					"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&dbEstimate)
				require.NoError(t, err)
				assert.True(t, dbEstimate.Valid)
				assert.Equal(t, format, dbEstimate.String)
			})
		}
	})

	t.Run("set various valid formats", func(t *testing.T) {
		formats := []string{"4h", "30m", "1w", "2d", "3m"}

		for _, format := range formats {
			format := format // capture loop variable
			t.Run(format, func(t *testing.T) {
				taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for "+format)

				err := svc.UpdateTaskEstimate(ctx, taskID, &format)
				assert.NoError(t, err)

				var dbEstimate sql.NullString
				err = db.QueryRowContext(ctx,
					"SELECT estimate FROM tasks WHERE id = ?", taskID).Scan(&dbEstimate)
				require.NoError(t, err)
				assert.True(t, dbEstimate.Valid)
				assert.Equal(t, format, dbEstimate.String)
			})
		}
	})
}

func TestUpdateTaskEstimate_Errors(t *testing.T) {
	t.Parallel()
	db := fixtures.SetupTestDB(t)

	svc, err := NewService(db, database.SQLite, nil, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Helper: create test project, column, and task for valid test cases
	projectID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Test Project")
	columnID := fixtures.GetColumnIDByName(t, db, fixtures.SQLiteDialect(), projectID, "Todo")

	t.Run("invalid task ID - zero", func(t *testing.T) {
		estimate := "2d"
		err := svc.UpdateTaskEstimate(ctx, 0, &estimate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("invalid task ID - negative", func(t *testing.T) {
		estimate := "2d"
		err := svc.UpdateTaskEstimate(ctx, -1, &estimate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("non-existent task", func(t *testing.T) {
		estimate := "2d"
		err := svc.UpdateTaskEstimate(ctx, 999999, &estimate)
		// SQLite UPDATE doesn't error on non-existent rows - it succeeds with 0 rows affected
		// This is expected behavior for SQL UPDATE statements
		assert.NoError(t, err)
	})

	t.Run("invalid estimate format - invalid unit", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for invalid unit")

		invalidEstimate := "2x"
		err := svc.UpdateTaskEstimate(ctx, taskID, &invalidEstimate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("invalid estimate format - no unit", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for no unit")

		invalidEstimate := "5"
		err := svc.UpdateTaskEstimate(ctx, taskID, &invalidEstimate)
		assert.Error(t, err)
	})

	t.Run("invalid estimate format - duplicate unit", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for duplicate")

		invalidEstimate := "1d2d"
		err := svc.UpdateTaskEstimate(ctx, taskID, &invalidEstimate)
		assert.Error(t, err)
	})

	t.Run("invalid estimate format - random text", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for random text")

		invalidEstimate := "invalid"
		err := svc.UpdateTaskEstimate(ctx, taskID, &invalidEstimate)
		assert.Error(t, err)
	})

	t.Run("invalid estimate format - spaces", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for spaces")

		invalidEstimate := "1d 2h"
		err := svc.UpdateTaskEstimate(ctx, taskID, &invalidEstimate)
		assert.Error(t, err)
	})

	t.Run("invalid estimate format - zero quantity", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for zero")

		invalidEstimate := "0d"
		err := svc.UpdateTaskEstimate(ctx, taskID, &invalidEstimate)
		// Zero quantities are now rejected as they are not meaningful for estimates
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidEstimateFormat)
	})

	t.Run("invalid estimate format - negative quantity", func(t *testing.T) {
		taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columnID, "Task for negative")

		invalidEstimate := "-1d"
		err := svc.UpdateTaskEstimate(ctx, taskID, &invalidEstimate)
		assert.Error(t, err)
	})
}
