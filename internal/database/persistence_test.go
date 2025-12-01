package database

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// Test 1: Task CRUD Persistence
func TestTaskCRUDPersistence(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	defer db.Close()

	// Create a column (using default project ID 1)
	col, err := CreateColumn(context.Background(), db, "Todo", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Create a task
	task, err := CreateTask(context.Background(), db, "Test task", "Test description", col.ID, 0)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Verify task exists in database
	tasks, err := GetTasksByColumn(context.Background(), db, col.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Title != "Test task" {
		t.Errorf("Expected title 'Test task', got '%s'", tasks[0].Title)
	}

	if tasks[0].Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", tasks[0].Description)
	}

	// Update task title
	if err := UpdateTaskTitle(context.Background(), db, task.ID, "Updated task"); err != nil {
		t.Fatalf("Failed to update task title: %v", err)
	}

	// Verify update persisted
	tasks, err = GetTasksByColumn(context.Background(), db, col.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks after update: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task after update, got %d", len(tasks))
	}

	if tasks[0].Title != "Updated task" {
		t.Errorf("Expected updated title 'Updated task', got '%s'", tasks[0].Title)
	}

	// Verify updated_at timestamp was updated
	if tasks[0].UpdatedAt.Before(tasks[0].CreatedAt) {
		t.Error("UpdatedAt should be >= CreatedAt")
	}

	// Delete task
	if err := DeleteTask(context.Background(), db, task.ID); err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify task no longer exists
	tasks, err = GetTasksByColumn(context.Background(), db, col.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks after delete: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after delete, got %d", len(tasks))
	}
}

// Test 2: Column CRUD Persistence with Linked List
func TestColumnCRUDPersistence(t *testing.T) {
	db, dbPath := setupTestDBFile(t)
	defer os.Remove(dbPath)

	// Create 3 columns (using default project ID 1)
	col1, err := CreateColumn(context.Background(), db, "Todo", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	col2, err := CreateColumn(context.Background(), db, "In Progress", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	col3, err := CreateColumn(context.Background(), db, "Done", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column 3: %v", err)
	}

	// Verify linked list pointers are correct
	verifyLinkedListIntegrity(t, db)

	columns, err := GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 3 {
		t.Fatalf("Expected 3 columns, got %d", len(columns))
	}

	// Verify order
	if columns[0].ID != col1.ID || columns[1].ID != col2.ID || columns[2].ID != col3.ID {
		t.Error("Columns not in expected order")
	}

	// Close and reopen database
	db = closeAndReopenDB(t, db, dbPath)
	defer db.Close()

	// Verify columns reload with correct order
	columns, err = GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns after reload: %v", err)
	}

	if len(columns) != 3 {
		t.Fatalf("Expected 3 columns after reload, got %d", len(columns))
	}

	if columns[0].ID != col1.ID || columns[1].ID != col2.ID || columns[2].ID != col3.ID {
		t.Error("Columns not in expected order after reload")
	}

	verifyLinkedListIntegrity(t, db)

	// Delete middle column
	if err := DeleteColumn(context.Background(), db, col2.ID); err != nil {
		t.Fatalf("Failed to delete column: %v", err)
	}

	// Verify pointers updated correctly
	columns, err = GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns after delete: %v", err)
	}

	if len(columns) != 2 {
		t.Fatalf("Expected 2 columns after delete, got %d", len(columns))
	}

	verifyLinkedListIntegrity(t, db)

	// Close and reopen
	db = closeAndReopenDB(t, db, dbPath)
	defer db.Close()

	// Verify deletion persisted
	columns, err = GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns after reload: %v", err)
	}

	if len(columns) != 2 {
		t.Fatalf("Expected 2 columns after reload, got %d", len(columns))
	}

	if columns[0].ID != col1.ID || columns[1].ID != col3.ID {
		t.Error("Remaining columns not as expected")
	}

	verifyLinkedListIntegrity(t, db)
}

// Test 3: Task Movement Persistence
func TestTaskMovementPersistence(t *testing.T) {
	db, dbPath := setupTestDBFile(t)
	defer os.Remove(dbPath)

	// Create database with 3 columns (using default project ID 1)
	col1, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)
	col2, _ := CreateColumn(context.Background(), db, "In Progress", 1, nil)
	col3, _ := CreateColumn(context.Background(), db, "Done", 1, nil)

	// Create task in first column
	task, err := CreateTask(context.Background(), db, "Test task", "Description", col1.ID, 0)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Move task to next column
	if err := MoveTaskToNextColumn(context.Background(), db, task.ID); err != nil {
		t.Fatalf("Failed to move task: %v", err)
	}

	// Verify task moved in database
	tasks, err := GetTasksByColumn(context.Background(), db, col2.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be in second column")
	}

	// Close and reopen database
	db = closeAndReopenDB(t, db, dbPath)
	defer db.Close()

	// Verify task is still in correct column
	tasks, err = GetTasksByColumn(context.Background(), db, col2.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks after reload: %v", err)
	}

	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should still be in second column after reload")
	}

	// Move task to previous column
	if err := MoveTaskToPrevColumn(context.Background(), db, task.ID); err != nil {
		t.Fatalf("Failed to move task back: %v", err)
	}

	// Verify movement persisted
	tasks, err = GetTasksByColumn(context.Background(), db, col1.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be back in first column")
	}

	// Verify not in col2 or col3
	tasks, _ = GetTasksByColumn(context.Background(), db, col2.ID)
	if len(tasks) != 0 {
		t.Error("Task should not be in second column")
	}

	tasks, _ = GetTasksByColumn(context.Background(), db, col3.ID)
	if len(tasks) != 0 {
		t.Error("Task should not be in third column")
	}
}

// Test 4: Column Insertion Persistence
func TestColumnInsertionPersistence(t *testing.T) {
	db, dbPath := setupTestDBFile(t)
	defer os.Remove(dbPath)

	// Create database with 2 columns (A, B) using default project ID 1
	colA, _ := CreateColumn(context.Background(), db, "Column A", 1, nil)
	colB, _ := CreateColumn(context.Background(), db, "Column B", 1, nil)

	// Insert new column after A (creating A, C, B)
	colC, err := CreateColumn(context.Background(), db, "Column C", 1, &colA.ID)
	if err != nil {
		t.Fatalf("Failed to insert column: %v", err)
	}

	// Verify linked list: A.next=C, C.prev=A, C.next=B, B.prev=C
	columns, err := GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 3 {
		t.Fatalf("Expected 3 columns, got %d", len(columns))
	}

	// Verify order is A, C, B
	if columns[0].Name != "Column A" || columns[1].Name != "Column C" || columns[2].Name != "Column B" {
		t.Errorf("Expected order A, C, B, got %s, %s, %s", columns[0].Name, columns[1].Name, columns[2].Name)
	}

	// Verify A.next = C
	if columns[0].NextID == nil || *columns[0].NextID != colC.ID {
		t.Error("Column A's NextID should point to Column C")
	}

	// Verify C.prev = A
	if columns[1].PrevID == nil || *columns[1].PrevID != colA.ID {
		t.Error("Column C's PrevID should point to Column A")
	}

	// Verify C.next = B
	if columns[1].NextID == nil || *columns[1].NextID != colB.ID {
		t.Error("Column C's NextID should point to Column B")
	}

	// Verify B.prev = C
	if columns[2].PrevID == nil || *columns[2].PrevID != colC.ID {
		t.Error("Column B's PrevID should point to Column C")
	}

	verifyLinkedListIntegrity(t, db)

	// Close and reopen database
	db = closeAndReopenDB(t, db, dbPath)
	defer db.Close()

	// Verify order is still A, C, B
	columns, err = GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns after reload: %v", err)
	}

	if len(columns) != 3 {
		t.Fatalf("Expected 3 columns after reload, got %d", len(columns))
	}

	if columns[0].Name != "Column A" || columns[1].Name != "Column C" || columns[2].Name != "Column B" {
		t.Errorf("Expected order A, C, B after reload, got %s, %s, %s", columns[0].Name, columns[1].Name, columns[2].Name)
	}

	// Verify pointers are correct
	verifyLinkedListIntegrity(t, db)
}

// Test 5: Cascade Deletion
func TestCascadeDeletion(t *testing.T) {
	db, dbPath := setupTestDBFile(t)
	defer os.Remove(dbPath)

	// Create database with 1 column (using default project ID 1)
	col, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)

	// Create 5 tasks in column
	for i := 0; i < 5; i++ {
		_, err := CreateTask(context.Background(), db, "Task", "Description", col.ID, i)
		if err != nil {
			t.Fatalf("Failed to create task %d: %v", i, err)
		}
	}

	// Verify tasks exist
	tasks, _ := GetTasksByColumn(context.Background(), db, col.ID)
	if len(tasks) != 5 {
		t.Fatalf("Expected 5 tasks, got %d", len(tasks))
	}

	// Delete column
	if err := DeleteColumn(context.Background(), db, col.ID); err != nil {
		t.Fatalf("Failed to delete column: %v", err)
	}

	// Verify all tasks are deleted
	tasks, _ = GetTasksByColumn(context.Background(), db, col.ID)
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after column deletion, got %d", len(tasks))
	}

	// Verify column doesn't exist
	columns, _ := GetAllColumns(context.Background(), db)
	if len(columns) != 0 {
		t.Errorf("Expected 0 columns, got %d", len(columns))
	}

	// Close and reopen database
	db = closeAndReopenDB(t, db, dbPath)
	defer db.Close()

	// Verify column and tasks don't exist
	columns, _ = GetAllColumns(context.Background(), db)
	if len(columns) != 0 {
		t.Errorf("Expected 0 columns after reload, got %d", len(columns))
	}

	// Try to query tasks (should return empty)
	tasks, _ = GetTasksByColumn(context.Background(), db, col.ID)
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after reload, got %d", len(tasks))
	}
}

// Test 6: Transaction Rollback on Error
func TestTransactionRollback(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	defer db.Close()

	// Create 2 columns (using default project ID 1)
	col1, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)
	col2, _ := CreateColumn(context.Background(), db, "Done", 1, nil)

	// Attempt to delete column with invalid ID
	err := DeleteColumn(context.Background(), db, 99999)
	if err == nil {
		t.Error("Expected error when deleting non-existent column")
	}

	// Verify transaction rolled back (no changes)
	columns, _ := GetAllColumns(context.Background(), db)
	if len(columns) != 2 {
		t.Errorf("Expected 2 columns, got %d (transaction should have rolled back)", len(columns))
	}

	// Attempt to create task with invalid column ID
	_, err = CreateTask(context.Background(), db, "Test", "Description", 99999, 0)
	if err == nil {
		t.Error("Expected error when creating task with invalid column ID")
	}

	// Verify no tasks were created
	count1, _ := GetTaskCountByColumn(context.Background(), db, col1.ID)
	count2, _ := GetTaskCountByColumn(context.Background(), db, col2.ID)
	if count1 != 0 || count2 != 0 {
		t.Error("No tasks should have been created")
	}
}

// Test 7: Sequential Bulk Operations
// Note: Tests creating many tasks rapidly (realistic for TUI event-driven app)
func TestSequentialBulkOperations(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	defer db.Close()

	// Create a column (using default project ID 1)
	col, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)

	// Create many tasks in sequence (like rapid user input)
	numTasks := 50
	for i := 0; i < numTasks; i++ {
		_, err := CreateTask(context.Background(), db, "Task", "Description", col.ID, i)
		if err != nil {
			t.Fatalf("Failed to create task %d: %v", i, err)
		}
	}

	// Verify all tasks were created
	tasks, err := GetTasksByColumn(context.Background(), db, col.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	if len(tasks) != numTasks {
		t.Errorf("Expected %d tasks, got %d", numTasks, len(tasks))
	}

	// Verify no data corruption (all tasks have valid unique IDs)
	taskIDs := make(map[int]bool)
	for _, task := range tasks {
		if task.ID == 0 {
			t.Error("Found task with ID 0 (data corruption)")
		}
		if taskIDs[task.ID] {
			t.Errorf("Found duplicate task ID %d (data corruption)", task.ID)
		}
		taskIDs[task.ID] = true
	}

	// Test rapid updates
	for i := 0; i < 10; i++ {
		err := UpdateTaskTitle(context.Background(), db, tasks[i].ID, "Updated")
		if err != nil {
			t.Fatalf("Failed to update task %d: %v", i, err)
		}
	}

	// Test rapid deletions
	for i := 10; i < 20; i++ {
		err := DeleteTask(context.Background(), db, tasks[i].ID)
		if err != nil {
			t.Fatalf("Failed to delete task %d: %v", i, err)
		}
	}

	// Verify final count
	tasks, _ = GetTasksByColumn(context.Background(), db, col.ID)
	expectedCount := numTasks - 10 // deleted 10 tasks
	if len(tasks) != expectedCount {
		t.Errorf("Expected %d tasks after deletions, got %d", expectedCount, len(tasks))
	}
}

// Test 8: Reload Full State
func TestReloadFullState(t *testing.T) {
	db, dbPath := setupTestDBFile(t)
	defer os.Remove(dbPath)

	// Create database with complex state:
	// - 5 columns with linked list
	columnNames := []string{"Backlog", "Todo", "In Progress", "Review", "Done"}
	var columnIDs []int

	for _, name := range columnNames {
		col, err := CreateColumn(context.Background(), db, name, 1, nil)
		if err != nil {
			t.Fatalf("Failed to create column %s: %v", name, err)
		}
		columnIDs = append(columnIDs, col.ID)
	}

	// - 20 tasks across columns (4 per column)
	taskCount := 0
	for colIdx, colID := range columnIDs {
		for i := 0; i < 4; i++ {
			_, err := CreateTask(context.Background(), db, "Task", "Description", colID, i)
			if err != nil {
				t.Fatalf("Failed to create task in column %d: %v", colIdx, err)
			}
			taskCount++
		}
	}

	// Verify linked list integrity before close
	verifyLinkedListIntegrity(t, db)

	// Close database
	db = closeAndReopenDB(t, db, dbPath)
	defer db.Close()

	// Load all columns (verify order)
	columns, err := GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns after reload: %v", err)
	}

	if len(columns) != len(columnNames) {
		t.Fatalf("Expected %d columns, got %d", len(columnNames), len(columns))
	}

	for i, col := range columns {
		if col.Name != columnNames[i] {
			t.Errorf("Column %d: expected name %s, got %s", i, columnNames[i], col.Name)
		}
	}

	// Load all tasks (verify correct columns)
	totalTasks := 0
	for _, colID := range columnIDs {
		tasks, err := GetTasksByColumn(context.Background(), db, colID)
		if err != nil {
			t.Fatalf("Failed to get tasks for column %d: %v", colID, err)
		}

		if len(tasks) != 4 {
			t.Errorf("Expected 4 tasks in column %d, got %d", colID, len(tasks))
		}

		totalTasks += len(tasks)

		// Verify all tasks have correct column_id
		for _, task := range tasks {
			if task.ColumnID != colID {
				t.Errorf("Task %d has wrong column_id: expected %d, got %d", task.ID, colID, task.ColumnID)
			}
		}
	}

	if totalTasks != taskCount {
		t.Errorf("Expected %d total tasks, got %d", taskCount, totalTasks)
	}

	// Verify linked list integrity after reload
	verifyLinkedListIntegrity(t, db)

	// Verify foreign keys intact by trying to create task with invalid column
	_, err = CreateTask(context.Background(), db, "Invalid", "Description", 99999, 0)
	if err == nil {
		t.Error("Should not be able to create task with invalid column_id")
	}
}

// Test 9: Migration Idempotency
func TestMigrationIdempotency(t *testing.T) {
	db, dbPath := setupTestDBFile(t)
	defer os.Remove(dbPath)

	// Create some data (using default project ID 1)
	col, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)
	_, _ = CreateTask(context.Background(), db, "Task", "Description", col.ID, 0)

	// Run migrations again
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations second time: %v", err)
	}

	// Verify no errors and schema is correct
	columns, err := GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 1 {
		t.Errorf("Expected 1 column, got %d", len(columns))
	}

	tasks, _ := GetTasksByColumn(context.Background(), db, col.ID)
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Run migrations a third time
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations third time: %v", err)
	}

	// Verify still works correctly
	columns, err = GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 1 {
		t.Errorf("Expected 1 column after third migration, got %d", len(columns))
	}
}

// Test 10: Empty Database Reload
func TestEmptyDatabaseReload(t *testing.T) {
	// Create a fresh database file WITHOUT clearing the default columns
	tmpfile, err := os.CreateTemp("", "paso-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	dbPath := tmpfile.Name()
	defer os.Remove(dbPath)

	// First open: initialize with migrations
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Close database immediately (fresh database with migrations)
	db.Close()

	// Reopen database
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db.Close()

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Verify default columns exist (from initial migration)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM columns").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query columns: %v", err)
	}

	// Should have 3 default columns (Todo, In Progress, Done)
	if count != 3 {
		t.Errorf("Expected 3 default columns, got %d", count)
	}

	// Get all columns and verify linked list
	columns, err := GetAllColumns(context.Background(), db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 3 {
		t.Fatalf("Expected 3 columns, got %d", len(columns))
	}

	// Verify linked list is correct
	verifyLinkedListIntegrity(t, db)

	// Verify column names are defaults
	expectedNames := map[string]bool{
		"Todo":        true,
		"In Progress": true,
		"Done":        true,
	}

	for _, col := range columns {
		if !expectedNames[col.Name] {
			t.Errorf("Unexpected column name: %s", col.Name)
		}
	}
}

// Test 11: Timestamps Persistence
func TestTimestampsPersistence(t *testing.T) {
	db, dbPath := setupTestDBFile(t)
	defer os.Remove(dbPath)

	// Create column and task (using default project ID 1)
	col, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)
	task, err := CreateTask(context.Background(), db, "Test task", "Description", col.ID, 0)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	originalCreatedAt := task.CreatedAt
	originalUpdatedAt := task.UpdatedAt

	// Wait to ensure timestamp difference (SQLite has second precision)
	time.Sleep(1100 * time.Millisecond)

	// Update task
	if err := UpdateTaskTitle(context.Background(), db, task.ID, "Updated task"); err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Reload task
	tasks, _ := GetTasksByColumn(context.Background(), db, col.ID)
	updatedTask := tasks[0]

	// Verify created_at hasn't changed
	if !updatedTask.CreatedAt.Equal(originalCreatedAt) {
		t.Error("CreatedAt should not change on update")
	}

	// Verify updated_at has changed or stayed the same (SQLite has second precision)
	if updatedTask.UpdatedAt.Before(originalUpdatedAt) {
		t.Error("UpdatedAt should not be before original timestamp")
	}

	// Close and reopen database
	db = closeAndReopenDB(t, db, dbPath)
	defer db.Close()

	// Verify timestamps persisted correctly
	tasks, _ = GetTasksByColumn(context.Background(), db, col.ID)
	if len(tasks) == 0 {
		t.Fatal("Task not found after reload")
	}

	reloadedTask := tasks[0]

	// Verify timestamps match
	if !reloadedTask.CreatedAt.Equal(originalCreatedAt) {
		t.Error("CreatedAt not persisted correctly")
	}

	if !reloadedTask.UpdatedAt.Equal(updatedTask.UpdatedAt) {
		t.Error("UpdatedAt not persisted correctly")
	}
}

// Test 12: Complex Movement Sequence Persistence
func TestComplexMovementSequencePersistence(t *testing.T) {
	db, dbPath := setupTestDBFile(t)
	defer os.Remove(dbPath)

	// Create 4 columns (using default project ID 1)
	col1, _ := CreateColumn(context.Background(), db, "Col1", 1, nil)
	col2, _ := CreateColumn(context.Background(), db, "Col2", 1, nil)
	col3, _ := CreateColumn(context.Background(), db, "Col3", 1, nil)
	col4, _ := CreateColumn(context.Background(), db, "Col4", 1, nil)

	// Create 3 tasks in col1
	task1, _ := CreateTask(context.Background(), db, "Task 1", "", col1.ID, 0)
	task2, _ := CreateTask(context.Background(), db, "Task 2", "", col1.ID, 1)
	task3, _ := CreateTask(context.Background(), db, "Task 3", "", col1.ID, 2)

	// Move tasks in complex pattern
	MoveTaskToNextColumn(context.Background(), db, task1.ID) // Task 1: Col1 -> Col2
	MoveTaskToNextColumn(context.Background(), db, task1.ID) // Task 1: Col2 -> Col3
	MoveTaskToNextColumn(context.Background(), db, task2.ID) // Task 2: Col1 -> Col2
	MoveTaskToPrevColumn(context.Background(), db, task1.ID) // Task 1: Col3 -> Col2
	MoveTaskToNextColumn(context.Background(), db, task3.ID) // Task 3: Col1 -> Col2

	// Expected state:
	// Col1: []
	// Col2: [Task 1, Task 2, Task 3]
	// Col3: []
	// Col4: []

	// Verify state before close
	tasks1, _ := GetTasksByColumn(context.Background(), db, col1.ID)
	tasks2, _ := GetTasksByColumn(context.Background(), db, col2.ID)
	tasks3, _ := GetTasksByColumn(context.Background(), db, col3.ID)
	tasks4, _ := GetTasksByColumn(context.Background(), db, col4.ID)

	if len(tasks1) != 0 || len(tasks2) != 3 || len(tasks3) != 0 || len(tasks4) != 0 {
		t.Errorf("Unexpected task distribution: col1=%d, col2=%d, col3=%d, col4=%d",
			len(tasks1), len(tasks2), len(tasks3), len(tasks4))
	}

	// Close and reopen
	db = closeAndReopenDB(t, db, dbPath)
	defer db.Close()

	// Verify state persisted
	tasks1, _ = GetTasksByColumn(context.Background(), db, col1.ID)
	tasks2, _ = GetTasksByColumn(context.Background(), db, col2.ID)
	tasks3, _ = GetTasksByColumn(context.Background(), db, col3.ID)
	tasks4, _ = GetTasksByColumn(context.Background(), db, col4.ID)

	if len(tasks1) != 0 || len(tasks2) != 3 || len(tasks3) != 0 || len(tasks4) != 0 {
		t.Errorf("Task distribution not persisted: col1=%d, col2=%d, col3=%d, col4=%d",
			len(tasks1), len(tasks2), len(tasks3), len(tasks4))
	}
}

// Test 13: Column Reordering Persistence
func TestColumnReorderingPersistence(t *testing.T) {
	db, dbPath := setupTestDBFile(t)
	defer os.Remove(dbPath)

	// Create 3 columns: A, B, C (using default project ID 1)
	colA, _ := CreateColumn(context.Background(), db, "A", 1, nil)
	_, _ = CreateColumn(context.Background(), db, "B", 1, nil)
	_, _ = CreateColumn(context.Background(), db, "C", 1, nil)

	// Insert D between A and B
	_, _ = CreateColumn(context.Background(), db, "D", 1, &colA.ID)

	// Expected order: A, D, B, C
	columns, _ := GetAllColumns(context.Background(), db)
	if len(columns) != 4 {
		t.Fatalf("Expected 4 columns, got %d", len(columns))
	}

	expectedOrder := []string{"A", "D", "B", "C"}
	for i, col := range columns {
		if col.Name != expectedOrder[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expectedOrder[i], col.Name)
		}
	}

	// Close and reopen
	db = closeAndReopenDB(t, db, dbPath)
	defer db.Close()

	// Verify order persisted
	columns, _ = GetAllColumns(context.Background(), db)
	if len(columns) != 4 {
		t.Fatalf("Expected 4 columns after reload, got %d", len(columns))
	}

	for i, col := range columns {
		if col.Name != expectedOrder[i] {
			t.Errorf("Position %d after reload: expected %s, got %s", i, expectedOrder[i], col.Name)
		}
	}

	verifyLinkedListIntegrity(t, db)
}

// Test 14: Update Task Column Directly
func TestUpdateTaskColumnDirectly(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	defer db.Close()

	// Create 2 columns (using default project ID 1)
	col1, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)
	col2, _ := CreateColumn(context.Background(), db, "Done", 1, nil)

	// Create task in col1
	task, _ := CreateTask(context.Background(), db, "Test", "Description", col1.ID, 0)

	// Move task using UpdateTaskColumn
	err := UpdateTaskColumn(context.Background(), db, task.ID, col2.ID, 0)
	if err != nil {
		t.Fatalf("Failed to update task column: %v", err)
	}

	// Verify task moved
	tasks1, _ := GetTasksByColumn(context.Background(), db, col1.ID)
	tasks2, _ := GetTasksByColumn(context.Background(), db, col2.ID)

	if len(tasks1) != 0 {
		t.Error("Task should not be in col1")
	}

	if len(tasks2) != 1 || tasks2[0].ID != task.ID {
		t.Error("Task should be in col2")
	}
}

// Test 15: Multiple Tasks in Column Order
func TestMultipleTasksInColumnOrder(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	defer db.Close()

	col, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)

	// Create tasks with specific positions
	task1, _ := CreateTask(context.Background(), db, "Task 1", "", col.ID, 0)
	task2, _ := CreateTask(context.Background(), db, "Task 2", "", col.ID, 1)
	task3, _ := CreateTask(context.Background(), db, "Task 3", "", col.ID, 2)

	// Retrieve tasks
	tasks, err := GetTasksByColumn(context.Background(), db, col.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(tasks))
	}

	// Verify order by position
	if tasks[0].ID != task1.ID || tasks[1].ID != task2.ID || tasks[2].ID != task3.ID {
		t.Error("Tasks not in expected order")
	}

	// Verify positions are correct
	if tasks[0].Position != 0 || tasks[1].Position != 1 || tasks[2].Position != 2 {
		t.Error("Task positions not as expected")
	}
}
