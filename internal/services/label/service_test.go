package label

import (
	"context"
	"database/sql"
	"testing"

	"github.com/thenoetrevino/paso/internal/testutil"
)

// ============================================================================
// TEST HELPERS
// ============================================================================

// setupTestDB creates an in-memory database with full schema using testutil
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return testutil.SetupTestDB(t)
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
func createTestTask(t *testing.T, db *sql.DB, projectID int) int {
	t.Helper()
	// First create a column for the task
	columnResult, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, "Default")
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}
	columnID, _ := columnResult.LastInsertId()

	// Create task in that column
	result, err := db.ExecContext(context.Background(), "INSERT INTO tasks (column_id, title, description, position) VALUES (?, ?, ?, ?)", columnID, "Test Task", "Test Description", 0)
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
	_, err := db.ExecContext(context.Background(), "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
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
	defer func() { _ = db.Close() }()

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

func TestCreateLabel_Validation(t *testing.T) {
	tests := []struct {
		name      string
		labelName string
		color     string
		projectID int
		wantErr   bool
		errType   error
		setupFn   func(*sql.DB) int // Returns project ID if needed
	}{
		{
			name:      "empty name",
			labelName: "",
			color:     "#FF5733",
			projectID: 1,
			wantErr:   true,
			errType:   ErrEmptyName,
		},
		{
			name:  "name too long",
			color: "#FF5733",
			setupFn: func(db *sql.DB) int {
				return createTestProject(t, db)
			},
			labelName: func() string {
				name := ""
				for i := 0; i < 51; i++ {
					name += "a"
				}
				return name
			}(),
			wantErr: true,
			errType: ErrNameTooLong,
		},
		{
			name:      "invalid project ID",
			labelName: "Bug",
			color:     "#FF5733",
			projectID: 0,
			wantErr:   true,
			errType:   ErrInvalidProjectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			projectID := tt.projectID
			if tt.setupFn != nil {
				projectID = tt.setupFn(db)
			}

			svc := NewService(db, nil)
			req := CreateLabelRequest{
				ProjectID: projectID,
				Name:      tt.labelName,
				Color:     tt.color,
			}

			_, err := svc.CreateLabel(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("CreateLabel() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestCreateLabel_InvalidColor(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })

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
			t.Parallel()
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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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

func TestUpdateLabel_Validation(t *testing.T) {
	tests := []struct {
		name     string
		labelID  int
		newName  *string
		newColor *string
		wantErr  bool
		errType  error
		setupFn  func(*sql.DB) int // Returns label ID if needed
	}{
		{
			name:    "empty name",
			newName: ptrStr(""),
			wantErr: true,
			errType: ErrEmptyName,
			setupFn: func(db *sql.DB) int {
				projectID := createTestProject(t, db)
				label, _ := NewService(db, nil).CreateLabel(context.Background(), CreateLabelRequest{
					ProjectID: projectID,
					Name:      "Bug",
					Color:     "#FF5733",
				})
				return label.ID
			},
		},
		{
			name:     "invalid color",
			newColor: ptrStr("FF5733"), // Missing hash
			wantErr:  true,
			errType:  ErrInvalidColor,
			setupFn: func(db *sql.DB) int {
				projectID := createTestProject(t, db)
				label, _ := NewService(db, nil).CreateLabel(context.Background(), CreateLabelRequest{
					ProjectID: projectID,
					Name:      "Bug",
					Color:     "#FF5733",
				})
				return label.ID
			},
		},
		{
			name:    "invalid ID",
			labelID: 0,
			newName: ptrStr("Updated Bug"),
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			labelID := tt.labelID
			if tt.setupFn != nil {
				labelID = tt.setupFn(db)
			}

			svc := NewService(db, nil)
			req := UpdateLabelRequest{
				ID:    labelID,
				Name:  tt.newName,
				Color: tt.newColor,
			}

			err := svc.UpdateLabel(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("UpdateLabel() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

// ptrStr is a helper function that returns a pointer to a string
func ptrStr(s string) *string {
	return &s
}

func TestDeleteLabel(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

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

func TestDeleteLabel_Validation(t *testing.T) {
	tests := []struct {
		name    string
		labelID int
		wantErr bool
		errType error
	}{
		{
			name:    "invalid ID",
			labelID: 0,
			wantErr: true,
			errType: ErrInvalidLabelID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			err := svc.DeleteLabel(context.Background(), tt.labelID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("DeleteLabel() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestDeleteLabel_CascadeToTaskLabels(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

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
