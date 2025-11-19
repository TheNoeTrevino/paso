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

	// Clear the default seeded columns for fresh tests
	_, err = db.Exec("DELETE FROM columns")
	if err != nil {
		t.Fatalf("Failed to clear columns: %v", err)
	}

	return db
}

// TestLinkedListTraversal tests that columns are created and traversed in correct order
func TestLinkedListTraversal(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create 3 columns
	col1, err := CreateColumn(db, "Todo", nil)
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	col2, err := CreateColumn(db, "In Progress", nil)
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	col3, err := CreateColumn(db, "Done", nil)
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
	col1, _ := CreateColumn(db, "Todo", nil)
	col3, _ := CreateColumn(db, "Done", nil)

	// Insert a column in the middle (after col1)
	col2, err := CreateColumn(db, "In Progress", &col1.ID)
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
	_, _ = CreateColumn(db, "Todo", nil)
	col2, _ := CreateColumn(db, "In Progress", nil)

	// Append a column to the end (pass nil for afterColumnID)
	col3, err := CreateColumn(db, "Done", nil)
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
	col1, _ := CreateColumn(db, "Todo", nil)
	col2, _ := CreateColumn(db, "In Progress", nil)
	col3, _ := CreateColumn(db, "Done", nil)

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
	col1, _ := CreateColumn(db, "Todo", nil)
	col2, _ := CreateColumn(db, "In Progress", nil)
	col3, _ := CreateColumn(db, "Done", nil)

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
	col1, _ := CreateColumn(db, "Todo", nil)
	col2, _ := CreateColumn(db, "In Progress", nil)
	col3, _ := CreateColumn(db, "Done", nil)

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
	col1, _ := CreateColumn(db, "Todo", nil)
	col2, _ := CreateColumn(db, "In Progress", nil)
	col3, _ := CreateColumn(db, "Done", nil)

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
	col, err := CreateColumn(db, "Only Column", nil)
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
