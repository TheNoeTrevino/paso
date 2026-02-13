package task

import (
	"fmt"
	"strconv"
)

// ReadyMoveInput represents the parsed input for moving a task to ready
type ReadyMoveInput struct {
	TaskID int
}

// ReadyMoveResult represents the output of a successful ready move operation
type ReadyMoveResult struct {
	TaskID     int
	FromColumn string
	ToColumn   string
}

// ParseReadyMoveArgs parses task ID from command arguments
func ParseReadyMoveArgs(args []string) (*ReadyMoveInput, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("task ID is required")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid task ID: %s", args[0])
	}

	return &ReadyMoveInput{
		TaskID: taskID,
	}, nil
}

// FormatReadyMoveOutput generates the human-readable output message
func FormatReadyMoveOutput(result *ReadyMoveResult) string {
	return fmt.Sprintf("Task %d moved to '%s'", result.TaskID, result.ToColumn)
}

// FormatReadyMoveJSON generates the JSON output structure
func FormatReadyMoveJSON(result *ReadyMoveResult) map[string]any {
	return FormatMoveResultJSON(result.TaskID, result.FromColumn, result.ToColumn)
}

// FormatReadyMoveQuiet generates the quiet output (just the ID)
func FormatReadyMoveQuiet(result *ReadyMoveResult) string {
	return fmt.Sprintf("%d\n", result.TaskID)
}
