package task

import (
	"fmt"
	"strconv"
)

// EstimateInput represents the parsed input for estimating a task
type EstimateInput struct {
	TaskID   int
	Estimate string
	Clear    bool
}

// EstimateResult represents the output of a successful estimate operation
type EstimateResult struct {
	TaskID   int
	Estimate string
	Cleared  bool
}

// ParseEstimateArgs parses task ID and estimate from command arguments
func ParseEstimateArgs(args []string, clearFlag bool) (*EstimateInput, error) {
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

	estimate := ""
	if len(args) > 1 {
		estimate = args[1]
	}

	return &EstimateInput{
		TaskID:   taskID,
		Estimate: estimate,
		Clear:    clearFlag,
	}, nil
}

// FormatEstimateOutput generates the human-readable output message
func FormatEstimateOutput(result *EstimateResult) string {
	if result.Cleared {
		return fmt.Sprintf("Task %d estimate cleared", result.TaskID)
	}
	return fmt.Sprintf("Task %d estimate set to %s", result.TaskID, result.Estimate)
}

// FormatEstimateJSON generates the JSON output structure
func FormatEstimateJSON(result *EstimateResult) map[string]any {
	output := map[string]any{
		"success": true,
		"task_id": result.TaskID,
	}
	if result.Cleared {
		output["estimate"] = nil
	} else {
		output["estimate"] = result.Estimate
	}
	return output
}
