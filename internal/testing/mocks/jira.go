package mocks

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/jira"
)

// Compile-time interface verification
var _ jira.IssueFetcher = (*MockJiraFetcher)(nil)

// MockJiraFetcher is a mock implementation of jira.IssueFetcher for testing.
type MockJiraFetcher struct {
	mu    sync.Mutex
	Calls []MockCall

	FetchIssueErr    error
	FetchIssueResult *jira.Issue
}

// NewMockJiraFetcher creates a new mock Jira fetcher.
func NewMockJiraFetcher() *MockJiraFetcher {
	return &MockJiraFetcher{
		Calls: make([]MockCall, 0),
	}
}

// Reset clears all recorded calls.
func (m *MockJiraFetcher) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
}

// GetCalls returns a copy of all recorded calls.
func (m *MockJiraFetcher) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called.
func (m *MockJiraFetcher) HasCall(method string) bool {
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
func (m *MockJiraFetcher) CallCount(method string) int {
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

func (m *MockJiraFetcher) CheckInstalled() error {
	return nil
}

func (m *MockJiraFetcher) FetchIssue(_ context.Context, issueKey string) (*jira.Issue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{
		Method: "FetchIssue",
		Args: map[string]any{
			"issueKey": issueKey,
		},
	})
	return m.FetchIssueResult, m.FetchIssueErr
}
