package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestUnit_TaskOp_AddTaskKeyOpensForm verifies that pressing the 'a' key
// (default AddTask binding) transitions the model into TicketFormMode.
func TestUnit_TaskOp_AddTaskKeyOpensForm(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Test Project"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)

	// Set terminal size so View() works and mode is NormalMode
	m.UIState.SetWidth(120)
	m.UIState.Height = 40
	require.Equal(t, state.NormalMode, m.UIState.Mode)

	// Press 'a' (default AddTask key)
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "a", Code: 'a'})
	updated := UpdateModelWithMessage(m, keyMsg)

	assert.Equal(t, state.TicketFormMode, updated.UIState.Mode,
		"pressing AddTask key should transition to TicketFormMode")
	assert.NotNil(t, updated.Forms.Form.TaskForm,
		"TaskForm should be initialized after pressing AddTask key")
	assert.Equal(t, 0, updated.Forms.Form.EditingTaskID,
		"EditingTaskID should be 0 for a new task (create mode)")
}

// TestUnit_TaskOp_AddTaskNoColumnsShowsError verifies that pressing the add
// task key when there are no columns shows an error notification instead of
// opening the form.
func TestUnit_TaskOp_AddTaskNoColumnsShowsError(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Test Project"},
	}
	// No columns
	svc.Column.GetColumnsByProjectResult = []*models.Column{}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	keyMsg := tea.KeyPressMsg(tea.Key{Text: "a", Code: 'a'})
	updated := UpdateModelWithMessage(m, keyMsg)

	assert.Equal(t, state.NormalMode, updated.UIState.Mode,
		"should stay in NormalMode when no columns exist")
}

// TestUnit_TaskOp_EditTaskEntersLoadingMode verifies that pressing the edit
// task key ('e') transitions into TicketFormLoadingMode when a task is selected.
func TestUnit_TaskOp_EditTaskEntersLoadingMode(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Test Project"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
		10: {
			{ID: 42, Title: "Test task", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"},
		},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Press 'e' (default EditTask key)
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "e", Code: 'e'})
	updated := UpdateModelWithMessage(m, keyMsg)

	assert.Equal(t, state.TicketFormLoadingMode, updated.UIState.Mode,
		"pressing EditTask key should enter TicketFormLoadingMode")
}

// TestUnit_TaskOp_EditTaskNoTaskShowsError verifies that pressing edit
// when no task is selected stays in NormalMode.
func TestUnit_TaskOp_EditTaskNoTaskShowsError(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Test Project"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	// No tasks in column
	svc.Task.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
		10: {},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	keyMsg := tea.KeyPressMsg(tea.Key{Text: "e", Code: 'e'})
	updated := UpdateModelWithMessage(m, keyMsg)

	assert.Equal(t, state.NormalMode, updated.UIState.Mode,
		"pressing EditTask with no task should stay in NormalMode")
}

// TestUnit_TaskOp_TaskDetailLoadError verifies that a taskDetailForEditError
// message returns the model to NormalMode and handles the error.
func TestUnit_TaskOp_TaskDetailLoadError(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Test Project"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Simulate being in loading mode
	m.UIState.Mode = state.TicketFormLoadingMode
	m.LoadingGitInfo = true

	// Send error message
	errMsg := taskDetailForEditError{err: errors.New("database connection lost")}
	updated := UpdateModelWithMessage(m, errMsg)

	assert.Equal(t, state.NormalMode, updated.UIState.Mode,
		"taskDetailForEditError should return to NormalMode")
	assert.False(t, updated.LoadingGitInfo,
		"loading flag should be cleared after error")
}

// TestUnit_TaskOp_DeleteTaskKeyEntersConfirmMode verifies that pressing 'd'
// enters DeleteConfirmMode when a task is selected.
func TestUnit_TaskOp_DeleteTaskKeyEntersConfirmMode(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Test Project"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
		10: {
			{ID: 42, Title: "Test task", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"},
		},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	keyMsg := tea.KeyPressMsg(tea.Key{Text: "d", Code: 'd'})
	updated := UpdateModelWithMessage(m, keyMsg)

	assert.Equal(t, state.DeleteConfirmMode, updated.UIState.Mode,
		"pressing DeleteTask key should enter DeleteConfirmMode")
}
