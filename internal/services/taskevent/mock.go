package taskevent

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

// EventCall records a single method call to the mock
type EventCall struct {
	Method string
	TaskID int
	Args   map[string]interface{}
}

// MockService is a mock implementation of taskevent.Service for testing
type MockService struct {
	mu    sync.Mutex
	Calls []EventCall

	// Optional: allow setting return errors for testing error paths
	CreateTaskCreatedEventErr       error
	CreateTaskMovedEventErr         error
	CreateTaskAssociatedEventErr    error
	CreateTaskDisassociatedEventErr error
	CreateLabelAddedEventErr        error
	CreateLabelRemovedEventErr      error
	CreatePriorityChangedEventErr   error
	CreateTypeChangedEventErr       error
	GetEventsByTaskErr              error
	GetEventsByTaskResult           []models.TaskEvent
}

// NewMockService creates a new mock service
func NewMockService() *MockService {
	return &MockService{
		Calls: make([]EventCall, 0),
	}
}

// Reset clears all recorded calls
func (m *MockService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]EventCall, 0)
}

// GetCalls returns a copy of all recorded calls
func (m *MockService) GetCalls() []EventCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]EventCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called with the given task ID
func (m *MockService) HasCall(method string, taskID int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.Calls {
		if call.Method == method && call.TaskID == taskID {
			return true
		}
	}
	return false
}

// CallCount returns the number of times a method was called
func (m *MockService) CallCount(method string) int {
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

func (m *MockService) recordCall(method string, taskID int, args map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, EventCall{
		Method: method,
		TaskID: taskID,
		Args:   args,
	})
}

func (m *MockService) CreateTaskCreatedEvent(ctx context.Context, qtx types.Querier, taskID int, title, author string) error {
	m.recordCall("CreateTaskCreatedEvent", taskID, map[string]interface{}{
		"title":  title,
		"author": author,
	})
	return m.CreateTaskCreatedEventErr
}

func (m *MockService) CreateTaskMovedEvent(ctx context.Context, qtx types.Querier, taskID int, fromColumn, toColumn, author string) error {
	m.recordCall("CreateTaskMovedEvent", taskID, map[string]interface{}{
		"fromColumn": fromColumn,
		"toColumn":   toColumn,
		"author":     author,
	})
	return m.CreateTaskMovedEventErr
}

func (m *MockService) CreateTaskAssociatedEvent(ctx context.Context, qtx types.Querier, taskID, relatedTaskID int, relatedTitle, relationLabel, author string) error {
	m.recordCall("CreateTaskAssociatedEvent", taskID, map[string]interface{}{
		"relatedTaskID": relatedTaskID,
		"relatedTitle":  relatedTitle,
		"relationLabel": relationLabel,
		"author":        author,
	})
	return m.CreateTaskAssociatedEventErr
}

func (m *MockService) CreateTaskDisassociatedEvent(ctx context.Context, qtx types.Querier, taskID, relatedTaskID int, author string) error {
	m.recordCall("CreateTaskDisassociatedEvent", taskID, map[string]interface{}{
		"relatedTaskID": relatedTaskID,
		"author":        author,
	})
	return m.CreateTaskDisassociatedEventErr
}

func (m *MockService) CreateLabelAddedEvent(ctx context.Context, qtx types.Querier, taskID int, labelName, author string) error {
	m.recordCall("CreateLabelAddedEvent", taskID, map[string]interface{}{
		"labelName": labelName,
		"author":    author,
	})
	return m.CreateLabelAddedEventErr
}

func (m *MockService) CreateLabelRemovedEvent(ctx context.Context, qtx types.Querier, taskID int, labelName, author string) error {
	m.recordCall("CreateLabelRemovedEvent", taskID, map[string]interface{}{
		"labelName": labelName,
		"author":    author,
	})
	return m.CreateLabelRemovedEventErr
}

func (m *MockService) CreatePriorityChangedEvent(ctx context.Context, qtx types.Querier, taskID int, oldPriority, newPriority, author string) error {
	m.recordCall("CreatePriorityChangedEvent", taskID, map[string]interface{}{
		"oldPriority": oldPriority,
		"newPriority": newPriority,
		"author":      author,
	})
	return m.CreatePriorityChangedEventErr
}

func (m *MockService) CreateTypeChangedEvent(ctx context.Context, qtx types.Querier, taskID int, oldType, newType, author string) error {
	m.recordCall("CreateTypeChangedEvent", taskID, map[string]interface{}{
		"oldType": oldType,
		"newType": newType,
		"author":  author,
	})
	return m.CreateTypeChangedEventErr
}

func (m *MockService) GetEventsByTask(ctx context.Context, taskID int) ([]models.TaskEvent, error) {
	m.recordCall("GetEventsByTask", taskID, nil)
	if m.GetEventsByTaskErr != nil {
		return nil, m.GetEventsByTaskErr
	}
	return m.GetEventsByTaskResult, nil
}
