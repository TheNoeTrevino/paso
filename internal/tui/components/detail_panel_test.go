package components

import (
	"strings"
	"testing"
	"time"

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
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestRenderDetailPanel_NilTask(t *testing.T) {
	t.Parallel()

	result := RenderDetailPanel(nil, 80, 40)

	if result == "" {
		t.Fatal("RenderDetailPanel(nil) should not return empty string")
	}

	if !strings.Contains(result, "Select a task to view details") {
		t.Error("RenderDetailPanel(nil) should contain empty state message")
	}
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

	if result == "" {
		t.Fatal("RenderDetailPanel should not return empty string")
	}

	// Should contain ticket number
	if !strings.Contains(result, "PROJ-42") {
		t.Error("Should contain ticket number PROJ-42")
	}

	// Should contain title
	if !strings.Contains(result, "Test Task") {
		t.Error("Should contain task title")
	}

	// Should contain status
	if !strings.Contains(result, "Status") {
		t.Error("Should contain Status label")
	}

	// Should contain footer
	if !strings.Contains(result, "Press Enter to view/edit") {
		t.Error("Should contain footer hint")
	}
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

	if result == "" {
		t.Fatal("RenderDetailPanel should not return empty string")
	}

	// Should contain ticket number
	if !strings.Contains(result, "TEST-123") {
		t.Error("Should contain ticket number")
	}

	// Should contain title
	if !strings.Contains(result, "Full Feature Task") {
		t.Error("Should contain task title")
	}

	// Should contain type
	if !strings.Contains(result, "Type") {
		t.Error("Should contain Type label")
	}

	// Should contain priority
	if !strings.Contains(result, "Priority") {
		t.Error("Should contain Priority label")
	}

	// Should contain blocked indicator
	if !strings.Contains(result, "BLOCKED") {
		t.Error("Should contain BLOCKED indicator")
	}

	// Should contain labels section
	if !strings.Contains(result, "Labels") {
		t.Error("Should contain Labels section")
	}

	// Should contain description section
	if !strings.Contains(result, "Description") {
		t.Error("Should contain Description section")
	}

	// Should contain relations section
	if !strings.Contains(result, "Relations") {
		t.Error("Should contain Relations section")
	}

	// Should contain comments section
	if !strings.Contains(result, "Comments") {
		t.Error("Should contain Comments section")
	}

	// Should show comment count (2 comments)
	if !strings.Contains(result, "2 comments") {
		t.Error("Should show correct comment count")
	}

	// Should contain timestamps section
	if !strings.Contains(result, "Created") {
		t.Error("Should contain Created timestamp")
	}
	if !strings.Contains(result, "Updated") {
		t.Error("Should contain Updated timestamp")
	}
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

	// Should NOT contain optional sections
	if strings.Contains(result, "Labels") {
		t.Error("Should not contain Labels section when no labels")
	}
	if strings.Contains(result, "Description") {
		t.Error("Should not contain Description section when no description")
	}
	if strings.Contains(result, "Relations") {
		t.Error("Should not contain Relations section when no relations")
	}
	if strings.Contains(result, "Comments") {
		t.Error("Should not contain Comments section when no comments")
	}
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

	// Should show singular "comment" not "comments"
	if !strings.Contains(result, "1 comment)") {
		t.Error("Should show singular 'comment' for single comment")
	}
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

	// Should contain ellipsis indicating truncation
	if !strings.Contains(result, "...") {
		t.Error("Long description should be truncated with ellipsis")
	}
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

	// Should truncate the comment preview
	if !strings.Contains(result, "...") {
		t.Error("Long comment should be truncated with ellipsis")
	}

	// Should not contain the full message
	if strings.Contains(result, longMessage) {
		t.Error("Should not contain full long comment message")
	}
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
			result := RenderDetailPanelLoading(80, 40, tt.spinnerFrame)

			if result == "" {
				t.Fatal("RenderDetailPanelLoading should not return empty string")
			}

			if !strings.Contains(result, "Loading task details") {
				t.Error("Should contain loading message")
			}
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

	if result == "" {
		t.Fatal("RenderDetailPanel should handle narrow width")
	}

	// Should still contain essential elements
	if !strings.Contains(result, "PROJECT-999") {
		t.Error("Should contain ticket number even in narrow panel")
	}
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

	// Should not panic and should return something
	if result == "" {
		t.Fatal("RenderDetailPanel should handle very small dimensions")
	}
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

	// Should contain relations section
	if !strings.Contains(result, "Relations") {
		t.Error("Should contain Relations section")
	}

	// Should contain parent arrow
	if !strings.Contains(result, "↑") {
		t.Error("Should contain upward arrow for parent relations")
	}
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

	// Should contain downward arrow for child relations
	if !strings.Contains(result, "↓") {
		t.Error("Should contain downward arrow for child relations")
	}
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

	// Should contain Labels section
	if !strings.Contains(result, "Labels") {
		t.Error("Should contain Labels section")
	}

	// Each label should be rendered (checking for label names)
	for _, label := range task.Labels {
		if !strings.Contains(result, label.Name) {
			t.Errorf("Should contain label name %q", label.Name)
		}
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

	// Should contain formatted timestamps
	if !strings.Contains(result, "Mar 15, 2024") {
		t.Error("Should contain formatted created date")
	}
	if !strings.Contains(result, "Jun 20, 2024") {
		t.Error("Should contain formatted updated date")
	}
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

	// Should NOT contain BLOCKED indicator
	if strings.Contains(result, "BLOCKED") {
		t.Error("Should not contain BLOCKED indicator when task is not blocked")
	}
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

	// Should contain type
	if !strings.Contains(result, "Type") {
		t.Error("Should contain Type label")
	}
	if !strings.Contains(result, "Bug") {
		t.Error("Should contain type description")
	}
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

	// Should contain priority
	if !strings.Contains(result, "Priority") {
		t.Error("Should contain Priority label")
	}
	if !strings.Contains(result, "Critical") {
		t.Error("Should contain priority description")
	}
}
