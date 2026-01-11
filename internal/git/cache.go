package git

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	DefaultCacheTTL = 45 * time.Second
)

type CacheEntry struct {
	GitInfo   GitInfo
	Branches  []BranchInfo
	Timestamp time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

func (c *Cache) Get(ctx context.Context, workDir string) (GitInfo, []BranchInfo, bool) {
	c.mu.RLock()

	entry, exists := c.entries[workDir]
	if !exists {
		c.mu.RUnlock()
		slog.Debug("cache miss: entry not found", "workDir", workDir)
		return GitInfo{}, nil, false
	}

	if c.isExpired(entry) {
		slog.Debug("cache miss: entry expired",
			"workDir", workDir,
			"age", time.Since(entry.Timestamp),
			"ttl", c.ttl)

		// Remove expired entry
		c.mu.RUnlock()
		c.mu.Lock()
		// Double-check entry still exists and is still expired
		if entry, exists := c.entries[workDir]; exists && c.isExpired(entry) {
			delete(c.entries, workDir)
			slog.Debug("removed expired cache entry", "workDir", workDir)
		}
		c.mu.Unlock()

		return GitInfo{}, nil, false
	}

	slog.Debug("cache hit",
		"workDir", workDir,
		"age", time.Since(entry.Timestamp),
		"branch", entry.GitInfo.CurrentBranch,
		"branchCount", len(entry.Branches))

	c.mu.RUnlock()
	return entry.GitInfo, entry.Branches, true
}

func (c *Cache) Set(workDir string, info GitInfo, branches []BranchInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[workDir] = &CacheEntry{
		GitInfo:   info,
		Branches:  branches,
		Timestamp: time.Now(),
	}

	slog.Debug("cache updated",
		"workDir", workDir,
		"branch", info.CurrentBranch,
		"branchCount", len(branches))
}

func (c *Cache) Invalidate(workDir string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, workDir)
	slog.Debug("cache invalidated", "workDir", workDir)
}

func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	slog.Debug("cache invalidated (all entries)")
}

func (c *Cache) isExpired(entry *CacheEntry) bool {
	return time.Since(entry.Timestamp) > c.ttl
}
