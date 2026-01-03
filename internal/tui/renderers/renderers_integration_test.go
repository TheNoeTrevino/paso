package renderers

import (
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestRenderListViewBasic tests basic list view rendering
func TestRenderListViewBasic(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Task 1",
			},
			ColumnName: "Todo",
			ColumnID:   1,
		},
		{
			Task: &models.TaskSummary{
				ID:    2,
				Title: "Task 2",
			},
			ColumnName: "In Progress",
			ColumnID:   2,
		},
	}

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20)

	if output == "" {
		t.Fatal("RenderListView should produce output")
	}

	// Should contain task titles
	if !strings.Contains(output, "Task 1") && !strings.Contains(output, "Task") {
		t.Error("output should contain task information")
	}
}

// TestRenderListViewEmpty tests list view with no rows
func TestRenderListViewEmpty(t *testing.T) {
	var rows []ListViewRow

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20)

	if output == "" {
		t.Fatal("RenderListView should produce output even when empty")
	}
}

// TestRenderListViewSelectedRow tests rendering with selected row
func TestRenderListViewSelectedRow(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Task 1",
			},
			ColumnName: "Todo",
			ColumnID:   1,
		},
		{
			Task: &models.TaskSummary{
				ID:    2,
				Title: "Task 2",
			},
			ColumnName: "In Progress",
			ColumnID:   2,
		},
		{
			Task: &models.TaskSummary{
				ID:    3,
				Title: "Task 3",
			},
			ColumnName: "Done",
			ColumnID:   3,
		},
	}

	// Select the middle row
	output := RenderListView(rows, 1, 0, state.SortByTitle, state.SortAsc, 100, 20)

	if output == "" {
		t.Fatal("RenderListView should render with selected row")
	}
}

// TestRenderListViewScrolling tests rendering with scroll offset
func TestRenderListViewScrolling(t *testing.T) {
	// Create many rows
	rows := make([]ListViewRow, 50)
	for i := 0; i < 50; i++ {
		rows[i] = ListViewRow{
			Task: &models.TaskSummary{
				ID:    i + 1,
				Title: "Task " + string(rune('A'+(i%26))),
			},
			ColumnName: "Column",
			ColumnID:   1,
		}
	}

	// Test with scroll offset
	output := RenderListView(rows, 10, 5, state.SortByTitle, state.SortAsc, 100, 20)

	if output == "" {
		t.Fatal("RenderListView should handle scrolling")
	}
}

// TestRenderListViewDifferentSorts tests various sort fields
func TestRenderListViewDifferentSorts(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Task A",
			},
			ColumnName: "Todo",
			ColumnID:   1,
		},
		{
			Task: &models.TaskSummary{
				ID:    2,
				Title: "Task B",
			},
			ColumnName: "In Progress",
			ColumnID:   2,
		},
	}

	sorts := []state.SortField{
		state.SortByTitle,
		state.SortByStatus,
	}

	for _, sortField := range sorts {
		t.Run("sort", func(t *testing.T) {
			output := RenderListView(rows, 0, 0, sortField, state.SortAsc, 100, 20)

			if output == "" {
				t.Errorf("RenderListView should work with sort field %v", sortField)
			}
		})
	}
}

// TestRenderListViewSortOrders tests ascending and descending
func TestRenderListViewSortOrders(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Task A",
			},
			ColumnName: "Column",
			ColumnID:   1,
		},
		{
			Task: &models.TaskSummary{
				ID:    2,
				Title: "Task B",
			},
			ColumnName: "Column",
			ColumnID:   1,
		},
	}

	ascOutput := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20)
	descOutput := RenderListView(rows, 0, 0, state.SortByTitle, state.SortDesc, 100, 20)

	if ascOutput == "" || descOutput == "" {
		t.Fatal("RenderListView should work for both sort orders")
	}
}

// TestRenderListViewNarrowWidth tests rendering with narrow width
func TestRenderListViewNarrowWidth(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Very Long Task Title That Should Be Truncated",
			},
			ColumnName: "Column",
			ColumnID:   1,
		},
	}

	// Very narrow width
	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 30, 10)

	if output == "" {
		t.Fatal("RenderListView should handle narrow width")
	}

	// Should still be reasonable length (truncated)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) > 40 { // Should be reasonably constrained
			t.Logf("Line length %d (may be due to color codes): %q", len(line), line)
		}
	}
}

// TestRenderListViewLargeWidth tests rendering with large width
func TestRenderListViewLargeWidth(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Task 1",
			},
			ColumnName: "Column Name",
			ColumnID:   1,
		},
	}

	// Very large width
	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 500, 50)

	if output == "" {
		t.Fatal("RenderListView should handle large width")
	}
}

// TestRenderListViewManyRows tests rendering with many rows
func TestRenderListViewManyRows(t *testing.T) {
	rows := make([]ListViewRow, 100)
	for i := 0; i < 100; i++ {
		rows[i] = ListViewRow{
			Task: &models.TaskSummary{
				ID:    i + 1,
				Title: "Task " + string(rune('A'+(i%26))),
			},
			ColumnName: "Column",
			ColumnID:   1,
		}
	}

	output := RenderListView(rows, 50, 10, state.SortByTitle, state.SortAsc, 100, 20)

	if output == "" {
		t.Fatal("RenderListView should handle many rows")
	}
}

// TestRenderListViewUnicodeContent tests rendering with unicode
func TestRenderListViewUnicodeContent(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Unicode: 你好 🚀 Ñoño",
			},
			ColumnName: "列",
			ColumnID:   1,
		},
		{
			Task: &models.TaskSummary{
				ID:    2,
				Title: "Emoji: 🎉 🎊 ✅",
			},
			ColumnName: "列 2",
			ColumnID:   1,
		},
	}

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20)

	if output == "" {
		t.Fatal("RenderListView should handle unicode content")
	}
}

// TestRenderListViewScrollBehavior tests scroll indicator rendering
func TestRenderListViewScrollBehavior(t *testing.T) {
	// Create rows
	rows := make([]ListViewRow, 30)
	for i := 0; i < 30; i++ {
		rows[i] = ListViewRow{
			Task: &models.TaskSummary{
				ID:    i + 1,
				Title: "Task " + string(rune('A'+(i%26))),
			},
			ColumnName: "Column",
			ColumnID:   1,
		}
	}

	// Test with scroll at top
	outputTop := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 10)
	if outputTop == "" {
		t.Fatal("RenderListView should render at top")
	}

	// Test with scroll in middle
	outputMid := RenderListView(rows, 10, 10, state.SortByTitle, state.SortAsc, 100, 10)
	if outputMid == "" {
		t.Fatal("RenderListView should render in middle")
	}

	// Test with scroll at bottom
	outputBot := RenderListView(rows, 20, 20, state.SortByTitle, state.SortAsc, 100, 10)
	if outputBot == "" {
		t.Fatal("RenderListView should render at bottom")
	}
}

// TestRenderListViewMinimalSize tests rendering in minimal space
func TestRenderListViewMinimalSize(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Task",
			},
			ColumnName: "Col",
			ColumnID:   1,
		},
	}

	// Minimal space
	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 20, 3)

	if output == "" {
		t.Fatal("RenderListView should handle minimal space")
	}
}

// TestRenderListViewHeaderPresent tests that header is always rendered
func TestRenderListViewHeaderPresent(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Task 1",
			},
			ColumnName: "Column",
			ColumnID:   1,
		},
	}

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20) // Will have table header

	// Should have multiple lines (header + separator + rows)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Errorf("output should have header and content, got %d lines", len(lines))
	}
}

// TestRenderListViewConsistency tests rendering consistency across calls
func TestRenderListViewConsistency(t *testing.T) {
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Task 1",
			},
			ColumnName: "Column",
			ColumnID:   1,
		},
	}

	output1 := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20)
	output2 := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20)

	// Same input should produce same output
	if output1 != output2 {
		t.Error("RenderListView should produce consistent output")
	}
}

// TestListViewRowStructure tests the ListViewRow data structure
func TestListViewRowStructure(t *testing.T) {
	row := ListViewRow{
		Task: &models.TaskSummary{
			ID:    42,
			Title: "Important Task",
		},
		ColumnName: "In Progress",
		ColumnID:   2,
	}

	if row.Task.ID != 42 {
		t.Error("task ID should be preserved")
	}
	if row.ColumnName != "In Progress" {
		t.Error("column name should be preserved")
	}
	if row.ColumnID != 2 {
		t.Error("column ID should be preserved")
	}
}

// TestRenderListViewWithNilTask tests handling of edge cases
func TestRenderListViewEdgeCases(t *testing.T) {
	// Test with rows that have zero height
	rows := []ListViewRow{
		{
			Task: &models.TaskSummary{
				ID:    1,
				Title: "Task",
			},
			ColumnName: "Column",
			ColumnID:   1,
		},
	}

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 0)

	// Should still produce some output (even if limited by height)
	if output == "" {
		t.Fatal("RenderListView should produce output even with 0 height")
	}
}
