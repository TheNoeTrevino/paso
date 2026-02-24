package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/events"

	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database with production migrations applied.
// This is a local helper to avoid importing testutil (which would create an import cycle).
func setupTestDB(tb testing.TB) *sql.DB {
	tb.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		tb.Fatalf("Failed to create test database: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON"); err != nil {
		tb.Fatalf("Failed to enable foreign keys: %v", err)
	}
	if err := applyMigrations(db, SQLite, false); err != nil {
		tb.Fatalf("Failed to run migrations: %v", err)
	}
	tb.Cleanup(func() {
		if err := db.Close(); err != nil {
			tb.Logf("failed to close test db: %v", err)
		}
	})
	return db
}

// setupPostgresTestDB connects to a PostgreSQL test database and applies production migrations.
// Returns nil if PostgreSQL is not available (the caller should skip).
func setupPostgresTestDB(tb testing.TB) *sql.DB {
	tb.Helper()

	host := getTestEnv("PG_HOST", "localhost")
	port := getTestEnv("PG_PORT", "5432")
	user := getTestEnv("PG_USER", "postgres")
	password := getTestEnv("PG_PASSWORD", "postgres")
	dbname := getTestEnv("PG_DATABASE", "paso_test")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		tb.Skipf("PostgreSQL not available: %v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		tb.Skipf("PostgreSQL connection failed: %v", err)
		return nil
	}

	// Drop all tables and goose state for a clean slate
	drops := `
	drop table if exists standup_logs cascade;
	drop table if exists task_events cascade;
	drop table if exists task_comments cascade;
	drop table if exists task_labels cascade;
	drop table if exists task_subtasks cascade;
	drop table if exists tasks cascade;
	drop table if exists assignees cascade;
	drop table if exists labels cascade;
	drop table if exists columns cascade;
	drop table if exists project_counters cascade;
	drop table if exists projects cascade;
	drop table if exists relation_types cascade;
	drop table if exists priorities cascade;
	drop table if exists types cascade;
	drop table if exists goose_db_version cascade;
	`
	if _, err := db.ExecContext(ctx, drops); err != nil {
		tb.Logf("warning: failed to drop tables during cleanup: %v", err)
	}

	if err := applyMigrations(db, PostgreSQL, false); err != nil {
		tb.Fatalf("Failed to run PostgreSQL migrations: %v", err)
	}

	tb.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func getTestEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func createTestProject(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	ctx := context.Background()
	result, err := db.ExecContext(ctx, "INSERT INTO projects (name) VALUES (?)", name)
	require.NoError(t, err)
	projectID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO project_counters (project_id) VALUES (?)", projectID)
	require.NoError(t, err)

	return int(projectID)
}

func TestWithTx_Success_Commit(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)

	ctx := context.Background()
	projectID := createTestProject(t, db, "Test Project")

	// Execute transaction that should commit
	err := WithTx(ctx, db, func(tx *sql.Tx) error {
		// Insert a column within transaction
		_, err := tx.ExecContext(ctx, "INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, "Test Column")
		return err
	})
	require.NoError(t, err)

	// Verify column was created (transaction committed)
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM columns WHERE name = ?", "Test Column").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestWithTx_Error_Rollback(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)

	ctx := context.Background()
	projectID := createTestProject(t, db, "Test Project")

	// Execute transaction that should rollback
	expectedErr := errors.New("intentional error")
	err := WithTx(ctx, db, func(tx *sql.Tx) error {
		// Insert a column within transaction
		_, err := tx.ExecContext(ctx, "INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, "Test Column")
		if err != nil {
			return err
		}
		// Return error to trigger rollback
		return expectedErr
	})

	require.Equal(t, expectedErr, err)

	// Verify column was NOT created (transaction rolled back)
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM columns WHERE name = ?", "Test Column").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestWithTx_Error_BeginFails(t *testing.T) {
	t.Parallel()
	// Create a closed database to trigger begin error
	db := setupTestDB(t)
	_ = db.Close()

	ctx := context.Background()
	err := WithTx(ctx, db, func(tx *sql.Tx) error {
		return nil
	})

	require.Error(t, err)
}

func TestNullInt64ToPtr_Valid(t *testing.T) {
	t.Parallel()
	nv := sql.NullInt64{Int64: 42, Valid: true}
	result := nullInt64ToPtr(nv)

	require.NotNil(t, result)
	assert.Equal(t, 42, *result)
}

func TestNullInt64ToPtr_Null(t *testing.T) {
	t.Parallel()
	nv := sql.NullInt64{Int64: 0, Valid: false}
	result := nullInt64ToPtr(nv)

	assert.Nil(t, result)
}

func TestNullStringToString_Valid(t *testing.T) {
	t.Parallel()
	ns := sql.NullString{String: "test string", Valid: true}
	result := NullStringToString(ns)

	assert.Equal(t, "test string", result)
}

func TestNullStringToString_Null(t *testing.T) {
	t.Parallel()
	ns := sql.NullString{String: "", Valid: false}
	result := NullStringToString(ns)

	assert.Equal(t, "", result)
}

func TestNullTimeToTime_Valid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	nt := sql.NullTime{Time: now, Valid: true}
	result := NullTimeToTime(nt)

	assert.True(t, result.Equal(now))
}

func TestNullTimeToTime_Null(t *testing.T) {
	t.Parallel()
	nt := sql.NullTime{Time: time.Time{}, Valid: false}
	result := NullTimeToTime(nt)

	assert.True(t, result.IsZero())
}

func TestAnyToIntPtr_Int64(t *testing.T) {
	t.Parallel()
	var val any = int64(123)
	result := AnyToIntPtr(val)

	require.NotNil(t, result)
	assert.Equal(t, 123, *result)
}

func TestAnyToIntPtr_Int(t *testing.T) {
	t.Parallel()
	var val any = int(456)
	result := AnyToIntPtr(val)

	require.NotNil(t, result)
	assert.Equal(t, 456, *result)
}

func TestAnyToIntPtr_Nil(t *testing.T) {
	t.Parallel()
	var val any = nil
	result := AnyToIntPtr(val)

	assert.Nil(t, result)
}

func TestAnyToIntPtr_InvalidType(t *testing.T) {
	t.Parallel()
	var val any = "not an int"
	result := AnyToIntPtr(val)

	assert.Nil(t, result)
}

type mockEventPublisher struct {
	sentEvents []events.Event
	shouldFail bool
}

func (m *mockEventPublisher) Connect(ctx context.Context) error { return nil }
func (m *mockEventPublisher) Listen(ctx context.Context) (<-chan events.Event, error) {
	return nil, nil
}
func (m *mockEventPublisher) Subscribe(projectID int) error        { return nil }
func (m *mockEventPublisher) SetNotifyFunc(fn events.NotifyFunc)   {}
func (m *mockEventPublisher) ForceFlush(ctx context.Context) error { return nil }
func (m *mockEventPublisher) Close() error                         { return nil }

func (m *mockEventPublisher) SendEvent(event events.Event) error {
	if m.shouldFail {
		return errors.New("mock send error")
	}
	m.sentEvents = append(m.sentEvents, event)
	return nil
}

func TestSendEvent_WithClient(t *testing.T) {
	t.Parallel()
	mock := &mockEventPublisher{sentEvents: []events.Event{}}
	projectID := 42

	sendEvent(mock, projectID)

	require.Len(t, mock.sentEvents, 1)

	event := mock.sentEvents[0]
	assert.Equal(t, events.EventDatabaseChanged, event.Type)
	assert.Equal(t, projectID, event.ProjectID)
}

func TestSendEvent_NilClient(t *testing.T) {
	t.Parallel()
	// Should not panic with nil client
	sendEvent(nil, 42)
}

func TestSendEvent_Error(t *testing.T) {
	t.Parallel()
	mock := &mockEventPublisher{shouldFail: true}

	// Should not panic or return error (errors are logged)
	sendEvent(mock, 42)
}
