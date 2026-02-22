package task

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestGetTaskTreeByProject_EmptyProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Get tree for empty project
		tree, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		assert.Len(t, tree, 0)
	})
}

func TestGetTaskTreeByProject_SimpleHierarchy(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create parent and child tasks
		parentID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Parent Task")
		childID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Child Task")

		// Add parent-child relation (non-blocking)
		_, err := env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, 1)",
				d.Placeholder(1), d.Placeholder(2)),
			parentID, childID)
		require.NoError(t, err, "failed to create relation")

		// Get tree
		tree, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		require.Len(t, tree, 1)

		root := tree[0]
		assert.Equal(t, parentID, root.ID)

		require.Len(t, root.Children, 1)

		child := root.Children[0]
		assert.Equal(t, childID, child.ID)
	})
}

func TestGetTaskTreeByProject_CircularDependency(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create three tasks: A -> B -> C -> A (circular)
		taskA := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task A")
		taskB := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task B")
		taskC := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task C")

		// A -> B -> C -> A (circular)
		_, err := env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, 1)",
				d.Placeholder(1), d.Placeholder(2)),
			taskA, taskB)
		require.NoError(t, err)
		_, err = env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, 1)",
				d.Placeholder(1), d.Placeholder(2)),
			taskB, taskC)
		require.NoError(t, err)
		_, err = env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, 1)",
				d.Placeholder(1), d.Placeholder(2)),
			taskC, taskA)
		require.NoError(t, err)

		// Get tree - should handle circular dependency gracefully
		tree, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		// With circular dependency, none are truly "root" tasks
		// This is expected behavior - the function should not crash or hang
		if len(tree) != 0 {
			t.Logf("Got %d root tasks with circular dependencies (expected)", len(tree))
		}
	})
}

func TestGetTaskTreeByProject_DeepNesting(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create a deep hierarchy: Task1 -> Task2 -> Task3 -> Task4 -> Task5
		tasks := make([]int, 5)
		for i := range 5 {
			tasks[i] = fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task "+string(rune('1'+i)))
		}

		// Link them in a chain
		for i := range 4 {
			_, err := env.DB.ExecContext(env.Ctx,
				fmt.Sprintf("INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, 1)",
					d.Placeholder(1), d.Placeholder(2)),
				tasks[i], tasks[i+1])
			require.NoError(t, err, "failed to create relation")
		}

		// Get tree
		tree, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		require.Len(t, tree, 1)

		// Traverse the tree and verify depth
		depth := 0
		current := tree[0]
		for {
			depth++
			if len(current.Children) == 0 {
				break
			}
			require.LessOrEqual(t, depth, 10, "Depth exceeds expected maximum")
			current = current.Children[0]
		}

		assert.Equal(t, 5, depth)
	})
}

func TestGetTaskTreeByProject_BlockingRelationship(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create parent and child tasks
		parentID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Parent Task")
		blockerID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Blocker Task")

		// Add blocking relation (relation_type_id = 2 is the blocker type)
		_, err := env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, 2)",
				d.Placeholder(1), d.Placeholder(2)),
			parentID, blockerID)
		require.NoError(t, err, "failed to create blocking relation")

		// Get tree
		tree, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		require.Len(t, tree, 1)

		root := tree[0]
		require.Len(t, root.Children, 1)

		blocker := root.Children[0]
		assert.True(t, blocker.IsBlocking)
	})
}

func TestGetTaskTreeByProject_SortedByTaskNumber(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create multiple root tasks (no parent relations)
		task1 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task 1")
		task2 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task 2")
		task3 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task 3")

		// Set task numbers (in database they auto-increment, but let's verify sorting)
		_, err := env.DB.ExecContext(env.Ctx, fmt.Sprintf("UPDATE tasks SET task_number = 3 WHERE id = %s", d.Placeholder(1)), task1)
		require.NoError(t, err)
		_, err = env.DB.ExecContext(env.Ctx, fmt.Sprintf("UPDATE tasks SET task_number = 1 WHERE id = %s", d.Placeholder(1)), task2)
		require.NoError(t, err)
		_, err = env.DB.ExecContext(env.Ctx, fmt.Sprintf("UPDATE tasks SET task_number = 2 WHERE id = %s", d.Placeholder(1)), task3)
		require.NoError(t, err)

		// Get tree
		tree, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		require.Len(t, tree, 3)

		// Verify sorted by task number
		assert.Equal(t, 1, tree[0].TaskNumber)
		assert.Equal(t, 2, tree[1].TaskNumber)
		assert.Equal(t, 3, tree[2].TaskNumber)
	})
}

func TestGetTaskTreeByProject_MultipleRootsWithChildren(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create two separate trees
		root1 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Root 1")
		child1 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Child 1")

		root2 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Root 2")
		child2 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Child 2")

		// Link them
		_, err := env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, 1)",
				d.Placeholder(1), d.Placeholder(2)),
			root1, child1)
		require.NoError(t, err)
		_, err = env.DB.ExecContext(env.Ctx,
			fmt.Sprintf("INSERT INTO task_subtasks (parent_id, child_id, relation_type_id) VALUES (%s, %s, 1)",
				d.Placeholder(1), d.Placeholder(2)),
			root2, child2)
		require.NoError(t, err)

		// Get tree
		tree, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Operation failed")

		require.Len(t, tree, 2)

		// Verify both roots have exactly one child
		for i, root := range tree {
			assert.Len(t, root.Children, 1, "Root %d", i)
		}
	})
}

func TestGetTaskTreeByProject_SingleTask(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create a single task
		task1ID := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task 1")

		// Get tree
		nodes, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		// Should have exactly one root node
		require.Len(t, nodes, 1)

		node := nodes[0]
		assert.Equal(t, task1ID, node.ID)
		assert.Equal(t, "Task 1", node.Title)
		assert.Len(t, node.Children, 0)
	})
}

// TestGetTaskTreeByProject_SimpleLinearTree tests a linear chain: A -> B -> C
func TestGetTaskTreeByProject_SimpleLinearTree(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create three tasks in a linear chain: A -> B -> C
		// A is parent (root), B is child of A, C is child of B
		taskA := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task A")
		taskB := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task B")
		taskC := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task C")

		// Create parent-child relationships
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskA, taskB, 1) // A -> B (parent-child)
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskB, taskC, 1) // B -> C (parent-child)

		// Get tree
		nodes, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		// Should have 1 root (Task A)
		require.Len(t, nodes, 1)

		// Check root node
		rootNode := nodes[0]
		assert.Equal(t, taskA, rootNode.ID)
		assert.Equal(t, "Task A", rootNode.Title)

		// Check first level child (B)
		require.Len(t, rootNode.Children, 1)
		childB := rootNode.Children[0]
		assert.Equal(t, taskB, childB.ID)
		assert.Equal(t, "Task B", childB.Title)

		// Check second level child (C)
		require.Len(t, childB.Children, 1)
		childC := childB.Children[0]
		assert.Equal(t, taskC, childC.ID)
		assert.Equal(t, "Task C", childC.Title)

		// Check that C has no children
		assert.Len(t, childC.Children, 0)
	})
}

// TestGetTaskTreeByProject_MultipleRoots tests multiple independent roots
func TestGetTaskTreeByProject_MultipleRoots(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create 3 independent root tasks
		task1 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Root 1")
		task2 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Root 2")
		task3 := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Root 3")

		// Set task numbers to establish deterministic order
		_, err := env.DB.ExecContext(env.Ctx, fmt.Sprintf("UPDATE tasks SET task_number = 1 WHERE id = %s", d.Placeholder(1)), task1)
		require.NoError(t, err)
		_, err = env.DB.ExecContext(env.Ctx, fmt.Sprintf("UPDATE tasks SET task_number = 2 WHERE id = %s", d.Placeholder(1)), task2)
		require.NoError(t, err)
		_, err = env.DB.ExecContext(env.Ctx, fmt.Sprintf("UPDATE tasks SET task_number = 3 WHERE id = %s", d.Placeholder(1)), task3)
		require.NoError(t, err)

		// Get tree
		nodes, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		// Should have 3 roots
		require.Len(t, nodes, 3)

		// Verify roots are sorted by task number (ascending)
		assert.Equal(t, 1, nodes[0].TaskNumber)
		assert.Equal(t, 2, nodes[1].TaskNumber)
		assert.Equal(t, 3, nodes[2].TaskNumber)

		// Verify the task IDs match what we created
		assert.Equal(t, task1, nodes[0].ID)
		assert.Equal(t, task2, nodes[1].ID)
		assert.Equal(t, task3, nodes[2].ID)

		for _, node := range nodes {
			assert.Len(t, node.Children, 0)
		}
	})
}

// TestGetTaskTreeByProject_DiamondDependencies tests diamond pattern: A -> (B,C) -> D
func TestGetTaskTreeByProject_DiamondDependencies(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create diamond pattern:
		//       A
		//      / \
		//     B   C
		//      \ /
		//       D
		taskA := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task A")
		taskB := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task B")
		taskC := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task C")
		taskD := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task D")

		// Build relationships
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskA, taskB, 1) // A -> B
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskA, taskC, 1) // A -> C
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskB, taskD, 1) // B -> D
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskC, taskD, 1) // C -> D

		// Get tree
		nodes, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		// Should have 1 root (Task A)
		require.Len(t, nodes, 1)

		rootA := nodes[0]
		assert.Equal(t, taskA, rootA.ID)

		// A should have 2 children (B and C)
		require.Len(t, rootA.Children, 2)

		// Check children are B and C (order may vary)
		childIDs := map[int]bool{rootA.Children[0].ID: true, rootA.Children[1].ID: true}
		assert.True(t, childIDs[taskB] && childIDs[taskC])

		// Both B and C should have D as child
		for _, child := range rootA.Children {
			require.Len(t, child.Children, 1)
			assert.Equal(t, taskD, child.Children[0].ID)
		}
	})
}

// TestGetTaskTreeByProject_SelfDependency tests self-referencing task (A -> A)
func TestGetTaskTreeByProject_SelfDependency(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskA := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task A")

		// Create self-dependency
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskA, taskA, 1) // A -> A

		// Should handle gracefully (not panic or hang)
		nodes, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		// With self-dependency, the task is marked as having a parent (itself)
		// So it won't appear as a root. This is a quirk of how circular dependencies are handled.
		// The important thing is that it doesn't panic or infinite loop.
		// Traverse the entire tree to ensure no infinite loops
		nodeCount := 0
		var traverse func(*models.TaskTreeNode)
		traverse = func(node *models.TaskTreeNode) {
			nodeCount++
			require.LessOrEqual(t, nodeCount, 100, "Tree traversal exceeded 100 nodes - likely infinite loop")
			for _, child := range node.Children {
				traverse(child)
			}
		}

		for _, root := range nodes {
			traverse(root)
		}

		// Even if task is not a root, total node count should be <= 1
		assert.LessOrEqual(t, nodeCount, 1)
	})
}

// TestGetTaskTreeByProject_BlockingRelationships tests tree with blocking relationships
func TestGetTaskTreeByProject_BlockingRelationships(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create tasks with blocking relationships
		taskA := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task A (Blocker)")
		taskB := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task B (Blocked)")

		// A blocks B (relation type 2)
		// Note: The tree building logic treats ALL relationships as parent-child hierarchies
		// So one will be a root and the other will be a child
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskB, taskA, 2) // B is blocked by A (A is parent)

		// Get tree
		nodes, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		// Should have exactly 1 root (the parent in the relationship)
		require.Len(t, nodes, 1)

		root := nodes[0]

		// Root should have 1 child (the child in the relationship)
		require.Len(t, root.Children, 1)

		child := root.Children[0]

		// Verify it's the A->B relationship with blocking flag
		taskIDs := map[int]bool{root.ID: true, child.ID: true}
		assert.True(t, taskIDs[taskA] && taskIDs[taskB])

		// The relation should be marked as blocking
		assert.True(t, child.IsBlocking)
	})
}

// TestGetTaskTreeByProject_MixedRelationships tests tree with both parent-child and blocking
func TestGetTaskTreeByProject_MixedRelationships(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		// Create 4 tasks
		taskA := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task A")
		taskB := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task B")
		taskC := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task C")
		taskD := fixtures.CreateTestTask(t, env.DB, env.Dialect, env.ColumnID, "Task D")

		// A -> B (parent-child), B -> C (parent-child), C blocks D
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskA, taskB, 1) // A -> B (parent-child)
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskB, taskC, 1) // B -> C (parent-child)
		fixtures.AddTaskSubtask(t, env.DB, env.Dialect, taskD, taskC, 2) // C is blocked by D

		// Get tree
		nodes, err := env.Svc.GetTaskTreeByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		// Should have 2 roots: A and D
		require.Len(t, nodes, 2)

		// Find root A
		var rootA *models.TaskTreeNode
		for _, node := range nodes {
			if node.ID == taskA {
				rootA = node
				break
			}
		}

		require.NotNil(t, rootA)

		// Verify A -> B -> C hierarchy
		require.Len(t, rootA.Children, 1)

		childB := rootA.Children[0]
		assert.Equal(t, taskB, childB.ID)

		require.Len(t, childB.Children, 1)

		childC := childB.Children[0]
		assert.Equal(t, taskC, childC.ID)
	})
}
