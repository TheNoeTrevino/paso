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
	result := renderTaskCardLabels(nil, bg, 30)

	assert.Contains(t, result, "no labels")
}

func TestRenderTaskCardLabels_AllFit(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	labels := []*models.Label{
		{Name: "bug", Color: "#FF0000"},
		{Name: "fix", Color: "#00FF00"},
	}

	result := renderTaskCardLabels(labels, bg, 40)

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

	// Use a narrow width that can't fit all labels
	result := renderTaskCardLabels(labels, bg, 20)

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

	widths := []int{10, 15, 20, 25, 30, 35, 40}
	for _, cardWidth := range widths {
		result := renderTaskCardLabels(labels, bg, cardWidth)
		// Strip the leading newline+space that the function prepends
		line := strings.TrimPrefix(result, "\n ")
		renderedWidth := lipgloss.Width(line)
		maxAllowed := cardWidth - 3 // safetyBuffer = 3
		assert.LessOrEqual(t, renderedWidth, maxAllowed,
			"label line width %d exceeded max %d at cardWidth=%d",
			renderedWidth, maxAllowed, cardWidth)
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

	widths := []int{15, 20, 25, 30, 35, 40, 50}
	for _, cardWidth := range widths {
		result := renderTaskSummaryMetadata(task, bg, cardWidth)
		line := strings.TrimPrefix(result, "\n ")
		renderedWidth := lipgloss.Width(line)
		maxAllowed := cardWidth - 3 // safetyBuffer = 3
		assert.LessOrEqual(t, renderedWidth, maxAllowed,
			"metadata line width %d exceeded max %d at cardWidth=%d",
			renderedWidth, maxAllowed, cardWidth)
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

	result := renderTaskSummaryMetadata(task, bg, 25)

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

	// 80 char wide card should fit everything
	result := renderTaskSummaryMetadata(task, bg, 80)

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

	// Width that allows partial "help wanted" to show
	// "bug duplicate enhancement help wa..." needs to be truncated
	// Full: "bug duplicate enhancement help wanted" = 3+1+9+1+11+1+11 = 37
	// Available after safetyBuffer (3): cardWidth - 3 = maxWidth
	// Set cardWidth=33, maxWidth=30
	// "bug duplicate enhancement help..." = 3+1+9+1+11+1+4+3 = 33 (doesn't fit in 30)
	// "bug duplicate enhancement hel..." = 3+1+9+1+11+1+3+3 = 32 (doesn't fit in 30)
	// "bug duplicate enhancement he..." = 3+1+9+1+11+1+2+3 = 31 (doesn't fit in 30)
	// "bug duplicate enhancement h..." = 3+1+9+1+11+1+1+3 = 30 (fits!)
	result := renderTaskCardLabels(labels, bg, 33)

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

	// Width that allows partial assignee
	// "task ∙ critical ∙ @noet" = 4+3+8+3+5 = 23
	// With ellipsis: 23+3 = 26
	result := renderTaskSummaryMetadata(task, bg, 29)

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

	// Test various widths to ensure character-level truncation works
	// Full: "bug enhancement" = 3+1+11 = 15, maxWidth after buffer = cardWidth-3
	// So cardWidth=18 gives maxWidth=15, which fits exactly
	testCases := []struct {
		width                int
		shouldContainBug     bool
		shouldContainPartial bool
		partialText          string
	}{
		// cardWidth=16, maxWidth=13: "bug enhanc..." = 3+1+6+3 = 13
		{width: 16, shouldContainBug: true, shouldContainPartial: true, partialText: "enhanc"},
		// cardWidth=14, maxWidth=11: "bug enha..." = 3+1+4+3 = 11
		{width: 14, shouldContainBug: true, shouldContainPartial: true, partialText: "enha"},
		// cardWidth=11, maxWidth=8: "bug en..." = 3+1+2+3 = 9 (doesn't fit), "bug e..." = 3+1+1+3 = 8
		{width: 11, shouldContainBug: true, shouldContainPartial: true, partialText: "e"},
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

	// Test various widths
	testCases := []struct {
		width             int
		shouldContainTask bool
		shouldContainCrit bool
	}{
		// Wide enough for task + critical + partial assignee
		{width: 35, shouldContainTask: true, shouldContainCrit: true},
		// Only task + critical
		{width: 25, shouldContainTask: true, shouldContainCrit: true},
	}

	for _, tc := range testCases {
		result := renderTaskSummaryMetadata(task, bg, tc.width)
		if tc.shouldContainTask {
			assert.Contains(t, result, "task", "width=%d should show 'task'", tc.width)
		}
		if tc.shouldContainCrit {
			assert.Contains(t, result, "critical", "width=%d should show 'critical'", tc.width)
		}
		assert.Contains(t, result, "...", "width=%d should show ellipsis", tc.width)
	}
}

func TestRenderTaskSummaryTitle_NoTruncation(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	task := &models.TaskSummary{
		Title: "Short task",
	}

	result := renderTaskSummaryTitle(task, bg, 50)

	assert.Contains(t, result, "Short task")
	assert.NotContains(t, result, "...")
}

func TestRenderTaskSummaryTitle_Truncation(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	task := &models.TaskSummary{
		Title: "This is a very long task title that should be truncated",
	}

	result := renderTaskSummaryTitle(task, bg, 20)

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
		width         int
		shouldContain string
	}{
		{width: 20, shouldContain: "Implement fe"},
		{width: 15, shouldContain: "Implement"},
		{width: 10, shouldContain: "Impl"},
	}

	for _, tc := range testCases {
		result := renderTaskSummaryTitle(task, bg, tc.width)
		assert.Contains(t, result, tc.shouldContain, "width=%d should show partial '%s'", tc.width, tc.shouldContain)
		assert.Contains(t, result, "...", "width=%d should show ellipsis", tc.width)
	}
}

func TestRenderTaskSummaryTitle_NeverExceedsWidth(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	task := &models.TaskSummary{
		Title: "This is a very long task title that needs to be truncated at various widths to ensure it never exceeds the card width",
	}

	widths := []int{10, 15, 20, 25, 30, 35, 40, 50}
	for _, cardWidth := range widths {
		result := renderTaskSummaryTitle(task, bg, cardWidth)
		line := strings.TrimPrefix(result, " ")
		renderedWidth := lipgloss.Width(line)
		maxAllowed := cardWidth - 3
		assert.LessOrEqual(t, renderedWidth, maxAllowed,
			"title line width %d exceeded max %d at cardWidth=%d",
			renderedWidth, maxAllowed, cardWidth)
	}
}

func TestRenderTaskSummaryTitle_EmptyTitle(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	task := &models.TaskSummary{
		Title: "",
	}

	result := renderTaskSummaryTitle(task, bg, 30)

	assert.NotContains(t, result, "...")
}
