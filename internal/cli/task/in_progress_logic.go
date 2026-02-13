package task

import (
	"fmt"
	"strconv"
	"strings"
)

// InProgressMode represents the operation mode (list or move)
type InProgressMode string

const (
	InProgressModeList InProgressMode = "list"
	InProgressModeMove InProgressMode = "move"
)

// InProgressInput represents the parsed input for the in-progress command
type InProgressInput struct {
	Mode      InProgressMode
	ProjectID int
	TaskID    int
}

// TaskDisplay represents a simplified task for display
type TaskDisplay struct {
	ID                  int    `json:"id"`
	TicketNumber        int    `json:"ticket_number"`
	Title               string `json:"title"`
	TypeDescription     string `json:"type_description"`
	PriorityDescription string `json:"priority_description"`
	PriorityColor       string `json:"priority_color"`
	IsBlocked           bool   `json:"is_blocked"`
}

// ListInProgressResult represents the output of listing in-progress tasks
type ListInProgressResult struct {
	Tasks []TaskDisplay
	Count int
}

// MoveInProgressResult represents the output of moving a task to in-progress
type MoveInProgressResult struct {
	TaskID     int
	FromColumn string
	ToColumn   string
}

// ParseInProgressArgs parses the command arguments and flags to determine the mode and inputs
func ParseInProgressArgs(args []string, projectID int) (*InProgressInput, error) {
	// List mode if projectID is provided
	if projectID > 0 {
		return &InProgressInput{
			Mode:      InProgressModeList,
			ProjectID: projectID,
		}, nil
	}

	// Move mode - require task ID
	if len(args) == 0 {
		return nil, fmt.Errorf("task ID is required (or use --project flag to list tasks)")
	}

	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid task ID '%s': must be a number", args[0])
	}

	if taskID <= 0 {
		return nil, fmt.Errorf("task ID must be a positive integer")
	}

	return &InProgressInput{
		Mode:   InProgressModeMove,
		TaskID: taskID,
	}, nil
}

// FormatListQuiet generates quiet output for listing (IDs only)
func FormatListQuiet(result *ListInProgressResult) string {
	var output strings.Builder
	for _, task := range result.Tasks {
		output.WriteString(fmt.Sprintf("%d\n", task.ID))
	}
	return output.String()
}

// FormatListJSON generates JSON output for listing
func FormatListJSON(result *ListInProgressResult) map[string]any {
	return map[string]any{
		"success": true,
		"tasks":   result.Tasks,
		"count":   result.Count,
	}
}

// FormatListHuman generates human-readable output for listing
func FormatListHuman(result *ListInProgressResult) string {
	if result.Count == 0 {
		return "No in-progress tasks found"
	}

	output := fmt.Sprintf("Found %d in-progress tasks:\n\n", result.Count)
	for _, task := range result.Tasks {
		priorityInfo := ""
		if ShouldDisplayPriority(task.PriorityDescription) {
			priorityInfo = fmt.Sprintf(" [%s]", task.PriorityDescription)
		}

		blockedInfo := ""
		if task.IsBlocked {
			blockedInfo = " ▲ BLOCKED"
		}

		output += fmt.Sprintf("  [%d] %s%s%s\n", task.ID, task.Title, priorityInfo, blockedInfo)
	}

	return output
}

// FormatMoveQuiet generates quiet output for moving (task ID only)
func FormatMoveQuiet(result *MoveInProgressResult) string {
	return fmt.Sprintf("%d\n", result.TaskID)
}

// FormatMoveJSON generates JSON output for moving
func FormatMoveJSON(result *MoveInProgressResult) map[string]any {
	return FormatMoveResultJSON(result.TaskID, result.FromColumn, result.ToColumn)
}

// FormatMoveHuman generates human-readable output for moving
func FormatMoveHuman(result *MoveInProgressResult) string {
	return fmt.Sprintf("Task %d moved to '%s'", result.TaskID, result.ToColumn)
}
