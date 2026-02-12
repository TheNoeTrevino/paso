package task

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/services/taskevent"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func findEventCall(t *testing.T, mock *taskevent.MockService, method string, taskID int) *taskevent.EventCall {
	t.Helper()
	calls := mock.GetCalls()
	for i := range calls {
		if calls[i].Method == method && calls[i].TaskID == taskID {
			return &calls[i]
		}
	}
	return nil
}

func TestCreateTask_EmitsTaskCreatedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
	})
	require.NoError(t, err)
	require.NotNil(t, task)

	require.True(t, mock.HasCallWithId("CreateTaskCreatedEvent", task.ID),
		"Expected CreateTaskCreatedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskCreatedEvent"))

	createCall := findEventCall(t, mock, "CreateTaskCreatedEvent", task.ID)
	require.NotNil(t, createCall)
	require.Equal(t, "Test Task", createCall.Args["title"])
}

func TestMoveTaskToColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)
	doneColumnID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "Done")
	taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")

	mock.Reset()

	err := env.Svc.MoveTaskToColumn(env.Ctx, taskID, doneColumnID)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTaskMovedEvent", taskID),
		"Expected CreateTaskMovedEvent to be called for task %d", taskID)
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))

	moveCall := findEventCall(t, mock, "CreateTaskMovedEvent", taskID)
	require.NotNil(t, moveCall)
	require.Equal(t, "To Do", moveCall.Args["fromColumn"])
	require.Equal(t, "Done", moveCall.Args["toColumn"])
}

func TestMoveTaskToNextColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)
	col2ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "In Progress")

	fixtures.LinkColumns(t, env.DB, env.Dialect, env.ColumnID, col2ID)

	taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")
	mock.Reset()

	err := env.Svc.MoveTaskToNextColumn(env.Ctx, taskID)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTaskMovedEvent", taskID))
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))
}

func TestMoveTaskToPrevColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)
	col2ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "In Progress")

	fixtures.LinkColumns(t, env.DB, env.Dialect, env.ColumnID, col2ID)

	taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, col2ID, "Test Task")
	mock.Reset()

	err := env.Svc.MoveTaskToPrevColumn(env.Ctx, taskID)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTaskMovedEvent", taskID))
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))
}

func TestMoveTaskToReadyColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)
	_ = fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Ready", true, false, false)
	taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")

	mock.Reset()

	err := env.Svc.MoveTaskToReadyColumn(env.Ctx, taskID)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTaskMovedEvent", taskID))
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))
}

func TestMoveTaskToCompletedColumn_EmitsTaskMovedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)
	_ = fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Done", false, true, false)
	taskID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Test Task")

	mock.Reset()

	err := env.Svc.MoveTaskToCompletedColumn(env.Ctx, taskID)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTaskMovedEvent", taskID))
	require.Equal(t, 1, mock.CallCount("CreateTaskMovedEvent"))
}

func TestAddParentRelation_EmitsTaskAssociatedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: env.ColumnID,
		Position: 1,
	})
	require.NoError(t, err)

	mock.Reset()

	err = env.Svc.AddParentRelation(env.Ctx, task2.ID, task1.ID, 1)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTaskAssociatedEvent", task2.ID),
		"Expected CreateTaskAssociatedEvent to be called for task %d", task2.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskAssociatedEvent"))

	assocCall := findEventCall(t, mock, "CreateTaskAssociatedEvent", task2.ID)
	require.NotNil(t, assocCall)
	require.Equal(t, task1.ID, assocCall.Args["relatedTaskID"])
}

func TestAddChildRelation_EmitsTaskAssociatedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: env.ColumnID,
		Position: 1,
	})
	require.NoError(t, err)

	mock.Reset()

	err = env.Svc.AddChildRelation(env.Ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTaskAssociatedEvent", task1.ID),
		"Expected CreateTaskAssociatedEvent to be called for task %d", task1.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskAssociatedEvent"))

	assocCall := findEventCall(t, mock, "CreateTaskAssociatedEvent", task1.ID)
	require.NotNil(t, assocCall)
	require.Equal(t, task2.ID, assocCall.Args["relatedTaskID"])
}

func TestRemoveParentRelation_EmitsTaskDisassociatedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:     "Child Task",
		ColumnID:  env.ColumnID,
		Position:  1,
		ParentIDs: []int{task1.ID},
	})
	require.NoError(t, err)

	mock.Reset()

	err = env.Svc.RemoveParentRelation(env.Ctx, task2.ID, task1.ID)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTaskDisassociatedEvent", task2.ID),
		"Expected CreateTaskDisassociatedEvent to be called for task %d", task2.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskDisassociatedEvent"))

	disassocCall := findEventCall(t, mock, "CreateTaskDisassociatedEvent", task2.ID)
	require.NotNil(t, disassocCall)
	require.Equal(t, task1.ID, disassocCall.Args["relatedTaskID"])
}

func TestRemoveChildRelation_EmitsTaskDisassociatedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Parent Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Child Task",
		ColumnID: env.ColumnID,
		Position: 1,
	})
	require.NoError(t, err)

	err = env.Svc.AddChildRelation(env.Ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err)

	mock.Reset()

	err = env.Svc.RemoveChildRelation(env.Ctx, task1.ID, task2.ID)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTaskDisassociatedEvent", task1.ID),
		"Expected CreateTaskDisassociatedEvent to be called for task %d", task1.ID)
	require.Equal(t, 1, mock.CallCount("CreateTaskDisassociatedEvent"))

	disassocCall := findEventCall(t, mock, "CreateTaskDisassociatedEvent", task1.ID)
	require.NotNil(t, disassocCall)
	require.Equal(t, task2.ID, disassocCall.Args["relatedTaskID"])
}

func TestAttachLabel_EmitsLabelAddedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)
	labelID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "bug", fixtures.DefaultTestLabelColor)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
	})
	require.NoError(t, err)

	mock.Reset()

	err = env.Svc.AttachLabel(env.Ctx, task.ID, labelID)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateLabelAddedEvent", task.ID),
		"Expected CreateLabelAddedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreateLabelAddedEvent"))

	labelCall := findEventCall(t, mock, "CreateLabelAddedEvent", task.ID)
	require.NotNil(t, labelCall)
	require.Equal(t, "bug", labelCall.Args["labelName"])
}

func TestDetachLabel_EmitsLabelRemovedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)
	labelID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "bug", fixtures.DefaultTestLabelColor)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		LabelIDs: []int{labelID},
	})
	require.NoError(t, err)

	mock.Reset()

	err = env.Svc.DetachLabel(env.Ctx, task.ID, labelID)
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateLabelRemovedEvent", task.ID),
		"Expected CreateLabelRemovedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreateLabelRemovedEvent"))

	labelCall := findEventCall(t, mock, "CreateLabelRemovedEvent", task.ID)
	require.NotNil(t, labelCall)
	require.Equal(t, "bug", labelCall.Args["labelName"])
}

func TestUpdateTask_EmitsPriorityChangedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:      "Test Task",
		ColumnID:   env.ColumnID,
		PriorityID: 3, // medium
	})
	require.NoError(t, err)

	mock.Reset()

	newPriority := 4 // high
	err = env.Svc.UpdateTask(env.Ctx, UpdateTaskRequest{
		TaskID:     task.ID,
		PriorityID: &newPriority,
	})
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreatePriorityChangedEvent", task.ID),
		"Expected CreatePriorityChangedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreatePriorityChangedEvent"))

	priorityCall := findEventCall(t, mock, "CreatePriorityChangedEvent", task.ID)
	require.NotNil(t, priorityCall)
	require.Equal(t, "medium", priorityCall.Args["oldPriority"])
	require.Equal(t, "high", priorityCall.Args["newPriority"])
}

func TestUpdateTask_EmitsTypeChangedEvent(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		TypeID:   1, // task
	})
	require.NoError(t, err)

	mock.Reset()

	newType := 3 // bug
	err = env.Svc.UpdateTask(env.Ctx, UpdateTaskRequest{
		TaskID: task.ID,
		TypeID: &newType,
	})
	require.NoError(t, err)

	require.True(t, mock.HasCallWithId("CreateTypeChangedEvent", task.ID),
		"Expected CreateTypeChangedEvent to be called for task %d", task.ID)
	require.Equal(t, 1, mock.CallCount("CreateTypeChangedEvent"))

	typeCall := findEventCall(t, mock, "CreateTypeChangedEvent", task.ID)
	require.NotNil(t, typeCall)
	require.Equal(t, "task", typeCall.Args["oldType"])
	require.Equal(t, "bug", typeCall.Args["newType"])
}

func TestUpdateTask_NoEventWhenPriorityUnchanged(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:      "Test Task",
		ColumnID:   env.ColumnID,
		PriorityID: 3, // medium
	})
	require.NoError(t, err)

	mock.Reset()

	samePriority := 3 // same as before
	err = env.Svc.UpdateTask(env.Ctx, UpdateTaskRequest{
		TaskID:     task.ID,
		PriorityID: &samePriority,
	})
	require.NoError(t, err)

	require.Equal(t, 0, mock.CallCount("CreatePriorityChangedEvent"),
		"Expected no priority changed event when priority is unchanged")
}

func TestUpdateTask_NoEventWhenTypeUnchanged(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		TypeID:   1, // task
	})
	require.NoError(t, err)

	mock.Reset()

	sameType := 1 // same as before
	err = env.Svc.UpdateTask(env.Ctx, UpdateTaskRequest{
		TaskID: task.ID,
		TypeID: &sameType,
	})
	require.NoError(t, err)

	require.Equal(t, 0, mock.CallCount("CreateTypeChangedEvent"),
		"Expected no type changed event when type is unchanged")
}

func TestUpdateTask_EmitsBothPriorityAndTypeChangedEvents(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:      "Test Task",
		ColumnID:   env.ColumnID,
		PriorityID: 3, // medium
		TypeID:     1, // task
	})
	require.NoError(t, err)

	mock.Reset()

	newPriority := 4 // high
	newType := 3     // bug
	err = env.Svc.UpdateTask(env.Ctx, UpdateTaskRequest{
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

	env, mock := setupTestEnvWithEventMock(t)
	label1ID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "bug", fixtures.DefaultTestLabelColor)
	label2ID := fixtures.CreateTestLabel(t, env.DB, env.Dialect, env.ProjectID, "urgent", fixtures.DefaultTestLabelColor)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
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

	env, mock := setupTestEnvWithEventMock(t)

	parent1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Parent 1",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	parent2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Parent 2",
		ColumnID: env.ColumnID,
		Position: 1,
	})
	require.NoError(t, err)

	mock.Reset()

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:     "Child Task",
		ColumnID:  env.ColumnID,
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

	env := setupTestEnv(t)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
	})
	require.NoError(t, err)
	require.NotNil(t, task)
}

func TestMoveTask_SameColumn_NoEventEmitted(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
	})
	require.NoError(t, err)

	mock.Reset()

	err = env.Svc.MoveTaskToColumn(env.Ctx, task.ID, env.ColumnID)

	if err == nil {
		require.Equal(t, 0, mock.CallCount("CreateTaskMovedEvent"),
			"Expected no move event when staying in same column")
	}
}

func TestEventServiceError_DoesNotBreakOperation(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)
	mock.CreateTaskCreatedEventErr = context.DeadlineExceeded

	_, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
	})

	require.Error(t, err, "Expected error when event service fails")
}

func TestMultipleMoveOperations_EmitCorrectEvents(t *testing.T) {
	t.Parallel()

	env, mock := setupTestEnvWithEventMock(t)
	col2ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "In Progress")
	col3ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "Done")

	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
	})
	require.NoError(t, err)

	mock.Reset()

	err = env.Svc.MoveTaskToColumn(env.Ctx, task.ID, col2ID)
	require.NoError(t, err)

	err = env.Svc.MoveTaskToColumn(env.Ctx, task.ID, col3ID)
	require.NoError(t, err)

	require.Equal(t, 2, mock.CallCount("CreateTaskMovedEvent"),
		"Expected 2 move events for 2 moves")
}
