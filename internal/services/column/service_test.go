package column

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// ============================================================================
// TEST HELPERS
// ============================================================================

// setupTestDB creates an in-memory database and runs migrations
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Enable foreign key constraints
	_, err = db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Run migrations inline (simplified version for tests)
	if err := createTestSchema(db); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// createTestSchema creates the minimal schema needed for column service tests
func createTestSchema(db *sql.DB) error {
	schema := `
	-- Create projects table
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create columns table (linked list structure)
	CREATE TABLE IF NOT EXISTS columns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		prev_id INTEGER,
		next_id INTEGER,
		holds_ready_tasks BOOLEAN NOT NULL DEFAULT 0,
		holds_completed_tasks BOOLEAN NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (prev_id) REFERENCES columns(id) ON DELETE SET NULL,
		FOREIGN KEY (next_id) REFERENCES columns(id) ON DELETE SET NULL
	);

	-- Unique partial index: only one column per project can have holds_ready_tasks = 1
	CREATE UNIQUE INDEX idx_columns_ready_per_project
	ON columns(project_id) WHERE holds_ready_tasks = 1;

	-- Unique partial index: only one column per project can have holds_completed_tasks = 1
	CREATE UNIQUE INDEX idx_columns_completed_per_project
	ON columns(project_id) WHERE holds_completed_tasks = 1;

	-- Create tasks table (for deletion constraint checking)
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		column_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		position INTEGER NOT NULL DEFAULT 0,
		ticket_number INTEGER,
		type_id INTEGER NOT NULL DEFAULT 1,
		priority_id INTEGER NOT NULL DEFAULT 3,
		status TEXT DEFAULT 'todo',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE,
		UNIQUE(column_id, position)
	);
	`

	_, err := db.ExecContext(context.Background(), schema)
	return err
}

// createTestProject creates a test project and returns its ID
func createTestProject(t *testing.T, db *sql.DB) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO projects (name, description) VALUES (?, ?)", "Test Project", "Test Description")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get project ID: %v", err)
	}
	return int(id)
}

// createTestTask creates a test task and returns its ID
func createTestTask(t *testing.T, db *sql.DB, columnID int) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO tasks (column_id, title, description, position) VALUES (?, ?, ?, ?)",
		columnID, "Test Task", "Test Description", 0)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get task ID: %v", err)
	}
	return int(id)
}

// ============================================================================
// TEST CASES
// ============================================================================

func TestCreateColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	}

	result, err := svc.CreateColumn(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected column result, got nil")
	}

	if result.Name != "To Do" {
		t.Errorf("Expected name 'To Do', got '%s'", result.Name)
	}

	if result.ProjectID != projectID {
		t.Errorf("Expected project ID %d, got %d", projectID, result.ProjectID)
	}

	if result.ID == 0 {
		t.Error("Expected column ID to be set")
	}

	// First column should have nil prev and next
	if result.PrevID != nil {
		t.Errorf("Expected prev_id nil for first column, got %v", result.PrevID)
	}

	if result.NextID != nil {
		t.Errorf("Expected next_id nil for first column, got %v", result.NextID)
	}
}

func TestCreateColumn_EmptyName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateColumnRequest{
		Name:      "", // Empty name
		ProjectID: projectID,
	}

	_, err := svc.CreateColumn(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for empty name")
	}

	if err != ErrEmptyName {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}
}

func TestCreateColumn_NameTooLong(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	longName := ""
	for i := 0; i < 51; i++ {
		longName += "a"
	}

	req := CreateColumnRequest{
		Name:      longName,
		ProjectID: projectID,
	}

	_, err := svc.CreateColumn(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for long name")
	}

	if err != ErrNameTooLong {
		t.Errorf("Expected ErrNameTooLong, got %v", err)
	}
}

func TestCreateColumn_InvalidProjectID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	req := CreateColumnRequest{
		Name:      "To Do",
		ProjectID: 0, // Invalid
	}

	_, err := svc.CreateColumn(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for invalid project ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}

func TestCreateColumn_LinkedList(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create first column
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Create second column (should append to end)
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "In Progress",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Create third column (should append to end)
	col3, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 3: %v", err)
	}

	// Verify linked list structure: col1 <-> col2 <-> col3

	// Get updated column 1
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get column 1: %v", err)
	}

	if col1Updated.PrevID != nil {
		t.Errorf("Expected col1 prev_id nil, got %v", col1Updated.PrevID)
	}
	if col1Updated.NextID == nil || *col1Updated.NextID != col2.ID {
		t.Errorf("Expected col1 next_id %d, got %v", col2.ID, col1Updated.NextID)
	}

	// Get updated column 2
	col2Updated, err := svc.GetColumnByID(context.Background(), col2.ID)
	if err != nil {
		t.Fatalf("Failed to get column 2: %v", err)
	}

	if col2Updated.PrevID == nil || *col2Updated.PrevID != col1.ID {
		t.Errorf("Expected col2 prev_id %d, got %v", col1.ID, col2Updated.PrevID)
	}
	if col2Updated.NextID == nil || *col2Updated.NextID != col3.ID {
		t.Errorf("Expected col2 next_id %d, got %v", col3.ID, col2Updated.NextID)
	}

	// Column 3
	if col3.PrevID == nil || *col3.PrevID != col2.ID {
		t.Errorf("Expected col3 prev_id %d, got %v", col2.ID, col3.PrevID)
	}
	if col3.NextID != nil {
		t.Errorf("Expected col3 next_id nil, got %v", col3.NextID)
	}
}

func TestCreateColumn_InsertAfter(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two columns
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	col3, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 3: %v", err)
	}

	// Insert column 2 after column 1
	afterID := col1.ID
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "In Progress",
		ProjectID: projectID,
		AfterID:   &afterID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2 after column 1: %v", err)
	}

	// Verify linked list: col1 <-> col2 <-> col3

	// Get updated columns
	col1Updated, _ := svc.GetColumnByID(context.Background(), col1.ID)
	col2Updated, _ := svc.GetColumnByID(context.Background(), col2.ID)
	col3Updated, _ := svc.GetColumnByID(context.Background(), col3.ID)

	if col1Updated.NextID == nil || *col1Updated.NextID != col2.ID {
		t.Errorf("Expected col1 next_id %d, got %v", col2.ID, col1Updated.NextID)
	}

	if col2Updated.PrevID == nil || *col2Updated.PrevID != col1.ID {
		t.Errorf("Expected col2 prev_id %d, got %v", col1.ID, col2Updated.PrevID)
	}
	if col2Updated.NextID == nil || *col2Updated.NextID != col3.ID {
		t.Errorf("Expected col2 next_id %d, got %v", col3.ID, col2Updated.NextID)
	}

	if col3Updated.PrevID == nil || *col3Updated.PrevID != col2.ID {
		t.Errorf("Expected col3 prev_id %d, got %v", col2.ID, col3Updated.PrevID)
	}
}

func TestGetColumnsByProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two columns
	_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	results, err := svc.GetColumnsByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(results))
	}

	if results[0].Name != "To Do" {
		t.Errorf("Expected first column name 'To Do', got '%s'", results[0].Name)
	}

	if results[1].Name != "Done" {
		t.Errorf("Expected second column name 'Done', got '%s'", results[1].Name)
	}
}

func TestGetColumnsByProject_Empty(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	results, err := svc.GetColumnsByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 columns, got %d", len(results))
	}
}

func TestGetColumnsByProject_InvalidProjectID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetColumnsByProject(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid project ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}

func TestGetColumnByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	result, err := svc.GetColumnByID(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, result.ID)
	}

	if result.Name != "To Do" {
		t.Errorf("Expected name 'To Do', got '%s'", result.Name)
	}
}

func TestGetColumnByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetColumnByID(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for non-existent column")
	}

	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetColumnByID_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetColumnByID(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestUpdateColumnName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	err = svc.UpdateColumnName(context.Background(), created.ID, "Backlog")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify update
	updated, err := svc.GetColumnByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated column: %v", err)
	}

	if updated.Name != "Backlog" {
		t.Errorf("Expected name 'Backlog', got '%s'", updated.Name)
	}
}

func TestUpdateColumnName_EmptyName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	err = svc.UpdateColumnName(context.Background(), created.ID, "")

	if err == nil {
		t.Fatal("Expected validation error for empty name")
	}

	if err != ErrEmptyName {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}
}

func TestUpdateColumnName_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.UpdateColumnName(context.Background(), 0, "Backlog")

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestDeleteColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	err = svc.DeleteColumn(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify column is deleted
	_, err = svc.GetColumnByID(context.Background(), created.ID)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows after deletion, got %v", err)
	}
}

func TestDeleteColumn_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.DeleteColumn(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestDeleteColumn_HasTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Create a task in the column
	createTestTask(t, db, created.ID)

	err = svc.DeleteColumn(context.Background(), created.ID)

	if err == nil {
		t.Fatal("Expected error for column with tasks")
	}

	if err != ErrColumnHasTasks {
		t.Errorf("Expected ErrColumnHasTasks, got %v", err)
	}
}

func TestDeleteColumn_LinkedListIntegrity(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create three columns: col1 <-> col2 <-> col3
	col1, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})

	col2, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "In Progress",
		ProjectID: projectID,
	})

	col3, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})

	// Delete middle column (col2)
	err := svc.DeleteColumn(context.Background(), col2.ID)
	if err != nil {
		t.Fatalf("Failed to delete column 2: %v", err)
	}

	// Verify linked list is repaired: col1 <-> col3

	col1Updated, _ := svc.GetColumnByID(context.Background(), col1.ID)
	col3Updated, _ := svc.GetColumnByID(context.Background(), col3.ID)

	if col1Updated.NextID == nil || *col1Updated.NextID != col3.ID {
		t.Errorf("Expected col1 next_id %d after deletion, got %v", col3.ID, col1Updated.NextID)
	}

	if col3Updated.PrevID == nil || *col3Updated.PrevID != col1.ID {
		t.Errorf("Expected col3 prev_id %d after deletion, got %v", col1.ID, col3Updated.PrevID)
	}
}

// ============================================================================
// TEST CASES - HOLDS_READY_TASKS FEATURE
// ============================================================================

func TestCreateColumn_WithHoldsReadyTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	}

	result, err := svc.CreateColumn(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsReadyTasks {
		t.Error("Expected HoldsReadyTasks to be true")
	}
}

func TestCreateColumn_HoldsReadyTasks_ClearsPrevious(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create first column with HoldsReadyTasks = true
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	if !col1.HoldsReadyTasks {
		t.Fatal("Expected col1 to hold ready tasks")
	}

	// Create second column with HoldsReadyTasks = true
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "In Progress",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	if !col2.HoldsReadyTasks {
		t.Error("Expected col2 to hold ready tasks")
	}

	// Verify col1 is no longer the ready column
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get updated col1: %v", err)
	}

	if col1Updated.HoldsReadyTasks {
		t.Error("Expected col1 to no longer hold ready tasks")
	}
}

func TestSetHoldsReadyTasks_Success(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two columns (neither ready)
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "Done",
		ProjectID:       projectID,
		HoldsReadyTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Set col1 as ready
	updated, err := svc.SetHoldsReadyTasks(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !updated.HoldsReadyTasks {
		t.Error("Expected column to hold ready tasks after SetHoldsReadyTasks")
	}
}

func TestSetHoldsReadyTasks_TransfersFromPrevious(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create col1 as ready
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Create col2 as not ready
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "In Progress",
		ProjectID:       projectID,
		HoldsReadyTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Set col2 as ready (should clear col1)
	updated, err := svc.SetHoldsReadyTasks(context.Background(), col2.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !updated.HoldsReadyTasks {
		t.Error("Expected col2 to hold ready tasks")
	}

	// Verify col1 is no longer ready
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get col1: %v", err)
	}

	if col1Updated.HoldsReadyTasks {
		t.Error("Expected col1 to no longer hold ready tasks")
	}
}

func TestGetColumnByID_IncludesHoldsReadyTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create column with HoldsReadyTasks = true
	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Fetch via GetColumnByID
	result, err := svc.GetColumnByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsReadyTasks {
		t.Error("Expected HoldsReadyTasks to be true")
	}
}

func TestGetColumnsByProject_IncludesHoldsReadyTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create one ready column and one not ready
	_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create ready column: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "Done",
		ProjectID:       projectID,
		HoldsReadyTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create non-ready column: %v", err)
	}

	// Fetch all columns
	results, err := svc.GetColumnsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(results))
	}

	// Verify first column (To Do) is ready
	if results[0].Name == "To Do" && !results[0].HoldsReadyTasks {
		t.Error("Expected 'To Do' column to hold ready tasks")
	}

	// Verify second column (Done) is not ready
	if results[1].Name == "Done" && results[1].HoldsReadyTasks {
		t.Error("Expected 'Done' column to not hold ready tasks")
	}
}

func TestSetHoldsReadyTasks_InvalidColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsReadyTasks(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid column ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestSetHoldsReadyTasks_ColumnNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsReadyTasks(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for non-existent column")
	}

	// Should get a wrapped sql.ErrNoRows
	if err != sql.ErrNoRows && !contains(err.Error(), "no rows") {
		t.Errorf("Expected sql.ErrNoRows or wrapped error, got %v", err)
	}
}

func TestCreateColumn_OnlyOneReadyPerProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)

	// Manually insert two columns with holds_ready_tasks = 1
	// This should violate the unique partial index constraint
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO columns (name, project_id, holds_ready_tasks) VALUES (?, ?, ?)",
		"To Do", projectID, true)
	if err != nil {
		t.Fatalf("Failed to insert first ready column: %v", err)
	}

	_, err = db.ExecContext(context.Background(),
		"INSERT INTO columns (name, project_id, holds_ready_tasks) VALUES (?, ?, ?)",
		"Review", projectID, true)

	if err == nil {
		t.Fatal("Expected database constraint violation for duplicate ready columns")
	}

	// Should get a constraint violation error
	if !contains(err.Error(), "UNIQUE") && !contains(err.Error(), "constraint") {
		t.Errorf("Expected UNIQUE constraint violation, got %v", err)
	}
}

// Helper function to check if error message contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================================
// TEST CASES - HOLDS_COMPLETED_TASKS FEATURE
// ============================================================================

func TestCreateColumn_WithHoldsCompletedTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	}

	result, err := svc.CreateColumn(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsCompletedTasks {
		t.Error("Expected HoldsCompletedTasks to be true")
	}
}

func TestCreateColumn_HoldsCompletedTasks_FailsWhenExists(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create first column with HoldsCompletedTasks = true
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	if !col1.HoldsCompletedTasks {
		t.Fatal("Expected col1 to hold completed tasks")
	}

	// Create second column with HoldsCompletedTasks = true (should fail)
	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Archive",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})

	if err == nil {
		t.Fatal("Expected error when creating second completed column")
	}

	if err != ErrCompletedColumnExists && !contains(err.Error(), "completed column already exists") {
		t.Errorf("Expected ErrCompletedColumnExists, got %v", err)
	}
}

func TestSetHoldsCompletedTasks_Success(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two columns (neither completed)
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Archive",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Set col1 as completed
	updated, err := svc.SetHoldsCompletedTasks(context.Background(), col1.ID, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !updated.HoldsCompletedTasks {
		t.Error("Expected column to hold completed tasks after SetHoldsCompletedTasks")
	}
}

func TestSetHoldsCompletedTasks_FailsWhenExistsWithoutForce(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create col1 as completed
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Create col2 as not completed
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Archive",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Try to set col2 as completed without force (should fail)
	_, err = svc.SetHoldsCompletedTasks(context.Background(), col2.ID, false)

	if err == nil {
		t.Fatal("Expected error when setting completed without force")
	}

	if err != ErrCompletedColumnExists && !contains(err.Error(), "completed column already exists") {
		t.Errorf("Expected ErrCompletedColumnExists, got %v", err)
	}

	// Verify col1 is still completed
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get col1: %v", err)
	}

	if !col1Updated.HoldsCompletedTasks {
		t.Error("Expected col1 to still hold completed tasks")
	}
}

func TestSetHoldsCompletedTasks_SucceedsWithForce(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create col1 as completed
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Create col2 as not completed
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Archive",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Set col2 as completed with force (should succeed)
	updated, err := svc.SetHoldsCompletedTasks(context.Background(), col2.ID, true)
	if err != nil {
		t.Fatalf("Expected no error with force flag, got %v", err)
	}

	if !updated.HoldsCompletedTasks {
		t.Error("Expected col2 to hold completed tasks")
	}

	// Verify col1 is no longer completed
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get col1: %v", err)
	}

	if col1Updated.HoldsCompletedTasks {
		t.Error("Expected col1 to no longer hold completed tasks")
	}
}

func TestGetColumnByID_IncludesHoldsCompletedTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create column with HoldsCompletedTasks = true
	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Fetch via GetColumnByID
	result, err := svc.GetColumnByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsCompletedTasks {
		t.Error("Expected HoldsCompletedTasks to be true")
	}
}

func TestGetColumnsByProject_IncludesHoldsCompletedTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create one completed column and one not completed
	_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create completed column: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Todo",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create non-completed column: %v", err)
	}

	// Fetch all columns
	results, err := svc.GetColumnsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(results))
	}

	// Verify first column (Done) is completed
	if results[0].Name == "Done" && !results[0].HoldsCompletedTasks {
		t.Error("Expected 'Done' column to hold completed tasks")
	}

	// Verify second column (Todo) is not completed
	if results[1].Name == "Todo" && results[1].HoldsCompletedTasks {
		t.Error("Expected 'Todo' column to not hold completed tasks")
	}
}

func TestSetHoldsCompletedTasks_InvalidColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsCompletedTasks(context.Background(), 0, false)

	if err == nil {
		t.Fatal("Expected error for invalid column ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestSetHoldsCompletedTasks_ColumnNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsCompletedTasks(context.Background(), 999, false)

	if err == nil {
		t.Fatal("Expected error for non-existent column")
	}

	// Should get a wrapped sql.ErrNoRows
	if err != sql.ErrNoRows && !contains(err.Error(), "no rows") {
		t.Errorf("Expected sql.ErrNoRows or wrapped error, got %v", err)
	}
}

func TestCreateColumn_OnlyOneCompletedPerProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)

	// Manually insert two columns with holds_completed_tasks = 1
	// This should violate the unique partial index constraint
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO columns (name, project_id, holds_completed_tasks) VALUES (?, ?, ?)",
		"Done", projectID, true)
	if err != nil {
		t.Fatalf("Failed to insert first completed column: %v", err)
	}

	_, err = db.ExecContext(context.Background(),
		"INSERT INTO columns (name, project_id, holds_completed_tasks) VALUES (?, ?, ?)",
		"Archive", projectID, true)

	if err == nil {
		t.Fatal("Expected database constraint violation for duplicate completed columns")
	}

	// Should get a constraint violation error
	if !contains(err.Error(), "UNIQUE") && !contains(err.Error(), "constraint") {
		t.Errorf("Expected UNIQUE constraint violation, got %v", err)
	}
}
