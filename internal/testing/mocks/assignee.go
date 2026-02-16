package mocks

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/assignee"
)

// Compile-time interface verification
var _ assignee.Service = (*MockAssigneeService)(nil)

// MockAssigneeService is a mock implementation of assignee.Service for testing.
// It records all method calls for verification in tests.
type MockAssigneeService struct {
	mu    sync.Mutex
	Calls []MockCall

	// Per-method error injection
	ListErr        error
	GetByNameErr   error
	GetByIDErr     error
	CreateErr      error
	GetOrCreateErr error
	DeleteErr      error

	// Per-method result injection
	ListResult        []*models.Assignee
	GetByNameResult   *models.Assignee
	GetByIDResult     *models.Assignee
	CreateResult      *models.Assignee
	GetOrCreateResult *models.Assignee
}

// NewMockAssigneeService creates a new mock assignee service.
func NewMockAssigneeService() *MockAssigneeService {
	return &MockAssigneeService{
		Calls: make([]MockCall, 0),
	}
}

func (m *MockAssigneeService) recordCall(method string, args map[string]any) {
	m.Calls = append(m.Calls, MockCall{
		Method: method,
		Args:   args,
	})
}

// Reset clears all recorded calls.
func (m *MockAssigneeService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
}

// GetCalls returns a copy of all recorded calls.
func (m *MockAssigneeService) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called.
func (m *MockAssigneeService) HasCall(method string) bool {
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
func (m *MockAssigneeService) CallCount(method string) int {
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

func (m *MockAssigneeService) List(ctx context.Context) ([]*models.Assignee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("List", nil)
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	return m.ListResult, nil
}

func (m *MockAssigneeService) GetByName(ctx context.Context, name string) (*models.Assignee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetByName", map[string]any{
		"name": name,
	})
	if m.GetByNameErr != nil {
		return nil, m.GetByNameErr
	}
	return m.GetByNameResult, nil
}

func (m *MockAssigneeService) GetByID(ctx context.Context, id int) (*models.Assignee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetByID", map[string]any{
		"id": id,
	})
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	return m.GetByIDResult, nil
}

func (m *MockAssigneeService) Create(ctx context.Context, name string) (*models.Assignee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("Create", map[string]any{
		"name": name,
	})
	if m.CreateErr != nil {
		return nil, m.CreateErr
	}
	return m.CreateResult, nil
}

func (m *MockAssigneeService) GetOrCreate(ctx context.Context, name string) (*models.Assignee, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetOrCreate", map[string]any{
		"name": name,
	})
	if m.GetOrCreateErr != nil {
		return nil, m.GetOrCreateErr
	}
	return m.GetOrCreateResult, nil
}

func (m *MockAssigneeService) Delete(ctx context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("Delete", map[string]any{
		"id": id,
	})
	return m.DeleteErr
}
