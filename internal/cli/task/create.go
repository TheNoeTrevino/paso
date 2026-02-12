// Package task holds all cli commands related to tasks
// e.g., paso task ...
package task

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/handler"
	"github.com/thenoetrevino/paso/internal/cli/styles"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
	taskservice "github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/user"
)

// CreateCmd returns the task create subcommand
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new task",
		Long: `Create a new task with specified attributes.

Examples:
  # Simple task using shorthand flags
  paso task create -t "Fix bug" -p 1 -b 15 -B 20

  # Full example with all options (shorthand)
  paso task create -t "Add authentication" -d "Implement JWT auth" -T feature -r high -P 3 -p 1

  # JSON output for agents
  paso task create -t "Fix bug" -p 1 -j

  # Quiet mode for bash capture
  TASK_ID=$(paso task create -t "Fix bug" -p 1 -q)

  # Long-form flags also supported
  paso task create --title="Fix bug" --project=1 --blocked-by 15 --blocks 20
`,
		RunE: handler.Command(&createHandler{}, parseCreateFlags),
	}

	// Required flags
	cmd.Flags().StringP("title", "t", "", "Task title (required)")
	if err := cmd.MarkFlagRequired("title"); err != nil {
		slog.Error("failed to mark flag as required", "error", err)
	}

	cmd.Flags().IntP("project", "p", 0, "Project ID (uses git branch association if not specified)")

	// Optional flags
	cmd.Flags().StringP("description", "d", "", "Task description (use - for stdin)")
	cmd.Flags().StringP("type", "T", "task", "Task type: task or feature")
	cmd.Flags().StringP("priority", "r", "medium", "Priority: trivial, low, medium, high, critical")
	cmd.Flags().IntP("parent", "P", 0, "Parent task ID (creates dependency)")
	cmd.Flags().IntP("blocked-by", "b", 0, "Task ID that blocks this task")
	cmd.Flags().IntP("blocks", "B", 0, "Task ID that is blocked by this task")
	cmd.Flags().StringP("column", "c", "", "Column name (defaults to first column)")
	cmd.Flags().StringP("assignee", "a", "", "Assignee name (defaults to active assignee)")
	cmd.Flags().StringP("estimate", "e", "", "Time estimate (e.g. 2h, 30m, 1d)")

	// Agent-friendly flags (REQUIRED on all commands)
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

// createHandler implements handler.Handler for task creation
type createHandler struct{}

// Execute implements the Handler interface
func (h *createHandler) Execute(ctx context.Context, args *handler.Arguments) (any, error) {
	// Get flag values from arguments
	taskTitle, err := args.MustGetString("title")
	if err != nil {
		return nil, err
	}
	taskDescription := args.GetString("description", "")
	taskType := args.GetString("type", "task")
	taskPriority := args.GetString("priority", "medium")
	taskParent := args.GetInt("parent", 0)
	taskBlockedBy := args.GetInt("blocked-by", 0)
	taskBlocks := args.GetInt("blocks", 0)
	taskColumn := args.GetString("column", "")
	taskAssignee := args.GetString("assignee", "")
	taskEstimate := args.GetString("estimate", "")

	// Initialize CLI first (uses injected instance from context if in test mode)
	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize CLI: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to close CLI", "error", err)
		}
	}()

	// Get project ID from flag or git branch
	cmd := args.GetCmd()
	taskProject, err := cli.GetProjectIDWithCLI(cmd, cliInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: no project specified: use --project flag or create a project associated with this branch")
	}

	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, taskProject)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: project %d not found", taskProject)
	}

	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, taskProject)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch columns: %w", err)
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("failed to create task: project has no columns")
	}

	var targetColumnID int
	if taskColumn == "" {
		targetColumnID = columns[0].ID
	} else {
		col, err := cli.FindColumnByName(columns, taskColumn)
		if err != nil {
			return nil, fmt.Errorf("failed to find column: column '%s' not found", taskColumn)
		}
		targetColumnID = col.ID
	}

	description := taskDescription
	if description == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
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

	// Resolve assignee: use flag value, fall back to active assignee from config
	var assigneeID int
	assigneeName := taskAssignee
	if assigneeName == "" {
		cfg, err := config.Load()
		if err == nil {
			assigneeName = cfg.GetActiveAssignee()
		}
		if assigneeName == "" {
			assigneeName = user.GetCurrentUsername()
		}
	}
	if assigneeName != "" {
		assignee, err := cliInstance.App.AssigneeService.GetOrCreate(ctx, assigneeName)
		if err != nil {
			slog.Warn("failed to resolve assignee", "assignee", assigneeName, "error", err)
		} else {
			assigneeID = assignee.ID
		}
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
		AssigneeID:  assigneeID,
		Estimate:    taskEstimate,
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
		return nil, fmt.Errorf("failed to create task: %w", err)
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

// PrettyPrint implements the PrettyPrintable interface for styled output
func (r *taskCreateResult) PrettyPrint(colorScheme colors.ColorScheme) string {
	details := []styles.Detail{
		{Key: "ID", Value: strconv.Itoa(r.ID)},
		{Key: "Title", Value: r.Title},
		{Key: "Project", Value: r.Project},
		{Key: "Priority", Value: r.Priority},
		{Key: "Type", Value: r.Type},
	}

	if r.Description != "" {
		truncated := styles.TruncateString(r.Description, 60)
		details = append(details, styles.Detail{Key: "Description", Value: truncated})
	}

	return styles.RenderSuccessWithDetails("Task created successfully", details, colorScheme)
}

func parseCreateFlags(cmd *cobra.Command) error {
	// Validate required flags
	title, _ := cmd.Flags().GetString("title")
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("failed to validate input: title is required")
	}
	return nil
}
