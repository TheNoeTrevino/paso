package database

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"
)

// TestMoveTaskBetweenColumns tests moving tasks using the linked list functions
func TestMoveTaskBetweenColumns(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create columns
	col1, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)
	col2, _ := CreateColumn(context.Background(), db, "In Progress", 1, nil)
	col3, _ := CreateColumn(context.Background(), db, "Done", 1, nil)

	// Create a task in the first column
	task, err := CreateTask(context.Background(), db, "Test Task", "Description", col1.ID, 0)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Move task to next column (col1 -> col2)
	err = MoveTaskToNextColumn(context.Background(), db, task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to next column: %v", err)
	}

	// Verify task is now in col2
	tasks, err := GetTasksByColumn(context.Background(), db, col2.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be in second column")
	}

	// Verify task is not in col1
	tasks, err = GetTasksByColumn(context.Background(), db, col1.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Error("Task should not be in first column")
	}

	// Move task to next column (col2 -> col3)
	err = MoveTaskToNextColumn(context.Background(), db, task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to third column: %v", err)
	}

	// Verify task is now in col3
	tasks, err = GetTasksByColumn(context.Background(), db, col3.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be in third column")
	}

	// Try to move beyond last column (should fail)
	err = MoveTaskToNextColumn(context.Background(), db, task.ID)
	if err == nil {
		t.Error("Should not be able to move task beyond last column")
	}

	// Move task back (col3 -> col2)
	err = MoveTaskToPrevColumn(context.Background(), db, task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to previous column: %v", err)
	}

	// Verify task is back in col2
	tasks, err = GetTasksByColumn(context.Background(), db, col2.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be back in second column")
	}

	// Move task back to col1
	err = MoveTaskToPrevColumn(context.Background(), db, task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to first column: %v", err)
	}

	// Try to move before first column (should fail)
	err = MoveTaskToPrevColumn(context.Background(), db, task.ID)
	if err == nil {
		t.Error("Should not be able to move task before first column")
	}
}

// TestTaskCreationPersistence tests that tasks are properly saved to the database
func TestTaskCreationPersistence(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a column first
	col, err := CreateColumn(context.Background(), db, "Todo", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Create a task with title and description
	task, err := CreateTask(context.Background(), db, "Test Task Title", "This is a test description", col.ID, 0)
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
	tasks, err := GetTasksByColumn(context.Background(), db, col.ID)
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
	defer db.Close()

	// Create column and task
	col, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)
	task, _ := CreateTask(context.Background(), db, "Original Title", "Original Description", col.ID, 0)

	// Update the task
	err := UpdateTask(context.Background(), db, task.ID, "Updated Title", "Updated Description")
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Retrieve and verify the update persisted
	detail, err := GetTaskDetail(context.Background(), db, task.ID)
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
	defer db.Close()

	// Create column, task with description, and labels
	col, _ := CreateColumn(context.Background(), db, "Todo", 1, nil)
	task, _ := CreateTask(context.Background(), db, "Full Task", "A complete description with details", col.ID, 0)
	label, _ := CreateLabel(context.Background(), db, 1, "Important", "#FFD700")
	SetTaskLabels(context.Background(), db, task.ID, []int{label.ID})

	// Get full detail
	detail, err := GetTaskDetail(context.Background(), db, task.ID)
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
