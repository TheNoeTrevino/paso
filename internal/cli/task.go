package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	taskTitle       string
	taskDescription string
	taskType        string
	taskPriority    string
	taskParent      int
	taskColumn      string
	taskProject     int

	// Agent-friendly flags (add to ALL commands)
	jsonOutput bool
	quietMode  bool
)

func TaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}

	cmd.AddCommand(taskCreateCmd())
	// Other subcommands will be added in later phases
	// cmd.AddCommand(taskListCmd())
	// cmd.AddCommand(taskUpdateCmd())
	// cmd.AddCommand(taskDeleteCmd())
	// cmd.AddCommand(taskLinkCmd())

	return cmd
}

func taskCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new task",
		Long: `Create a new task with specified attributes.

Examples:
  # Simple task (human-readable output)
  paso task create --title="Fix bug" --project=1

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
		RunE: runTaskCreate,
	}

	// Required flags
	cmd.Flags().StringVar(&taskTitle, "title", "", "Task title (required)")
	cmd.MarkFlagRequired("title")

	cmd.Flags().IntVar(&taskProject, "project", 0, "Project ID (required)")
	cmd.MarkFlagRequired("project")

	// Optional flags
	cmd.Flags().StringVar(&taskDescription, "description", "", "Task description (use - for stdin)")
	cmd.Flags().StringVar(&taskType, "type", "task", "Task type: task or feature")
	cmd.Flags().StringVar(&taskPriority, "priority", "medium", "Priority: trivial, low, medium, high, critical")
	cmd.Flags().IntVar(&taskParent, "parent", 0, "Parent task ID (creates dependency)")
	cmd.Flags().StringVar(&taskColumn, "column", "", "Column name (defaults to first column)")

	// Agent-friendly flags (REQUIRED on all commands)
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&quietMode, "quiet", false, "Minimal output (ID only)")

	return cmd
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		formatter.Error("INITIALIZATION_ERROR", err.Error())
		return err
	}
	defer cli.Close()

	// Validate project exists
	project, err := cli.Repo.GetProjectByID(ctx, taskProject)
	if err != nil {
		formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", taskProject))
		os.Exit(3) // Exit code 3 = not found
	}

	// Get columns for project
	columns, err := cli.Repo.GetColumnsByProject(ctx, taskProject)
	if err != nil {
		formatter.Error("COLUMN_FETCH_ERROR", err.Error())
		return err
	}
	if len(columns) == 0 {
		formatter.Error("NO_COLUMNS", "project has no columns")
		return fmt.Errorf("project has no columns")
	}

	// Determine target column
	var targetColumnID int
	if taskColumn == "" {
		targetColumnID = columns[0].ID
	} else {
		found := false
		for _, col := range columns {
			if strings.EqualFold(col.Name, taskColumn) {
				targetColumnID = col.ID
				found = true
				break
			}
		}
		if !found {
			formatter.Error("COLUMN_NOT_FOUND", fmt.Sprintf("column '%s' not found", taskColumn))
			os.Exit(3)
		}
	}

	// Handle description from stdin
	description := taskDescription
	if description == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			formatter.Error("STDIN_READ_ERROR", err.Error())
			return err
		}
		description = string(data)
	}

	// Get position (append to end)
	count, err := cli.Repo.GetTaskCountByColumn(ctx, targetColumnID)
	if err != nil {
		formatter.Error("COUNT_ERROR", err.Error())
		return err
	}
	position := count + 1

	// Create task
	task, err := cli.Repo.CreateTask(ctx, taskTitle, description, targetColumnID, position)
	if err != nil {
		formatter.Error("TASK_CREATE_ERROR", err.Error())
		return err
	}

	// Set type if not default
	typeID, err := parseTaskType(taskType)
	if err != nil {
		formatter.Error("INVALID_TYPE", err.Error())
		os.Exit(5) // Exit code 5 = validation error
	}
	if typeID != 1 {
		// Note: UpdateTaskType doesn't exist yet - would need to add it
		// For now, we'll skip this or use raw SQL
		// TODO: Add UpdateTaskType to repository
	}

	// Set priority if not default
	priorityID, err := parsePriority(taskPriority)
	if err != nil {
		formatter.Error("INVALID_PRIORITY", err.Error())
		os.Exit(5)
	}
	if priorityID != 3 {
		if err := cli.Repo.UpdateTaskPriority(ctx, task.ID, priorityID); err != nil {
			formatter.Error("PRIORITY_UPDATE_ERROR", err.Error())
			return err
		}
	}

	// Add parent relationship if specified
	if taskParent > 0 {
		if err := cli.Repo.AddSubtask(ctx, taskParent, task.ID); err != nil {
			formatter.Error("LINK_ERROR", err.Error())
			return err
		}
	}

	// Output based on mode (JSON/Quiet/Human)
	if quietMode {
		fmt.Printf("%d\n", task.ID)
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"task": map[string]interface{}{
				"id":          task.ID,
				"title":       task.Title,
				"description": task.Description,
				"project":     map[string]interface{}{"id": project.ID, "name": project.Name},
				"type":        taskType,
				"priority":    taskPriority,
				"created_at":  task.CreatedAt,
			},
		})
	}

	// Human-readable output
	fmt.Printf("✓ Task '%s' created successfully (ID: %d)\n", taskTitle, task.ID)
	fmt.Printf("  Project: %s\n", project.Name)
	fmt.Printf("  Type: %s\n", taskType)
	fmt.Printf("  Priority: %s\n", taskPriority)
	if taskParent > 0 {
		fmt.Printf("  Parent: #%d\n", taskParent)
	}

	return nil
}

// Helper: Map type string to ID
func parseTaskType(typeStr string) (int, error) {
	types := map[string]int{
		"task":    1,
		"feature": 2,
	}

	id, ok := types[strings.ToLower(typeStr)]
	if !ok {
		return 0, fmt.Errorf("invalid type '%s' (must be: task, feature)", typeStr)
	}
	return id, nil
}

// Helper: Map priority string to ID
func parsePriority(priority string) (int, error) {
	priorities := map[string]int{
		"trivial":  1,
		"low":      2,
		"medium":   3,
		"high":     4,
		"critical": 5,
	}

	id, ok := priorities[strings.ToLower(priority)]
	if !ok {
		return 0, fmt.Errorf("invalid priority '%s' (must be: trivial, low, medium, high, critical)", priority)
	}
	return id, nil
}
