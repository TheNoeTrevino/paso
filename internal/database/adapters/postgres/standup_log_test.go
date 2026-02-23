package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	postgres_adapter "github.com/thenoetrevino/paso/internal/database/adapters/postgres"
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

// TestStandupLogDateRangeQuery_Postgres verifies that the PostgreSQL adapter correctly
// filters standup logs by a UTC date range.
//
// PostgreSQL uses TIMESTAMP WITHOUT TIME ZONE. The lib/pq driver sends time.Time params
// correctly in wire format and the server handles timezone-aware comparison. However,
// we still validate that UTC-normalized query bounds correctly bracket stored values.
//
// These tests require PostgreSQL to be running (see fixtures.SetupPostgresTestDB).
// They are skipped automatically when PostgreSQL is unavailable.
func TestStandupLogDateRangeQuery_Postgres(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	d := fixtures.PostgresDialect()

	t.Run("UTC range matches logs stored by DEFAULT CURRENT_TIMESTAMP", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-tz-test"))

		insertAt := func(content string, at time.Time) {
			t.Helper()
			_, err := db.ExecContext(ctx,
				"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
				projectID, content, at.UTC())
			require.NoError(t, err)
		}

		// Log: Feb 23, 2026 at 17:13 UTC (exact timestamp from the bug report)
		insertAt("finished this and that", time.Date(2026, 2, 23, 17, 13, 2, 0, time.UTC))
		// Log: Feb 22, 2026 at 10:00 UTC
		insertAt("previous day log", time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC))
		// Log: Feb 15, 2026 at 09:00 UTC — outside the 1-week range
		insertAt("old log", time.Date(2026, 2, 15, 9, 0, 0, 0, time.UTC))

		// 1 week back from Feb 23 → since Feb 16 00:00 UTC
		since := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
		until := time.Date(2026, 2, 24, 0, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		require.Len(t, results, 2, "should return both Feb 22 and Feb 23 logs")

		contents := make([]string, len(results))
		for i, r := range results {
			contents[i] = r.Content
		}
		assert.Contains(t, contents, "finished this and that")
		assert.Contains(t, contents, "previous day log")
		assert.NotContains(t, contents, "old log")
	})

	t.Run("returns empty when range is in the future", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-future"))

		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "a log", time.Date(2026, 2, 23, 17, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
			Until:     time.Date(2027, 1, 8, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("returns empty when range predates all logs", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-past"))

		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "a log", time.Date(2026, 2, 23, 17, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Until:     time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("until boundary is exclusive", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-until-exclusive"))

		logAt := time.Date(2026, 2, 23, 17, 13, 2, 0, time.UTC)
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "test", logAt)
		require.NoError(t, err)

		// until = exactly the log timestamp — should NOT be included (<, not <=)
		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC),
			Until:     logAt,
		})
		require.NoError(t, err)
		assert.Empty(t, results, "log at exact 'until' boundary should be excluded")
	})

	t.Run("since boundary is inclusive", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-since-inclusive"))

		logAt := time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "boundary log", logAt)
		require.NoError(t, err)

		// since = exactly the log timestamp — should be included (>=)
		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     logAt,
			Until:     time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "boundary log", results[0].Content)
	})

	t.Run("only returns logs for the specified project", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-project-filter"))
		otherProjectID := int64(fixtures.CreateTestProject(t, db, d, "pg-other-project"))

		logAt := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "my log", logAt)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			otherProjectID, "other log", logAt)
		require.NoError(t, err)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC),
			Until:     time.Date(2026, 2, 24, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, projectID, results[0].ProjectID)
	})

	t.Run("results are ordered newest first", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-ordering"))

		for _, entry := range []struct {
			content string
			at      time.Time
		}{
			{"oldest", time.Date(2026, 2, 20, 8, 0, 0, 0, time.UTC)},
			{"middle", time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)},
			{"newest", time.Date(2026, 2, 23, 17, 0, 0, 0, time.UTC)},
		} {
			_, err := db.ExecContext(ctx,
				"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
				projectID, entry.content, entry.at)
			require.NoError(t, err)
		}

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC),
			Until:     time.Date(2026, 2, 24, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "newest", results[0].Content)
		assert.Equal(t, "middle", results[1].Content)
		assert.Equal(t, "oldest", results[2].Content)
	})

	t.Run("returned created_at is correct UTC time", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-created-at-utc"))

		logAt := time.Date(2026, 2, 23, 17, 13, 2, 0, time.UTC)
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "test", logAt)
		require.NoError(t, err)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     time.Date(2026, 2, 23, 17, 0, 0, 0, time.UTC),
			Until:     time.Date(2026, 2, 23, 18, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		require.Len(t, results, 1)

		got := results[0].CreatedAt
		require.True(t, got.Valid)
		utc := got.Time.UTC()
		assert.Equal(t, 2026, utc.Year())
		assert.Equal(t, time.February, utc.Month())
		assert.Equal(t, 23, utc.Day())
		assert.Equal(t, 17, utc.Hour())
		assert.Equal(t, 13, utc.Minute())
		assert.Equal(t, 2, utc.Second())
	})
}

// TestStandupLogDateRangeQuery_Postgres_TimezoneEdgeCases tests UTC boundary edge cases
// for PostgreSQL. The key invariant: UTC-normalized since/until must correctly bracket
// values stored via DEFAULT CURRENT_TIMESTAMP (server-local time in Postgres TIMESTAMP).
//
// Note: PostgreSQL TIMESTAMP WITHOUT TIME ZONE stores values in the server's local time,
// but the lib/pq driver automatically handles conversion to/from UTC for time.Time params.
func TestStandupLogDateRangeQuery_Postgres_TimezoneEdgeCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	d := fixtures.PostgresDialect()

	t.Run("UTC midnight boundaries span the full day", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-midnight-span"))

		// User in UTC-6: logs at 5:13 PM local (CST) = 23:13 UTC on Feb 22
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "evening local time", time.Date(2026, 2, 22, 23, 13, 0, 0, time.UTC))
		require.NoError(t, err)
		// User in UTC+9 (JST): logs at 8 AM local Feb 23 = Feb 22 23:00 UTC
		_, err = db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "morning JST same UTC day", time.Date(2026, 2, 22, 23, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		since := time.Date(2026, 2, 22, 0, 0, 0, 0, time.UTC)
		until := time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		assert.Len(t, results, 2, "both logs at 23:00 and 23:13 UTC on Feb 22 should be found")
	})

	t.Run("log at 23:59:59 UTC is within same-day UTC range", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-late-night-utc"))

		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "late night log", time.Date(2026, 2, 22, 23, 59, 59, 0, time.UTC))
		require.NoError(t, err)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     time.Date(2026, 2, 22, 0, 0, 0, 0, time.UTC),
			Until:     time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "late night log", results[0].Content)
	})

	t.Run("log at midnight UTC is NOT in previous day range", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		adapter := postgres_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-midnight-boundary"))

		// Log exactly at Feb 23 00:00:00 UTC — the exclusive 'until' for Feb 22
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES ($1, $2, $3)",
			projectID, "midnight log", time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		// Range is Feb 22 only
		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     time.Date(2026, 2, 22, 0, 0, 0, 0, time.UTC),
			Until:     time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		assert.Empty(t, results, "log at midnight (exclusive 'until') should not appear in Feb 22 range")

		// But it SHOULD appear in the Feb 23 range
		results, err = adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC),
			Until:     time.Date(2026, 2, 24, 0, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "midnight log", results[0].Content)
	})
}

// TestStandupLogCreateAndRetrieve_Postgres verifies the round-trip create + retrieve.
func TestStandupLogCreateAndRetrieve_Postgres(t *testing.T) {
	t.Parallel()

	db := fixtures.SetupPostgresTestDB(t)
	if db == nil {
		return
	}
	adapter := postgres_adapter.New(db)
	ctx := context.Background()

	d := fixtures.PostgresDialect()
	projectID := int64(fixtures.CreateTestProject(t, db, d, "pg-create-retrieve"))

	before := time.Now().UTC().Truncate(time.Second)

	created, err := adapter.CreateStandupLog(ctx, types.CreateStandupLogParams{
		ProjectID: projectID,
		Content:   "test entry",
	})
	require.NoError(t, err)

	after := time.Now().UTC().Add(time.Second)

	assert.Equal(t, projectID, created.ProjectID)
	assert.Equal(t, "test entry", created.Content)
	require.True(t, created.CreatedAt.Valid, "created_at should be valid (non-NULL)")
	assert.False(t, created.CreatedAt.Time.IsZero(), "created_at should not be zero")

	utcTime := created.CreatedAt.Time.UTC()
	assert.True(t, !utcTime.Before(before) && !utcTime.After(after),
		fmt.Sprintf("created_at (%v) should be between %v and %v", utcTime, before, after))
}
