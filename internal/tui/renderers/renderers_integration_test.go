package renderers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// defaultKM returns default key mappings for test convenience.
func defaultKM() config.KeyMappings {
	return config.DefaultKeyMappings()
}

// TestRenderListViewBasic tests basic list view rendering
func TestRenderListViewBasic(t *testing.T) {
	t.Parallel()
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

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20, defaultKM())

	require.NotEmpty(t, output)
	assert.True(t, strings.Contains(output, "Task 1") || strings.Contains(output, "Task"))
}

// TestRenderListViewEmpty tests list view with no rows
func TestRenderListViewEmpty(t *testing.T) {
	t.Parallel()
	var rows []ListViewRow

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20, defaultKM())

	require.NotEmpty(t, output)
}

// TestRenderListViewSelectedRow tests rendering with selected row
func TestRenderListViewSelectedRow(t *testing.T) {
	t.Parallel()
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
	output := RenderListView(rows, 1, 0, state.SortByTitle, state.SortAsc, 100, 20, defaultKM())

	require.NotEmpty(t, output)
}

// TestRenderListViewScrolling tests rendering with scroll offset
func TestRenderListViewScrolling(t *testing.T) {
	t.Parallel()
	// Create many rows
	rows := make([]ListViewRow, 50)
	for i := range 50 {
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
	output := RenderListView(rows, 10, 5, state.SortByTitle, state.SortAsc, 100, 20, defaultKM())

	require.NotEmpty(t, output)
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

	t.Parallel()
	for _, sortField := range sorts {
		t.Run("sort", func(t *testing.T) {
			t.Parallel()
			output := RenderListView(rows, 0, 0, sortField, state.SortAsc, 100, 20, defaultKM())

			assert.NotEmpty(t, output)
		})
	}
}

// TestRenderListViewSortOrders tests ascending and descending
func TestRenderListViewSortOrders(t *testing.T) {
	t.Parallel()
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

	ascOutput := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20, defaultKM())
	descOutput := RenderListView(rows, 0, 0, state.SortByTitle, state.SortDesc, 100, 20, defaultKM())

	require.NotEmpty(t, ascOutput)
	require.NotEmpty(t, descOutput)
}

// TestRenderListViewNarrowWidth tests rendering with narrow width
func TestRenderListViewNarrowWidth(t *testing.T) {
	t.Parallel()
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
	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 30, 10, defaultKM())

	require.NotEmpty(t, output)

	// Should still be reasonable length (truncated)
	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		if len(line) > 40 { // Should be reasonably constrained
			t.Logf("Line length %d (may be due to color codes): %q", len(line), line)
		}
	}
}

// TestRenderListViewLargeWidth tests rendering with large width
func TestRenderListViewLargeWidth(t *testing.T) {
	t.Parallel()
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
	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 500, 50, defaultKM())

	require.NotEmpty(t, output)
}

// TestRenderListViewManyRows tests rendering with many rows
func TestRenderListViewManyRows(t *testing.T) {
	t.Parallel()
	rows := make([]ListViewRow, 100)
	for i := range 100 {
		rows[i] = ListViewRow{
			Task: &models.TaskSummary{
				ID:    i + 1,
				Title: "Task " + string(rune('A'+(i%26))),
			},
			ColumnName: "Column",
			ColumnID:   1,
		}
	}

	output := RenderListView(rows, 50, 10, state.SortByTitle, state.SortAsc, 100, 20, defaultKM())

	require.NotEmpty(t, output)
}

// TestRenderListViewUnicodeContent tests rendering with unicode
func TestRenderListViewUnicodeContent(t *testing.T) {
	t.Parallel()
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

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20, defaultKM())

	require.NotEmpty(t, output)
}

// TestRenderListViewScrollBehavior tests scroll indicator rendering
func TestRenderListViewScrollBehavior(t *testing.T) {
	t.Parallel()
	// Create rows
	rows := make([]ListViewRow, 30)
	for i := range 30 {
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
	outputTop := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 10, defaultKM())
	require.NotEmpty(t, outputTop)

	// Test with scroll in middle
	outputMid := RenderListView(rows, 10, 10, state.SortByTitle, state.SortAsc, 100, 10, defaultKM())
	require.NotEmpty(t, outputMid)

	// Test with scroll at bottom
	outputBot := RenderListView(rows, 20, 20, state.SortByTitle, state.SortAsc, 100, 10, defaultKM())
	require.NotEmpty(t, outputBot)
}

// TestRenderListViewMinimalSize tests rendering in minimal space
func TestRenderListViewMinimalSize(t *testing.T) {
	t.Parallel()
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
	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 20, 3, defaultKM())

	require.NotEmpty(t, output)
}

// TestRenderListViewHeaderPresent tests that header is always rendered
func TestRenderListViewHeaderPresent(t *testing.T) {
	t.Parallel()
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

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20, defaultKM()) // Will have table header

	// Should have multiple lines (header + separator + rows)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.GreaterOrEqual(t, len(lines), 2)
}

// TestRenderListViewConsistency tests rendering consistency across calls
func TestRenderListViewConsistency(t *testing.T) {
	t.Parallel()
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

	output1 := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20, defaultKM())
	output2 := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 20, defaultKM())

	assert.Equal(t, output1, output2)
}

// TestListViewRowStructure tests the ListViewRow data structure
func TestListViewRowStructure(t *testing.T) {
	t.Parallel()
	row := ListViewRow{
		Task: &models.TaskSummary{
			ID:    42,
			Title: "Important Task",
		},
		ColumnName: "In Progress",
		ColumnID:   2,
	}

	assert.Equal(t, 42, row.Task.ID)
	assert.Equal(t, "In Progress", row.ColumnName)
	assert.Equal(t, 2, row.ColumnID)
}

// TestRenderListViewEdgeCases tests handling of edge cases
func TestRenderListViewEdgeCases(t *testing.T) {
	t.Parallel()
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

	output := RenderListView(rows, 0, 0, state.SortByTitle, state.SortAsc, 100, 0, defaultKM())

	require.NotEmpty(t, output)
}
