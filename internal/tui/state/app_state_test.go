package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
)

// TestGetCurrentProject_EmptyProjects ensures project access with no projects returns nil.
// Edge case: Application startup with no projects in database.
func TestGetCurrentProject_EmptyProjects(t *testing.T) {
	t.Parallel()
	state := NewAppState(
		[]*models.Project{}, // Empty projects
		0,
		nil,
		nil,
		nil,
	)

	got := state.GetCurrentProject()
	assert.Nil(t, got)
}

// TestGetCurrentProject_InvalidIndex ensures project access with out-of-bounds index returns nil.
// Edge case: Corrupted state with invalid selectedProject index.
func TestGetCurrentProject_InvalidIndex(t *testing.T) {
	t.Parallel()
	projects := []*models.Project{
		{ID: 1, Name: "Project 1"},
		{ID: 2, Name: "Project 2"},
	}

	// Test with index beyond slice length
	state := NewAppState(projects, 5, nil, nil, nil)
	got := state.GetCurrentProject()
	assert.Nil(t, got, "GetCurrentProject() with index=5 (out of bounds)")

	// Test with negative index
	state.SetSelectedProject(-1)
	got = state.GetCurrentProject()
	assert.Nil(t, got, "GetCurrentProject() with index=-1")

	// Test with valid index for comparison
	state.SetSelectedProject(1)
	got = state.GetCurrentProject()
	require.NotNil(t, got, "GetCurrentProject() with valid index=1 = nil, want non-nil")
	assert.Equal(t, 2, got.ID)
}

// TestGetCurrentProjectID_NilProject ensures project ID returns 0 when no project selected.
// Edge case: GetCurrentProject() returns nil.
func TestGetCurrentProjectID_NilProject(t *testing.T) {
	t.Parallel()
	// Empty projects -> GetCurrentProject returns nil
	state := NewAppState([]*models.Project{}, 0, nil, nil, nil)

	got := state.GetCurrentProjectID()
	assert.Equal(t, 0, got)

	// Valid project for comparison
	projects := []*models.Project{{ID: 42, Name: "Test Project"}}
	state = NewAppState(projects, 0, nil, nil, nil)
	got = state.GetCurrentProjectID()
	assert.Equal(t, 42, got)
}

// TestNewAppState_NilTasks ensures constructor initializes nil tasks map to empty map.
// Edge case: Constructor called with nil tasks map.
func TestNewAppState_NilTasks(t *testing.T) {
	t.Parallel()
	// Pass nil for tasks map
	state := NewAppState(nil, 0, nil, nil, nil)

	// Accessing the tasks map should not panic
	tasks := state.Tasks()
	require.NotNil(t, tasks, "NewAppState() with nil tasks map resulted in nil map, want initialized empty map")

	// Should be able to write to the map without panic
	tasks[1] = []*models.TaskSummary{{ID: 1, Title: "Test Task"}}
	assert.Len(t, tasks, 1)

	// SetTasks should also handle nil
	state.SetTasks(nil)
	tasks = state.Tasks()
	assert.NotNil(t, tasks, "SetTasks(nil) resulted in nil map, want initialized empty map")
}

// TestGetColumnByID_ValidColumn ensures O(1) column lookup works correctly.
// Performance value: Verifies the columnByID map provides correct lookups.
func TestGetColumnByID_ValidColumn(t *testing.T) {
	t.Parallel()
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "In Progress"},
		{ID: 3, Name: "Done"},
	}

	state := NewAppState(nil, 0, columns, nil, nil)

	// Test valid lookups
	col := state.GetColumnByID(2)
	require.NotNil(t, col, "GetColumnByID(2) = nil, want non-nil")
	assert.Equal(t, 2, col.ID)
	assert.Equal(t, "In Progress", col.Name)
}

// TestGetColumnByID_InvalidColumn ensures lookup for non-existent column returns nil.
// Edge case: Querying for a column ID that doesn't exist.
func TestGetColumnByID_InvalidColumn(t *testing.T) {
	t.Parallel()
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "Done"},
	}

	state := NewAppState(nil, 0, columns, nil, nil)

	// Test invalid lookup
	col := state.GetColumnByID(999)
	assert.Nil(t, col)
}

// TestGetColumnByID_EmptyColumns ensures lookup with no columns returns nil.
// Edge case: Application state with no columns loaded.
func TestGetColumnByID_EmptyColumns(t *testing.T) {
	t.Parallel()
	state := NewAppState(nil, 0, []*models.Column{}, nil, nil)

	col := state.GetColumnByID(1)
	assert.Nil(t, col)
}

// TestSetColumns_UpdatesMap ensures SetColumns rebuilds the columnByID map.
// Performance value: Verifies map stays in sync when columns change.
func TestSetColumns_UpdatesMap(t *testing.T) {
	t.Parallel()
	initialColumns := []*models.Column{
		{ID: 1, Name: "Original"},
	}
	state := NewAppState(nil, 0, initialColumns, nil, nil)

	// Verify initial state
	col := state.GetColumnByID(1)
	require.NotNil(t, col, "Initial column lookup failed")
	require.Equal(t, "Original", col.Name, "Initial column lookup failed")

	// Update columns
	newColumns := []*models.Column{
		{ID: 2, Name: "Updated"},
		{ID: 3, Name: "New"},
	}
	state.SetColumns(newColumns)

	// Old column should no longer be found
	col = state.GetColumnByID(1)
	assert.Nil(t, col, "GetColumnByID(1) after SetColumns should return nil, map not updated")

	// New columns should be found
	col = state.GetColumnByID(2)
	require.NotNil(t, col, "GetColumnByID(2) after SetColumns failed, map not rebuilt")
	assert.Equal(t, "Updated", col.Name)

	col = state.GetColumnByID(3)
	require.NotNil(t, col, "GetColumnByID(3) after SetColumns failed, map not rebuilt")
	assert.Equal(t, "New", col.Name)
}
