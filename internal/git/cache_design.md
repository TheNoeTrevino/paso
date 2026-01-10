# Git Information Cache Design

## Overview
Design for caching git repository information to improve TUI responsiveness and reduce redundant git command execution.

## Problem Statement
Git operations (DetectGitInfo, ListBranches) are currently executed synchronously on every project form open and branch picker interaction. In repositories with many branches or slow git operations, this causes UI freezes and poor user experience.

## Goals
1. Reduce UI blocking by caching git information
2. Support multi-repository workflows (different working directories)
3. Provide manual refresh capability (Ctrl+R)
4. Integrate seamlessly with existing async git command pattern (task #23)
5. Maintain data freshness while minimizing redundant git calls

## Cache Architecture

### Location: Separate Cache Package
**Decision**: Create `internal/git/cache.go` in the git package

**Rationale**:
- Keeps git-related logic together
- Cache is tightly coupled to git operations
- Can be unit tested independently
- Easier to mock for testing
- Follows single responsibility principle

**Alternative Considered**: Embed in TUI Model state
- Rejected because: Tight coupling to TUI, harder to test, violates separation of concerns

### Cache Structure

```go
package git

import (
    "context"
    "sync"
    "time"
)

// CacheEntry holds cached git information with timestamp
type CacheEntry struct {
    GitInfo   GitInfo
    Branches  []BranchInfo
    Timestamp time.Time
}

// Cache provides thread-safe caching of git repository information
// keyed by working directory to support multi-repo workflows
type Cache struct {
    mu      sync.RWMutex
    entries map[string]*CacheEntry // key: absolute working directory path
    ttl     time.Duration
}

// NewCache creates a cache with specified TTL
func NewCache(ttl time.Duration) *Cache {
    return &Cache{
        entries: make(map[string]*CacheEntry),
        ttl:     ttl,
    }
}
```

### Cache Key
**Key**: Absolute path to current working directory

**Rationale**:
- Supports multi-repository workflows (user switches between different repos)
- Git operations are directory-specific
- Natural boundary for cache invalidation
- Can use `os.Getwd()` to get current directory

**Example Keys**:
- `/home/user/projects/paso`
- `/home/user/projects/myapp`

### Cache Operations

```go
// Get retrieves cached git info and branches for current directory
// Returns (GitInfo, []BranchInfo, bool) where bool indicates cache hit
func (c *Cache) Get(ctx context.Context, workDir string) (GitInfo, []BranchInfo, bool)

// Set stores git info and branches for current directory
func (c *Cache) Set(workDir string, info GitInfo, branches []BranchInfo)

// Invalidate removes cache entry for current directory
func (c *Cache) Invalidate(workDir string)

// InvalidateAll clears entire cache (for global refresh)
func (c *Cache) InvalidateAll()

// isExpired checks if cache entry has exceeded TTL
func (c *Cache) isExpired(entry *CacheEntry) bool
```

### TTL (Time-To-Live)

**Recommended Value**: 45 seconds

**Rationale**:
- Git repository state doesn't change frequently during active TUI sessions
- Long enough to prevent redundant calls during rapid UI interactions
- Short enough to detect external changes (other terminal windows, IDE changes)
- Balances between freshness and performance
- User can manually refresh anytime with Ctrl+R

**Configurable**: TTL should be configurable via const in cache.go for easy tuning

```go
const (
    DefaultCacheTTL = 45 * time.Second
)
```

### Thread Safety

**Approach**: sync.RWMutex

**Rationale**:
- Git operations may be called from multiple goroutines (async tea.Cmd)
- RWMutex allows concurrent reads, exclusive writes
- Prevents race conditions
- Standard Go pattern for this use case

**Usage Pattern**:
- Read lock for Get() - allows concurrent cache hits
- Write lock for Set(), Invalidate() - exclusive access for modifications

## Integration with Async Flow (Task #23)

### Current Async Pattern
From `internal/tui/git_commands.go`:
```go
func detectGitInfoCmd(ctx context.Context) tea.Cmd {
    return func() tea.Msg {
        info := git.DetectGitInfo(ctx)
        return gitInfoMsg{info: &info, err: nil}
    }
}

func listBranchesCmd(ctx context.Context) tea.Cmd {
    return func() tea.Msg {
        branches, err := git.ListBranches(ctx)
        return gitBranchesMsg{branches: branches, err: err}
    }
}
```

### Cache Integration Pattern

**High-Level Flow**:
1. User triggers action requiring git info (e.g., open project form)
2. TUI dispatches async command that checks cache first
3. On cache hit: Return cached data immediately (fast path)
4. On cache miss or expired: Execute git commands, populate cache, return data (slow path)

**Modified Commands**:
```go
// Pass cache instance to commands
func detectGitInfoCmd(ctx context.Context, cache *git.Cache, workDir string) tea.Cmd {
    return func() tea.Msg {
        // Fast path: check cache
        if info, _, hit := cache.Get(ctx, workDir); hit {
            return gitInfoMsg{info: &info, err: nil}
        }

        // Slow path: fetch from git
        info := git.DetectGitInfo(ctx)
        cache.Set(workDir, info, nil) // Update cache (branches set separately)
        return gitInfoMsg{info: &info, err: nil}
    }
}

func listBranchesCmd(ctx context.Context, cache *git.Cache, workDir string) tea.Cmd {
    return func() tea.Msg {
        // Fast path: check cache
        if _, branches, hit := cache.Get(ctx, workDir); hit && branches != nil {
            return gitBranchesMsg{branches: branches, err: nil}
        }

        // Slow path: fetch from git
        branches, err := git.ListBranches(ctx)
        if err == nil {
            // Update cache (info should already be cached or will be cached separately)
            info, _, _ := cache.Get(ctx, workDir)
            cache.Set(workDir, info, branches)
        }
        return gitBranchesMsg{branches: branches, err: err}
    }
}
```

### Cache Instance Lifecycle

**Storage**: Add to Model struct
```go
type Model struct {
    // ... existing fields
    GitCache *git.Cache  // Git information cache
}
```

**Initialization**: In `InitialModel()`
```go
func InitialModel(ctx context.Context, application *app.App, ...) Model {
    // ...
    gitCache := git.NewCache(git.DefaultCacheTTL)

    return Model{
        // ... existing fields
        GitCache: gitCache,
    }
}
```

## Invalidation Strategy

### Manual Invalidation (Ctrl+R)

**User Action**: User presses Ctrl+R in project form or branch picker

**Behavior**:
1. Call `cache.Invalidate(currentWorkDir)` or `cache.InvalidateAll()`
2. Re-dispatch async git commands to fetch fresh data
3. Update cache with fresh data
4. Re-render UI with new data

**Location**: Handle in `update_normal.go` or relevant form update handlers

### Automatic Invalidation (TTL Expiry)

**Behavior**:
- Checked on every `cache.Get()` call
- If `time.Since(entry.Timestamp) > cache.ttl`, treat as cache miss
- Triggers fresh git fetch
- Updates cache with new timestamp

**No background expiry**: Keep it simple - check on access, not on timer

### Scope of Invalidation

**Per-Directory**: `Invalidate(workDir)` - Invalidates single repository
**Global**: `InvalidateAll()` - Clears entire cache

**Use Cases**:
- Ctrl+R in specific form: per-directory invalidation
- Application-wide refresh command: global invalidation
- Switching working directories: automatic (different cache key)

## Cache Miss Handling

### Flow
1. Check cache with `Get()`
2. If miss or expired, execute git commands
3. Store results with `Set()`
4. Return data to caller

### Error Handling
- If git command fails, **do not cache error state**
- Return error to caller without updating cache
- Next call will retry git operation
- Prevents caching transient failures

## Data Consistency

### Race Conditions
**Scenario**: Concurrent requests for same repository

**Mitigation**:
- RWMutex ensures thread-safe access
- Last writer wins for cache updates
- Acceptable because git state should be consistent within TTL window

### Stale Data
**Scenario**: External git operations (other terminal, IDE) modify repository

**Mitigation**:
- TTL ensures cache refreshes periodically (45s default)
- User can manually refresh with Ctrl+R
- Trade-off: slight staleness for better performance

## Future Enhancements (Out of Scope for Task #27)

1. **File System Watching**: Use fsnotify to invalidate cache on .git/ changes
2. **Adaptive TTL**: Shorter TTL for active repos, longer for inactive
3. **Partial Cache Updates**: Update only GitInfo or Branches separately
4. **Cache Persistence**: Save cache to disk for faster startup (probably overkill)
5. **Cache Metrics**: Track hit/miss rates for optimization

## Testing Strategy

### Unit Tests (cache.go)
- Test cache hit/miss scenarios
- Test TTL expiration logic
- Test thread safety (concurrent Get/Set)
- Test invalidation (per-directory, global)
- Test working directory isolation

### Integration Tests
- Test cache integration with async commands
- Test manual refresh (Ctrl+R)
- Test multi-repository scenarios
- Test error handling (git failures)

## Implementation Checklist (Task #28)

1. Create `internal/git/cache.go` with Cache struct
2. Implement cache operations (Get, Set, Invalidate, InvalidateAll)
3. Add cache instance to Model struct
4. Initialize cache in InitialModel()
5. Modify detectGitInfoCmd to use cache
6. Modify listBranchesCmd to use cache
7. Add manual refresh handler (Ctrl+R)
8. Write unit tests for cache
9. Write integration tests for cache + async flow
10. Update documentation (if needed)

## Summary

**Cache Location**: `internal/git/cache.go` (separate package)
**Cache Key**: Absolute working directory path
**TTL**: 45 seconds (configurable)
**Thread Safety**: sync.RWMutex
**Invalidation**: Manual (Ctrl+R) + Automatic (TTL expiry)
**Integration**: Seamless with existing async tea.Cmd pattern

This design provides a clean, testable, and maintainable caching solution that integrates naturally with the existing Bubble Tea architecture and async git operation pattern.
