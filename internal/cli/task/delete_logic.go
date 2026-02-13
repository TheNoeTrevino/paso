package task

import (
	"fmt"
	"strconv"
	"strings"
)

// DeleteInput represents the parsed input for deleting a task
type DeleteInput struct {
	TaskID int
}

// DeleteResult represents the output of a successful delete operation
type DeleteResult struct {
	TaskID int
}

// ParseDeleteArgs parses task ID from command arguments
func ParseDeleteArgs(args []string) (*DeleteInput, error) {
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

	return &DeleteInput{
		TaskID: taskID,
	}, nil
}

// FormatDeleteOutput generates the human-readable output message
func FormatDeleteOutput(result *DeleteResult) string {
	return fmt.Sprintf("Task %d deleted successfully", result.TaskID)
}

// FormatDeleteJSON generates the JSON output structure
func FormatDeleteJSON(result *DeleteResult) map[string]any {
	return map[string]any{
		"success": true,
		"task_id": result.TaskID,
	}
}

// FormatConfirmationPrompt generates the confirmation prompt message
func FormatConfirmationPrompt(taskID int, taskTitle string) string {
	return fmt.Sprintf("Delete task #%d: '%s'? (y/N): ", taskID, taskTitle)
}

// IsConfirmationYes checks if the response is a confirmation (y/yes)
func IsConfirmationYes(response string) bool {
	normalized := strings.ToLower(strings.TrimSpace(response))
	return normalized == "y" || normalized == "yes"
}
