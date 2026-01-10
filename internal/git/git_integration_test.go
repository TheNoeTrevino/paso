package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Integration Test Helpers
// ============================================================================

// setupGitRepo creates a temporary git repository for testing
func setupGitRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	ctx := context.Background()

	// Initialize git repo
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = tmpDir
	err := cmd.Run()
	require.NoError(t, err, "Failed to initialize git repository")

	// Configure git for testing
	configCmds := [][]string{
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
	}

	for _, cmdArgs := range configCmds {
		cfgCmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		cfgCmd.Dir = tmpDir
		err := cfgCmd.Run()
		require.NoError(t, err, "Failed to configure git")
	}

	return tmpDir
}

// createCommit creates a commit in the git repository
func createCommit(t *testing.T, repoDir, message string) {
	t.Helper()

	ctx := context.Background()

	// Create a file
	filePath := filepath.Join(repoDir, "test.txt")
	err := os.WriteFile(filePath, []byte("test content\n"), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Add file
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = repoDir
	err = cmd.Run()
	require.NoError(t, err, "Failed to git add")

	// Commit
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	commitCmd.Dir = repoDir
	err = commitCmd.Run()
	require.NoError(t, err, "Failed to git commit")
}

// createBranch creates a new branch in the git repository
func createBranch(t *testing.T, repoDir, branchName string) {
	t.Helper()

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
	cmd.Dir = repoDir
	err := cmd.Run()
	require.NoError(t, err, "Failed to create branch")
}

// checkoutBranch checks out an existing branch
func checkoutBranch(t *testing.T, repoDir, branchName string) {
	t.Helper()

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "checkout", branchName)
	cmd.Dir = repoDir
	err := cmd.Run()
	require.NoError(t, err, "Failed to checkout branch")
}

// detachHead puts the repository in detached HEAD state
func detachHead(t *testing.T, repoDir string) {
	t.Helper()

	ctx := context.Background()

	// Get the current commit hash
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to get HEAD commit")

	commitHash := string(output[:len(output)-1]) // Remove newline

	// Checkout the commit directly (detached HEAD)
	checkoutCmd := exec.CommandContext(ctx, "git", "checkout", commitHash)
	checkoutCmd.Dir = repoDir
	err = checkoutCmd.Run()
	require.NoError(t, err, "Failed to detach HEAD")
}

// ============================================================================
// DetectGitInfo Integration Tests
// ============================================================================

func TestDetectGitInfo_NormalRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repoDir := setupGitRepo(t)
	createCommit(t, repoDir, "Initial commit")

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(repoDir)
	require.NoError(t, err)

	// Test DetectGitInfo
	ctx := context.Background()
	info := DetectGitInfo(ctx)

	// Assertions for normal repo with commits
	assert.True(t, info.IsRepo, "Should detect as git repository")
	assert.True(t, info.HasCommits, "Should detect commits")
	assert.False(t, info.IsDetached, "Should not be in detached HEAD state")
	assert.NotEmpty(t, info.CurrentBranch, "Should have a current branch")
	// Git init creates 'master' or 'main' depending on git config
	assert.Contains(t, []string{"master", "main"}, info.CurrentBranch, "Should be on default branch")
}

func TestDetectGitInfo_FeatureBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repoDir := setupGitRepo(t)
	createCommit(t, repoDir, "Initial commit")
	createBranch(t, repoDir, "feature/my-awesome-feature")

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(repoDir)
	require.NoError(t, err)

	ctx := context.Background()
	info := DetectGitInfo(ctx)

	assert.True(t, info.IsRepo, "Should detect as git repository")
	assert.True(t, info.HasCommits, "Should detect commits")
	assert.False(t, info.IsDetached, "Should not be in detached HEAD state")
	assert.Equal(t, "feature/my-awesome-feature", info.CurrentBranch, "Should detect feature branch")
}

func TestDetectGitInfo_DetachedHead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repoDir := setupGitRepo(t)
	createCommit(t, repoDir, "Initial commit")
	detachHead(t, repoDir)

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(repoDir)
	require.NoError(t, err)

	ctx := context.Background()
	info := DetectGitInfo(ctx)

	assert.True(t, info.IsRepo, "Should detect as git repository")
	assert.True(t, info.HasCommits, "Should detect commits")
	assert.True(t, info.IsDetached, "Should be in detached HEAD state")
	assert.Empty(t, info.CurrentBranch, "Should not have a current branch in detached state")
}

func TestDetectGitInfo_EmptyRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repoDir := setupGitRepo(t)
	// Don't create any commits

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(repoDir)
	require.NoError(t, err)

	ctx := context.Background()
	info := DetectGitInfo(ctx)

	assert.True(t, info.IsRepo, "Should detect as git repository")
	assert.False(t, info.HasCommits, "Should not detect commits in empty repo")
	assert.False(t, info.IsDetached, "Empty repo should not be detached")
	// Empty repos might have a default branch name or empty string
	// Implementation will determine the behavior
}

func TestDetectGitInfo_NotARepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	// Don't initialize git in this directory

	// Change to non-repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	info := DetectGitInfo(ctx)

	assert.False(t, info.IsRepo, "Should not detect as git repository")
	assert.False(t, info.HasCommits, "Non-repo should not have commits")
	assert.False(t, info.IsDetached, "Non-repo should not be detached")
	assert.Empty(t, info.CurrentBranch, "Non-repo should not have a branch")
}

func TestDetectGitInfo_BareRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	ctx := context.Background()

	// Initialize bare repository
	cmd := exec.CommandContext(ctx, "git", "init", "--bare")
	cmd.Dir = tmpDir
	err := cmd.Run()
	require.NoError(t, err, "Failed to initialize bare repository")

	// Change to bare repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	info := DetectGitInfo(ctx)

	// Bare repositories should be handled gracefully
	// The implementation might detect it as a repo or not
	// Either way, it should not crash
	assert.NotNil(t, info, "Should return a GitInfo struct for bare repo")
}

func TestDetectGitInfo_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repoDir := setupGitRepo(t)
	createCommit(t, repoDir, "Initial commit")

	// Change to repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(repoDir)
	require.NoError(t, err)

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(1 * time.Millisecond)

	// DetectGitInfo should handle timeout gracefully
	info := DetectGitInfo(ctx)

	// Should return valid GitInfo even with timeout
	// Implementation should use its own timeout (2 seconds per plan)
	assert.NotNil(t, info, "Should return GitInfo even with expired context")
}

func TestDetectGitInfo_BranchWithSpecialChars(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repoDir := setupGitRepo(t)
	createCommit(t, repoDir, "Initial commit")

	// Test various branch names with special characters
	branchNames := []string{
		"feature/AUTH-123-user-login",
		"fix/bug-in-parser",
		"release/v1.0.0",
		"feature/user_authentication",
		"hotfix/critical-bug",
	}

	for _, branchName := range branchNames {
		t.Run(branchName, func(t *testing.T) {
			createBranch(t, repoDir, branchName)

			// Change to repo directory
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(originalDir)
				assert.NoError(t, err)
			}()
			err = os.Chdir(repoDir)
			require.NoError(t, err)

			ctx := context.Background()
			info := DetectGitInfo(ctx)

			assert.True(t, info.IsRepo, "Should detect as git repository")
			assert.Equal(t, branchName, info.CurrentBranch, "Should detect branch name correctly")

			// Switch back to main/master for next iteration
			checkoutBranch(t, repoDir, "master")
		})
	}
}

func TestDetectGitInfo_Subdirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repoDir := setupGitRepo(t)
	createCommit(t, repoDir, "Initial commit")
	createBranch(t, repoDir, "feature/test")

	// Create a subdirectory
	subDir := filepath.Join(repoDir, "subdir", "nested")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err, "Failed to create subdirectory")

	// Change to subdirectory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(subDir)
	require.NoError(t, err)

	ctx := context.Background()
	info := DetectGitInfo(ctx)

	// Should detect git info even from subdirectory
	assert.True(t, info.IsRepo, "Should detect as git repository from subdirectory")
	assert.Equal(t, "feature/test", info.CurrentBranch, "Should detect branch from subdirectory")
}

func TestDetectGitInfo_MultipleWorktrees(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test verifies behavior with git worktrees
	// Each worktree should be detected independently
	repoDir := setupGitRepo(t)
	createCommit(t, repoDir, "Initial commit")
	createBranch(t, repoDir, "feature/branch1")
	checkoutBranch(t, repoDir, "master")

	// Create a worktree (requires git 2.5+)
	worktreeDir := filepath.Join(t.TempDir(), "worktree")
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", worktreeDir, "feature/branch1")
	cmd.Dir = repoDir
	err := cmd.Run()
	if err != nil {
		// Git worktree might not be available in all environments
		t.Skip("Git worktree not available, skipping test")
	}

	// Test from main repo
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	infoMain := DetectGitInfo(ctx)
	assert.Contains(t, []string{"master", "main"}, infoMain.CurrentBranch, "Should detect main branch in main repo")

	// Test from worktree
	err = os.Chdir(worktreeDir)
	require.NoError(t, err)

	infoWorktree := DetectGitInfo(ctx)
	assert.Equal(t, "feature/branch1", infoWorktree.CurrentBranch, "Should detect feature branch in worktree")
}

// ============================================================================
// Edge Cases and Error Handling
// ============================================================================

func TestDetectGitInfo_VeryLongBranchName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repoDir := setupGitRepo(t)
	createCommit(t, repoDir, "Initial commit")

	// Create a branch with a very long name (but valid for git)
	longBranchName := "feature/" + string(make([]byte, 200))
	for i := range longBranchName[8:] {
		longBranchName = longBranchName[:8+i] + "a" + longBranchName[8+i+1:]
	}

	createBranch(t, repoDir, longBranchName)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(repoDir)
	require.NoError(t, err)

	ctx := context.Background()
	info := DetectGitInfo(ctx)

	assert.True(t, info.IsRepo, "Should detect as git repository")
	assert.NotEmpty(t, info.CurrentBranch, "Should detect long branch name")
	sanitized, err := SanitizeBranchName(info.CurrentBranch)
	assert.NoError(t, err, "Should not error on sanitizing branch name")
	assert.LessOrEqual(t, len(sanitized), 255, "Sanitized branch name should be <= 255 characters")
}

func TestDetectGitInfo_CorruptedRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repoDir := setupGitRepo(t)
	createCommit(t, repoDir, "Initial commit")

	// Corrupt the repository by removing .git/HEAD
	headPath := filepath.Join(repoDir, ".git", "HEAD")
	err := os.Remove(headPath)
	require.NoError(t, err, "Failed to corrupt repository")

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		assert.NoError(t, err)
	}()
	err = os.Chdir(repoDir)
	require.NoError(t, err)

	ctx := context.Background()
	info := DetectGitInfo(ctx)

	// Should handle corrupted repo gracefully
	// Might detect as repo but fail on other operations
	assert.NotNil(t, info, "Should return GitInfo even for corrupted repo")
}
