package column

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

func TestCreateColumn(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		}

		result, err := env.Svc.CreateColumn(env.Ctx, req)
		require.NoError(t, err)

		require.NotNil(t, result)

		assert.Equal(t, "To Do", result.Name)

		assert.Equal(t, env.ProjectID, result.ProjectID)

		assert.NotZero(t, result.ID)

		// First column should have nil prev and next
		assert.Nil(t, result.PrevID)

		assert.Nil(t, result.NextID)
	})
}

func TestCreateColumn_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		colName   string
		projectID int
		wantErr   bool
		errType   error
		setupFn   func(*sql.DB, fixtures.Dialect, database.DatabaseType) int // Returns project ID if needed
	}{
		{
			name:      "empty name",
			colName:   "",
			projectID: 1,
			wantErr:   true,
			errType:   ErrEmptyName,
		},
		{
			name: "name too long",
			setupFn: func(db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) int {
				return fixtures.CreateBareProject(t, db, d, "Test Project")
			},
			colName: strings.Repeat("a", 51),
			wantErr: true,
			errType: ErrNameTooLong,
		},
		{
			name:      "invalid project ID",
			colName:   "To Do",
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
				if tt.setupFn != nil {
					projectID = tt.setupFn(db, d, dbType)
				}

				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)
				req := CreateColumnRequest{
					Name:      tt.colName,
					ProjectID: projectID,
				}

				_, err = svc.CreateColumn(context.Background(), req)

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

func TestCreateColumn_LinkedList(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create first column
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column 1")

		// Create second column (should append to end)
		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "In Progress",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column 2")

		// Create third column (should append to end)
		col3, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "Done",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column 3")

		// Verify linked list structure: col1 <-> col2 <-> col3

		// Get updated column 1
		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err, "Failed to get column 1")

		assert.Nil(t, col1Updated.PrevID)
		require.NotNil(t, col1Updated.NextID)
		assert.Equal(t, col2.ID, *col1Updated.NextID)

		// Get updated column 2
		col2Updated, err := env.Svc.GetColumnByID(env.Ctx, col2.ID)
		require.NoError(t, err, "Failed to get column 2")

		require.NotNil(t, col2Updated.PrevID)
		assert.Equal(t, col1.ID, *col2Updated.PrevID)
		require.NotNil(t, col2Updated.NextID)
		assert.Equal(t, col3.ID, *col2Updated.NextID)

		// Column 3
		require.NotNil(t, col3.PrevID)
		assert.Equal(t, col2.ID, *col3.PrevID)
		assert.Nil(t, col3.NextID)
	})
}

func TestCreateColumn_InsertAfter(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create two columns
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column 1")

		col3, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "Done",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column 3")

		// Insert column 2 after column 1
		afterID := col1.ID
		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "In Progress",
			ProjectID: env.ProjectID,
			AfterID:   &afterID,
		})
		require.NoError(t, err, "Failed to create column 2 after column 1")

		// Verify linked list: col1 <-> col2 <-> col3

		// Get updated columns
		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err)
		col2Updated, err := env.Svc.GetColumnByID(env.Ctx, col2.ID)
		require.NoError(t, err)
		col3Updated, err := env.Svc.GetColumnByID(env.Ctx, col3.ID)
		require.NoError(t, err)

		require.NotNil(t, col1Updated.NextID)
		assert.Equal(t, col2.ID, *col1Updated.NextID)

		require.NotNil(t, col2Updated.PrevID)
		assert.Equal(t, col1.ID, *col2Updated.PrevID)
		require.NotNil(t, col2Updated.NextID)
		assert.Equal(t, col3.ID, *col2Updated.NextID)

		require.NotNil(t, col3Updated.PrevID)
		assert.Equal(t, col2.ID, *col3Updated.PrevID)
	})
}

func TestGetColumnsByProject(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create two columns
		_, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column 1")

		_, err = env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "Done",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column 2")

		results, err := env.Svc.GetColumnsByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		require.Len(t, results, 2)

		assert.Equal(t, "To Do", results[0].Name)

		assert.Equal(t, "Done", results[1].Name)
	})
}

func TestGetColumnsByProject_Empty(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		results, err := env.Svc.GetColumnsByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		assert.Len(t, results, 0)
	})
}

func TestGetColumnsByProject_InvalidProjectID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetColumnsByProject(env.Ctx, 0)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidProjectID)
	})
}

func TestGetColumnByID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column")

		result, err := env.Svc.GetColumnByID(env.Ctx, created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, result.ID)

		assert.Equal(t, "To Do", result.Name)
	})
}

func TestGetColumnByID_NotFound(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetColumnByID(env.Ctx, 999)

		require.Error(t, err)

		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestGetColumnByID_InvalidID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetColumnByID(env.Ctx, 0)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestUpdateColumnName(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column")

		err = env.Svc.UpdateColumnName(env.Ctx, created.ID, "Backlog")
		require.NoError(t, err)

		// Verify update
		updated, err := env.Svc.GetColumnByID(env.Ctx, created.ID)
		require.NoError(t, err, "Failed to get updated column")

		assert.Equal(t, "Backlog", updated.Name)
	})
}

func TestUpdateColumnName_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		columnID int
		newName  string
		wantErr  bool
		errType  error
		setupFn  func(*sql.DB, fixtures.Dialect, database.DatabaseType) int // Returns column ID if needed
	}{
		{
			name:    "empty name",
			newName: "",
			wantErr: true,
			errType: ErrEmptyName,
			setupFn: func(db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) int {
				projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)
				col, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
					Name:      "To Do",
					ProjectID: projectID,
				})
				require.NoError(t, err)
				return col.ID
			},
		},
		{
			name:     "invalid ID",
			columnID: 0,
			newName:  "Backlog",
			wantErr:  true,
			errType:  ErrInvalidColumnID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				columnID := tt.columnID
				if tt.setupFn != nil {
					columnID = tt.setupFn(db, d, dbType)
				}

				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)
				err = svc.UpdateColumnName(context.Background(), columnID, tt.newName)

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

func TestDeleteColumn(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column")

		err = env.Svc.DeleteColumn(env.Ctx, created.ID)
		require.NoError(t, err)

		// Verify column is deleted
		_, err = env.Svc.GetColumnByID(env.Ctx, created.ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestDeleteColumn_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		columnID int
		wantErr  bool
		errType  error
		setupFn  func(*sql.DB, fixtures.Dialect, database.DatabaseType) int // Returns column ID if needed
	}{
		{
			name:     "invalid ID",
			columnID: 0,
			wantErr:  true,
			errType:  ErrInvalidColumnID,
		},
		{
			name:    "column has tasks",
			wantErr: true,
			errType: ErrColumnHasTasks,
			setupFn: func(db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) int {
				projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)
				col, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
					Name:      "To Do",
					ProjectID: projectID,
				})
				require.NoError(t, err)
				fixtures.CreateTestTask(t, db, d, col.ID, "Test Task")
				return col.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				columnID := tt.columnID
				if tt.setupFn != nil {
					columnID = tt.setupFn(db, d, dbType)
				}

				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)
				err = svc.DeleteColumn(context.Background(), columnID)

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

func TestDeleteColumn_LinkedListIntegrity(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create three columns: col1 <-> col2 <-> col3
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "In Progress",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		col3, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "Done",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		// Delete middle column (col2)
		err = env.Svc.DeleteColumn(env.Ctx, col2.ID)
		require.NoError(t, err, "Failed to delete column 2")

		// Verify linked list is repaired: col1 <-> col3

		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err)
		col3Updated, err := env.Svc.GetColumnByID(env.Ctx, col3.ID)
		require.NoError(t, err)

		require.NotNil(t, col1Updated.NextID)
		assert.Equal(t, col3.ID, *col1Updated.NextID)

		require.NotNil(t, col3Updated.PrevID)
		assert.Equal(t, col1.ID, *col3Updated.PrevID)
	})
}

func TestCreateColumn_WithHoldsReadyTasks(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateColumnRequest{
			Name:            "To Do",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: true,
		}

		result, err := env.Svc.CreateColumn(env.Ctx, req)
		require.NoError(t, err)

		assert.True(t, result.HoldsReadyTasks)
	})
}

func TestCreateColumn_HoldsReadyTasks_ClearsPrevious(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create first column with HoldsReadyTasks = true
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:            "To Do",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: true,
		})
		require.NoError(t, err, "Failed to create column 1")

		require.True(t, col1.HoldsReadyTasks)

		// Create second column with HoldsReadyTasks = true
		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:            "In Progress",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: true,
		})
		require.NoError(t, err, "Failed to create column 2")

		assert.True(t, col2.HoldsReadyTasks)

		// Verify col1 is no longer the ready column
		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err, "Failed to get updated col1")

		assert.False(t, col1Updated.HoldsReadyTasks)
	})
}

func TestSetHoldsReadyTasks_Success(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create two columns (neither ready)
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:            "To Do",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: false,
		})
		require.NoError(t, err, "Failed to create column 1")

		_, err = env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:            "Done",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: false,
		})
		require.NoError(t, err, "Failed to create column 2")

		// Set col1 as ready
		updated, err := env.Svc.SetHoldsReadyTasks(env.Ctx, col1.ID)
		require.NoError(t, err)

		assert.True(t, updated.HoldsReadyTasks)
	})
}

func TestSetHoldsReadyTasks_TransfersFromPrevious(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create col1 as ready
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:            "To Do",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: true,
		})
		require.NoError(t, err, "Failed to create column 1")

		// Create col2 as not ready
		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:            "In Progress",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: false,
		})
		require.NoError(t, err, "Failed to create column 2")

		// Attempt to set col2 as ready - should fail because col1 already holds ready tasks
		_, err = env.Svc.SetHoldsReadyTasks(env.Ctx, col2.ID)
		require.Error(t, err)

		// Verify error message includes the existing column info
		assert.ErrorContains(t, err, "To Do")

		// Verify col1 still holds ready tasks
		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err, "Failed to get col1")

		assert.True(t, col1Updated.HoldsReadyTasks)

		// Verify col2 does not hold ready tasks
		col2Updated, err := env.Svc.GetColumnByID(env.Ctx, col2.ID)
		require.NoError(t, err, "Failed to get col2")

		assert.False(t, col2Updated.HoldsReadyTasks)
	})
}

func TestGetColumnByID_IncludesHoldsReadyTasks(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create column with HoldsReadyTasks = true
		created, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:            "To Do",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: true,
		})
		require.NoError(t, err, "Failed to create column")

		// Fetch via GetColumnByID
		result, err := env.Svc.GetColumnByID(env.Ctx, created.ID)
		require.NoError(t, err)

		assert.True(t, result.HoldsReadyTasks)
	})
}

func TestGetColumnsByProject_IncludesHoldsReadyTasks(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create one ready column and one not ready
		_, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:            "To Do",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: true,
		})
		require.NoError(t, err, "Failed to create ready column")

		_, err = env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:            "Done",
			ProjectID:       env.ProjectID,
			HoldsReadyTasks: false,
		})
		require.NoError(t, err, "Failed to create non-ready column")

		// Fetch all columns
		results, err := env.Svc.GetColumnsByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		require.Len(t, results, 2)

		// Verify first column (To Do) is ready
		if results[0].Name == "To Do" {
			assert.True(t, results[0].HoldsReadyTasks)
		}

		// Verify second column (Done) is not ready
		if results[1].Name == "Done" {
			assert.False(t, results[1].HoldsReadyTasks)
		}
	})
}

func TestSetHoldsReadyTasks_InvalidColumnID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.SetHoldsReadyTasks(env.Ctx, 0)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestSetHoldsReadyTasks_ColumnNotFound(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.SetHoldsReadyTasks(env.Ctx, 999)

		require.Error(t, err)

		// Should get a wrapped sql.ErrNoRows
		assert.True(t, errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows"))
	})
}

func TestCreateColumn_OnlyOneReadyPerProject(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		ctx := context.Background()

		// Manually insert two columns with holds_ready_tasks = true
		// This should violate the unique partial index constraint
		_, err := db.ExecContext(ctx,
			fmt.Sprintf("INSERT INTO columns (name, project_id, holds_ready_tasks) VALUES (%s, %s, %s)", d.Placeholder(1), d.Placeholder(2), d.Placeholder(3)),
			"To Do", projectID, true)
		require.NoError(t, err, "Failed to insert first ready column")

		_, err = db.ExecContext(ctx,
			fmt.Sprintf("INSERT INTO columns (name, project_id, holds_ready_tasks) VALUES (%s, %s, %s)", d.Placeholder(1), d.Placeholder(2), d.Placeholder(3)),
			"Review", projectID, true)

		require.Error(t, err)

		// Should get a constraint violation error
		assert.True(t, strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "constraint"))
	})
}

func TestGetColumnByID_NegativeID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetColumnByID(env.Ctx, -1)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestGetColumnsByProject_NegativeProjectID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.GetColumnsByProject(env.Ctx, -1)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidProjectID)
	})
}

func TestGetColumnsByProject_NonExistentProject(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Non-existent project should return empty list, not error
		results, err := env.Svc.GetColumnsByProject(env.Ctx, 999999)
		require.NoError(t, err)

		assert.Len(t, results, 0)
	})
}

func TestUpdateColumnName_NegativeID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		err := env.Svc.UpdateColumnName(env.Ctx, -1, "New Name")

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestUpdateColumnName_NonExistentColumn(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		err := env.Svc.UpdateColumnName(env.Ctx, 999999, "New Name")

		require.Error(t, err)

		assert.ErrorContains(t, err, "failed to get column")
	})
}

func TestUpdateColumnName_NameTooLong(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column")

		// Create a name that's 51 characters long
		longName := strings.Repeat("a", 51)

		err = env.Svc.UpdateColumnName(env.Ctx, created.ID, longName)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrNameTooLong)
	})
}

func TestDeleteColumn_NegativeID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		err := env.Svc.DeleteColumn(env.Ctx, -1)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestDeleteColumn_NonExistentColumn(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		err := env.Svc.DeleteColumn(env.Ctx, 999999)

		require.Error(t, err)

		assert.ErrorContains(t, err, "failed to get column info")
	})
}

func TestCreateColumn_InvalidAfterID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		afterID int
		wantErr bool
		errType error
		setupFn func(*sql.DB, fixtures.Dialect, database.DatabaseType) int // Returns project ID
	}{
		{
			name:    "zero after ID",
			afterID: 0,
			wantErr: true,
			errType: ErrInvalidColumnID,
			setupFn: func(db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) int {
				return fixtures.CreateBareProject(t, db, d, "Test Project")
			},
		},
		{
			name:    "negative after ID",
			afterID: -1,
			wantErr: true,
			errType: ErrInvalidColumnID,
			setupFn: func(db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) int {
				return fixtures.CreateBareProject(t, db, d, "Test Project")
			},
		},
		{
			name:    "non-existent after ID",
			afterID: 999999,
			wantErr: false, // Service doesn't validate existence of afterID before attempting operations
			setupFn: func(db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) int {
				return fixtures.CreateBareProject(t, db, d, "Test Project")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
				projectID := tt.setupFn(db, d, dbType)
				svc, err := NewService(db, dbType, nil)
				require.NoError(t, err)

				_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
					Name:      "Test Column",
					ProjectID: projectID,
					AfterID:   &tt.afterID,
				})

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

func TestCreateColumn_AfterColumnFromDifferentProject(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		// Create two projects
		projectID1 := fixtures.CreateBareProject(t, db, d, "Test Project")
		projectID2 := fixtures.CreateBareProject(t, db, d, "Test Project")

		// Create column in project 1
		col1, err := svc.CreateColumn(ctx, CreateColumnRequest{
			Name:      "Column 1",
			ProjectID: projectID1,
		})
		require.NoError(t, err, "Failed to create column 1")

		// Try to create column in project 2 after column from project 1
		// This should succeed because the service doesn't validate project matching
		afterID := col1.ID
		col2, err := svc.CreateColumn(ctx, CreateColumnRequest{
			Name:      "Column 2",
			ProjectID: projectID2,
			AfterID:   &afterID,
		})
		require.NoError(t, err)

		// Verify the column was created successfully
		require.NotNil(t, col2)

		// The column should be in project 2
		assert.Equal(t, projectID2, col2.ProjectID)
	})
}

func TestDeleteColumn_FirstColumn(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create three columns: col1 <-> col2 <-> col3
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "In Progress",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		col3, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "Done",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		// Delete first column (col1)
		err = env.Svc.DeleteColumn(env.Ctx, col1.ID)
		require.NoError(t, err, "Failed to delete column 1")

		// Verify linked list is repaired: col2 <-> col3

		col2Updated, err := env.Svc.GetColumnByID(env.Ctx, col2.ID)
		require.NoError(t, err)
		col3Updated, err := env.Svc.GetColumnByID(env.Ctx, col3.ID)
		require.NoError(t, err)

		assert.Nil(t, col2Updated.PrevID)

		require.NotNil(t, col2Updated.NextID)
		assert.Equal(t, col3.ID, *col2Updated.NextID)

		require.NotNil(t, col3Updated.PrevID)
		assert.Equal(t, col2.ID, *col3Updated.PrevID)
	})
}

func TestDeleteColumn_LastColumn(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create three columns: col1 <-> col2 <-> col3
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "In Progress",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		col3, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "Done",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		// Delete last column (col3)
		err = env.Svc.DeleteColumn(env.Ctx, col3.ID)
		require.NoError(t, err, "Failed to delete column 3")

		// Verify linked list is repaired: col1 <-> col2

		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err)
		col2Updated, err := env.Svc.GetColumnByID(env.Ctx, col2.ID)
		require.NoError(t, err)

		require.NotNil(t, col1Updated.NextID)
		assert.Equal(t, col2.ID, *col1Updated.NextID)

		require.NotNil(t, col2Updated.PrevID)
		assert.Equal(t, col1.ID, *col2Updated.PrevID)

		assert.Nil(t, col2Updated.NextID)
	})
}

func TestDeleteColumn_OnlyColumn(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create single column
		col, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err)

		// Delete it
		err = env.Svc.DeleteColumn(env.Ctx, col.ID)
		require.NoError(t, err, "Failed to delete only column")

		// Verify project has no columns
		columns, err := env.Svc.GetColumnsByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err, "Failed to get columns")

		assert.Len(t, columns, 0)
	})
}

func TestSetHoldsReadyTasks_NegativeColumnID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.SetHoldsReadyTasks(env.Ctx, -1)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestCreateColumn_MultipleReadyColumns_DifferentProjects(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		// Create two projects
		projectID1 := fixtures.CreateBareProject(t, db, d, "Test Project")
		projectID2 := fixtures.CreateBareProject(t, db, d, "Test Project")

		// .*
		col1, err := svc.CreateColumn(ctx, CreateColumnRequest{
			Name:            "To Do",
			ProjectID:       projectID1,
			HoldsReadyTasks: true,
		})
		require.NoError(t, err, "Failed to create ready column in project 1")

		require.True(t, col1.HoldsReadyTasks)

		// Create ready column in project 2 - should succeed
		col2, err := svc.CreateColumn(ctx, CreateColumnRequest{
			Name:            "To Do",
			ProjectID:       projectID2,
			HoldsReadyTasks: true,
		})
		require.NoError(t, err, "Failed to create ready column in project 2")

		require.True(t, col2.HoldsReadyTasks)

		// Verify both columns are still ready
		col1Check, err := svc.GetColumnByID(ctx, col1.ID)
		require.NoError(t, err)
		col2Check, err := svc.GetColumnByID(ctx, col2.ID)
		require.NoError(t, err)

		assert.True(t, col1Check.HoldsReadyTasks)

		assert.True(t, col2Check.HoldsReadyTasks)
	})
}

func TestSetHoldsCompletedTasks_NegativeColumnID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.SetHoldsCompletedTasks(env.Ctx, -1, false)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestCreateColumn_MultipleCompletedColumns_DifferentProjects(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		// Create two projects
		projectID1 := fixtures.CreateBareProject(t, db, d, "Test Project")
		projectID2 := fixtures.CreateBareProject(t, db, d, "Test Project")

		// .*
		col1, err := svc.CreateColumn(ctx, CreateColumnRequest{
			Name:                "Done",
			ProjectID:           projectID1,
			HoldsCompletedTasks: true,
		})
		require.NoError(t, err, "Failed to create completed column in project 1")

		require.True(t, col1.HoldsCompletedTasks)

		// Create completed column in project 2 - should succeed
		col2, err := svc.CreateColumn(ctx, CreateColumnRequest{
			Name:                "Done",
			ProjectID:           projectID2,
			HoldsCompletedTasks: true,
		})
		require.NoError(t, err, "Failed to create completed column in project 2")

		require.True(t, col2.HoldsCompletedTasks)

		// Verify both columns are still completed
		col1Check, err := svc.GetColumnByID(ctx, col1.ID)
		require.NoError(t, err)
		col2Check, err := svc.GetColumnByID(ctx, col2.ID)
		require.NoError(t, err)

		assert.True(t, col1Check.HoldsCompletedTasks)

		assert.True(t, col2Check.HoldsCompletedTasks)
	})
}

func TestCreateColumn_WithHoldsInProgressTasks(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateColumnRequest{
			Name:                 "In Progress",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: true,
		}

		result, err := env.Svc.CreateColumn(env.Ctx, req)
		require.NoError(t, err)

		assert.True(t, result.HoldsInProgressTasks)
	})
}

func TestCreateColumn_HoldsInProgressTasks_ClearsPrevious(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create first column with HoldsInProgressTasks = true
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                 "In Progress",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: true,
		})
		require.NoError(t, err, "Failed to create column 1")

		require.True(t, col1.HoldsInProgressTasks)

		// Create second column with HoldsInProgressTasks = true
		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                 "Doing",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: true,
		})
		require.NoError(t, err, "Failed to create column 2")

		assert.True(t, col2.HoldsInProgressTasks)

		// Verify col1 is no longer the in-progress column
		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err, "Failed to get updated col1")

		assert.False(t, col1Updated.HoldsInProgressTasks)
	})
}

func TestSetHoldsInProgressTasks_Success(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create two columns (neither in-progress)
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                 "In Progress",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: false,
		})
		require.NoError(t, err, "Failed to create column 1")

		_, err = env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                 "Done",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: false,
		})
		require.NoError(t, err, "Failed to create column 2")

		// Set col1 as in-progress
		updated, err := env.Svc.SetHoldsInProgressTasks(env.Ctx, col1.ID)
		require.NoError(t, err)

		assert.True(t, updated.HoldsInProgressTasks)
	})
}

func TestSetHoldsInProgressTasks_FailsWhenExists(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create col1 as in-progress
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                 "In Progress",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: true,
		})
		require.NoError(t, err, "Failed to create column 1")

		// Create col2 as not in-progress
		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                 "Doing",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: false,
		})
		require.NoError(t, err, "Failed to create column 2")

		// Attempt to set col2 as in-progress - should fail
		_, err = env.Svc.SetHoldsInProgressTasks(env.Ctx, col2.ID)
		require.Error(t, err)

		// Verify error message includes the existing column info
		assert.ErrorContains(t, err, "In Progress")

		// Verify col1 still holds in-progress tasks
		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err, "Failed to get col1")

		assert.True(t, col1Updated.HoldsInProgressTasks)

		// Verify col2 does not hold in-progress tasks
		col2Updated, err := env.Svc.GetColumnByID(env.Ctx, col2.ID)
		require.NoError(t, err, "Failed to get col2")

		assert.False(t, col2Updated.HoldsInProgressTasks)
	})
}

func TestSetHoldsInProgressTasks_InvalidColumnID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.SetHoldsInProgressTasks(env.Ctx, 0)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestSetHoldsInProgressTasks_NegativeColumnID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.SetHoldsInProgressTasks(env.Ctx, -1)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestSetHoldsInProgressTasks_ColumnNotFound(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.SetHoldsInProgressTasks(env.Ctx, 999)

		require.Error(t, err)

		// Should get a wrapped sql.ErrNoRows
		assert.True(t, errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows"))
	})
}

func TestGetColumnByID_IncludesHoldsInProgressTasks(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create column with HoldsInProgressTasks = true
		created, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                 "In Progress",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: true,
		})
		require.NoError(t, err, "Failed to create column")

		// Fetch via GetColumnByID
		result, err := env.Svc.GetColumnByID(env.Ctx, created.ID)
		require.NoError(t, err)

		assert.True(t, result.HoldsInProgressTasks)
	})
}

func TestGetColumnsByProject_IncludesHoldsInProgressTasks(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create one in-progress column and one not in-progress
		_, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                 "In Progress",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: true,
		})
		require.NoError(t, err, "Failed to create in-progress column")

		_, err = env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                 "Done",
			ProjectID:            env.ProjectID,
			HoldsInProgressTasks: false,
		})
		require.NoError(t, err, "Failed to create non-in-progress column")

		// Fetch all columns
		results, err := env.Svc.GetColumnsByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		require.Len(t, results, 2)

		// Verify first column (In Progress) is in-progress
		if results[0].Name == "In Progress" {
			assert.True(t, results[0].HoldsInProgressTasks)
		}

		// Verify second column (Done) is not in-progress
		if results[1].Name == "Done" {
			assert.False(t, results[1].HoldsInProgressTasks)
		}
	})
}

func TestCreateColumn_MultipleInProgressColumns_DifferentProjects(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		svc, err := NewService(db, dbType, nil)
		require.NoError(t, err)
		ctx := context.Background()

		// Create two projects
		projectID1 := fixtures.CreateBareProject(t, db, d, "Test Project")
		projectID2 := fixtures.CreateBareProject(t, db, d, "Test Project")

		// .*
		col1, err := svc.CreateColumn(ctx, CreateColumnRequest{
			Name:                 "In Progress",
			ProjectID:            projectID1,
			HoldsInProgressTasks: true,
		})
		require.NoError(t, err, "Failed to create in-progress column in project 1")

		require.True(t, col1.HoldsInProgressTasks)

		// Create in-progress column in project 2 - should succeed
		col2, err := svc.CreateColumn(ctx, CreateColumnRequest{
			Name:                 "In Progress",
			ProjectID:            projectID2,
			HoldsInProgressTasks: true,
		})
		require.NoError(t, err, "Failed to create in-progress column in project 2")

		require.True(t, col2.HoldsInProgressTasks)

		// Verify both columns are still in-progress
		col1Check, err := svc.GetColumnByID(ctx, col1.ID)
		require.NoError(t, err)
		col2Check, err := svc.GetColumnByID(ctx, col2.ID)
		require.NoError(t, err)

		assert.True(t, col1Check.HoldsInProgressTasks)

		assert.True(t, col2Check.HoldsInProgressTasks)
	})
}

func TestCreateColumn_NameExactly50Characters(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create a name that's exactly 50 characters
		name50 := strings.Repeat("a", 50)

		req := CreateColumnRequest{
			Name:      name50,
			ProjectID: env.ProjectID,
		}

		result, err := env.Svc.CreateColumn(env.Ctx, req)
		require.NoError(t, err)

		assert.Equal(t, name50, result.Name)
	})
}

func TestCreateColumn_NonExistentProject(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Try to create column in non-existent project
		_, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: 999999,
		})

		require.Error(t, err)

		// Error should be related to foreign key constraint
		assert.True(t, strings.Contains(err.Error(), "failed to create column"))
	})
}

func TestUpdateColumnName_Exact50Characters(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		created, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: env.ProjectID,
		})
		require.NoError(t, err, "Failed to create column")

		// Create a name that's exactly 50 characters
		name50 := strings.Repeat("a", 50)

		err = env.Svc.UpdateColumnName(env.Ctx, created.ID, name50)
		require.NoError(t, err)

		// Verify update
		updated, err := env.Svc.GetColumnByID(env.Ctx, created.ID)
		require.NoError(t, err, "Failed to get updated column")

		assert.Equal(t, name50, updated.Name)
	})
}

func TestCreateColumn_AllSpecialFlagsTrue(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Try to create column with all special flags set to true
		req := CreateColumnRequest{
			Name:                 "Multi-purpose",
			ProjectID:            env.ProjectID,
			HoldsReadyTasks:      true,
			HoldsCompletedTasks:  true,
			HoldsInProgressTasks: true,
		}

		result, err := env.Svc.CreateColumn(env.Ctx, req)
		require.NoError(t, err)

		// Verify all flags are set
		assert.True(t, result.HoldsReadyTasks)

		assert.True(t, result.HoldsCompletedTasks)

		assert.True(t, result.HoldsInProgressTasks)
	})
}

func TestCreateColumn_ProjectIDMaxInt(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Try to create column with max int project ID (won't exist)
		_, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:      "To Do",
			ProjectID: 2147483647, // max int32
		})

		require.Error(t, err)

		// Error should be related to foreign key constraint
		assert.True(t, strings.Contains(err.Error(), "failed to create column"))
	})
}

func TestCreateColumn_WithHoldsCompletedTasks(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		req := CreateColumnRequest{
			Name:                "Done",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: true,
		}

		result, err := env.Svc.CreateColumn(env.Ctx, req)
		require.NoError(t, err)

		assert.True(t, result.HoldsCompletedTasks)
	})
}

func TestCreateColumn_HoldsCompletedTasks_FailsWhenExists(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create first column with HoldsCompletedTasks = true
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Done",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: true,
		})
		require.NoError(t, err, "Failed to create column 1")

		require.True(t, col1.HoldsCompletedTasks)

		// Create second column with HoldsCompletedTasks = true (should fail)
		_, err = env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Archive",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: true,
		})

		require.Error(t, err)

		assert.True(t, errors.Is(err, ErrCompletedColumnExists) || strings.Contains(err.Error(), "completed column already exists"))
	})
}

func TestSetHoldsCompletedTasks_Success(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create two columns (neither completed)
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Done",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: false,
		})
		require.NoError(t, err, "Failed to create column 1")

		_, err = env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Archive",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: false,
		})
		require.NoError(t, err, "Failed to create column 2")

		// Set col1 as completed
		updated, err := env.Svc.SetHoldsCompletedTasks(env.Ctx, col1.ID, false)
		require.NoError(t, err)

		assert.True(t, updated.HoldsCompletedTasks)
	})
}

func TestSetHoldsCompletedTasks_FailsWhenExistsWithoutForce(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create col1 as completed
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Done",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: true,
		})
		require.NoError(t, err, "Failed to create column 1")

		// Create col2 as not completed
		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Archive",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: false,
		})
		require.NoError(t, err, "Failed to create column 2")

		// Try to set col2 as completed without force (should fail)
		_, err = env.Svc.SetHoldsCompletedTasks(env.Ctx, col2.ID, false)

		require.Error(t, err)

		assert.True(t, errors.Is(err, ErrCompletedColumnExists) || strings.Contains(err.Error(), "completed column already exists"))

		// Verify col1 is still completed
		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err, "Failed to get col1")

		assert.True(t, col1Updated.HoldsCompletedTasks)
	})
}

func TestSetHoldsCompletedTasks_SucceedsWithForce(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create col1 as completed
		col1, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Done",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: true,
		})
		require.NoError(t, err, "Failed to create column 1")

		// Create col2 as not completed
		col2, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Archive",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: false,
		})
		require.NoError(t, err, "Failed to create column 2")

		// Set col2 as completed with force (should succeed)
		updated, err := env.Svc.SetHoldsCompletedTasks(env.Ctx, col2.ID, true)
		require.NoError(t, err)

		assert.True(t, updated.HoldsCompletedTasks)

		// Verify col1 is no longer completed
		col1Updated, err := env.Svc.GetColumnByID(env.Ctx, col1.ID)
		require.NoError(t, err, "Failed to get col1")

		assert.False(t, col1Updated.HoldsCompletedTasks)
	})
}

func TestGetColumnByID_IncludesHoldsCompletedTasks(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create column with HoldsCompletedTasks = true
		created, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Done",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: true,
		})
		require.NoError(t, err, "Failed to create column")

		// Fetch via GetColumnByID
		result, err := env.Svc.GetColumnByID(env.Ctx, created.ID)
		require.NoError(t, err)

		assert.True(t, result.HoldsCompletedTasks)
	})
}

func TestGetColumnsByProject_IncludesHoldsCompletedTasks(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		// Create one completed column and one not completed
		_, err := env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Done",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: true,
		})
		require.NoError(t, err, "Failed to create completed column")

		_, err = env.Svc.CreateColumn(env.Ctx, CreateColumnRequest{
			Name:                "Todo",
			ProjectID:           env.ProjectID,
			HoldsCompletedTasks: false,
		})
		require.NoError(t, err, "Failed to create non-completed column")

		// Fetch all columns
		results, err := env.Svc.GetColumnsByProject(env.Ctx, env.ProjectID)
		require.NoError(t, err)

		require.Len(t, results, 2)

		// Verify first column (Done) is completed
		if results[0].Name == "Done" {
			assert.True(t, results[0].HoldsCompletedTasks)
		}

		// Verify second column (Todo) is not completed
		if results[1].Name == "Todo" {
			assert.False(t, results[1].HoldsCompletedTasks)
		}
	})
}

func TestSetHoldsCompletedTasks_InvalidColumnID(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.SetHoldsCompletedTasks(env.Ctx, 0, false)

		require.Error(t, err)

		assert.ErrorIs(t, err, ErrInvalidColumnID)
	})
}

func TestSetHoldsCompletedTasks_ColumnNotFound(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		env := setupTestEnv(t, db, d, dbType)

		_, err := env.Svc.SetHoldsCompletedTasks(env.Ctx, 999, false)

		require.Error(t, err)

		// Should get a wrapped sql.ErrNoRows
		assert.True(t, errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows"))
	})
}

func TestCreateColumn_OnlyOneCompletedPerProject(t *testing.T) {
	t.Parallel()

	fixtures.RunDatabaseTests(t, func(t *testing.T, db *sql.DB, d fixtures.Dialect, dbType database.DatabaseType) {
		projectID := fixtures.CreateBareProject(t, db, d, "Test Project")
		ctx := context.Background()

		// Manually insert two columns with holds_completed_tasks = true
		// This should violate the unique partial index constraint
		_, err := db.ExecContext(ctx,
			fmt.Sprintf("INSERT INTO columns (name, project_id, holds_completed_tasks) VALUES (%s, %s, %s)", d.Placeholder(1), d.Placeholder(2), d.Placeholder(3)),
			"Done", projectID, true)
		require.NoError(t, err, "Failed to insert first completed column")

		_, err = db.ExecContext(ctx,
			fmt.Sprintf("INSERT INTO columns (name, project_id, holds_completed_tasks) VALUES (%s, %s, %s)", d.Placeholder(1), d.Placeholder(2), d.Placeholder(3)),
			"Archive", projectID, true)

		require.Error(t, err)

		// Should get a constraint violation error
		assert.True(t, strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "constraint"))
	})
}
