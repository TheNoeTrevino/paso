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
