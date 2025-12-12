package database

import (
	"context"
	"log"
	"testing"

	_ "modernc.org/sqlite"
)

// TestMoveTaskBetweenColumns tests moving tasks using the linked list functions
func TestMoveTaskBetweenColumns(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create columns
	col1, _ := repo.CreateColumn(context.Background(), "Todo", 1, nil)
	col2, _ := repo.CreateColumn(context.Background(), "In Progress", 1, nil)
	col3, _ := repo.CreateColumn(context.Background(), "Done", 1, nil)

	// Create a task in the first column
	task, err := repo.CreateTask(context.Background(), "Test Task", "Description", col1.ID, 0)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Move task to next column (col1 -> col2)
	err = repo.MoveTaskToNextColumn(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to next column: %v", err)
	}

	// Verify task is now in col2
	tasks, err := repo.GetTasksByColumn(context.Background(), col2.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be in second column")
	}

	// Verify task is not in col1
	tasks, err = repo.GetTasksByColumn(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Error("Task should not be in first column")
	}

	// Move task to next column (col2 -> col3)
	err = repo.MoveTaskToNextColumn(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to third column: %v", err)
	}

	// Verify task is now in col3
	tasks, err = repo.GetTasksByColumn(context.Background(), col3.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be in third column")
	}

	// Try to move beyond last column (should fail)
	err = repo.MoveTaskToNextColumn(context.Background(), task.ID)
	if err == nil {
		t.Error("Should not be able to move task beyond last column")
	}

	// Move task back (col3 -> col2)
	err = repo.MoveTaskToPrevColumn(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to previous column: %v", err)
	}

	// Verify task is back in col2
	tasks, err = repo.GetTasksByColumn(context.Background(), col2.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be back in second column")
	}

	// Move task back to col1
	err = repo.MoveTaskToPrevColumn(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to first column: %v", err)
	}

	// Try to move before first column (should fail)
	err = repo.MoveTaskToPrevColumn(context.Background(), task.ID)
	if err == nil {
		t.Error("Should not be able to move task before first column")
	}
}

// TestTaskCreationPersistence tests that tasks are properly saved to the database
func TestTaskCreationPersistence(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create a column first
	col, err := repo.CreateColumn(context.Background(), "Todo", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Create a task with title and description
	task, err := repo.CreateTask(context.Background(), "Test Task Title", "This is a test description", col.ID, 0)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Verify task was created with correct data
	if task.ID == 0 {
		t.Error("Task should have a valid ID")
	}
	if task.Title != "Test Task Title" {
		t.Errorf("Expected title 'Test Task Title', got '%s'", task.Title)
	}
	if task.Description != "This is a test description" {
		t.Errorf("Expected description 'This is a test description', got '%s'", task.Description)
	}
	if task.ColumnID != col.ID {
		t.Errorf("Expected column ID %d, got %d", col.ID, task.ColumnID)
	}

	// Verify task can be retrieved from database
	tasks, err := repo.GetTasksByColumn(context.Background(), col.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Test Task Title" {
		t.Errorf("Retrieved task has wrong title: %s", tasks[0].Title)
	}
	if tasks[0].Description != "This is a test description" {
		t.Errorf("Retrieved task has wrong description: %s", tasks[0].Description)
	}
}

// TestTaskUpdatePersistence tests that task updates are properly saved
func TestTaskUpdatePersistence(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create column and task
	col, _ := repo.CreateColumn(context.Background(), "Todo", 1, nil)
	task, _ := repo.CreateTask(context.Background(), "Original Title", "Original Description", col.ID, 0)

	// Update the task
	err := repo.UpdateTask(context.Background(), task.ID, "Updated Title", "Updated Description")
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Retrieve and verify the update persisted
	detail, err := repo.GetTaskDetail(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to get task detail: %v", err)
	}
	if detail.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", detail.Title)
	}
	if detail.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got '%s'", detail.Description)
	}
}

// TestTaskDetailIncludesAllFields tests that GetTaskDetail returns all fields
func TestTaskDetailIncludesAllFields(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create column, task with description, and labels
	col, _ := repo.CreateColumn(context.Background(), "Todo", 1, nil)
	task, _ := repo.CreateTask(context.Background(), "Full Task", "A complete description with details", col.ID, 0)
	label, _ := repo.CreateLabel(context.Background(), 1, "Important", "#FFD700")
	if err := repo.SetTaskLabels(context.Background(), task.ID, []int{label.ID}); err != nil {
		t.Fatalf("Failed to set task labels: %v", err)
	}

	// Get full detail
	detail, err := repo.GetTaskDetail(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to get task detail: %v", err)
	}

	// Verify all fields
	if detail.ID != task.ID {
		t.Errorf("Wrong ID: expected %d, got %d", task.ID, detail.ID)
	}
	if detail.Title != "Full Task" {
		t.Errorf("Wrong title: %s", detail.Title)
	}
	if detail.Description != "A complete description with details" {
		t.Errorf("Wrong description: %s", detail.Description)
	}
	if detail.ColumnID != col.ID {
		t.Errorf("Wrong column ID: expected %d, got %d", col.ID, detail.ColumnID)
	}
	if len(detail.Labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(detail.Labels))
	}
	if detail.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if detail.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

// TestSwapTaskUp tests moving tasks up within a column
func TestSwapTaskUp(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create a column
	col, _ := repo.CreateColumn(context.Background(), "Todo", 1, nil)

	// Create three tasks
	task1, _ := repo.CreateTask(context.Background(), "Task 1", "", col.ID, 0)
	task2, _ := repo.CreateTask(context.Background(), "Task 2", "", col.ID, 1)
	task3, _ := repo.CreateTask(context.Background(), "Task 3", "", col.ID, 2)

	// Move task2 up (should swap with task1)
	err := repo.SwapTaskUp(context.Background(), task2.ID)
	if err != nil {
		t.Fatalf("Failed to swap task up: %v", err)
	}

	// Verify new order: task2, task1, task3
	tasks, _ := repo.GetTasksByColumn(context.Background(), col.ID)
	if len(tasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != task2.ID {
		t.Errorf("Expected task2 at position 0, got task %d", tasks[0].ID)
	}
	if tasks[1].ID != task1.ID {
		t.Errorf("Expected task1 at position 1, got task %d", tasks[1].ID)
	}
	if tasks[2].ID != task3.ID {
		t.Errorf("Expected task3 at position 2, got task %d", tasks[2].ID)
	}

	// Verify positions are correct
	if tasks[0].Position != 0 {
		t.Errorf("Task at index 0 should have position 0, got %d", tasks[0].Position)
	}
	if tasks[1].Position != 1 {
		t.Errorf("Task at index 1 should have position 1, got %d", tasks[1].Position)
	}
	if tasks[2].Position != 2 {
		t.Errorf("Task at index 2 should have position 2, got %d", tasks[2].Position)
	}

	// Try to move task2 up again (now at top, should fail)
	err = repo.SwapTaskUp(context.Background(), task2.ID)
	if err == nil {
		t.Error("Expected error when swapping up from top position")
	}
}

// TestSwapTaskDown tests moving tasks down within a column
func TestSwapTaskDown(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create a column
	col, _ := repo.CreateColumn(context.Background(), "Todo", 1, nil)

	// Create three tasks
	task1, _ := repo.CreateTask(context.Background(), "Task 1", "", col.ID, 0)
	task2, _ := repo.CreateTask(context.Background(), "Task 2", "", col.ID, 1)
	task3, _ := repo.CreateTask(context.Background(), "Task 3", "", col.ID, 2)

	// Move task2 down (should swap with task3)
	err := repo.SwapTaskDown(context.Background(), task2.ID)
	if err != nil {
		t.Fatalf("Failed to swap task down: %v", err)
	}

	// Verify new order: task1, task3, task2
	tasks, _ := repo.GetTasksByColumn(context.Background(), col.ID)
	if len(tasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != task1.ID {
		t.Errorf("Expected task1 at position 0, got task %d", tasks[0].ID)
	}
	if tasks[1].ID != task3.ID {
		t.Errorf("Expected task3 at position 1, got task %d", tasks[1].ID)
	}
	if tasks[2].ID != task2.ID {
		t.Errorf("Expected task2 at position 2, got task %d", tasks[2].ID)
	}

	// Verify positions are correct
	if tasks[0].Position != 0 {
		t.Errorf("Task at index 0 should have position 0, got %d", tasks[0].Position)
	}
	if tasks[1].Position != 1 {
		t.Errorf("Task at index 1 should have position 1, got %d", tasks[1].Position)
	}
	if tasks[2].Position != 2 {
		t.Errorf("Task at index 2 should have position 2, got %d", tasks[2].Position)
	}

	// Try to move task2 down again (now at bottom, should fail)
	err = repo.SwapTaskDown(context.Background(), task2.ID)
	if err == nil {
		t.Error("Expected error when swapping down from bottom position")
	}
}

// TestSwapTaskUpAndDown tests multiple swap operations
func TestSwapTaskUpAndDown(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create a column
	col, _ := repo.CreateColumn(context.Background(), "Todo", 1, nil)

	// Create four tasks
	task1, _ := repo.CreateTask(context.Background(), "Task 1", "", col.ID, 0)
	task2, _ := repo.CreateTask(context.Background(), "Task 2", "", col.ID, 1)
	task3, _ := repo.CreateTask(context.Background(), "Task 3", "", col.ID, 2)
	task4, _ := repo.CreateTask(context.Background(), "Task 4", "", col.ID, 3)

	// Move task3 up twice (should end up at position 0)
	if err := repo.SwapTaskUp(context.Background(), task3.ID); err != nil {
		t.Fatalf("Failed to swap task up: %v", err)
	}
	if err := repo.SwapTaskUp(context.Background(), task3.ID); err != nil {
		t.Fatalf("Failed to swap task up second time: %v", err)
	}

	// Verify order: task3, task1, task2, task4
	tasks, _ := repo.GetTasksByColumn(context.Background(), col.ID)
	if tasks[0].ID != task3.ID || tasks[1].ID != task1.ID || tasks[2].ID != task2.ID || tasks[3].ID != task4.ID {
		t.Error("Tasks not in expected order after moving up")
	}

	// Move task2 down (should swap with task4)
	if err := repo.SwapTaskDown(context.Background(), task2.ID); err != nil {
		t.Fatalf("Failed to swap task down: %v", err)
	}

	// Verify order: task3, task1, task4, task2
	tasks, _ = repo.GetTasksByColumn(context.Background(), col.ID)
	if tasks[0].ID != task3.ID || tasks[1].ID != task1.ID || tasks[2].ID != task4.ID || tasks[3].ID != task2.ID {
		t.Error("Tasks not in expected order after moving down")
	}
}
