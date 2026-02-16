package mocks

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/project"
)

// Compile-time interface verification
var _ project.Service = (*MockProjectService)(nil)

// MockProjectService is a mock implementation of project.Service for testing.
// It records all method calls and supports per-method error/result injection.
type MockProjectService struct {
	mu    sync.Mutex
	Calls []MockCall

	// Per-method error injection
	GetAllProjectsErr        error
	GetProjectByIDErr        error
	GetProjectByGitBranchErr error
	GetTaskCountErr          error
	CreateProjectErr         error
	UpdateProjectErr         error
	DeleteProjectErr         error

	// Per-method result injection
	GetAllProjectsResult        []*models.Project
	GetProjectByIDResult        *models.Project
	GetProjectByGitBranchResult *models.Project
	GetTaskCountResult          int
	CreateProjectResult         *models.Project
}

// NewMockProjectService creates a new mock project service.
func NewMockProjectService() *MockProjectService {
	return &MockProjectService{
		Calls: make([]MockCall, 0),
	}
}

// Reset clears all recorded calls.
func (m *MockProjectService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
}

// GetCalls returns a copy of all recorded calls.
func (m *MockProjectService) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called.
func (m *MockProjectService) HasCall(method string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.Calls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// CallCount returns the number of times a method was called.
func (m *MockProjectService) CallCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

func (m *MockProjectService) recordCall(method string, args map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{
		Method: method,
		Args:   args,
	})
}

func (m *MockProjectService) GetAllProjects(ctx context.Context) ([]*models.Project, error) {
	m.recordCall("GetAllProjects", nil)
	if m.GetAllProjectsErr != nil {
		return nil, m.GetAllProjectsErr
	}
	return m.GetAllProjectsResult, nil
}

func (m *MockProjectService) GetProjectByID(ctx context.Context, id int) (*models.Project, error) {
	m.recordCall("GetProjectByID", map[string]any{
		"id": id,
	})
	if m.GetProjectByIDErr != nil {
		return nil, m.GetProjectByIDErr
	}
	return m.GetProjectByIDResult, nil
}

func (m *MockProjectService) GetProjectByGitBranch(ctx context.Context, gitBranch string) (*models.Project, error) {
	m.recordCall("GetProjectByGitBranch", map[string]any{
		"gitBranch": gitBranch,
	})
	if m.GetProjectByGitBranchErr != nil {
		return nil, m.GetProjectByGitBranchErr
	}
	return m.GetProjectByGitBranchResult, nil
}

func (m *MockProjectService) GetTaskCount(ctx context.Context, projectID int) (int, error) {
	m.recordCall("GetTaskCount", map[string]any{
		"projectID": projectID,
	})
	if m.GetTaskCountErr != nil {
		return 0, m.GetTaskCountErr
	}
	return m.GetTaskCountResult, nil
}

func (m *MockProjectService) CreateProject(ctx context.Context, req project.CreateProjectRequest) (*models.Project, error) {
	m.recordCall("CreateProject", map[string]any{
		"name":        req.Name,
		"description": req.Description,
		"gitBranch":   req.GitBranch,
	})
	if m.CreateProjectErr != nil {
		return nil, m.CreateProjectErr
	}
	return m.CreateProjectResult, nil
}

func (m *MockProjectService) UpdateProject(ctx context.Context, req project.UpdateProjectRequest) error {
	m.recordCall("UpdateProject", map[string]any{
		"id":          req.ID,
		"name":        req.Name,
		"description": req.Description,
		"gitBranch":   req.GitBranch,
	})
	return m.UpdateProjectErr
}

func (m *MockProjectService) DeleteProject(ctx context.Context, id int, force bool) error {
	m.recordCall("DeleteProject", map[string]any{
		"id":    id,
		"force": force,
	})
	return m.DeleteProjectErr
}
