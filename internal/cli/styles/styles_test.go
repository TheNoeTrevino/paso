package styles

import (
	"testing"

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

func TestTreeConstants(t *testing.T) {
	t.Parallel()
	assert.NotEmpty(t, TreeBranch)
	assert.NotEmpty(t, TreeLastBranch)
	assert.NotEmpty(t, TreeVertical)
	assert.NotEmpty(t, TreeSpace)
	assert.Greater(t, CompletedDimIntensity, 0.0)
	assert.LessOrEqual(t, CompletedDimIntensity, 1.0)
}
