package project

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestCreateProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
		}

		result, err := env.Svc.CreateProject(env.Ctx, req)
		require.NoError(t, err, "Failed to create project")

		require.NotNil(t, result, "Expected project result, got nil")

		assert.Equal(t, "Test Project", result.Name)
		assert.Equal(t, "Test Description", result.Description)
		assert.NotZero(t, result.ID, "Expected project ID to be set")
	})
}

func TestCreateProject_EmptyName(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateProjectRequest{
			Name:        "",
			Description: "Test Description",
		}

		_, err := env.Svc.CreateProject(env.Ctx, req)

		require.Error(t, err, "Expected validation error for empty name")
		assert.ErrorIs(t, err, ErrEmptyName)
	})
}

func TestCreateProject_NameTooLong(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		longName := strings.Repeat("a", 101)

		req := CreateProjectRequest{
			Name:        longName,
			Description: "Test Description",
		}

		_, err := env.Svc.CreateProject(env.Ctx, req)

		require.Error(t, err, "Expected validation error for long name")
		assert.ErrorIs(t, err, ErrNameTooLong)
	})
}

func TestGetAllProjects(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 1",
			Description: "Desc 1",
		})
		require.NoError(t, err, "Failed to create project 1")

		_, err = env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 2",
			Description: "Desc 2",
		})
		require.NoError(t, err, "Failed to create project 2")

		results, err := env.Svc.GetAllProjects(env.Ctx)
		require.NoError(t, err, "Failed to get all projects")

		require.Len(t, results, 2)
		assert.Equal(t, "Project 1", results[0].Name)
		assert.Equal(t, "Project 2", results[1].Name)
	})
}

func TestGetAllProjects_Empty(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		results, err := env.Svc.GetAllProjects(env.Ctx)

		require.NoError(t, err, "Failed to get all projects")

		assert.Len(t, results, 0)
	})
}

func TestGetProjectByID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
		})
		require.NoError(t, err, "Failed to create project")

		result, err := env.Svc.GetProjectByID(env.Ctx, created.ID)
		require.NoError(t, err, "Failed to get project by ID")

		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, "Test Project", result.Name)
		assert.Equal(t, "Test Description", result.Description)
	})
}

func TestGetProjectByID_NotFound(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetProjectByID(env.Ctx, 999)

		require.Error(t, err, "Expected error for non-existent project")
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestGetProjectByID_InvalidID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetProjectByID(env.Ctx, 0)

		require.Error(t, err, "Expected error for invalid ID")
		assert.ErrorIs(t, err, ErrInvalidProjectID)
	})
}

func TestUpdateProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Old Name",
			Description: "Old Description",
		})
		require.NoError(t, err, "Failed to create project")

		newName := "Updated Project"
		req := UpdateProjectRequest{
			ID:   created.ID,
			Name: &newName,
		}

		err = env.Svc.UpdateProject(env.Ctx, req)
		require.NoError(t, err, "Failed to update project")

		updated, err := env.Svc.GetProjectByID(env.Ctx, created.ID)
		require.NoError(t, err, "Failed to get updated project")

		assert.Equal(t, "Updated Project", updated.Name)
		assert.Equal(t, "Old Description", updated.Description)
	})
}

func TestUpdateProject_EmptyName(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Old Name",
			Description: "Old Description",
		})
		require.NoError(t, err, "Failed to create project")

		emptyName := ""
		req := UpdateProjectRequest{
			ID:   created.ID,
			Name: &emptyName,
		}

		err = env.Svc.UpdateProject(env.Ctx, req)

		require.Error(t, err, "Expected validation error for empty name")
		assert.ErrorIs(t, err, ErrEmptyName)
	})
}

func TestUpdateProject_InvalidID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		newName := "Updated Project"
		req := UpdateProjectRequest{
			ID:   0,
			Name: &newName,
		}

		err := env.Svc.UpdateProject(env.Ctx, req)

		require.Error(t, err, "Expected error for invalid ID")
		assert.ErrorIs(t, err, ErrInvalidProjectID)
	})
}

func TestDeleteProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
		})
		require.NoError(t, err, "Failed to create project")

		err = env.Svc.DeleteProject(env.Ctx, created.ID, false)
		require.NoError(t, err, "Failed to delete project")

		_, err = env.Svc.GetProjectByID(env.Ctx, created.ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestDeleteProject_WithTasks(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
		})
		require.NoError(t, err, "Failed to create project")

		columnID := fixtures.CreateTestColumn(t, db, d, created.ID, "Test Column")
		fixtures.CreateTestTask(t, db, d, columnID, "Test Task")

		err = env.Svc.DeleteProject(env.Ctx, created.ID, false)

		require.Error(t, err, "Expected error when deleting project with tasks")
		assert.ErrorIs(t, err, ErrProjectHasTasks)
	})
}

func TestDeleteProject_WithTasksForce(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
		})
		require.NoError(t, err, "Failed to create project")

		columnID := fixtures.CreateTestColumn(t, db, d, created.ID, "Test Column")
		fixtures.CreateTestTask(t, db, d, columnID, "Test Task")

		err = env.Svc.DeleteProject(env.Ctx, created.ID, true)
		require.NoError(t, err, "Failed to delete project with force=true")

		_, err = env.Svc.GetProjectByID(env.Ctx, created.ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestDeleteProject_InvalidID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		err := env.Svc.DeleteProject(env.Ctx, 0, false)

		require.Error(t, err, "Expected error for invalid ID")
		assert.ErrorIs(t, err, ErrInvalidProjectID)
	})
}

func TestGetTaskCount(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
		})
		require.NoError(t, err, "Failed to create project")

		count, err := env.Svc.GetTaskCount(env.Ctx, created.ID)
		require.NoError(t, err, "Failed to get task count")

		assert.Equal(t, 0, count)
	})
}

func TestGetTaskCount_InvalidID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetTaskCount(env.Ctx, 0)

		require.Error(t, err, "Expected error for invalid ID")
		assert.ErrorIs(t, err, ErrInvalidProjectID)
	})
}

func TestCreateProject_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         CreateProjectRequest
		expectedErr error
	}{
		{
			name:        "name_with_unicode",
			req:         CreateProjectRequest{Name: "\u30d7\u30ed\u30b8\u30a7\u30af\u30c8", Description: "Unicode test"},
			expectedErr: nil,
		},
		{
			name:        "name_with_special_chars",
			req:         CreateProjectRequest{Name: "Project-2024_v1.0", Description: "Special chars"},
			expectedErr: nil,
		},
		{
			name:        "name_with_emojis",
			req:         CreateProjectRequest{Name: "Project \U0001f680", Description: "Emoji test"},
			expectedErr: nil,
		},
		{
			name:        "name_exactly_100_chars",
			req:         CreateProjectRequest{Name: strings.Repeat("a", 100), Description: "Boundary test"},
			expectedErr: nil,
		},
		{
			name:        "name_101_chars",
			req:         CreateProjectRequest{Name: strings.Repeat("a", 101), Description: "Boundary test"},
			expectedErr: ErrNameTooLong,
		},
		{
			name:        "empty_description",
			req:         CreateProjectRequest{Name: "Test Project", Description: ""},
			expectedErr: nil,
		},
		{
			name:        "very_long_description",
			req:         CreateProjectRequest{Name: "Test Project", Description: strings.Repeat("x", 10000)},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				env := setupTestEnv(t, db, d, dbType)

				result, err := env.Svc.CreateProject(env.Ctx, tt.req)

				if tt.expectedErr != nil {
					require.Error(t, err)
					assert.ErrorIs(t, err, tt.expectedErr)
				} else {
					require.NoError(t, err, "Failed to create project")
					require.NotNil(t, result, "Expected project result, got nil")
					assert.NotZero(t, result.ID, "Expected project ID to be set")
				}
			})
		})
	}
}

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
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				env := setupTestEnv(t, db, d, dbType)

				_, err := env.Svc.GetProjectByID(env.Ctx, tt.id)

				require.Error(t, err, "Expected error for negative ID")
				assert.ErrorIs(t, err, ErrInvalidProjectID)
			})
		})
	}
}

func TestGetProjectByID_VeryLargeID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetProjectByID(env.Ctx, 999999999)

		require.Error(t, err, "Expected error for non-existent project")
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestUpdateProject_ErrorCases(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Original Name",
			Description: "Original Description",
		})
		require.NoError(t, err, "Failed to create project")

		tests := []struct {
			name        string
			req         UpdateProjectRequest
			expectedErr error
			checkErr    func(error) bool
		}{
			{
				name:        "negative_id",
				req:         UpdateProjectRequest{ID: -1, Name: strPtr("New Name")},
				expectedErr: ErrInvalidProjectID,
			},
			{
				name:     "nonexistent_project",
				req:      UpdateProjectRequest{ID: 999999, Name: strPtr("New Name")},
				checkErr: func(err error) bool { return err != nil && err.Error() != "" },
			},
			{
				name:        "name_too_long",
				req:         UpdateProjectRequest{ID: created.ID, Name: strPtr(strings.Repeat("a", 101))},
				expectedErr: ErrNameTooLong,
			},
			{
				name:        "unicode_name",
				req:         UpdateProjectRequest{ID: created.ID, Name: strPtr("\u30d7\u30ed\u30b8\u30a7\u30af\u30c8\u66f4\u65b0")},
				expectedErr: nil,
			},
			{
				name:        "empty_description",
				req:         UpdateProjectRequest{ID: created.ID, Description: strPtr("")},
				expectedErr: nil,
			},
			{
				name:        "very_long_description",
				req:         UpdateProjectRequest{ID: created.ID, Description: strPtr(strings.Repeat("x", 10000))},
				expectedErr: nil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := env.Svc.UpdateProject(env.Ctx, tt.req)

				if tt.expectedErr != nil {
					require.Error(t, err)
					assert.ErrorIs(t, err, tt.expectedErr)
				} else if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err))
				} else {
					require.NoError(t, err, "Failed to update project")
				}
			})
		}
	})
}

func TestUpdateProject_NonExistentProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		newName := "Updated Name"
		req := UpdateProjectRequest{
			ID:   999999,
			Name: &newName,
		}

		err := env.Svc.UpdateProject(env.Ctx, req)

		require.Error(t, err, "Expected error when updating non-existent project")
		assert.NotEmpty(t, err.Error(), "Expected non-empty error message")
	})
}

func TestDeleteProject_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		projectID   int
		force       bool
		expectedErr error
	}{
		{
			name:        "negative_id",
			projectID:   -1,
			force:       false,
			expectedErr: ErrInvalidProjectID,
		},
		{
			name:        "negative_id_with_force",
			projectID:   -5,
			force:       true,
			expectedErr: ErrInvalidProjectID,
		},
		{
			name:        "very_large_nonexistent_id",
			projectID:   999999999,
			force:       false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				env := setupTestEnv(t, db, d, dbType)

				err := env.Svc.DeleteProject(env.Ctx, tt.projectID, tt.force)

				if tt.expectedErr != nil {
					require.Error(t, err)
					assert.ErrorIs(t, err, tt.expectedErr)
				} else {
					require.NoError(t, err, "Failed to delete project")
				}
			})
		})
	}
}

func TestDeleteProject_NonExistentProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		err := env.Svc.DeleteProject(env.Ctx, 999999, false)

		require.NoError(t, err, "Expected no error when deleting non-existent project (idempotent)")
	})
}

func TestGetTaskCount_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		projectID   int
		expectedErr error
	}{
		{
			name:        "zero_id",
			projectID:   0,
			expectedErr: ErrInvalidProjectID,
		},
		{
			name:        "negative_id",
			projectID:   -1,
			expectedErr: ErrInvalidProjectID,
		},
		{
			name:        "negative_large_id",
			projectID:   -999999,
			expectedErr: ErrInvalidProjectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				env := setupTestEnv(t, db, d, dbType)

				_, err := env.Svc.GetTaskCount(env.Ctx, tt.projectID)

				require.Error(t, err, "Expected error for invalid project ID")
				assert.ErrorIs(t, err, tt.expectedErr)
			})
		})
	}
}

func TestGetTaskCount_NonExistentProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		count, err := env.Svc.GetTaskCount(env.Ctx, 999999)

		require.NoError(t, err, "Expected no error for non-existent project")

		assert.Equal(t, 0, count)
	})
}

func TestGetAllProjects_AfterDelete(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		proj1, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 1",
			Description: "Desc 1",
		})
		require.NoError(t, err, "Failed to create project 1")

		proj2, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 2",
			Description: "Desc 2",
		})
		require.NoError(t, err, "Failed to create project 2")

		err = env.Svc.DeleteProject(env.Ctx, proj1.ID, false)
		require.NoError(t, err, "Failed to delete project 1")

		results, err := env.Svc.GetAllProjects(env.Ctx)
		require.NoError(t, err, "Failed to get all projects")

		require.Len(t, results, 1)
		assert.Equal(t, proj2.ID, results[0].ID)
	})
}

func strPtr(s string) *string {
	return &s
}

func TestGetProjectByGitBranch_Found(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			GitBranch:   "feature/my-feature",
		})
		require.NoError(t, err, "Failed to create project with git branch")

		project, err := env.Svc.GetProjectByGitBranch(env.Ctx, "feature/my-feature")
		require.NoError(t, err, "Should find project by git branch")
		require.NotNil(t, project, "Project should not be nil")

		assert.Equal(t, created.ID, project.ID, "Should return the correct project")
		assert.Equal(t, "Test Project", project.Name, "Project name should match")
		assert.Equal(t, "feature/my-feature", project.GitBranch, "Git branch should match")
	})
}

func TestGetProjectByGitBranch_NotFound(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
		})
		require.NoError(t, err, "Failed to create project")

		project, err := env.Svc.GetProjectByGitBranch(env.Ctx, "feature/non-existent")
		require.NoError(t, err, "Should not return error for not found (nil is valid)")
		assert.Nil(t, project, "Project should be nil when not found")
	})
}

func TestGetProjectByGitBranch_EmptyBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		project, err := env.Svc.GetProjectByGitBranch(env.Ctx, "")
		require.NoError(t, err, "Should not error on empty branch")
		assert.Nil(t, project, "Should return nil for empty branch")
	})
}

func TestGetProjectByGitBranch_MultipleProjectsOneWithBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 1",
			Description: "No branch",
		})
		require.NoError(t, err)

		proj2, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 2",
			Description: "With branch",
			GitBranch:   "feature/target",
		})
		require.NoError(t, err)

		_, err = env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 3",
			Description: "Different branch",
			GitBranch:   "feature/other",
		})
		require.NoError(t, err)

		project, err := env.Svc.GetProjectByGitBranch(env.Ctx, "feature/target")
		require.NoError(t, err)
		require.NotNil(t, project)
		assert.Equal(t, proj2.ID, project.ID, "Should return the correct project")
		assert.Equal(t, "feature/target", project.GitBranch)
	})
}

func TestCreateProject_WithGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			GitBranch:   "feature/my-feature",
		}

		result, err := env.Svc.CreateProject(env.Ctx, req)
		require.NoError(t, err, "Failed to create project with git branch")

		require.NotNil(t, result, "Expected project result, got nil")
		assert.Equal(t, "Test Project", result.Name)
		assert.Equal(t, "feature/my-feature", result.GitBranch, "Git branch should be set")

		found, err := env.Svc.GetProjectByGitBranch(env.Ctx, "feature/my-feature")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, result.ID, found.ID)
	})
}

func TestCreateProject_WithoutGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			GitBranch:   "",
		}

		result, err := env.Svc.CreateProject(env.Ctx, req)
		require.NoError(t, err, "Failed to create project without git branch")

		require.NotNil(t, result)
		assert.Equal(t, "", result.GitBranch, "Git branch should be empty")
	})
}

func TestCreateProject_DuplicateGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "First Project",
			Description: "First",
			GitBranch:   "feature/duplicate",
		})
		require.NoError(t, err, "Failed to create first project")

		_, err = env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Second Project",
			Description: "Second",
			GitBranch:   "feature/duplicate",
		})

		assert.Error(t, err, "Should error on duplicate git branch")
		assert.ErrorIs(t, err, ErrGitBranchAlreadyAssociated, "Should return ErrGitBranchAlreadyAssociated")
	})
}

func TestCreateProject_MultipleProjectsWithoutBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		for i := 0; i < 5; i++ {
			_, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
				Name:        fmt.Sprintf("Project %d", i),
				Description: "No branch",
				GitBranch:   "",
			})
			require.NoError(t, err, "Should allow multiple projects without git branch")
		}

		projects, err := env.Svc.GetAllProjects(env.Ctx)
		require.NoError(t, err)
		assert.Len(t, projects, 5, "Should have 5 projects")
	})
}

func TestCreateProject_VeryLongGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		longBranch := "feature/" + strings.Repeat("a", 300)

		req := CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			GitBranch:   longBranch,
		}

		result, err := env.Svc.CreateProject(env.Ctx, req)

		if err != nil {
			assert.Error(t, err, "Should error on very long git branch")
		} else {
			require.NotNil(t, result)
			assert.LessOrEqual(t, len(result.GitBranch), 255, "Git branch should be <= 255 chars")
		}
	})
}

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
			gitBranch:  "feature/\u7279\u6027",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				env := setupTestEnv(t, db, d, dbType)

				req := CreateProjectRequest{
					Name:        "Test Project",
					Description: "Test Description",
					GitBranch:   tt.gitBranch,
				}

				result, err := env.Svc.CreateProject(env.Ctx, req)

				if tt.shouldFail {
					assert.Error(t, err, "Should fail for branch: %s", tt.gitBranch)
				} else {
					require.NoError(t, err, "Should succeed for branch: %s", tt.gitBranch)
					require.NotNil(t, result)
					assert.Equal(t, tt.gitBranch, result.GitBranch)
				}
			})
		})
	}
}

func TestUpdateProject_SetGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
		})
		require.NoError(t, err)

		newBranch := "feature/new-branch"
		err = env.Svc.UpdateProject(env.Ctx, UpdateProjectRequest{
			ID:        created.ID,
			GitBranch: &newBranch,
		})
		require.NoError(t, err, "Should be able to set git branch on update")

		updated, err := env.Svc.GetProjectByID(env.Ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "feature/new-branch", updated.GitBranch, "Git branch should be updated")

		found, err := env.Svc.GetProjectByGitBranch(env.Ctx, "feature/new-branch")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
	})
}

func TestUpdateProject_ClearGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			GitBranch:   "feature/to-clear",
		})
		require.NoError(t, err)

		emptyBranch := ""
		err = env.Svc.UpdateProject(env.Ctx, UpdateProjectRequest{
			ID:        created.ID,
			GitBranch: &emptyBranch,
		})
		require.NoError(t, err, "Should be able to clear git branch")

		updated, err := env.Svc.GetProjectByID(env.Ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "", updated.GitBranch, "Git branch should be empty")

		found, err := env.Svc.GetProjectByGitBranch(env.Ctx, "feature/to-clear")
		require.NoError(t, err)
		assert.Nil(t, found, "Should not find project by old branch")
	})
}

func TestUpdateProject_ChangeGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			GitBranch:   "feature/old-branch",
		})
		require.NoError(t, err)

		newBranch := "feature/new-branch"
		err = env.Svc.UpdateProject(env.Ctx, UpdateProjectRequest{
			ID:        created.ID,
			GitBranch: &newBranch,
		})
		require.NoError(t, err, "Should be able to change git branch")

		updated, err := env.Svc.GetProjectByID(env.Ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "feature/new-branch", updated.GitBranch)

		found, err := env.Svc.GetProjectByGitBranch(env.Ctx, "feature/old-branch")
		require.NoError(t, err)
		assert.Nil(t, found, "Should not find by old branch")

		found, err = env.Svc.GetProjectByGitBranch(env.Ctx, "feature/new-branch")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
	})
}

func TestUpdateProject_DuplicateGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		proj1, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 1",
			Description: "First",
			GitBranch:   "feature/taken",
		})
		require.NoError(t, err)

		proj2, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 2",
			Description: "Second",
			GitBranch:   "feature/other",
		})
		require.NoError(t, err)

		takenBranch := "feature/taken"
		err = env.Svc.UpdateProject(env.Ctx, UpdateProjectRequest{
			ID:        proj2.ID,
			GitBranch: &takenBranch,
		})

		assert.Error(t, err, "Should error on duplicate git branch")
		assert.ErrorIs(t, err, ErrGitBranchAlreadyAssociated, "Should return ErrGitBranchAlreadyAssociated")

		updated, err := env.Svc.GetProjectByID(env.Ctx, proj2.ID)
		require.NoError(t, err)
		assert.Equal(t, "feature/other", updated.GitBranch, "Git branch should not have changed")

		found, err := env.Svc.GetProjectByGitBranch(env.Ctx, "feature/taken")
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, proj1.ID, found.ID)
	})
}

func TestUpdateProject_SameGitBranchNoChange(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			GitBranch:   "feature/my-branch",
		})
		require.NoError(t, err)

		sameBranch := "feature/my-branch"
		err = env.Svc.UpdateProject(env.Ctx, UpdateProjectRequest{
			ID:        created.ID,
			GitBranch: &sameBranch,
		})
		require.NoError(t, err, "Should succeed when updating to same branch")

		updated, err := env.Svc.GetProjectByID(env.Ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "feature/my-branch", updated.GitBranch)
	})
}

func TestGetAllProjects_IncludesGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 1",
			Description: "With branch",
			GitBranch:   "feature/one",
		})
		require.NoError(t, err)

		_, err = env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 2",
			Description: "Without branch",
		})
		require.NoError(t, err)

		_, err = env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Project 3",
			Description: "With branch",
			GitBranch:   "feature/three",
		})
		require.NoError(t, err)

		projects, err := env.Svc.GetAllProjects(env.Ctx)
		require.NoError(t, err)
		assert.Len(t, projects, 3)

		assert.Equal(t, "feature/one", projects[0].GitBranch)
		assert.Equal(t, "", projects[1].GitBranch)
		assert.Equal(t, "feature/three", projects[2].GitBranch)
	})
}

func TestGetProjectByID_IncludesGitBranch(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateProject(env.Ctx, CreateProjectRequest{
			Name:        "Test Project",
			Description: "Test Description",
			GitBranch:   "feature/test",
		})
		require.NoError(t, err)

		project, err := env.Svc.GetProjectByID(env.Ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "feature/test", project.GitBranch, "Git branch should be included")
	})
}
