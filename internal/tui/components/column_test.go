package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestRenderColumnHeader(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		column    *models.Column
		taskCount int
		wantText  string
	}{
		{
			name:      "empty column",
			column:    &models.Column{Name: "Backlog"},
			taskCount: 0,
			wantText:  "Backlog (0)",
		},
		{
			name:      "single task",
			column:    &models.Column{Name: "In Progress"},
			taskCount: 1,
			wantText:  "In Progress (1)",
		},
		{
			name:      "multiple tasks",
			column:    &models.Column{Name: "Done"},
			taskCount: 42,
			wantText:  "Done (42)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			result := renderColumnHeader(tt.column, tt.taskCount)
			assert.Contains(t, result, tt.wantText)
		})
	}
}

func TestRenderScrollIndicator_Show(t *testing.T) {
	t.Parallel()
	result := renderScrollIndicator(true, "▲ more above")

	assert.Contains(t, result, "▲")
	assert.Contains(t, result, "more above")
	assert.True(t, strings.HasSuffix(result, "\n"))
}

func TestRenderScrollIndicator_Hide(t *testing.T) {
	t.Parallel()
	result := renderScrollIndicator(false, "▲ more above")

	assert.Equal(t, "\n", result)
}

func TestRenderEmptyColumnContent_Structure(t *testing.T) {
	t.Parallel()
	header := "Test Header"

	result := renderEmptyColumnContent(header)

	assert.Contains(t, result, header)
	assert.Contains(t, result, "No tasks")

	lines := strings.Split(result, "\n")
	require.GreaterOrEqual(t, len(lines), 3)
}

func TestRenderEmptyColumnContent_PaddingCalculation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		height int
	}{
		{"small height", 10},
		{"medium height", 30},
		{"large height", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			result := renderEmptyColumnContent("Header")
			lines := strings.Split(result, "\n")

			require.NotEmpty(t, lines)

			content := strings.Join(lines, "\n")
			assert.Contains(t, content, "Header")
			assert.Contains(t, content, "No tasks")
		})
	}
}

func TestApplyColumnStyle_Selection(t *testing.T) {
	t.Parallel()
	content := "test content"
	width := 40

	// Test with selection
	selected := applyColumnStyle(content, true, 30, width)
	assert.NotEmpty(t, selected)
	assert.Contains(t, selected, content)

	// Test without selection
	notSelected := applyColumnStyle(content, false, 30, width)
	assert.NotEmpty(t, notSelected)
	assert.Contains(t, notSelected, content)
}

func TestApplyColumnStyle_Height(t *testing.T) {
	t.Parallel()
	content := "test content"
	width := 40

	// Test with height
	withHeight := applyColumnStyle(content, false, 30, width)
	assert.NotEmpty(t, withHeight)

	// Test with auto height (0)
	autoHeight := applyColumnStyle(content, false, 0, width)
	assert.NotEmpty(t, autoHeight)
}

func TestRenderColumnWithTasksContent_VisibleTaskCalculation(t *testing.T) {
	t.Parallel()
	// Create test tasks
	tasks := make([]*models.TaskSummary, 10)
	for i := range tasks {
		tasks[i] = &models.TaskSummary{
			ID:    i + 1,
			Title: "Test Task",
		}
	}

	header := "Test Header"
	height := 30
	scrollOffset := 0
	taskCardWidth := 32 // typical task card width

	result := renderColumnWithTasksContent(header, tasks, false, -1, height, scrollOffset, taskCardWidth)

	assert.Contains(t, result, header)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "\n")
}

func TestRenderColumnWithTasksContent_ScrollIndicators(t *testing.T) {
	t.Parallel()
	tasks := make([]*models.TaskSummary, 20)
	for i := range tasks {
		tasks[i] = &models.TaskSummary{
			ID:    i + 1,
			Title: "Test Task",
		}
	}

	header := "Test"
	height := 30
	taskCardWidth := 32 // typical task card width

	// Test scrolled down (should show top indicator)
	scrolledDown := renderColumnWithTasksContent(header, tasks, false, -1, height, 5, taskCardWidth)
	assert.Contains(t, scrolledDown, "▲")

	// Test at top (should not show top indicator in indicator line)
	atTop := renderColumnWithTasksContent(header, tasks, false, -1, height, 0, taskCardWidth)
	// The upward arrow should not appear since we're at the top
	lines := strings.Split(atTop, "\n")
	hasTopIndicator := false
	for _, line := range lines[:5] { // Check first few lines
		if strings.Contains(line, "▲") {
			hasTopIndicator = true
			break
		}
	}
	assert.False(t, hasTopIndicator)
}
