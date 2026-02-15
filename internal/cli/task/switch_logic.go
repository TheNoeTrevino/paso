package task

import (
	"fmt"
	"strconv"
)

// SwitchInput represents the parsed input for moving a task to a project
type SwitchInput struct {
	TaskID    int
	ProjectID int
}

// SwitchResult represents the output of a successful switch operation
type SwitchResult struct {
	TaskID      int
	ProjectID   int
	ProjectName string
}

// ParseSwitchArgs parses task ID and project ID from command arguments
func ParseSwitchArgs(args []string) (*SwitchInput, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("task ID and project ID are required")
	}

	if len(args) > 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid task ID: %s", args[0])
	}

	if taskID <= 0 {
		return nil, fmt.Errorf("task ID must be a positive integer")
	}

	projectID, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %s", args[1])
	}

	if projectID <= 0 {
		return nil, fmt.Errorf("project ID must be a positive integer")
	}

	return &SwitchInput{
		TaskID:    taskID,
		ProjectID: projectID,
	}, nil
}

// FormatSwitchOutput generates the human-readable output message
func FormatSwitchOutput(result *SwitchResult) string {
	return fmt.Sprintf("Task %d successfully moved to project %s", result.TaskID, result.ProjectName)
}

// FormatSwitchJSON generates the JSON output structure
func FormatSwitchJSON(result *SwitchResult) map[string]any {
	return map[string]any{
		"task_id":      result.TaskID,
		"project_id":   result.ProjectID,
		"project_name": result.ProjectName,
	}
}
