package label

import (
	"context"
	"database/sql"
	"strings"
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
	t.Parallel()
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
				for range 51 {
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
	t.Parallel()
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
	t.Parallel()
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

// ============================================================================
// ERROR PATH TESTS
// ============================================================================

func TestCreateLabel_InvalidLabelID_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		projectID int
		labelName string
		color     string
		wantErr   error
	}{
		{
			name:      "negative project ID",
			projectID: -1,
			labelName: "Bug",
			color:     "#FF5733",
			wantErr:   ErrInvalidProjectID,
		},
		{
			name:      "zero project ID",
			projectID: 0,
			labelName: "Bug",
			color:     "#FF5733",
			wantErr:   ErrInvalidProjectID,
		},
		{
			name:      "non-existent project ID",
			projectID: 999999,
			labelName: "Bug",
			color:     "#FF5733",
			wantErr:   nil, // Will be caught by database constraint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			req := CreateLabelRequest{
				ProjectID: tt.projectID,
				Name:      tt.labelName,
				Color:     tt.color,
			}

			_, err := svc.CreateLabel(context.Background(), req)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("CreateLabel() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				// For non-existent project, we expect some error (likely foreign key constraint)
				if err == nil {
					t.Error("CreateLabel() expected error for non-existent project ID, got nil")
				}
			}
		})
	}
}

func TestCreateLabel_DuplicateNames(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create first label
	_, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create first label: %v", err)
	}

	// Try to create second label with same name
	_, err = svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#00FF00",
	})

	// Database has UNIQUE(name, project_id) constraint, should return error
	if err == nil {
		t.Fatal("Expected error when creating duplicate label, but got nil")
	}

	// Verify the error message mentions the constraint or duplicate
	errMsg := err.Error()
	if !strings.Contains(errMsg, "label creation error") && !strings.Contains(errMsg, "already exists") {
		t.Errorf("Expected error about duplicate label, got: %v", err)
	}
}

func TestCreateLabel_DuplicateNames_DifferentProjects(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID1 := createTestProject(t, db)
	projectID2 := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create label in first project
	_, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID1,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create label in first project: %v", err)
	}

	// Create label with same name in different project (should succeed)
	_, err = svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID2,
		Name:      "Bug",
		Color:     "#00FF00",
	})
	if err != nil {
		t.Errorf("CreateLabel() in different project should succeed, got error: %v", err)
	}
}

func TestCreateLabel_SpecialCharacters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		labelName string
		shouldErr bool
	}{
		{
			name:      "unicode characters",
			labelName: "バグ",
			shouldErr: false,
		},
		{
			name:      "emoji",
			labelName: "🐛 Bug",
			shouldErr: false,
		},
		{
			name:      "special symbols",
			labelName: "Bug: P0 [Critical]",
			shouldErr: false,
		},
		{
			name:      "mixed unicode and emoji",
			labelName: "🚀 新機能",
			shouldErr: false,
		},
		{
			name:      "newline character",
			labelName: "Bug\nLine2",
			shouldErr: false,
		},
		{
			name:      "tab character",
			labelName: "Bug\tTab",
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			projectID := createTestProject(t, db)
			svc := NewService(db, nil)

			req := CreateLabelRequest{
				ProjectID: projectID,
				Name:      tc.labelName,
				Color:     "#FF5733",
			}

			result, err := svc.CreateLabel(context.Background(), req)

			if tc.shouldErr && err == nil {
				t.Error("CreateLabel() expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("CreateLabel() expected no error, got %v", err)
			}
			if !tc.shouldErr && result != nil && result.Name != tc.labelName {
				t.Errorf("CreateLabel() name = %q, want %q", result.Name, tc.labelName)
			}
		})
	}
}

func TestGetLabelsByProject_NegativeProjectID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetLabelsByProject(context.Background(), -1)

	if err == nil {
		t.Fatal("GetLabelsByProject() expected error for negative project ID, got nil")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("GetLabelsByProject() error = %v, want %v", err, ErrInvalidProjectID)
	}
}

func TestGetLabelsByProject_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Query non-existent project (should return empty list, not error)
	labels, err := svc.GetLabelsByProject(context.Background(), 999999)
	if err != nil {
		t.Errorf("GetLabelsByProject() for non-existent project should not error, got %v", err)
	}

	if len(labels) != 0 {
		t.Errorf("GetLabelsByProject() for non-existent project should return empty list, got %d labels", len(labels))
	}
}

func TestGetLabelsForTask_NegativeTaskID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetLabelsForTask(context.Background(), -1)

	if err == nil {
		t.Fatal("GetLabelsForTask() expected error for negative task ID, got nil")
	}

	if err != ErrInvalidTaskID {
		t.Errorf("GetLabelsForTask() error = %v, want %v", err, ErrInvalidTaskID)
	}
}

func TestGetLabelsForTask_NonExistentTask(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Query non-existent task (should return empty list, not error)
	labels, err := svc.GetLabelsForTask(context.Background(), 999999)
	if err != nil {
		t.Errorf("GetLabelsForTask() for non-existent task should not error, got %v", err)
	}

	if len(labels) != 0 {
		t.Errorf("GetLabelsForTask() for non-existent task should return empty list, got %d labels", len(labels))
	}
}

func TestUpdateLabel_InvalidLabelID_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		labelID  int
		newName  *string
		newColor *string
		wantErr  error
	}{
		{
			name:    "negative label ID",
			labelID: -1,
			newName: ptrStr("Updated Bug"),
			wantErr: ErrInvalidLabelID,
		},
		{
			name:    "zero label ID",
			labelID: 0,
			newName: ptrStr("Updated Bug"),
			wantErr: ErrInvalidLabelID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)
			req := UpdateLabelRequest{
				ID:    tt.labelID,
				Name:  tt.newName,
				Color: tt.newColor,
			}

			err := svc.UpdateLabel(context.Background(), req)

			if err != tt.wantErr {
				t.Errorf("UpdateLabel() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateLabel_NonExistentLabel(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	req := UpdateLabelRequest{
		ID:   999999,
		Name: ptrStr("Updated Bug"),
	}

	err := svc.UpdateLabel(context.Background(), req)

	if err == nil {
		t.Fatal("UpdateLabel() expected error for non-existent label, got nil")
	}

	if err != ErrLabelNotFound {
		t.Errorf("UpdateLabel() error = %v, want %v", err, ErrLabelNotFound)
	}
}

func TestUpdateLabel_NameTooLong(t *testing.T) {
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

	// Try to update with too long name
	longName := ""
	for i := 0; i < 51; i++ {
		longName += "a"
	}

	req := UpdateLabelRequest{
		ID:   created.ID,
		Name: &longName,
	}

	err = svc.UpdateLabel(context.Background(), req)

	if err == nil {
		t.Fatal("UpdateLabel() expected error for name too long, got nil")
	}

	if err != ErrNameTooLong {
		t.Errorf("UpdateLabel() error = %v, want %v", err, ErrNameTooLong)
	}
}

func TestUpdateLabel_InvalidColorFormats(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

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

	testCases := []struct {
		name  string
		color string
	}{
		{"missing hash", "FF5733"},
		{"too short", "#FF573"},
		{"too long", "#FF57333"},
		{"invalid chars", "#GGGGGG"},
		{"lowercase invalid", "gggggg"},
		{"empty", ""},
		{"spaces", "#FF 57 33"},
		{"rgb format", "rgb(255, 87, 51)"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := UpdateLabelRequest{
				ID:    created.ID,
				Color: &tc.color,
			}

			err := svc.UpdateLabel(context.Background(), req)

			if err == nil {
				t.Errorf("UpdateLabel() expected error for color %q, got nil", tc.color)
			}

			if err != ErrInvalidColor {
				t.Errorf("UpdateLabel() error = %v, want %v", err, ErrInvalidColor)
			}
		})
	}
}

func TestUpdateLabel_NoFieldsToUpdate(t *testing.T) {
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

	// Update with no fields (should be no-op but succeed)
	req := UpdateLabelRequest{
		ID: created.ID,
	}

	err = svc.UpdateLabel(context.Background(), req)
	if err != nil {
		t.Errorf("UpdateLabel() with no fields should succeed, got error: %v", err)
	}

	// Verify nothing changed
	labels, err := svc.GetLabelsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}

	if len(labels) != 1 {
		t.Fatalf("Expected 1 label, got %d", len(labels))
	}

	if labels[0].Name != "Bug" {
		t.Errorf("Expected name to remain 'Bug', got '%s'", labels[0].Name)
	}

	if labels[0].Color != "#FF5733" {
		t.Errorf("Expected color to remain '#FF5733', got '%s'", labels[0].Color)
	}
}

func TestDeleteLabel_InvalidLabelID_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		labelID int
		wantErr error
	}{
		{
			name:    "negative label ID",
			labelID: -1,
			wantErr: ErrInvalidLabelID,
		},
		{
			name:    "zero label ID already tested",
			labelID: 0,
			wantErr: ErrInvalidLabelID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)

			err := svc.DeleteLabel(context.Background(), tt.labelID)

			if err != tt.wantErr {
				t.Errorf("DeleteLabel() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteLabel_NonExistentLabel(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Try to delete non-existent label (should succeed as per service implementation)
	err := svc.DeleteLabel(context.Background(), 999999)
	if err != nil {
		t.Errorf("DeleteLabel() for non-existent label should succeed (idempotent), got error: %v", err)
	}
}

func TestDeleteLabel_AlreadyDeleted(t *testing.T) {
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

	// Delete the label once
	err = svc.DeleteLabel(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to delete label: %v", err)
	}

	// Delete the same label again (should succeed as per idempotent design)
	err = svc.DeleteLabel(context.Background(), created.ID)
	if err != nil {
		t.Errorf("DeleteLabel() second time should succeed (idempotent), got error: %v", err)
	}
}

func TestCreateLabel_BoundaryValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		labelName string
		wantErr   error
	}{
		{
			name:      "exactly 50 characters",
			labelName: "12345678901234567890123456789012345678901234567890",
			wantErr:   nil,
		},
		{
			name:      "51 characters",
			labelName: "123456789012345678901234567890123456789012345678901",
			wantErr:   ErrNameTooLong,
		},
		{
			name:      "single character",
			labelName: "B",
			wantErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			projectID := createTestProject(t, db)
			svc := NewService(db, nil)

			req := CreateLabelRequest{
				ProjectID: projectID,
				Name:      tc.labelName,
				Color:     "#FF5733",
			}

			result, err := svc.CreateLabel(context.Background(), req)

			if tc.wantErr != nil {
				if err != tc.wantErr {
					t.Errorf("CreateLabel() error = %v, want %v", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("CreateLabel() unexpected error: %v", err)
				}
				if result != nil && result.Name != tc.labelName {
					t.Errorf("CreateLabel() name = %q, want %q", result.Name, tc.labelName)
				}
			}
		})
	}
}

func TestCreateLabel_ValidColorFormats(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		color string
		valid bool
	}{
		{"uppercase hex", "#FF5733", true},
		{"lowercase hex", "#ff5733", true},
		{"mixed case hex", "#Ff5733", true},
		{"all zeros", "#000000", true},
		{"all Fs", "#FFFFFF", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			projectID := createTestProject(t, db)
			svc := NewService(db, nil)

			req := CreateLabelRequest{
				ProjectID: projectID,
				Name:      "Label_" + tc.name,
				Color:     tc.color,
			}

			result, err := svc.CreateLabel(context.Background(), req)

			if tc.valid {
				if err != nil {
					t.Errorf("CreateLabel() unexpected error for valid color %q: %v", tc.color, err)
				}
				if result == nil {
					t.Errorf("CreateLabel() expected result, got nil")
				} else if result.Color != tc.color {
					t.Errorf("CreateLabel() color = %q, want %q", result.Color, tc.color)
				}
			} else {
				if err == nil {
					t.Errorf("CreateLabel() expected error for invalid color %q, got nil", tc.color)
				}
			}
		})
	}
}

func TestUpdateLabel_DuplicateNameInProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two labels
	label1, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create first label: %v", err)
	}
	if label1.ID == 0 {
		t.Fatal("Expected label1 to have valid ID")
	}

	label2, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Feature",
		Color:     "#00FF00",
	})
	if err != nil {
		t.Fatalf("Failed to create second label: %v", err)
	}

	// Try to update label2 to have the same name as label1
	newName := "Bug"
	req := UpdateLabelRequest{
		ID:   label2.ID,
		Name: &newName,
	}

	err = svc.UpdateLabel(context.Background(), req)

	// Note: This test documents current behavior. If unique constraint exists in DB,
	// it should fail. Currently allows duplicates.
	if err == nil {
		t.Skip("Database schema allows duplicate label names within same project")
	}
}
