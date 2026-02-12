package project

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/git"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
)

// testEnv encapsulates common test dependencies for project service tests.
type testEnv struct {
	DB      *sql.DB
	Dialect fixtures.Dialect
	Svc     Service
	GitMock *mockGitDetector
	Ctx     context.Context
}

// setupTestEnv creates a new test environment with all necessary dependencies.
func setupTestEnv(tb testing.TB) *testEnv {
	tb.Helper()
	db := fixtures.SetupTestDB(tb)
	dialect := fixtures.SQLiteDialect()
	gitMock := newMockGitDetector()
	svc, err := NewService(db, database.SQLite, nil, gitMock)
	require.NoError(tb, err, "failed to create test service")
	return &testEnv{
		DB:      db,
		Dialect: dialect,
		Svc:     svc,
		GitMock: gitMock,
		Ctx:     context.Background(),
	}
}

// mockGitDetector is a mock implementation of git.Detector for testing
type mockGitDetector struct {
	branches map[string]bool
	gitInfo  git.GitInfo
}

func newMockGitDetector() *mockGitDetector {
	return &mockGitDetector{
		branches: make(map[string]bool),
		gitInfo:  git.GitInfo{}, // default: not in a repo
	}
}

func (m *mockGitDetector) DetectGitInfo(_ context.Context) git.GitInfo {
	return m.gitInfo
}

func (m *mockGitDetector) ValidateBranchName(_ context.Context, branchName string) error {
	if strings.TrimSpace(branchName) == "" {
		return git.ErrEmptyBranchName
	}
	return nil
}

func (m *mockGitDetector) BranchExists(_ context.Context, branchName string) (bool, error) {
	exists, ok := m.branches[branchName]
	if !ok {
		return true, nil
	}
	return exists, nil
}
