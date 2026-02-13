package task

import (
	"fmt"

	"github.com/thenoetrevino/paso/internal/models"
)

const (
	unknownColumnName = "Unknown"
	moveTargetNext    = "next"
	moveTargetPrev    = "prev"
)

// findNextColumnName finds the name of the next column in the linked list
func findNextColumnName(columns []*models.Column, currentColumnID int) string {
	for _, col := range columns {
		if col.ID == currentColumnID && col.NextID != nil {
			for _, nextCol := range columns {
				if nextCol.ID == *col.NextID {
					return nextCol.Name
				}
			}
		}
	}
	return unknownColumnName
}

// findPrevColumnName finds the name of the previous column in the linked list
func findPrevColumnName(columns []*models.Column, currentColumnID int) string {
	for _, col := range columns {
		if col.ID == currentColumnID && col.PrevID != nil {
			for _, prevCol := range columns {
				if prevCol.ID == *col.PrevID {
					return prevCol.Name
				}
			}
		}
	}
	return unknownColumnName
}

// parseMoveTarget determines the type of move target (next, prev, or column name)
// and returns the target type and normalized target string
func parseMoveTarget(target string) (targetType string, normalized string) {
	switch target {
	case "next", "Next", "NEXT":
		return moveTargetNext, moveTargetNext
	case "prev", "Prev", "PREV", "previous", "Previous", "PREVIOUS":
		return moveTargetPrev, moveTargetPrev
	default:
		return "column", target
	}
}

// validateMoveTarget checks if a move is valid and returns an error message if not
// Returns empty string if valid
func validateMoveTarget(targetType, toColumnName, currentColumnName string) string {
	if toColumnName != unknownColumnName {
		return ""
	}

	switch targetType {
	case moveTargetNext:
		return fmt.Sprintf("task is already in the last column (%s)", currentColumnName)
	case moveTargetPrev:
		return fmt.Sprintf("task is already in the first column (%s)", currentColumnName)
	}
	return ""
}

// formatMoveMessage generates the human-readable message for a move operation
func formatMoveMessage(taskID int, fromColumn, toColumn string, dryRun bool) string {
	isSameColumn := fromColumn == toColumn

	if dryRun {
		if isSameColumn {
			return fmt.Sprintf("Would keep task %d in '%s' (already there)", taskID, toColumn)
		}
		return fmt.Sprintf("Would move task %d from '%s' to '%s'", taskID, fromColumn, toColumn)
	}

	if isSameColumn {
		return fmt.Sprintf("Task %d is already in '%s'", taskID, toColumn)
	}
	return fmt.Sprintf("Task %d moved to '%s'", taskID, toColumn)
}

// MoveResult represents the result of a move operation
type MoveResult struct {
	Success    bool   `json:"success"`
	TaskID     int    `json:"task_id"`
	FromColumn string `json:"from_column"`
	ToColumn   string `json:"to_column"`
	DryRun     bool   `json:"dry_run,omitempty"`
}

// createMoveResult creates a MoveResult struct
func createMoveResult(taskID int, fromColumn, toColumn string, dryRun bool) MoveResult {
	return MoveResult{
		Success:    true,
		TaskID:     taskID,
		FromColumn: fromColumn,
		ToColumn:   toColumn,
		DryRun:     dryRun,
	}
}
