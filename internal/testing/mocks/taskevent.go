package mocks

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/taskevent"
)

// Compile-time interface verification
var _ taskevent.Service = (*MockTaskEventService)(nil)

// MockTaskEventService is a mock implementation of taskevent.Service for testing.
// It records all method calls and allows injection of return values and errors.
//
// This mock is intended for cross-package tests (CLI, TUI, events, etc.).
// For service-level integration tests, use taskevent.MockService instead to avoid import cycles.
type MockTaskEventService struct {
	mu    sync.Mutex
	Calls []MockCall

	// Error injection for each method
	CreateTaskCreatedEventErr       error
	CreateTaskMovedEventErr         error
	CreateTaskAssociatedEventErr    error
	CreateTaskDisassociatedEventErr error
	CreateLabelAddedEventErr        error
	CreateLabelRemovedEventErr      error
	CreatePriorityChangedEventErr   error
	CreateTypeChangedEventErr       error
	GetEventsByTaskErr              error

	// Result injection
	GetEventsByTaskResult []models.TaskEvent
}

// NewMockTaskEventService creates a new mock task event service.
func NewMockTaskEventService() *MockTaskEventService {
	return &MockTaskEventService{
		Calls: make([]MockCall, 0),
	}
}

func (m *MockTaskEventService) recordCall(method string, taskID int, args map[string]interface{}) {
	if args == nil {
		args = make(map[string]interface{})
	}
	args["taskID"] = taskID // Keep for backward compatibility
	m.Calls = append(m.Calls, MockCall{
		Method: method,
		TaskID: taskID,
		Args:   args,
	})
}

// Reset clears all recorded calls.
func (m *MockTaskEventService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
}

// GetCalls returns a copy of all recorded calls.
func (m *MockTaskEventService) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called with the given task ID.
func (m *MockTaskEventService) HasCall(method string, taskID int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.Calls {
		if call.Method == method && call.TaskID == taskID {
			return true
		}
	}
	return false
}

// CallCount returns the number of times a method was called.
func (m *MockTaskEventService) CallCount(method string) int {
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

// Service interface methods

func (m *MockTaskEventService) CreateTaskCreatedEvent(ctx context.Context, qtx types.Querier, taskID int, title, author string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateTaskCreatedEvent", taskID, map[string]interface{}{
		"title":  title,
		"author": author,
	})
	return m.CreateTaskCreatedEventErr
}

func (m *MockTaskEventService) CreateTaskMovedEvent(ctx context.Context, qtx types.Querier, taskID int, fromColumn, toColumn, author string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateTaskMovedEvent", taskID, map[string]interface{}{
		"fromColumn": fromColumn,
		"toColumn":   toColumn,
		"author":     author,
	})
	return m.CreateTaskMovedEventErr
}

func (m *MockTaskEventService) CreateTaskAssociatedEvent(ctx context.Context, qtx types.Querier, taskID, relatedTaskID int, relatedTitle, relationLabel, author string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateTaskAssociatedEvent", taskID, map[string]interface{}{
		"relatedTaskID": relatedTaskID,
		"relatedTitle":  relatedTitle,
		"relationLabel": relationLabel,
		"author":        author,
	})
	return m.CreateTaskAssociatedEventErr
}

func (m *MockTaskEventService) CreateTaskDisassociatedEvent(ctx context.Context, qtx types.Querier, taskID, relatedTaskID int, author string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateTaskDisassociatedEvent", taskID, map[string]interface{}{
		"relatedTaskID": relatedTaskID,
		"author":        author,
	})
	return m.CreateTaskDisassociatedEventErr
}

func (m *MockTaskEventService) CreateLabelAddedEvent(ctx context.Context, qtx types.Querier, taskID int, labelName, author string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateLabelAddedEvent", taskID, map[string]interface{}{
		"labelName": labelName,
		"author":    author,
	})
	return m.CreateLabelAddedEventErr
}

func (m *MockTaskEventService) CreateLabelRemovedEvent(ctx context.Context, qtx types.Querier, taskID int, labelName, author string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateLabelRemovedEvent", taskID, map[string]interface{}{
		"labelName": labelName,
		"author":    author,
	})
	return m.CreateLabelRemovedEventErr
}

func (m *MockTaskEventService) CreatePriorityChangedEvent(ctx context.Context, qtx types.Querier, taskID int, oldPriority, newPriority, author string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreatePriorityChangedEvent", taskID, map[string]interface{}{
		"oldPriority": oldPriority,
		"newPriority": newPriority,
		"author":      author,
	})
	return m.CreatePriorityChangedEventErr
}

func (m *MockTaskEventService) CreateTypeChangedEvent(ctx context.Context, qtx types.Querier, taskID int, oldType, newType, author string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateTypeChangedEvent", taskID, map[string]interface{}{
		"oldType": oldType,
		"newType": newType,
		"author":  author,
	})
	return m.CreateTypeChangedEventErr
}

func (m *MockTaskEventService) GetEventsByTask(ctx context.Context, taskID int) ([]models.TaskEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetEventsByTask", taskID, nil)
	if m.GetEventsByTaskErr != nil {
		return nil, m.GetEventsByTaskErr
	}
	return m.GetEventsByTaskResult, nil
}
