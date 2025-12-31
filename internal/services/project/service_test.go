package project

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

// createTestSchema creates the minimal schema needed for project service tests
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

	-- Create project_counters table (for ticket numbers)
	CREATE TABLE IF NOT EXISTS project_counters (
		project_id INTEGER PRIMARY KEY,
		next_ticket_number INTEGER DEFAULT 1,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	-- Create columns table (for deletion constraint checking)
	CREATE TABLE IF NOT EXISTS columns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		prev_id INTEGER,
		next_id INTEGER,
		holds_ready_tasks BOOLEAN NOT NULL DEFAULT 0,
		holds_completed_tasks BOOLEAN NOT NULL DEFAULT 0,
		holds_in_progress_tasks BOOLEAN NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	-- Create tasks table (for task count queries)
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		column_id INTEGER,
		title TEXT NOT NULL,
		description TEXT,
		ticket_number INTEGER,
		status TEXT DEFAULT 'todo',
		priority_id INTEGER,
		type_id INTEGER,
		position REAL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE SET NULL
	);
	`

	_, err := db.ExecContext(context.Background(), schema)
	return err
}

// ============================================================================
// TEST CASES
// ============================================================================

func TestCreateProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil) // nil event publisher is OK

	req := CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	}

	result, err := svc.CreateProject(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected project result, got nil")
	}

	if result.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", result.Name)
	}

	if result.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", result.Description)
	}

	if result.ID == 0 {
		t.Error("Expected project ID to be set")
	}
}

func TestCreateProject_EmptyName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	req := CreateProjectRequest{
		Name:        "", // Empty name
		Description: "Test Description",
	}

	_, err := svc.CreateProject(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for empty name")
	}

	if err != ErrEmptyName {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}
}

func TestCreateProject_NameTooLong(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}

	req := CreateProjectRequest{
		Name:        longName,
		Description: "Test Description",
	}

	_, err := svc.CreateProject(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for long name")
	}

	if err != ErrNameTooLong {
		t.Errorf("Expected ErrNameTooLong, got %v", err)
	}
}

func TestGetAllProjects(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create two projects
	_, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Project 1",
		Description: "Desc 1",
	})
	if err != nil {
		t.Fatalf("Failed to create project 1: %v", err)
	}

	_, err = svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Project 2",
		Description: "Desc 2",
	})
	if err != nil {
		t.Fatalf("Failed to create project 2: %v", err)
	}

	results, err := svc.GetAllProjects(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 projects, got %d", len(results))
	}

	if results[0].Name != "Project 1" {
		t.Errorf("Expected first project name 'Project 1', got '%s'", results[0].Name)
	}

	if results[1].Name != "Project 2" {
		t.Errorf("Expected second project name 'Project 2', got '%s'", results[1].Name)
	}
}

func TestGetAllProjects_Empty(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	results, err := svc.GetAllProjects(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 projects, got %d", len(results))
	}
}

func TestGetProjectByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create a project
	created, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	result, err := svc.GetProjectByID(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, result.ID)
	}

	if result.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", result.Name)
	}

	if result.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", result.Description)
	}
}

func TestGetProjectByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetProjectByID(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for non-existent project")
	}

	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetProjectByID_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetProjectByID(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}

func TestUpdateProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create a project
	created, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Old Name",
		Description: "Old Description",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	newName := "Updated Project"
	req := UpdateProjectRequest{
		ID:   created.ID,
		Name: &newName,
	}

	err = svc.UpdateProject(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify update
	updated, err := svc.GetProjectByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated project: %v", err)
	}

	if updated.Name != "Updated Project" {
		t.Errorf("Expected name 'Updated Project', got '%s'", updated.Name)
	}

	if updated.Description != "Old Description" {
		t.Errorf("Expected description to remain 'Old Description', got '%s'", updated.Description)
	}
}

func TestUpdateProject_EmptyName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create a project
	created, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Old Name",
		Description: "Old Description",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	emptyName := ""
	req := UpdateProjectRequest{
		ID:   created.ID,
		Name: &emptyName,
	}

	err = svc.UpdateProject(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for empty name")
	}

	if err != ErrEmptyName {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}
}

func TestUpdateProject_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	newName := "Updated Project"
	req := UpdateProjectRequest{
		ID:   0,
		Name: &newName,
	}

	err := svc.UpdateProject(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}

func TestDeleteProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create a project (which will have default columns)
	created, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Delete should succeed since project has no tasks (columns don't matter)
	err = svc.DeleteProject(context.Background(), created.ID, false)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify project is deleted
	_, err = svc.GetProjectByID(context.Background(), created.ID)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows after deletion, got %v", err)
	}
}

func TestDeleteProject_WithTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create a project
	created, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a column first (tasks are associated via column)
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name) VALUES (?, ?)", created.ID, "Test Column")
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}
	columnID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get column ID: %v", err)
	}

	// Create a task in the column
	_, err = db.ExecContext(context.Background(), "INSERT INTO tasks (project_id, column_id, title) VALUES (?, ?, ?)", created.ID, columnID, "Test Task")
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// This should fail because project has tasks and force=false
	err = svc.DeleteProject(context.Background(), created.ID, false)

	if err == nil {
		t.Fatal("Expected error when deleting project with tasks")
	}

	if err != ErrProjectHasTasks {
		t.Errorf("Expected ErrProjectHasTasks, got %v", err)
	}
}

func TestDeleteProject_WithTasksForce(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create a project
	created, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a column first (tasks are associated via column)
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name) VALUES (?, ?)", created.ID, "Test Column")
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}
	columnID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get column ID: %v", err)
	}

	// Create a task in the column
	_, err = db.ExecContext(context.Background(), "INSERT INTO tasks (project_id, column_id, title) VALUES (?, ?, ?)", created.ID, columnID, "Test Task")
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// This should succeed because force=true
	err = svc.DeleteProject(context.Background(), created.ID, true)

	if err != nil {
		t.Fatalf("Expected no error with force=true, got %v", err)
	}

	// Verify project is deleted
	_, err = svc.GetProjectByID(context.Background(), created.ID)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows after deletion, got %v", err)
	}
}

func TestDeleteProject_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.DeleteProject(context.Background(), 0, false)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}

func TestGetTaskCount(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create a project
	created, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Initially should have 0 tasks
	count, err := svc.GetTaskCount(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestGetTaskCount_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetTaskCount(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}
