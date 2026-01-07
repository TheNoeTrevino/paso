package git

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// GitInfo contains information about the current git repository state
type GitInfo struct {
	IsRepo        bool   // True if current directory is in a git repository
	CurrentBranch string // Name of current branch (empty if detached or not a repo)
	IsDetached    bool   // True if in detached HEAD state
	HasCommits    bool   // True if repository has at least one commit
}

// IsValidForAssociation checks if the git repository state is valid for project association
// Returns true if in a git repo, on a named branch (not detached), and branch name is not empty
func (g GitInfo) IsValidForAssociation() bool {
	return g.IsRepo && g.CurrentBranch != "" && !g.IsDetached
}

// DetectGitInfo detects git repository information from the current working directory
// All git commands are executed with a 2-second timeout for reliability.
// The passed context is respected for cancellation, with per-operation timeouts added.
func DetectGitInfo(ctx context.Context) GitInfo {
	info := GitInfo{
		IsRepo:        false,
		CurrentBranch: "",
		IsDetached:    false,
		HasCommits:    false,
	}

	// Check if we're in a git repository
	if !isGitRepo(ctx) {
		return info
	}
	info.IsRepo = true

	// Check if repository has commits
	info.HasCommits = hasCommits(ctx)

	// Get current branch name and detect detached HEAD
	branchName, isDetached := getCurrentBranch(ctx)
	info.CurrentBranch = SanitizeBranchName(branchName)
	info.IsDetached = isDetached

	return info
}

// isGitRepo checks if the current directory is inside a git repository
// Respects the passed context while adding a 2-second timeout for the git operation
func isGitRepo(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// hasCommits checks if the repository has at least one commit
// Respects the passed context while adding a 2-second timeout for the git operation
func hasCommits(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	err := cmd.Run()
	return err == nil
}

// getCurrentBranch gets the current branch name and detects detached HEAD state
// Returns (branchName, isDetached)
// Respects the passed context while adding a 2-second timeout for the git operation
func getCurrentBranch(ctx context.Context) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Try to get symbolic ref (branch name)
	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "--short", "HEAD")
	output, err := cmd.Output()

	if err != nil {
		// If symbolic-ref fails, we're in detached HEAD state
		return "", true
	}

	// Parse branch name (remove trailing newline)
	branchName := strings.TrimSpace(string(output))
	return branchName, false
}

// SanitizeBranchName sanitizes a git branch name for safe storage
// - Trims leading/trailing whitespace
// - Truncates to maximum 255 characters (database column limit)
// - Preserves valid git branch characters (slashes, hyphens, underscores, dots, etc.)
func SanitizeBranchName(name string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(name)

	// Truncate to 255 characters (database TEXT column limit)
	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
	}

	return sanitized
}
