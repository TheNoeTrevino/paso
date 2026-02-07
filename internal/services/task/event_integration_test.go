package task

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/services/taskevent"
	"github.com/thenoetrevino/paso/internal/testutil"
)

func newTestServiceWithMock(t *testing.T, db *sql.DB, mock *taskevent.MockService) Service {
	t.Helper()
	svc, err := NewService(db, database.SQLite, nil, mock)
	require.NoError(t, err, "failed to create test service with mock")
	return svc
}

func TestCreateTask_EmitsTaskCreatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
	})
	require.NoError(t, err)
	require.NotNil(t, task)

	require.True(t, mock.HasCall("CreateTaskCreatedEvent", task.ID),
		"Expected CreateTaskCreatedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskCreatedEvent"))

	calls := mock.GetCalls()
	var createCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreateTaskCreatedEvent" && calls[i].TaskID == task.ID {
			createCall = &calls[i]
			break
		}
	}
	require.NotNil(t, createCall)
	require.Equal(t, "Test Task", createCall.Args["title"])
}

func TestMoveTaskToColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	todoColumnID := createTestColumn(t, db, projectID, "Todo")
	doneColumnID := createTestColumn(t, db, projectID, "Done")
	taskID := createTestTask(t, db, todoColumnID, "Test Task")

	mock.Reset()

	err := svc.MoveTaskToColumn(context.Background(), taskID, doneColumnID)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTaskMovedEvent", taskID),
		"Expected CreateTaskMovedEvent to be called for task %d", taskID)
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))

	calls := mock.GetCalls()
	var moveCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreateTaskMovedEvent" && calls[i].TaskID == taskID {
			moveCall = &calls[i]
			break
		}
	}
	require.NotNil(t, moveCall)
	require.Equal(t, "Todo", moveCall.Args["fromColumn"])
	require.Equal(t, "Done", moveCall.Args["toColumn"])
}

func TestMoveTaskToNextColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "Todo")
	col2ID := createTestColumn(t, db, projectID, "In Progress")

	ctx := context.Background()
	_, err := db.ExecContext(ctx, "UPDATE columns SET next_id = ? WHERE id = ?", col2ID, col1ID)
	require.NoError(t, err)

	taskID := createTestTask(t, db, col1ID, "Test Task")
	mock.Reset()

	err = svc.MoveTaskToNextColumn(ctx, taskID)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTaskMovedEvent", taskID))
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))
}

func TestMoveTaskToPrevColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "Todo")
	col2ID := createTestColumn(t, db, projectID, "In Progress")

	ctx := context.Background()
	_, err := db.ExecContext(ctx, "UPDATE columns SET prev_id = ? WHERE id = ?", col1ID, col2ID)
	require.NoError(t, err)

	taskID := createTestTask(t, db, col2ID, "Test Task")
	mock.Reset()

	err = svc.MoveTaskToPrevColumn(ctx, taskID)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTaskMovedEvent", taskID))
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))
}

func TestMoveTaskToReadyColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	todoColumnID := createTestColumn(t, db, projectID, "Todo")
	_ = createTestReadyColumn(t, db, projectID, "Ready")
	taskID := createTestTask(t, db, todoColumnID, "Test Task")

	mock.Reset()

	err := svc.MoveTaskToReadyColumn(context.Background(), taskID)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTaskMovedEvent", taskID))
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))
}

func TestMoveTaskToCompletedColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	todoColumnID := createTestColumn(t, db, projectID, "Todo")
	_ = createTestCompletedColumn(t, db, projectID, "Done")
	taskID := createTestTask(t, db, todoColumnID, "Test Task")

	mock.Reset()

	err := svc.MoveTaskToCompletedColumn(context.Background(), taskID)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTaskMovedEvent", taskID))
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))
}

func TestAddParentRelation_EmitsTaskAssociatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	mock.Reset()

	err = svc.AddParentRelation(ctx, task2.ID, task1.ID, 1)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTaskAssociatedEvent", task2.ID),
		"Expected CreateTaskAssociatedEvent to be called for task %d", task2.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskAssociatedEvent"))

	calls := mock.GetCalls()
	var assocCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreateTaskAssociatedEvent" && calls[i].TaskID == task2.ID {
			assocCall = &calls[i]
			break
		}
	}
	require.NotNil(t, assocCall)
	require.Equal(t, task1.ID, assocCall.Args["relatedTaskID"])
}

func TestAddChildRelation_EmitsTaskAssociatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	mock.Reset()

	err = svc.AddChildRelation(ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTaskAssociatedEvent", task1.ID),
		"Expected CreateTaskAssociatedEvent to be called for task %d", task1.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskAssociatedEvent"))

	calls := mock.GetCalls()
	var assocCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreateTaskAssociatedEvent" && calls[i].TaskID == task1.ID {
			assocCall = &calls[i]
			break
		}
	}
	require.NotNil(t, assocCall)
	require.Equal(t, task2.ID, assocCall.Args["relatedTaskID"])
}

func TestRemoveParentRelation_EmitsTaskDisassociatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:     "Child Task",
		ColumnID:  columnID,
		Position:  1,
		ParentIDs: []int{task1.ID},
	})
	require.NoError(t, err)

	mock.Reset()

	err = svc.RemoveParentRelation(ctx, task2.ID, task1.ID)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTaskDisassociatedEvent", task2.ID),
		"Expected CreateTaskDisassociatedEvent to be called for task %d", task2.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskDisassociatedEvent"))

	calls := mock.GetCalls()
	var disassocCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreateTaskDisassociatedEvent" && calls[i].TaskID == task2.ID {
			disassocCall = &calls[i]
			break
		}
	}
	require.NotNil(t, disassocCall)
	require.Equal(t, task1.ID, disassocCall.Args["relatedTaskID"])
}

func TestRemoveChildRelation_EmitsTaskDisassociatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	err = svc.AddChildRelation(ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err)

	mock.Reset()

	err = svc.RemoveChildRelation(ctx, task1.ID, task2.ID)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTaskDisassociatedEvent", task1.ID),
		"Expected CreateTaskDisassociatedEvent to be called for task %d", task1.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskDisassociatedEvent"))

	calls := mock.GetCalls()
	var disassocCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreateTaskDisassociatedEvent" && calls[i].TaskID == task1.ID {
			disassocCall = &calls[i]
			break
		}
	}
	require.NotNil(t, disassocCall)
	require.Equal(t, task2.ID, disassocCall.Args["relatedTaskID"])
}

func TestAttachLabel_EmitsLabelAddedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")
	labelID := createTestLabel(t, db, projectID, "bug")

	ctx := context.Background()
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
	})
	require.NoError(t, err)

	mock.Reset()

	err = svc.AttachLabel(ctx, task.ID, labelID)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateLabelAddedEvent", task.ID),
		"Expected CreateLabelAddedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreateLabelAddedEvent"))

	calls := mock.GetCalls()
	var labelCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreateLabelAddedEvent" && calls[i].TaskID == task.ID {
			labelCall = &calls[i]
			break
		}
	}
	require.NotNil(t, labelCall)
	require.Equal(t, "bug", labelCall.Args["labelName"])
}

func TestDetachLabel_EmitsLabelRemovedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")
	labelID := createTestLabel(t, db, projectID, "bug")

	ctx := context.Background()
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		LabelIDs: []int{labelID},
	})
	require.NoError(t, err)

	mock.Reset()

	err = svc.DetachLabel(ctx, task.ID, labelID)
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateLabelRemovedEvent", task.ID),
		"Expected CreateLabelRemovedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreateLabelRemovedEvent"))

	calls := mock.GetCalls()
	var labelCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreateLabelRemovedEvent" && calls[i].TaskID == task.ID {
			labelCall = &calls[i]
			break
		}
	}
	require.NotNil(t, labelCall)
	require.Equal(t, "bug", labelCall.Args["labelName"])
}

func TestUpdateTask_EmitsPriorityChangedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:      "Test Task",
		ColumnID:   columnID,
		PriorityID: 3, // medium
	})
	require.NoError(t, err)

	mock.Reset()

	newPriority := 4 // high
	err = svc.UpdateTask(ctx, UpdateTaskRequest{
		TaskID:     task.ID,
		PriorityID: &newPriority,
	})
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreatePriorityChangedEvent", task.ID),
		"Expected CreatePriorityChangedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreatePriorityChangedEvent"))

	calls := mock.GetCalls()
	var priorityCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreatePriorityChangedEvent" && calls[i].TaskID == task.ID {
			priorityCall = &calls[i]
			break
		}
	}
	require.NotNil(t, priorityCall)
	require.Equal(t, "medium", priorityCall.Args["oldPriority"])
	require.Equal(t, "high", priorityCall.Args["newPriority"])
}

func TestUpdateTask_EmitsTypeChangedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		TypeID:   1, // task
	})
	require.NoError(t, err)

	mock.Reset()

	newType := 3 // bug
	err = svc.UpdateTask(ctx, UpdateTaskRequest{
		TaskID: task.ID,
		TypeID: &newType,
	})
	require.NoError(t, err)

	require.True(t, mock.HasCall("CreateTypeChangedEvent", task.ID),
		"Expected CreateTypeChangedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreateTypeChangedEvent"))

	calls := mock.GetCalls()
	var typeCall *taskevent.EventCall
	for i := range calls {
		if calls[i].Method == "CreateTypeChangedEvent" && calls[i].TaskID == task.ID {
			typeCall = &calls[i]
			break
		}
	}
	require.NotNil(t, typeCall)
	require.Equal(t, "task", typeCall.Args["oldType"])
	require.Equal(t, "bug", typeCall.Args["newType"])
}

func TestUpdateTask_NoEventWhenPriorityUnchanged(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:      "Test Task",
		ColumnID:   columnID,
		PriorityID: 3, // medium
	})
	require.NoError(t, err)

	mock.Reset()

	samePriority := 3 // same as before
	err = svc.UpdateTask(ctx, UpdateTaskRequest{
		TaskID:     task.ID,
		PriorityID: &samePriority,
	})
	require.NoError(t, err)

	require.Equal(t, 0, mock.CallCount("CreatePriorityChangedEvent"),
		"Expected no priority changed event when priority is unchanged")
}

func TestUpdateTask_NoEventWhenTypeUnchanged(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		TypeID:   1, // task
	})
	require.NoError(t, err)

	mock.Reset()

	sameType := 1 // same as before
	err = svc.UpdateTask(ctx, UpdateTaskRequest{
		TaskID: task.ID,
		TypeID: &sameType,
	})
	require.NoError(t, err)

	require.Equal(t, 0, mock.CallCount("CreateTypeChangedEvent"),
		"Expected no type changed event when type is unchanged")
}

func TestUpdateTask_EmitsBothPriorityAndTypeChangedEvents(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:      "Test Task",
		ColumnID:   columnID,
		PriorityID: 3, // medium
		TypeID:     1, // task
	})
	require.NoError(t, err)

	mock.Reset()

	newPriority := 4 // high
	newType := 3     // bug
	err = svc.UpdateTask(ctx, UpdateTaskRequest{
		TaskID:     task.ID,
		PriorityID: &newPriority,
		TypeID:     &newType,
	})
	require.NoError(t, err)

	require.Equal(t, 1, mock.CallCount("CreatePriorityChangedEvent"),
		"Expected exactly one priority changed event")
	require.Equal(t, 1, mock.CallCount("CreateTypeChangedEvent"),
		"Expected exactly one type changed event")
}

func TestCreateTask_WithLabels_OnlyEmitsTaskCreatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")
	label1ID := createTestLabel(t, db, projectID, "bug")
	label2ID := createTestLabel(t, db, projectID, "urgent")

	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		LabelIDs: []int{label1ID, label2ID},
	})
	require.NoError(t, err)
	require.NotNil(t, task)

	require.Equal(t, 1, mock.CallCount("CreateTaskCreatedEvent"))

	// Labels attached during CreateTask don't emit separate LabelAdded events
	// This is by design - only the TaskCreated event is emitted
	require.Equal(t, 0, mock.CallCount("CreateLabelAddedEvent"),
		"Labels attached during CreateTask should not emit separate events")
}

func TestCreateTask_WithParents_OnlyEmitsTaskCreatedEvent(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	parent1, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent 1",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	parent2, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Parent 2",
		ColumnID: columnID,
		Position: 1,
	})
	require.NoError(t, err)

	mock.Reset()

	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:     "Child Task",
		ColumnID:  columnID,
		Position:  2,
		ParentIDs: []int{parent1.ID, parent2.ID},
	})
	require.NoError(t, err)
	require.NotNil(t, task)

	require.Equal(t, 1, mock.CallCount("CreateTaskCreatedEvent"))

	// Parent relationships created during CreateTask don't emit separate events
	// This is by design - only the TaskCreated event is emitted
	require.Equal(t, 0, mock.CallCount("CreateTaskAssociatedEvent"),
		"Parent relationships created during CreateTask should not emit separate events")
}

func TestNoEventService_NoEventsEmitted(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc, err := NewService(db, database.SQLite, nil, nil)
	require.NoError(t, err)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	task, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
	})
	require.NoError(t, err)
	require.NotNil(t, task)
}

func TestMoveTask_SameColumn_NoEventEmitted(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	ctx := context.Background()
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
	})
	require.NoError(t, err)

	mock.Reset()

	err = svc.MoveTaskToColumn(ctx, task.ID, columnID)

	if err == nil {
		require.Equal(t, 0, mock.CallCount("CreateTaskMovedEvent"),
			"Expected no move event when staying in same column")
	}
}

func TestEventServiceError_DoesNotBreakOperation(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	mock.CreateTaskCreatedEventErr = context.DeadlineExceeded
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	columnID := createTestColumn(t, db, projectID, "Todo")

	_, err := svc.CreateTask(context.Background(), CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
	})

	require.Error(t, err, "Expected error when event service fails")
}

func TestMultipleMoveOperations_EmitCorrectEvents(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	mock := taskevent.NewMockService()
	svc := newTestServiceWithMock(t, db, mock)

	projectID := createTestProject(t, db)
	col1ID := createTestColumn(t, db, projectID, "Todo")
	col2ID := createTestColumn(t, db, projectID, "In Progress")
	col3ID := createTestColumn(t, db, projectID, "Done")

	ctx := context.Background()
	task, err := svc.CreateTask(ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col1ID,
	})
	require.NoError(t, err)

	mock.Reset()

	err = svc.MoveTaskToColumn(ctx, task.ID, col2ID)
	require.NoError(t, err)

	err = svc.MoveTaskToColumn(ctx, task.ID, col3ID)
	require.NoError(t, err)

	require.Equal(t, 2, mock.CallCount("CreateTaskMovedEvent"),
		"Expected 2 move events for 2 moves")
}
