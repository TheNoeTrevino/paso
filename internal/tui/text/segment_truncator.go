package text

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// Segment represents a piece of content with plain text and a render function.
type Segment struct {
	Text   string
	Render func(text string) string
}

// NewStyledSegment creates a Segment with the given text, style, and background color.
func NewStyledSegment(text string, style lipgloss.Style, bg string) Segment {
	return Segment{
		Text: text,
		Render: func(t string) string {
			return style.Background(lipgloss.Color(bg)).Render(t)
		},
	}
}

// TruncateTextToWidth returns the longest prefix of text that fits within maxWidth,
// or empty string if nothing fits. Uses runewidth for accurate Unicode width measurement.
func TruncateTextToWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	w := 0
	for i, r := range text {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxWidth {
			return text[:i]
		}
		w += rw
	}
	return text
}

// segmentTruncator holds all the precomputed state needed to truncate
// a sequence of styled segments to fit within a maximum display width.
type segmentTruncator struct {
	segments        []Segment
	prefixWidth     []int
	maxWidth        int
	sepWidth        int
	ellipsisWidth   int
	styledSeparator string
	ellipsis        string
}

func newSegmentTruncator(segments []Segment, separator string, maxWidth int, bg string) segmentTruncator {
	const ellipsisText = "..."

	ellipsisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
	sepWidth := runewidth.StringWidth(separator)

	n := len(segments)
	prefixWidth := make([]int, n+1)
	for i, seg := range segments {
		prefixWidth[i+1] = prefixWidth[i] + runewidth.StringWidth(seg.Text)
		if i > 0 {
			prefixWidth[i+1] += sepWidth
		}
	}

	return segmentTruncator{
		segments:      segments,
		prefixWidth:   prefixWidth,
		maxWidth:      maxWidth,
		sepWidth:      sepWidth,
		ellipsisWidth: runewidth.StringWidth(ellipsisText),
		styledSeparator: lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(theme.Subtle)).
			Render(separator),
		ellipsis: ellipsisStyle.Background(lipgloss.Color(bg)).Render(ellipsisText),
	}
}

// renderJoined renders all segments and joins them with the styled separator.
func (t *segmentTruncator) renderJoined(segs []Segment) string {
	rendered := make([]string, len(segs))
	for i, seg := range segs {
		rendered[i] = seg.Render(seg.Text)
	}
	return strings.Join(rendered, t.styledSeparator)
}

// tryFitAll returns the fully rendered string if every segment fits within maxWidth.
func (t *segmentTruncator) tryFitAll() (string, bool) {
	totalWidth := t.prefixWidth[len(t.segments)]
	if totalWidth > t.maxWidth {
		return "", false
	}
	return t.renderJoined(t.segments), true
}

// tryFitWithTruncation walks backwards through segments to find how many complete
// segments fit, and whether a partial last segment can be shown.
func (t *segmentTruncator) tryFitWithTruncation() string {
	for i := len(t.segments) - 1; i >= 1; i-- {
		width := t.prefixWidth[i]

		remainingWidth := t.maxWidth - width - t.sepWidth - t.ellipsisWidth
		truncated := TruncateTextToWidth(t.segments[i].Text, remainingWidth)
		if truncated != "" {
			prefix := t.renderJoined(t.segments[:i])
			partial := t.segments[i].Render(truncated)
			if prefix != "" {
				return prefix + t.styledSeparator + partial + t.ellipsis
			}
			return partial + t.ellipsis
		}

		if width+t.ellipsisWidth <= t.maxWidth {
			return t.renderJoined(t.segments[:i]) + t.ellipsis
		}
	}

	return ""
}

// tryPartialFirst attempts to fit a truncated version of the first segment.
func (t *segmentTruncator) tryPartialFirst() string {
	truncated := TruncateTextToWidth(t.segments[0].Text, t.maxWidth-t.ellipsisWidth)
	if truncated != "" {
		return t.segments[0].Render(truncated) + t.ellipsis
	}
	return ""
}

// TruncateSegments joins segments with a separator, truncating the last
// visible segment at the character level if needed. Measures plain text widths
// and only renders with lipgloss after determining what will fit.
func TruncateSegments(segments []Segment, separator string, maxWidth int, bg string) string {
	if len(segments) == 0 {
		return ""
	}

	t := newSegmentTruncator(segments, separator, maxWidth, bg)

	if result, ok := t.tryFitAll(); ok {
		return result
	}
	if result := t.tryFitWithTruncation(); result != "" {
		return result
	}
	if result := t.tryPartialFirst(); result != "" {
		return result
	}

	return t.ellipsis
}
