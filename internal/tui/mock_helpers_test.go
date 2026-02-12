package tui

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

// MockServices holds mock service instances for TUI testing.
type MockServices struct {
	Task     *mocks.MockTaskService
	Project  *mocks.MockProjectService
	Column   *mocks.MockColumnService
	Label    *mocks.MockLabelService
	Assignee *mocks.MockAssigneeService
}

// NewDefaultMockServices creates a set of mock services with sensible defaults.
// All list/query methods return empty slices so InitialModel won't nil-panic.
func NewDefaultMockServices() *MockServices {
	taskSvc := mocks.NewMockTaskService()
	taskSvc.GetTaskSummariesByProjectResult = make(map[int][]*models.TaskSummary)

	projectSvc := mocks.NewMockProjectService()
	projectSvc.GetAllProjectsResult = []*models.Project{}

	columnSvc := mocks.NewMockColumnService()
	columnSvc.GetColumnsByProjectResult = []*models.Column{}

	labelSvc := mocks.NewMockLabelService()
	labelSvc.GetLabelsByProjectResult = []*models.Label{}

	assigneeSvc := mocks.NewMockAssigneeService()
	assigneeSvc.ListResult = []*models.Assignee{}

	return &MockServices{
		Task:     taskSvc,
		Project:  projectSvc,
		Column:   columnSvc,
		Label:    labelSvc,
		Assignee: assigneeSvc,
	}
}

// SetupTestModelWithMocks creates a TUI Model backed by mock services.
// It builds an *app.App with the provided mocks and calls InitialModel
// with no event client and no database connection.
func SetupTestModelWithMocks(t *testing.T, svc *MockServices) Model {
	t.Helper()

	appContainer := &app.App{
		TaskService:     svc.Task,
		ProjectService:  svc.Project,
		ColumnService:   svc.Column,
		LabelService:    svc.Label,
		AssigneeService: svc.Assignee,
	}

	cfg, err := config.Load()
	require.NoError(t, err, "failed to load config")

	ctx := context.Background()
	m := InitialModel(ctx, appContainer, cfg, nil, nil)
	return m
}

func TestSetupTestModelWithMocks(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()
	m := SetupTestModelWithMocks(t, svc)

	require.NotNil(t, m.App)
	require.NotNil(t, m.AppState)
	require.NotNil(t, m.UIState)
	require.NotNil(t, m.Config)

	// InitialModel should have called these during init
	require.True(t, svc.Project.HasCall("GetAllProjects"), "expected GetAllProjects to be called during init")
	require.True(t, svc.Column.HasCall("GetColumnsByProject"), "expected GetColumnsByProject to be called during init")
	require.True(t, svc.Label.HasCall("GetLabelsByProject"), "expected GetLabelsByProject to be called during init")
}
