package styles

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

func TestInit(t *testing.T) {
	t.Parallel()
	scheme := colors.ColorScheme{
		Title:     "#FFFFFF",
		Accent:    "#00FF00",
		Normal:    "#CCCCCC",
		Subtle:    "#666666",
		ErrorFg:   "#FF0000",
		ErrorBg:   "#330000",
		WarningFg: "#FFFF00",
		WarningBg: "#333300",
		InfoFg:    "#00FFFF",
		InfoBg:    "#003333",
	}

	Init(scheme)

	// Verify styles were set (non-zero width means initialized)
	assert.Equal(t, 80, CardWidth)
}

func TestColoredText(t *testing.T) {
	t.Parallel()
	result := ColoredText("hello", "#FF0000")
	assert.Contains(t, result, "hello")
}

func TestBoldColoredText(t *testing.T) {
	t.Parallel()
	result := BoldColoredText("bold text", "#00FF00")
	assert.Contains(t, result, "bold text")
}

func TestRenderLabelChip(t *testing.T) {
	t.Parallel()
	label := &models.Label{
		ID:    1,
		Name:  "bug",
		Color: "#FF0000",
	}

	result := RenderLabelChip(label)
	assert.Contains(t, result, "[bug]")
}

func TestRenderTaskReference(t *testing.T) {
	t.Parallel()
	ref := &models.TaskReference{
		ID:            1,
		TaskNumber:    42,
		Title:         "Fix the thing",
		ProjectName:   "MyProject",
		RelationColor: "#0000FF",
	}

	result := RenderTaskReference(ref)
	assert.Contains(t, result, "MyProject-42")
	assert.Contains(t, result, "Fix the thing")
}

func TestRenderTaskReferenceWithLabel(t *testing.T) {
	t.Parallel()
	ref := &models.TaskReference{
		ID:            1,
		TaskNumber:    42,
		Title:         "Fix the thing",
		ProjectName:   "MyProject",
		RelationLabel: "blocks",
		RelationColor: "#FF0000",
	}

	result := RenderTaskReferenceWithLabel(ref)
	assert.Contains(t, result, "MyProject-42")
	assert.Contains(t, result, "blocks")
	assert.Contains(t, result, "Fix the thing")
}

func TestRenderCard(t *testing.T) {
	t.Parallel()
	// Init styles first so CardStyle is set
	scheme := colors.ColorScheme{
		Title:     "#FFFFFF",
		Accent:    "#00FF00",
		Normal:    "#CCCCCC",
		Subtle:    "#666666",
		ErrorFg:   "#FF0000",
		ErrorBg:   "#330000",
		WarningFg: "#FFFF00",
		WarningBg: "#333300",
		InfoFg:    "#00FFFF",
		InfoBg:    "#003333",
	}
	Init(scheme)

	result := RenderCard("card content")
	assert.Contains(t, result, "card content")
}

func TestCardContentWidth(t *testing.T) {
	t.Parallel()
	// CardWidth includes the border and horizontal padding on both sides, so
	// the usable text width is narrower than CardWidth itself.
	assert.Equal(t, CardWidth-(CardBorderX+CardPaddingX)*2, CardContentWidth())
	assert.Less(t, CardContentWidth(), CardWidth)
}

func TestWrapText(t *testing.T) {
	t.Parallel()

	t.Run("wraps on word boundaries within the limit", func(t *testing.T) {
		t.Parallel()
		lines := WrapText("the quick brown fox jumps", 10)
		for _, l := range lines {
			assert.LessOrEqual(t, len(l), 10, "line %q exceeds limit", l)
		}
		assert.Equal(t, "the quick brown fox jumps", strings.Join(lines, " "))
	})

	t.Run("hard-breaks words longer than the limit", func(t *testing.T) {
		t.Parallel()
		lines := WrapText("supercalifragilistic", 5)
		assert.Greater(t, len(lines), 1)
		for _, l := range lines {
			assert.LessOrEqual(t, len(l), 5, "line %q exceeds limit", l)
		}
	})

	t.Run("preserves existing newlines as hard breaks", func(t *testing.T) {
		t.Parallel()
		lines := WrapText("- one\n- two", 40)
		assert.Equal(t, []string{"- one", "- two"}, lines)
	})
}

func TestRenderTaskReferenceWrapsWithHangingIndent(t *testing.T) {
	t.Parallel()
	ref := &models.TaskReference{
		TaskNumber:    42,
		ProjectName:   "MyProject",
		Title:         strings.Repeat("word ", 40), // forces wrapping
		RelationColor: "#0000FF",
	}

	lines := strings.Split(stripANSI(RenderTaskReference(ref)), "\n")
	assert.Greater(t, len(lines), 1, "long title should wrap onto multiple lines")

	// First line carries the bullet at a two-space indent.
	assert.True(t, strings.HasPrefix(lines[0], "  • "), "got %q", lines[0])
	// Continuation lines hang under the bullet text at a four-space indent
	// (and never collapse back to the card margin).
	for _, l := range lines[1:] {
		assert.True(t, strings.HasPrefix(l, "    "), "continuation %q lost its hanging indent", l)
		assert.False(t, strings.HasPrefix(l, "    •"), "continuation %q should not repeat the bullet", l)
	}
	// Every line stays within the card content width. Measure visual columns
	// (not bytes), matching how the wrap library and CardContentWidth count.
	for _, l := range lines {
		assert.LessOrEqual(t, runewidth.StringWidth(l), CardContentWidth(), "line %q exceeds card width", l)
	}
}

func TestTreeConstants(t *testing.T) {
	t.Parallel()
	assert.NotEmpty(t, TreeBranch)
	assert.NotEmpty(t, TreeLastBranch)
	assert.NotEmpty(t, TreeVertical)
	assert.NotEmpty(t, TreeSpace)
	assert.Greater(t, CompletedDimIntensity, 0.0)
	assert.LessOrEqual(t, CompletedDimIntensity, 1.0)
}
