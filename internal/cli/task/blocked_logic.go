package task

import (
	"fmt"
	"strings"

	"github.com/thenoetrevino/paso/internal/models"
)

// BlockedResult represents the output of a successful blocked operation
type BlockedResult struct {
	Tasks []*models.TaskSummary
	Count int
}

// FilterBlockedTasks filters a task map to return only blocked tasks
func FilterBlockedTasks(tasksByColumn map[int][]*models.TaskSummary) []*models.TaskSummary {
	var blockedTasks []*models.TaskSummary
	for _, columnTasks := range tasksByColumn {
		for _, task := range columnTasks {
			if task.IsBlocked {
				blockedTasks = append(blockedTasks, task)
			}
		}
	}
	return blockedTasks
}

// FormatBlockedOutput generates the human-readable output message
func FormatBlockedOutput(result *BlockedResult) string {
	if result.Count == 0 {
		return "No blocked tasks found"
	}

	var output strings.Builder
	fmt.Fprintf(&output, "Found %d blocked tasks:\n\n", result.Count)
	for _, t := range result.Tasks {
		priorityInfo := ""
		if ShouldDisplayPriority(t.PriorityDescription) {
			priorityInfo = fmt.Sprintf(" [%s]", t.PriorityDescription)
		}
		fmt.Fprintf(&output, "  [%d] %s%s (BLOCKED)\n", t.ID, t.Title, priorityInfo)
	}
	return output.String()
}

// FormatBlockedJSON generates the JSON output structure
func FormatBlockedJSON(result *BlockedResult) map[string]any {
	return map[string]any{
		"success": true,
		"tasks":   result.Tasks,
		"count":   result.Count,
	}
}

// FormatBlockedQuiet generates the quiet output (IDs only)
func FormatBlockedQuiet(result *BlockedResult) string {
	var output strings.Builder
	for _, t := range result.Tasks {
		fmt.Fprintf(&output, "%d\n", t.ID)
	}
	return output.String()
}
