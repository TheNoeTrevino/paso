package task

import (
	"fmt"

	"github.com/thenoetrevino/paso/internal/models"
)

// ReadyResult represents the result of a ready command
type ReadyResult struct {
	Tasks []*models.TaskSummary
	Count int
}

// FormatReadyOutput formats the ready result for human-readable output
func FormatReadyOutput(result *ReadyResult) string {
	if result.Count == 0 {
		return "No ready tasks found"
	}

	output := fmt.Sprintf("Found %d ready tasks:\n\n", result.Count)
	for _, t := range result.Tasks {
		priorityInfo := ""
		if t.PriorityDescription != "" && t.PriorityDescription != "medium" {
			priorityInfo = fmt.Sprintf(" [%s]", t.PriorityDescription)
		}
		output += fmt.Sprintf("  [%d] %s%s\n", t.ID, t.Title, priorityInfo)
	}

	return output
}

// FormatReadyJSON formats the ready result as a JSON-serializable map
func FormatReadyJSON(result *ReadyResult) map[string]any {
	return map[string]any{
		"success": true,
		"tasks":   result.Tasks,
		"count":   result.Count,
	}
}

// FormatReadyQuiet formats the ready result for quiet mode (IDs only)
func FormatReadyQuiet(result *ReadyResult) string {
	output := ""
	for _, t := range result.Tasks {
		output += fmt.Sprintf("%d\n", t.ID)
	}
	return output
}
