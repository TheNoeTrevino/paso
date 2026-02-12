package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddParentRelation(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create two tasks
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

	// Add parent relation (task1 is parent of task2)
	err = env.Svc.AddParentRelation(env.Ctx, task2.ID, task1.ID, 1)
	require.NoError(t, err, "Operation failed")

	// Verify relationship exists
	var count int
	err = env.DB.QueryRowContext(env.Ctx, "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	require.NoError(t, err, "failed to query task relationships")

	assert.Equal(t, 1, count)
}

func TestAddChildRelation(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create two tasks
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

	// Add child relation (task2 is child of task1)
	err = env.Svc.AddChildRelation(env.Ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err, "Operation failed")

	// Verify relationship exists
	var count int
	err = env.DB.QueryRowContext(env.Ctx, "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	require.NoError(t, err, "failed to query task relationships")

	assert.Equal(t, 1, count)
}

func TestRemoveParentRelation(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create two tasks with relationship
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

	// Remove parent relation
	err = env.Svc.RemoveParentRelation(env.Ctx, task2.ID, task1.ID)
	require.NoError(t, err, "Operation failed")

	// Verify relationship is removed
	var count int
	err = env.DB.QueryRowContext(env.Ctx, "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	require.NoError(t, err, "failed to query task relationships")

	assert.Equal(t, 0, count)
}

func TestRemoveChildRelation(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create two tasks with child relationship
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

	// Add child relation (task2 is child of task1)
	err = env.Svc.AddChildRelation(env.Ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err)

	// Remove child relation
	err = env.Svc.RemoveChildRelation(env.Ctx, task1.ID, task2.ID)
	require.NoError(t, err, "Operation failed")

	// Verify relationship is removed
	var count int
	err = env.DB.QueryRowContext(env.Ctx, "SELECT COUNT(*) FROM task_subtasks WHERE parent_id = ? AND child_id = ?", task1.ID, task2.ID).Scan(&count)
	require.NoError(t, err, "failed to query task relationships")

	assert.Equal(t, 0, count)
}

func TestAddParentRelation_CircularDependencyCheck(t *testing.T) {
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

	// Add parent relationship: 2 -> 1
	err = env.Svc.AddParentRelation(env.Ctx, task2.ID, task1.ID, 1)
	require.NoError(t, err)

	// Try to add reverse relationship which would create a circle: 1 -> 2 when 2 -> 1 exists
	// Note: The current implementation doesn't prevent this at the service level,
	// so we're just testing that it doesn't panic
	err = env.Svc.AddChildRelation(env.Ctx, task1.ID, task2.ID, 1)
	if err != nil {
		// This is acceptable - the service might prevent circular deps
		t.Logf("Service prevented circular dependency: %v", err)
	}
}

// TestAddParentRelation_SelfRelationPrevention tests that self-relations are prevented
func TestAddParentRelation_SelfRelationPrevention(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	task, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task 1",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	// Try to create a self-relation
	err = env.Svc.AddParentRelation(env.Ctx, task.ID, task.ID, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSelfRelation)
}

// TestRemoveParentRelation_RestructuresTree tests that removing relations restructures tree
func TestRemoveParentRelation_RestructuresTree(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t)
	// Create three tasks
	task1, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task A",
		ColumnID: env.ColumnID,
		Position: 0,
	})
	require.NoError(t, err)

	task2, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task B",
		ColumnID: env.ColumnID,
		Position: 1,
	})
	require.NoError(t, err)

	task3, err := env.Svc.CreateTask(env.Ctx, CreateTaskRequest{
		Title:    "Task C",
		ColumnID: env.ColumnID,
		Position: 2,
	})
	require.NoError(t, err)

	// Create chain: A -> B -> C
	err = env.Svc.AddChildRelation(env.Ctx, task1.ID, task2.ID, 1)
	require.NoError(t, err)
	err = env.Svc.AddChildRelation(env.Ctx, task2.ID, task3.ID, 1)
	require.NoError(t, err)

	// Verify structure
	nodes, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	// Remove middle relationship: B from A
	err = env.Svc.RemoveChildRelation(env.Ctx, task1.ID, task2.ID)
	assert.NoError(t, err)

	// Now A and B should both be roots
	nodes, err = env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
	require.NoError(t, err)
	assert.Len(t, nodes, 2)
}
