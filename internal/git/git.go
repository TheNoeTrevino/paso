package git

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
	"unicode"
)

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

	// Check if we're in a git repository
	if !isGitRepo(ctx) {
		return info
	}
	info.IsRepo = true

	// Check if repository has commits
	info.HasCommits = hasCommits(ctx)

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

// isGitRepo checks if the current directory is inside a git repository
// Respects the passed context while adding a 5-second timeout for the git operation
func isGitRepo(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// hasCommits checks if the repository has at least one commit
// Respects the passed context while adding a 5-second timeout for the git operation
func hasCommits(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	err := cmd.Run()
	return err == nil
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
func getCurrentBranch(ctx context.Context) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

// ListBranches returns a list of all local git branches with their status markers.
// Branch names are sanitized for safe storage.
// Markers: "* " (current branch), "+ " (worktree branch), "  " (normal branch)
// Respects the passed context while adding a 5-second timeout for the git operation.
func ListBranches(ctx context.Context) ([]BranchInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "branch", "--list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	branches := make([]BranchInfo, 0, len(lines))

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		// Extract marker (first 2 characters: "* ", "+ ", or "  ")
		marker := line[:2]
		branchName := strings.TrimSpace(line[2:])

		if branchName == "" {
			continue
		}

		sanitized, err := SanitizeBranchName(branchName)
		if err != nil || sanitized == "" {
			continue
		}

		branches = append(branches, BranchInfo{
			Name:   sanitized,
			Marker: marker,
		})
	}

	// Sort by branch name
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Name < branches[j].Name
	})

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
			return "", fmt.Errorf("branch name contains invalid character: :")
		case ' ':
			return "", fmt.Errorf("branch name contains invalid character: space")
		}
	}

	if strings.Contains(sanitized, "..") {
		return "", fmt.Errorf("branch name contains invalid sequence: ..")
	}

	if strings.Contains(sanitized, "@{") {
		return "", fmt.Errorf("branch name contains invalid sequence: @{")
	}

	if strings.Contains(sanitized, "//") {
		return "", fmt.Errorf("branch name contains invalid sequence: //")
	}

	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
	}

	return sanitized, nil
}

// ValidateBranchName validates a branch name using git's own validation rules.
// Uses `git check-ref-format --branch` to ensure the name conforms to git's branch naming conventions.
// Returns an error if the branch name is invalid, git is not installed, or the operation times out.
// Respects the passed context while adding a 5-second timeout for the git operation.
func ValidateBranchName(ctx context.Context, branchName string) error {
	if strings.TrimSpace(branchName) == "" {
		return exec.Command("").Run()
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "check-ref-format", "--branch", branchName)
	return cmd.Run()
}
