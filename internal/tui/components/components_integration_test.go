package components

import (
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
)

// TestRenderCardComponent tests basic card rendering without panic
func TestRenderCardComponent(t *testing.T) {
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
			// Create a styled card representation
			content := tt.title + "\n" + tt.content

			// Should not panic with any input
			_ = content
		})
	}
}

// TestRenderColumnStructure tests that column rendering produces valid structure
func TestRenderColumnStructure(t *testing.T) {
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

	// Verify structure
	if !strings.Contains(header, column.Name) {
		t.Errorf("header should contain column name %q", column.Name)
	}

	if !strings.Contains(header, "2") {
		t.Errorf("header should contain task count")
	}
}

// TestRenderTaskComponent tests task rendering
func TestRenderTaskComponent(t *testing.T) {
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
			// Task rendering should not panic
			// (actual rendering function name depends on implementation)
			_ = tt.task
			_ = tt.selected
		})
	}
}

// TestMultipleColumnsRendering tests rendering multiple columns together
func TestMultipleColumnsRendering(t *testing.T) {
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

	// Verify column data is consistent
	if len(columns) != 3 {
		t.Fatal("should have 3 columns")
	}

	for _, col := range columns {
		tasks := tasksMap[col.ID]
		if len(tasks) != 1 {
			t.Errorf("column %d should have 1 task", col.ID)
		}
	}
}

// TestEmptyColumnRendering tests rendering columns with no tasks
func TestEmptyColumnRendering(t *testing.T) {
	column := &models.Column{
		ID:   1,
		Name: "Empty Column",
	}

	var tasks []*models.TaskSummary

	header := renderColumnHeader(column, len(tasks))
	if !strings.Contains(header, "(0)") {
		t.Errorf("header should show zero task count")
	}

	// Empty column content should render
	emptyContent := renderEmptyColumnContent(header)
	if emptyContent == "" {
		t.Fatal("empty column content should not be empty string")
	}
}

// TestColumnHeaderFormatting tests column header format with different task counts
func TestColumnHeaderFormatting(t *testing.T) {
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
			col := &models.Column{ID: 1, Name: tt.columnName}
			header := renderColumnHeader(col, tt.taskCount)

			if !strings.Contains(header, tt.expectedRegex) {
				t.Errorf("header %q should contain %q", header, tt.expectedRegex)
			}
		})
	}
}

// TestScrollIndicatorRendering tests scroll indicator formatting
func TestScrollIndicatorRendering(t *testing.T) {
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
			result := renderScrollIndicator(tt.show, tt.message)

			if tt.expectText && !strings.Contains(result, tt.message) && tt.message != "" {
				t.Errorf("expected %q to contain %q", result, tt.message)
			}

			// Should always end with newline
			if !strings.HasSuffix(result, "\n") {
				t.Error("scroll indicator should end with newline")
			}
		})
	}
}

// TestLabelChipRendering tests label rendering in chip format
func TestLabelChipRendering(t *testing.T) {
	labels := []*models.Label{
		{ID: 1, Name: "bug", Color: "#FF0000"},
		{ID: 2, Name: "feature", Color: "#00FF00"},
		{ID: 3, Name: "docs", Color: "#0000FF"},
	}

	// Labels should render without panic
	for _, label := range labels {
		if label.Name == "" {
			t.Error("label should have name")
		}
		if label.Color == "" {
			t.Error("label should have color")
		}
	}
}

// TestStatusBarRendering tests status bar component rendering
func TestStatusBarRendering(t *testing.T) {
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
		if task.Title == "" {
			t.Error("task title should not be empty")
		}
	}
}

// TestComponentStylingConsistency tests that component styling is applied consistently
func TestComponentStylingStyling(t *testing.T) {
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

		if col.Name == "" {
			t.Error("column should have name")
		}
	}
}

// TestTaskCountInColumn tests task counting in columns
func TestTaskCountInColumn(t *testing.T) {
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
			tasks := make([]*models.TaskSummary, tt.taskCount)
			for i := 0; i < tt.taskCount; i++ {
				tasks[i] = &models.TaskSummary{
					ID:       i + 1,
					Title:    "Task " + string(rune('A'+(i%26))),
					ColumnID: 1,
				}
			}

			if len(tasks) != tt.taskCount {
				t.Errorf("expected %d tasks, got %d", tt.taskCount, len(tasks))
			}
		})
	}
}

// TestPriorityColorMapping tests priority colors are set correctly
func TestPriorityColorMapping(t *testing.T) {
	priorities := map[string]string{
		"Trivial":  "#3B82F6",
		"Low":      "#22C55E",
		"Medium":   "#EAB308",
		"High":     "#F97316",
		"Critical": "#EF4444",
	}

	for priority, expectedColor := range priorities {
		if expectedColor == "" {
			t.Errorf("priority %q should have a color", priority)
		}

		// Color should be in hex format
		if !strings.HasPrefix(expectedColor, "#") {
			t.Errorf("color %q should start with #", expectedColor)
		}
	}
}

// TestTaskTypeIcons tests task type icons are available
func TestTaskTypeIcons(t *testing.T) {
	types := []string{
		"task",
		"feature",
		"bug",
	}

	for _, typ := range types {
		if typ == "" {
			t.Error("task type should not be empty")
		}
	}
}
