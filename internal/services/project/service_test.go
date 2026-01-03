package project

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "Failed to create project")

	require.NotNil(t, result, "Expected project result, got nil")

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
	require.NoError(t, err, "Failed to create project 1")

	_, err = svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Project 2",
		Description: "Desc 2",
	})
	require.NoError(t, err, "Failed to create project 2")

	results, err := svc.GetAllProjects(context.Background())
	assert.NoError(t, err, "Failed to get all projects")

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

	assert.NoError(t, err, "Failed to get all projects")

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
	require.NoError(t, err, "Failed to create project")

	result, err := svc.GetProjectByID(context.Background(), created.ID)
	assert.NoError(t, err, "Failed to get project by ID")

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
	require.NoError(t, err, "Failed to create project")

	newName := "Updated Project"
	req := UpdateProjectRequest{
		ID:   created.ID,
		Name: &newName,
	}

	err = svc.UpdateProject(context.Background(), req)
	assert.NoError(t, err, "Failed to update project")

	// Verify update
	updated, err := svc.GetProjectByID(context.Background(), created.ID)
	require.NoError(t, err, "Failed to get updated project")

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
	require.NoError(t, err, "Failed to create project")

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
	require.NoError(t, err, "Failed to create project")

	// Delete should succeed since project has no tasks (columns don't matter)
	err = svc.DeleteProject(context.Background(), created.ID, false)
	assert.NoError(t, err, "Failed to delete project")

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
	require.NoError(t, err, "Failed to create project")

	// Create a column first (tasks are associated via column)
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name) VALUES (?, ?)", created.ID, "Test Column")
	require.NoError(t, err, "Failed to create column")
	columnID, err := result.LastInsertId()
	require.NoError(t, err, "Failed to get column ID")

	// Create a task in the column
	_, err = db.ExecContext(context.Background(), "INSERT INTO tasks (column_id, title, position) VALUES (?, ?, ?)", columnID, "Test Task", 0)
	require.NoError(t, err, "Failed to create task")

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
	require.NoError(t, err, "Failed to create project")

	// Create a column first (tasks are associated via column)
	result, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name) VALUES (?, ?)", created.ID, "Test Column")
	require.NoError(t, err, "Failed to create column")
	columnID, err := result.LastInsertId()
	require.NoError(t, err, "Failed to get column ID")

	// Create a task in the column
	_, err = db.ExecContext(context.Background(), "INSERT INTO tasks (column_id, title, position) VALUES (?, ?, ?)", columnID, "Test Task", 0)
	require.NoError(t, err, "Failed to create task")

	// This should succeed because force=true
	err = svc.DeleteProject(context.Background(), created.ID, true)
	assert.NoError(t, err, "Failed to delete project with force=true")

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
	require.NoError(t, err, "Failed to create project")

	// Initially should have 0 tasks
	count, err := svc.GetTaskCount(context.Background(), created.ID)
	assert.NoError(t, err, "Failed to get task count")

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

// ============================================================================
// ADDITIONAL ERROR PATH TESTS
// ============================================================================

// TestCreateProject_ErrorCases tests various error scenarios for CreateProject using table-driven tests
func TestCreateProject_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         CreateProjectRequest
		expectedErr error
		description string
	}{
		{
			name:        "name_with_unicode",
			req:         CreateProjectRequest{Name: "プロジェクト", Description: "Unicode test"},
			expectedErr: nil, // Unicode should be allowed
			description: "Unicode characters in name should be accepted",
		},
		{
			name:        "name_with_special_chars",
			req:         CreateProjectRequest{Name: "Project-2024_v1.0", Description: "Special chars"},
			expectedErr: nil, // Special chars should be allowed
			description: "Special characters in name should be accepted",
		},
		{
			name:        "name_with_emojis",
			req:         CreateProjectRequest{Name: "Project 🚀", Description: "Emoji test"},
			expectedErr: nil, // Emojis should be allowed
			description: "Emojis in name should be accepted",
		},
		{
			name:        "name_exactly_100_chars",
			req:         CreateProjectRequest{Name: "a" + string(make([]byte, 99)), Description: "Boundary test"},
			expectedErr: nil, // Exactly 100 chars should be allowed
			description: "Name with exactly 100 characters should be accepted",
		},
		{
			name:        "name_101_chars",
			req:         CreateProjectRequest{Name: "a" + string(make([]byte, 100)), Description: "Boundary test"},
			expectedErr: ErrNameTooLong,
			description: "Name with 101 characters should be rejected",
		},
		{
			name:        "empty_description",
			req:         CreateProjectRequest{Name: "Test Project", Description: ""},
			expectedErr: nil, // Empty description should be allowed
			description: "Empty description should be accepted",
		},
		{
			name:        "very_long_description",
			req:         CreateProjectRequest{Name: "Test Project", Description: string(make([]byte, 10000))},
			expectedErr: nil, // Long descriptions should be allowed (no validation on description length)
			description: "Very long description should be accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)

			result, err := svc.CreateProject(context.Background(), tt.req)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectedErr)
				}
				if err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err, "Failed to create project")
				require.NotNil(t, result, "Expected project result, got nil")
				if result.ID == 0 {
					t.Error("Expected project ID to be set")
				}
			}
		})
	}
}

// TestGetProjectByID_NegativeID tests that negative IDs are rejected
func TestGetProjectByID_NegativeID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   int
	}{
		{"negative_one", -1},
		{"negative_hundred", -100},
		{"negative_max_int", -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)

			_, err := svc.GetProjectByID(context.Background(), tt.id)

			if err == nil {
				t.Fatal("Expected error for negative ID")
			}

			if err != ErrInvalidProjectID {
				t.Errorf("Expected ErrInvalidProjectID, got %v", err)
			}
		})
	}
}

// TestGetProjectByID_VeryLargeID tests that non-existent large IDs return sql.ErrNoRows
func TestGetProjectByID_VeryLargeID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetProjectByID(context.Background(), 999999999)

	if err == nil {
		t.Fatal("Expected error for non-existent project")
	}

	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

// TestUpdateProject_ErrorCases tests various error scenarios for UpdateProject
func TestUpdateProject_ErrorCases(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })

	svc := NewService(db, nil)

	// Create a project for update tests
	created, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Original Name",
		Description: "Original Description",
	})
	require.NoError(t, err, "Failed to create project")

	tests := []struct {
		name        string
		req         UpdateProjectRequest
		expectedErr error
		checkErr    func(error) bool
		description string
	}{
		{
			name:        "negative_id",
			req:         UpdateProjectRequest{ID: -1, Name: strPtr("New Name")},
			expectedErr: ErrInvalidProjectID,
			description: "Negative ID should be rejected",
		},
		{
			name:        "nonexistent_project",
			req:         UpdateProjectRequest{ID: 999999, Name: strPtr("New Name")},
			checkErr:    func(err error) bool { return err != nil && err.Error() != "" },
			description: "Non-existent project should return error",
		},
		{
			name:        "name_too_long",
			req:         UpdateProjectRequest{ID: created.ID, Name: strPtr(string(make([]byte, 101)))},
			expectedErr: ErrNameTooLong,
			description: "Name exceeding 100 characters should be rejected",
		},
		{
			name:        "unicode_name",
			req:         UpdateProjectRequest{ID: created.ID, Name: strPtr("プロジェクト更新")},
			expectedErr: nil,
			description: "Unicode characters should be accepted",
		},
		{
			name:        "empty_description",
			req:         UpdateProjectRequest{ID: created.ID, Description: strPtr("")},
			expectedErr: nil,
			description: "Empty description should be accepted",
		},
		{
			name:        "very_long_description",
			req:         UpdateProjectRequest{ID: created.ID, Description: strPtr(string(make([]byte, 10000)))},
			expectedErr: nil,
			description: "Very long description should be accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpdateProject(context.Background(), tt.req)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectedErr)
				}
				if err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else if tt.checkErr != nil {
				if !tt.checkErr(err) {
					t.Errorf("Error check failed for %v", err)
				}
			} else {
				require.NoError(t, err, "Failed to update project")
			}
		})
	}
}

// TestUpdateProject_NonExistentProject explicitly tests updating a non-existent project
func TestUpdateProject_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	newName := "Updated Name"
	req := UpdateProjectRequest{
		ID:   999999,
		Name: &newName,
	}

	err := svc.UpdateProject(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error when updating non-existent project")
	}

	// Should get sql.ErrNoRows wrapped in a fmt error
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestDeleteProject_ErrorCases tests various error scenarios for DeleteProject
func TestDeleteProject_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(*testing.T, Service) int
		projectID   int
		force       bool
		expectedErr error
		description string
	}{
		{
			name:        "negative_id",
			setupFunc:   nil,
			projectID:   -1,
			force:       false,
			expectedErr: ErrInvalidProjectID,
			description: "Negative ID should be rejected",
		},
		{
			name:        "negative_id_with_force",
			setupFunc:   nil,
			projectID:   -5,
			force:       true,
			expectedErr: ErrInvalidProjectID,
			description: "Negative ID should be rejected even with force=true",
		},
		{
			name:        "very_large_nonexistent_id",
			setupFunc:   nil,
			projectID:   999999999,
			force:       false,
			expectedErr: nil, // Non-existent project deletion should succeed (idempotent)
			description: "Non-existent project deletion should be idempotent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)

			var projectID int
			if tt.setupFunc != nil {
				projectID = tt.setupFunc(t, svc)
			} else {
				projectID = tt.projectID
			}

			err := svc.DeleteProject(context.Background(), projectID, tt.force)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectedErr)
				}
				if err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				require.NoError(t, err, "Failed to delete project")
			}
		})
	}
}

// TestDeleteProject_NonExistentProject tests deleting a non-existent project
func TestDeleteProject_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Deleting a non-existent project should succeed (idempotent operation)
	err := svc.DeleteProject(context.Background(), 999999, false)

	assert.NoError(t, err, "Expected no error when deleting non-existent project (idempotent)")
}

// TestGetTaskCount_ErrorCases tests various error scenarios for GetTaskCount
func TestGetTaskCount_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		projectID   int
		expectedErr error
		description string
	}{
		{
			name:        "zero_id",
			projectID:   0,
			expectedErr: ErrInvalidProjectID,
			description: "Zero ID should be rejected",
		},
		{
			name:        "negative_id",
			projectID:   -1,
			expectedErr: ErrInvalidProjectID,
			description: "Negative ID should be rejected",
		},
		{
			name:        "negative_large_id",
			projectID:   -999999,
			expectedErr: ErrInvalidProjectID,
			description: "Large negative ID should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			svc := NewService(db, nil)

			_, err := svc.GetTaskCount(context.Background(), tt.projectID)

			if err == nil {
				t.Fatal("Expected error for invalid project ID")
			}

			if err != tt.expectedErr {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

// TestGetTaskCount_NonExistentProject tests getting task count for non-existent project
func TestGetTaskCount_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Getting task count for non-existent project should return 0
	count, err := svc.GetTaskCount(context.Background(), 999999)

	assert.NoError(t, err, "Expected no error for non-existent project")

	if count != 0 {
		t.Errorf("Expected count 0 for non-existent project, got %d", count)
	}
}

// TestGetAllProjects_AfterDelete tests that deleted projects don't appear in list
func TestGetAllProjects_AfterDelete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create two projects
	proj1, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Project 1",
		Description: "Desc 1",
	})
	require.NoError(t, err, "Failed to create project 1")

	proj2, err := svc.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "Project 2",
		Description: "Desc 2",
	})
	require.NoError(t, err, "Failed to create project 2")

	// Delete first project
	err = svc.DeleteProject(context.Background(), proj1.ID, false)
	require.NoError(t, err, "Failed to delete project 1")

	// Get all projects
	results, err := svc.GetAllProjects(context.Background())
	require.NoError(t, err, "Failed to get all projects")

	// Should only have project 2
	if len(results) != 1 {
		t.Fatalf("Expected 1 project after deletion, got %d", len(results))
	}

	if results[0].ID != proj2.ID {
		t.Errorf("Expected project ID %d, got %d", proj2.ID, results[0].ID)
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
