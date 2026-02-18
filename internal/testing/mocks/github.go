package mocks

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/github"
)

// Compile-time interface verification
var _ github.IssueFetcher = (*MockGitHubFetcher)(nil)

// MockGitHubFetcher is a mock implementation of github.IssueFetcher for testing.
type MockGitHubFetcher struct {
	mu    sync.Mutex
	Calls []MockCall

	FetchIssueErr    error
	FetchIssueResult *github.Issue
}

// NewMockGitHubFetcher creates a new mock GitHub fetcher.
func NewMockGitHubFetcher() *MockGitHubFetcher {
	return &MockGitHubFetcher{
		Calls: make([]MockCall, 0),
	}
}

// Reset clears all recorded calls.
func (m *MockGitHubFetcher) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
}

// GetCalls returns a copy of all recorded calls.
func (m *MockGitHubFetcher) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called.
func (m *MockGitHubFetcher) HasCall(method string) bool {
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
func (m *MockGitHubFetcher) CallCount(method string) int {
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

func (m *MockGitHubFetcher) FetchIssue(_ context.Context, issueNumber int) (*github.Issue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{
		Method: "FetchIssue",
		Args: map[string]any{
			"issueNumber": issueNumber,
		},
	})
	return m.FetchIssueResult, m.FetchIssueErr
}
