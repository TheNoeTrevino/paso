package task

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// UpdateInput represents the parsed input for updating a task
type UpdateInput struct {
	TaskID      int
	Title       *string
	Description *string
	Priority    *string
}

// UpdateResult represents the output of a successful update operation
type UpdateResult struct {
	TaskID int
}

// ParseUpdateID parses the task ID from command arguments
func ParseUpdateID(args []string) (int, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("task ID is required")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid ID '%s': must be a number", args[0])
	}

	return taskID, nil
}

// ParseUpdateFlags extracts update fields from command flags
func ParseUpdateFlags(cmd *cobra.Command) (*UpdateInput, error) {
	taskTitle, _ := cmd.Flags().GetString("title")
	taskDescription, _ := cmd.Flags().GetString("description")
	taskPriority, _ := cmd.Flags().GetString("priority")

	titleFlag := cmd.Flags().Lookup("title")
	descFlag := cmd.Flags().Lookup("description")
	priorityFlag := cmd.Flags().Lookup("priority")

	// At least one update field must be provided
	if !titleFlag.Changed && !descFlag.Changed && !priorityFlag.Changed {
		return nil, fmt.Errorf("at least one of --title, --description, or --priority must be specified")
	}

	input := &UpdateInput{}

	if titleFlag.Changed {
		input.Title = &taskTitle
	}
	if descFlag.Changed {
		input.Description = &taskDescription
	}
	if priorityFlag.Changed {
		input.Priority = &taskPriority
	}

	return input, nil
}

// HasContentUpdate checks if title or description updates are present
func HasContentUpdate(input *UpdateInput) bool {
	return input.Title != nil || input.Description != nil
}

// HasPriorityUpdate checks if priority update is present
func HasPriorityUpdate(input *UpdateInput) bool {
	return input.Priority != nil
}

// FormatUpdateOutput generates the human-readable output message
func FormatUpdateOutput(result *UpdateResult) string {
	return fmt.Sprintf("Task %d updated successfully", result.TaskID)
}

// FormatUpdateJSON generates the JSON output structure
func FormatUpdateJSON(result *UpdateResult) map[string]any {
	return map[string]any{
		"success": true,
		"task_id": result.TaskID,
	}
}
