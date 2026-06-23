package task

import (
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/thenoetrevino/paso/internal/cli/styles"
)

// listMarkerRe matches a leading list marker so wrapped continuation lines can
// be hang-indented under the marker's text. Covers "- ", "* ", "• ", checkbox
// "[ ] "/"[x] ", and numbered "1. "/"1) " markers, with optional leading
// whitespace.
var listMarkerRe = regexp.MustCompile(`^(\s*(?:[-*•]|\[.?\]|\d+[.)])\s+)`)

// ShouldDisplayPriority returns true if priority should be shown in output
// (i.e., it's set and not the default "medium")
func ShouldDisplayPriority(priorityDesc string) bool {
	return priorityDesc != "" && priorityDesc != "medium"
}

// FormatMoveResultJSON creates a standardized JSON response for move operations
func FormatMoveResultJSON(taskID int, fromColumn, toColumn string) map[string]any {
	return map[string]any{
		"success":     true,
		"task_id":     taskID,
		"from_column": fromColumn,
		"to_column":   toColumn,
	}
}

// DisplayOrDefault returns the value pointed to by ptr if it's non-nil and non-empty,
// otherwise returns the default value
func DisplayOrDefault(ptr *string, defaultVal string) string {
	if ptr != nil && *ptr != "" {
		return *ptr
	}
	return defaultVal
}

// WriteIndentedLines writes text to a builder with each line indented and
// styled. Lines wider than the card content area are word-wrapped to the same
// indent so they don't get re-wrapped (and lose their indentation) when the
// card border is rendered.
func WriteIndentedLines(builder *strings.Builder, text string, indent string, style lipgloss.Style) {
	limit := styles.CardContentWidth() - len(indent)
	for src := range strings.SplitSeq(text, "\n") {
		// Hang-indent wrapped continuation lines under a list marker's text so a
		// bullet whose body wraps still reads as one item rather than several.
		hang := ""
		if m := listMarkerRe.FindString(src); m != "" {
			hang = strings.Repeat(" ", len(m))
		}
		// Reserve the hang width so wrapped lines fit once it's prepended.
		for i, line := range styles.WrapText(src, limit-len(hang)) {
			lineIndent := indent
			if i > 0 {
				lineIndent = indent + hang
			}
			builder.WriteString(lineIndent + style.Render(line) + "\n")
		}
	}
}
