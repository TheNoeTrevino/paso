package task

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
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
		RunE: runCreate,
	}

	// Required flags
	cmd.Flags().String("title", "", "Task title (required)")
	if err := cmd.MarkFlagRequired("title"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Int("project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

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

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	taskTitle, _ := cmd.Flags().GetString("title")
	taskDescription, _ := cmd.Flags().GetString("description")
	taskType, _ := cmd.Flags().GetString("type")
	taskPriority, _ := cmd.Flags().GetString("priority")
	taskParent, _ := cmd.Flags().GetInt("parent")
	taskBlockedBy, _ := cmd.Flags().GetInt("blocked-by")
	taskBlocks, _ := cmd.Flags().GetInt("blocks")
	taskColumn, _ := cmd.Flags().GetString("column")
	taskProject, _ := cmd.Flags().GetInt("project")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

	formatter := &cli.OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cliInstance, err := cli.NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Validate project exists
	project, err := cliInstance.App.ProjectService.GetProjectByID(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("PROJECT_NOT_FOUND",
			fmt.Sprintf("project %d not found", taskProject),
			"Use 'paso project list' to see available projects or 'paso project create' to create a new one"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitNotFound)
	}

	// Get columns for project
	columns, err := cliInstance.App.ColumnService.GetColumnsByProject(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.Error("COLUMN_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	if len(columns) == 0 {
		if fmtErr := formatter.ErrorWithSuggestion("NO_COLUMNS",
			"project has no columns",
			"Create columns using 'paso column create --name=<name> --project="+fmt.Sprintf("%d", taskProject)+"'"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
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
			availableColumns := make([]string, len(columns))
			for i, col := range columns {
				availableColumns[i] = col.Name
			}
			if fmtErr := formatter.ErrorWithSuggestion("COLUMN_NOT_FOUND",
				fmt.Sprintf("column '%s' not found", taskColumn),
				fmt.Sprintf("Available columns: %s", strings.Join(availableColumns, ", "))); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			os.Exit(cli.ExitNotFound)
		}
	}

	// Handle description from stdin
	description := taskDescription
	if description == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			if fmtErr := formatter.Error("STDIN_READ_ERROR", err.Error()); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			return err
		}
		description = string(data)
	}

	// Parse type
	ttype := taskType
	if ttype == "" {
		ttype = "task"
	}
	typeID, err := cli.ParseTaskType(ttype)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("INVALID_TYPE", err.Error(),
			"Valid types are: task, feature"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitValidation)
	}

	// Parse priority
	priority := taskPriority
	if priority == "" {
		priority = "medium"
	}
	priorityID, err := cli.ParsePriority(priority)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("INVALID_PRIORITY", err.Error(),
			"Valid priorities are: trivial, low, medium, high, critical"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(cli.ExitValidation)
	}

	// Create task with all parameters
	// Position set to 9999 to append to end (will be adjusted if needed)
	req := taskservice.CreateTaskRequest{
		Title:       taskTitle,
		Description: description,
		ColumnID:    targetColumnID,
		Position:    9999,
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
		if fmtErr := formatter.Error("TASK_CREATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
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
	if taskBlockedBy > 0 {
		fmt.Printf("  Blocked by: #%d\n", taskBlockedBy)
	}
	if taskBlocks > 0 {
		fmt.Printf("  Blocks: #%d\n", taskBlocks)
	}

	return nil
}
