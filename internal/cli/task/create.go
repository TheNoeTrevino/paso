// Package task holds all cli commands related to tasks
// e.g., paso task ...
package task

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/handler"
	"github.com/thenoetrevino/paso/internal/models"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
)

// CreateCmd returns the task create subcommand
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new task",
		Long: `Create a new task with specified attributes.

Examples:
  # Simple task (human-readable output)
  paso task create --title="Fix bug" --project=1 --blocked-by 15 --blocks 20

  # JSON output for agents
  paso task create --title="Fix bug" --project=1 --json

  # Quiet mode for bash capture
  TASK_ID=$(paso task create --title="Fix bug" --project=1 --quiet)

  # Full example with all options
  paso task create \
    --title="Add authentication" \
    --description="Implement JWT auth" \
    --type=feature \
    --priority=high \
    --parent=3 \
    --project=1
`,
		RunE: handler.Command(&createHandler{}, parseCreateFlags),
	}

	// Required flags
	cmd.Flags().String("title", "", "Task title (required)")
	if err := cmd.MarkFlagRequired("title"); err != nil {
		slog.Error("Error marking flag as required", "error", err)
	}

	cmd.Flags().Int("project", 0, "Project ID (uses PASO_PROJECT env var if not specified)")

	// Optional flags
	cmd.Flags().String("description", "", "Task description (use - for stdin)")
	cmd.Flags().String("type", "task", "Task type: task or feature")
	cmd.Flags().String("priority", "medium", "Priority: trivial, low, medium, high, critical")
	cmd.Flags().Int("parent", 0, "Parent task ID (creates dependency)")
	cmd.Flags().Int("blocked-by", 0, "Task ID that blocks this task")
	cmd.Flags().Int("blocks", 0, "Task ID that is blocked by this task")
	cmd.Flags().String("column", "", "Column name (defaults to first column)")

	// Agent-friendly flags (REQUIRED on all commands)
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

// createHandler implements handler.Handler for task creation
type createHandler struct{}

// Execute implements the Handler interface
func (h *createHandler) Execute(ctx context.Context, args *handler.Arguments) (interface{}, error) {
	// Get flag values from arguments
	taskTitle := args.MustGetString("title")
	taskDescription := args.GetString("description", "")
	taskType := args.GetString("type", "task")
	taskPriority := args.GetString("priority", "medium")
	taskParent := args.GetInt("parent", 0)
	taskBlockedBy := args.GetInt("blocked-by", 0)
	taskBlocks := args.GetInt("blocks", 0)
	taskColumn := args.GetString("column", "")

	// Get project ID from flag or environment variable
	cmd := args.GetCmd()
	taskProject, err := cli.GetProjectID(cmd)
	if err != nil {
		return nil, fmt.Errorf("no project specified: use --project flag or set with 'eval $(paso use project <project-id>)'")
	}

	// Initialize CLI (uses injected instance from context if in test mode)
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialization error: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("Error closing CLI", "error", err)
		}
	}()

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, taskProject)
	if err != nil {
		return nil, fmt.Errorf("project %d not found", taskProject)
	}

	// Get columns for project
	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, taskProject)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch columns: %w", err)
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("project has no columns")
	}

	// Determine target column
	var targetColumnID int
	if taskColumn == "" {
		targetColumnID = columns[0].ID
	} else {
		col, err := cli.FindColumnByName(columns, taskColumn)
		if err != nil {
			return nil, fmt.Errorf("column '%s' not found", taskColumn)
		}
		targetColumnID = col.ID
	}

	// Handle description from stdin
	description := taskDescription
	if description == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("stdin read error: %w", err)
		}
		description = string(data)
	}

	// Parse type
	typeID, err := cli.ParseTaskType(taskType)
	if err != nil {
		return nil, err
	}

	// Parse priority
	priorityID, err := cli.ParsePriority(taskPriority)
	if err != nil {
		return nil, err
	}

	// Create task with all parameters
	// Position set to DefaultTaskPosition to append to end (will be adjusted if needed)
	req := taskservice.CreateTaskRequest{
		Title:       taskTitle,
		Description: description,
		ColumnID:    targetColumnID,
		Position:    models.DefaultTaskPosition,
		PriorityID:  priorityID,
		TypeID:      typeID,
	}

	// Add parent relationship if specified
	if taskParent > 0 {
		req.ParentIDs = []int{taskParent}
	}

	// Add blocking relationships if specified
	if taskBlockedBy > 0 {
		req.BlockedByIDs = []int{taskBlockedBy}
	}
	if taskBlocks > 0 {
		req.BlocksIDs = []int{taskBlocks}
	}

	task, err := cliInstance.App.TaskService.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("task creation error: %w", err)
	}

	return &taskCreateResult{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Project:     project.Name,
		Type:        taskType,
		Priority:    taskPriority,
		CreatedAt:   task.CreatedAt.String(),
	}, nil
}

// taskCreateResult represents the result of task creation
type taskCreateResult struct {
	ID          int
	Title       string
	Description string
	Project     string
	Type        string
	Priority    string
	CreatedAt   string
}

// GetID implements the GetID interface for quiet mode output
func (r *taskCreateResult) GetID() int {
	return r.ID
}

func parseCreateFlags(cmd *cobra.Command) error {
	// Validate required flags
	title, _ := cmd.Flags().GetString("title")
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}
