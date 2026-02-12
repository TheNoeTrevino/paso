package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
)

// helper function to create a test TaskDetail with a given ID
func makeTestDetail(id int) *models.TaskDetail {
	return &models.TaskDetail{
		ID:    id,
		Title: "Test Task",
	}
}

// TestNewTaskDetailCache_DefaultSize ensures default cache size is used for invalid maxSize.
// Edge case: User provides zero or negative maxSize.
func TestNewTaskDetailCache_DefaultSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		maxSize  int
		wantSize int
	}{
		{"zero maxSize", 0, DefaultCacheSize},
		{"negative maxSize", -5, DefaultCacheSize},
		{"positive maxSize", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewTaskDetailCache(tt.maxSize)
			assert.Equal(t, tt.wantSize, cache.maxSize)
		})
	}
}

// TestSet_LRUEvictionOrder verifies oldest entries are evicted first when cache exceeds maxSize.
// This tests Task 221: Fill cache to max, add one more, verify FIRST entry was evicted.
func TestSet_LRUEvictionOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		maxSize        int
		insertIDs      []int
		wantEvictedIDs []int // IDs that should NOT be in cache after inserts
		wantCachedIDs  []int // IDs that should be in cache after inserts
	}{
		{
			name:           "evict first when full",
			maxSize:        3,
			insertIDs:      []int{1, 2, 3, 4}, // 4th insert should evict ID 1
			wantEvictedIDs: []int{1},
			wantCachedIDs:  []int{2, 3, 4},
		},
		{
			name:           "evict multiple oldest",
			maxSize:        3,
			insertIDs:      []int{1, 2, 3, 4, 5}, // Should evict 1, then 2
			wantEvictedIDs: []int{1, 2},
			wantCachedIDs:  []int{3, 4, 5},
		},
		{
			name:           "single capacity cache",
			maxSize:        1,
			insertIDs:      []int{1, 2, 3},
			wantEvictedIDs: []int{1, 2},
			wantCachedIDs:  []int{3},
		},
		{
			name:           "no eviction when under capacity",
			maxSize:        5,
			insertIDs:      []int{1, 2, 3},
			wantEvictedIDs: []int{},
			wantCachedIDs:  []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewTaskDetailCache(tt.maxSize)

			for _, id := range tt.insertIDs {
				cache.Set(id, makeTestDetail(id))
			}

			// Verify evicted IDs are NOT in cache
			for _, id := range tt.wantEvictedIDs {
				assert.False(t, cache.Has(id), "ID %d should have been evicted but is still in cache", id)
			}

			// Verify expected IDs ARE in cache
			for _, id := range tt.wantCachedIDs {
				assert.True(t, cache.Has(id), "ID %d should be in cache but was not found", id)
			}

			// Verify cache size doesn't exceed maxSize
			assert.LessOrEqual(t, len(cache.CachedIDs()), tt.maxSize)
		})
	}
}

// TestGet_UpdatesLRUPosition verifies that accessing a cached item moves it to most recently used.
// This tests Task 222: Add A, B, C; access A; add D; verify B was evicted (not A).
func TestGet_UpdatesLRUPosition(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	// Add entries A(1), B(2), C(3) in that order
	cache.Set(1, makeTestDetail(1)) // A - oldest
	cache.Set(2, makeTestDetail(2)) // B
	cache.Set(3, makeTestDetail(3)) // C - newest

	// Access A (ID 1), which should move it to most recently used
	detail, found := cache.Get(1)
	require.True(t, found, "Get(1) should find entry A")
	assert.Equal(t, 1, detail.ID)

	// Now order should be: B(2), C(3), A(1) - with B as oldest
	// Add D (ID 4), which should evict B (the oldest untouched)
	cache.Set(4, makeTestDetail(4))

	// Verify B was evicted (not A)
	assert.False(t, cache.Has(2), "B (ID 2) should have been evicted but is still in cache")

	// Verify A is still in cache (it was accessed recently)
	assert.True(t, cache.Has(1), "A (ID 1) should still be in cache after Get() updated its LRU position")

	// Verify C and D are in cache
	assert.True(t, cache.Has(3), "C (ID 3) should be in cache")
	assert.True(t, cache.Has(4), "D (ID 4) should be in cache")
}

// TestGet_UpdatesLRUPosition_MultipleAccesses tests multiple Get() calls affecting eviction order.
func TestGet_UpdatesLRUPosition_MultipleAccesses(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	// Add A, B, C
	cache.Set(1, makeTestDetail(1))
	cache.Set(2, makeTestDetail(2))
	cache.Set(3, makeTestDetail(3))

	// Access B, then A (order becomes: C, B, A)
	cache.Get(2)
	cache.Get(1)

	// Add D and E (should evict C first, then B)
	cache.Set(4, makeTestDetail(4))
	cache.Set(5, makeTestDetail(5))

	// C and B should be evicted
	assert.False(t, cache.Has(3), "C (ID 3) should have been evicted first")
	assert.False(t, cache.Has(2), "B (ID 2) should have been evicted second")

	// A, D, E should remain
	assert.True(t, cache.Has(1), "A (ID 1) should still be in cache")
	assert.True(t, cache.Has(4), "D (ID 4) should be in cache")
	assert.True(t, cache.Has(5), "E (ID 5) should be in cache")
}

// TestSet_NilDetail ensures setting nil detail is a no-op.
// Edge case: Caller passes nil detail, should not add entry or cause panic.
func TestSet_NilDetail(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	// Set a valid entry first
	cache.Set(1, makeTestDetail(1))

	// Try to set nil detail
	cache.Set(2, nil)

	// Verify nil was not added
	assert.False(t, cache.Has(2), "Set(2, nil) should not add entry to cache")

	// Verify original entry is still there
	assert.True(t, cache.Has(1), "Original entry (ID 1) should still be in cache")

	// Verify cache size
	ids := cache.CachedIDs()
	assert.Len(t, ids, 1)
}

// TestSet_UpdateExisting ensures updating existing entry refreshes LRU position.
func TestSet_UpdateExisting(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	// Add A, B, C
	cache.Set(1, makeTestDetail(1))
	cache.Set(2, makeTestDetail(2))
	cache.Set(3, makeTestDetail(3))

	// Update A with new detail (should move to most recently used)
	updatedDetail := &models.TaskDetail{
		ID:    1,
		Title: "Updated Task",
	}
	cache.Set(1, updatedDetail)

	// Add D (should evict B, not A since A was just updated)
	cache.Set(4, makeTestDetail(4))

	// Verify B was evicted
	assert.False(t, cache.Has(2), "B (ID 2) should have been evicted")

	// Verify A has updated value
	detail, found := cache.Get(1)
	require.True(t, found, "A (ID 1) should still be in cache")
	assert.Equal(t, "Updated Task", detail.Title)
}

// TestInvalidate_NonExistentID ensures Invalidate on non-existent ID is safe (no panic).
// Edge case: Caller tries to invalidate an ID that was never cached or already evicted.
func TestInvalidate_NonExistentID(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	// Add some entries
	cache.Set(1, makeTestDetail(1))
	cache.Set(2, makeTestDetail(2))

	// Invalidate a non-existent ID (should not panic)
	cache.Invalidate(999)

	// Verify existing entries are unaffected
	assert.True(t, cache.Has(1), "ID 1 should still be in cache after invalidating non-existent ID")
	assert.True(t, cache.Has(2), "ID 2 should still be in cache after invalidating non-existent ID")
}

// TestInvalidate_EmptyCache ensures Invalidate on empty cache is safe.
func TestInvalidate_EmptyCache(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	// Invalidate on empty cache (should not panic)
	cache.Invalidate(1)

	// Verify cache is still empty
	ids := cache.CachedIDs()
	assert.Len(t, ids, 0)
}

// TestCachedIDs_EmptyCache ensures CachedIDs returns empty slice on empty cache.
func TestCachedIDs_EmptyCache(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	ids := cache.CachedIDs()

	require.NotNil(t, ids, "CachedIDs() should return non-nil empty slice, not nil")
	assert.Len(t, ids, 0)
}

// TestCachedIDs_ReturnsCopy ensures returned slice is a copy that can be safely modified.
func TestCachedIDs_ReturnsCopy(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)
	cache.Set(1, makeTestDetail(1))
	cache.Set(2, makeTestDetail(2))

	ids := cache.CachedIDs()

	// Modify the returned slice
	ids[0] = 999

	// Verify internal state is unaffected
	internalIDs := cache.CachedIDs()
	assert.NotEqual(t, 999, internalIDs[0], "CachedIDs() should return a copy, not the internal slice")
}

// TestSetBatch_EmptyMap ensures SetBatch with empty map is safe.
func TestSetBatch_EmptyMap(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	// Add an entry first
	cache.Set(1, makeTestDetail(1))

	// SetBatch with empty map (should be no-op)
	cache.SetBatch(map[int]*models.TaskDetail{})

	// Verify existing entry is unaffected
	assert.True(t, cache.Has(1), "Existing entry should be unaffected by empty SetBatch")

	ids := cache.CachedIDs()
	assert.Len(t, ids, 1)
}

// TestSetBatch_NilMap ensures SetBatch with nil map is safe.
func TestSetBatch_NilMap(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)
	cache.Set(1, makeTestDetail(1))

	// SetBatch with nil map (should be no-op, range over nil map is safe in Go)
	cache.SetBatch(nil)

	assert.True(t, cache.Has(1), "Existing entry should be unaffected by nil SetBatch")
}

// TestGet_NonExistentID ensures Get on non-existent ID returns nil, false.
func TestGet_NonExistentID(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)
	cache.Set(1, makeTestDetail(1))

	detail, found := cache.Get(999)

	assert.False(t, found, "Get(999) should return false for non-existent ID")
	assert.Nil(t, detail, "Get(999) should return nil detail for non-existent ID")
}

// TestGet_EmptyCache ensures Get on empty cache returns nil, false.
func TestGet_EmptyCache(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	detail, found := cache.Get(1)

	assert.False(t, found, "Get(1) on empty cache should return false")
	assert.Nil(t, detail, "Get(1) on empty cache should return nil")
}

// TestHas_EmptyCache ensures Has on empty cache returns false.
func TestHas_EmptyCache(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	assert.False(t, cache.Has(1), "Has(1) on empty cache should return false")
}

// TestClear_EmptyCache ensures Clear on empty cache is safe.
func TestClear_EmptyCache(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	// Clear empty cache (should not panic)
	cache.Clear()

	ids := cache.CachedIDs()
	assert.Len(t, ids, 0)
}

// TestClear_WithEntries ensures Clear removes all entries.
func TestClear_WithEntries(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)
	cache.Set(1, makeTestDetail(1))
	cache.Set(2, makeTestDetail(2))
	cache.Set(3, makeTestDetail(3))

	cache.Clear()

	assert.False(t, cache.Has(1) || cache.Has(2) || cache.Has(3), "Clear() should remove all entries")

	ids := cache.CachedIDs()
	assert.Len(t, ids, 0)
}

// TestInvalidate_RemovesFromOrder ensures Invalidate properly removes from LRU order.
func TestInvalidate_RemovesFromOrder(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(3)

	// Add A, B, C
	cache.Set(1, makeTestDetail(1))
	cache.Set(2, makeTestDetail(2))
	cache.Set(3, makeTestDetail(3))

	// Invalidate B (middle entry)
	cache.Invalidate(2)

	// Verify B is removed
	assert.False(t, cache.Has(2), "B (ID 2) should be removed after Invalidate")

	// Add D and E (should evict A first, then C - not cause issues due to removed B)
	cache.Set(4, makeTestDetail(4))
	cache.Set(5, makeTestDetail(5))
	cache.Set(6, makeTestDetail(6))

	// A and C should be evicted (B was already removed)
	assert.False(t, cache.Has(1), "A (ID 1) should have been evicted")
	assert.False(t, cache.Has(3), "C (ID 3) should have been evicted")

	// D, E, F should remain
	assert.True(t, cache.Has(4) && cache.Has(5) && cache.Has(6), "D, E, F should all be in cache")
}

// TestCachedIDs_PreservesLRUOrder ensures CachedIDs returns IDs in LRU order (oldest first).
func TestCachedIDs_PreservesLRUOrder(t *testing.T) {
	t.Parallel()
	cache := NewTaskDetailCache(5)

	// Add in order: 1, 2, 3
	cache.Set(1, makeTestDetail(1))
	cache.Set(2, makeTestDetail(2))
	cache.Set(3, makeTestDetail(3))

	// Access 1 (moves to end)
	cache.Get(1)

	// Order should now be: 2, 3, 1
	ids := cache.CachedIDs()

	expected := []int{2, 3, 1}
	require.Len(t, ids, len(expected))

	for i, id := range ids {
		assert.Equal(t, expected[i], id, "CachedIDs()[%d]", i)
	}
}
