package taskevent

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func createTestTaskWithSetup(t *testing.T, db *sql.DB, d fixtures.Dialect) int {
	t.Helper()
	projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
	columnID := fixtures.CreateTestColumn(t, db, d, projectID, "Todo")
	return fixtures.CreateTestTask(t, db, d, columnID, "Test Task")
}

func TestCreateTaskCreatedEvent(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTaskCreatedEvent(ctx, q, taskID, "My New Task", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, "Task created: My New Task", events[0].Content)
		require.Equal(t, "testuser", events[0].Author)
	})
}

func TestCreateTaskMovedEvent(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTaskMovedEvent(ctx, q, taskID, "Todo", "In Progress", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, "Moved from 'Todo' to 'In Progress'", events[0].Content)
	})
}

func TestCreateTaskAssociatedEvent(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTaskAssociatedEvent(ctx, q, taskID, 42, "Other Task", "Blocked By", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, "Associated to task #42 (Other Task) as: Blocked By", events[0].Content)
	})
}

func TestCreateTaskDisassociatedEvent(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTaskDisassociatedEvent(ctx, q, taskID, 42, "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, "Removed association to task #42", events[0].Content)
	})
}

func TestCreateLabelAddedEvent(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateLabelAddedEvent(ctx, q, taskID, "urgent", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, "Added label: urgent", events[0].Content)
	})
}

func TestCreateLabelRemovedEvent(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateLabelRemovedEvent(ctx, q, taskID, "urgent", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, "Removed label: urgent", events[0].Content)
	})
}

func TestCreatePriorityChangedEvent(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreatePriorityChangedEvent(ctx, q, taskID, "medium", "high", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, "Changed priority from 'medium' to 'high'", events[0].Content)
	})
}

func TestCreateTypeChangedEvent(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTypeChangedEvent(ctx, q, taskID, "task", "bug", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, "Changed type from 'task' to 'bug'", events[0].Content)
	})
}

func TestGetEventsByTask_MultipleEvents(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTaskCreatedEvent(ctx, q, taskID, "Task", "user1")
		require.NoError(t, err)
		err = svc.CreateLabelAddedEvent(ctx, q, taskID, "urgent", "user2")
		require.NoError(t, err)
		err = svc.CreateTaskMovedEvent(ctx, q, taskID, "Todo", "Done", "user3")
		require.NoError(t, err)

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
	})
}

func TestGetEventsByTask_EmptyList(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)

		events, err := svc.GetEventsByTask(context.Background(), taskID)
		require.NoError(t, err)
		require.Len(t, events, 0)
	})
}

func TestGetEventsByTask_InvalidTaskID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, _ fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		_, err = svc.GetEventsByTask(ctx, 0)
		require.ErrorIs(t, err, ErrInvalidTaskID)

		_, err = svc.GetEventsByTask(ctx, -1)
		require.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestCreateEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, _ fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTaskCreatedEvent(ctx, q, 0, "Task", "user")
		require.ErrorIs(t, err, ErrInvalidTaskID)

		err = svc.CreateTaskCreatedEvent(ctx, q, -1, "Task", "user")
		require.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestCreateTaskMovedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, _ fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)

		err = svc.CreateTaskMovedEvent(context.Background(), q, 0, "Todo", "Done", "user")
		require.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestCreateTaskAssociatedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, _ fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)

		err = svc.CreateTaskAssociatedEvent(context.Background(), q, 0, 42, "Other", "Blocked By", "user")
		require.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestCreateTaskDisassociatedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, _ fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)

		err = svc.CreateTaskDisassociatedEvent(context.Background(), q, 0, 42, "user")
		require.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestCreateLabelAddedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, _ fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)

		err = svc.CreateLabelAddedEvent(context.Background(), q, 0, "urgent", "user")
		require.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestCreateLabelRemovedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, _ fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)

		err = svc.CreateLabelRemovedEvent(context.Background(), q, 0, "urgent", "user")
		require.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestCreatePriorityChangedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, _ fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)

		err = svc.CreatePriorityChangedEvent(context.Background(), q, 0, "medium", "high", "user")
		require.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestCreateTypeChangedEvent_InvalidTaskID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, _ fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)

		err = svc.CreateTypeChangedEvent(context.Background(), q, 0, "task", "bug", "user")
		require.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestEventAuthorPreserved(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		// Create events with different authors
		err = svc.CreateTaskCreatedEvent(ctx, q, taskID, "Task", "alice")
		require.NoError(t, err)
		err = svc.CreateLabelAddedEvent(ctx, q, taskID, "urgent", "bob")
		require.NoError(t, err)
		err = svc.CreateTaskMovedEvent(ctx, q, taskID, "Todo", "Done", "charlie")
		require.NoError(t, err)

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
	})
}

func TestEventTimestampSet(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTaskCreatedEvent(ctx, q, taskID, "Task", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)

		// CreatedAt should be set
		require.False(t, events[0].CreatedAt.IsZero(), "CreatedAt should be set")
	})
}

func TestEventTaskIDPreserved(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTaskCreatedEvent(ctx, q, taskID, "Task", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, taskID, events[0].TaskID)
	})
}

func TestEventIDAssigned(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		taskID := createTestTaskWithSetup(t, db, d)
		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		err = svc.CreateTaskCreatedEvent(ctx, q, taskID, "Task", "testuser")
		require.NoError(t, err)

		events, err := svc.GetEventsByTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Greater(t, events[0].ID, 0, "Event ID should be assigned")
	})
}

func TestGetEventsByTask_OnlyReturnsEventsForSpecificTask(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType)
		require.NoError(t, err)
		ctx := context.Background()

		// Create two tasks
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		columnID := fixtures.CreateTestColumn(t, db, d, projectID, "Todo")
		taskID1 := fixtures.CreateTestTask(t, db, d, columnID, "Task 1")
		taskID2 := fixtures.CreateTestTask(t, db, d, columnID, "Task 2")

		q, err := database.NewQuerier(db, dbType)
		require.NoError(t, err)

		// Create events for both tasks
		err = svc.CreateTaskCreatedEvent(ctx, q, taskID1, "Task 1", "user1")
		require.NoError(t, err)
		err = svc.CreateTaskCreatedEvent(ctx, q, taskID2, "Task 2", "user2")
		require.NoError(t, err)
		err = svc.CreateLabelAddedEvent(ctx, q, taskID1, "urgent", "user1")
		require.NoError(t, err)

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
	})
}
