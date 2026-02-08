package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestRenderTaskCardLabels_NoLabels(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	result := renderTaskCardLabels(nil, bg, 27)

	assert.Contains(t, result, "no labels")
}

func TestRenderTaskCardLabels_AllFit(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	labels := []*models.Label{
		{Name: "bug", Color: "#FF0000"},
		{Name: "fix", Color: "#00FF00"},
	}

	result := renderTaskCardLabels(labels, bg, 37)

	assert.Contains(t, result, "bug")
	assert.Contains(t, result, "fix")
	assert.NotContains(t, result, "...")
}

func TestRenderTaskCardLabels_Overflow(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	labels := []*models.Label{
		{Name: "bug", Color: "#FF0000"},
		{Name: "duplicate", Color: "#00FF00"},
		{Name: "enhancement", Color: "#0000FF"},
		{Name: "help wanted", Color: "#FFFF00"},
	}

	// Use a narrow maxWidth that can't fit all labels
	result := renderTaskCardLabels(labels, bg, 17)

	assert.Contains(t, result, "...")
	// At least one label should render
	assert.Contains(t, result, "bug")
	// "help wanted" should NOT appear since it's at the end and we're truncating
	assert.NotContains(t, result, "help wanted")
}

func TestRenderTaskCardLabels_NeverExceedsWidth(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	labels := []*models.Label{
		{Name: "bug", Color: "#FF0000"},
		{Name: "duplicate", Color: "#00FF00"},
		{Name: "enhancement", Color: "#0000FF"},
		{Name: "help wanted", Color: "#FFFF00"},
		{Name: "good first issue", Color: "#7057FF"},
	}

	widths := []int{7, 12, 17, 22, 27, 32, 37}
	for _, maxWidth := range widths {
		result := renderTaskCardLabels(labels, bg, maxWidth)
		// Strip the leading newline+space that the function prepends
		line := strings.TrimPrefix(result, "\n ")
		renderedWidth := lipgloss.Width(line)
		assert.LessOrEqual(t, renderedWidth, maxWidth,
			"label line width %d exceeded max %d",
			renderedWidth, maxWidth)
	}
}

func TestRenderTaskSummaryMetadata_NeverExceedsWidth(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	assignee := "noetrevino"
	task := &models.TaskSummary{
		TypeDescription:     "task",
		PriorityDescription: "critical",
		PriorityColor:       "#FF0000",
		AssigneeName:        &assignee,
		IsBlocked:           true,
	}

	widths := []int{12, 17, 22, 27, 32, 37, 47}
	for _, maxWidth := range widths {
		result := renderTaskSummaryMetadata(task, bg, maxWidth)
		line := strings.TrimPrefix(result, "\n ")
		renderedWidth := lipgloss.Width(line)
		assert.LessOrEqual(t, renderedWidth, maxWidth,
			"metadata line width %d exceeded max %d",
			renderedWidth, maxWidth)
	}
}

func TestRenderTaskSummaryMetadata_LongAssigneeTruncated(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	longAssignee := "averylongusernamethatwilloverflow"
	task := &models.TaskSummary{
		TypeDescription:     "task",
		PriorityDescription: "critical",
		PriorityColor:       "#FF0000",
		AssigneeName:        &longAssignee,
	}

	result := renderTaskSummaryMetadata(task, bg, 22)

	assert.Contains(t, result, "...")
	assert.NotContains(t, result, longAssignee)
}

func TestRenderTaskSummaryMetadata_AllFitWideCard(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	assignee := "noe"
	task := &models.TaskSummary{
		TypeDescription:     "task",
		PriorityDescription: "high",
		PriorityColor:       "#FF0000",
		AssigneeName:        &assignee,
	}

	// Wide enough to fit everything
	result := renderTaskSummaryMetadata(task, bg, 77)

	assert.Contains(t, result, "task")
	assert.Contains(t, result, "high")
	assert.Contains(t, result, "@noe")
	assert.NotContains(t, result, "...")
}

func TestRenderTask_MetadataDoesNotWrap(t *testing.T) {
	t.Parallel()

	assignee := "noetrevino"
	task := &models.TaskSummary{
		Title:               "Test Task",
		TypeDescription:     "task",
		PriorityDescription: "critical",
		PriorityColor:       "#FF0000",
		AssigneeName:        &assignee,
		Labels: []*models.Label{
			{Name: "bug", Color: "#FF0000"},
		},
	}

	// TaskCardHeight = 5: top border + 3 content lines + bottom border.
	// If metadata wraps, the card will have more than 5 lines.
	widths := []int{31, 32, 35, 40, 50}
	for _, width := range widths {
		rendered := RenderTask(task, false, width)
		lines := strings.Split(rendered, "\n")
		assert.Equal(t, TaskCardHeight, len(lines),
			"card at width=%d has %d lines (expected %d), metadata likely wrapped",
			width, len(lines), TaskCardHeight)
	}
}

func TestRenderTaskCardLabels_ShowsPartialLastLabel(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	labels := []*models.Label{
		{Name: "bug", Color: "#FF0000"},
		{Name: "duplicate", Color: "#00FF00"},
		{Name: "enhancement", Color: "#0000FF"},
		{Name: "help wanted", Color: "#FFFF00"},
	}

	// maxWidth=30 allows partial "help wanted" to show
	// Full: "bug duplicate enhancement help wanted" = 3+1+9+1+11+1+11 = 37
	// "bug duplicate enhancement h..." = 3+1+9+1+11+1+1+3 = 30 (fits!)
	result := renderTaskCardLabels(labels, bg, 30)

	assert.Contains(t, result, "bug")
	assert.Contains(t, result, "duplicate")
	assert.Contains(t, result, "enhancement")
	assert.Contains(t, result, "h")
	assert.Contains(t, result, "...")
	// The full "help wanted" should not appear
	assert.NotContains(t, result, "help wanted")
}

func TestRenderTaskSummaryMetadata_ShowsPartialAssignee(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	longAssignee := "noetrevino"
	task := &models.TaskSummary{
		TypeDescription:     "task",
		PriorityDescription: "critical",
		PriorityColor:       "#FF0000",
		AssigneeName:        &longAssignee,
	}

	// maxWidth=26 allows partial assignee
	// "task | critical | @noet" = 4+3+8+3+5 = 23
	// With ellipsis: 23+3 = 26
	result := renderTaskSummaryMetadata(task, bg, 26)

	assert.Contains(t, result, "task")
	assert.Contains(t, result, "critical")
	assert.Contains(t, result, "@noet")
	assert.Contains(t, result, "...")
	// The full assignee should not appear
	assert.NotContains(t, result, "@noetrevino")
}

func TestRenderTaskCardLabels_CharacterLevelTruncation(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	labels := []*models.Label{
		{Name: "bug", Color: "#FF0000"},
		{Name: "enhancement", Color: "#0000FF"},
	}

	// Test various maxWidths to ensure character-level truncation works
	// Full: "bug enhancement" = 3+1+11 = 15
	testCases := []struct {
		width                int
		shouldContainBug     bool
		shouldContainPartial bool
		partialText          string
	}{
		// maxWidth=13: "bug enhanc..." = 3+1+6+3 = 13
		{width: 13, shouldContainBug: true, shouldContainPartial: true, partialText: "enhanc"},
		// maxWidth=11: "bug enha..." = 3+1+4+3 = 11
		{width: 11, shouldContainBug: true, shouldContainPartial: true, partialText: "enha"},
		// maxWidth=8: "bug e..." = 3+1+1+3 = 8
		{width: 8, shouldContainBug: true, shouldContainPartial: true, partialText: "e"},
	}

	for _, tc := range testCases {
		result := renderTaskCardLabels(labels, bg, tc.width)
		if tc.shouldContainBug {
			assert.Contains(t, result, "bug", "width=%d should show 'bug'", tc.width)
		}
		if tc.shouldContainPartial {
			assert.Contains(t, result, tc.partialText, "width=%d should show partial '%s'", tc.width, tc.partialText)
			assert.Contains(t, result, "...", "width=%d should show ellipsis", tc.width)
		}
		assert.NotContains(t, result, "enhancement", "width=%d should not show full 'enhancement'", tc.width)
	}
}

func TestRenderTaskSummaryMetadata_CharacterLevelTruncation(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	assignee := "claude"
	task := &models.TaskSummary{
		TypeDescription:     "task",
		PriorityDescription: "critical",
		PriorityColor:       "#FF0000",
		AssigneeName:        &assignee,
		IsBlocked:           true,
	}

	// Test various maxWidths
	testCases := []struct {
		maxWidth          int
		shouldContainTask bool
		shouldContainCrit bool
	}{
		// Wide enough for task + critical + partial assignee
		{maxWidth: 32, shouldContainTask: true, shouldContainCrit: true},
		// Only task + critical
		{maxWidth: 22, shouldContainTask: true, shouldContainCrit: true},
	}

	for _, tc := range testCases {
		result := renderTaskSummaryMetadata(task, bg, tc.maxWidth)
		if tc.shouldContainTask {
			assert.Contains(t, result, "task", "maxWidth=%d should show 'task'", tc.maxWidth)
		}
		if tc.shouldContainCrit {
			assert.Contains(t, result, "critical", "maxWidth=%d should show 'critical'", tc.maxWidth)
		}
		assert.Contains(t, result, "...", "maxWidth=%d should show ellipsis", tc.maxWidth)
	}
}

func TestRenderTaskSummaryTitle_NoTruncation(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	task := &models.TaskSummary{
		Title: "Short task",
	}

	result := renderTaskSummaryTitle(task, bg, 47)

	assert.Contains(t, result, "Short task")
	assert.NotContains(t, result, "...")
}

func TestRenderTaskSummaryTitle_Truncation(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	task := &models.TaskSummary{
		Title: "This is a very long task title that should be truncated",
	}

	result := renderTaskSummaryTitle(task, bg, 17)

	assert.Contains(t, result, "...")
	assert.NotContains(t, result, "This is a very long task title that should be truncated")
}

func TestRenderTaskSummaryTitle_CharacterLevelTruncation(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	task := &models.TaskSummary{
		Title: "Implement feature XYZ",
	}

	testCases := []struct {
		maxWidth      int
		shouldContain string
	}{
		{maxWidth: 17, shouldContain: "Implement feat"},
		{maxWidth: 12, shouldContain: "Implement"},
		{maxWidth: 7, shouldContain: "Impl"},
	}

	for _, tc := range testCases {
		result := renderTaskSummaryTitle(task, bg, tc.maxWidth)
		assert.Contains(t, result, tc.shouldContain, "maxWidth=%d should show partial '%s'", tc.maxWidth, tc.shouldContain)
		assert.Contains(t, result, "...", "maxWidth=%d should show ellipsis", tc.maxWidth)
	}
}

func TestRenderTaskSummaryTitle_NeverExceedsWidth(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	task := &models.TaskSummary{
		Title: "This is a very long task title that needs to be truncated at various widths to ensure it never exceeds the card width",
	}

	widths := []int{7, 12, 17, 22, 27, 32, 37, 47}
	for _, maxWidth := range widths {
		result := renderTaskSummaryTitle(task, bg, maxWidth)
		line := strings.TrimPrefix(result, " ")
		renderedWidth := lipgloss.Width(line)
		assert.LessOrEqual(t, renderedWidth, maxWidth,
			"title line width %d exceeded max %d",
			renderedWidth, maxWidth)
	}
}

func TestRenderTaskSummaryTitle_EmptyTitle(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	task := &models.TaskSummary{
		Title: "",
	}

	result := renderTaskSummaryTitle(task, bg, 27)

	assert.NotContains(t, result, "...")
}
