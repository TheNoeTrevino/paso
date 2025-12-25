package label

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

// createTestSchema creates the minimal schema needed for label service tests
func createTestSchema(db *sql.DB) error {
	schema := `
	-- Create projects table (labels belong to projects)
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create labels table
	CREATE TABLE IF NOT EXISTS labels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		color TEXT NOT NULL DEFAULT '#7D56F4',
		project_id INTEGER NOT NULL,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	-- Create tasks table (for GetLabelsForTask)
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT DEFAULT 'todo',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	-- Create task_labels join table
	CREATE TABLE IF NOT EXISTS task_labels (
		task_id INTEGER NOT NULL,
		label_id INTEGER NOT NULL,
		PRIMARY KEY (task_id, label_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
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
func createTestTask(t *testing.T, db *sql.DB, projectID int) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO tasks (project_id, title, description) VALUES (?, ?, ?)", projectID, "Test Task", "Test Description")
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get task ID: %v", err)
	}
	return int(id)
}

// attachLabelToTask attaches a label to a task
func attachLabelToTask(t *testing.T, db *sql.DB, taskID, labelID int) {
	t.Helper()
	_, err := db.Exec("INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
	if err != nil {
		t.Fatalf("Failed to attach label to task: %v", err)
	}
}

// ============================================================================
// TEST CASES
// ============================================================================

func TestCreateLabel(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	}

	result, err := svc.CreateLabel(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected label result, got nil")
	}

	if result.Name != "Bug" {
		t.Errorf("Expected name 'Bug', got '%s'", result.Name)
	}

	if result.Color != "#FF5733" {
		t.Errorf("Expected color '#FF5733', got '%s'", result.Color)
	}

	if result.ProjectID != projectID {
		t.Errorf("Expected project ID %d, got %d", projectID, result.ProjectID)
	}

	if result.ID == 0 {
		t.Error("Expected label ID to be set")
	}
}

func TestCreateLabel_EmptyName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateLabelRequest{
		ProjectID: projectID,
		Name:      "", // Empty name
		Color:     "#FF5733",
	}

	_, err := svc.CreateLabel(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for empty name")
	}

	if err != ErrEmptyName {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}
}

func TestCreateLabel_NameTooLong(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	longName := ""
	for i := 0; i < 51; i++ {
		longName += "a"
	}

	req := CreateLabelRequest{
		ProjectID: projectID,
		Name:      longName,
		Color:     "#FF5733",
	}

	_, err := svc.CreateLabel(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for long name")
	}

	if err != ErrNameTooLong {
		t.Errorf("Expected ErrNameTooLong, got %v", err)
	}
}

func TestCreateLabel_InvalidColor(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	testCases := []struct {
		name  string
		color string
	}{
		{"missing hash", "FF5733"},
		{"too short", "#FF573"},
		{"too long", "#FF57333"},
		{"invalid chars", "#GG5733"},
		{"lowercase invalid", "#gg5733"},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateLabelRequest{
				ProjectID: projectID,
				Name:      "Bug",
				Color:     tc.color,
			}

			_, err := svc.CreateLabel(context.Background(), req)

			if err == nil {
				t.Fatalf("Expected validation error for color %q", tc.color)
			}

			if err != ErrInvalidColor {
				t.Errorf("Expected ErrInvalidColor, got %v", err)
			}
		})
	}
}

func TestCreateLabel_InvalidProjectID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	req := CreateLabelRequest{
		ProjectID: 0, // Invalid
		Name:      "Bug",
		Color:     "#FF5733",
	}

	_, err := svc.CreateLabel(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for invalid project ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}

func TestGetLabelsByProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two labels
	_, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create label 1: %v", err)
	}

	_, err = svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Feature",
		Color:     "#33FF57",
	})
	if err != nil {
		t.Fatalf("Failed to create label 2: %v", err)
	}

	results, err := svc.GetLabelsByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 labels, got %d", len(results))
	}

	if results[0].Name != "Bug" {
		t.Errorf("Expected first label name 'Bug', got '%s'", results[0].Name)
	}

	if results[1].Name != "Feature" {
		t.Errorf("Expected second label name 'Feature', got '%s'", results[1].Name)
	}
}

func TestGetLabelsByProject_Empty(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	results, err := svc.GetLabelsByProject(context.Background(), projectID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 labels, got %d", len(results))
	}
}

func TestGetLabelsByProject_InvalidProjectID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	_, err := svc.GetLabelsByProject(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid project ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}

func TestGetLabelsForTask(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	taskID := createTestTask(t, db, projectID)
	svc := NewService(db, nil)

	// Create two labels and attach them to the task
	label1, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create label 1: %v", err)
	}

	label2, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Critical",
		Color:     "#FF0000",
	})
	if err != nil {
		t.Fatalf("Failed to create label 2: %v", err)
	}

	attachLabelToTask(t, db, taskID, label1.ID)
	attachLabelToTask(t, db, taskID, label2.ID)

	results, err := svc.GetLabelsForTask(context.Background(), taskID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 labels, got %d", len(results))
	}

	// Check that both labels are present (order may vary)
	labelNames := map[string]bool{}
	for _, label := range results {
		labelNames[label.Name] = true
	}

	if !labelNames["Bug"] {
		t.Error("Expected to find 'Bug' label")
	}

	if !labelNames["Critical"] {
		t.Error("Expected to find 'Critical' label")
	}
}

func TestGetLabelsForTask_Empty(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	taskID := createTestTask(t, db, projectID)
	svc := NewService(db, nil)

	results, err := svc.GetLabelsForTask(context.Background(), taskID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 labels, got %d", len(results))
	}
}

func TestGetLabelsForTask_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	_, err := svc.GetLabelsForTask(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid task ID")
	}

	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

func TestUpdateLabel(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create a label
	created, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	newName := "Critical Bug"
	newColor := "#FF0000"
	req := UpdateLabelRequest{
		ID:    created.ID,
		Name:  &newName,
		Color: &newColor,
	}

	err = svc.UpdateLabel(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify update
	labels, err := svc.GetLabelsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}

	if len(labels) != 1 {
		t.Fatalf("Expected 1 label, got %d", len(labels))
	}

	if labels[0].Name != "Critical Bug" {
		t.Errorf("Expected name 'Critical Bug', got '%s'", labels[0].Name)
	}

	if labels[0].Color != "#FF0000" {
		t.Errorf("Expected color '#FF0000', got '%s'", labels[0].Color)
	}
}

func TestUpdateLabel_OnlyName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create a label
	created, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	newName := "Updated Bug"
	req := UpdateLabelRequest{
		ID:   created.ID,
		Name: &newName,
	}

	err = svc.UpdateLabel(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify update - color should remain unchanged
	labels, err := svc.GetLabelsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}

	if labels[0].Name != "Updated Bug" {
		t.Errorf("Expected name 'Updated Bug', got '%s'", labels[0].Name)
	}

	if labels[0].Color != "#FF5733" {
		t.Errorf("Expected color to remain '#FF5733', got '%s'", labels[0].Color)
	}
}

func TestUpdateLabel_EmptyName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create a label
	created, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	emptyName := ""
	req := UpdateLabelRequest{
		ID:   created.ID,
		Name: &emptyName,
	}

	err = svc.UpdateLabel(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for empty name")
	}

	if err != ErrEmptyName {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}
}

func TestUpdateLabel_InvalidColor(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create a label
	created, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	invalidColor := "FF5733" // Missing hash
	req := UpdateLabelRequest{
		ID:    created.ID,
		Color: &invalidColor,
	}

	err = svc.UpdateLabel(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error for invalid color")
	}

	if err != ErrInvalidColor {
		t.Errorf("Expected ErrInvalidColor, got %v", err)
	}
}

func TestUpdateLabel_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	newName := "Updated Bug"
	req := UpdateLabelRequest{
		ID:   0,
		Name: &newName,
	}

	err := svc.UpdateLabel(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidLabelID {
		t.Errorf("Expected ErrInvalidLabelID, got %v", err)
	}
}

func TestDeleteLabel(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create a label
	created, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	err = svc.DeleteLabel(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify label is deleted
	labels, err := svc.GetLabelsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}

	if len(labels) != 0 {
		t.Errorf("Expected 0 labels after deletion, got %d", len(labels))
	}
}

func TestDeleteLabel_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, nil)

	err := svc.DeleteLabel(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidLabelID {
		t.Errorf("Expected ErrInvalidLabelID, got %v", err)
	}
}

func TestDeleteLabel_CascadeToTaskLabels(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	projectID := createTestProject(t, db)
	taskID := createTestTask(t, db, projectID)
	svc := NewService(db, nil)

	// Create a label and attach it to a task
	created, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	attachLabelToTask(t, db, taskID, created.ID)

	// Verify label is attached
	taskLabels, err := svc.GetLabelsForTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("Failed to get task labels: %v", err)
	}
	if len(taskLabels) != 1 {
		t.Fatalf("Expected 1 task label before deletion, got %d", len(taskLabels))
	}

	// Delete the label
	err = svc.DeleteLabel(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify task_labels entry is also deleted (cascade)
	taskLabels, err = svc.GetLabelsForTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("Failed to get task labels after deletion: %v", err)
	}
	if len(taskLabels) != 0 {
		t.Errorf("Expected 0 task labels after cascade delete, got %d", len(taskLabels))
	}
}
