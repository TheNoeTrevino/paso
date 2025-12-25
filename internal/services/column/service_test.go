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
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (prev_id) REFERENCES columns(id) ON DELETE SET NULL,
		FOREIGN KEY (next_id) REFERENCES columns(id) ON DELETE SET NULL
	);

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

	_, err := db.Exec(schema)
	return err
}

// createTestProject creates a test project and returns its ID
func createTestProject(t *testing.T, db *sql.DB) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO projects (name, description) VALUES (?, ?)", "Test Project", "Test Description")
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
	result, err := db.Exec("INSERT INTO tasks (column_id, title, description, position) VALUES (?, ?, ?, ?)",
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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
