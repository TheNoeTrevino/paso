package taskevent

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testutil"
)

func newTestService(t *testing.T, db *sql.DB) Service {
	t.Helper()
	svc, err := NewService(db, database.SQLite)
	require.NoError(t, err, "failed to create test service")
	return svc
}

func newTestQuerier(t *testing.T, db *sql.DB) database.Querier {
	t.Helper()
	q, err := database.NewQuerier(db, database.SQLite)
	require.NoError(t, err, "failed to create test querier")
	return q
}

func createTestProject(t *testing.T, db *sql.DB) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO projects (name, description) VALUES (?, ?)", "Test Project", "Description")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	id, _ := result.LastInsertId()
	_, err = db.ExecContext(context.Background(), "INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)", id)
	if err != nil {
		t.Fatalf("Failed to initialize project counter: %v", err)
	}

	return int(id)
}

func createTestColumn(t *testing.T, db *sql.DB, projectID int, name string) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name, holds_ready_tasks) VALUES (?, ?, ?)", projectID, name, false)
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

func createTestTask(t *testing.T, db *sql.DB, columnID int, title string) int {
	t.Helper()

	var maxPos sql.NullInt64
	err := db.QueryRowContext(context.Background(),
		"SELECT MAX(position) FROM tasks WHERE column_id = ?", columnID).Scan(&maxPos)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Failed to get max position: %v", err)
	}

	nextPos := 0
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	result, err := db.ExecContext(context.Background(),
		"INSERT INTO tasks (column_id, title, position, type_id, priority_id) VALUES (?, ?, ?, 1, 3)",
		columnID, title, nextPos)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

func createTestTaskWithProjectSetup(t *testing.T, db *sql.DB) int {
	t.Helper()
	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")
	return createTestTask(t, db, columnID, "Test Task")
}

func TestCreateTaskCreatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateTaskCreatedEvent(ctx, queries, taskID, "My New Task", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "Task created: My New Task", events[0].Content)
	require.Equal(t, "testuser", events[0].Author)
}

func TestCreateTaskMovedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateTaskMovedEvent(ctx, queries, taskID, "Todo", "In Progress", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "Moved from 'Todo' to 'In Progress'", events[0].Content)
}

func TestCreateTaskAssociatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateTaskAssociatedEvent(ctx, queries, taskID, 42, "Other Task", "Blocked By", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "Associated to task #42 (Other Task) as: Blocked By", events[0].Content)
}

func TestCreateTaskDisassociatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateTaskDisassociatedEvent(ctx, queries, taskID, 42, "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "Removed association to task #42", events[0].Content)
}

func TestCreateLabelAddedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateLabelAddedEvent(ctx, queries, taskID, "urgent", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "Added label: urgent", events[0].Content)
}

func TestCreateLabelRemovedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateLabelRemovedEvent(ctx, queries, taskID, "urgent", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "Removed label: urgent", events[0].Content)
}

func TestCreatePriorityChangedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreatePriorityChangedEvent(ctx, queries, taskID, "medium", "high", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "Changed priority from 'medium' to 'high'", events[0].Content)
}

func TestCreateTypeChangedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateTypeChangedEvent(ctx, queries, taskID, "task", "bug", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "Changed type from 'task' to 'bug'", events[0].Content)
}

func TestGetEventsByTask_MultipleEvents(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	_ = svc.CreateTaskCreatedEvent(ctx, queries, taskID, "Task", "user1")
	_ = svc.CreateLabelAddedEvent(ctx, queries, taskID, "urgent", "user2")
	_ = svc.CreateTaskMovedEvent(ctx, queries, taskID, "Todo", "Done", "user3")

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 3)

	// Verify all events are present (don't check order since timestamps may be identical)
	contents := make([]string, 3)
	for i, e := range events {
		contents[i] = e.Content
	}

	foundMoved := false
	foundLabel := false
	foundCreated := false
	for _, c := range contents {
		if c == "Moved from 'Todo' to 'Done'" {
			foundMoved = true
		}
		if c == "Added label: urgent" {
			foundLabel = true
		}
		if c == "Task created: Task" {
			foundCreated = true
		}
	}
	require.True(t, foundMoved, "Should contain 'Moved' event")
	require.True(t, foundLabel, "Should contain 'Added label' event")
	require.True(t, foundCreated, "Should contain 'Task created' event")
}

func TestGetEventsByTask_EmptyList(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)

	events, err := svc.GetEventsByTask(context.Background(), taskID)
	require.NoError(t, err)
	require.Len(t, events, 0)
}

func TestGetEventsByTask_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	_, err := svc.GetEventsByTask(ctx, 0)
	require.ErrorIs(t, err, ErrInvalidTaskID)

	_, err = svc.GetEventsByTask(ctx, -1)
	require.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestCreateEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateTaskCreatedEvent(ctx, queries, 0, "Task", "user")
	require.ErrorIs(t, err, ErrInvalidTaskID)

	err = svc.CreateTaskCreatedEvent(ctx, queries, -1, "Task", "user")
	require.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestCreateTaskMovedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	queries := newTestQuerier(t, db)

	err := svc.CreateTaskMovedEvent(context.Background(), queries, 0, "Todo", "Done", "user")
	require.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestCreateTaskAssociatedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	queries := newTestQuerier(t, db)

	err := svc.CreateTaskAssociatedEvent(context.Background(), queries, 0, 42, "Other", "Blocked By", "user")
	require.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestCreateTaskDisassociatedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	queries := newTestQuerier(t, db)

	err := svc.CreateTaskDisassociatedEvent(context.Background(), queries, 0, 42, "user")
	require.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestCreateLabelAddedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	queries := newTestQuerier(t, db)

	err := svc.CreateLabelAddedEvent(context.Background(), queries, 0, "urgent", "user")
	require.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestCreateLabelRemovedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	queries := newTestQuerier(t, db)

	err := svc.CreateLabelRemovedEvent(context.Background(), queries, 0, "urgent", "user")
	require.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestCreatePriorityChangedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	queries := newTestQuerier(t, db)

	err := svc.CreatePriorityChangedEvent(context.Background(), queries, 0, "medium", "high", "user")
	require.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestCreateTypeChangedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	queries := newTestQuerier(t, db)

	err := svc.CreateTypeChangedEvent(context.Background(), queries, 0, "task", "bug", "user")
	require.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestEventAuthorPreserved(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	// Create events with different authors
	_ = svc.CreateTaskCreatedEvent(ctx, queries, taskID, "Task", "alice")
	_ = svc.CreateLabelAddedEvent(ctx, queries, taskID, "urgent", "bob")
	_ = svc.CreateTaskMovedEvent(ctx, queries, taskID, "Todo", "Done", "charlie")

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 3)

	// Verify authors are preserved (check all authors present, order may vary with same timestamps)
	authorMap := make(map[string]bool)
	for _, e := range events {
		authorMap[e.Author] = true
	}
	require.True(t, authorMap["alice"], "alice should be present")
	require.True(t, authorMap["bob"], "bob should be present")
	require.True(t, authorMap["charlie"], "charlie should be present")
}

func TestEventTimestampSet(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateTaskCreatedEvent(ctx, queries, taskID, "Task", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)

	// CreatedAt should be set
	require.False(t, events[0].CreatedAt.IsZero(), "CreatedAt should be set")
}

func TestEventTaskIDPreserved(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateTaskCreatedEvent(ctx, queries, taskID, "Task", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, taskID, events[0].TaskID)
}

func TestEventIDAssigned(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	taskID := createTestTaskWithProjectSetup(t, db)
	queries := newTestQuerier(t, db)
	ctx := context.Background()

	err := svc.CreateTaskCreatedEvent(ctx, queries, taskID, "Task", "testuser")
	require.NoError(t, err)

	events, err := svc.GetEventsByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Greater(t, events[0].ID, 0, "Event ID should be assigned")
}

func TestGetEventsByTask_OnlyReturnsEventsForSpecificTask(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two tasks
	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")
	taskID1 := createTestTask(t, db, columnID, "Task 1")
	taskID2 := createTestTask(t, db, columnID, "Task 2")

	queries := newTestQuerier(t, db)

	// Create events for both tasks
	_ = svc.CreateTaskCreatedEvent(ctx, queries, taskID1, "Task 1", "user1")
	_ = svc.CreateTaskCreatedEvent(ctx, queries, taskID2, "Task 2", "user2")
	_ = svc.CreateLabelAddedEvent(ctx, queries, taskID1, "urgent", "user1")

	// Get events for task 1 only
	events, err := svc.GetEventsByTask(ctx, taskID1)
	require.NoError(t, err)
	require.Len(t, events, 2)

	// Verify all events belong to task 1
	for _, event := range events {
		require.Equal(t, taskID1, event.TaskID)
	}

	// Get events for task 2 only
	events, err = svc.GetEventsByTask(ctx, taskID2)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, taskID2, events[0].TaskID)
}
