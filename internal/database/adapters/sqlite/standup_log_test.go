package sqlite_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sqlite_adapter "github.com/thenoetrevino/paso/internal/database/adapters/sqlite"
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

// TestStandupLogDateRangeQuery_SQLite verifies that the SQLite adapter correctly filters
// standup logs by a UTC date range. This is the core regression test for the timezone
// mismatch bug where CURRENT_TIMESTAMP stores UTC bare strings but query params were
// constructed in local time.
func TestStandupLogDateRangeQuery_SQLite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d := fixtures.SQLiteDialect()

	t.Run("UTC range matches UTC-stored logs", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "tz-test"))

		insertAt := func(content string, at time.Time) {
			t.Helper()
			ts := at.UTC().Format("2006-01-02 15:04:05")
			_, err := db.ExecContext(ctx,
				"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
				projectID, content, ts)
			require.NoError(t, err)
		}

		// Log: Feb 23, 2026 at 17:13 UTC (the exact timestamp from the bug report)
		insertAt("finished this and that", time.Date(2026, 2, 23, 17, 13, 2, 0, time.UTC))
		// Log: Feb 22, 2026 at 10:00 UTC
		insertAt("previous day log", time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC))
		// Log: Feb 15, 2026 at 09:00 UTC — outside the 1-week range
		insertAt("old log", time.Date(2026, 2, 15, 9, 0, 0, 0, time.UTC))

		// Simulate what generate.go does after the fix: time.Now().UTC() as reference.
		// 1 week back from Feb 23 = since Feb 16 00:00 UTC.
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
		assert.NotContains(t, contents, "old log", "Feb 15 log should be outside the range")
	})

	t.Run("returns empty when range is in the future", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "future-range"))

		ts := time.Date(2026, 2, 23, 17, 13, 2, 0, time.UTC).Format("2006-01-02 15:04:05")
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
			projectID, "a log", ts)
		require.NoError(t, err)

		since := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
		until := time.Date(2027, 1, 8, 0, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("returns empty when range predates all logs", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "past-range"))

		ts := time.Date(2026, 2, 23, 17, 13, 2, 0, time.UTC).Format("2006-01-02 15:04:05")
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
			projectID, "a log", ts)
		require.NoError(t, err)

		since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		until := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("until boundary is exclusive", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "until-exclusive"))

		logAt := time.Date(2026, 2, 23, 17, 13, 2, 0, time.UTC)
		ts := logAt.Format("2006-01-02 15:04:05")
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
			projectID, "test", ts)
		require.NoError(t, err)

		// until = exactly the log timestamp — should NOT be included (<, not <=)
		since := time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC)
		until := logAt

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		assert.Empty(t, results, "log at exact 'until' boundary should be excluded")
	})

	t.Run("since boundary is inclusive", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "since-inclusive"))

		logAt := time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC)
		ts := logAt.Format("2006-01-02 15:04:05")
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
			projectID, "boundary log", ts)
		require.NoError(t, err)

		// since = exactly the log timestamp — should be included (>=)
		since := logAt
		until := time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		require.Len(t, results, 1, "log at exact 'since' boundary should be included")
		assert.Equal(t, "boundary log", results[0].Content)
	})

	t.Run("only returns logs for the specified project", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "project-filter"))
		otherProjectID := int64(fixtures.CreateTestProject(t, db, d, "other-project"))

		ts := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
			projectID, "my log", ts)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
			otherProjectID, "other log", ts)
		require.NoError(t, err)

		since := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
		until := time.Date(2026, 2, 24, 0, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, projectID, results[0].ProjectID)
	})

	t.Run("results are ordered newest first", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "ordering"))

		for _, entry := range []struct {
			content string
			at      time.Time
		}{
			{"oldest", time.Date(2026, 2, 20, 8, 0, 0, 0, time.UTC)},
			{"middle", time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)},
			{"newest", time.Date(2026, 2, 23, 17, 0, 0, 0, time.UTC)},
		} {
			ts := entry.at.Format("2006-01-02 15:04:05")
			_, err := db.ExecContext(ctx,
				"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
				projectID, entry.content, ts)
			require.NoError(t, err)
		}

		since := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
		until := time.Date(2026, 2, 24, 0, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "newest", results[0].Content)
		assert.Equal(t, "middle", results[1].Content)
		assert.Equal(t, "oldest", results[2].Content)
	})

	t.Run("returned created_at is parseable as UTC", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "created-at-utc"))

		logAt := time.Date(2026, 2, 23, 17, 13, 2, 0, time.UTC)
		ts := logAt.Format("2006-01-02 15:04:05")
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
			projectID, "test", ts)
		require.NoError(t, err)

		since := time.Date(2026, 2, 23, 17, 0, 0, 0, time.UTC)
		until := time.Date(2026, 2, 23, 18, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		require.Len(t, results, 1)

		got := results[0].CreatedAt
		require.True(t, got.Valid, "created_at should be non-NULL")
		utc := got.Time.UTC()
		assert.Equal(t, 2026, utc.Year())
		assert.Equal(t, time.February, utc.Month())
		assert.Equal(t, 23, utc.Day())
		assert.Equal(t, 17, utc.Hour())
		assert.Equal(t, 13, utc.Minute())
		assert.Equal(t, 2, utc.Second())
	})
}

// TestStandupLogDateRangeQuery_SQLite_TimezoneEdgeCases tests edge cases around UTC
// timezone boundaries — the root cause of the original bug.
func TestStandupLogDateRangeQuery_SQLite_TimezoneEdgeCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	d := fixtures.SQLiteDialect()

	t.Run("UTC midnight boundaries span the full day", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "midnight-span"))

		insertAt := func(content string, at time.Time) {
			t.Helper()
			ts := at.UTC().Format("2006-01-02 15:04:05")
			_, err := db.ExecContext(ctx,
				"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
				projectID, content, ts)
			require.NoError(t, err)
		}

		// User in UTC-6: logs at 5:13 PM local (CST) = 23:13 UTC on Feb 22.
		insertAt("evening local time", time.Date(2026, 2, 22, 23, 13, 0, 0, time.UTC))
		// User in UTC+9 (JST): logs at 8 AM local Feb 23 = Feb 22 23:00 UTC.
		insertAt("morning JST same UTC day", time.Date(2026, 2, 22, 23, 0, 0, 0, time.UTC))

		// UTC-based range: midnight-to-midnight on Feb 22 UTC
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
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "late-night-utc"))

		ts := time.Date(2026, 2, 22, 23, 59, 59, 0, time.UTC).Format("2006-01-02 15:04:05")
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
			projectID, "late night log", ts)
		require.NoError(t, err)

		since := time.Date(2026, 2, 22, 0, 0, 0, 0, time.UTC)
		until := time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
		})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "late night log", results[0].Content)
	})

	t.Run("log at midnight UTC is NOT in previous day range", func(t *testing.T) {
		t.Parallel()
		db := fixtures.SetupTestDB(t)
		adapter := sqlite_adapter.New(db)
		projectID := int64(fixtures.CreateTestProject(t, db, d, "midnight-boundary"))

		// Log exactly at Feb 23 00:00:00 UTC — the exclusive 'until' boundary for Feb 22
		ts := time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")
		_, err := db.ExecContext(ctx,
			"INSERT INTO standup_logs (project_id, content, created_at) VALUES (?, ?, ?)",
			projectID, "midnight log", ts)
		require.NoError(t, err)

		// Range is Feb 22 only
		since := time.Date(2026, 2, 22, 0, 0, 0, 0, time.UTC)
		until := time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC)

		results, err := adapter.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
			ProjectID: projectID,
			Since:     since,
			Until:     until,
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

// TestStandupLogCreateAndRetrieve_SQLite verifies the round-trip: create via DEFAULT
// CURRENT_TIMESTAMP (UTC in SQLite) and retrieve. The created_at should be valid and recent.
func TestStandupLogCreateAndRetrieve_SQLite(t *testing.T) {
	t.Parallel()

	db := fixtures.SetupTestDB(t)
	adapter := sqlite_adapter.New(db)
	ctx := context.Background()

	d := fixtures.SQLiteDialect()
	projectID := int64(fixtures.CreateTestProject(t, db, d, "create-retrieve"))

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
