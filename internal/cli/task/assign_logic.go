package task

import (
	"fmt"
	"strconv"
)

// AssignInput represents the parsed input for assigning a task
type AssignInput struct {
	TaskID       int
	AssigneeName string
	Clear        bool
}

// AssignResult represents the output of a successful assign operation
type AssignResult struct {
	TaskID       int
	AssigneeName string
	Cleared      bool
}

// ParseAssignArgs parses task ID and assignee name from command arguments
func ParseAssignArgs(args []string) (*AssignInput, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("task ID is required")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid task ID: %s", args[0])
	}

	assigneeName := ""
	if len(args) > 1 {
		assigneeName = args[1]
	}

	return &AssignInput{
		TaskID:       taskID,
		AssigneeName: assigneeName,
	}, nil
}

// FormatAssignOutput generates the human-readable output message
func FormatAssignOutput(result *AssignResult) string {
	if result.Cleared {
		return fmt.Sprintf("Task %d assignee cleared", result.TaskID)
	}
	return fmt.Sprintf("Task %d assigned to @%s", result.TaskID, result.AssigneeName)
}

// FormatAssignJSON generates the JSON output structure
func FormatAssignJSON(result *AssignResult) map[string]any {
	output := map[string]any{
		"success": true,
		"task_id": result.TaskID,
	}
	if result.Cleared {
		output["assignee"] = nil
	} else {
		output["assignee"] = result.AssigneeName
	}
	return output
}
