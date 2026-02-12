package components

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

func init() {
	// Initialize styles for tests that need rendering
	InitStyles(*colors.Default())
}

func TestTruncateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "string shorter than maxLen",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "string equal to maxLen",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "string longer than maxLen",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "maxLen is zero",
			input:  "hello",
			maxLen: 0,
			want:   "",
		},
		{
			name:   "maxLen is negative",
			input:  "hello",
			maxLen: -5,
			want:   "",
		},
		{
			name:   "maxLen is 1",
			input:  "hello",
			maxLen: 1,
			want:   "h",
		},
		{
			name:   "maxLen is 2",
			input:  "hello",
			maxLen: 2,
			want:   "he",
		},
		{
			name:   "maxLen is 3",
			input:  "hello",
			maxLen: 3,
			want:   "hel",
		},
		{
			name:   "maxLen is 4 (first to use ellipsis)",
			input:  "hello world",
			maxLen: 4,
			want:   "h...",
		},
		{
			name:   "very long string",
			input:  strings.Repeat("a", 1000),
			maxLen: 10,
			want:   "aaaaaaa...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			got := truncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRenderDetailPanel_NilTask(t *testing.T) {
	t.Parallel()

	result := RenderDetailPanel(nil, 80, 40)

	require.NotEmpty(t, result)
	assert.Contains(t, result, "Select a task to view details")
}

func TestRenderDetailPanel_MinimalTask(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Test Task",
		TicketNumber: 42,
		ProjectName:  "PROJ",
		ColumnName:   "Todo",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result := RenderDetailPanel(task, 80, 40)

	require.NotEmpty(t, result)
	assert.Contains(t, result, "PROJ-42")
	assert.Contains(t, result, "Test Task")
	assert.Contains(t, result, "Status")
	assert.Contains(t, result, "Press Enter to view/edit")
}

func TestRenderDetailPanel_FullTask(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:                  1,
		Title:               "Full Feature Task",
		Description:         "This is a detailed description of the task that explains what needs to be done.",
		TicketNumber:        123,
		ProjectName:         "TEST",
		ColumnName:          "In Progress",
		TypeDescription:     "Feature",
		PriorityDescription: "High",
		PriorityColor:       "#FF0000",
		IsBlocked:           true,
		Labels: []*models.Label{
			{ID: 1, Name: "frontend", Color: "#3B82F6"},
			{ID: 2, Name: "urgent", Color: "#EF4444"},
		},
		ParentTasks: []*models.TaskReference{
			{
				ID:            10,
				TicketNumber:  100,
				Title:         "Parent Task",
				ProjectName:   "TEST",
				RelationLabel: "Blocks",
				RelationColor: "#FF0000",
				IsBlocking:    true,
			},
		},
		ChildTasks: []*models.TaskReference{
			{
				ID:            20,
				TicketNumber:  200,
				Title:         "Child Task",
				ProjectName:   "TEST",
				RelationLabel: "Blocked by",
				RelationColor: "#00FF00",
				IsBlocking:    false,
			},
		},
		Comments: []*models.Comment{
			{ID: 1, Author: "Alice", Message: "First comment"},
			{ID: 2, Author: "Bob", Message: "Second comment that is the most recent"},
		},
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 20, 14, 45, 0, 0, time.UTC),
	}

	result := RenderDetailPanel(task, 100, 50)

	require.NotEmpty(t, result)
	assert.Contains(t, result, "TEST-123")
	assert.Contains(t, result, "Full Feature Task")
	assert.Contains(t, result, "Type")
	assert.Contains(t, result, "Priority")
	assert.Contains(t, result, "BLOCKED")
	assert.Contains(t, result, "Labels")
	assert.Contains(t, result, "Description")
	assert.Contains(t, result, "Relations")
	assert.Contains(t, result, "Comments")
	assert.Contains(t, result, "2 comments")
	assert.Contains(t, result, "Created")
	assert.Contains(t, result, "Updated")
}

func TestRenderDetailPanel_WithoutOptionalSections(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Simple Task",
		TicketNumber: 1,
		ProjectName:  "PROJ",
		ColumnName:   "Done",
		// No description, labels, relations, or comments
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := RenderDetailPanel(task, 80, 40)

	assert.NotContains(t, result, "Labels")
	assert.NotContains(t, result, "Description")
	assert.NotContains(t, result, "Relations")
	assert.NotContains(t, result, "Comments")
}

func TestRenderDetailPanel_SingleComment(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Task with one comment",
		TicketNumber: 1,
		ProjectName:  "PROJ",
		ColumnName:   "Todo",
		Comments: []*models.Comment{
			{ID: 1, Author: "Alice", Message: "Only comment"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := RenderDetailPanel(task, 80, 40)

	assert.Contains(t, result, "1 comment)")
}

func TestRenderDetailPanel_LongDescription(t *testing.T) {
	t.Parallel()

	// Create a description that will exceed 6 lines when wrapped
	longDescription := strings.Repeat("This is a long line of text that will wrap. ", 20)

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Task",
		Description:  longDescription,
		TicketNumber: 1,
		ProjectName:  "PROJ",
		ColumnName:   "Todo",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result := RenderDetailPanel(task, 80, 60)

	assert.Contains(t, result, "...")
}

func TestRenderDetailPanel_LongCommentPreview(t *testing.T) {
	t.Parallel()

	// Create a comment longer than commentPreviewLength (80)
	longMessage := strings.Repeat("a", 100)

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Task",
		TicketNumber: 1,
		ProjectName:  "PROJ",
		ColumnName:   "Todo",
		Comments: []*models.Comment{
			{ID: 1, Author: "Alice", Message: longMessage},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := RenderDetailPanel(task, 100, 50)

	assert.Contains(t, result, "...")
	assert.NotContains(t, result, longMessage)
}

func TestRenderDetailPanelLoading(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		spinnerFrame int
	}{
		{"frame 0", 0},
		{"frame 5", 5},
		{"frame 11", 11},
		{"frame wraps at 12", 12},
		{"large frame number", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			result := RenderDetailPanelLoading(80, 40, tt.spinnerFrame)

			require.NotEmpty(t, result)
			assert.Contains(t, result, "Loading task details")
		})
	}
}

func TestRenderDetailPanel_NarrowWidth(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:           1,
		Title:        "This is a very long task title that should wrap properly in narrow panels",
		TicketNumber: 999,
		ProjectName:  "PROJECT",
		ColumnName:   "In Progress",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Very narrow width
	result := RenderDetailPanel(task, 50, 30)

	require.NotEmpty(t, result)
	assert.Contains(t, result, "PROJECT-999")
}

func TestRenderDetailPanel_VerySmallDimensions(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Task",
		TicketNumber: 1,
		ProjectName:  "P",
		ColumnName:   "Todo",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Minimum viable dimensions
	result := RenderDetailPanel(task, 20, 10)

	require.NotEmpty(t, result)
}

func TestRenderDetailPanel_RelationsWithBlockingParent(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Blocked Task",
		TicketNumber: 1,
		ProjectName:  "PROJ",
		ColumnName:   "Todo",
		ParentTasks: []*models.TaskReference{
			{
				ID:            10,
				TicketNumber:  10,
				Title:         "Blocking Parent",
				ProjectName:   "PROJ",
				RelationLabel: "Blocked by",
				RelationColor: "#FF0000",
				IsBlocking:    true,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := RenderDetailPanel(task, 100, 50)

	assert.Contains(t, result, "Relations")
	assert.Contains(t, result, "↑")
}

func TestRenderDetailPanel_RelationsWithChild(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Parent Task",
		TicketNumber: 1,
		ProjectName:  "PROJ",
		ColumnName:   "Todo",
		ChildTasks: []*models.TaskReference{
			{
				ID:            20,
				TicketNumber:  20,
				Title:         "Child Task",
				ProjectName:   "PROJ",
				RelationLabel: "Blocks",
				RelationColor: "#00FF00",
				IsBlocking:    false,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := RenderDetailPanel(task, 100, 50)

	assert.Contains(t, result, "↓")
}

func TestRenderDetailPanel_MultipleLabels(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Task with many labels",
		TicketNumber: 1,
		ProjectName:  "PROJ",
		ColumnName:   "Todo",
		Labels: []*models.Label{
			{ID: 1, Name: "bug", Color: "#FF0000"},
			{ID: 2, Name: "urgent", Color: "#FFA500"},
			{ID: 3, Name: "frontend", Color: "#00FF00"},
			{ID: 4, Name: "backend", Color: "#0000FF"},
			{ID: 5, Name: "docs", Color: "#800080"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := RenderDetailPanel(task, 100, 50)

	assert.Contains(t, result, "Labels")

	// Each label should be rendered (checking for label names)
	for _, label := range task.Labels {
		assert.Contains(t, result, label.Name)
	}
}

func TestRenderDetailPanel_Timestamps(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2024, 3, 15, 9, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC)

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Task",
		TicketNumber: 1,
		ProjectName:  "PROJ",
		ColumnName:   "Done",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	result := RenderDetailPanel(task, 100, 50)

	assert.Contains(t, result, "Mar 15, 2024")
	assert.Contains(t, result, "Jun 20, 2024")
}

func TestRenderDetailPanel_NotBlocked(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:           1,
		Title:        "Normal Task",
		TicketNumber: 1,
		ProjectName:  "PROJ",
		ColumnName:   "Todo",
		IsBlocked:    false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result := RenderDetailPanel(task, 100, 50)

	assert.NotContains(t, result, "BLOCKED")
}

func TestRenderDetailPanel_WithTypeOnly(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:              1,
		Title:           "Typed Task",
		TicketNumber:    1,
		ProjectName:     "PROJ",
		ColumnName:      "Todo",
		TypeDescription: "Bug",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	result := RenderDetailPanel(task, 100, 50)

	assert.Contains(t, result, "Type")
	assert.Contains(t, result, "Bug")
}

func TestRenderDetailPanel_WithPriorityOnly(t *testing.T) {
	t.Parallel()

	task := &models.TaskDetail{
		ID:                  1,
		Title:               "Priority Task",
		TicketNumber:        1,
		ProjectName:         "PROJ",
		ColumnName:          "Todo",
		PriorityDescription: "Critical",
		PriorityColor:       "#EF4444",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	result := RenderDetailPanel(task, 100, 50)

	assert.Contains(t, result, "Priority")
	assert.Contains(t, result, "Critical")
}
