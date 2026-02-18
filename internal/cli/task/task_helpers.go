package task

import (
	"strings"

	"charm.land/lipgloss/v2"
)

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

// WriteIndentedLines writes text to a builder with each line indented and styled
func WriteIndentedLines(builder *strings.Builder, text string, indent string, style lipgloss.Style) {
	for line := range strings.SplitSeq(text, "\n") {
		builder.WriteString(indent + style.Render(line) + "\n")
	}
}
