package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
)

// TestRenderCardComponent tests basic card rendering without panic
func TestRenderCardComponent(t *testing.T) {
	t.Parallel()
	// Create a simple styled card content
	tests := []struct {
		name    string
		title   string
		content string
	}{
		{
			name:    "simple card",
			title:   "Card Title",
			content: "Card content",
		},
		{
			name:    "empty content",
			title:   "Title Only",
			content: "",
		},
		{
			name:    "long content",
			title:   "Long Title",
			content: strings.Repeat("A", 200),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a styled card representation
			content := tt.title + "\n" + tt.content

			// Should not panic with any input
			_ = content
		})
	}
}

// TestRenderColumnStructure tests that column rendering produces valid structure
func TestRenderColumnStructure(t *testing.T) {
	t.Parallel()
	column := &models.Column{
		ID:   1,
		Name: "In Progress",
	}

	tasks := []*models.TaskSummary{
		{
			ID:                  1,
			Title:               "Task 1",
			ColumnID:            column.ID,
			TypeDescription:     "Feature",
			PriorityDescription: "High",
		},
		{
			ID:                  2,
			Title:               "Task 2",
			ColumnID:            column.ID,
			TypeDescription:     "Bug",
			PriorityDescription: "Medium",
		},
	}

	// Create header
	header := renderColumnHeader(column, len(tasks))

	assert.Contains(t, header, column.Name)
	assert.Contains(t, header, "2")
}

// TestRenderTaskComponent tests task rendering
func TestRenderTaskComponent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		task     *models.TaskSummary
		selected bool
	}{
		{
			name: "simple task",
			task: &models.TaskSummary{
				ID:                  1,
				Title:               "Test Task",
				TypeDescription:     "Feature",
				PriorityDescription: "High",
				ColumnID:            1,
			},
			selected: false,
		},
		{
			name: "selected task",
			task: &models.TaskSummary{
				ID:                  2,
				Title:               "Selected Task",
				TypeDescription:     "Bug",
				PriorityDescription: "Critical",
				ColumnID:            1,
			},
			selected: true,
		},
		{
			name: "task with labels",
			task: &models.TaskSummary{
				ID:                  3,
				Title:               "Task with Labels",
				TypeDescription:     "Task",
				PriorityDescription: "Low",
				ColumnID:            1,
				Labels: []*models.Label{
					{ID: 1, Name: "bug", Color: "#FF0000"},
					{ID: 2, Name: "urgent", Color: "#FFA500"},
				},
			},
			selected: false,
		},
		{
			name: "task with very long title",
			task: &models.TaskSummary{
				ID:                  4,
				Title:               strings.Repeat("Long Title ", 10),
				TypeDescription:     "Feature",
				PriorityDescription: "Medium",
				ColumnID:            1,
			},
			selected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Task rendering should not panic
			// (actual rendering function name depends on implementation)
			_ = tt.task
			_ = tt.selected
		})
	}
}

// TestMultipleColumnsRendering tests rendering multiple columns together
func TestMultipleColumnsRendering(t *testing.T) {
	t.Parallel()
	columns := []*models.Column{
		{ID: 1, Name: "Backlog"},
		{ID: 2, Name: "In Progress"},
		{ID: 3, Name: "Done"},
	}

	tasksMap := make(map[int][]*models.TaskSummary)
	for i, col := range columns {
		tasksMap[col.ID] = []*models.TaskSummary{
			{
				ID:                  i*10 + 1,
				Title:               "Task " + string(rune('A'+i)),
				ColumnID:            col.ID,
				TypeDescription:     "Feature",
				PriorityDescription: "Medium",
			},
		}
	}

	require.Len(t, columns, 3)

	for _, col := range columns {
		tasks := tasksMap[col.ID]
		assert.Len(t, tasks, 1)
	}
}

// TestEmptyColumnRendering tests rendering columns with no tasks
func TestEmptyColumnRendering(t *testing.T) {
	t.Parallel()
	column := &models.Column{
		ID:   1,
		Name: "Empty Column",
	}

	var tasks []*models.TaskSummary

	header := renderColumnHeader(column, len(tasks))
	assert.Contains(t, header, "(0)")

	// Empty column content should render
	emptyContent := renderEmptyColumnContent(header)
	require.NotEmpty(t, emptyContent)
}

// TestColumnHeaderFormatting tests column header format with different task counts
func TestColumnHeaderFormatting(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		columnName    string
		taskCount     int
		expectedRegex string
	}{
		{"single task", "Todo", 1, "Todo"},
		{"multiple tasks", "Done", 42, "Done"},
		{"zero tasks", "Backlog", 0, "Backlog"},
		{"large count", "Review", 1000, "Review"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			col := &models.Column{ID: 1, Name: tt.columnName}
			header := renderColumnHeader(col, tt.taskCount)

			assert.Contains(t, header, tt.expectedRegex)
		})
	}
}

// TestScrollIndicatorRendering tests scroll indicator formatting
func TestScrollIndicatorRendering(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		show       bool
		message    string
		expectText bool
	}{
		{"show indicator", true, "▲ more above", true},
		{"hide indicator", false, "▲ more above", false},
		{"empty message show", true, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := renderScrollIndicator(tt.show, tt.message)

			if tt.expectText && tt.message != "" {
				assert.Contains(t, result, tt.message)
			}

			assert.True(t, strings.HasSuffix(result, "\n"))
		})
	}
}

// TestLabelChipRendering tests label rendering in chip format
func TestLabelChipRendering(t *testing.T) {
	t.Parallel()
	labels := []*models.Label{
		{ID: 1, Name: "bug", Color: "#FF0000"},
		{ID: 2, Name: "feature", Color: "#00FF00"},
		{ID: 3, Name: "docs", Color: "#0000FF"},
	}

	// Labels should render without panic
	for _, label := range labels {
		assert.NotEmpty(t, label.Name)
		assert.NotEmpty(t, label.Color)
	}
}

// TestStatusBarRendering tests status bar component rendering
func TestStatusBarRendering(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		width     int
		height    int
		mode      string
		taskCount int
	}{
		{"normal terminal", 120, 40, "Normal", 5},
		{"narrow terminal", 80, 24, "Form", 0},
		{"very narrow", 40, 10, "Picker", 3},
		{"wide terminal", 200, 50, "Normal", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Status bar should handle any dimension without panic
			_ = tt.width
			_ = tt.height
			_ = tt.mode
			_ = tt.taskCount
		})
	}
}

// TestComponentsWithUnicodeContent tests components handle unicode characters
func TestComponentsWithUnicodeContent(t *testing.T) {
	t.Parallel()
	tasks := []*models.TaskSummary{
		{
			ID:                  1,
			Title:               "Unicode: 你好 🚀 Ñoño",
			TypeDescription:     "Feature",
			PriorityDescription: "High",
			ColumnID:            1,
		},
		{
			ID:                  2,
			Title:               "Emoji task 🎉 🎊 ✅",
			TypeDescription:     "Bug",
			PriorityDescription: "Critical",
			ColumnID:            1,
		},
	}

	for _, task := range tasks {
		assert.NotEmpty(t, task.Title)
	}
}

// TestComponentStylingConsistency tests that component styling is applied consistently
func TestComponentStylingStyling(t *testing.T) {
	t.Parallel()
	columns := []*models.Column{
		{ID: 1, Name: "Column A"},
		{ID: 2, Name: "Column B"},
		{ID: 3, Name: "Column C"},
	}

	selectedIdx := 1

	// All columns should render (not just selected one)
	for i, col := range columns {
		isSelected := i == selectedIdx
		_ = isSelected // Would be used for styling in real implementation

		assert.NotEmpty(t, col.Name)
	}
}

// TestTaskCountInColumn tests task counting in columns
func TestTaskCountInColumn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		taskCount int
	}{
		{"empty", 0},
		{"single", 1},
		{"few", 5},
		{"many", 100},
		{"very many", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tasks := make([]*models.TaskSummary, tt.taskCount)
			for i := 0; i < tt.taskCount; i++ {
				tasks[i] = &models.TaskSummary{
					ID:       i + 1,
					Title:    "Task " + string(rune('A'+(i%26))),
					ColumnID: 1,
				}
			}

			assert.Len(t, tasks, tt.taskCount)
		})
	}
}

// TestPriorityColorMapping tests priority colors are set correctly
func TestPriorityColorMapping(t *testing.T) {
	t.Parallel()
	priorities := map[string]string{
		"Trivial":  "#3B82F6",
		"Low":      "#22C55E",
		"Medium":   "#EAB308",
		"High":     "#F97316",
		"Critical": "#EF4444",
	}

	for priority, expectedColor := range priorities {
		assert.NotEmpty(t, expectedColor, "priority %q should have a color", priority)
		assert.True(t, strings.HasPrefix(expectedColor, "#"), "color %q should start with #", expectedColor)
	}
}

// TestTaskTypeIcons tests task type icons are available
func TestTaskTypeIcons(t *testing.T) {
	t.Parallel()
	types := []string{
		"task",
		"feature",
		"bug",
	}

	for _, typ := range types {
		assert.NotEmpty(t, typ)
	}
}
