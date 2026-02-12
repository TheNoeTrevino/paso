package task

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestMoveTaskToNextColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	col1ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "To Do")
	col2ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "In Progress")
	fixtures.LinkColumns(t, env.DB, env.Dialect, col1ID, col2ID)
	// Create task in first column
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col1ID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to next column
	err = env.Svc.MoveTaskToNextColumn(env.Ctx, task.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved to col2
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task.ID, col2ID)
}

func TestMoveTaskToNextColumn_LastColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	columnID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "Done")
	// Create task in last column (no next_id)
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: columnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to next column (should fail)
	err = env.Svc.MoveTaskToNextColumn(env.Ctx, task.ID)

	require.Error(t, err)
}

func TestMoveTaskToNextColumn_InvalidID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	err := env.Svc.MoveTaskToNextColumn(env.Ctx, 999)

	require.Error(t, err)
}

func TestMoveTaskToPrevColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	col1ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "To Do")
	col2ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "In Progress")
	fixtures.LinkColumns(t, env.DB, env.Dialect, col1ID, col2ID)
	// Create task in second column
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col2ID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to previous column
	err = env.Svc.MoveTaskToPrevColumn(env.Ctx, task.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved to col1
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task.ID, col1ID)
}

func TestMoveTaskToPrevColumn_FirstColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create task in first column (no prev_id)
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to previous column (should fail)
	err = env.Svc.MoveTaskToPrevColumn(env.Ctx, task.ID)

	require.Error(t, err)
}

func TestMoveTaskToPrevColumn_InvalidID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	err := env.Svc.MoveTaskToPrevColumn(env.Ctx, 999)

	require.Error(t, err)
}

func TestMoveTaskToColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	col1ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "To Do")
	col2ID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "Done")
	// Create task in first column
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: col1ID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to specific column
	err = env.Svc.MoveTaskToColumn(env.Ctx, task.ID, col2ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task.ID, col2ID)
}

func TestMoveTaskToColumn_InvalidColumnID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create task
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to invalid column
	err = env.Svc.MoveTaskToColumn(env.Ctx, task.ID, 999)

	require.Error(t, err)
}

func TestMoveTaskToColumn_InvalidTaskID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Try to move invalid task
	err := env.Svc.MoveTaskToColumn(env.Ctx, 999, env.ColumnID)

	require.Error(t, err)
}

func TestMoveTaskUp(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create two tasks
	task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: env.ColumnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Move task2 up (should swap positions with task1)
	err = env.Svc.MoveTaskUp(env.Ctx, task2.ID)
	require.NoError(t, err, "Operation failed")

	// Verify positions swapped
	var pos1, pos2 int64
	err = env.DB.QueryRowContext(env.Ctx, "SELECT position FROM tasks WHERE id = ?", task1.ID).Scan(&pos1)
	require.NoError(t, err)

	err = env.DB.QueryRowContext(env.Ctx, "SELECT position FROM tasks WHERE id = ?", task2.ID).Scan(&pos2)
	require.NoError(t, err)

	assert.Less(t, pos2, pos1)
}

func TestMoveTaskUp_FirstPosition(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create task at first position
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move up (should fail - no task above)
	err = env.Svc.MoveTaskUp(env.Ctx, task.ID)

	require.Error(t, err)
}

func TestMoveTaskUp_InvalidID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	err := env.Svc.MoveTaskUp(env.Ctx, 999)

	require.Error(t, err)
}

func TestMoveTaskDown(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create two tasks
	task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: env.ColumnID,
		Position: 1,
	})
	require.NoError(t, err)

	// Move task1 down (should swap positions with task2)
	err = env.Svc.MoveTaskDown(env.Ctx, task1.ID)
	require.NoError(t, err, "Operation failed")

	// Verify positions swapped
	var pos1, pos2 int64
	err = env.DB.QueryRowContext(env.Ctx, "SELECT position FROM tasks WHERE id = ?", task1.ID).Scan(&pos1)
	require.NoError(t, err)

	err = env.DB.QueryRowContext(env.Ctx, "SELECT position FROM tasks WHERE id = ?", task2.ID).Scan(&pos2)
	require.NoError(t, err)

	assert.Greater(t, pos1, pos2)
}

func TestMoveTaskDown_LastPosition(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create task at last position
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move down (should fail - no task below)
	err = env.Svc.MoveTaskDown(env.Ctx, task.ID)

	require.Error(t, err)
}

func TestMoveTaskDown_InvalidID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	err := env.Svc.MoveTaskDown(env.Ctx, 999)

	require.Error(t, err)
}

func TestMoveTaskToReadyColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	todoColID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "To Do")
	readyColID := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Ready", true, false, false)
	// Create task in To Do column
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: todoColID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to ready column
	err = env.Svc.MoveTaskToReadyColumn(env.Ctx, task.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved to ready column
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task.ID, readyColID)
}

func TestMoveTaskToReadyColumn_InvalidTaskID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Try to move non-existent task
	err := env.Svc.MoveTaskToReadyColumn(env.Ctx, 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidTaskID) || errors.Is(err, sql.ErrNoRows))
}

func TestMoveTaskToReadyColumn_NoReadyColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create task
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to ready column when none exists
	err = env.Svc.MoveTaskToReadyColumn(env.Ctx, task.ID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows) || err.Error() == "no ready column configured for this project")
}

func TestMoveTaskToReadyColumn_AlreadyInReadyColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	readyColID := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Ready", true, false, false)
	// Create task already in ready column
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: readyColID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to ready column (already there)
	err = env.Svc.MoveTaskToReadyColumn(env.Ctx, task.ID)

	assert.ErrorIs(t, err, ErrTaskAlreadyInTargetColumn)
}

func TestMoveTaskToReadyColumn_ZeroTaskID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Try to move task with ID 0
	err := env.Svc.MoveTaskToReadyColumn(env.Ctx, 0)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestMoveTaskToReadyColumn_NegativeTaskID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Try to move task with negative ID
	err := env.Svc.MoveTaskToReadyColumn(env.Ctx, -1)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestMoveTaskToCompletedColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	todoColID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "To Do")
	completedColID := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Done", false, true, false)
	// Create task in To Do column
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: todoColID,
		Position: 0,
	})
	require.NoError(t, err)

	// Move to completed column
	err = env.Svc.MoveTaskToCompletedColumn(env.Ctx, task.ID)
	require.NoError(t, err, "Operation failed")

	// Verify task moved to completed column
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task.ID, completedColID)
}

func TestMoveTaskToCompletedColumn_InvalidTaskID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Try to move non-existent task
	err := env.Svc.MoveTaskToCompletedColumn(env.Ctx, 999)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidTaskID) || errors.Is(err, sql.ErrNoRows))
}

func TestMoveTaskToCompletedColumn_NoCompletedColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create task
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to completed column when none exists
	err = env.Svc.MoveTaskToCompletedColumn(env.Ctx, task.ID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows) || err.Error() == "no completed column configured for this project")
}

func TestMoveTaskToCompletedColumn_AlreadyInCompletedColumn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	completedColID := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Done", false, true, false)
	// Create task already in completed column
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: completedColID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to move to completed column (already there)
	err = env.Svc.MoveTaskToCompletedColumn(env.Ctx, task.ID)

	assert.ErrorIs(t, err, ErrTaskAlreadyInTargetColumn)
}

func TestMoveTaskToCompletedColumn_ZeroTaskID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Try to move task with ID 0
	err := env.Svc.MoveTaskToCompletedColumn(env.Ctx, 0)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestMoveTaskToCompletedColumn_NegativeTaskID(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Try to move task with negative ID
	err := env.Svc.MoveTaskToCompletedColumn(env.Ctx, -1)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestMoveTaskToCompletedColumn_MultipleTasksInProject(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	todoColID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "To Do")
	inProgressColID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "In Progress")
	completedColID := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Done", false, true, false)
	// Create multiple tasks in different columns
	task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: todoColID,
		Position: 0,
	})
	require.NoError(t, err)
	task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: inProgressColID,
		Position: 0,
	})
	require.NoError(t, err)
	task3, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 3",
		ColumnID: todoColID,
		Position: 1,
	})
	require.NoError(t, err)

	// Move task2 to completed
	err = env.Svc.MoveTaskToCompletedColumn(env.Ctx, task2.ID)
	require.NoError(t, err)

	// Verify task2 is in completed column
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task2.ID, completedColID)

	// Verify other tasks are unchanged
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task1.ID, todoColID)
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task3.ID, todoColID)
}

func TestMoveTaskToReadyColumn_MultipleTasksInProject(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	todoColID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "To Do")
	inProgressColID := fixtures.CreateTestColumn(t, env.DB, env.Dialect, env.ProjectID, "In Progress")
	readyColID := fixtures.CreateColumnWithFlags(t, env.DB, env.Dialect, env.ProjectID, "Ready", true, false, false)
	// Create multiple tasks in different columns
	task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: todoColID,
		Position: 0,
	})
	require.NoError(t, err)
	task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 2",
		ColumnID: inProgressColID,
		Position: 0,
	})
	require.NoError(t, err)
	task3, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 3",
		ColumnID: todoColID,
		Position: 1,
	})
	require.NoError(t, err)

	// Move task2 to ready
	err = env.Svc.MoveTaskToReadyColumn(env.Ctx, task2.ID)
	require.NoError(t, err)

	// Verify task2 is in ready column
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task2.ID, readyColID)

	// Verify other tasks are unchanged
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task1.ID, todoColID)
	fixtures.AssertTaskInColumn(t, env.DB, env.Dialect, task3.ID, todoColID)
}
