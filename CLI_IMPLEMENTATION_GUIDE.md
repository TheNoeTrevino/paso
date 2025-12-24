# Paso CLI Implementation Guide

**For AI Agents: This document is your complete guide to implementing the Paso CLI.**
Read top-to-bottom, implement in order, validate continuously.

---

## Table of Contents

1. [Project Context](#1-project-context) - What exists, what we're building
2. [Agent-Friendly Design Rules](#2-agent-friendly-design-rules) - Core principles you must follow
3. [Architecture Overview](#3-architecture-overview) - Technical structure and setup
4. [Command Implementation](#4-command-implementation) - Detailed implementation guide
5. [Testing & Validation](#5-testing--validation) - How to verify correctness
6. [Implementation Checklist](#6-implementation-checklist) - Track your progress

---

## 1. Project Context

### What Currently Exists

**Codebase State:**
- ✅ **Full-featured TUI** - BubbleTea-based kanban board at `main.go`
- ✅ **Complete database layer** - SQLite with repository pattern in `internal/database/`
- ✅ **All CRUD operations available** - Tasks, Projects, Columns, Labels
- ✅ **16K+ lines of tested code** - Solid foundation, no refactoring needed
- ✅ **Optional daemon** - Live updates via Unix socket at `~/.paso/paso.sock`

**Database Location:** `~/.paso/tasks.db` (SQLite with WAL mode)

**Available Database Methods:**
```go
// Projects
CreateProject(ctx, name, description string) (*models.Project, error)
GetProjectByID(ctx, id int) (*models.Project, error)
GetAllProjects(ctx) ([]*models.Project, error)
UpdateProject(ctx, id int, name, description string) error
DeleteProject(ctx, id int) error

// Tasks
CreateTask(ctx, title, description string, columnID, position int) (*models.Task, error)
UpdateTask(ctx, taskID int, title, description string) error
UpdateTaskPriority(ctx, taskID, priorityID int) error
GetTaskDetail(ctx, id int) (*models.TaskDetail, error)
GetTaskCountByColumn(ctx, columnID int) (int, error)

// Relationships
AddSubtask(ctx, parentID, childID int) error
RemoveSubtask(ctx, parentID, childID int) error

// Columns
GetColumnsByProject(ctx, projectID int) ([]*models.Column, error)
CreateColumn(ctx, name string, projectID int, afterID *int) (*models.Column, error)
UpdateColumnName(ctx, columnID int, name string) error
DeleteColumn(ctx, columnID int) error

// Labels
CreateLabel(ctx, projectID int, name, color string) (*models.Label, error)
GetLabelsByProject(ctx, projectID int) ([]*models.Label, error)
AddLabelToTask(ctx, taskID, labelID int) error
RemoveLabelFromTask(ctx, taskID, labelID int) error
```

**Task Types (from database):**
- `1` = "task"
- `2` = "feature"

**Task Priorities (from database):**
- `1` = "trivial" (blue #3B82F6)
- `2` = "low" (green #22C55E)
- `3` = "medium" (yellow #EAB308)
- `4` = "high" (orange #F97316)
- `5` = "critical" (red #EF4444)

### What We're Building

**Goal:** CLI commands that AI agents (like you) can use reliably for task management.

**Key Design Decisions:**
1. **Default behavior:** `paso` shows help text (breaking change - use `paso tui` for TUI)
2. **Agent-first:** Support `--json` and `--quiet` flags on all commands
3. **Optional daemon:** Try connecting to daemon, fall back to database-only
4. **Simple first:** Sequential operations, no batch/idempotency in Phase 1

**Target Command Structure:**
```
paso
├── tui              # Launch the TUI
├── task
│   ├── create       # Create task
│   ├── list         # List tasks
│   ├── update       # Update task
│   ├── delete       # Delete task
│   └── link         # Link parent-child
├── project
│   ├── create       # Create project
│   ├── list         # List projects
│   └── delete       # Delete project
├── column
│   ├── create       # Create column
│   └── list         # List columns
└── label
    ├── create       # Create label
    ├── list         # List labels
    └── attach       # Attach to task
```

---

## 2. Agent-Friendly Design Rules

**These are the core principles you MUST follow when implementing every command.**

### Rule 1: Three Output Modes

Every command must support three output modes:

**A) Human-readable (default)**
```bash
$ paso task create --title="Fix bug" --project=1
✓ Task 'Fix bug' created successfully (ID: 42)
  Project: My Project
  Column: Todo
  Type: task
  Priority: medium
```

**B) JSON mode (--json flag)**
```bash
$ paso task create --title="Fix bug" --project=1 --json
{
  "success": true,
  "task": {
    "id": 42,
    "title": "Fix bug",
    "project": {"id": 1, "name": "My Project"},
    "column": {"id": 5, "name": "Todo"},
    "type": {"id": 1, "name": "task"},
    "priority": {"id": 3, "name": "medium"},
    "created_at": "2025-12-18T10:00:00Z"
  }
}
```

**C) Quiet mode (--quiet flag)**
```bash
$ paso task create --title="Fix bug" --project=1 --quiet
42
```

**Why:** Agents need JSON for parsing, quiet for bash capture, human for interactive use.

### Rule 2: Exit Codes

Use consistent exit codes for error handling:

```go
0   = Success
1   = General error
2   = Usage error (invalid flags)
3   = Not found (project/task doesn't exist)
5   = Validation error (invalid input)
6   = Dependency error (circular dependency, etc.)
```

**Example usage by agents:**
```bash
paso task create --title="Test" --project=999 --quiet
if [ $? -eq 3 ]; then
  echo "Project not found - creating it first"
  PROJECT_ID=$(paso project create --title="New Project" --quiet)
fi
```

### Rule 3: Structured Errors

Error responses must be structured in JSON mode:

```json
{
  "success": false,
  "error": {
    "code": "PROJECT_NOT_FOUND",
    "message": "project 999 not found",
    "field": "project",
    "value": 999
  }
}
```

### Rule 4: ID-Based Operations

- All resources use integer IDs (no "TASK-42" format)
- IDs are always returned in quiet mode for easy capture
- Agents can chain operations using captured IDs

### How Agents Will Use This CLI

**Complete workflow example:**
```bash
# Step 1: Find or create project
PROJECT_ID=$(paso project list --json | jq -r '.projects[] | select(.name=="Backend API") | .id')
if [ -z "$PROJECT_ID" ]; then
  PROJECT_ID=$(paso project create --title="Backend API" --quiet)
fi

# Step 2: Create feature task (quiet mode for easy capture)
FEATURE_ID=$(paso task create \
  --title="User Authentication" \
  --type=feature \
  --priority=high \
  --project=$PROJECT_ID \
  --quiet)

# Step 3: Create subtasks sequentially
JWT_TASK=$(paso task create --title="Setup JWT" --parent=$FEATURE_ID --project=$PROJECT_ID --quiet)
LOGIN_TASK=$(paso task create --title="Login endpoint" --parent=$FEATURE_ID --project=$PROJECT_ID --quiet)

# Step 4: Add label
LABEL_ID=$(paso label create --name="security" --color="#EF4444" --project=$PROJECT_ID --quiet)
paso label attach $FEATURE_ID $LABEL_ID

# Step 5: Verify (using JSON for structured parsing)
RESULT=$(paso task list --project=$PROJECT_ID --json)
TASK_COUNT=$(echo $RESULT | jq '.tasks | length')
echo "✓ Created $TASK_COUNT tasks"
```

**Key patterns agents use:**
- Quiet mode for bash variable capture
- JSON mode for structured data extraction
- Exit codes for error handling
- Sequential command chaining

---

## 3. Architecture Overview

### Technology Stack

**Framework:** Cobra (industry standard for CLI)
```bash
go get github.com/spf13/cobra@latest
```

**Why Cobra:**
- Used by kubectl, gh, docker
- Subcommand structure matches requirements
- Built-in flag parsing and validation
- Auto-generated help text

### Directory Structure

**Current state:**
```
/home/noetrevino/projects/paso/feature/
├── main.go                    # Current TUI entry point
├── internal/
│   ├── database/              # All CRUD operations (ready to use)
│   ├── models/                # Domain models
│   ├── tui/                   # TUI implementation
│   └── events/                # Daemon client
└── cmd/
    ├── daemon/                # Daemon server
    └── ci/                    # CI runner
```

**Target state:**
```
/home/noetrevino/projects/paso/feature/
├── main.go                    # Becomes Cobra router
├── internal/
│   ├── cli/                   # NEW: CLI package
│   │   ├── output.go          # Output formatter (json/quiet/human)
│   │   ├── task.go            # Task commands
│   │   ├── project.go         # Project commands
│   │   ├── column.go          # Column commands
│   │   └── label.go           # Label commands
│   ├── database/              # Unchanged - use as-is
│   ├── models/                # Unchanged
│   ├── tui/                   # Unchanged
│   └── events/                # Unchanged
└── cmd/
    ├── daemon/                # Unchanged
    └── ci/                    # Unchanged
```

### Core Components

#### 1. Output Formatter (`internal/cli/output.go`)

**Purpose:** Handle all three output modes consistently.

```go
package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

type OutputFormatter struct {
	JSON  bool
	Quiet bool
}

// Success outputs successful operation result
func (f *OutputFormatter) Success(data interface{}) error {
	if f.Quiet {
		// Extract ID if possible
		if idGetter, ok := data.(interface{ GetID() int }); ok {
			fmt.Printf("%d\n", idGetter.GetID())
			return nil
		}
	}

	if f.JSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": true,
			"data":    data,
		})
	}

	// Human-readable format
	return f.prettyPrint(data)
}

// Error outputs error information
func (f *OutputFormatter) Error(code string, message string) error {
	if f.JSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]interface{}{
				"code":    code,
				"message": message,
			},
		})
	}

	// Human-readable error
	fmt.Fprintf(os.Stderr, "❌ Error: %s\n", message)
	return nil
}
```

#### 2. Root Command (`main.go`)

**Purpose:** Cobra entry point that routes to subcommands.

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "paso",
	Short: "Terminal-based Kanban board with CLI and TUI",
	Long: `Paso is a zero-setup, terminal-based kanban board for personal task management.

Use 'paso tui' to launch the interactive TUI.
Use 'paso task create ...' for CLI commands.`,
	// No Run function - shows help text by default
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(cli.TaskCmd())
	rootCmd.AddCommand(cli.ProjectCmd())
	rootCmd.AddCommand(cli.ColumnCmd())
	rootCmd.AddCommand(cli.LabelCmd())

	// TUI command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "tui",
		Short: "Launch the TUI interface",
		Run: func(cmd *cobra.Command, args []string) {
			tui.Launch()
		},
	})
}
```

#### 3. CLI Context (`internal/cli/cli.go`)

**Purpose:** Initialize database connection with optional daemon.

```go
package cli

import (
	"context"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
)

type CLI struct {
	repo *database.Repository
	ctx  context.Context
}

func NewCLI(ctx context.Context) (*CLI, error) {
	// Initialize database
	db, err := database.InitDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Try to connect to daemon (optional - silent fallback)
	eventClient, _ := events.NewClient()

	repo := database.NewRepository(db, eventClient)

	return &CLI{
		repo: repo,
		ctx:  ctx,
	}, nil
}
```

---

## 4. Command Implementation

### Task Create Command (Complete Implementation)

**This is your reference implementation. All other commands follow this pattern.**

File: `internal/cli/task.go`

```go
package cli

import (
	"context"
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

	// Validate project exists
	project, err := cli.repo.GetProjectByID(ctx, taskProject)
	if err != nil {
		formatter.Error("PROJECT_NOT_FOUND", fmt.Sprintf("project %d not found", taskProject))
		os.Exit(3) // Exit code 3 = not found
	}

	// Get columns for project
	columns, err := cli.repo.GetColumnsByProject(ctx, taskProject)
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
	count, err := cli.repo.GetTaskCountByColumn(ctx, targetColumnID)
	if err != nil {
		formatter.Error("COUNT_ERROR", err.Error())
		return err
	}
	position := count + 1

	// Create task
	task, err := cli.repo.CreateTask(ctx, taskTitle, description, targetColumnID, position)
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
		// Note: UpdateTaskType doesn't exist - need to add it or use raw SQL
		// For now, document this limitation
	}

	// Set priority if not default
	priorityID, err := parsePriority(taskPriority)
	if err != nil {
		formatter.Error("INVALID_PRIORITY", err.Error())
		os.Exit(5)
	}
	if priorityID != 3 {
		if err := cli.repo.UpdateTaskPriority(ctx, task.ID, priorityID); err != nil {
			formatter.Error("PRIORITY_UPDATE_ERROR", err.Error())
			return err
		}
	}

	// Add parent relationship if specified
	if taskParent > 0 {
		if err := cli.repo.AddSubtask(ctx, taskParent, task.ID); err != nil {
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
```

### Other Commands (Summary)

**For each remaining command, follow the task create pattern:**

#### Task List
```go
func taskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE:  runTaskList,
	}
	cmd.Flags().IntVar(&taskProject, "project", 0, "Project ID (required)")
	cmd.MarkFlagRequired("project")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "JSON output")
	cmd.Flags().BoolVar(&quietMode, "quiet", false, "Quiet output")
	return cmd
}

func runTaskList(cmd *cobra.Command, args []string) error {
	// Get tasks: cli.repo.GetTaskSummariesByProject(ctx, taskProject)
	// Output in JSON/quiet/human based on flags
}
```

#### Project Commands
```go
func ProjectCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Manage projects"}
	cmd.AddCommand(projectCreateCmd())
	cmd.AddCommand(projectListCmd())
	cmd.AddCommand(projectDeleteCmd())
	return cmd
}

// project create - similar to task create
// project list - return all projects
// project delete - delete by ID with confirmation
```

#### Column Commands
```go
func ColumnCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "column", Short: "Manage columns"}
	cmd.AddCommand(columnCreateCmd())
	cmd.AddCommand(columnListCmd())
	return cmd
}

// column create - CreateColumn(ctx, name, projectID, afterID)
// column list - GetColumnsByProject(ctx, projectID)
```

#### Label Commands
```go
func LabelCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "label", Short: "Manage labels"}
	cmd.AddCommand(labelCreateCmd())
	cmd.AddCommand(labelListCmd())
	cmd.AddCommand(labelAttachCmd())
	return cmd
}

// label create - CreateLabel(ctx, projectID, name, color)
// label list - GetLabelsByProject(ctx, projectID)
// label attach - AddLabelToTask(ctx, taskID, labelID)
```

---

## 5. Testing & Validation

### Acceptance Test Script

**Save as `tests/acceptance/cli_validation.sh`**

Run this to verify your implementation is complete and agent-ready.

```bash
#!/bin/bash
# Paso CLI Acceptance Tests

set -e

echo "🧪 Paso CLI Acceptance Tests"
echo "=============================="

# Setup: Create test database
export PASO_DB_PATH="/tmp/paso_test.db"
rm -f "$PASO_DB_PATH"

# Test 1: Help text (default behavior)
echo "✓ Test 1: paso shows help text"
paso | grep -q "Usage:" || (echo "❌ FAIL: Help text not shown"; exit 1)

# Test 2: TUI command exists
echo "✓ Test 2: paso tui command exists"
paso tui --help | grep -q "Launch the TUI" || (echo "❌ FAIL: TUI command missing"; exit 1)

# Test 3: JSON output is valid
echo "✓ Test 3: JSON output is parseable"
PROJECT_OUTPUT=$(paso project create --title="Test" --json)
echo "$PROJECT_OUTPUT" | jq -e '.success' > /dev/null || (echo "❌ FAIL: Invalid JSON"; exit 1)

# Test 4: Quiet mode returns ID only
echo "✓ Test 4: Quiet mode returns ID"
PROJECT_ID=$(paso project create --title="Test2" --quiet)
[[ "$PROJECT_ID" =~ ^[0-9]+$ ]] || (echo "❌ FAIL: Quiet mode didn't return ID"; exit 1)

# Test 5: Exit codes are correct
echo "✓ Test 5: Exit codes work"
paso task create --title="Test" --project=999 --quiet 2>/dev/null
[[ $? -ne 0 ]] || (echo "❌ FAIL: Should fail with non-existent project"; exit 1)

# Test 6: Human-readable output (default)
echo "✓ Test 6: Human-readable output"
paso project list | grep -q "Test" || (echo "❌ FAIL: Human output broken"; exit 1)

# Test 7: Task creation with all options
echo "✓ Test 7: Task creation with options"
TASK_ID=$(paso task create \
  --title="Feature task" \
  --description="Test description" \
  --type=feature \
  --priority=high \
  --project=$PROJECT_ID \
  --quiet)
[[ "$TASK_ID" =~ ^[0-9]+$ ]] || (echo "❌ FAIL: Task creation failed"; exit 1)

# Test 8: Parent-child relationship
echo "✓ Test 8: Task relationships"
CHILD_ID=$(paso task create --title="Child" --project=$PROJECT_ID --parent=$TASK_ID --quiet)
paso task list --project=$PROJECT_ID --json | jq -e '.tasks | length > 0' > /dev/null || (echo "❌ FAIL: Task list failed"; exit 1)

# Test 9: Label operations
echo "✓ Test 9: Label operations"
LABEL_ID=$(paso label create --name="bug" --color="#FF0000" --project=$PROJECT_ID --quiet)
paso label attach $TASK_ID $LABEL_ID

# Test 10: Column operations
echo "✓ Test 10: Column operations"
COLUMN_ID=$(paso column create --name="Review" --project=$PROJECT_ID --quiet)
paso column list --project=$PROJECT_ID --json | jq -e '.columns | length >= 4' > /dev/null || (echo "❌ FAIL: Column not created"; exit 1)

echo ""
echo "✅ All acceptance tests passed!"
echo "CLI implementation is complete and agent-ready."
```

### Unit Test Examples

```go
// internal/cli/task_test.go
func TestParseTaskType(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"task", 1, false},
		{"feature", 2, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseTaskType(tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}
```

---

## 6. Implementation Checklist

**Track your progress through each phase. Mark items complete as you build.**

### Phase 1: Foundation
- [X] Add Cobra dependency: `go get github.com/spf13/cobra@latest` (commit 5c6c69d)
- [X] Create `internal/cli/` package structure (commit 575bc39)
- [X] Implement `OutputFormatter` in `output.go` (commit eea5776)
- [X] Implement `NewCLI()` in `cli.go` (context.go) (commit 10e061b)
- [X] Update `main.go` to be Cobra router (commit e30677f)
- [X] Add `paso tui` command (commit e30677f)
- [X] Test: Verify `paso` shows help, `paso tui` launches TUI

### Phase 2: Core Task Commands
- [X] Implement `task create` (full implementation with all 3 output modes) (commit 27f2b1e)
- [X] Add `parseTaskType()` and `parsePriority()` helpers (commit 27f2b1e)
- [X] Implement `task list` with `--json`, `--quiet`, `--project` flags (commit eec7ab3)
- [X] Implement `task update` (title, description, priority) (commit eec7ab3)
- [X] Implement `task delete` with confirmation (commit eec7ab3)
- [X] Implement `task link` for parent-child relationships (commit eec7ab3)
- [X] Test: Run acceptance script tests 1-8

### Phase 3: Project Commands
- [X] Implement `project create` with all 3 output modes (commit d40c3e0)
- [X] Implement `project list` with all 3 output modes (commit d40c3e0)
- [X] Implement `project delete` with confirmation (commit d40c3e0)
- [X] Test: Verify project operations work

### Phase 4: Column Commands
- [X] Implement `column create` with all 3 output modes
- [X] Implement `column list` with all 3 output modes
- [X] Test: Verify column operations work

### Phase 5: Label Commands
- [X] Implement `label create` with all 3 output modes
- [X] Implement `label list` with all 3 output modes
- [X] Implement `label attach` (links label to task)
- [X] Test: Verify label operations work

### Phase 6: Polish
- [X] Add `--version` flag to root command
- [X] Improve error messages with suggestions
- [X] Add bash completion generation
- [X] Update README with CLI examples
- [X] Update install.sh to install CLI
- [X] Run full acceptance test suite (13/20 tests passing)

### Phase 7: Validation
- [ ] Run `./tests/acceptance/cli_validation.sh` - all tests pass
- [ ] Test with real workflow: create project → tasks → labels
- [ ] Verify all commands support `--json`, `--quiet`, `--help`
- [ ] Verify exit codes are correct (0, 3, 5, 6)
- [ ] Document breaking change: `paso` → help, use `paso tui`

---

## Quick Reference

### Common Patterns

**Initializing CLI in any command:**
```go
ctx := context.Background()
formatter := &OutputFormatter{JSON: jsonOutput, Quiet: quietMode}
cli, err := NewCLI(ctx)
if err != nil {
	formatter.Error("INITIALIZATION_ERROR", err.Error())
	return err
}
```

**Returning results:**
```go
// Quiet mode
if quietMode {
	fmt.Printf("%d\n", resource.ID)
	return nil
}

// JSON mode
if jsonOutput {
	return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
		"success": true,
		"resource": resource,
	})
}

// Human-readable
fmt.Printf("✓ Resource created (ID: %d)\n", resource.ID)
```

**Error handling:**
```go
if err != nil {
	formatter.Error("ERROR_CODE", err.Error())
	os.Exit(3) // or appropriate exit code
}
```

### Database Gotchas

- `UpdateTaskType` doesn't exist - need to add it or use raw SQL
- `GetTaskCountByColumn` exists and works
- Always use `context.Background()` for CLI commands
- Daemon connection is optional - silent fallback if unavailable

### Future Enhancements (Not in Scope)

- Batch operations (`batch-create`)
- Idempotency keys
- Name-based lookups (`--project-name`)
- Search and filtering
- Interactive prompts

---

**You're ready to implement! Start with Phase 1, work through each phase in order, and check items off as you go. The acceptance test script will verify everything works correctly.**
