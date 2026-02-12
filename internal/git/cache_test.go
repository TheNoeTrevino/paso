package git

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_Get_Miss(t *testing.T) {
	t.Parallel()

	cache := NewCache(DefaultCacheTTL)
	ctx := context.Background()

	gitInfo, branches, hit := cache.Get(ctx, "/nonexistent/path")

	assert.False(t, hit, "Get on empty cache should return false")
	assert.Equal(t, GitInfo{}, gitInfo, "Get miss should return empty GitInfo")
	assert.Nil(t, branches, "Get miss should return nil branches")
}

func TestCache_Get_Hit(t *testing.T) {
	t.Parallel()

	cache := NewCache(DefaultCacheTTL)
	ctx := context.Background()
	workDir := "/test/repo"

	expectedInfo := GitInfo{
		IsRepo:        true,
		CurrentBranch: "main",
		IsDetached:    false,
		HasCommits:    true,
	}
	expectedBranches := []BranchInfo{
		{Name: "main", Marker: "* "},
		{Name: "feature/test", Marker: "  "},
	}

	cache.Set(workDir, expectedInfo, expectedBranches)

	gitInfo, branches, hit := cache.Get(ctx, workDir)

	assert.True(t, hit, "Get should return true for cached entry")
	assert.Equal(t, expectedInfo, gitInfo, "Get should return cached GitInfo")
	assert.Equal(t, expectedBranches, branches, "Get should return cached branches")
}

func TestCache_Get_Expired(t *testing.T) {
	t.Parallel()

	ttl := 10 * time.Millisecond
	cache := NewCache(ttl)
	ctx := context.Background()
	workDir := "/test/repo"

	info := GitInfo{
		IsRepo:        true,
		CurrentBranch: "main",
		HasCommits:    true,
	}
	branches := []BranchInfo{{Name: "main", Marker: "* "}}

	cache.Set(workDir, info, branches)

	gitInfo, branchList, hit := cache.Get(ctx, workDir)
	assert.True(t, hit, "Should hit before TTL expires")
	assert.Equal(t, info, gitInfo)
	assert.Equal(t, branches, branchList)

	time.Sleep(15 * time.Millisecond)

	gitInfo, branchList, hit = cache.Get(ctx, workDir)
	assert.False(t, hit, "Should miss after TTL expires")
	assert.Equal(t, GitInfo{}, gitInfo, "Expired entry should return empty GitInfo")
	assert.Nil(t, branchList, "Expired entry should return nil branches")
}

func TestCache_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		workDir  string
		info     GitInfo
		branches []BranchInfo
	}{
		{
			name:    "basic_set",
			workDir: "/home/user/project",
			info: GitInfo{
				IsRepo:        true,
				CurrentBranch: "main",
				HasCommits:    true,
			},
			branches: []BranchInfo{
				{Name: "main", Marker: "* "},
			},
		},
		{
			name:    "set_with_multiple_branches",
			workDir: "/home/user/another-project",
			info: GitInfo{
				IsRepo:        true,
				CurrentBranch: "feature/auth",
				HasCommits:    true,
			},
			branches: []BranchInfo{
				{Name: "main", Marker: "  "},
				{Name: "develop", Marker: "  "},
				{Name: "feature/auth", Marker: "* "},
			},
		},
		{
			name:    "set_with_empty_branches",
			workDir: "/empty/repo",
			info: GitInfo{
				IsRepo:        true,
				CurrentBranch: "main",
				HasCommits:    false,
			},
			branches: nil,
		},
		{
			name:    "overwrite_existing",
			workDir: "/overwrite/test",
			info: GitInfo{
				IsRepo:        true,
				CurrentBranch: "updated-branch",
				HasCommits:    true,
			},
			branches: []BranchInfo{
				{Name: "updated-branch", Marker: "* "},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()

			cache := NewCache(DefaultCacheTTL)
			ctx := context.Background()

			if tt.name == "overwrite_existing" {
				oldInfo := GitInfo{IsRepo: true, CurrentBranch: "old-branch"}
				cache.Set(tt.workDir, oldInfo, []BranchInfo{{Name: "old-branch"}})
			}

			cache.Set(tt.workDir, tt.info, tt.branches)

			gotInfo, gotBranches, hit := cache.Get(ctx, tt.workDir)

			assert.True(t, hit, "Should hit after Set")
			assert.Equal(t, tt.info, gotInfo, "Stored GitInfo should match")
			assert.Equal(t, tt.branches, gotBranches, "Stored branches should match")
		})
	}
}

func TestCache_Invalidate(t *testing.T) {
	t.Parallel()

	cache := NewCache(DefaultCacheTTL)
	ctx := context.Background()

	workDir1 := "/project/one"
	workDir2 := "/project/two"

	info1 := GitInfo{IsRepo: true, CurrentBranch: "branch-one", HasCommits: true}
	info2 := GitInfo{IsRepo: true, CurrentBranch: "branch-two", HasCommits: true}
	branches1 := []BranchInfo{{Name: "branch-one", Marker: "* "}}
	branches2 := []BranchInfo{{Name: "branch-two", Marker: "* "}}

	cache.Set(workDir1, info1, branches1)
	cache.Set(workDir2, info2, branches2)

	_, _, hit1 := cache.Get(ctx, workDir1)
	_, _, hit2 := cache.Get(ctx, workDir2)
	assert.True(t, hit1, "workDir1 should be in cache")
	assert.True(t, hit2, "workDir2 should be in cache")

	cache.Invalidate(workDir1)

	_, _, hit1 = cache.Get(ctx, workDir1)
	gotInfo2, gotBranches2, hit2 := cache.Get(ctx, workDir2)

	assert.False(t, hit1, "workDir1 should be invalidated")
	assert.True(t, hit2, "workDir2 should still be in cache")
	assert.Equal(t, info2, gotInfo2, "workDir2 data should remain intact")
	assert.Equal(t, branches2, gotBranches2, "workDir2 branches should remain intact")
}

func TestCache_InvalidateAll(t *testing.T) {
	t.Parallel()

	cache := NewCache(DefaultCacheTTL)
	ctx := context.Background()

	workDirs := []string{
		"/project/one",
		"/project/two",
		"/project/three",
	}

	for i, wd := range workDirs {
		info := GitInfo{IsRepo: true, CurrentBranch: "branch", HasCommits: true}
		branches := []BranchInfo{{Name: "branch", Marker: "* "}}
		_ = i
		cache.Set(wd, info, branches)
	}

	for _, wd := range workDirs {
		_, _, hit := cache.Get(ctx, wd)
		assert.True(t, hit, "Entry should exist before InvalidateAll")
	}

	cache.InvalidateAll()

	for _, wd := range workDirs {
		_, _, hit := cache.Get(ctx, wd)
		assert.False(t, hit, "Entry should not exist after InvalidateAll")
	}
}

func TestCache_Concurrent(t *testing.T) {
	t.Parallel()

	cache := NewCache(DefaultCacheTTL)
	ctx := context.Background()
	workDir := "/concurrent/test"

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				switch j % 4 {
				case 0:
					info := GitInfo{
						IsRepo:        true,
						CurrentBranch: "branch",
						HasCommits:    true,
					}
					branches := []BranchInfo{{Name: "branch", Marker: "* "}}
					cache.Set(workDir, info, branches)
				case 1:
					cache.Get(ctx, workDir)
				case 2:
					cache.Invalidate(workDir)
				case 3:
					cache.InvalidateAll()
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestCache_ConcurrentDifferentKeys(t *testing.T) {
	t.Parallel()

	cache := NewCache(DefaultCacheTTL)
	ctx := context.Background()

	var wg sync.WaitGroup
	numKeys := 50

	for i := 0; i < numKeys; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			workDir := "/project/" + string(rune('a'+id%26)) + string(rune('0'+id/26))

			info := GitInfo{
				IsRepo:        true,
				CurrentBranch: "branch",
				HasCommits:    true,
			}
			branches := []BranchInfo{{Name: "branch"}}

			cache.Set(workDir, info, branches)

			gotInfo, gotBranches, hit := cache.Get(ctx, workDir)
			if hit {
				assert.Equal(t, info.CurrentBranch, gotInfo.CurrentBranch)
				assert.Equal(t, len(branches), len(gotBranches))
			}

			cache.Invalidate(workDir)
		}(i)
	}

	wg.Wait()
}
