package state

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
)

// TestGetCurrentProject_EmptyProjects ensures project access with no projects returns nil.
// Edge case: Application startup with no projects in database.
// Security value: Prevents nil pointer dereference.
func TestGetCurrentProject_EmptyProjects(t *testing.T) {
	state := NewAppState(
		[]*models.Project{}, // Empty projects
		0,
		nil,
		nil,
		nil,
	)

	got := state.GetCurrentProject()
	if got != nil {
		t.Errorf("GetCurrentProject() with empty projects = %v, want nil", got)
	}
}

// TestGetCurrentProject_InvalidIndex ensures project access with out-of-bounds index returns nil.
// Edge case: Corrupted state with invalid selectedProject index.
// Security value: Prevents index out of bounds panic.
func TestGetCurrentProject_InvalidIndex(t *testing.T) {
	projects := []*models.Project{
		{ID: 1, Name: "Project 1"},
		{ID: 2, Name: "Project 2"},
	}

	// Test with index beyond slice length
	state := NewAppState(projects, 5, nil, nil, nil)
	got := state.GetCurrentProject()
	if got != nil {
		t.Errorf("GetCurrentProject() with index=5 (out of bounds) = %v, want nil", got)
	}

	// Test with negative index
	state.SetSelectedProject(-1)
	got = state.GetCurrentProject()
	if got != nil {
		t.Errorf("GetCurrentProject() with index=-1 = %v, want nil", got)
	}

	// Test with valid index for comparison
	state.SetSelectedProject(1)
	got = state.GetCurrentProject()
	if got == nil {
		t.Error("GetCurrentProject() with valid index=1 = nil, want non-nil")
	} else if got.ID != 2 {
		t.Errorf("GetCurrentProject() with index=1, got project ID %d, want 2", got.ID)
	}
}

// TestGetCurrentProjectID_NilProject ensures project ID returns 0 when no project selected.
// Edge case: GetCurrentProject() returns nil.
// Security value: Returns safe default (0) instead of panicking.
func TestGetCurrentProjectID_NilProject(t *testing.T) {
	// Empty projects -> GetCurrentProject returns nil
	state := NewAppState([]*models.Project{}, 0, nil, nil, nil)

	got := state.GetCurrentProjectID()
	if got != 0 {
		t.Errorf("GetCurrentProjectID() with nil project = %d, want 0", got)
	}

	// Valid project for comparison
	projects := []*models.Project{{ID: 42, Name: "Test Project"}}
	state = NewAppState(projects, 0, nil, nil, nil)
	got = state.GetCurrentProjectID()
	if got != 42 {
		t.Errorf("GetCurrentProjectID() with valid project = %d, want 42", got)
	}
}

// TestNewAppState_NilTasks ensures constructor initializes nil tasks map to empty map.
// Edge case: Constructor called with nil tasks map.
// Security value: Prevents nil map write panic (assigning to nil map causes panic).
func TestNewAppState_NilTasks(t *testing.T) {
	// Pass nil for tasks map
	state := NewAppState(nil, 0, nil, nil, nil)

	// Accessing the tasks map should not panic
	tasks := state.Tasks()
	if tasks == nil {
		t.Error("NewAppState() with nil tasks map resulted in nil map, want initialized empty map")
	}

	// Should be able to write to the map without panic
	tasks[1] = []*models.TaskSummary{{ID: 1, Title: "Test Task"}}
	if len(tasks) != 1 {
		t.Errorf("Writing to tasks map failed, got length %d, want 1", len(tasks))
	}

	// SetTasks should also handle nil
	state.SetTasks(nil)
	tasks = state.Tasks()
	if tasks == nil {
		t.Error("SetTasks(nil) resulted in nil map, want initialized empty map")
	}
}

// TestGetColumnByID_ValidColumn ensures O(1) column lookup works correctly.
// Performance value: Verifies the columnByID map provides correct lookups.
func TestGetColumnByID_ValidColumn(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "In Progress"},
		{ID: 3, Name: "Done"},
	}

	state := NewAppState(nil, 0, columns, nil, nil)

	// Test valid lookups
	col := state.GetColumnByID(2)
	if col == nil {
		t.Fatal("GetColumnByID(2) = nil, want non-nil")
	}
	if col.ID != 2 {
		t.Errorf("GetColumnByID(2) returned column with ID=%d, want 2", col.ID)
	}
	if col.Name != "In Progress" {
		t.Errorf("GetColumnByID(2) returned column with Name=%s, want 'In Progress'", col.Name)
	}
}

// TestGetColumnByID_InvalidColumn ensures lookup for non-existent column returns nil.
// Edge case: Querying for a column ID that doesn't exist.
// Security value: Prevents nil pointer dereference.
func TestGetColumnByID_InvalidColumn(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "Done"},
	}

	state := NewAppState(nil, 0, columns, nil, nil)

	// Test invalid lookup
	col := state.GetColumnByID(999)
	if col != nil {
		t.Errorf("GetColumnByID(999) = %v, want nil", col)
	}
}

// TestGetColumnByID_EmptyColumns ensures lookup with no columns returns nil.
// Edge case: Application state with no columns loaded.
// Security value: Prevents nil pointer dereference.
func TestGetColumnByID_EmptyColumns(t *testing.T) {
	state := NewAppState(nil, 0, []*models.Column{}, nil, nil)

	col := state.GetColumnByID(1)
	if col != nil {
		t.Errorf("GetColumnByID(1) with empty columns = %v, want nil", col)
	}
}

// TestSetColumns_UpdatesMap ensures SetColumns rebuilds the columnByID map.
// Performance value: Verifies map stays in sync when columns change.
func TestSetColumns_UpdatesMap(t *testing.T) {
	initialColumns := []*models.Column{
		{ID: 1, Name: "Original"},
	}
	state := NewAppState(nil, 0, initialColumns, nil, nil)

	// Verify initial state
	col := state.GetColumnByID(1)
	if col == nil || col.Name != "Original" {
		t.Fatal("Initial column lookup failed")
	}

	// Update columns
	newColumns := []*models.Column{
		{ID: 2, Name: "Updated"},
		{ID: 3, Name: "New"},
	}
	state.SetColumns(newColumns)

	// Old column should no longer be found
	col = state.GetColumnByID(1)
	if col != nil {
		t.Error("GetColumnByID(1) after SetColumns should return nil, map not updated")
	}

	// New columns should be found
	col = state.GetColumnByID(2)
	if col == nil || col.Name != "Updated" {
		t.Error("GetColumnByID(2) after SetColumns failed, map not rebuilt")
	}

	col = state.GetColumnByID(3)
	if col == nil || col.Name != "New" {
		t.Error("GetColumnByID(3) after SetColumns failed, map not rebuilt")
	}
}
