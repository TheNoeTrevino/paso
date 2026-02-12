package mocks

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/column"
)

// Compile-time interface verification
var _ column.Service = (*MockColumnService)(nil)

// MockColumnService is a mock implementation of column.Service for testing.
// It records all method calls for verification in tests.
type MockColumnService struct {
	mu    sync.Mutex
	Calls []MockCall

	// Per-method error injection
	GetColumnsByProjectErr     error
	GetColumnByIDErr           error
	CreateColumnErr            error
	UpdateColumnNameErr        error
	SetHoldsReadyTasksErr      error
	SetHoldsCompletedTasksErr  error
	SetHoldsInProgressTasksErr error
	DeleteColumnErr            error

	// Per-method result injection
	GetColumnsByProjectResult     []*models.Column
	GetColumnByIDResult           *models.Column
	CreateColumnResult            *models.Column
	SetHoldsReadyTasksResult      *models.Column
	SetHoldsCompletedTasksResult  *models.Column
	SetHoldsInProgressTasksResult *models.Column
}

// NewMockColumnService creates a new mock column service.
func NewMockColumnService() *MockColumnService {
	return &MockColumnService{
		Calls: make([]MockCall, 0),
	}
}

// Reset clears all recorded calls.
func (m *MockColumnService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
}

// GetCalls returns a copy of all recorded calls.
func (m *MockColumnService) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called.
func (m *MockColumnService) HasCall(method string) bool {
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
func (m *MockColumnService) CallCount(method string) int {
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

func (m *MockColumnService) recordCall(method string, args map[string]interface{}) {
	m.Calls = append(m.Calls, MockCall{
		Method: method,
		Args:   args,
	})
}

func (m *MockColumnService) GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetColumnsByProject", map[string]interface{}{
		"projectID": projectID,
	})
	if m.GetColumnsByProjectErr != nil {
		return nil, m.GetColumnsByProjectErr
	}
	return m.GetColumnsByProjectResult, nil
}

func (m *MockColumnService) GetColumnByID(ctx context.Context, id int) (*models.Column, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetColumnByID", map[string]interface{}{
		"id": id,
	})
	if m.GetColumnByIDErr != nil {
		return nil, m.GetColumnByIDErr
	}
	return m.GetColumnByIDResult, nil
}

func (m *MockColumnService) CreateColumn(ctx context.Context, req column.CreateColumnRequest) (*models.Column, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateColumn", map[string]interface{}{
		"name":                 req.Name,
		"projectID":            req.ProjectID,
		"afterID":              req.AfterID,
		"holdsReadyTasks":      req.HoldsReadyTasks,
		"holdsCompletedTasks":  req.HoldsCompletedTasks,
		"holdsInProgressTasks": req.HoldsInProgressTasks,
	})
	if m.CreateColumnErr != nil {
		return nil, m.CreateColumnErr
	}
	return m.CreateColumnResult, nil
}

func (m *MockColumnService) UpdateColumnName(ctx context.Context, id int, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("UpdateColumnName", map[string]interface{}{
		"id":   id,
		"name": name,
	})
	return m.UpdateColumnNameErr
}

func (m *MockColumnService) SetHoldsReadyTasks(ctx context.Context, columnID int) (*models.Column, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("SetHoldsReadyTasks", map[string]interface{}{
		"columnID": columnID,
	})
	if m.SetHoldsReadyTasksErr != nil {
		return nil, m.SetHoldsReadyTasksErr
	}
	return m.SetHoldsReadyTasksResult, nil
}

func (m *MockColumnService) SetHoldsCompletedTasks(ctx context.Context, columnID int, force bool) (*models.Column, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("SetHoldsCompletedTasks", map[string]interface{}{
		"columnID": columnID,
		"force":    force,
	})
	if m.SetHoldsCompletedTasksErr != nil {
		return nil, m.SetHoldsCompletedTasksErr
	}
	return m.SetHoldsCompletedTasksResult, nil
}

func (m *MockColumnService) SetHoldsInProgressTasks(ctx context.Context, columnID int) (*models.Column, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("SetHoldsInProgressTasks", map[string]interface{}{
		"columnID": columnID,
	})
	if m.SetHoldsInProgressTasksErr != nil {
		return nil, m.SetHoldsInProgressTasksErr
	}
	return m.SetHoldsInProgressTasksResult, nil
}

func (m *MockColumnService) DeleteColumn(ctx context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DeleteColumn", map[string]interface{}{
		"id": id,
	})
	return m.DeleteColumnErr
}
