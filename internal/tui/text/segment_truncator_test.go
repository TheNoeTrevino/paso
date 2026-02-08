package text

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestTruncateSegments_EmptySlice(t *testing.T) {
	t.Parallel()
	result := TruncateSegments(nil, " ", 40, "#000000")
	assert.Equal(t, "", result)
}

func TestTruncateSegments_AllFit(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "aaa", Render: func(t string) string { return t }},
		{Text: "bbb", Render: func(t string) string { return t }},
		{Text: "ccc", Render: func(t string) string { return t }},
	}
	separator := " "
	// "aaa bbb ccc" = 11 chars
	result := TruncateSegments(segments, separator, 20, "#000000")

	assert.Contains(t, result, "aaa")
	assert.Contains(t, result, "bbb")
	assert.Contains(t, result, "ccc")
	assert.NotContains(t, result, "...")
}

func TestTruncateSegments_Truncation(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "aaa", Render: func(t string) string { return t }},
		{Text: "bbb", Render: func(t string) string { return t }},
		{Text: "ccc", Render: func(t string) string { return t }},
		{Text: "ddd", Render: func(t string) string { return t }},
		{Text: "eee", Render: func(t string) string { return t }},
	}
	separator := " "
	// Each segment is 3 wide, separator is 1.
	// "aaa bbb ccc..." = 3+1+3+1+3+3 = 14
	// maxWidth = 14 should fit 3 segments + ellipsis
	result := TruncateSegments(segments, separator, 14, "#000000")

	assert.Contains(t, result, "aaa")
	assert.Contains(t, result, "bbb")
	assert.Contains(t, result, "ccc")
	assert.Contains(t, result, "...")
	assert.NotContains(t, result, "ddd")
	assert.NotContains(t, result, "eee")
}

func TestTruncateSegments_OnlyFirstFits(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "aaa", Render: func(t string) string { return t }},
		{Text: "bbb", Render: func(t string) string { return t }},
		{Text: "ccc", Render: func(t string) string { return t }},
	}
	separator := " "
	// "aaa..." = 3+3 = 6
	result := TruncateSegments(segments, separator, 6, "#000000")

	assert.Contains(t, result, "aaa")
	assert.Contains(t, result, "...")
	assert.NotContains(t, result, "bbb")
}

func TestTruncateSegments_NoneCanFit(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "aaaaaaaaaa", Render: func(t string) string { return t }},
		{Text: "bbb", Render: func(t string) string { return t }},
	}
	separator := " "
	// First segment alone (10) + ellipsis (3) = 13, exceeds maxWidth of 5
	result := TruncateSegments(segments, separator, 5, "#000000")

	assert.Contains(t, result, "...")
	assert.NotContains(t, result, "aaa")
	assert.NotContains(t, result, "bbb")
}

func TestTruncateSegments_SingleSegmentFits(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "hello", Render: func(t string) string { return t }},
	}
	result := TruncateSegments(segments, " ", 20, "#000000")

	assert.Equal(t, "hello", result)
	assert.NotContains(t, result, "...")
}

func TestTruncateSegments_SingleSegmentTooWide(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "hello world this is long", Render: func(t string) string { return t }},
	}
	result := TruncateSegments(segments, " ", 5, "#000000")

	assert.Contains(t, result, "...")
	// With character-level truncation, we should now see partial content
	assert.Contains(t, result, "he")
}

func TestTruncateSegments_ExactFit(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "aa", Render: func(t string) string { return t }},
		{Text: "bb", Render: func(t string) string { return t }},
	}
	separator := " "
	// "aa bb" = 2+1+2 = 5
	result := TruncateSegments(segments, separator, 5, "#000000")

	assert.Contains(t, result, "aa")
	assert.Contains(t, result, "bb")
	assert.NotContains(t, result, "...")
}

func TestTruncateSegments_WideSeparator(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "aa", Render: func(t string) string { return t }},
		{Text: "bb", Render: func(t string) string { return t }},
		{Text: "cc", Render: func(t string) string { return t }},
	}
	separator := " | "
	// "aa | bb | cc" = 2+3+2+3+2 = 12
	// maxWidth 11 means third segment won't fit completely
	// Now with character-level truncation: "aa | bb | c..." = 2+3+2+3+1+3 = 14, doesn't fit
	// "aa | bb..." = 2+3+2+3 = 10, fits in 11
	result := TruncateSegments(segments, separator, 11, "#000000")

	assert.Contains(t, result, "aa")
	assert.Contains(t, result, "bb")
	assert.Contains(t, result, "...")
}

func TestTruncateSegments_StyledSegments(t *testing.T) {
	t.Parallel()

	bg := "#333333"
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Background(lipgloss.Color(bg))

	segments := []Segment{
		{Text: "task", Render: func(t string) string { return style.Render(t) }},
		{Text: "critical", Render: func(t string) string { return style.Render(t) }},
		{Text: "@averylongusernamethatisverylong", Render: func(t string) string { return style.Render(t) }},
	}
	separator := " | "

	// Use a width that fits the first two segments and part of the third
	// "task" = 4, " | " = 3, "critical" = 8, " | " = 3, partial username
	maxWidth := 4 + 3 + 8 + 3 + 5 + 3 // enough for "@aver..."

	result := TruncateSegments(segments, separator, maxWidth, bg)

	assert.Contains(t, result, "task")
	assert.Contains(t, result, "critical")
	assert.Contains(t, result, "...")
	// Should show partial truncation of the username
	assert.Contains(t, result, "@aver")
	assert.NotContains(t, result, "@averylongusernamethatisverylong")
}

func TestTruncateSegments_PartialLastSegment(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "bug", Render: func(t string) string { return t }},
		{Text: "duplicate", Render: func(t string) string { return t }},
		{Text: "enhancement", Render: func(t string) string { return t }},
		{Text: "help wanted", Render: func(t string) string { return t }},
	}
	separator := " "

	// "bug duplicate enhancement h..." = 3+1+9+1+11+1+1+3 = 30
	// Set maxWidth to show partial last segment
	maxWidth := 30

	result := TruncateSegments(segments, separator, maxWidth, "#333333")

	assert.Contains(t, result, "bug")
	assert.Contains(t, result, "duplicate")
	assert.Contains(t, result, "enhancement")
	// Only 1 character of "help wanted" fits before ellipsis
	assert.Contains(t, result, "h")
	assert.Contains(t, result, "...")
	assert.NotContains(t, result, "help wanted")
}

func TestTruncateSegments_MetadataPartialAssignee(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "task", Render: func(t string) string { return t }},
		{Text: "critical", Render: func(t string) string { return t }},
		{Text: "@noetrevino", Render: func(t string) string { return t }},
	}
	separator := " ∙ "

	// "task ∙ critical ∙ @noet..." should fit
	maxWidth := 28

	result := TruncateSegments(segments, separator, maxWidth, "#333333")

	assert.Contains(t, result, "task")
	assert.Contains(t, result, "critical")
	assert.Contains(t, result, "@noet")
	assert.Contains(t, result, "...")
	assert.NotContains(t, result, "@noetrevino")
}

func TestTruncateSegments_MaxWidthEqualsEllipsisWidth(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "hello", Render: func(t string) string { return t }},
		{Text: "world", Render: func(t string) string { return t }},
	}

	result := TruncateSegments(segments, " ", 3, "#000000")

	assert.Contains(t, result, "...")
	assert.NotContains(t, result, "hello")
	assert.NotContains(t, result, "world")
}

func TestTruncateSegments_EmptyTextSegments(t *testing.T) {
	t.Parallel()

	t.Run("single empty segment", func(t *testing.T) {
		t.Parallel()
		segments := []Segment{
			{Text: "", Render: func(t string) string { return t }},
		}
		result := TruncateSegments(segments, " ", 20, "#000000")
		assert.NotContains(t, result, "...")
	})

	t.Run("multiple segments with some empty", func(t *testing.T) {
		t.Parallel()
		segments := []Segment{
			{Text: "aaa", Render: func(t string) string { return t }},
			{Text: "", Render: func(t string) string { return t }},
			{Text: "bbb", Render: func(t string) string { return t }},
		}
		result := TruncateSegments(segments, " ", 20, "#000000")
		assert.Contains(t, result, "aaa")
		assert.Contains(t, result, "bbb")
		assert.NotContains(t, result, "...")
	})

	t.Run("all segments empty", func(t *testing.T) {
		t.Parallel()
		segments := []Segment{
			{Text: "", Render: func(t string) string { return t }},
			{Text: "", Render: func(t string) string { return t }},
			{Text: "", Render: func(t string) string { return t }},
		}
		result := TruncateSegments(segments, " ", 10, "#000000")
		assert.NotContains(t, result, "...")
	})
}

func TestTruncateSegments_VerySmallMaxWidth(t *testing.T) {
	t.Parallel()

	segments := []Segment{
		{Text: "hello", Render: func(t string) string { return t }},
	}

	t.Run("maxWidth zero", func(t *testing.T) {
		t.Parallel()
		result := TruncateSegments(segments, " ", 0, "#000000")
		assert.Contains(t, result, "...")
	})

	t.Run("maxWidth negative", func(t *testing.T) {
		t.Parallel()
		result := TruncateSegments(segments, " ", -1, "#000000")
		assert.Contains(t, result, "...")
	})

	t.Run("maxWidth 1", func(t *testing.T) {
		t.Parallel()
		result := TruncateSegments(segments, " ", 1, "#000000")
		assert.Contains(t, result, "...")
	})

	t.Run("maxWidth 2", func(t *testing.T) {
		t.Parallel()
		result := TruncateSegments(segments, " ", 2, "#000000")
		assert.Contains(t, result, "...")
	})
}

func TestTruncateTextToWidth(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		text     string
		maxWidth int
		expected string
	}{
		{
			name:     "text fits within maxWidth",
			text:     "hello",
			maxWidth: 10,
			expected: "hello",
		},
		{
			name:     "text exactly equals maxWidth",
			text:     "hello",
			maxWidth: 5,
			expected: "hello",
		},
		{
			name:     "text truncated to maxWidth",
			text:     "hello world",
			maxWidth: 5,
			expected: "hello",
		},
		{
			name:     "empty string",
			text:     "",
			maxWidth: 10,
			expected: "",
		},
		{
			name:     "maxWidth zero",
			text:     "hello",
			maxWidth: 0,
			expected: "",
		},
		{
			name:     "maxWidth negative",
			text:     "hello",
			maxWidth: -1,
			expected: "",
		},
		{
			name:     "maxWidth one",
			text:     "hello",
			maxWidth: 1,
			expected: "h",
		},
		{
			name:     "CJK double-width characters fit exactly",
			text:     "\u4e16\u754c",
			maxWidth: 4,
			expected: "\u4e16\u754c",
		},
		{
			name:     "CJK double-width characters truncated",
			text:     "\u4e16\u754c\u4f60\u597d",
			maxWidth: 5,
			expected: "\u4e16\u754c",
		},
		{
			name:     "CJK boundary odd maxWidth",
			text:     "\u4e16\u754c\u4f60",
			maxWidth: 3,
			expected: "\u4e16",
		},
		{
			name:     "mixed ASCII and CJK",
			text:     "ab\u4e16cd",
			maxWidth: 4,
			expected: "ab\u4e16",
		},
		{
			name:     "mixed ASCII and CJK truncated at CJK boundary",
			text:     "ab\u4e16cd",
			maxWidth: 3,
			expected: "ab",
		},
		{
			name:     "single CJK char does not fit in width 1",
			text:     "\u4e16",
			maxWidth: 1,
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := TruncateTextToWidth(tc.text, tc.maxWidth)
			assert.Equal(t, tc.expected, result)
		})
	}
}
