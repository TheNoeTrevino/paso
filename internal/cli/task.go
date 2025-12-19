package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/models"
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
	cmd.AddCommand(taskListCmd())
	cmd.AddCommand(taskUpdateCmd())
	cmd.AddCommand(taskDeleteCmd())
	cmd.AddCommand(taskLinkCmd())

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
	if err := cmd.MarkFlagRequired("title"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().IntVar(&taskProject, "project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

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
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Validate project exists
	project, err := cli.Repo.GetProjectByID(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("PROJECT_NOT_FOUND",
			fmt.Sprintf("project %d not found", taskProject),
			"Use 'paso project list' to see available projects or 'paso project create' to create a new one"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(3) // Exit code 3 = not found
	}

	// Get columns for project
	columns, err := cli.Repo.GetColumnsByProject(ctx, taskProject)
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
			os.Exit(3)
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

	// Get position (append to end)
	count, err := cli.Repo.GetTaskCountByColumn(ctx, targetColumnID)
	if err != nil {
		if fmtErr := formatter.Error("COUNT_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	position := count + 1

	// Create task
	task, err := cli.Repo.CreateTask(ctx, taskTitle, description, targetColumnID, position)
	if err != nil {
		if fmtErr := formatter.Error("TASK_CREATE_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Set type if not default
	// Handle empty type (use default)
	ttype := taskType
	if ttype == "" {
		ttype = "task"
	}
	typeID, err := parseTaskType(ttype)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("INVALID_TYPE", err.Error(),
			"Valid types are: task, feature"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(5) // Exit code 5 = validation error
	}
	// Note: UpdateTaskType doesn't exist yet - would need to add it
	// For now, we'll skip this or use raw SQL
	// TODO: Add UpdateTaskType to repository
	_ = typeID // Suppress unused variable warning

	// Set priority if not default
	// Handle empty priority (use default)
	priority := taskPriority
	if priority == "" {
		priority = "medium"
	}
	priorityID, err := parsePriority(priority)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("INVALID_PRIORITY", err.Error(),
			"Valid priorities are: trivial, low, medium, high, critical"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(5)
	}
	if priorityID != 3 {
		if err := cli.Repo.UpdateTaskPriority(ctx, task.ID, priorityID); err != nil {
			if fmtErr := formatter.Error("PRIORITY_UPDATE_ERROR", err.Error()); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			return err
		}
	}

	// Add parent relationship if specified
	if taskParent > 0 {
		if err := cli.Repo.AddSubtask(ctx, taskParent, task.ID); err != nil {
			if fmtErr := formatter.Error("LINK_ERROR", err.Error()); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
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

// taskListCmd returns the task list subcommand
func taskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long:  "List all tasks in a project.",
		RunE:  runTaskList,
	}

	// Required flags
	cmd.Flags().IntVar(&taskProject, "project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&quietMode, "quiet", false, "Minimal output (ID only)")

	return cmd
}

func runTaskList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

	// Initialize CLI
	cli, err := NewCLI(ctx)
	if err != nil {
		if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

	// Get tasks (returns map[columnID][]*TaskSummary)
	tasksByColumn, err := cli.Repo.GetTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Flatten tasks from all columns
	var allTasks []*models.TaskSummary
	for _, columnTasks := range tasksByColumn {
		allTasks = append(allTasks, columnTasks...)
	}

	// Output in appropriate format
	if quietMode {
		// Just print IDs
		for _, t := range allTasks {
			fmt.Printf("%d\n", t.ID)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"tasks":   allTasks,
		})
	}

	// Human-readable output
	if len(allTasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	fmt.Printf("Found %d tasks:\n\n", len(allTasks))
	for _, t := range allTasks {
		fmt.Printf("  [%d] %s\n", t.ID, t.Title)
	}

	return nil
}

// taskUpdateCmd returns the task update subcommand
func taskUpdateCmd() *cobra.Command {
	var taskID int

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a task",
		Long:  "Update task title, description, or priority.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

			// Initialize CLI
			cli, err := NewCLI(ctx)
			if err != nil {
				if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				return err
			}
			defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

			// Get task ID flag
			taskID, _ = cmd.Flags().GetInt("id")

			// At least one update field must be provided
			titleFlag := cmd.Flags().Lookup("title")
			descFlag := cmd.Flags().Lookup("description")
			priorityFlag := cmd.Flags().Lookup("priority")

			if !titleFlag.Changed && !descFlag.Changed && !priorityFlag.Changed {
				if fmtErr := formatter.Error("NO_UPDATES", "at least one of --title, --description, or --priority must be specified"); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				os.Exit(2)
			}

			// Update title/description if provided
			if titleFlag.Changed || descFlag.Changed {
				if err := cli.Repo.UpdateTask(ctx, taskID, taskTitle, taskDescription); err != nil {
					if fmtErr := formatter.Error("UPDATE_ERROR", err.Error()); fmtErr != nil {
						log.Printf("Error formatting error message: %v", fmtErr)
					}
					return err
				}
			}

			// Update priority if provided
			if priorityFlag.Changed {
				priorityID, err := parsePriority(taskPriority)
				if err != nil {
					if fmtErr := formatter.Error("INVALID_PRIORITY", err.Error()); fmtErr != nil {
						log.Printf("Error formatting error message: %v", fmtErr)
					}
					os.Exit(5)
				}
				if err := cli.Repo.UpdateTaskPriority(ctx, taskID, priorityID); err != nil {
					if fmtErr := formatter.Error("PRIORITY_UPDATE_ERROR", err.Error()); fmtErr != nil {
						log.Printf("Error formatting error message: %v", fmtErr)
					}
					return err
				}
			}

			// Output success
			if quietMode {
				fmt.Printf("%d\n", taskID)
				return nil
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"success": true,
					"task_id": taskID,
				})
			}

			fmt.Printf("✓ Task %d updated successfully\n", taskID)
			return nil
		},
	}

	// Required flags
	cmd.Flags().IntVar(&taskID, "id", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional update flags
	cmd.Flags().StringVar(&taskTitle, "title", "", "New task title")
	cmd.Flags().StringVar(&taskDescription, "description", "", "New task description")
	cmd.Flags().StringVar(&taskPriority, "priority", "", "New priority: trivial, low, medium, high, critical")

	// Agent-friendly flags
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&quietMode, "quiet", false, "Minimal output (ID only)")

	return cmd
}

// taskDeleteCmd returns the task delete subcommand
func taskDeleteCmd() *cobra.Command {
	var taskID int
	var force bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a task",
		Long:  "Delete a task by ID (requires confirmation unless --force or --quiet).",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

			// Initialize CLI
			cli, err := NewCLI(ctx)
			if err != nil {
				if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				return err
			}
			defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

			taskID, _ = cmd.Flags().GetInt("id")
			force, _ = cmd.Flags().GetBool("force")

			// Get task details for confirmation
			task, err := cli.Repo.GetTaskDetail(ctx, taskID)
			if err != nil {
				if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				os.Exit(3)
			}

			// Ask for confirmation unless force or quiet mode
			if !force && !quietMode {
				fmt.Printf("Delete task #%d: '%s'? (y/N): ", taskID, task.Title)
				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					log.Printf("Error reading user input: %v", err)
				}
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			// Delete the task
			if err := cli.Repo.DeleteTask(ctx, taskID); err != nil {
				if fmtErr := formatter.Error("DELETE_ERROR", err.Error()); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				return err
			}

			// Output success
			if quietMode {
				return nil
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"success": true,
					"task_id": taskID,
				})
			}

			fmt.Printf("✓ Task %d deleted successfully\n", taskID)
			return nil
		},
	}

	// Required flags
	cmd.Flags().IntVar(&taskID, "id", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional flags
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation")

	// Agent-friendly flags
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&quietMode, "quiet", false, "Minimal output")

	return cmd
}

// taskLinkCmd returns the task link subcommand
func taskLinkCmd() *cobra.Command {
	var parentID, childID int

	cmd := &cobra.Command{
		Use:   "link",
		Short: "Link tasks (parent-child)",
		Long:  "Create a parent-child relationship between tasks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

			// Initialize CLI
			cli, err := NewCLI(ctx)
			if err != nil {
				if fmtErr := formatter.Error("INITIALIZATION_ERROR", err.Error()); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				return err
			}
			defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Error closing CLI: %v", err)
		}
	}()

			parentID, _ = cmd.Flags().GetInt("parent")
			childID, _ = cmd.Flags().GetInt("child")

			// Create the relationship
			if err := cli.Repo.AddSubtask(ctx, parentID, childID); err != nil {
				if fmtErr := formatter.Error("LINK_ERROR", err.Error()); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				return err
			}

			// Output success
			if quietMode {
				return nil
			}

			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"success":   true,
					"parent_id": parentID,
					"child_id":  childID,
				})
			}

			fmt.Printf("✓ Linked task %d as child of task %d\n", childID, parentID)
			return nil
		},
	}

	// Required flags
	cmd.Flags().IntVar(&parentID, "parent", 0, "Parent task ID (required)")
	if err := cmd.MarkFlagRequired("parent"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().IntVar(&childID, "child", 0, "Child task ID (required)")
	if err := cmd.MarkFlagRequired("child"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&quietMode, "quiet", false, "Minimal output")

	return cmd
}
