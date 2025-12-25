package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/thenoetrevino/paso/internal/events"

	_ "modernc.org/sqlite"
)

// ============================================================================
// Local Test Helpers (to avoid import cycle with testutil)
// ============================================================================

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	_, err = db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err := createTestSchema(db); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func createTestSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS project_counters (
		project_id INTEGER PRIMARY KEY,
		next_ticket_number INTEGER DEFAULT 1,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS columns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		prev_id INTEGER,
		next_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);
	`
	_, err := db.Exec(schema)
	return err
}

func createTestProject(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO projects (name) VALUES (?)", name)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}
	projectID, _ := result.LastInsertId()

	_, err = db.Exec("INSERT INTO project_counters (project_id) VALUES (?)", projectID)
	if err != nil {
		t.Fatalf("Failed to create project counter: %v", err)
	}

	return int(projectID)
}

func createTestColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, name)
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}
	columnID, _ := result.LastInsertId()
	return int(columnID)
}

// ============================================================================
// Transaction Helper Tests
// ============================================================================

func TestWithTx_Success_Commit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	projectID := createTestProject(t, db, "Test Project")

	// Execute transaction that should commit
	err := withTx(ctx, db, func(tx *sql.Tx) error {
		// Insert a column within transaction
		_, err := tx.Exec("INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, "Test Column")
		return err
	})

	if err != nil {
		t.Fatalf("Expected transaction to succeed, got error: %v", err)
	}

	// Verify column was created (transaction committed)
	var count int
	db.QueryRow("SELECT COUNT(*) FROM columns WHERE name = ?", "Test Column").Scan(&count)
	if count != 1 {
		t.Errorf("Expected 1 column, got %d", count)
	}
}

func TestWithTx_Error_Rollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	projectID := createTestProject(t, db, "Test Project")

	// Execute transaction that should rollback
	expectedErr := errors.New("intentional error")
	err := withTx(ctx, db, func(tx *sql.Tx) error {
		// Insert a column within transaction
		_, err := tx.Exec("INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, "Test Column")
		if err != nil {
			return err
		}
		// Return error to trigger rollback
		return expectedErr
	})

	if err != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify column was NOT created (transaction rolled back)
	var count int
	db.QueryRow("SELECT COUNT(*) FROM columns WHERE name = ?", "Test Column").Scan(&count)
	if count != 0 {
		t.Errorf("Expected 0 columns (rollback), got %d", count)
	}
}

func TestWithTx_Error_BeginFails(t *testing.T) {
	// Create a closed database to trigger begin error
	db := setupTestDB(t)
	db.Close()

	ctx := context.Background()
	err := withTx(ctx, db, func(tx *sql.Tx) error {
		return nil
	})

	if err == nil {
		t.Fatal("Expected error when beginning transaction on closed DB, got nil")
	}
}

// ============================================================================
// Null Conversion Tests
// ============================================================================

func TestNullInt64ToPtr_Valid(t *testing.T) {
	nv := sql.NullInt64{Int64: 42, Valid: true}
	result := nullInt64ToPtr(nv)

	if result == nil {
		t.Fatal("Expected non-nil pointer, got nil")
	}
	if *result != 42 {
		t.Errorf("Expected 42, got %d", *result)
	}
}

func TestNullInt64ToPtr_Null(t *testing.T) {
	nv := sql.NullInt64{Int64: 0, Valid: false}
	result := nullInt64ToPtr(nv)

	if result != nil {
		t.Errorf("Expected nil for SQL NULL, got %v", result)
	}
}

func TestNullStringToString_Valid(t *testing.T) {
	ns := sql.NullString{String: "test string", Valid: true}
	result := NullStringToString(ns)

	if result != "test string" {
		t.Errorf("Expected 'test string', got '%s'", result)
	}
}

func TestNullStringToString_Null(t *testing.T) {
	ns := sql.NullString{String: "", Valid: false}
	result := NullStringToString(ns)

	if result != "" {
		t.Errorf("Expected empty string for SQL NULL, got '%s'", result)
	}
}

func TestNullTimeToTime_Valid(t *testing.T) {
	now := time.Now()
	nt := sql.NullTime{Time: now, Valid: true}
	result := NullTimeToTime(nt)

	if !result.Equal(now) {
		t.Errorf("Expected %v, got %v", now, result)
	}
}

func TestNullTimeToTime_Null(t *testing.T) {
	nt := sql.NullTime{Time: time.Time{}, Valid: false}
	result := NullTimeToTime(nt)

	if !result.IsZero() {
		t.Errorf("Expected zero time for SQL NULL, got %v", result)
	}
}

func TestInterfaceToIntPtr_Int64(t *testing.T) {
	var val interface{} = int64(123)
	result := InterfaceToIntPtr(val)

	if result == nil {
		t.Fatal("Expected non-nil pointer, got nil")
	}
	if *result != 123 {
		t.Errorf("Expected 123, got %d", *result)
	}
}

func TestInterfaceToIntPtr_Int(t *testing.T) {
	var val interface{} = int(456)
	result := InterfaceToIntPtr(val)

	if result == nil {
		t.Fatal("Expected non-nil pointer, got nil")
	}
	if *result != 456 {
		t.Errorf("Expected 456, got %d", *result)
	}
}

func TestInterfaceToIntPtr_Nil(t *testing.T) {
	var val interface{} = nil
	result := InterfaceToIntPtr(val)

	if result != nil {
		t.Errorf("Expected nil for nil interface, got %v", result)
	}
}

func TestInterfaceToIntPtr_InvalidType(t *testing.T) {
	var val interface{} = "not an int"
	result := InterfaceToIntPtr(val)

	if result != nil {
		t.Errorf("Expected nil for invalid type, got %v", result)
	}
}

// ============================================================================
// Event Sending Tests
// ============================================================================

type mockEventPublisher struct {
	sentEvents []events.Event
	shouldFail bool
}

func (m *mockEventPublisher) Connect(ctx context.Context) error { return nil }
func (m *mockEventPublisher) Listen(ctx context.Context) (<-chan events.Event, error) {
	return nil, nil
}
func (m *mockEventPublisher) Subscribe(projectID int) error      { return nil }
func (m *mockEventPublisher) SetNotifyFunc(fn events.NotifyFunc) {}
func (m *mockEventPublisher) Close() error                       { return nil }

func (m *mockEventPublisher) SendEvent(event events.Event) error {
	if m.shouldFail {
		return errors.New("mock send error")
	}
	m.sentEvents = append(m.sentEvents, event)
	return nil
}

func TestSendEvent_WithClient(t *testing.T) {
	mock := &mockEventPublisher{sentEvents: []events.Event{}}
	projectID := 42

	sendEvent(mock, projectID)

	if len(mock.sentEvents) != 1 {
		t.Fatalf("Expected 1 event to be sent, got %d", len(mock.sentEvents))
	}

	event := mock.sentEvents[0]
	if event.Type != events.EventDatabaseChanged {
		t.Errorf("Expected event type %s, got %s", events.EventDatabaseChanged, event.Type)
	}
	if event.ProjectID != projectID {
		t.Errorf("Expected project ID %d, got %d", projectID, event.ProjectID)
	}
}

func TestSendEvent_NilClient(t *testing.T) {
	// Should not panic with nil client
	sendEvent(nil, 42)
}

func TestSendEvent_Error(t *testing.T) {
	mock := &mockEventPublisher{shouldFail: true}

	// Should not panic or return error (errors are logged)
	sendEvent(mock, 42)
}

// ============================================================================
// Project ID Lookup Tests
// ============================================================================

func TestGetProjectIDFromTable_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	projectID := createTestProject(t, db, "Test Project")
	columnID := createTestColumn(t, db, projectID, "Test Column")

	// Get project ID from columns table
	result, err := getProjectIDFromTable(ctx, db, "columns", columnID)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if result != projectID {
		t.Errorf("Expected project ID %d, got %d", projectID, result)
	}
}

func TestGetProjectIDFromTable_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Try to get project ID for non-existent column
	_, err := getProjectIDFromTable(ctx, db, "columns", 9999)
	if err == nil {
		t.Fatal("Expected error for non-existent entity, got nil")
	}
}

func TestGetProjectIDFromTable_InvalidTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Try to get project ID from non-existent table
	_, err := getProjectIDFromTable(ctx, db, "nonexistent_table", 1)
	if err == nil {
		t.Fatal("Expected error for invalid table, got nil")
	}
}
