package helpers

import (
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/commands"
)

// GetAdjacentTaskIDs calculates the task IDs adjacent to the currently selected task.
// Adjacent tasks are those directly above, below, left, and right in the kanban board.
//
// Parameters:
//   - columns: Ordered slice of columns in the board
//   - tasks: Map of column ID to tasks in that column (ordered by position)
//   - selectedCol: Index of the currently selected column in the columns slice
//   - selectedTask: Index of the currently selected task within the column
//
// Returns an AdjacentTasks struct with:
//   - Current: The ID of the selected task (0 if selection is invalid)
//   - Above: ID of task above in same column (0 if at top or column empty)
//   - Below: ID of task below in same column (0 if at bottom or column empty)
//   - Left: ID of first task in column to the left (0 if no left column or empty)
//   - Right: ID of first task in column to the right (0 if no right column or empty)
func GetAdjacentTaskIDs(
	columns []*models.Column,
	tasks map[int][]*models.TaskSummary,
	selectedCol int,
	selectedTask int,
) commands.AdjacentTasks {
	result := commands.AdjacentTasks{}

	// Validate column selection
	if len(columns) == 0 || selectedCol < 0 || selectedCol >= len(columns) {
		return result
	}

	currentColumn := columns[selectedCol]
	currentTasks := tasks[currentColumn.ID]

	// Validate task selection within current column
	if len(currentTasks) == 0 || selectedTask < 0 || selectedTask >= len(currentTasks) {
		return result
	}

	// Set current task
	result.Current = currentTasks[selectedTask].ID

	// Task above in same column
	if selectedTask > 0 {
		result.Above = currentTasks[selectedTask-1].ID
	}

	// Task below in same column
	if selectedTask < len(currentTasks)-1 {
		result.Below = currentTasks[selectedTask+1].ID
	}

	// Task in column to the left
	if selectedCol > 0 {
		leftColumn := columns[selectedCol-1]
		leftTasks := tasks[leftColumn.ID]
		if len(leftTasks) > 0 {
			// Use task at same index if available, otherwise first task
			idx := min(selectedTask, len(leftTasks)-1)
			result.Left = leftTasks[idx].ID
		}
	}

	// Task in column to the right
	if selectedCol < len(columns)-1 {
		rightColumn := columns[selectedCol+1]
		rightTasks := tasks[rightColumn.ID]
		if len(rightTasks) > 0 {
			// Use task at same index if available, otherwise first task
			idx := min(selectedTask, len(rightTasks)-1)
			result.Right = rightTasks[idx].ID
		}
	}

	return result
}
