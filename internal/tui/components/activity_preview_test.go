package components

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

func init() {
	InitStyles(*colors.Default())
}

func TestRenderActivityBadge_Comment(t *testing.T) {
	t.Parallel()

	result := RenderActivityBadge(models.ActivityTypeComment)
	assert.Contains(t, result, "Comment")
}

func TestRenderActivityBadge_Event(t *testing.T) {
	t.Parallel()

	result := RenderActivityBadge(models.ActivityTypeEvent)
	assert.Contains(t, result, "Event")
}

func TestRenderActivityPreviewItem_Comment(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 3, 15, 9, 30, 0, 0, time.UTC)
	item := models.ActivityItem{
		ID:        1,
		TaskID:    42,
		Type:      models.ActivityTypeComment,
		Content:   "Hello world",
		Author:    "alice",
		CreatedAt: created,
	}

	result := RenderActivityPreviewItem(item, 60)

	assert.Contains(t, result, "Comment")
	assert.Contains(t, result, "alice")
	assert.Contains(t, result, "Mar 15 09:30")
	assert.Contains(t, result, "Hello world")
}

func TestRenderActivityPreviewItem_LongContent(t *testing.T) {
	t.Parallel()

	// Build content that will definitely wrap beyond 2 lines at width=24.
	content := strings.Repeat("word ", 40)
	item := models.ActivityItem{
		Type:      models.ActivityTypeComment,
		Content:   content,
		Author:    "bob",
		CreatedAt: time.Date(2026, 3, 15, 9, 30, 0, 0, time.UTC),
	}

	result := RenderActivityPreviewItem(item, 24)

	assert.Contains(t, result, "...", "long content should be truncated with ellipsis")
}

func TestRenderActivityPreviews_FitsAll(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 3, 15, 9, 30, 0, 0, time.UTC)
	items := []models.ActivityItem{
		{Type: models.ActivityTypeComment, Content: "first", Author: "alice", CreatedAt: created},
		{Type: models.ActivityTypeComment, Content: "second", Author: "bob", CreatedAt: created},
		{Type: models.ActivityTypeEvent, Content: "third", Author: "carol", CreatedAt: created},
	}

	// 100 lines of available height is plenty for 3 items.
	result := RenderActivityPreviews(items, 60, 100)

	assert.Contains(t, result, "first")
	assert.Contains(t, result, "second")
	assert.Contains(t, result, "third")
}

func TestRenderActivityPreviews_TruncatesWhenTall(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 3, 15, 9, 30, 0, 0, time.UTC)
	items := []models.ActivityItem{
		{Type: models.ActivityTypeComment, Content: "first-item", Author: "alice", CreatedAt: created},
		{Type: models.ActivityTypeComment, Content: "second-item", Author: "bob", CreatedAt: created},
		{Type: models.ActivityTypeComment, Content: "third-item", Author: "carol", CreatedAt: created},
	}

	// availableHeight=4 => maxItems = 4/4 = 1, so only the first item should appear.
	result := RenderActivityPreviews(items, 60, 4)

	assert.Contains(t, result, "first-item")
	assert.NotContains(t, result, "second-item")
	assert.NotContains(t, result, "third-item")
}

func TestRenderActivityPreviews_MinimumHeight(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 3, 15, 9, 30, 0, 0, time.UTC)
	items := []models.ActivityItem{
		{Type: models.ActivityTypeComment, Content: "only-item", Author: "alice", CreatedAt: created},
		{Type: models.ActivityTypeComment, Content: "skipped", Author: "bob", CreatedAt: created},
	}

	// availableHeight=1 => maxItems = max(1/4, 1) = 1, guaranteed at least 1 item.
	result := RenderActivityPreviews(items, 60, 1)

	assert.Contains(t, result, "only-item")
	assert.NotContains(t, result, "skipped")
}
