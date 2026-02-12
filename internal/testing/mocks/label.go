package mocks

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/label"
)

// Compile-time interface verification
var _ label.Service = (*MockLabelService)(nil)

// MockLabelService is a mock implementation of label.Service for testing.
// It records all method calls for verification in tests.
type MockLabelService struct {
	mu    sync.Mutex
	Calls []MockCall

	// Per-method error injection
	GetLabelsByProjectErr error
	GetLabelsForTaskErr   error
	CountTasksByLabelErr  error
	CreateLabelErr        error
	UpdateLabelErr        error
	DeleteLabelErr        error

	// Per-method result injection
	GetLabelsByProjectResult []*models.Label
	GetLabelsForTaskResult   []*models.Label
	CountTasksByLabelResult  int
	CreateLabelResult        *models.Label
}

// NewMockLabelService creates a new mock label service.
func NewMockLabelService() *MockLabelService {
	return &MockLabelService{
		Calls: make([]MockCall, 0),
	}
}

func (m *MockLabelService) recordCall(method string, args map[string]interface{}) {
	m.Calls = append(m.Calls, MockCall{
		Method: method,
		Args:   args,
	})
}

// Reset clears all recorded calls.
func (m *MockLabelService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
}

// GetCalls returns a copy of all recorded calls.
func (m *MockLabelService) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called.
func (m *MockLabelService) HasCall(method string) bool {
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
func (m *MockLabelService) CallCount(method string) int {
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

func (m *MockLabelService) GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetLabelsByProject", map[string]interface{}{
		"projectID": projectID,
	})
	if m.GetLabelsByProjectErr != nil {
		return nil, m.GetLabelsByProjectErr
	}
	return m.GetLabelsByProjectResult, nil
}

func (m *MockLabelService) GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetLabelsForTask", map[string]interface{}{
		"taskID": taskID,
	})
	if m.GetLabelsForTaskErr != nil {
		return nil, m.GetLabelsForTaskErr
	}
	return m.GetLabelsForTaskResult, nil
}

func (m *MockLabelService) CountTasksByLabel(ctx context.Context, labelID int) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CountTasksByLabel", map[string]interface{}{
		"labelID": labelID,
	})
	if m.CountTasksByLabelErr != nil {
		return 0, m.CountTasksByLabelErr
	}
	return m.CountTasksByLabelResult, nil
}

func (m *MockLabelService) CreateLabel(ctx context.Context, req label.CreateLabelRequest) (*models.Label, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateLabel", map[string]interface{}{
		"projectID": req.ProjectID,
		"name":      req.Name,
		"color":     req.Color,
	})
	if m.CreateLabelErr != nil {
		return nil, m.CreateLabelErr
	}
	return m.CreateLabelResult, nil
}

func (m *MockLabelService) UpdateLabel(ctx context.Context, req label.UpdateLabelRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("UpdateLabel", map[string]interface{}{
		"id":    req.ID,
		"name":  req.Name,
		"color": req.Color,
	})
	return m.UpdateLabelErr
}

func (m *MockLabelService) DeleteLabel(ctx context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DeleteLabel", map[string]interface{}{
		"id": id,
	})
	return m.DeleteLabelErr
}
