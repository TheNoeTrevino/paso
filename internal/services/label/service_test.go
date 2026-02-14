package label

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

// createTestTaskInProject creates a column and task in the given project, returning the task ID.
func createTestTaskInProject(t *testing.T, db *sql.DB, d fixtures.Dialect, projectID int) int {
	t.Helper()
	columnID := fixtures.CreateTestColumn(t, db, d, projectID, "Default")
	return fixtures.CreateTestTask(t, db, d, columnID, "Test Task")
}

func TestCreateLabel(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateLabelRequest{
			ProjectID: env.ProjectID,
			Name:      "Bug",
			Color:     "#FF5733",
		}

		result, err := env.Svc.CreateLabel(env.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Bug", result.Name)
		assert.Equal(t, "#FF5733", result.Color)
		assert.Equal(t, env.ProjectID, result.ProjectID)
		assert.NotZero(t, result.ID)
	})
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
		needsProj bool // whether to create a real project
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
			name:      "name too long",
			color:     "#FF5733",
			labelName: strings.Repeat("a", 51),
			wantErr:   true,
			errType:   ErrNameTooLong,
			needsProj: true,
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
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				projectID := tt.projectID
				if tt.needsProj {
					projectID = fixtures.CreateBareProject(t, db, d, "Test Project")
				}

				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)
				req := CreateLabelRequest{
					ProjectID: projectID,
					Name:      tt.labelName,
					Color:     tt.color,
				}

				_, err = svc.CreateLabel(context.Background(), req)

				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}

				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			})
		})
	}
}

func TestCreateLabel_InvalidColor(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

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
					ProjectID: env.ProjectID,
					Name:      "Bug",
					Color:     tc.color,
				}

				_, err := env.Svc.CreateLabel(env.Ctx, req)

				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidColor)
			})
		}
	})
}

func TestCreateLabel_InvalidProjectID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateLabelRequest{
			ProjectID: 0,
			Name:      "Bug",
			Color:     "#FF5733",
		}

		_, err := env.Svc.CreateLabel(env.Ctx, req)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidProjectID)
	})
}

func TestGetLabelsByProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.CreateLabel(env.Ctx, CreateLabelRequest{
			ProjectID: env.ProjectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		_, err = env.Svc.CreateLabel(env.Ctx, CreateLabelRequest{
			ProjectID: env.ProjectID,
			Name:      "Feature",
			Color:     "#33FF57",
		})
		require.NoError(t, err)

		results, err := env.Svc.GetLabelsByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.Equal(t, "Bug", results[0].Name)
		assert.Equal(t, "Feature", results[1].Name)
	})
}

func TestGetLabelsByProject_Empty(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		results, err := env.Svc.GetLabelsByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

func TestGetLabelsByProject_InvalidProjectID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetLabelsByProject(env.Ctx, 0)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidProjectID)
	})
}

func TestGetLabelsForTask(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskID := createTestTaskInProject(t, env.DB, env.Dialect, env.ProjectID)

		label1, err := env.Svc.CreateLabel(env.Ctx, CreateLabelRequest{
			ProjectID: env.ProjectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		label2, err := env.Svc.CreateLabel(env.Ctx, CreateLabelRequest{
			ProjectID: env.ProjectID,
			Name:      "Critical",
			Color:     "#FF0000",
		})
		require.NoError(t, err)

		fixtures.AttachLabelToTask(t, env.DB, env.Dialect, taskID, label1.ID)
		fixtures.AttachLabelToTask(t, env.DB, env.Dialect, taskID, label2.ID)

		results, err := env.Svc.GetLabelsForTask(env.Ctx, taskID)
		require.NoError(t, err)
		require.Len(t, results, 2)

		labelNames := map[string]bool{}
		for _, label := range results {
			labelNames[label.Name] = true
		}
		assert.True(t, labelNames["Bug"])
		assert.True(t, labelNames["Critical"])
	})
}

func TestGetLabelsForTask_Empty(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)
		taskID := createTestTaskInProject(t, env.DB, env.Dialect, env.ProjectID)

		results, err := env.Svc.GetLabelsForTask(env.Ctx, taskID)
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

func TestGetLabelsForTask_InvalidTaskID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetLabelsForTask(env.Ctx, 0)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestUpdateLabel(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		created, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		newName := "Critical Bug"
		newColor := "#FF0000"
		req := UpdateLabelRequest{
			ID:    created.ID,
			Name:  &newName,
			Color: &newColor,
		}

		err = svc.UpdateLabel(ctx, req)
		require.NoError(t, err)

		labels, err := svc.GetLabelsByProject(ctx, projectID)
		require.NoError(t, err)
		require.Len(t, labels, 1)
		assert.Equal(t, "Critical Bug", labels[0].Name)
		assert.Equal(t, "#FF0000", labels[0].Color)
	})
}

func TestUpdateLabel_OnlyName(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		created, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		newName := "Updated Bug"
		req := UpdateLabelRequest{
			ID:   created.ID,
			Name: &newName,
		}

		err = svc.UpdateLabel(ctx, req)
		require.NoError(t, err)

		labels, err := svc.GetLabelsByProject(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Bug", labels[0].Name)
		assert.Equal(t, "#FF5733", labels[0].Color)
	})
}

func TestUpdateLabel_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		labelID   int
		newName   *string
		newColor  *string
		wantErr   bool
		errType   error
		needsReal bool // whether to create a real label
	}{
		{
			name:      "empty name",
			newName:   ptrStr(""),
			wantErr:   true,
			errType:   ErrEmptyName,
			needsReal: true,
		},
		{
			name:      "invalid color",
			newColor:  ptrStr("FF5733"),
			wantErr:   true,
			errType:   ErrInvalidColor,
			needsReal: true,
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
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)

				labelID := tt.labelID
				if tt.needsReal {
					projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
					label, err := svc.CreateLabel(context.Background(), CreateLabelRequest{
						ProjectID: projectID,
						Name:      "Bug",
						Color:     "#FF5733",
					})
					require.NoError(t, err)
					labelID = label.ID
				}

				req := UpdateLabelRequest{
					ID:    labelID,
					Name:  tt.newName,
					Color: tt.newColor,
				}

				err = svc.UpdateLabel(context.Background(), req)

				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}

				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			})
		})
	}
}

func ptrStr(s string) *string {
	return &s
}

func TestDeleteLabel(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		created, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		err = svc.DeleteLabel(ctx, created.ID)
		require.NoError(t, err)

		labels, err := svc.GetLabelsByProject(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, labels, 0)
	})
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
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)
				err = svc.DeleteLabel(context.Background(), tt.labelID)

				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}

				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			})
		})
	}
}

func TestDeleteLabel_CascadeToTaskLabels(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		taskID := createTestTaskInProject(t, db, d, projectID)
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		created, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		fixtures.AttachLabelToTask(t, db, d, taskID, created.ID)

		taskLabels, err := svc.GetLabelsForTask(ctx, taskID)
		require.NoError(t, err)
		require.Len(t, taskLabels, 1)

		err = svc.DeleteLabel(ctx, created.ID)
		require.NoError(t, err)

		taskLabels, err = svc.GetLabelsForTask(ctx, taskID)
		require.NoError(t, err)
		assert.Len(t, taskLabels, 0)
	})
}

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
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)
				req := CreateLabelRequest{
					ProjectID: tt.projectID,
					Name:      tt.labelName,
					Color:     tt.color,
				}

				_, err = svc.CreateLabel(context.Background(), req)

				if tt.wantErr != nil {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.Error(t, err)
				}
			})
		})
	}
}

func TestCreateLabel_DuplicateNames(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		_, err = svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		_, err = svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#00FF00",
		})

		require.Error(t, err)
		errMsg := err.Error()
		assert.True(t, strings.Contains(errMsg, "label creation error") || strings.Contains(errMsg, "already exists"))
	})
}

func TestCreateLabel_DuplicateNames_DifferentProjects(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID1 := fixtures.CreateBareProject(t, db, d, "Test Project 1")
		projectID2 := fixtures.CreateBareProject(t, db, d, "Test Project 2")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		_, err = svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID1,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		_, err = svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID2,
			Name:      "Bug",
			Color:     "#00FF00",
		})
		assert.NoError(t, err)
	})
}

func TestCreateLabel_SpecialCharacters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		labelName string
		shouldErr bool
	}{
		{name: "unicode characters", labelName: "\u30d0\u30b0", shouldErr: false},
		{name: "emoji", labelName: "\U0001f41b Bug", shouldErr: false},
		{name: "special symbols", labelName: "Bug: P0 [Critical]", shouldErr: false},
		{name: "mixed unicode and emoji", labelName: "\U0001f680 \u65b0\u6a5f\u80fd", shouldErr: false},
		{name: "newline character", labelName: "Bug\nLine2", shouldErr: false},
		{name: "tab character", labelName: "Bug\tTab", shouldErr: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)

				req := CreateLabelRequest{
					ProjectID: projectID,
					Name:      tc.labelName,
					Color:     "#FF5733",
				}

				result, err := svc.CreateLabel(context.Background(), req)

				if tc.shouldErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				if !tc.shouldErr && result != nil {
					assert.Equal(t, tc.labelName, result.Name)
				}
			})
		})
	}
}

func TestGetLabelsByProject_NegativeProjectID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)

		_, err = svc.GetLabelsByProject(context.Background(), -1)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidProjectID)
	})
}

func TestGetLabelsByProject_NonExistentProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)

		labels, err := svc.GetLabelsByProject(context.Background(), 999999)
		assert.NoError(t, err)
		assert.Len(t, labels, 0)
	})
}

func TestGetLabelsForTask_NegativeTaskID(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)

		_, err = svc.GetLabelsForTask(context.Background(), -1)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTaskID)
	})
}

func TestGetLabelsForTask_NonExistentTask(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)

		labels, err := svc.GetLabelsForTask(context.Background(), 999999)
		assert.NoError(t, err)
		assert.Len(t, labels, 0)
	})
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
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)
				req := UpdateLabelRequest{
					ID:    tt.labelID,
					Name:  tt.newName,
					Color: tt.newColor,
				}

				err = svc.UpdateLabel(context.Background(), req)

				assert.ErrorIs(t, err, tt.wantErr)
			})
		})
	}
}

func TestUpdateLabel_NonExistentLabel(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)

		req := UpdateLabelRequest{
			ID:   999999,
			Name: ptrStr("Updated Bug"),
		}

		err = svc.UpdateLabel(context.Background(), req)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrLabelNotFound)
	})
}

func TestUpdateLabel_NameTooLong(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		created, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		longName := strings.Repeat("a", 51)
		req := UpdateLabelRequest{
			ID:   created.ID,
			Name: &longName,
		}

		err = svc.UpdateLabel(ctx, req)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNameTooLong)
	})
}

func TestUpdateLabel_InvalidColorFormats(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		created, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

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
				req := UpdateLabelRequest{
					ID:    created.ID,
					Color: &tc.color,
				}

				err := svc.UpdateLabel(ctx, req)
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidColor)
			})
		}
	})
}

func TestUpdateLabel_NoFieldsToUpdate(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		created, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		req := UpdateLabelRequest{
			ID: created.ID,
		}

		err = svc.UpdateLabel(ctx, req)
		assert.NoError(t, err)

		labels, err := svc.GetLabelsByProject(ctx, projectID)
		require.NoError(t, err)
		require.Len(t, labels, 1)
		assert.Equal(t, "Bug", labels[0].Name)
		assert.Equal(t, "#FF5733", labels[0].Color)
	})
}

func TestDeleteLabel_InvalidLabelID_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		labelID int
		wantErr error
	}{
		{name: "negative label ID", labelID: -1, wantErr: ErrInvalidLabelID},
		{name: "zero label ID already tested", labelID: 0, wantErr: ErrInvalidLabelID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)

				err = svc.DeleteLabel(context.Background(), tt.labelID)
				assert.ErrorIs(t, err, tt.wantErr)
			})
		})
	}
}

func TestDeleteLabel_NonExistentLabel(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)

		err = svc.DeleteLabel(context.Background(), 999999)
		assert.NoError(t, err)
	})
}

func TestDeleteLabel_AlreadyDeleted(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		created, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		err = svc.DeleteLabel(ctx, created.ID)
		require.NoError(t, err)

		err = svc.DeleteLabel(ctx, created.ID)
		assert.NoError(t, err)
	})
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
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)

				req := CreateLabelRequest{
					ProjectID: projectID,
					Name:      tc.labelName,
					Color:     "#FF5733",
				}

				result, err := svc.CreateLabel(context.Background(), req)

				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
				} else {
					assert.NoError(t, err)
					if result != nil {
						assert.Equal(t, tc.labelName, result.Name)
					}
				}
			})
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
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)

				req := CreateLabelRequest{
					ProjectID: projectID,
					Name:      "Label_" + tc.name,
					Color:     tc.color,
				}

				result, err := svc.CreateLabel(context.Background(), req)

				if tc.valid {
					assert.NoError(t, err)
					require.NotNil(t, result)
					assert.Equal(t, tc.color, result.Color)
				} else {
					assert.Error(t, err)
				}
			})
		})
	}
}

func TestUpdateLabel_DuplicateNameInProject(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		label1, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)
		require.NotZero(t, label1.ID)

		label2, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Feature",
			Color:     "#00FF00",
		})
		require.NoError(t, err)

		newName := "Bug"
		req := UpdateLabelRequest{
			ID:   label2.ID,
			Name: &newName,
		}

		err = svc.UpdateLabel(ctx, req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestUpdateLabel_DuplicateName_DifferentProjects(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		project1ID := fixtures.CreateBareProject(t, db, d, "Test Project 1")
		project2ID := fixtures.CreateBareProject(t, db, d, "Test Project 2")

		_, err = svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: project1ID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		label2, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: project2ID,
			Name:      "Feature",
			Color:     "#00FF00",
		})
		require.NoError(t, err)

		newName := "Bug"
		req := UpdateLabelRequest{
			ID:   label2.ID,
			Name: &newName,
		}

		err = svc.UpdateLabel(ctx, req)
		require.NoError(t, err)

		labels, err := svc.GetLabelsByProject(ctx, project2ID)
		require.NoError(t, err)
		require.Len(t, labels, 1)
		assert.Equal(t, "Bug", labels[0].Name)
	})
}

func TestUpdateLabel_SameNameNoChange(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")

		label, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		sameName := "Bug"
		req := UpdateLabelRequest{
			ID:   label.ID,
			Name: &sameName,
		}

		err = svc.UpdateLabel(ctx, req)
		require.NoError(t, err)
	})
}

func TestUpdateLabel_CaseVariation(t *testing.T) {
	t.Parallel()
	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")

		_, err = svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "bug",
			Color:     "#FF5733",
		})
		require.NoError(t, err)

		label2, err := svc.CreateLabel(ctx, CreateLabelRequest{
			ProjectID: projectID,
			Name:      "Feature",
			Color:     "#00FF00",
		})
		require.NoError(t, err)

		newName := "BUG"
		req := UpdateLabelRequest{
			ID:   label2.ID,
			Name: &newName,
		}

		err = svc.UpdateLabel(ctx, req)
		require.NoError(t, err)

		labels, err := svc.GetLabelsByProject(ctx, projectID)
		require.NoError(t, err)
		require.Len(t, labels, 2)

		names := make(map[string]bool)
		for _, l := range labels {
			names[l.Name] = true
		}
		assert.True(t, names["bug"] && names["BUG"])
	})
}
