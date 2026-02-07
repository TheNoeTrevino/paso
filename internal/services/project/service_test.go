package project

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// mockGitChecker is a mock implementation of GitChecker for testing
type mockGitChecker struct {
	branches map[string]bool
}

func newMockGitChecker() *mockGitChecker {
	return &mockGitChecker{
		branches: make(map[string]bool),
	}
}

func (m *mockGitChecker) BranchExists(_ context.Context, branchName string) (bool, error) {
	exists, ok := m.branches[branchName]
	if !ok {
		return true, nil
	}
	return exists, nil
}

// newTestService creates a new service for testing (panics on error since tests use valid SQLite)
func newTestService(t *testing.T, db *sql.DB) Service {
	t.Helper()
	mockGit := newMockGitChecker()
	svc, err := NewService(db, database.SQLite, nil, mockGit)
	require.NoError(t, err, "failed to create test service")
	return svc
}

func TestCreateProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db) // nil event publisher is OK

	req := CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	}

	result, err := svc.CreateProject(context.Background(), req)
	require.NoError(t, err, "Failed to create project")

	require.NotNil(t, result, "Expected project result, got nil")

	assert.Equal(t, "Test Project", result.Name)
	assert.Equal(t, "Test Description", result.Description)
	assert.NotZero(t, result.ID, "Expected project ID to be set")
}

func TestCreateProject_EmptyName(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	req := CreateProjectRequest{
		Name:        "", // Empty name
		Description: "Test Description",
	}

	_, err := svc.CreateProject(context.Background(), req)

	require.Error(t, err, "Expected validation error for empty name")
	assert.Equal(t, ErrEmptyName, err)
}

func TestCreateProject_NameTooLong(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}

	req := CreateProjectRequest{
		Name:        longName,
		Description: "Test Description",
	}

	_, err := svc.CreateProject(context.Background(), req)

	require.Error(t, err, "Expected validation error for long name")
	assert.Equal(t, ErrNameTooLong, err)
}

func TestGetAllProjects(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two projects
	_, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 1",
		Description: "Desc 1",
	})
	require.NoError(t, err, "Failed to create project 1")

	_, err = svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 2",
		Description: "Desc 2",
	})
	require.NoError(t, err, "Failed to create project 2")

	results, err := svc.GetAllProjects(ctx)
	require.NoError(t, err, "Failed to get all projects")

	require.Len(t, results, 2)
	assert.Equal(t, "Project 1", results[0].Name)
	assert.Equal(t, "Project 2", results[1].Name)
}

func TestGetAllProjects_Empty(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	results, err := svc.GetAllProjects(context.Background())

	require.NoError(t, err, "Failed to get all projects")

	assert.Len(t, results, 0)
}

func TestGetProjectByID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	require.NoError(t, err, "Failed to create project")

	result, err := svc.GetProjectByID(ctx, created.ID)
	require.NoError(t, err, "Failed to get project by ID")

	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, "Test Project", result.Name)
	assert.Equal(t, "Test Description", result.Description)
}

func TestGetProjectByID_NotFound(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetProjectByID(context.Background(), 999)

	require.Error(t, err, "Expected error for non-existent project")
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestGetProjectByID_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetProjectByID(context.Background(), 0)

	require.Error(t, err, "Expected error for invalid ID")
	assert.Equal(t, ErrInvalidProjectID, err)
}

func TestUpdateProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Old Name",
		Description: "Old Description",
	})
	require.NoError(t, err, "Failed to create project")

	newName := "Updated Project"
	req := UpdateProjectRequest{
		ID:   created.ID,
		Name: &newName,
	}

	err = svc.UpdateProject(ctx, req)
	require.NoError(t, err, "Failed to update project")

	// Verify update
	updated, err := svc.GetProjectByID(ctx, created.ID)
	require.NoError(t, err, "Failed to get updated project")

	assert.Equal(t, "Updated Project", updated.Name)
	assert.Equal(t, "Old Description", updated.Description)
}

func TestUpdateProject_EmptyName(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Old Name",
		Description: "Old Description",
	})
	require.NoError(t, err, "Failed to create project")

	emptyName := ""
	req := UpdateProjectRequest{
		ID:   created.ID,
		Name: &emptyName,
	}

	err = svc.UpdateProject(ctx, req)

	require.Error(t, err, "Expected validation error for empty name")
	assert.Equal(t, ErrEmptyName, err)
}

func TestUpdateProject_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	newName := "Updated Project"
	req := UpdateProjectRequest{
		ID:   0,
		Name: &newName,
	}

	err := svc.UpdateProject(context.Background(), req)

	require.Error(t, err, "Expected error for invalid ID")
	assert.Equal(t, ErrInvalidProjectID, err)
}

func TestDeleteProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project (which will have default columns)
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	require.NoError(t, err, "Failed to create project")

	// Delete should succeed since project has no tasks (columns don't matter)
	err = svc.DeleteProject(ctx, created.ID, false)
	require.NoError(t, err, "Failed to delete project")

	// Verify project is deleted
	_, err = svc.GetProjectByID(ctx, created.ID)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteProject_WithTasks(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	require.NoError(t, err, "Failed to create project")

	// Create a column first (tasks are associated via column)
	result, err := db.ExecContext(ctx, "INSERT INTO columns (project_id, name) VALUES (?, ?)", created.ID, "Test Column")
	require.NoError(t, err, "Failed to create column")
	columnID, err := result.LastInsertId()
	require.NoError(t, err, "Failed to get column ID")

	// Create a task in the column
	_, err = db.ExecContext(ctx, "INSERT INTO tasks (column_id, title, position) VALUES (?, ?, ?)", columnID, "Test Task", 0)
	require.NoError(t, err, "Failed to create task")

	// This should fail because project has tasks and force=false
	err = svc.DeleteProject(ctx, created.ID, false)

	require.Error(t, err, "Expected error when deleting project with tasks")
	assert.Equal(t, ErrProjectHasTasks, err)
}

func TestDeleteProject_WithTasksForce(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	require.NoError(t, err, "Failed to create project")

	// Create a column first (tasks are associated via column)
	result, err := db.ExecContext(ctx, "INSERT INTO columns (project_id, name) VALUES (?, ?)", created.ID, "Test Column")
	require.NoError(t, err, "Failed to create column")
	columnID, err := result.LastInsertId()
	require.NoError(t, err, "Failed to get column ID")

	// Create a task in the column
	_, err = db.ExecContext(ctx, "INSERT INTO tasks (column_id, title, position) VALUES (?, ?, ?)", columnID, "Test Task", 0)
	require.NoError(t, err, "Failed to create task")

	// This should succeed because force=true
	err = svc.DeleteProject(ctx, created.ID, true)
	require.NoError(t, err, "Failed to delete project with force=true")

	// Verify project is deleted
	_, err = svc.GetProjectByID(ctx, created.ID)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteProject_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	err := svc.DeleteProject(context.Background(), 0, false)

	require.Error(t, err, "Expected error for invalid ID")
	assert.Equal(t, ErrInvalidProjectID, err)
}

func TestGetTaskCount(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	require.NoError(t, err, "Failed to create project")

	// Initially should have 0 tasks
	count, err := svc.GetTaskCount(ctx, created.ID)
	require.NoError(t, err, "Failed to get task count")

	assert.Equal(t, 0, count)
}

func TestGetTaskCount_InvalidID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetTaskCount(context.Background(), 0)

	require.Error(t, err, "Expected error for invalid ID")
	assert.Equal(t, ErrInvalidProjectID, err)
}

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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)

			result, err := svc.CreateProject(context.Background(), tt.req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err, "Failed to create project")
				require.NotNil(t, result, "Expected project result, got nil")
				assert.NotZero(t, result.ID, "Expected project ID to be set")
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)

			_, err := svc.GetProjectByID(context.Background(), tt.id)

			require.Error(t, err, "Expected error for negative ID")
			assert.Equal(t, ErrInvalidProjectID, err)
		})
	}
}

// TestGetProjectByID_VeryLargeID tests that non-existent large IDs return sql.ErrNoRows
func TestGetProjectByID_VeryLargeID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetProjectByID(context.Background(), 999999999)

	require.Error(t, err, "Expected error for non-existent project")
	assert.Equal(t, sql.ErrNoRows, err)
}

// TestUpdateProject_ErrorCases tests various error scenarios for UpdateProject
func TestUpdateProject_ErrorCases(t *testing.T) {
	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project for update tests
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
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
			err := svc.UpdateProject(ctx, tt.req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else if tt.checkErr != nil {
				assert.True(t, tt.checkErr(err))
			} else {
				require.NoError(t, err, "Failed to update project")
			}
		})
	}
}

// TestUpdateProject_NonExistentProject explicitly tests updating a non-existent project
func TestUpdateProject_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	newName := "Updated Name"
	req := UpdateProjectRequest{
		ID:   999999,
		Name: &newName,
	}

	err := svc.UpdateProject(context.Background(), req)

	require.Error(t, err, "Expected error when updating non-existent project")

	// Should get sql.ErrNoRows wrapped in a fmt error
	assert.NotEmpty(t, err.Error(), "Expected non-empty error message")
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)

			var projectID int
			if tt.setupFunc != nil {
				projectID = tt.setupFunc(t, svc)
			} else {
				projectID = tt.projectID
			}

			err := svc.DeleteProject(context.Background(), projectID, tt.force)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err, "Failed to delete project")
			}
		})
	}
}

// TestDeleteProject_NonExistentProject tests deleting a non-existent project
func TestDeleteProject_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Deleting a non-existent project should succeed (idempotent operation)
	err := svc.DeleteProject(context.Background(), 999999, false)

	require.NoError(t, err, "Expected no error when deleting non-existent project (idempotent)")
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)

			_, err := svc.GetTaskCount(context.Background(), tt.projectID)

			require.Error(t, err, "Expected error for invalid project ID")
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

// TestGetTaskCount_NonExistentProject tests getting task count for non-existent project
func TestGetTaskCount_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Getting task count for non-existent project should return 0
	count, err := svc.GetTaskCount(context.Background(), 999999)

	require.NoError(t, err, "Expected no error for non-existent project")

	assert.Equal(t, 0, count)
}

// TestGetAllProjects_AfterDelete tests that deleted projects don't appear in list
func TestGetAllProjects_AfterDelete(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two projects
	proj1, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 1",
		Description: "Desc 1",
	})
	require.NoError(t, err, "Failed to create project 1")

	proj2, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 2",
		Description: "Desc 2",
	})
	require.NoError(t, err, "Failed to create project 2")

	// Delete first project
	err = svc.DeleteProject(ctx, proj1.ID, false)
	require.NoError(t, err, "Failed to delete project 1")

	// Get all projects
	results, err := svc.GetAllProjects(ctx)
	require.NoError(t, err, "Failed to get all projects")

	// Should only have project 2
	require.Len(t, results, 1)
	assert.Equal(t, proj2.ID, results[0].ID)
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

// TestGetProjectByGitBranch_Found tests finding a project by git branch
func TestGetProjectByGitBranch_Found(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project with git branch
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
		GitBranch:   "feature/my-feature",
	})
	require.NoError(t, err, "Failed to create project with git branch")

	// Find project by git branch
	project, err := svc.GetProjectByGitBranch(ctx, "feature/my-feature")
	require.NoError(t, err, "Should find project by git branch")
	require.NotNil(t, project, "Project should not be nil")

	assert.Equal(t, created.ID, project.ID, "Should return the correct project")
	assert.Equal(t, "Test Project", project.Name, "Project name should match")
	assert.Equal(t, "feature/my-feature", project.GitBranch, "Git branch should match")
}

// TestGetProjectByGitBranch_NotFound tests when no project is associated with the branch
func TestGetProjectByGitBranch_NotFound(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a project without git branch
	_, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	require.NoError(t, err, "Failed to create project")

	// Try to find project by non-existent branch
	project, err := svc.GetProjectByGitBranch(ctx, "feature/non-existent")
	require.NoError(t, err, "Should not return error for not found (nil is valid)")
	assert.Nil(t, project, "Project should be nil when not found")
}

// TestGetProjectByGitBranch_EmptyBranch tests searching with empty branch name
func TestGetProjectByGitBranch_EmptyBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Try to find project with empty branch name
	project, err := svc.GetProjectByGitBranch(context.Background(), "")
	require.NoError(t, err, "Should not error on empty branch")
	assert.Nil(t, project, "Should return nil for empty branch")
}

// TestGetProjectByGitBranch_MultipleProjectsOneWithBranch tests that only the project with the branch is returned
func TestGetProjectByGitBranch_MultipleProjectsOneWithBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create multiple projects, only one with git branch
	_, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 1",
		Description: "No branch",
	})
	require.NoError(t, err)

	proj2, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 2",
		Description: "With branch",
		GitBranch:   "feature/target",
	})
	require.NoError(t, err)

	_, err = svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 3",
		Description: "Different branch",
		GitBranch:   "feature/other",
	})
	require.NoError(t, err)

	// Find the specific project
	project, err := svc.GetProjectByGitBranch(ctx, "feature/target")
	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, proj2.ID, project.ID, "Should return the correct project")
	assert.Equal(t, "feature/target", project.GitBranch)
}

// TestCreateProject_WithGitBranch tests creating a project with a git branch
func TestCreateProject_WithGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	req := CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
		GitBranch:   "feature/my-feature",
	}

	result, err := svc.CreateProject(ctx, req)
	require.NoError(t, err, "Failed to create project with git branch")

	require.NotNil(t, result, "Expected project result, got nil")
	assert.Equal(t, "Test Project", result.Name)
	assert.Equal(t, "feature/my-feature", result.GitBranch, "Git branch should be set")

	// Verify project can be retrieved by git branch
	found, err := svc.GetProjectByGitBranch(ctx, "feature/my-feature")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, result.ID, found.ID)
}

// TestCreateProject_WithoutGitBranch tests creating a project without a git branch
func TestCreateProject_WithoutGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	req := CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
		GitBranch:   "", // Empty git branch
	}

	result, err := svc.CreateProject(context.Background(), req)
	require.NoError(t, err, "Failed to create project without git branch")

	require.NotNil(t, result)
	assert.Equal(t, "", result.GitBranch, "Git branch should be empty")
}

// TestCreateProject_DuplicateGitBranch tests creating two projects with the same git branch (should fail)
func TestCreateProject_DuplicateGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create first project with git branch
	_, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "First Project",
		Description: "First",
		GitBranch:   "feature/duplicate",
	})
	require.NoError(t, err, "Failed to create first project")

	// Try to create second project with same git branch
	_, err = svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Second Project",
		Description: "Second",
		GitBranch:   "feature/duplicate",
	})

	// Should get unique constraint violation error
	assert.Error(t, err, "Should error on duplicate git branch")
	assert.ErrorIs(t, err, ErrGitBranchAlreadyAssociated, "Should return ErrGitBranchAlreadyAssociated")
}

// TestCreateProject_MultipleProjectsWithoutBranch tests that multiple projects without git branches are allowed
func TestCreateProject_MultipleProjectsWithoutBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create multiple projects without git branches
	for i := 0; i < 5; i++ {
		_, err := svc.CreateProject(ctx, CreateProjectRequest{
			Name:        fmt.Sprintf("Project %d", i),
			Description: "No branch",
			GitBranch:   "", // Empty is allowed multiple times (NULL in DB)
		})
		require.NoError(t, err, "Should allow multiple projects without git branch")
	}

	// Verify all were created
	projects, err := svc.GetAllProjects(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, 5, "Should have 5 projects")
}

// TestCreateProject_VeryLongGitBranch tests creating a project with a very long git branch name
func TestCreateProject_VeryLongGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Create a branch name longer than 255 characters
	longBranch := "feature/" + strings.Repeat("a", 300)

	req := CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
		GitBranch:   longBranch,
	}

	result, err := svc.CreateProject(context.Background(), req)

	// Should either:
	// 1. Truncate to 255 chars and succeed, OR
	// 2. Return a validation error
	// The implementation will determine which
	if err != nil {
		// If error, should be a validation error
		assert.Error(t, err, "Should error on very long git branch")
	} else {
		// If success, should be truncated
		require.NotNil(t, result)
		assert.LessOrEqual(t, len(result.GitBranch), 255, "Git branch should be <= 255 chars")
	}
}

// TestCreateProject_GitBranchWithSpecialChars tests creating a project with special characters in branch name
func TestCreateProject_GitBranchWithSpecialChars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		gitBranch  string
		shouldFail bool
	}{
		{
			name:       "branch_with_slashes",
			gitBranch:  "feature/my-feature",
			shouldFail: false,
		},
		{
			name:       "branch_with_hyphens",
			gitBranch:  "my-awesome-feature",
			shouldFail: false,
		},
		{
			name:       "branch_with_underscores",
			gitBranch:  "my_feature_branch",
			shouldFail: false,
		},
		{
			name:       "branch_with_dots",
			gitBranch:  "release/v1.0.0",
			shouldFail: false,
		},
		{
			name:       "branch_with_unicode",
			gitBranch:  "feature/特性",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)

			req := CreateProjectRequest{
				Name:        "Test Project",
				Description: "Test Description",
				GitBranch:   tt.gitBranch,
			}

			result, err := svc.CreateProject(context.Background(), req)

			if tt.shouldFail {
				assert.Error(t, err, "Should fail for branch: %s", tt.gitBranch)
			} else {
				require.NoError(t, err, "Should succeed for branch: %s", tt.gitBranch)
				require.NotNil(t, result)
				assert.Equal(t, tt.gitBranch, result.GitBranch)
			}
		})
	}
}

// TestUpdateProject_SetGitBranch tests setting a git branch on an existing project
func TestUpdateProject_SetGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create project without git branch
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
	})
	require.NoError(t, err)

	// Update to add git branch
	newBranch := "feature/new-branch"
	err = svc.UpdateProject(ctx, UpdateProjectRequest{
		ID:        created.ID,
		GitBranch: &newBranch,
	})
	require.NoError(t, err, "Should be able to set git branch on update")

	// Verify update
	updated, err := svc.GetProjectByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "feature/new-branch", updated.GitBranch, "Git branch should be updated")

	// Verify can be found by branch
	found, err := svc.GetProjectByGitBranch(ctx, "feature/new-branch")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
}

// TestUpdateProject_ClearGitBranch tests clearing a git branch from a project
func TestUpdateProject_ClearGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create project with git branch
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
		GitBranch:   "feature/to-clear",
	})
	require.NoError(t, err)

	// Update to clear git branch
	emptyBranch := ""
	err = svc.UpdateProject(ctx, UpdateProjectRequest{
		ID:        created.ID,
		GitBranch: &emptyBranch,
	})
	require.NoError(t, err, "Should be able to clear git branch")

	// Verify update
	updated, err := svc.GetProjectByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "", updated.GitBranch, "Git branch should be empty")

	// Verify cannot be found by old branch
	found, err := svc.GetProjectByGitBranch(ctx, "feature/to-clear")
	require.NoError(t, err)
	assert.Nil(t, found, "Should not find project by old branch")
}

// TestUpdateProject_ChangeGitBranch tests changing the git branch of a project
func TestUpdateProject_ChangeGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create project with git branch
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
		GitBranch:   "feature/old-branch",
	})
	require.NoError(t, err)

	// Update to different git branch
	newBranch := "feature/new-branch"
	err = svc.UpdateProject(ctx, UpdateProjectRequest{
		ID:        created.ID,
		GitBranch: &newBranch,
	})
	require.NoError(t, err, "Should be able to change git branch")

	// Verify update
	updated, err := svc.GetProjectByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "feature/new-branch", updated.GitBranch)

	// Verify old branch doesn't work
	found, err := svc.GetProjectByGitBranch(ctx, "feature/old-branch")
	require.NoError(t, err)
	assert.Nil(t, found, "Should not find by old branch")

	// Verify new branch works
	found, err = svc.GetProjectByGitBranch(ctx, "feature/new-branch")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
}

// TestUpdateProject_DuplicateGitBranch tests updating to a branch that's already taken
func TestUpdateProject_DuplicateGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two projects with different branches
	proj1, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 1",
		Description: "First",
		GitBranch:   "feature/taken",
	})
	require.NoError(t, err)

	proj2, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 2",
		Description: "Second",
		GitBranch:   "feature/other",
	})
	require.NoError(t, err)

	// Try to update proj2 to use proj1's branch
	takenBranch := "feature/taken"
	err = svc.UpdateProject(ctx, UpdateProjectRequest{
		ID:        proj2.ID,
		GitBranch: &takenBranch,
	})

	// Should fail with constraint violation
	assert.Error(t, err, "Should error on duplicate git branch")
	assert.ErrorIs(t, err, ErrGitBranchAlreadyAssociated, "Should return ErrGitBranchAlreadyAssociated")

	// Verify proj2 still has original branch
	updated, err := svc.GetProjectByID(ctx, proj2.ID)
	require.NoError(t, err)
	assert.Equal(t, "feature/other", updated.GitBranch, "Git branch should not have changed")

	// Verify proj1 still owns the taken branch
	found, err := svc.GetProjectByGitBranch(ctx, "feature/taken")
	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, proj1.ID, found.ID)
}

// TestUpdateProject_SameGitBranchNoChange tests updating a project with the same git branch (should succeed)
func TestUpdateProject_SameGitBranchNoChange(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create project with git branch
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
		GitBranch:   "feature/my-branch",
	})
	require.NoError(t, err)

	// Update with the same git branch
	sameBranch := "feature/my-branch"
	err = svc.UpdateProject(ctx, UpdateProjectRequest{
		ID:        created.ID,
		GitBranch: &sameBranch,
	})
	require.NoError(t, err, "Should succeed when updating to same branch")

	// Verify branch unchanged
	updated, err := svc.GetProjectByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "feature/my-branch", updated.GitBranch)
}

// TestGetAllProjects_IncludesGitBranch tests that GetAllProjects returns git branch info
func TestGetAllProjects_IncludesGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create projects with and without git branches
	_, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 1",
		Description: "With branch",
		GitBranch:   "feature/one",
	})
	require.NoError(t, err)

	_, err = svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 2",
		Description: "Without branch",
	})
	require.NoError(t, err)

	_, err = svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Project 3",
		Description: "With branch",
		GitBranch:   "feature/three",
	})
	require.NoError(t, err)

	// Get all projects
	projects, err := svc.GetAllProjects(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, 3)

	// Verify git branches are included
	assert.Equal(t, "feature/one", projects[0].GitBranch)
	assert.Equal(t, "", projects[1].GitBranch)
	assert.Equal(t, "feature/three", projects[2].GitBranch)
}

// TestGetProjectByID_IncludesGitBranch tests that GetProjectByID returns git branch info
func TestGetProjectByID_IncludesGitBranch(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)
	ctx := context.Background()

	// Create project with git branch
	created, err := svc.CreateProject(ctx, CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test Description",
		GitBranch:   "feature/test",
	})
	require.NoError(t, err)

	// Get project by ID
	project, err := svc.GetProjectByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "feature/test", project.GitBranch, "Git branch should be included")
}
