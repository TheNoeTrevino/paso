package git

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

var ErrEmptyBranchName = errors.New("branch name cannot be empty")

// CACHE DESIGN (Task #27)
//
// This package will implement caching for git operations to improve TUI performance.
//
// Cache Structure:
//   - Location: internal/git/cache.go (separate file in this package)
//   - Key: Absolute working directory path (supports multi-repo workflows)
//   - Data: CacheEntry{GitInfo, []BranchInfo, Timestamp}
//   - TTL: 45 seconds (configurable via DefaultCacheTTL constant)
//   - Thread Safety: sync.RWMutex for concurrent read/exclusive write access
//
// Cache Operations:
//   - Get(ctx, workDir) -> (GitInfo, []BranchInfo, hit bool)
//   - Set(workDir, GitInfo, []BranchInfo)
//   - Invalidate(workDir) - per-directory invalidation
//   - InvalidateAll() - global cache clear
//
// Integration with TUI:
//   - Cache instance stored in Model.GitCache
//   - Async commands (detectGitInfoCmd, listBranchesCmd) check cache first
//   - Cache hit: return cached data immediately (fast path)
//   - Cache miss/expired: execute git commands, update cache, return data (slow path)
//
// Invalidation Strategy:
//   - Automatic: TTL expiry checked on Get() - if time.Since(timestamp) > TTL, treat as miss
//   - Manual: Ctrl+R triggers cache.Invalidate(currentWorkDir) and re-fetches git data
//   - No background timers: check expiry on access only (simpler, less overhead)
//
// Why This Design:
//   - Separate package: testable, follows single responsibility, loosely coupled
//   - Working directory key: natural multi-repo support, aligns with git semantics
//   - 45s TTL: balances freshness vs performance (most UI interactions happen in bursts)
//   - RWMutex: standard Go pattern for read-heavy concurrent access
//   - Async integration: seamless with existing Bubble Tea tea.Cmd pattern
//
// See cache_design.md for full design document.

// GitInfo contains information about the current git repository state
type GitInfo struct {
	IsRepo        bool   // True if current directory is in a git repository
	CurrentBranch string // Name of current branch (empty if detached or not a repo)
	IsDetached    bool   // True if in detached HEAD state
	HasCommits    bool   // True if repository has at least one commit
}

// BranchInfo contains information about a git branch
type BranchInfo struct {
	Name   string // Branch name (sanitized)
	Marker string // Visual marker: "* " (current), "+ " (worktree), "  " (normal)
}

// IsValidForAssociation checks if the git repository state is valid for project association
// Returns true if in a git repo, on a named branch (not detached), and branch name is not empty
func (g GitInfo) IsValidForAssociation() bool {
	return g.IsRepo && g.CurrentBranch != "" && !g.IsDetached
}

// DetectGitInfo detects git repository information from the current working directory
// All git commands are executed with a 5-second timeout for reliability.
// The passed context is respected for cancellation, with per-operation timeouts added.
func DetectGitInfo(ctx context.Context) GitInfo {
	info := GitInfo{
		IsRepo:        false,
		CurrentBranch: "",
		IsDetached:    false,
		HasCommits:    false,
	}

	// Batch check: is repo + has commits in single command (reduces exec overhead from 2 calls to 1)
	isRepo, hasCommits := checkRepoAndCommits(ctx)
	if !isRepo {
		return info
	}
	info.IsRepo = true
	info.HasCommits = hasCommits

	// Get current branch name and detect detached HEAD
	branchName, isDetached := getCurrentBranch(ctx)
	sanitized, err := SanitizeBranchName(branchName)
	if err != nil {
		sanitized = ""
	}
	info.CurrentBranch = sanitized
	info.IsDetached = isDetached

	return info
}

// checkRepoAndCommits checks if we're in a git repo and if it has commits in a single exec call.
// Returns (isRepo, hasCommits). If isRepo is false, hasCommits is meaningless.
// Uses "git rev-parse --git-dir HEAD" which succeeds fully only if both conditions are met.
func checkRepoAndCommits(ctx context.Context) (isRepo bool, hasCommits bool) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// --git-dir succeeds if in a repo, HEAD succeeds if commits exist
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// Could be "not a repo" or "no commits" - need to distinguish
		fallbackCtx, fallbackCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer fallbackCancel()
		repoCheck := exec.CommandContext(fallbackCtx, "git", "rev-parse", "--git-dir")
		if repoCheck.Run() != nil {
			return false, false
		}
		return true, false
	}

	// Both succeeded if we got output containing both paths
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return true, len(lines) >= 2
}

// BranchExists checks if a branch with the given name exists in the repository
// Respects the passed context while adding a 5-second timeout for the git operation
func BranchExists(ctx context.Context, branchName string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "refs/heads/"+branchName)
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// getCurrentBranch gets the current branch name and detects detached HEAD state
// Returns (branchName, isDetached)
// Respects the passed context while adding a 5-second timeout for the git operation
// TODO: should we add typesafety to this boolean?
func getCurrentBranch(ctx context.Context) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to get symbolic ref (branch name)
	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "--short", "HEAD")
	rawBranchname, err := cmd.Output()
	if err != nil {
		// If symbolic-ref fails, we're in detached HEAD state
		return "", true
	}

	// Parse branch name (remove trailing newline)
	branchName := strings.TrimSpace(string(rawBranchname))
	return branchName, false
}

// ListBranches returns a list of all local git branches with their status markers.
// Branch names are sanitized for safe storage.
// Markers: "* " (current branch), "+ " (worktree branch), "  " (normal branch)
// Respects the passed context while adding a 5-second timeout for the git operation.
func ListBranches(ctx context.Context) ([]BranchInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// let git handle sorting
	cmd := exec.CommandContext(ctx, "git", "branch", "--list", "--sort=refname")
	rawBranches, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(rawBranches), "\n")
	branches := make([]BranchInfo, 0, len(lines))

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		// Extract branch marker (first 2 characters: "* ", "+ ", or "  ")
		marker := line[:2]
		branchName := strings.TrimSpace(line[2:])

		if branchName == "" {
			continue
		}

		sanitized, err := SanitizeBranchName(branchName)
		if err != nil || sanitized == "" {
			slog.Debug("skipping invalid branch name",
				"raw_name", branchName,
				"error", err)
			continue
		}

		branches = append(branches, BranchInfo{
			Name:   sanitized,
			Marker: marker,
		})
	}

	return branches, nil
}

// SanitizeBranchName sanitizes a git branch name for safe storage
// - Trims leading/trailing whitespace
// - Rejects control characters (newlines, tabs, null bytes, carriage returns)
// - Rejects git-invalid characters: * ? [ ] \ ^ ~ : space
// - Rejects git-invalid sequences: .. @{ //
// - Truncates to maximum 255 characters (database column limit)
// - Preserves valid git branch characters (slashes, hyphens, underscores, dots, etc.)
func SanitizeBranchName(name string) (string, error) {
	sanitized := strings.TrimSpace(name)

	for _, r := range sanitized {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("branch name contains control character: %q", r)
		}

		switch r {
		case '*':
			return "", fmt.Errorf("branch name contains invalid character: *")
		case '?':
			return "", fmt.Errorf("branch name contains invalid character: ?")
		case '[':
			return "", fmt.Errorf("branch name contains invalid character: [")
		case ']':
			return "", fmt.Errorf("branch name contains invalid character: ]")
		case '\\':
			return "", fmt.Errorf("branch name contains invalid character: \\")
		case '^':
			return "", fmt.Errorf("branch name contains invalid character: ^")
		case '~':
			return "", fmt.Errorf("branch name contains invalid character: ~")
		case ':':
			return "", fmt.Errorf("branch name contains invalid character")
		case ' ':
			return "", fmt.Errorf("branch name contains invalid character: space")
		}
	}

	if strings.Contains(sanitized, "..") {
		return "", fmt.Errorf("branch name contains invalid sequence")
	}

	if strings.Contains(sanitized, "@{") {
		return "", fmt.Errorf("branch name contains invalid sequence: @{")
	}

	if strings.Contains(sanitized, "//") {
		return "", fmt.Errorf("branch name contains invalid sequence: //")
	}

	if len(sanitized) > 255 {
		return "", fmt.Errorf("branch name too long: %d characters (max 255)", len(sanitized))
	}

	return sanitized, nil
}

// ValidateBranchName validates a branch name using git's own validation rules.
// Uses `git check-ref-format --branch` to ensure the name conforms to git's branch naming conventions.
// Returns an error if the branch name is invalid, git is not installed, or the operation times out.
// Respects the passed context while adding a 5-second timeout for the git operation.
func ValidateBranchName(ctx context.Context, branchName string) error {
	if strings.TrimSpace(branchName) == "" {
		return ErrEmptyBranchName
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "check-ref-format", "--branch", branchName)
	return cmd.Run()
}
