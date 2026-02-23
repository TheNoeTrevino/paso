package task

import (
	"fmt"
	"strconv"
)

// ArchiveInput represents the parsed input for archiving a task
type ArchiveInput struct {
	TaskID int
}

// ArchiveResult represents the output of a successful archive operation
type ArchiveResult struct {
	TaskID      int
	WasArchived bool
}

// ParseArchiveArgs parses task ID from command arguments
func ParseArchiveArgs(args []string) (*ArchiveInput, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("task ID is required")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid task ID: %s", args[0])
	}

	if taskID <= 0 {
		return nil, fmt.Errorf("task ID must be a positive integer")
	}

	return &ArchiveInput{
		TaskID: taskID,
	}, nil
}

// FormatArchiveOutput generates the human-readable output message
func FormatArchiveOutput(result *ArchiveResult) string {
	if result.WasArchived {
		return fmt.Sprintf("Task %d archived", result.TaskID)
	}
	return fmt.Sprintf("Task %d unarchived", result.TaskID)
}

// FormatArchiveJSON generates the JSON output structure
func FormatArchiveJSON(result *ArchiveResult) map[string]any {
	return map[string]any{
		"task_id":  result.TaskID,
		"archived": result.WasArchived,
	}
}
