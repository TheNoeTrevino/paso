package database

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations to set up schema
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clear the default seeded columns and labels for fresh tests
	_, err = db.Exec("DELETE FROM columns")
	if err != nil {
		t.Fatalf("Failed to clear columns: %v", err)
	}
	_, err = db.Exec("DELETE FROM labels")
	if err != nil {
		t.Fatalf("Failed to clear labels: %v", err)
	}

	return db
}

// TestLinkedListTraversal tests that columns are created and traversed in correct order
func TestLinkedListTraversal(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create 3 columns
	col1, err := CreateColumn(db, "Todo", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	col2, err := CreateColumn(db, "In Progress", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	col3, err := CreateColumn(db, "Done", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column 3: %v", err)
	}

	// Retrieve all columns and verify order
	columns, err := GetAllColumns(db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}

	// Verify order and linked list structure
	if columns[0].ID != col1.ID || columns[0].Name != "Todo" {
		t.Errorf("First column should be Todo, got %s", columns[0].Name)
	}
	if columns[0].PrevID != nil {
		t.Error("First column should have nil PrevID")
	}
	if columns[0].NextID == nil || *columns[0].NextID != col2.ID {
		t.Error("First column's NextID should point to second column")
	}

	if columns[1].ID != col2.ID || columns[1].Name != "In Progress" {
		t.Errorf("Second column should be In Progress, got %s", columns[1].Name)
	}
	if columns[1].PrevID == nil || *columns[1].PrevID != col1.ID {
		t.Error("Second column's PrevID should point to first column")
	}
	if columns[1].NextID == nil || *columns[1].NextID != col3.ID {
		t.Error("Second column's NextID should point to third column")
	}

	if columns[2].ID != col3.ID || columns[2].Name != "Done" {
		t.Errorf("Third column should be Done, got %s", columns[2].Name)
	}
	if columns[2].PrevID == nil || *columns[2].PrevID != col2.ID {
		t.Error("Third column's PrevID should point to second column")
	}
	if columns[2].NextID != nil {
		t.Error("Third column should have nil NextID")
	}
}

// TestInsertColumnMiddle tests inserting a column in the middle of the list
func TestInsertColumnMiddle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create initial columns
	col1, _ := CreateColumn(db, "Todo", 1, nil)
	col3, _ := CreateColumn(db, "Done", 1, nil)

	// Insert a column in the middle (after col1)
	col2, err := CreateColumn(db, "In Progress", 1, &col1.ID)
	if err != nil {
		t.Fatalf("Failed to insert column in middle: %v", err)
	}

	// Verify the linked list structure
	columns, err := GetAllColumns(db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}

	// Verify order: Todo -> In Progress -> Done
	if columns[0].Name != "Todo" || columns[1].Name != "In Progress" || columns[2].Name != "Done" {
		t.Errorf("Column order incorrect: %s, %s, %s", columns[0].Name, columns[1].Name, columns[2].Name)
	}

	// Verify pointers
	if columns[0].NextID == nil || *columns[0].NextID != col2.ID {
		t.Error("First column should point to inserted column")
	}
	if columns[1].PrevID == nil || *columns[1].PrevID != col1.ID {
		t.Error("Inserted column should point back to first column")
	}
	if columns[1].NextID == nil || *columns[1].NextID != col3.ID {
		t.Error("Inserted column should point forward to third column")
	}
	if columns[2].PrevID == nil || *columns[2].PrevID != col2.ID {
		t.Error("Third column should point back to inserted column")
	}
}

// TestInsertColumnEnd tests appending a column to the end
func TestInsertColumnEnd(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create initial columns
	_, _ = CreateColumn(db, "Todo", 1, nil)
	col2, _ := CreateColumn(db, "In Progress", 1, nil)

	// Append a column to the end (pass nil for afterColumnID)
	col3, err := CreateColumn(db, "Done", 1, nil)
	if err != nil {
		t.Fatalf("Failed to append column: %v", err)
	}

	// Verify the linked list structure
	columns, err := GetAllColumns(db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}

	// Verify last column
	lastCol := columns[2]
	if lastCol.ID != col3.ID {
		t.Error("Last column should be the newly appended column")
	}
	if lastCol.NextID != nil {
		t.Error("Last column should have nil NextID")
	}
	if lastCol.PrevID == nil || *lastCol.PrevID != col2.ID {
		t.Error("Last column should point back to previous column")
	}
}

// TestDeleteColumnMiddle tests deleting a middle column
func TestDeleteColumnMiddle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create three columns
	col1, _ := CreateColumn(db, "Todo", 1, nil)
	col2, _ := CreateColumn(db, "In Progress", 1, nil)
	col3, _ := CreateColumn(db, "Done", 1, nil)

	// Delete the middle column
	err := DeleteColumn(db, col2.ID)
	if err != nil {
		t.Fatalf("Failed to delete column: %v", err)
	}

	// Verify only two columns remain
	columns, err := GetAllColumns(db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 2 {
		t.Errorf("Expected 2 columns after deletion, got %d", len(columns))
	}

	// Verify the linked list is correct: Todo -> Done
	if columns[0].ID != col1.ID || columns[1].ID != col3.ID {
		t.Error("Remaining columns should be Todo and Done")
	}

	// Verify pointers are updated
	if columns[0].NextID == nil || *columns[0].NextID != col3.ID {
		t.Error("First column should now point to third column")
	}
	if columns[1].PrevID == nil || *columns[1].PrevID != col1.ID {
		t.Error("Third column should now point back to first column")
	}
}

// TestDeleteColumnHead tests deleting the head column
func TestDeleteColumnHead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create three columns
	col1, _ := CreateColumn(db, "Todo", 1, nil)
	col2, _ := CreateColumn(db, "In Progress", 1, nil)
	col3, _ := CreateColumn(db, "Done", 1, nil)

	// Delete the head column
	err := DeleteColumn(db, col1.ID)
	if err != nil {
		t.Fatalf("Failed to delete head column: %v", err)
	}

	// Verify two columns remain
	columns, err := GetAllColumns(db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 2 {
		t.Errorf("Expected 2 columns after deletion, got %d", len(columns))
	}

	// Verify new head
	if columns[0].ID != col2.ID {
		t.Error("Second column should now be the head")
	}
	if columns[0].PrevID != nil {
		t.Error("New head should have nil PrevID")
	}
	if columns[0].NextID == nil || *columns[0].NextID != col3.ID {
		t.Error("New head should point to third column")
	}
}

// TestDeleteColumnTail tests deleting the tail column
func TestDeleteColumnTail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create three columns
	col1, _ := CreateColumn(db, "Todo", 1, nil)
	col2, _ := CreateColumn(db, "In Progress", 1, nil)
	col3, _ := CreateColumn(db, "Done", 1, nil)

	// Delete the tail column
	err := DeleteColumn(db, col3.ID)
	if err != nil {
		t.Fatalf("Failed to delete tail column: %v", err)
	}

	// Verify two columns remain
	columns, err := GetAllColumns(db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 2 {
		t.Errorf("Expected 2 columns after deletion, got %d", len(columns))
	}

	// Verify new tail
	if columns[1].ID != col2.ID {
		t.Error("Second column should now be the tail")
	}
	if columns[1].NextID != nil {
		t.Error("New tail should have nil NextID")
	}
	if columns[1].PrevID == nil || *columns[1].PrevID != col1.ID {
		t.Error("New tail should point back to first column")
	}
}

// TestMoveTaskBetweenColumns tests moving tasks using the linked list functions
func TestMoveTaskBetweenColumns(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create columns
	col1, _ := CreateColumn(db, "Todo", 1, nil)
	col2, _ := CreateColumn(db, "In Progress", 1, nil)
	col3, _ := CreateColumn(db, "Done", 1, nil)

	// Create a task in the first column
	task, err := CreateTask(db, "Test Task", "Description", col1.ID, 0)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Move task to next column (col1 -> col2)
	err = MoveTaskToNextColumn(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to next column: %v", err)
	}

	// Verify task is now in col2
	tasks, err := GetTasksByColumn(db, col2.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be in second column")
	}

	// Verify task is not in col1
	tasks, err = GetTasksByColumn(db, col1.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Error("Task should not be in first column")
	}

	// Move task to next column (col2 -> col3)
	err = MoveTaskToNextColumn(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to third column: %v", err)
	}

	// Verify task is now in col3
	tasks, err = GetTasksByColumn(db, col3.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be in third column")
	}

	// Try to move beyond last column (should fail)
	err = MoveTaskToNextColumn(db, task.ID)
	if err == nil {
		t.Error("Should not be able to move task beyond last column")
	}

	// Move task back (col3 -> col2)
	err = MoveTaskToPrevColumn(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to previous column: %v", err)
	}

	// Verify task is back in col2
	tasks, err = GetTasksByColumn(db, col2.ID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Error("Task should be back in second column")
	}

	// Move task back to col1
	err = MoveTaskToPrevColumn(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to move task to first column: %v", err)
	}

	// Try to move before first column (should fail)
	err = MoveTaskToPrevColumn(db, task.ID)
	if err == nil {
		t.Error("Should not be able to move task before first column")
	}
}

// TestEmptyList tests operations on an empty column list
func TestEmptyList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Get columns from empty database
	columns, err := GetAllColumns(db)
	if err != nil {
		t.Fatalf("Failed to get columns from empty DB: %v", err)
	}

	if len(columns) != 0 {
		t.Errorf("Expected 0 columns in empty DB, got %d", len(columns))
	}
}

// TestSingleColumn tests operations with a single column
func TestSingleColumn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create single column
	col, err := CreateColumn(db, "Only Column", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create single column: %v", err)
	}

	// Verify it's both head and tail
	if col.PrevID != nil || col.NextID != nil {
		t.Error("Single column should have nil PrevID and NextID")
	}

	// Get all columns
	columns, err := GetAllColumns(db)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 1 {
		t.Errorf("Expected 1 column, got %d", len(columns))
	}

	// Create a task and try to move it (should fail in both directions)
	task, _ := CreateTask(db, "Test Task", "", col.ID, 0)

	err = MoveTaskToNextColumn(db, task.ID)
	if err == nil {
		t.Error("Should not be able to move task right from single column")
	}

	err = MoveTaskToPrevColumn(db, task.ID)
	if err == nil {
		t.Error("Should not be able to move task left from single column")
	}
}

// =============================================================================
// Task Persistence Tests
// =============================================================================

// TestTaskCreationPersistence tests that tasks are properly saved to the database
func TestTaskCreationPersistence(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a column first
	col, err := CreateColumn(db, "Todo", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Create a task with title and description
	task, err := CreateTask(db, "Test Task Title", "This is a test description", col.ID, 0)
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
	tasks, err := GetTasksByColumn(db, col.ID)
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
	col, _ := CreateColumn(db, "Todo", 1, nil)
	task, _ := CreateTask(db, "Original Title", "Original Description", col.ID, 0)

	// Update the task
	err := UpdateTask(db, task.ID, "Updated Title", "Updated Description")
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Retrieve and verify the update persisted
	detail, err := GetTaskDetail(db, task.ID)
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

// TestLabelPersistence tests that labels are properly saved and retrieved
func TestLabelPersistence(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a label (projectID 1 is created by migrations)
	label, err := CreateLabel(db, 1, "Bug", "#FF0000")
	if err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	if label.ID == 0 {
		t.Error("Label should have a valid ID")
	}
	if label.Name != "Bug" {
		t.Errorf("Expected label name 'Bug', got '%s'", label.Name)
	}
	if label.Color != "#FF0000" {
		t.Errorf("Expected label color '#FF0000', got '%s'", label.Color)
	}
	if label.ProjectID != 1 {
		t.Errorf("Expected label project ID 1, got %d", label.ProjectID)
	}

	// Retrieve all labels
	labels, err := GetAllLabels(db)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("Expected 1 label, got %d", len(labels))
	}
	if labels[0].Name != "Bug" {
		t.Errorf("Retrieved label has wrong name: %s", labels[0].Name)
	}
}

// TestTaskLabelAssociation tests the many-to-many relationship between tasks and labels
func TestTaskLabelAssociation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create column, task, and labels
	col, _ := CreateColumn(db, "Todo", 1, nil)
	task, _ := CreateTask(db, "Test Task", "Description", col.ID, 0)
	label1, _ := CreateLabel(db, 1, "Bug", "#FF0000")
	label2, _ := CreateLabel(db, 1, "Feature", "#00FF00")

	// Associate labels with task
	err := SetTaskLabels(db, task.ID, []int{label1.ID, label2.ID})
	if err != nil {
		t.Fatalf("Failed to set task labels: %v", err)
	}

	// Retrieve labels for task
	labels, err := GetLabelsForTask(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to get labels for task: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("Expected 2 labels, got %d", len(labels))
	}

	// Verify task summary includes labels
	summaries, err := GetTaskSummariesByColumn(db, col.ID)
	if err != nil {
		t.Fatalf("Failed to get task summaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("Expected 1 summary, got %d", len(summaries))
	}
	if len(summaries[0].Labels) != 2 {
		t.Errorf("Expected summary to have 2 labels, got %d", len(summaries[0].Labels))
	}

	// Verify task detail includes labels
	detail, err := GetTaskDetail(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to get task detail: %v", err)
	}
	if len(detail.Labels) != 2 {
		t.Errorf("Expected detail to have 2 labels, got %d", len(detail.Labels))
	}
}

// TestSetTaskLabelsReplaces tests that SetTaskLabels replaces existing labels
func TestSetTaskLabelsReplaces(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create column, task, and labels
	col, _ := CreateColumn(db, "Todo", 1, nil)
	task, _ := CreateTask(db, "Test Task", "", col.ID, 0)
	label1, _ := CreateLabel(db, 1, "Bug", "#FF0000")
	label2, _ := CreateLabel(db, 1, "Feature", "#00FF00")
	label3, _ := CreateLabel(db, 1, "Enhancement", "#0000FF")

	// Set initial labels
	SetTaskLabels(db, task.ID, []int{label1.ID, label2.ID})

	// Replace with different labels
	err := SetTaskLabels(db, task.ID, []int{label3.ID})
	if err != nil {
		t.Fatalf("Failed to replace task labels: %v", err)
	}

	// Verify only the new label is associated
	labels, err := GetLabelsForTask(db, task.ID)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("Expected 1 label after replacement, got %d", len(labels))
	}
	if labels[0].ID != label3.ID {
		t.Errorf("Expected label ID %d, got %d", label3.ID, labels[0].ID)
	}
}

// TestTaskDetailIncludesAllFields tests that GetTaskDetail returns all fields
func TestTaskDetailIncludesAllFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create column, task with description, and labels
	col, _ := CreateColumn(db, "Todo", 1, nil)
	task, _ := CreateTask(db, "Full Task", "A complete description with details", col.ID, 0)
	label, _ := CreateLabel(db, 1, "Important", "#FFD700")
	SetTaskLabels(db, task.ID, []int{label.ID})

	// Get full detail
	detail, err := GetTaskDetail(db, task.ID)
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

// TestProjectSpecificLabels tests that labels are properly scoped to projects
func TestProjectSpecificLabels(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Project 1 is already created by migrations
	// Create a second project
	project2, err := CreateProject(db, "Project 2", "Second project")
	if err != nil {
		t.Fatalf("Failed to create project 2: %v", err)
	}

	// Create labels for project 1
	label1, err := CreateLabel(db, 1, "Bug", "#FF0000")
	if err != nil {
		t.Fatalf("Failed to create label for project 1: %v", err)
	}
	if label1.ProjectID != 1 {
		t.Errorf("Expected project ID 1, got %d", label1.ProjectID)
	}

	// Create labels for project 2
	label2, err := CreateLabel(db, project2.ID, "Feature", "#00FF00")
	if err != nil {
		t.Fatalf("Failed to create label for project 2: %v", err)
	}
	if label2.ProjectID != project2.ID {
		t.Errorf("Expected project ID %d, got %d", project2.ID, label2.ProjectID)
	}

	// GetLabelsByProject should return only project-specific labels
	labelsP1, err := GetLabelsByProject(db, 1)
	if err != nil {
		t.Fatalf("Failed to get labels for project 1: %v", err)
	}
	if len(labelsP1) != 1 {
		t.Errorf("Expected 1 label for project 1, got %d", len(labelsP1))
	}
	if labelsP1[0].Name != "Bug" {
		t.Errorf("Expected label 'Bug', got '%s'", labelsP1[0].Name)
	}

	labelsP2, err := GetLabelsByProject(db, project2.ID)
	if err != nil {
		t.Fatalf("Failed to get labels for project 2: %v", err)
	}
	if len(labelsP2) != 1 {
		t.Errorf("Expected 1 label for project 2, got %d", len(labelsP2))
	}
	if labelsP2[0].Name != "Feature" {
		t.Errorf("Expected label 'Feature', got '%s'", labelsP2[0].Name)
	}

	// GetAllLabels should return all labels
	allLabels, err := GetAllLabels(db)
	if err != nil {
		t.Fatalf("Failed to get all labels: %v", err)
	}
	if len(allLabels) != 2 {
		t.Errorf("Expected 2 total labels, got %d", len(allLabels))
	}
}
