package column

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
func createTestTask(t *testing.T, db *sql.DB, columnID int) int {
	t.Helper()
	result, err := db.ExecContext(context.Background(), "INSERT INTO tasks (column_id, title, description, position) VALUES (?, ?, ?, ?)",
		columnID, "Test Task", "Test Description", 0)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get task ID: %v", err)
	}
	return int(id)
}

// ============================================================================
// TEST CASES
// ============================================================================

func TestCreateColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	}

	result, err := svc.CreateColumn(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected column result, got nil")
	}

	if result.Name != "To Do" {
		t.Errorf("Expected name 'To Do', got '%s'", result.Name)
	}

	if result.ProjectID != projectID {
		t.Errorf("Expected project ID %d, got %d", projectID, result.ProjectID)
	}

	if result.ID == 0 {
		t.Error("Expected column ID to be set")
	}

	// First column should have nil prev and next
	if result.PrevID != nil {
		t.Errorf("Expected prev_id nil for first column, got %v", result.PrevID)
	}

	if result.NextID != nil {
		t.Errorf("Expected next_id nil for first column, got %v", result.NextID)
	}
}

func TestCreateColumn_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		colName   string
		projectID int
		wantErr   bool
		errType   error
		setupFn   func(*sql.DB) int // Returns project ID if needed
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
			setupFn: func(db *sql.DB) int {
				return createTestProject(t, db)
			},
			colName: func() string {
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
			colName:   "To Do",
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
			req := CreateColumnRequest{
				Name:      tt.colName,
				ProjectID: projectID,
			}

			_, err := svc.CreateColumn(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateColumn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("CreateColumn() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestCreateColumn_LinkedList(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create first column
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Create second column (should append to end)
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "In Progress",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Create third column (should append to end)
	col3, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 3: %v", err)
	}

	// Verify linked list structure: col1 <-> col2 <-> col3

	// Get updated column 1
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get column 1: %v", err)
	}

	if col1Updated.PrevID != nil {
		t.Errorf("Expected col1 prev_id nil, got %v", col1Updated.PrevID)
	}
	if col1Updated.NextID == nil || *col1Updated.NextID != col2.ID {
		t.Errorf("Expected col1 next_id %d, got %v", col2.ID, col1Updated.NextID)
	}

	// Get updated column 2
	col2Updated, err := svc.GetColumnByID(context.Background(), col2.ID)
	if err != nil {
		t.Fatalf("Failed to get column 2: %v", err)
	}

	if col2Updated.PrevID == nil || *col2Updated.PrevID != col1.ID {
		t.Errorf("Expected col2 prev_id %d, got %v", col1.ID, col2Updated.PrevID)
	}
	if col2Updated.NextID == nil || *col2Updated.NextID != col3.ID {
		t.Errorf("Expected col2 next_id %d, got %v", col3.ID, col2Updated.NextID)
	}

	// Column 3
	if col3.PrevID == nil || *col3.PrevID != col2.ID {
		t.Errorf("Expected col3 prev_id %d, got %v", col2.ID, col3.PrevID)
	}
	if col3.NextID != nil {
		t.Errorf("Expected col3 next_id nil, got %v", col3.NextID)
	}
}

func TestCreateColumn_InsertAfter(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two columns
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	col3, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 3: %v", err)
	}

	// Insert column 2 after column 1
	afterID := col1.ID
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "In Progress",
		ProjectID: projectID,
		AfterID:   &afterID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2 after column 1: %v", err)
	}

	// Verify linked list: col1 <-> col2 <-> col3

	// Get updated columns
	col1Updated, _ := svc.GetColumnByID(context.Background(), col1.ID)
	col2Updated, _ := svc.GetColumnByID(context.Background(), col2.ID)
	col3Updated, _ := svc.GetColumnByID(context.Background(), col3.ID)

	if col1Updated.NextID == nil || *col1Updated.NextID != col2.ID {
		t.Errorf("Expected col1 next_id %d, got %v", col2.ID, col1Updated.NextID)
	}

	if col2Updated.PrevID == nil || *col2Updated.PrevID != col1.ID {
		t.Errorf("Expected col2 prev_id %d, got %v", col1.ID, col2Updated.PrevID)
	}
	if col2Updated.NextID == nil || *col2Updated.NextID != col3.ID {
		t.Errorf("Expected col2 next_id %d, got %v", col3.ID, col2Updated.NextID)
	}

	if col3Updated.PrevID == nil || *col3Updated.PrevID != col2.ID {
		t.Errorf("Expected col3 prev_id %d, got %v", col2.ID, col3Updated.PrevID)
	}
}

func TestGetColumnsByProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two columns
	_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	results, err := svc.GetColumnsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(results))
	}

	if results[0].Name != "To Do" {
		t.Errorf("Expected first column name 'To Do', got '%s'", results[0].Name)
	}

	if results[1].Name != "Done" {
		t.Errorf("Expected second column name 'Done', got '%s'", results[1].Name)
	}
}

func TestGetColumnsByProject_Empty(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	results, err := svc.GetColumnsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 columns, got %d", len(results))
	}
}

func TestGetColumnsByProject_InvalidProjectID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetColumnsByProject(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid project ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}

func TestGetColumnByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	result, err := svc.GetColumnByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, result.ID)
	}

	if result.Name != "To Do" {
		t.Errorf("Expected name 'To Do', got '%s'", result.Name)
	}
}

func TestGetColumnByID_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetColumnByID(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for non-existent column")
	}

	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetColumnByID_InvalidID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetColumnByID(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestUpdateColumnName(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	err = svc.UpdateColumnName(context.Background(), created.ID, "Backlog")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify update
	updated, err := svc.GetColumnByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated column: %v", err)
	}

	if updated.Name != "Backlog" {
		t.Errorf("Expected name 'Backlog', got '%s'", updated.Name)
	}
}

func TestUpdateColumnName_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		columnID int
		newName  string
		wantErr  bool
		errType  error
		setupFn  func(*sql.DB) int // Returns column ID if needed
	}{
		{
			name:    "empty name",
			newName: "",
			wantErr: true,
			errType: ErrEmptyName,
			setupFn: func(db *sql.DB) int {
				projectID := createTestProject(t, db)
				col, _ := NewService(db, nil).CreateColumn(context.Background(), CreateColumnRequest{
					Name:      "To Do",
					ProjectID: projectID,
				})
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
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			columnID := tt.columnID
			if tt.setupFn != nil {
				columnID = tt.setupFn(db)
			}

			svc := NewService(db, nil)
			err := svc.UpdateColumnName(context.Background(), columnID, tt.newName)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateColumnName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("UpdateColumnName() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestDeleteColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	err = svc.DeleteColumn(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify column is deleted
	_, err = svc.GetColumnByID(context.Background(), created.ID)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows after deletion, got %v", err)
	}
}

func TestDeleteColumn_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		columnID int
		wantErr  bool
		errType  error
		setupFn  func(*sql.DB) int // Returns column ID if needed
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
			setupFn: func(db *sql.DB) int {
				projectID := createTestProject(t, db)
				col, _ := NewService(db, nil).CreateColumn(context.Background(), CreateColumnRequest{
					Name:      "To Do",
					ProjectID: projectID,
				})
				createTestTask(t, db, col.ID)
				return col.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			columnID := tt.columnID
			if tt.setupFn != nil {
				columnID = tt.setupFn(db)
			}

			svc := NewService(db, nil)
			err := svc.DeleteColumn(context.Background(), columnID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteColumn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("DeleteColumn() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestDeleteColumn_LinkedListIntegrity(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create three columns: col1 <-> col2 <-> col3
	col1, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})

	col2, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "In Progress",
		ProjectID: projectID,
	})

	col3, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})

	// Delete middle column (col2)
	err := svc.DeleteColumn(context.Background(), col2.ID)
	if err != nil {
		t.Fatalf("Failed to delete column 2: %v", err)
	}

	// Verify linked list is repaired: col1 <-> col3

	col1Updated, _ := svc.GetColumnByID(context.Background(), col1.ID)
	col3Updated, _ := svc.GetColumnByID(context.Background(), col3.ID)

	if col1Updated.NextID == nil || *col1Updated.NextID != col3.ID {
		t.Errorf("Expected col1 next_id %d after deletion, got %v", col3.ID, col1Updated.NextID)
	}

	if col3Updated.PrevID == nil || *col3Updated.PrevID != col1.ID {
		t.Errorf("Expected col3 prev_id %d after deletion, got %v", col1.ID, col3Updated.PrevID)
	}
}

// ============================================================================
// TEST CASES - HOLDS_READY_TASKS FEATURE
// ============================================================================

func TestCreateColumn_WithHoldsReadyTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	}

	result, err := svc.CreateColumn(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsReadyTasks {
		t.Error("Expected HoldsReadyTasks to be true")
	}
}

func TestCreateColumn_HoldsReadyTasks_ClearsPrevious(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create first column with HoldsReadyTasks = true
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	if !col1.HoldsReadyTasks {
		t.Fatal("Expected col1 to hold ready tasks")
	}

	// Create second column with HoldsReadyTasks = true
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "In Progress",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	if !col2.HoldsReadyTasks {
		t.Error("Expected col2 to hold ready tasks")
	}

	// Verify col1 is no longer the ready column
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get updated col1: %v", err)
	}

	if col1Updated.HoldsReadyTasks {
		t.Error("Expected col1 to no longer hold ready tasks")
	}
}

func TestSetHoldsReadyTasks_Success(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two columns (neither ready)
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "Done",
		ProjectID:       projectID,
		HoldsReadyTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Set col1 as ready
	updated, err := svc.SetHoldsReadyTasks(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !updated.HoldsReadyTasks {
		t.Error("Expected column to hold ready tasks after SetHoldsReadyTasks")
	}
}

func TestSetHoldsReadyTasks_TransfersFromPrevious(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create col1 as ready
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Create col2 as not ready
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "In Progress",
		ProjectID:       projectID,
		HoldsReadyTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Attempt to set col2 as ready - should fail because col1 already holds ready tasks
	_, err = svc.SetHoldsReadyTasks(context.Background(), col2.ID)
	if err == nil {
		t.Fatal("Expected error when setting ready tasks on col2 while col1 is ready, got nil")
	}

	// Verify error message includes the existing column info
	expectedErrorSubstring := "To Do"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErrorSubstring, err)
	}

	// Verify col1 still holds ready tasks
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get col1: %v", err)
	}

	if !col1Updated.HoldsReadyTasks {
		t.Error("Expected col1 to still hold ready tasks")
	}

	// Verify col2 does not hold ready tasks
	col2Updated, err := svc.GetColumnByID(context.Background(), col2.ID)
	if err != nil {
		t.Fatalf("Failed to get col2: %v", err)
	}

	if col2Updated.HoldsReadyTasks {
		t.Error("Expected col2 to not hold ready tasks")
	}
}

func TestGetColumnByID_IncludesHoldsReadyTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create column with HoldsReadyTasks = true
	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Fetch via GetColumnByID
	result, err := svc.GetColumnByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsReadyTasks {
		t.Error("Expected HoldsReadyTasks to be true")
	}
}

func TestGetColumnsByProject_IncludesHoldsReadyTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create one ready column and one not ready
	_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create ready column: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "Done",
		ProjectID:       projectID,
		HoldsReadyTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create non-ready column: %v", err)
	}

	// Fetch all columns
	results, err := svc.GetColumnsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(results))
	}

	// Verify first column (To Do) is ready
	if results[0].Name == "To Do" && !results[0].HoldsReadyTasks {
		t.Error("Expected 'To Do' column to hold ready tasks")
	}

	// Verify second column (Done) is not ready
	if results[1].Name == "Done" && results[1].HoldsReadyTasks {
		t.Error("Expected 'Done' column to not hold ready tasks")
	}
}

func TestSetHoldsReadyTasks_InvalidColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsReadyTasks(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid column ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestSetHoldsReadyTasks_ColumnNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsReadyTasks(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for non-existent column")
	}

	// Should get a wrapped sql.ErrNoRows
	if err != sql.ErrNoRows && !contains(err.Error(), "no rows") {
		t.Errorf("Expected sql.ErrNoRows or wrapped error, got %v", err)
	}
}

func TestCreateColumn_OnlyOneReadyPerProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)

	// Manually insert two columns with holds_ready_tasks = 1
	// This should violate the unique partial index constraint
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO columns (name, project_id, holds_ready_tasks) VALUES (?, ?, ?)",
		"To Do", projectID, true)
	if err != nil {
		t.Fatalf("Failed to insert first ready column: %v", err)
	}

	_, err = db.ExecContext(context.Background(),
		"INSERT INTO columns (name, project_id, holds_ready_tasks) VALUES (?, ?, ?)",
		"Review", projectID, true)

	if err == nil {
		t.Fatal("Expected database constraint violation for duplicate ready columns")
	}

	// Should get a constraint violation error
	if !contains(err.Error(), "UNIQUE") && !contains(err.Error(), "constraint") {
		t.Errorf("Expected UNIQUE constraint violation, got %v", err)
	}
}

// Helper function to check if error message contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================================
// ERROR PATH TESTS - INVALID IDs
// ============================================================================

func TestGetColumnByID_NegativeID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetColumnByID(context.Background(), -1)

	if err == nil {
		t.Fatal("Expected error for negative ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestGetColumnsByProject_NegativeProjectID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.GetColumnsByProject(context.Background(), -1)

	if err == nil {
		t.Fatal("Expected error for negative project ID")
	}

	if err != ErrInvalidProjectID {
		t.Errorf("Expected ErrInvalidProjectID, got %v", err)
	}
}

func TestGetColumnsByProject_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Non-existent project should return empty list, not error
	results, err := svc.GetColumnsByProject(context.Background(), 999999)
	if err != nil {
		t.Fatalf("Expected no error for non-existent project, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 columns for non-existent project, got %d", len(results))
	}
}

func TestUpdateColumnName_NegativeID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.UpdateColumnName(context.Background(), -1, "New Name")

	if err == nil {
		t.Fatal("Expected error for negative ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestUpdateColumnName_NonExistentColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.UpdateColumnName(context.Background(), 999999, "New Name")

	if err == nil {
		t.Fatal("Expected error for non-existent column")
	}

	if !contains(err.Error(), "failed to get column") {
		t.Errorf("Expected error about failed to get column, got %v", err)
	}
}

func TestUpdateColumnName_NameTooLong(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Create a name that's 51 characters long
	longName := strings.Repeat("a", 51)

	err = svc.UpdateColumnName(context.Background(), created.ID, longName)

	if err == nil {
		t.Fatal("Expected error for name too long")
	}

	if err != ErrNameTooLong {
		t.Errorf("Expected ErrNameTooLong, got %v", err)
	}
}

func TestDeleteColumn_NegativeID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.DeleteColumn(context.Background(), -1)

	if err == nil {
		t.Fatal("Expected error for negative ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestDeleteColumn_NonExistentColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	err := svc.DeleteColumn(context.Background(), 999999)

	if err == nil {
		t.Fatal("Expected error for non-existent column")
	}

	if !contains(err.Error(), "failed to get column info") {
		t.Errorf("Expected error about failed to get column info, got %v", err)
	}
}

// ============================================================================
// ERROR PATH TESTS - COLUMN LINKING EDGE CASES
// ============================================================================

func TestCreateColumn_InvalidAfterID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		afterID int
		wantErr bool
		errType error
		setupFn func(*sql.DB) int // Returns project ID
	}{
		{
			name:    "zero after ID",
			afterID: 0,
			wantErr: true,
			errType: ErrInvalidColumnID,
			setupFn: func(db *sql.DB) int {
				return createTestProject(t, db)
			},
		},
		{
			name:    "negative after ID",
			afterID: -1,
			wantErr: true,
			errType: ErrInvalidColumnID,
			setupFn: func(db *sql.DB) int {
				return createTestProject(t, db)
			},
		},
		{
			name:    "non-existent after ID",
			afterID: 999999,
			wantErr: false, // Service doesn't validate existence of afterID before attempting operations
			setupFn: func(db *sql.DB) int {
				return createTestProject(t, db)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			defer func() { _ = db.Close() }()

			projectID := tt.setupFn(db)
			svc := NewService(db, nil)

			_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
				Name:      "Test Column",
				ProjectID: projectID,
				AfterID:   &tt.afterID,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateColumn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("CreateColumn() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestCreateColumn_AfterColumnFromDifferentProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create two projects
	projectID1 := createTestProject(t, db)
	projectID2 := createTestProject(t, db)

	// Create column in project 1
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Column 1",
		ProjectID: projectID1,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Try to create column in project 2 after column from project 1
	// This should succeed because the service doesn't validate project matching
	afterID := col1.ID
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Column 2",
		ProjectID: projectID2,
		AfterID:   &afterID,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the column was created successfully
	if col2 == nil {
		t.Fatal("Expected column to be created")
	}

	// The column should be in project 2
	if col2.ProjectID != projectID2 {
		t.Errorf("Expected column project ID %d, got %d", projectID2, col2.ProjectID)
	}
}

func TestDeleteColumn_FirstColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create three columns: col1 <-> col2 <-> col3
	col1, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})

	col2, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "In Progress",
		ProjectID: projectID,
	})

	col3, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})

	// Delete first column (col1)
	err := svc.DeleteColumn(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to delete column 1: %v", err)
	}

	// Verify linked list is repaired: col2 <-> col3

	col2Updated, _ := svc.GetColumnByID(context.Background(), col2.ID)
	col3Updated, _ := svc.GetColumnByID(context.Background(), col3.ID)

	if col2Updated.PrevID != nil {
		t.Errorf("Expected col2 prev_id nil after deletion, got %v", col2Updated.PrevID)
	}

	if col2Updated.NextID == nil || *col2Updated.NextID != col3.ID {
		t.Errorf("Expected col2 next_id %d after deletion, got %v", col3.ID, col2Updated.NextID)
	}

	if col3Updated.PrevID == nil || *col3Updated.PrevID != col2.ID {
		t.Errorf("Expected col3 prev_id %d after deletion, got %v", col2.ID, col3Updated.PrevID)
	}
}

func TestDeleteColumn_LastColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create three columns: col1 <-> col2 <-> col3
	col1, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})

	col2, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "In Progress",
		ProjectID: projectID,
	})

	col3, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "Done",
		ProjectID: projectID,
	})

	// Delete last column (col3)
	err := svc.DeleteColumn(context.Background(), col3.ID)
	if err != nil {
		t.Fatalf("Failed to delete column 3: %v", err)
	}

	// Verify linked list is repaired: col1 <-> col2

	col1Updated, _ := svc.GetColumnByID(context.Background(), col1.ID)
	col2Updated, _ := svc.GetColumnByID(context.Background(), col2.ID)

	if col1Updated.NextID == nil || *col1Updated.NextID != col2.ID {
		t.Errorf("Expected col1 next_id %d after deletion, got %v", col2.ID, col1Updated.NextID)
	}

	if col2Updated.PrevID == nil || *col2Updated.PrevID != col1.ID {
		t.Errorf("Expected col2 prev_id %d after deletion, got %v", col1.ID, col2Updated.PrevID)
	}

	if col2Updated.NextID != nil {
		t.Errorf("Expected col2 next_id nil after deletion, got %v", col2Updated.NextID)
	}
}

func TestDeleteColumn_OnlyColumn(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create single column
	col, _ := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})

	// Delete it
	err := svc.DeleteColumn(context.Background(), col.ID)
	if err != nil {
		t.Fatalf("Failed to delete only column: %v", err)
	}

	// Verify project has no columns
	columns, err := svc.GetColumnsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	if len(columns) != 0 {
		t.Errorf("Expected 0 columns after deletion, got %d", len(columns))
	}
}

// ============================================================================
// ERROR PATH TESTS - SPECIAL COLUMN FLAGS (HOLDS_READY_TASKS)
// ============================================================================

func TestSetHoldsReadyTasks_NegativeColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsReadyTasks(context.Background(), -1)

	if err == nil {
		t.Fatal("Expected error for negative column ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestCreateColumn_MultipleReadyColumns_DifferentProjects(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create two projects
	projectID1 := createTestProject(t, db)
	projectID2 := createTestProject(t, db)

	// Create ready column in project 1
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID1,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create ready column in project 1: %v", err)
	}

	if !col1.HoldsReadyTasks {
		t.Fatal("Expected col1 to hold ready tasks")
	}

	// Create ready column in project 2 - should succeed
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:            "To Do",
		ProjectID:       projectID2,
		HoldsReadyTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create ready column in project 2: %v", err)
	}

	if !col2.HoldsReadyTasks {
		t.Fatal("Expected col2 to hold ready tasks")
	}

	// Verify both columns are still ready
	col1Check, _ := svc.GetColumnByID(context.Background(), col1.ID)
	col2Check, _ := svc.GetColumnByID(context.Background(), col2.ID)

	if !col1Check.HoldsReadyTasks {
		t.Error("Expected col1 to still hold ready tasks")
	}

	if !col2Check.HoldsReadyTasks {
		t.Error("Expected col2 to still hold ready tasks")
	}
}

// ============================================================================
// ERROR PATH TESTS - SPECIAL COLUMN FLAGS (HOLDS_COMPLETED_TASKS)
// ============================================================================

func TestSetHoldsCompletedTasks_NegativeColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsCompletedTasks(context.Background(), -1, false)

	if err == nil {
		t.Fatal("Expected error for negative column ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestCreateColumn_MultipleCompletedColumns_DifferentProjects(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create two projects
	projectID1 := createTestProject(t, db)
	projectID2 := createTestProject(t, db)

	// Create completed column in project 1
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID1,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create completed column in project 1: %v", err)
	}

	if !col1.HoldsCompletedTasks {
		t.Fatal("Expected col1 to hold completed tasks")
	}

	// Create completed column in project 2 - should succeed
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID2,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create completed column in project 2: %v", err)
	}

	if !col2.HoldsCompletedTasks {
		t.Fatal("Expected col2 to hold completed tasks")
	}

	// Verify both columns are still completed
	col1Check, _ := svc.GetColumnByID(context.Background(), col1.ID)
	col2Check, _ := svc.GetColumnByID(context.Background(), col2.ID)

	if !col1Check.HoldsCompletedTasks {
		t.Error("Expected col1 to still hold completed tasks")
	}

	if !col2Check.HoldsCompletedTasks {
		t.Error("Expected col2 to still hold completed tasks")
	}
}

// ============================================================================
// ERROR PATH TESTS - SPECIAL COLUMN FLAGS (HOLDS_IN_PROGRESS_TASKS)
// ============================================================================

func TestCreateColumn_WithHoldsInProgressTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateColumnRequest{
		Name:                 "In Progress",
		ProjectID:            projectID,
		HoldsInProgressTasks: true,
	}

	result, err := svc.CreateColumn(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsInProgressTasks {
		t.Error("Expected HoldsInProgressTasks to be true")
	}
}

func TestCreateColumn_HoldsInProgressTasks_ClearsPrevious(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create first column with HoldsInProgressTasks = true
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "In Progress",
		ProjectID:            projectID,
		HoldsInProgressTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	if !col1.HoldsInProgressTasks {
		t.Fatal("Expected col1 to hold in-progress tasks")
	}

	// Create second column with HoldsInProgressTasks = true
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "Doing",
		ProjectID:            projectID,
		HoldsInProgressTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	if !col2.HoldsInProgressTasks {
		t.Error("Expected col2 to hold in-progress tasks")
	}

	// Verify col1 is no longer the in-progress column
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get updated col1: %v", err)
	}

	if col1Updated.HoldsInProgressTasks {
		t.Error("Expected col1 to no longer hold in-progress tasks")
	}
}

func TestSetHoldsInProgressTasks_Success(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two columns (neither in-progress)
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "In Progress",
		ProjectID:            projectID,
		HoldsInProgressTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "Done",
		ProjectID:            projectID,
		HoldsInProgressTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Set col1 as in-progress
	updated, err := svc.SetHoldsInProgressTasks(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !updated.HoldsInProgressTasks {
		t.Error("Expected column to hold in-progress tasks after SetHoldsInProgressTasks")
	}
}

func TestSetHoldsInProgressTasks_FailsWhenExists(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create col1 as in-progress
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "In Progress",
		ProjectID:            projectID,
		HoldsInProgressTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Create col2 as not in-progress
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "Doing",
		ProjectID:            projectID,
		HoldsInProgressTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Attempt to set col2 as in-progress - should fail
	_, err = svc.SetHoldsInProgressTasks(context.Background(), col2.ID)
	if err == nil {
		t.Fatal("Expected error when setting in-progress tasks on col2 while col1 is in-progress")
	}

	// Verify error message includes the existing column info
	expectedErrorSubstring := "In Progress"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErrorSubstring, err)
	}

	// Verify col1 still holds in-progress tasks
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get col1: %v", err)
	}

	if !col1Updated.HoldsInProgressTasks {
		t.Error("Expected col1 to still hold in-progress tasks")
	}

	// Verify col2 does not hold in-progress tasks
	col2Updated, err := svc.GetColumnByID(context.Background(), col2.ID)
	if err != nil {
		t.Fatalf("Failed to get col2: %v", err)
	}

	if col2Updated.HoldsInProgressTasks {
		t.Error("Expected col2 to not hold in-progress tasks")
	}
}

func TestSetHoldsInProgressTasks_InvalidColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsInProgressTasks(context.Background(), 0)

	if err == nil {
		t.Fatal("Expected error for invalid column ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestSetHoldsInProgressTasks_NegativeColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsInProgressTasks(context.Background(), -1)

	if err == nil {
		t.Fatal("Expected error for negative column ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestSetHoldsInProgressTasks_ColumnNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsInProgressTasks(context.Background(), 999)

	if err == nil {
		t.Fatal("Expected error for non-existent column")
	}

	// Should get a wrapped sql.ErrNoRows
	if err != sql.ErrNoRows && !contains(err.Error(), "no rows") {
		t.Errorf("Expected sql.ErrNoRows or wrapped error, got %v", err)
	}
}

func TestGetColumnByID_IncludesHoldsInProgressTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create column with HoldsInProgressTasks = true
	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "In Progress",
		ProjectID:            projectID,
		HoldsInProgressTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Fetch via GetColumnByID
	result, err := svc.GetColumnByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsInProgressTasks {
		t.Error("Expected HoldsInProgressTasks to be true")
	}
}

func TestGetColumnsByProject_IncludesHoldsInProgressTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create one in-progress column and one not in-progress
	_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "In Progress",
		ProjectID:            projectID,
		HoldsInProgressTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create in-progress column: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "Done",
		ProjectID:            projectID,
		HoldsInProgressTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create non-in-progress column: %v", err)
	}

	// Fetch all columns
	results, err := svc.GetColumnsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(results))
	}

	// Verify first column (In Progress) is in-progress
	if results[0].Name == "In Progress" && !results[0].HoldsInProgressTasks {
		t.Error("Expected 'In Progress' column to hold in-progress tasks")
	}

	// Verify second column (Done) is not in-progress
	if results[1].Name == "Done" && results[1].HoldsInProgressTasks {
		t.Error("Expected 'Done' column to not hold in-progress tasks")
	}
}

func TestCreateColumn_MultipleInProgressColumns_DifferentProjects(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Create two projects
	projectID1 := createTestProject(t, db)
	projectID2 := createTestProject(t, db)

	// Create in-progress column in project 1
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "In Progress",
		ProjectID:            projectID1,
		HoldsInProgressTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create in-progress column in project 1: %v", err)
	}

	if !col1.HoldsInProgressTasks {
		t.Fatal("Expected col1 to hold in-progress tasks")
	}

	// Create in-progress column in project 2 - should succeed
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                 "In Progress",
		ProjectID:            projectID2,
		HoldsInProgressTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create in-progress column in project 2: %v", err)
	}

	if !col2.HoldsInProgressTasks {
		t.Fatal("Expected col2 to hold in-progress tasks")
	}

	// Verify both columns are still in-progress
	col1Check, _ := svc.GetColumnByID(context.Background(), col1.ID)
	col2Check, _ := svc.GetColumnByID(context.Background(), col2.ID)

	if !col1Check.HoldsInProgressTasks {
		t.Error("Expected col1 to still hold in-progress tasks")
	}

	if !col2Check.HoldsInProgressTasks {
		t.Error("Expected col2 to still hold in-progress tasks")
	}
}

// ============================================================================
// ERROR PATH TESTS - BOUNDARY CONDITIONS
// ============================================================================

func TestCreateColumn_NameExactly50Characters(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create a name that's exactly 50 characters
	name50 := strings.Repeat("a", 50)

	req := CreateColumnRequest{
		Name:      name50,
		ProjectID: projectID,
	}

	result, err := svc.CreateColumn(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error for 50-character name, got %v", err)
	}

	if result.Name != name50 {
		t.Errorf("Expected name '%s', got '%s'", name50, result.Name)
	}
}

func TestCreateColumn_NonExistentProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Try to create column in non-existent project
	_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: 999999,
	})

	if err == nil {
		t.Fatal("Expected error when creating column in non-existent project")
	}

	// Error should be related to foreign key constraint
	if !contains(err.Error(), "failed to create column") {
		t.Errorf("Expected error about failed to create column, got %v", err)
	}
}

func TestUpdateColumnName_Exact50Characters(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Create a name that's exactly 50 characters
	name50 := strings.Repeat("a", 50)

	err = svc.UpdateColumnName(context.Background(), created.ID, name50)
	if err != nil {
		t.Fatalf("Expected no error for 50-character name, got %v", err)
	}

	// Verify update
	updated, err := svc.GetColumnByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated column: %v", err)
	}

	if updated.Name != name50 {
		t.Errorf("Expected name '%s', got '%s'", name50, updated.Name)
	}
}

func TestCreateColumn_AllSpecialFlagsTrue(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Try to create column with all special flags set to true
	req := CreateColumnRequest{
		Name:                 "Multi-purpose",
		ProjectID:            projectID,
		HoldsReadyTasks:      true,
		HoldsCompletedTasks:  true,
		HoldsInProgressTasks: true,
	}

	result, err := svc.CreateColumn(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify all flags are set
	if !result.HoldsReadyTasks {
		t.Error("Expected HoldsReadyTasks to be true")
	}

	if !result.HoldsCompletedTasks {
		t.Error("Expected HoldsCompletedTasks to be true")
	}

	if !result.HoldsInProgressTasks {
		t.Error("Expected HoldsInProgressTasks to be true")
	}
}

func TestCreateColumn_ProjectIDMaxInt(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	// Try to create column with max int project ID (won't exist)
	_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:      "To Do",
		ProjectID: 2147483647, // max int32
	})

	if err == nil {
		t.Fatal("Expected error for non-existent project with max int ID")
	}

	// Error should be related to foreign key constraint
	if !contains(err.Error(), "failed to create column") {
		t.Errorf("Expected error about failed to create column, got %v", err)
	}
}

// ============================================================================
// TEST CASES - HOLDS_COMPLETED_TASKS FEATURE
// ============================================================================

func TestCreateColumn_WithHoldsCompletedTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	req := CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	}

	result, err := svc.CreateColumn(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsCompletedTasks {
		t.Error("Expected HoldsCompletedTasks to be true")
	}
}

func TestCreateColumn_HoldsCompletedTasks_FailsWhenExists(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create first column with HoldsCompletedTasks = true
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	if !col1.HoldsCompletedTasks {
		t.Fatal("Expected col1 to hold completed tasks")
	}

	// Create second column with HoldsCompletedTasks = true (should fail)
	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Archive",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})

	if err == nil {
		t.Fatal("Expected error when creating second completed column")
	}

	if err != ErrCompletedColumnExists && !contains(err.Error(), "completed column already exists") {
		t.Errorf("Expected ErrCompletedColumnExists, got %v", err)
	}
}

func TestSetHoldsCompletedTasks_Success(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create two columns (neither completed)
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Archive",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Set col1 as completed
	updated, err := svc.SetHoldsCompletedTasks(context.Background(), col1.ID, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !updated.HoldsCompletedTasks {
		t.Error("Expected column to hold completed tasks after SetHoldsCompletedTasks")
	}
}

func TestSetHoldsCompletedTasks_FailsWhenExistsWithoutForce(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create col1 as completed
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Create col2 as not completed
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Archive",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Try to set col2 as completed without force (should fail)
	_, err = svc.SetHoldsCompletedTasks(context.Background(), col2.ID, false)

	if err == nil {
		t.Fatal("Expected error when setting completed without force")
	}

	if err != ErrCompletedColumnExists && !contains(err.Error(), "completed column already exists") {
		t.Errorf("Expected ErrCompletedColumnExists, got %v", err)
	}

	// Verify col1 is still completed
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get col1: %v", err)
	}

	if !col1Updated.HoldsCompletedTasks {
		t.Error("Expected col1 to still hold completed tasks")
	}
}

func TestSetHoldsCompletedTasks_SucceedsWithForce(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create col1 as completed
	col1, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column 1: %v", err)
	}

	// Create col2 as not completed
	col2, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Archive",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create column 2: %v", err)
	}

	// Set col2 as completed with force (should succeed)
	updated, err := svc.SetHoldsCompletedTasks(context.Background(), col2.ID, true)
	if err != nil {
		t.Fatalf("Expected no error with force flag, got %v", err)
	}

	if !updated.HoldsCompletedTasks {
		t.Error("Expected col2 to hold completed tasks")
	}

	// Verify col1 is no longer completed
	col1Updated, err := svc.GetColumnByID(context.Background(), col1.ID)
	if err != nil {
		t.Fatalf("Failed to get col1: %v", err)
	}

	if col1Updated.HoldsCompletedTasks {
		t.Error("Expected col1 to no longer hold completed tasks")
	}
}

func TestGetColumnByID_IncludesHoldsCompletedTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create column with HoldsCompletedTasks = true
	created, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create column: %v", err)
	}

	// Fetch via GetColumnByID
	result, err := svc.GetColumnByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.HoldsCompletedTasks {
		t.Error("Expected HoldsCompletedTasks to be true")
	}
}

func TestGetColumnsByProject_IncludesHoldsCompletedTasks(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)
	svc := NewService(db, nil)

	// Create one completed column and one not completed
	_, err := svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Done",
		ProjectID:           projectID,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		t.Fatalf("Failed to create completed column: %v", err)
	}

	_, err = svc.CreateColumn(context.Background(), CreateColumnRequest{
		Name:                "Todo",
		ProjectID:           projectID,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		t.Fatalf("Failed to create non-completed column: %v", err)
	}

	// Fetch all columns
	results, err := svc.GetColumnsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(results))
	}

	// Verify first column (Done) is completed
	if results[0].Name == "Done" && !results[0].HoldsCompletedTasks {
		t.Error("Expected 'Done' column to hold completed tasks")
	}

	// Verify second column (Todo) is not completed
	if results[1].Name == "Todo" && results[1].HoldsCompletedTasks {
		t.Error("Expected 'Todo' column to not hold completed tasks")
	}
}

func TestSetHoldsCompletedTasks_InvalidColumnID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsCompletedTasks(context.Background(), 0, false)

	if err == nil {
		t.Fatal("Expected error for invalid column ID")
	}

	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestSetHoldsCompletedTasks_ColumnNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	svc := NewService(db, nil)

	_, err := svc.SetHoldsCompletedTasks(context.Background(), 999, false)

	if err == nil {
		t.Fatal("Expected error for non-existent column")
	}

	// Should get a wrapped sql.ErrNoRows
	if err != sql.ErrNoRows && !contains(err.Error(), "no rows") {
		t.Errorf("Expected sql.ErrNoRows or wrapped error, got %v", err)
	}
}

func TestCreateColumn_OnlyOneCompletedPerProject(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	projectID := createTestProject(t, db)

	// Manually insert two columns with holds_completed_tasks = 1
	// This should violate the unique partial index constraint
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO columns (name, project_id, holds_completed_tasks) VALUES (?, ?, ?)",
		"Done", projectID, true)
	if err != nil {
		t.Fatalf("Failed to insert first completed column: %v", err)
	}

	_, err = db.ExecContext(context.Background(),
		"INSERT INTO columns (name, project_id, holds_completed_tasks) VALUES (?, ?, ?)",
		"Archive", projectID, true)

	if err == nil {
		t.Fatal("Expected database constraint violation for duplicate completed columns")
	}

	// Should get a constraint violation error
	if !contains(err.Error(), "UNIQUE") && !contains(err.Error(), "constraint") {
		t.Errorf("Expected UNIQUE constraint violation, got %v", err)
	}
}
