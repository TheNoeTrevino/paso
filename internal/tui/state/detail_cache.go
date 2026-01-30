package state

import "github.com/thenoetrevino/paso/internal/models"

// DefaultCacheSize is the default maximum number of task details to cache.
const DefaultCacheSize = 25

// TaskDetailCache provides LRU caching for task details.
// It maintains a fixed-size cache that evicts the least recently used entries
// when the cache exceeds its maximum size.
type TaskDetailCache struct {
	cache   map[int]*models.TaskDetail
	order   []int // LRU order: oldest at front, most recent at end
	maxSize int
}

// NewTaskDetailCache creates a new TaskDetailCache with the specified maximum size.
// If maxSize is <= 0, it defaults to DefaultCacheSize.
func NewTaskDetailCache(maxSize int) *TaskDetailCache {
	if maxSize <= 0 {
		maxSize = DefaultCacheSize
	}
	return &TaskDetailCache{
		cache:   make(map[int]*models.TaskDetail),
		order:   make([]int, 0, maxSize),
		maxSize: maxSize,
	}
}

// Get retrieves a task detail from the cache and updates its LRU position.
// Returns the detail and true if found, nil and false otherwise.
func (c *TaskDetailCache) Get(taskID int) (*models.TaskDetail, bool) {
	detail, exists := c.cache[taskID]
	if !exists {
		return nil, false
	}
	c.moveToEnd(taskID)
	return detail, true
}

// Set adds or updates a task detail in the cache.
// If the cache is full, the least recently used entry is evicted.
func (c *TaskDetailCache) Set(taskID int, detail *models.TaskDetail) {
	if detail == nil {
		return
	}

	if _, exists := c.cache[taskID]; exists {
		c.cache[taskID] = detail
		c.moveToEnd(taskID)
		return
	}

	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[taskID] = detail
	c.order = append(c.order, taskID)
}

// SetBatch adds multiple task details to the cache efficiently.
// Entries are added in the order provided by iterating over the map.
func (c *TaskDetailCache) SetBatch(details map[int]*models.TaskDetail) {
	for taskID, detail := range details {
		c.Set(taskID, detail)
	}
}

// Invalidate removes a single entry from the cache.
func (c *TaskDetailCache) Invalidate(taskID int) {
	if _, exists := c.cache[taskID]; !exists {
		return
	}
	delete(c.cache, taskID)
	c.removeFromOrder(taskID)
}

// Clear removes all entries from the cache.
func (c *TaskDetailCache) Clear() {
	c.cache = make(map[int]*models.TaskDetail)
	c.order = make([]int, 0, c.maxSize)
}

// Has checks if a task ID exists in the cache without updating LRU order.
func (c *TaskDetailCache) Has(taskID int) bool {
	_, exists := c.cache[taskID]
	return exists
}

// CachedIDs returns all task IDs currently in the cache.
// The returned slice is a copy and can be safely modified.
func (c *TaskDetailCache) CachedIDs() []int {
	ids := make([]int, len(c.order))
	copy(ids, c.order)
	return ids
}

// moveToEnd moves a task ID to the end of the order slice (most recently used).
func (c *TaskDetailCache) moveToEnd(taskID int) {
	c.removeFromOrder(taskID)
	c.order = append(c.order, taskID)
}

// removeFromOrder removes a task ID from the order slice.
func (c *TaskDetailCache) removeFromOrder(taskID int) {
	for i, id := range c.order {
		if id == taskID {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}

// evictOldest removes the least recently used entry from the cache.
func (c *TaskDetailCache) evictOldest() {
	if len(c.order) == 0 {
		return
	}
	oldest := c.order[0]
	c.order = c.order[1:]
	delete(c.cache, oldest)
}
