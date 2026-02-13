package task

import (
	"fmt"
	"strconv"
)

// DoneInput represents the parsed input for marking a task as done
type DoneInput struct {
	TaskID int
}

// DoneResult represents the output of a successful done operation
type DoneResult struct {
	TaskID     int
	FromColumn string
	ToColumn   string
}

// ParseDoneArgs parses task ID from command arguments
func ParseDoneArgs(args []string) (*DoneInput, error) {
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

	return &DoneInput{
		TaskID: taskID,
	}, nil
}

// FormatDoneOutput generates the human-readable output message
func FormatDoneOutput(result *DoneResult) string {
	return fmt.Sprintf("Task %d moved from %s to %s", result.TaskID, result.FromColumn, result.ToColumn)
}

// FormatDoneJSON generates the JSON output structure
func FormatDoneJSON(result *DoneResult) map[string]any {
	return FormatMoveResultJSON(result.TaskID, result.FromColumn, result.ToColumn)
}
