package label

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// newTestService creates a new service for testing (panics on error since tests use valid SQLite)
func newTestService(t *testing.T, db *sql.DB) Service {
	t.Helper()
	svc, err := NewService(db, database.SQLite, nil)
	require.NoError(t, err, "failed to create test service")
	return svc
}

// createTestProject creates a test project and returns its ID
func createTestProject(t *testing.T, db *sql.DB) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO projects (name, description) VALUES (?, ?)", "Test Project", "Test Description")
	require.NoError(t, err)
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

// createTestTask creates a test task and returns its ID
func createTestTask(t *testing.T, db *sql.DB, projectID int) int {
	t.Helper()
	// First create a column for the task
	columnResult, err := db.ExecContext(context.Background(), "INSERT INTO columns (project_id, name) VALUES (?, ?)", projectID, "Default")
	require.NoError(t, err)
	columnID, _ := columnResult.LastInsertId()

	// Create task in that column
	result, err := db.ExecContext(context.Background(), "INSERT INTO tasks (column_id, title, description, position) VALUES (?, ?, ?, ?)", columnID, "Test Task", "Test Description", 0)
	require.NoError(t, err)
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

// attachLabelToTask attaches a label to a task
func attachLabelToTask(t *testing.T, db *sql.DB, taskID, labelID int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), "INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)", taskID, labelID)
	require.NoError(t, err)
}

func TestCreateLabel(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)

	req := CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	}

	result, err := svc.CreateLabel(context.Background(), req)
	require.NoError(t, err)

	require.NotNil(t, result)

	assert.Equal(t, "Bug", result.Name)

	assert.Equal(t, "#FF5733", result.Color)

	assert.Equal(t, projectID, result.ProjectID)

	assert.NotZero(t, result.ID)
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

			db := testutil.SetupTestDB(t)

			projectID := tt.projectID
			if tt.setupFn != nil {
				projectID = tt.setupFn(db)
			}

			svc := newTestService(t, db)
			req := CreateLabelRequest{
				ProjectID: projectID,
				Name:      tt.labelName,
				Color:     tt.color,
			}

			_, err := svc.CreateLabel(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

func TestCreateLabel_InvalidColor(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)

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

			require.Error(t, err)

			assert.ErrorIs(t, err, ErrInvalidColor)
		})
	}
}

func TestCreateLabel_InvalidProjectID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	req := CreateLabelRequest{
		ProjectID: 0, // Invalid
		Name:      "Bug",
		Color:     "#FF5733",
	}

	_, err := svc.CreateLabel(context.Background(), req)

	require.Error(t, err)

	assert.ErrorIs(t, err, ErrInvalidProjectID)
}

func TestGetLabelsByProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two labels
	_, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	_, err = svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Feature",
		Color:     "#33FF57",
	})
	require.NoError(t, err)

	results, err := svc.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	require.Len(t, results, 2)

	assert.Equal(t, "Bug", results[0].Name)

	assert.Equal(t, "Feature", results[1].Name)
}

func TestGetLabelsByProject_Empty(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)

	results, err := svc.GetLabelsByProject(context.Background(), projectID)
	require.NoError(t, err)

	assert.Len(t, results, 0)
}

func TestGetLabelsByProject_InvalidProjectID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetLabelsByProject(context.Background(), 0)

	require.Error(t, err)

	assert.ErrorIs(t, err, ErrInvalidProjectID)
}

func TestGetLabelsForTask(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	taskID := createTestTask(t, db, projectID)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two labels and attach them to the task
	label1, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	label2, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Critical",
		Color:     "#FF0000",
	})
	require.NoError(t, err)

	attachLabelToTask(t, db, taskID, label1.ID)
	attachLabelToTask(t, db, taskID, label2.ID)

	results, err := svc.GetLabelsForTask(ctx, taskID)
	require.NoError(t, err)

	require.Len(t, results, 2)

	// Check that both labels are present (order may vary)
	labelNames := map[string]bool{}
	for _, label := range results {
		labelNames[label.Name] = true
	}

	assert.True(t, labelNames["Bug"])

	assert.True(t, labelNames["Critical"])
}

func TestGetLabelsForTask_Empty(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	taskID := createTestTask(t, db, projectID)
	svc := newTestService(t, db)

	results, err := svc.GetLabelsForTask(context.Background(), taskID)
	require.NoError(t, err)

	assert.Len(t, results, 0)
}

func TestGetLabelsForTask_InvalidTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetLabelsForTask(context.Background(), 0)

	require.Error(t, err)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestUpdateLabel(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a label
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

	// Verify update
	labels, err := svc.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	require.Len(t, labels, 1)

	assert.Equal(t, "Critical Bug", labels[0].Name)

	assert.Equal(t, "#FF0000", labels[0].Color)
}

func TestUpdateLabel_OnlyName(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a label
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

	// Verify update - color should remain unchanged
	labels, err := svc.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Bug", labels[0].Name)

	assert.Equal(t, "#FF5733", labels[0].Color)
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
				label, _ := newTestService(t, db).CreateLabel(context.Background(), CreateLabelRequest{
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
				label, _ := newTestService(t, db).CreateLabel(context.Background(), CreateLabelRequest{
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

			db := testutil.SetupTestDB(t)

			labelID := tt.labelID
			if tt.setupFn != nil {
				labelID = tt.setupFn(db)
			}

			svc := newTestService(t, db)
			req := UpdateLabelRequest{
				ID:    labelID,
				Name:  tt.newName,
				Color: tt.newColor,
			}

			err := svc.UpdateLabel(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
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

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a label
	created, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	err = svc.DeleteLabel(ctx, created.ID)
	require.NoError(t, err)

	// Verify label is deleted
	labels, err := svc.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	assert.Len(t, labels, 0)
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			err := svc.DeleteLabel(context.Background(), tt.labelID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

func TestDeleteLabel_CascadeToTaskLabels(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	taskID := createTestTask(t, db, projectID)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a label and attach it to a task
	created, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	attachLabelToTask(t, db, taskID, created.ID)

	// Verify label is attached
	taskLabels, err := svc.GetLabelsForTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, taskLabels, 1)

	// Delete the label
	err = svc.DeleteLabel(ctx, created.ID)
	require.NoError(t, err)

	// Verify task_labels entry is also deleted (cascade)
	taskLabels, err = svc.GetLabelsForTask(ctx, taskID)
	require.NoError(t, err)
	assert.Len(t, taskLabels, 0)
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
			wantErr:   nil, // Will be caught by database constraint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			req := CreateLabelRequest{
				ProjectID: tt.projectID,
				Name:      tt.labelName,
				Color:     tt.color,
			}

			_, err := svc.CreateLabel(context.Background(), req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				// For non-existent project, we expect some error (likely foreign key constraint)
				assert.Error(t, err)
			}
		})
	}
}

func TestCreateLabel_DuplicateNames(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create first label
	_, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	// Try to create second label with same name
	_, err = svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#00FF00",
	})

	// Database has UNIQUE(name, project_id) constraint, should return error
	require.Error(t, err)

	// Verify the error message mentions the constraint or duplicate
	errMsg := err.Error()
	assert.True(t, strings.Contains(errMsg, "label creation error") || strings.Contains(errMsg, "already exists"))
}

func TestCreateLabel_DuplicateNames_DifferentProjects(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID1 := createTestProject(t, db)
	projectID2 := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create label in first project
	_, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID1,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	// Create label with same name in different project (should succeed)
	_, err = svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID2,
		Name:      "Bug",
		Color:     "#00FF00",
	})
	assert.NoError(t, err)
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
			labelName: "\u30d0\u30b0",
			shouldErr: false,
		},
		{
			name:      "emoji",
			labelName: "\U0001f41b Bug",
			shouldErr: false,
		},
		{
			name:      "special symbols",
			labelName: "Bug: P0 [Critical]",
			shouldErr: false,
		},
		{
			name:      "mixed unicode and emoji",
			labelName: "\U0001f680 \u65b0\u6a5f\u80fd",
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

			db := testutil.SetupTestDB(t)

			projectID := createTestProject(t, db)
			svc := newTestService(t, db)

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
	}
}

func TestGetLabelsByProject_NegativeProjectID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetLabelsByProject(context.Background(), -1)

	require.Error(t, err)

	assert.ErrorIs(t, err, ErrInvalidProjectID)
}

func TestGetLabelsByProject_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Query non-existent project (should return empty list, not error)
	labels, err := svc.GetLabelsByProject(context.Background(), 999999)
	assert.NoError(t, err)

	assert.Len(t, labels, 0)
}

func TestGetLabelsForTask_NegativeTaskID(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	_, err := svc.GetLabelsForTask(context.Background(), -1)

	require.Error(t, err)

	assert.ErrorIs(t, err, ErrInvalidTaskID)
}

func TestGetLabelsForTask_NonExistentTask(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Query non-existent task (should return empty list, not error)
	labels, err := svc.GetLabelsForTask(context.Background(), 999999)
	assert.NoError(t, err)

	assert.Len(t, labels, 0)
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)
			req := UpdateLabelRequest{
				ID:    tt.labelID,
				Name:  tt.newName,
				Color: tt.newColor,
			}

			err := svc.UpdateLabel(context.Background(), req)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUpdateLabel_NonExistentLabel(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	req := UpdateLabelRequest{
		ID:   999999,
		Name: ptrStr("Updated Bug"),
	}

	err := svc.UpdateLabel(context.Background(), req)

	require.Error(t, err)

	assert.ErrorIs(t, err, ErrLabelNotFound)
}

func TestUpdateLabel_NameTooLong(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a label
	created, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	// Try to update with too long name
	longName := ""
	for i := 0; i < 51; i++ {
		longName += "a"
	}

	req := UpdateLabelRequest{
		ID:   created.ID,
		Name: &longName,
	}

	err = svc.UpdateLabel(ctx, req)

	require.Error(t, err)

	assert.ErrorIs(t, err, ErrNameTooLong)
}

func TestUpdateLabel_InvalidColorFormats(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a label
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
			t.Parallel()
			req := UpdateLabelRequest{
				ID:    created.ID,
				Color: &tc.color,
			}

			err := svc.UpdateLabel(ctx, req)

			assert.Error(t, err)

			assert.ErrorIs(t, err, ErrInvalidColor)
		})
	}
}

func TestUpdateLabel_NoFieldsToUpdate(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a label
	created, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	// Update with no fields (should be no-op but succeed)
	req := UpdateLabelRequest{
		ID: created.ID,
	}

	err = svc.UpdateLabel(ctx, req)
	assert.NoError(t, err)

	// Verify nothing changed
	labels, err := svc.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)

	require.Len(t, labels, 1)

	assert.Equal(t, "Bug", labels[0].Name)

	assert.Equal(t, "#FF5733", labels[0].Color)
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

			db := testutil.SetupTestDB(t)

			svc := newTestService(t, db)

			err := svc.DeleteLabel(context.Background(), tt.labelID)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDeleteLabel_NonExistentLabel(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	svc := newTestService(t, db)

	// Try to delete non-existent label (should succeed as per service implementation)
	err := svc.DeleteLabel(context.Background(), 999999)
	assert.NoError(t, err)
}

func TestDeleteLabel_AlreadyDeleted(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create a label
	created, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	// Delete the label once
	err = svc.DeleteLabel(ctx, created.ID)
	require.NoError(t, err)

	// Delete the same label again (should succeed as per idempotent design)
	err = svc.DeleteLabel(ctx, created.ID)
	assert.NoError(t, err)
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

			db := testutil.SetupTestDB(t)

			projectID := createTestProject(t, db)
			svc := newTestService(t, db)

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

			db := testutil.SetupTestDB(t)

			projectID := createTestProject(t, db)
			svc := newTestService(t, db)

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
	}
}

func TestUpdateLabel_DuplicateNameInProject(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)

	projectID := createTestProject(t, db)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two labels
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

	// Try to update label2 to have the same name as label1
	newName := "Bug"
	req := UpdateLabelRequest{
		ID:   label2.ID,
		Name: &newName,
	}

	err = svc.UpdateLabel(ctx, req)

	// Should return error due to unique constraint on (name, project_id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestUpdateLabel_DuplicateName_DifferentProjects(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)
	svc := newTestService(t, db)
	ctx := context.Background()

	// Create two projects
	project1ID := createTestProject(t, db)
	project2ID := createTestProject(t, db)

	// Create label "Bug" in project1
	_, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: project1ID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	// Create label "Feature" in project2
	label2, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: project2ID,
		Name:      "Feature",
		Color:     "#00FF00",
	})
	require.NoError(t, err)

	// Update label2 to "Bug" - should succeed since it's in a different project
	newName := "Bug"
	req := UpdateLabelRequest{
		ID:   label2.ID,
		Name: &newName,
	}

	err = svc.UpdateLabel(ctx, req)
	require.NoError(t, err)

	// Verify the update
	labels, err := svc.GetLabelsByProject(ctx, project2ID)
	require.NoError(t, err)
	require.Len(t, labels, 1)
	assert.Equal(t, "Bug", labels[0].Name)
}

func TestUpdateLabel_SameNameNoChange(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)
	svc := newTestService(t, db)
	ctx := context.Background()

	projectID := createTestProject(t, db)

	// Create label
	label, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	// Update label to its own name (should succeed - not a duplicate)
	sameName := "Bug"
	req := UpdateLabelRequest{
		ID:   label.ID,
		Name: &sameName,
	}

	err = svc.UpdateLabel(ctx, req)
	require.NoError(t, err)
}

func TestUpdateLabel_CaseVariation(t *testing.T) {
	t.Parallel()

	db := testutil.SetupTestDB(t)
	svc := newTestService(t, db)
	ctx := context.Background()

	projectID := createTestProject(t, db)

	// Create label "bug" (lowercase)
	_, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "bug",
		Color:     "#FF5733",
	})
	require.NoError(t, err)

	// Create label "Feature"
	label2, err := svc.CreateLabel(ctx, CreateLabelRequest{
		ProjectID: projectID,
		Name:      "Feature",
		Color:     "#00FF00",
	})
	require.NoError(t, err)

	// Try to update to "BUG" (uppercase)
	// Note: SQLite UNIQUE constraint is case-sensitive by default
	// So "bug" and "BUG" are treated as different values
	newName := "BUG"
	req := UpdateLabelRequest{
		ID:   label2.ID,
		Name: &newName,
	}

	err = svc.UpdateLabel(ctx, req)
	// Should succeed - SQLite treats "bug" and "BUG" as different
	require.NoError(t, err)

	// Verify we now have both "bug" and "BUG"
	labels, err := svc.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, labels, 2)

	names := make(map[string]bool)
	for _, l := range labels {
		names[l.Name] = true
	}
	assert.True(t, names["bug"] && names["BUG"])
}
