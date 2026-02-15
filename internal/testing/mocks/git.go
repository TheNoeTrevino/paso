package mocks

import (
	"context"
	"sync"

	"github.com/thenoetrevino/paso/internal/git"
)

// MockGitDetector is a thread-safe mock implementation of git.Detector.
// It allows tests to inject git repository state and branch existence checks.
type MockGitDetector struct {
	mu sync.Mutex

	// Info is the result returned by DetectGitInfo
	Info git.GitInfo

	// Branches maps branch names to existence status
	// If a branch is not in the map, BranchExists returns true (default: exists)
	Branches map[string]bool

	// ValidateBranchNameErr is the error to return from ValidateBranchName
	ValidateBranchNameErr error

	// BranchExistsErr is the error to return from BranchExists
	BranchExistsErr error

	// calls records all method invocations
	calls []MockCall
}

// Compile-time check that MockGitDetector implements git.Detector
var _ git.Detector = (*MockGitDetector)(nil)

// NewMockGitDetector creates a new MockGitDetector with default values.
// By default, it returns an empty GitInfo (not in a repo) and all branches exist.
func NewMockGitDetector() *MockGitDetector {
	return &MockGitDetector{
		Info:     git.GitInfo{}, // default: not in a repo
		Branches: make(map[string]bool),
		calls:    []MockCall{},
	}
}

// DetectGitInfo returns the configured Info.
func (m *MockGitDetector) DetectGitInfo(_ context.Context) git.GitInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("DetectGitInfo", nil)
	return m.Info
}

// ValidateBranchName validates the branch name.
// Returns git.ErrEmptyBranchName if the branch name is empty.
// Returns ValidateBranchNameErr if set, otherwise returns nil.
func (m *MockGitDetector) ValidateBranchName(_ context.Context, branchName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("ValidateBranchName", map[string]any{
		"branchName": branchName,
	})

	// Replicate the standard validation behavior
	if branchName == "" {
		return git.ErrEmptyBranchName
	}

	if m.ValidateBranchNameErr != nil {
		return m.ValidateBranchNameErr
	}

	return nil
}

// BranchExists checks if a branch exists based on the Branches map.
// If the branch is not in the map, it returns true (default: branch exists).
// Returns BranchExistsErr if set.
func (m *MockGitDetector) BranchExists(_ context.Context, branchName string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("BranchExists", map[string]any{
		"branchName": branchName,
	})

	if m.BranchExistsErr != nil {
		return false, m.BranchExistsErr
	}

	exists, ok := m.Branches[branchName]
	if !ok {
		// Default: branch exists (matches old mock behavior)
		return true, nil
	}

	return exists, nil
}

// Reset clears all recorded calls and resets error fields.
// Info and Branches are NOT reset, as tests typically configure these once.
func (m *MockGitDetector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = []MockCall{}
	m.ValidateBranchNameErr = nil
	m.BranchExistsErr = nil
}

// GetCalls returns a copy of all recorded calls.
func (m *MockGitDetector) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()

	calls := make([]MockCall, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// HasCall checks if a method was called with the given name.
func (m *MockGitDetector) HasCall(method string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, call := range m.calls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// CallCount returns the number of times a method was called.
func (m *MockGitDetector) CallCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, call := range m.calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// recordCall records a method call. Must be called with lock held.
func (m *MockGitDetector) recordCall(method string, args map[string]any) {
	m.calls = append(m.calls, MockCall{
		Method: method,
		Args:   args,
	})
}
