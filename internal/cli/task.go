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
	cmd.AddCommand(taskReadyCmd())
	cmd.AddCommand(taskBlockedCmd())

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
	cmd.Flags().String("column", "", "Column name (defaults to first column)")

	// Agent-friendly flags (REQUIRED on all commands)
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	taskTitle, _ := cmd.Flags().GetString("title")
	taskDescription, _ := cmd.Flags().GetString("description")
	taskType, _ := cmd.Flags().GetString("type")
	taskPriority, _ := cmd.Flags().GetString("priority")
	taskParent, _ := cmd.Flags().GetInt("parent")
	taskColumn, _ := cmd.Flags().GetString("column")
	taskProject, _ := cmd.Flags().GetInt("project")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

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
		os.Exit(ExitNotFound)
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
			os.Exit(ExitNotFound)
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
		os.Exit(ExitValidation)
	}
	if typeID != 1 {
		if err := cli.Repo.UpdateTaskType(ctx, task.ID, typeID); err != nil {
			if fmtErr := formatter.Error("TYPE_UPDATE_ERROR", err.Error()); fmtErr != nil {
				log.Printf("Error formatting error message: %v", fmtErr)
			}
			return err
		}
	}

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
		os.Exit(ExitValidation)
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
	cmd.Flags().Int("project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runTaskList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	taskProject, _ := cmd.Flags().GetInt("project")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

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

// taskReadyCmd returns the task ready subcommand
func taskReadyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ready",
		Short: "List tasks ready to work on",
		Long: `List all tasks that have no blocking dependencies.

These are tasks that can be started immediately as they are not
waiting on any other tasks to be completed.

Examples:
  # Human-readable output
  paso task ready --project=1

  # JSON output for agents
  paso task ready --project=1 --json

  # Quiet mode for bash capture
  TASK_IDS=$(paso task ready --project=1 --quiet)
`,
		RunE: runTaskReady,
	}

	// Required flags
	cmd.Flags().Int("project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runTaskReady(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	taskProject, _ := cmd.Flags().GetInt("project")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

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
	_, err = cli.Repo.GetProjectByID(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("PROJECT_NOT_FOUND",
			fmt.Sprintf("project %d not found", taskProject),
			"Use 'paso project list' to see available projects"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(ExitNotFound)
	}

	// Get all tasks for project (includes IsBlocked field)
	tasksByColumn, err := cli.Repo.GetTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Filter for ready tasks (IsBlocked == false)
	var readyTasks []*models.TaskSummary
	for _, columnTasks := range tasksByColumn {
		for _, task := range columnTasks {
			if !task.IsBlocked {
				readyTasks = append(readyTasks, task)
			}
		}
	}

	// Output in appropriate format
	if quietMode {
		// Just print IDs
		for _, t := range readyTasks {
			fmt.Printf("%d\n", t.ID)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"tasks":   readyTasks,
			"count":   len(readyTasks),
		})
	}

	// Human-readable output
	if len(readyTasks) == 0 {
		fmt.Println("No ready tasks found")
		return nil
	}

	fmt.Printf("Found %d ready tasks:\n\n", len(readyTasks))
	for _, t := range readyTasks {
		// Include priority if set
		priorityInfo := ""
		if t.PriorityDescription != "" && t.PriorityDescription != "medium" {
			priorityInfo = fmt.Sprintf(" [%s]", t.PriorityDescription)
		}
		fmt.Printf("  [%d] %s%s\n", t.ID, t.Title, priorityInfo)
	}

	return nil
}

// taskBlockedCmd returns the task blocked subcommand
func taskBlockedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocked",
		Short: "List blocked tasks",
		Long: `List all tasks that are blocked by dependencies.

These are tasks that cannot be started until their blocking
dependencies are completed.

Examples:
  # Human-readable output
  paso task blocked --project=1

  # JSON output for agents
  paso task blocked --project=1 --json

  # Quiet mode for bash capture
  TASK_IDS=$(paso task blocked --project=1 --quiet)
`,
		RunE: runTaskBlocked,
	}

	// Required flags
	cmd.Flags().Int("project", 0, "Project ID (required)")
	if err := cmd.MarkFlagRequired("project"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

func runTaskBlocked(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	taskProject, _ := cmd.Flags().GetInt("project")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quietMode, _ := cmd.Flags().GetBool("quiet")

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
	_, err = cli.Repo.GetProjectByID(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.ErrorWithSuggestion("PROJECT_NOT_FOUND",
			fmt.Sprintf("project %d not found", taskProject),
			"Use 'paso project list' to see available projects"); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		os.Exit(ExitNotFound)
	}

	// Get all tasks for project (includes IsBlocked field)
	tasksByColumn, err := cli.Repo.GetTaskSummariesByProject(ctx, taskProject)
	if err != nil {
		if fmtErr := formatter.Error("TASK_FETCH_ERROR", err.Error()); fmtErr != nil {
			log.Printf("Error formatting error message: %v", fmtErr)
		}
		return err
	}

	// Filter for blocked tasks (IsBlocked == true)
	var blockedTasks []*models.TaskSummary
	for _, columnTasks := range tasksByColumn {
		for _, task := range columnTasks {
			if task.IsBlocked {
				blockedTasks = append(blockedTasks, task)
			}
		}
	}

	// Output in appropriate format
	if quietMode {
		// Just print IDs
		for _, t := range blockedTasks {
			fmt.Printf("%d\n", t.ID)
		}
		return nil
	}

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"tasks":   blockedTasks,
			"count":   len(blockedTasks),
		})
	}

	// Human-readable output
	if len(blockedTasks) == 0 {
		fmt.Println("No blocked tasks found")
		return nil
	}

	fmt.Printf("Found %d blocked tasks:\n\n", len(blockedTasks))
	for _, t := range blockedTasks {
		// Include priority if set
		priorityInfo := ""
		if t.PriorityDescription != "" && t.PriorityDescription != "medium" {
			priorityInfo = fmt.Sprintf(" [%s]", t.PriorityDescription)
		}
		fmt.Printf("  [%d] %s%s (BLOCKED)\n", t.ID, t.Title, priorityInfo)
	}

	return nil
}

// taskUpdateCmd returns the task update subcommand
func taskUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a task",
		Long:  "Update task title, description, or priority.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			taskID, _ := cmd.Flags().GetInt("id")
			taskTitle, _ := cmd.Flags().GetString("title")
			taskDescription, _ := cmd.Flags().GetString("description")
			taskPriority, _ := cmd.Flags().GetString("priority")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			quietMode, _ := cmd.Flags().GetBool("quiet")

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

			// At least one update field must be provided
			titleFlag := cmd.Flags().Lookup("title")
			descFlag := cmd.Flags().Lookup("description")
			priorityFlag := cmd.Flags().Lookup("priority")

			if !titleFlag.Changed && !descFlag.Changed && !priorityFlag.Changed {
				if fmtErr := formatter.Error("NO_UPDATES", "at least one of --title, --description, or --priority must be specified"); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				os.Exit(ExitUsage)
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
					os.Exit(ExitValidation)
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
	cmd.Flags().Int("id", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Optional update flags
	cmd.Flags().String("title", "", "New task title")
	cmd.Flags().String("description", "", "New task description")
	cmd.Flags().String("priority", "", "New priority: trivial, low, medium, high, critical")

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output (ID only)")

	return cmd
}

// taskDeleteCmd returns the task delete subcommand
func taskDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a task",
		Long:  "Delete a task by ID (requires confirmation unless --force or --quiet).",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			taskID, _ := cmd.Flags().GetInt("id")
			force, _ := cmd.Flags().GetBool("force")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			quietMode, _ := cmd.Flags().GetBool("quiet")

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

			// Get task details for confirmation
			task, err := cli.Repo.GetTaskDetail(ctx, taskID)
			if err != nil {
				if fmtErr := formatter.Error("TASK_NOT_FOUND", fmt.Sprintf("task %d not found", taskID)); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				os.Exit(ExitNotFound)
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

	cmd.Flags().Int("id", 0, "Task ID (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Bool("force", false, "Skip confirmation")

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	return cmd
}

// taskLinkCmd returns the task link subcommand
func taskLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Link tasks with relationships",
		Long: `Create a relationship between two tasks.

Relationship Types:
  (default)  Parent-Child: Non-blocking hierarchical relationship
  --blocker  Blocked By/Blocker: Blocking relationship (parent blocked by child)
  --related  Related To: Non-blocking associative relationship

The --blocker and --related flags are mutually exclusive. If neither is specified,
a parent-child relationship is created.

Examples:
  # Parent-child relationship (default)
  paso task link --parent=5 --child=3

  # Blocking relationship (task 5 blocked by task 3)
  paso task link --parent=5 --child=3 --blocker

  # Related relationship
  paso task link --parent=5 --child=3 --related
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			parentID, _ := cmd.Flags().GetInt("parent")
			childID, _ := cmd.Flags().GetInt("child")
			blocker, _ := cmd.Flags().GetBool("blocker")
			related, _ := cmd.Flags().GetBool("related")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			quietMode, _ := cmd.Flags().GetBool("quiet")

			formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}

			// Validate mutually exclusive flags
			if blocker && related {
				if fmtErr := formatter.Error("INVALID_FLAGS",
					"cannot specify both --blocker and --related flags"); fmtErr != nil {
					log.Printf("Error formatting error message: %v", fmtErr)
				}
				os.Exit(ExitUsage)
			}

			// Determine relation type ID
			relationTypeID := 1 // Default: Parent/Child
			relationTypeName := "parent-child"

			if blocker {
				relationTypeID = 2 // Blocked By/Blocker
				relationTypeName = "blocking"
			} else if related {
				relationTypeID = 3 // Related To
				relationTypeName = "related"
			}

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

			// Create the relationship with specific type
			if err := cli.Repo.AddSubtaskWithRelationType(ctx, parentID, childID, relationTypeID); err != nil {
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
					"success":          true,
					"parent_id":        parentID,
					"child_id":         childID,
					"relation_type_id": relationTypeID,
					"relation_type":    relationTypeName,
				})
			}

			// Human-readable output with relationship type
			switch relationTypeID {
			case 2:
				fmt.Printf("✓ Created blocking relationship: task %d is blocked by task %d\n", parentID, childID)
			case 3:
				fmt.Printf("✓ Created related relationship between task %d and task %d\n", parentID, childID)
			default:
				fmt.Printf("✓ Linked task %d as child of task %d\n", childID, parentID)
			}

			return nil
		},
	}

	// Required flags
	cmd.Flags().Int("parent", 0, "Parent task ID (required)")
	if err := cmd.MarkFlagRequired("parent"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	cmd.Flags().Int("child", 0, "Child task ID (required)")
	if err := cmd.MarkFlagRequired("child"); err != nil {
		log.Printf("Error marking flag as required: %v", err)
	}

	// Agent-friendly flags
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("quiet", false, "Minimal output")

	// Relationship type flags (mutually exclusive)
	cmd.Flags().Bool("blocker", false, "Create blocking relationship (Blocked By/Blocker)")
	cmd.Flags().Bool("related", false, "Create related relationship (Related To)")

	return cmd
}
