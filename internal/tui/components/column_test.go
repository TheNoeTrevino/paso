package components

import (
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
)

func TestRenderColumnHeader(t *testing.T) {
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
			result := renderColumnHeader(tt.column, tt.taskCount)
			if !strings.Contains(result, tt.wantText) {
				t.Errorf("renderColumnHeader() = %q, want to contain %q", result, tt.wantText)
			}
		})
	}
}

func TestRenderScrollIndicator_Show(t *testing.T) {
	result := renderScrollIndicator(true, "▲ more above")

	if !strings.Contains(result, "▲") {
		t.Errorf("renderScrollIndicator(true, ...) = %q, want to contain ▲", result)
	}

	if !strings.Contains(result, "more above") {
		t.Errorf("renderScrollIndicator(true, ...) = %q, want to contain 'more above'", result)
	}

	if !strings.HasSuffix(result, "\n") {
		t.Errorf("renderScrollIndicator(true, ...) should end with newline")
	}
}

func TestRenderScrollIndicator_Hide(t *testing.T) {
	result := renderScrollIndicator(false, "▲ more above")

	if result != "\n" {
		t.Errorf("renderScrollIndicator(false, ...) = %q, want single newline", result)
	}
}

func TestRenderEmptyColumnContent_Structure(t *testing.T) {
	header := "Test Header"

	result := renderEmptyColumnContent(header)

	// Should contain the header
	if !strings.Contains(result, header) {
		t.Errorf("renderEmptyColumnContent() should contain header %q", header)
	}

	// Should contain "No tasks" message
	if !strings.Contains(result, "No tasks") {
		t.Error("renderEmptyColumnContent() should contain 'No tasks' message")
	}

	// Should have proper spacing
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Errorf("renderEmptyColumnContent() should have at least 3 lines, got %d", len(lines))
	}
}

func TestRenderEmptyColumnContent_PaddingCalculation(t *testing.T) {
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
			result := renderEmptyColumnContent("Header")
			lines := strings.Split(result, "\n")

			// Should have content (not crash or return empty string)
			if len(lines) == 0 {
				t.Error("renderEmptyColumnContent() returned empty content")
			}

			// Should contain the expected elements
			content := strings.Join(lines, "\n")
			if !strings.Contains(content, "Header") {
				t.Error("Missing header in output")
			}
			if !strings.Contains(content, "No tasks") {
				t.Error("Missing 'No tasks' message in output")
			}
		})
	}
}

func TestApplyColumnStyle_Selection(t *testing.T) {
	content := "test content"

	// Test with selection
	selected := applyColumnStyle(content, true, 30)
	if selected == "" {
		t.Error("applyColumnStyle(true) returned empty string")
	}
	if !strings.Contains(selected, content) {
		t.Error("applyColumnStyle(true) should contain original content")
	}

	// Test without selection
	notSelected := applyColumnStyle(content, false, 30)
	if notSelected == "" {
		t.Error("applyColumnStyle(false) returned empty string")
	}
	if !strings.Contains(notSelected, content) {
		t.Error("applyColumnStyle(false) should contain original content")
	}
}

func TestApplyColumnStyle_Height(t *testing.T) {
	content := "test content"

	// Test with height
	withHeight := applyColumnStyle(content, false, 30)
	if withHeight == "" {
		t.Error("applyColumnStyle() with height returned empty string")
	}

	// Test with auto height (0)
	autoHeight := applyColumnStyle(content, false, 0)
	if autoHeight == "" {
		t.Error("applyColumnStyle() with auto height returned empty string")
	}
}

func TestRenderColumnWithTasksContent_VisibleTaskCalculation(t *testing.T) {
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

	result := renderColumnWithTasksContent(header, tasks, false, -1, height, scrollOffset)

	// Should contain header
	if !strings.Contains(result, header) {
		t.Error("renderColumnWithTasksContent() should contain header")
	}

	// Should have content
	if result == "" {
		t.Error("renderColumnWithTasksContent() returned empty string")
	}

	// Should have newlines (structure)
	if !strings.Contains(result, "\n") {
		t.Error("renderColumnWithTasksContent() should have line breaks")
	}
}

func TestRenderColumnWithTasksContent_ScrollIndicators(t *testing.T) {
	tasks := make([]*models.TaskSummary, 20)
	for i := range tasks {
		tasks[i] = &models.TaskSummary{
			ID:    i + 1,
			Title: "Test Task",
		}
	}

	header := "Test"
	height := 30

	// Test scrolled down (should show top indicator)
	scrolledDown := renderColumnWithTasksContent(header, tasks, false, -1, height, 5)
	if !strings.Contains(scrolledDown, "▲") {
		t.Error("Should show top indicator when scrolled down")
	}

	// Test at top (should not show top indicator in indicator line)
	atTop := renderColumnWithTasksContent(header, tasks, false, -1, height, 0)
	// The ▲ should not appear since we're at the top
	lines := strings.Split(atTop, "\n")
	hasTopIndicator := false
	for _, line := range lines[:5] { // Check first few lines
		if strings.Contains(line, "▲") {
			hasTopIndicator = true
			break
		}
	}
	if hasTopIndicator {
		t.Error("Should not show top indicator when at top of list")
	}
}
